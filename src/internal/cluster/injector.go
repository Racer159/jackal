// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package cluster contains Zarf-specific cluster management functions.
package cluster

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/k8s"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/transform"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/types"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/mholt/archiver/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// The chunk size for the tarball chunks.
var payloadChunkSize = 1024 * 768

// StartInjectionMadness initializes a Zarf injection into the cluster.
func (c *Cluster) StartInjectionMadness(tempPath types.TempPaths, injectorSeedTags []string) {
	spinner := message.NewProgressSpinner("Attempting to bootstrap the seed image into the cluster")
	defer spinner.Stop()

	var err error
	var images k8s.ImageNodeMap
	var payloadConfigmaps []string
	var sha256sum string
	var seedImages []transform.Image

	// Get all the images from the cluster
	timeout := 5 * time.Minute
	spinner.Updatef("Getting the list of existing cluster images (%s timeout)", timeout.String())
	if images, err = c.GetAllImages(timeout); err != nil {
		spinner.Fatalf(err, "Unable to generate a list of candidate images to perform the registry injection")
	}

	spinner.Updatef("Creating the injector configmap")
	if err = c.createInjectorConfigmap(tempPath); err != nil {
		spinner.Fatalf(err, "Unable to create the injector configmap")
	}

	spinner.Updatef("Creating the injector service")
	if service, err := c.createService(); err != nil {
		spinner.Fatalf(err, "Unable to create the injector service")
	} else {
		config.ZarfSeedPort = fmt.Sprintf("%d", service.Spec.Ports[0].NodePort)
	}

	spinner.Updatef("Loading the seed image from the package")
	if seedImages, err = c.loadSeedImages(tempPath, injectorSeedTags, spinner); err != nil {
		spinner.Fatalf(err, "Unable to load the injector seed image from the package")
	}

	spinner.Updatef("Loading the seed registry configmaps")
	if payloadConfigmaps, sha256sum, err = c.createPayloadConfigmaps(tempPath, spinner); err != nil {
		spinner.Fatalf(err, "Unable to generate the injector payload configmaps")
	}

	// https://regex101.com/r/eLS3at/1
	zarfImageRegex := regexp.MustCompile(`(?m)^127\.0\.0\.1:`)

	// Try to create an injector pod using an existing image in the cluster
	for image, node := range images {
		// Don't try to run against the seed image if this is a secondary zarf init run
		if zarfImageRegex.MatchString(image) {
			continue
		}

		spinner.Updatef("Attempting to bootstrap with the %s/%s", node, image)

		// Make sure the pod is not there first
		_ = c.DeletePod(ZarfNamespaceName, "injector")

		// Update the podspec image path and use the first node found
		pod, err := c.buildInjectionPod(node[0], image, payloadConfigmaps, sha256sum)
		if err != nil {
			// Just debug log the output because failures just result in trying the next image
			message.Debug(err)
			continue
		}

		// Create the pod in the cluster
		pod, err = c.CreatePod(pod)
		if err != nil {
			// Just debug log the output because failures just result in trying the next image
			message.Debug(pod, err)
			continue
		}

		// if no error, try and wait for a seed image to be present, return if successful
		if c.injectorIsReady(seedImages, spinner) {
			spinner.Success()
			return
		}

		// Otherwise just continue to try next image
	}

	// All images were exhausted and still no happiness
	spinner.Fatalf(nil, "Unable to perform the injection")
}

// StopInjectionMadness handles cleanup once the seed registry is up.
func (c *Cluster) StopInjectionMadness() error {
	// Try to kill the injector pod now
	if err := c.DeletePod(ZarfNamespaceName, "injector"); err != nil {
		return err
	}

	// Remove the configmaps
	labelMatch := map[string]string{"zarf-injector": "payload"}
	if err := c.DeleteConfigMapsByLabel(ZarfNamespaceName, labelMatch); err != nil {
		return err
	}

	// Remove the injector service
	return c.DeleteService(ZarfNamespaceName, "zarf-injector")
}

func (c *Cluster) loadSeedImages(tempPath types.TempPaths, injectorSeedTags []string, spinner *message.Spinner) ([]transform.Image, error) {
	seedImages := []transform.Image{}
	tagToDigest := make(map[string]string)

	// Load the injector-specific images and save them as seed-images
	for _, src := range injectorSeedTags {
		spinner.Updatef("Loading the seed image '%s' from the package", src)

		img, err := utils.LoadOCIImage(tempPath.Images, src)
		if err != nil {
			return seedImages, err
		}

		crane.SaveOCI(img, tempPath.SeedImages)

		imgRef, err := transform.ParseImageRef(src)
		if err != nil {
			return seedImages, err
		}
		seedImages = append(seedImages, imgRef)

		// Get the image digest so we can set an annotation in the image.json later
		imgDigest, err := img.Digest()
		if err != nil {
			return seedImages, err
		}
		// This is done _without_ the domain (different from pull.go) since the injector only handles local images
		tagToDigest[imgRef.Path+imgRef.TagOrDigest] = imgDigest.String()
	}

	if err := utils.AddImageNameAnnotation(tempPath.SeedImages, tagToDigest); err != nil {
		return seedImages, fmt.Errorf("unable to format OCI layout: %w", err)
	}

	return seedImages, nil
}

func (c *Cluster) createPayloadConfigmaps(tempPath types.TempPaths, spinner *message.Spinner) ([]string, string, error) {
	var configMaps []string

	// Chunk size has to accommodate base64 encoding & etcd 1MB limit
	tarPath := filepath.Join(tempPath.Base, "payload.tgz")
	tarFileList, err := filepath.Glob(filepath.Join(tempPath.SeedImages, "*"))
	if err != nil {
		return configMaps, "", err
	}

	spinner.Updatef("Creating the seed registry archive to send to the cluster")
	// Create a tar archive of the injector payload
	if err := archiver.Archive(tarFileList, tarPath); err != nil {
		return configMaps, "", err
	}

	chunks, sha256sum, err := utils.SplitFile(tarPath, payloadChunkSize)
	if err != nil {
		return configMaps, "", err
	}

	spinner.Updatef("Splitting the archive into binary configmaps")

	chunkCount := len(chunks)

	// Loop over all chunks and generate configmaps
	for idx, data := range chunks {
		// Create a cat-friendly filename
		fileName := fmt.Sprintf("zarf-payload-%03d", idx)

		// Store the binary data
		configData := map[string][]byte{
			fileName: data,
		}

		spinner.Updatef("Adding archive binary configmap %d of %d to the cluster", idx+1, chunkCount)

		// Attempt to create the configmap in the cluster
		if _, err = c.ReplaceConfigmap(ZarfNamespaceName, fileName, configData); err != nil {
			return configMaps, "", err
		}

		// Add the configmap to the configmaps slice for later usage in the pod
		configMaps = append(configMaps, fileName)

		// Give the control plane a 250ms buffer between each configmap
		time.Sleep(250 * time.Millisecond)
	}

	return configMaps, sha256sum, nil
}

// Test for pod readiness and seed image presence.
func (c *Cluster) injectorIsReady(seedImages []transform.Image, spinner *message.Spinner) bool {
	// Establish the zarf connect tunnel
	tunnel, err := NewZarfTunnel()
	if err != nil {
		message.Warnf("Unable to establish a tunnel to look for seed images: %#v", err)
		return false
	}
	tunnel.AddSpinner(spinner)
	err = tunnel.Connect(ZarfInjector, false)
	if err != nil {
		return false
	}
	defer tunnel.Close()

	spinner.Updatef("Testing the injector for seed image availability")

	for _, seedImage := range seedImages {
		seedRegistry := fmt.Sprintf("%s/v2/%s/manifests/%s", tunnel.HTTPEndpoint(), seedImage.Path, seedImage.Tag)
		if resp, err := http.Get(seedRegistry); err != nil || resp.StatusCode != 200 {
			// Just debug log the output because failures just result in trying the next image
			message.Debug(resp, err)
			return false
		}
	}

	spinner.Updatef("Seed image found, injector is ready")
	return true
}

func (c *Cluster) createInjectorConfigmap(tempPath types.TempPaths) error {
	var err error
	configData := make(map[string][]byte)

	// Add the injector binary data to the configmap
	if configData["zarf-injector"], err = os.ReadFile(tempPath.InjectBinary); err != nil {
		return err
	}

	// Try to delete configmap silently
	_ = c.DeleteConfigmap(ZarfNamespaceName, "rust-binary")

	// Attempt to create the configmap in the cluster
	if _, err = c.CreateConfigmap(ZarfNamespaceName, "rust-binary", configData); err != nil {
		return err
	}

	return nil
}

func (c *Cluster) createService() (*corev1.Service, error) {
	service := c.GenerateService(ZarfNamespaceName, "zarf-injector")

	service.Spec.Type = corev1.ServiceTypeNodePort
	service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
		Port: int32(5000),
	})
	service.Spec.Selector = map[string]string{
		"app": "zarf-injector",
	}

	// Attempt to purse the service silently
	_ = c.DeleteService(ZarfNamespaceName, "zarf-injector")

	return c.CreateService(service)
}

// buildInjectionPod return a pod for injection with the appropriate containers to perform the injection.
func (c *Cluster) buildInjectionPod(node, image string, payloadConfigmaps []string, payloadShasum string) (*corev1.Pod, error) {
	pod := c.GeneratePod("injector", ZarfNamespaceName)
	executeMode := int32(0777)

	pod.Labels["app"] = "zarf-injector"

	// Ensure zarf agent doesn't break the injector on future runs
	pod.Labels[agentLabel] = "ignore"

	// Bind the pod to the node the image was found on
	pod.Spec.NodeSelector = map[string]string{"kubernetes.io/hostname": node}

	// Do not try to restart the pod as it will be deleted/re-created instead
	pod.Spec.RestartPolicy = corev1.RestartPolicyNever

	pod.Spec.Containers = []corev1.Container{
		{
			Name: "injector",

			// An existing image already present on the cluster
			Image: image,

			// PullIfNotPresent because some distros provide a way (even in airgap) to pull images from local or direct-connected registries
			ImagePullPolicy: corev1.PullIfNotPresent,

			// This directory's contents come from the init container output
			WorkingDir: "/zarf-init",

			// Call the injector with shasum of the tarball
			Command: []string{"/zarf-init/zarf-injector", payloadShasum},

			// Shared mount between the init and regular containers
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "init",
					MountPath: "/zarf-init/zarf-injector",
					SubPath:   "zarf-injector",
				},
				{
					Name:      "seed",
					MountPath: "/zarf-seed",
				},
			},

			// Readiness probe to optimize the pod startup time
			ReadinessProbe: &corev1.Probe{
				PeriodSeconds:    2,
				SuccessThreshold: 1,
				FailureThreshold: 10,
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/v2/",               // path to health check
						Port: intstr.FromInt(5000), // port to health check
					},
				},
			},

			// Keep resources as light as possible as we aren't actually running the container's other binaries
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(".5"),
					corev1.ResourceMemory: resource.MustParse("64Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
			},
		},
	}

	pod.Spec.Volumes = []corev1.Volume{
		// Contains the rust binary and collection of configmaps from the tarball (seed image).
		{
			Name: "init",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "rust-binary",
					},
					DefaultMode: &executeMode,
				},
			},
		},
		// Empty directory to hold the seed image (new dir to avoid permission issues)
		{
			Name: "seed",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	// Iterate over all the payload configmaps and add their mounts.
	for _, filename := range payloadConfigmaps {
		// Create the configmap volume from the given filename.
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
			Name: filename,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: filename,
					},
				},
			},
		})

		// Create the volume mount to place the new volume in the working directory
		pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      filename,
			MountPath: fmt.Sprintf("/zarf-init/%s", filename),
			SubPath:   filename,
		})
	}

	return pod, nil
}

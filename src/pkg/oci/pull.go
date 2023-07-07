// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package oci contains functions for interacting with Zarf packages stored in OCI registries.
package oci

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/pkg/utils/helpers"
	"github.com/defenseunicorns/zarf/src/types"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pterm/pterm"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
)

var (
	// AlwaysPull is a list of paths that will always be pulled from the remote repository.
	AlwaysPull = []string{config.ZarfYAML, config.ZarfChecksumsTxt, config.ZarfYAMLSignature}
)

// LayersFromPaths returns the descriptors for the given paths from the root manifest.
func (o *OrasRemote) LayersFromPaths(requestedPaths []string) (layers []ocispec.Descriptor, err error) {
	manifest, err := o.FetchRoot()
	if err != nil {
		return nil, err
	}
	for _, path := range requestedPaths {
		layer := manifest.Locate(path)
		if o.isEmptyDescriptor(layer) {
			return nil, fmt.Errorf("path %s does not exist in this package", path)
		}
		layers = append(layers, layer)
	}
	return layers, nil
}

// LayersFromRequestedComponents returns the descriptors for the given components from the root manifest.
//
// It also retrieves the descriptors for all image layers that are required by the components.
//
// It also respects the `required` flag on components, and will retrieve all necessary layers for required components.
func (o *OrasRemote) LayersFromRequestedComponents(requestedComponents []string) (layers []ocispec.Descriptor, err error) {
	root, err := o.FetchRoot()
	if err != nil {
		return nil, err
	}

	pkg, err := o.FetchZarfYAML(root)
	if err != nil {
		return nil, err
	}
	images := map[string]bool{}
	tarballFormat := "%s.tar"
	for _, name := range requestedComponents {
		component := helpers.Find(pkg.Components, func(component types.ZarfComponent) bool {
			return component.Name == name
		})
		if component.Name == "" {
			return nil, fmt.Errorf("component %s does not exist in this package", name)
		}
	}
	for _, component := range pkg.Components {
		// If we requested this component, or it is required, we need to pull its images and tarball
		if helpers.SliceContains(requestedComponents, component.Name) || component.Required {
			for _, image := range component.Images {
				images[image] = true
			}
			layers = append(layers, root.Locate(filepath.Join(config.ZarfComponentsDir, fmt.Sprintf(tarballFormat, component.Name))))
		}
	}
	// Append the sboms.tar layer if it exists
	//
	// Since sboms.tar is not a heavy addition 99% of the time, we'll just always pull it
	sbomsDescriptor := root.Locate(config.ZarfSBOMTar)
	if !o.isEmptyDescriptor(sbomsDescriptor) {
		layers = append(layers, sbomsDescriptor)
	}
	if len(images) > 0 {
		// Add the image index and the oci-layout layers
		layers = append(layers, root.Locate(root.indexPath), root.Locate(root.ociLayoutPath))
		index, err := o.FetchImagesIndex(root)
		if err != nil {
			return nil, err
		}
		for image := range images {
			manifestDescriptor := helpers.Find(index.Manifests, func(layer ocispec.Descriptor) bool {
				return layer.Annotations[ocispec.AnnotationBaseImageName] == image
			})
			manifest, err := o.FetchManifest(manifestDescriptor)
			if err != nil {
				return nil, err
			}
			// Add the manifest and the manifest config layers
			layers = append(layers, root.Locate(filepath.Join(root.imagesBlobsDir, manifestDescriptor.Digest.Encoded())))
			layers = append(layers, root.Locate(filepath.Join(root.imagesBlobsDir, manifest.Config.Digest.Encoded())))

			// Add all the layers from the manifest
			for _, layer := range manifest.Layers {
				layerPath := filepath.Join(root.imagesBlobsDir, layer.Digest.Encoded())
				layers = append(layers, root.Locate(layerPath))
			}
		}
	}
	return layers, nil
}

// PullPackage pulls the package from the remote repository and saves it to the given path.
//
// layersToPull is an optional parameter that allows the caller to specify which layers to pull.
//
// The following layers will ALWAYS be pulled if they exist:
//   - zarf.yaml
//   - checksums.txt
//   - zarf.yaml.sig
func (o *OrasRemote) PullPackage(destinationDir string, concurrency int, layersToPull ...ocispec.Descriptor) (partialPaths []string, err error) {
	isPartialPull := len(layersToPull) > 0
	ref := o.Reference

	pterm.Println()
	message.Debugf("Pulling %s", ref.String())
	message.Infof("Pulling Zarf package from %s", ref)

	manifest, err := o.FetchRoot()
	if err != nil {
		return partialPaths, err
	}

	estimatedBytes := int64(0)
	if isPartialPull {
		for _, path := range AlwaysPull {
			desc := manifest.Locate(path)
			layersToPull = append(layersToPull, desc)
		}
		for _, desc := range layersToPull {
			estimatedBytes += desc.Size
		}
	} else {
		estimatedBytes = manifest.SumLayersSize()
	}
	estimatedBytes += manifest.Config.Size

	dst, err := file.New(destinationDir)
	if err != nil {
		return partialPaths, err
	}
	defer dst.Close()

	copyOpts := o.CopyOpts
	copyOpts.Concurrency = concurrency
	if isPartialPull {
		paths := []string{}
		for _, layer := range layersToPull {
			paths = append(paths, layer.Annotations[ocispec.AnnotationTitle])
		}
		copyOpts.FindSuccessors = func(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
			nodes, err := content.Successors(ctx, fetcher, desc)
			if err != nil {
				return nil, err
			}
			var ret []ocispec.Descriptor
			for _, node := range nodes {
				if helpers.SliceContains(paths, node.Annotations[ocispec.AnnotationTitle]) {
					ret = append(ret, node)
				}
			}
			return ret, nil
		}
	}

	// Create a thread to update a progress bar as we save the package to disk
	doneSaving := make(chan int)
	var wg sync.WaitGroup
	wg.Add(1)
	go utils.RenderProgressBarForLocalDirWrite(destinationDir, estimatedBytes, &wg, doneSaving, "Pulling Zarf package data")
	_, err = oras.Copy(o.Context, o.Repository, ref.String(), dst, ref.String(), copyOpts)
	if err != nil {
		return partialPaths, err
	}

	// Send a signal to the progress bar that we're done and wait for it to finish
	doneSaving <- 1
	wg.Wait()

	message.Debugf("Pulled %s", ref.String())
	message.Successf("Pulled %s", ref.String())

	if isPartialPull {
		for _, layer := range layersToPull {
			if layer.Annotations[ocispec.AnnotationTitle] != "" {
				partialPaths = append(partialPaths, layer.Annotations[ocispec.AnnotationTitle])
			}
		}
	}

	return partialPaths, nil
}

// PullLayer pulls a layer from the remote repository and saves it to `destinationDir/annotationTitle`.
func (o *OrasRemote) PullLayer(desc ocispec.Descriptor, destinationDir string) error {
	if desc.MediaType != ZarfLayerMediaTypeBlob {
		return fmt.Errorf("invalid media type for file layer: %s", desc.MediaType)
	}
	b, err := o.FetchLayer(desc)
	if err != nil {
		return err
	}
	return utils.WriteFile(filepath.Join(destinationDir, desc.Annotations[ocispec.AnnotationTitle]), b)
}

// PullPackageMetadata pulls the package metadata from the remote repository and saves it to `destinationDir`.
func (o *OrasRemote) PullPackageMetadata(destinationDir string) (err error) {
	root, err := o.FetchRoot()
	if err != nil {
		return err
	}
	for _, path := range AlwaysPull {
		desc := root.Locate(path)
		if !o.isEmptyDescriptor(desc) {
			err = o.PullLayer(desc, destinationDir)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

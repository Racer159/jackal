// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package oci contains functions for interacting with Zarf packages stored in OCI registries.
package ocizarf

import (
	"fmt"
	"path/filepath"
	"slices"
	"sync"

	"github.com/defenseunicorns/zarf/src/pkg/layout"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/oci"
	"github.com/defenseunicorns/zarf/src/pkg/transform"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/pkg/utils/helpers"
	"github.com/defenseunicorns/zarf/src/types"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var (
	// PackageAlwaysPull is a list of paths that will always be pulled from the remote repository.
	PackageAlwaysPull = []string{layout.ZarfYAML, layout.Checksums, layout.Signature}
	// ZarfPackageIndexPath is the path to the index.json file in the OCI package.
	ZarfPackageIndexPath = filepath.Join("images", "index.json")
	// ZarfPackageImagesBlobsDir is the path to the directory containing the image blobs in the OCI package.
	ZarfPackageImagesBlobsDir = filepath.Join("images", "blobs", "sha256")
	// ZarfPackageLayoutPath is the path to the oci-layout file in the OCI package.
	ZarfPackageLayoutPath = filepath.Join("images", "oci-layout")
)

// PullPackage pulls the package from the remote repository and saves it to the given path.
//
// layersToPull is an optional parameter that allows the caller to specify which layers to pull.
//
// The following layers will ALWAYS be pulled if they exist:
//   - zarf.yaml
//   - checksums.txt
//   - zarf.yaml.sig
func (o *ZarfOrasRemote) PullPackage(destinationDir string, concurrency int, layersToPull ...ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	isPartialPull := len(layersToPull) > 0
	message.Debugf("Pulling %s", o.Repo().Reference)

	manifest, err := o.FetchRoot()
	if err != nil {
		return nil, err
	}

	if isPartialPull {
		for _, path := range PackageAlwaysPull {
			desc := manifest.Locate(path)
			layersToPull = append(layersToPull, desc)
		}
	} else {
		layersToPull = append(layersToPull, manifest.Layers...)
	}
	layersToPull = append(layersToPull, manifest.Config)

	// Create a thread to update a progress bar as we save the package to disk
	doneSaving := make(chan int)
	encounteredErr := make(chan int)
	var wg sync.WaitGroup
	wg.Add(1)
	successText := fmt.Sprintf("Pulling %q", helpers.OCIURLPrefix+o.Repo().Reference.String())

	layerSize := oci.SumLayersSize(layersToPull)
	go utils.RenderProgressBarForLocalDirWrite(destinationDir, layerSize, &wg, doneSaving, encounteredErr, "Pulling", successText)

	return o.PullLayers(destinationDir, concurrency, layersToPull, doneSaving, encounteredErr, &wg)
}

// LayersFromRequestedComponents returns the descriptors for the given components from the root manifest.
// It also retrieves the descriptors for all image layers that are required by the components.
//
// It also respects the `required` flag on components, and will retrieve all necessary layers for required components.
func LayersFromRequestedComponents(o *ZarfOrasRemote, requestedComponents []string) (layers []ocispec.Descriptor, err error) {
	root, err := o.FetchRoot()
	if err != nil {
		return nil, err
	}

	pkg, err := o.FetchZarfYAML()
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
		if slices.Contains(requestedComponents, component.Name) || component.Required {
			for _, image := range component.Images {
				images[image] = true
			}
			layers = append(layers, root.Locate(filepath.Join(layout.ComponentsDir, fmt.Sprintf(tarballFormat, component.Name))))
		}
	}
	// Append the sboms.tar layer if it exists
	//
	// Since sboms.tar is not a heavy addition 99% of the time, we'll just always pull it
	sbomsDescriptor := root.Locate(layout.SBOMTar)
	if !oci.IsEmptyDescriptor(sbomsDescriptor) {
		layers = append(layers, sbomsDescriptor)
	}
	if len(images) > 0 {
		// Add the image index and the oci-layout layers
		layers = append(layers, root.Locate(ZarfPackageIndexPath), root.Locate(ZarfPackageLayoutPath))
		index, err := o.FetchImagesIndex()
		if err != nil {
			return nil, err
		}
		for image := range images {
			// use docker's transform lib to parse the image ref
			// this properly mirrors the logic within create
			refInfo, err := transform.ParseImageRef(image)
			if err != nil {
				return nil, fmt.Errorf("failed to parse image ref %q: %w", image, err)
			}

			manifestDescriptor := helpers.Find(index.Manifests, func(layer ocispec.Descriptor) bool {
				return layer.Annotations[ocispec.AnnotationBaseImageName] == refInfo.Reference ||
					// A backwards compatibility shim for older Zarf versions that would leave docker.io off of image annotations
					(layer.Annotations[ocispec.AnnotationBaseImageName] == refInfo.Path+refInfo.TagOrDigest && refInfo.Host == "docker.io")
			})

			// even though these are technically image manifests, we store them as Zarf blobs
			manifestDescriptor.MediaType = ZarfLayerMediaTypeBlob

			manifest, err := o.FetchManifest(manifestDescriptor)
			if err != nil {
				return nil, err
			}
			// Add the manifest and the manifest config layers
			layers = append(layers, root.Locate(filepath.Join(ZarfPackageImagesBlobsDir, manifestDescriptor.Digest.Encoded())))
			layers = append(layers, root.Locate(filepath.Join(ZarfPackageImagesBlobsDir, manifest.Config.Digest.Encoded())))

			// Add all the layers from the manifest
			for _, layer := range manifest.Layers {
				layerPath := filepath.Join(ZarfPackageImagesBlobsDir, layer.Digest.Encoded())
				layers = append(layers, root.Locate(layerPath))
			}
		}
	}
	return layers, nil
}

// PullPackageMetadata pulls the package metadata from the remote repository and saves it to `destinationDir`.
func (o *ZarfOrasRemote) PullPackageMetadata(destinationDir string) ([]ocispec.Descriptor, error) {
	return o.PullFilesAtPaths(PackageAlwaysPull, destinationDir)
}

// PullPackageSBOM pulls the package's sboms.tar from the remote repository and saves it to `destinationDir`.
func (o *ZarfOrasRemote) PullPackageSBOM(destinationDir string) ([]ocispec.Descriptor, error) {
	return o.PullFilesAtPaths([]string{layout.SBOMTar}, destinationDir)
}

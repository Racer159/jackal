// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package zoci contains functions for interacting with Jackal packages stored in OCI registries.
package zoci

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/defenseunicorns/pkg/helpers"
	"github.com/defenseunicorns/pkg/oci"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/racer159/jackal/src/pkg/layout"
	"github.com/racer159/jackal/src/pkg/transform"
	"github.com/racer159/jackal/src/pkg/utils"
	"github.com/racer159/jackal/src/types"
	"oras.land/oras-go/v2/content/file"
)

var (
	// PackageAlwaysPull is a list of paths that will always be pulled from the remote repository.
	PackageAlwaysPull = []string{layout.JackalYAML, layout.Checksums, layout.Signature}
)

// PullPackage pulls the package from the remote repository and saves it to the given path.
//
// layersToPull is an optional parameter that allows the caller to specify which layers to pull.
//
// The following layers will ALWAYS be pulled if they exist:
//   - jackal.yaml
//   - checksums.txt
//   - jackal.yaml.sig
func (r *Remote) PullPackage(ctx context.Context, destinationDir string, concurrency int, layersToPull ...ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	isPartialPull := len(layersToPull) > 0
	r.Log().Debug(fmt.Sprintf("Pulling %s", r.Repo().Reference))

	manifest, err := r.FetchRoot(ctx)
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
	doneSaving := make(chan error)
	successText := fmt.Sprintf("Pulling %q", helpers.OCIURLPrefix+r.Repo().Reference.String())

	layerSize := oci.SumDescsSize(layersToPull)
	go utils.RenderProgressBarForLocalDirWrite(destinationDir, layerSize, doneSaving, "Pulling", successText)

	dst, err := file.New(destinationDir)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	copyOpts := r.GetDefaultCopyOpts()
	copyOpts.Concurrency = concurrency

	err = r.CopyToTarget(ctx, layersToPull, dst, copyOpts)
	doneSaving <- err
	<-doneSaving
	return layersToPull, err
}

// LayersFromRequestedComponents returns the descriptors for the given components from the root manifest.
//
// It also retrieves the descriptors for all image layers that are required by the components.
func (r *Remote) LayersFromRequestedComponents(ctx context.Context, requestedComponents []types.JackalComponent) (layers []ocispec.Descriptor, err error) {
	root, err := r.FetchRoot(ctx)
	if err != nil {
		return nil, err
	}

	pkg, err := r.FetchJackalYAML(ctx)
	if err != nil {
		return nil, err
	}
	tarballFormat := "%s.tar"
	images := map[string]bool{}
	for _, rc := range requestedComponents {
		component := helpers.Find(pkg.Components, func(component types.JackalComponent) bool {
			return component.Name == rc.Name
		})
		if component.Name == "" {
			return nil, fmt.Errorf("component %s does not exist in this package", rc.Name)
		}
		for _, image := range component.Images {
			images[image] = true
		}
		layers = append(layers, root.Locate(filepath.Join(layout.ComponentsDir, fmt.Sprintf(tarballFormat, component.Name))))
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
		layers = append(layers, root.Locate(layout.IndexPath), root.Locate(layout.OCILayoutPath))
		index, err := r.FetchImagesIndex(ctx)
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
					// A backwards compatibility shim for older Jackal versions that would leave docker.io off of image annotations
					(layer.Annotations[ocispec.AnnotationBaseImageName] == refInfo.Path+refInfo.TagOrDigest && refInfo.Host == "docker.io")
			})

			// even though these are technically image manifests, we store them as Jackal blobs
			manifestDescriptor.MediaType = JackalLayerMediaTypeBlob

			manifest, err := r.FetchManifest(ctx, manifestDescriptor)
			if err != nil {
				return nil, err
			}
			// Add the manifest and the manifest config layers
			layers = append(layers, root.Locate(filepath.Join(layout.ImagesBlobsDir, manifestDescriptor.Digest.Encoded())))
			layers = append(layers, root.Locate(filepath.Join(layout.ImagesBlobsDir, manifest.Config.Digest.Encoded())))

			// Add all the layers from the manifest
			for _, layer := range manifest.Layers {
				layerPath := filepath.Join(layout.ImagesBlobsDir, layer.Digest.Encoded())
				layers = append(layers, root.Locate(layerPath))
			}
		}
	}
	return layers, nil
}

// PullPackageMetadata pulls the package metadata from the remote repository and saves it to `destinationDir`.
func (r *Remote) PullPackageMetadata(ctx context.Context, destinationDir string) ([]ocispec.Descriptor, error) {
	return r.PullPaths(ctx, destinationDir, PackageAlwaysPull)
}

// PullPackageSBOM pulls the package's sboms.tar from the remote repository and saves it to `destinationDir`.
func (r *Remote) PullPackageSBOM(ctx context.Context, destinationDir string) ([]ocispec.Descriptor, error) {
	return r.PullPaths(ctx, destinationDir, []string{layout.SBOMTar})
}

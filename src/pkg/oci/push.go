// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package oci contains functions for interacting with Zarf packages stored in OCI registries.
package oci

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/defenseunicorns/zarf/src/pkg/layout"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/types"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
)

// ConfigPartial is a partial OCI config that is used to create the manifest config.
//
// Unless specified, an empty manifest config will be used: `{}`
// which causes an error on Google Artifact Registry
//
// to negate this, we create a simple manifest config with some build metadata
//
// the contents of this file are not used by Zarf
type ConfigPartial struct {
	Architecture string            `json:"architecture"`
	OCIVersion   string            `json:"ociVersion"`
	Annotations  map[string]string `json:"annotations,omitempty"`
}

// PushLayer pushes the given layer (bytes) to the remote repository.
func (o *OrasRemote) PushLayer(b []byte, mediaType string) (ocispec.Descriptor, error) {
	desc := content.NewDescriptorFromBytes(mediaType, b)
	return desc, o.repo.Push(o.ctx, desc, bytes.NewReader(b))
}

// pushManifestConfigFromMetadata pushes the manifest config with metadata to the remote repository.
func (o *OrasRemote) pushManifestConfigFromMetadata(metadata *types.ZarfMetadata, build *types.ZarfBuildData) (ocispec.Descriptor, error) {
	annotations := map[string]string{
		ocispec.AnnotationTitle:       metadata.Name,
		ocispec.AnnotationDescription: metadata.Description,
	}
	manifestConfig := ConfigPartial{
		Architecture: build.Architecture,
		OCIVersion:   "1.0.1",
		Annotations:  annotations,
	}
	manifestConfigBytes, err := json.Marshal(manifestConfig)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	return o.PushLayer(manifestConfigBytes, ZarfConfigMediaType)
}

// manifestAnnotationsFromMetadata returns the annotations for the manifest from the given metadata.
func (o *OrasRemote) manifestAnnotationsFromMetadata(metadata *types.ZarfMetadata) map[string]string {
	annotations := map[string]string{
		ocispec.AnnotationDescription: metadata.Description,
	}

	if url := metadata.URL; url != "" {
		annotations[ocispec.AnnotationURL] = url
	}
	if authors := metadata.Authors; authors != "" {
		annotations[ocispec.AnnotationAuthors] = authors
	}
	if documentation := metadata.Documentation; documentation != "" {
		annotations[ocispec.AnnotationDocumentation] = documentation
	}
	if source := metadata.Source; source != "" {
		annotations[ocispec.AnnotationSource] = source
	}
	if vendor := metadata.Vendor; vendor != "" {
		annotations[ocispec.AnnotationVendor] = vendor
	}

	return annotations
}

func (o *OrasRemote) generatePackManifest(src *file.Store, descs []ocispec.Descriptor, configDesc *ocispec.Descriptor, metadata *types.ZarfMetadata) (ocispec.Descriptor, error) {
	packOpts := oras.PackOptions{}
	packOpts.ConfigDescriptor = configDesc
	packOpts.PackImageManifest = true
	packOpts.ManifestAnnotations = o.manifestAnnotationsFromMetadata(metadata)

	root, err := oras.Pack(o.ctx, src, ocispec.MediaTypeImageManifest, descs, packOpts)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	if err = src.Tag(o.ctx, root, root.Digest.String()); err != nil {
		return ocispec.Descriptor{}, err
	}

	return root, nil
}

// PublishPackage publishes the package to the remote repository.
func (o *OrasRemote) PublishPackage(pkg *types.ZarfPackage, paths *layout.PackagePaths, concurrency int) error {
	ctx := o.ctx
	// source file store
	src, err := file.New(paths.Base)
	if err != nil {
		return err
	}
	defer src.Close()

	message.Infof("Publishing package to %s", o.repo.Reference)
	spinner := message.NewProgressSpinner("")
	defer spinner.Stop()

	// Get all of the layers in the package
	var descs []ocispec.Descriptor
	for name, path := range paths.Files() {
		spinner.Updatef("Preparing layer %s", utils.First30last30(name))

		mediaType := ZarfLayerMediaTypeBlob

		desc, err := src.Add(ctx, name, mediaType, path)
		if err != nil {
			return err
		}
		descs = append(descs, desc)
	}
	spinner.Successf("Prepared all layers")

	copyOpts := o.CopyOpts
	copyOpts.Concurrency = concurrency
	var total int64
	for _, desc := range descs {
		total += desc.Size
	}
	// assumes referrers API is not supported since OCI artifact
	// media type is not supported
	o.repo.SetReferrersCapability(false)

	// push the manifest config
	// since this config is so tiny, and the content is not used again
	// it is not logged to the progress, but will error if it fails
	manifestConfigDesc, err := o.pushManifestConfigFromMetadata(&pkg.Metadata, &pkg.Build)
	if err != nil {
		return err
	}
	root, err := o.generatePackManifest(src, descs, &manifestConfigDesc, &pkg.Metadata)
	if err != nil {
		return err
	}
	total += root.Size + manifestConfigDesc.Size

	o.Transport.ProgressBar = message.NewProgressBar(total, fmt.Sprintf("Publishing %s:%s", o.repo.Reference.Repository, o.repo.Reference.Reference))
	defer o.Transport.ProgressBar.Stop()
	// attempt to push the image manifest
	_, err = oras.Copy(ctx, src, root.Digest.String(), o.repo, o.repo.Reference.Reference, copyOpts)
	if err != nil {
		return err
	}

	o.Transport.ProgressBar.Successf("Published %s [%s]", o.repo.Reference, root.MediaType)

	return nil
}

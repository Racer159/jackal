// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package packager contains functions for interacting with, managing and deploying Zarf packages.
package packager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/types"
	"github.com/mholt/archiver/v3"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry"
)

// ZarfLayerMediaTypeBlob is the media type for all Zarf layers due to the range of possible content
const (
	ZarfLayerMediaTypeBlob = "application/vnd.zarf.layer.v1.blob"
	skeletonSuffix         = "skeleton"
)

// Publish publishes the package to a registry
//
// This is a wrapper around the oras library
// and much of the code was adapted from the oras CLI - https://github.com/oras-project/oras/blob/main/cmd/oras/push.go
//
// Authentication is handled via the Docker config file created w/ `zarf tools registry login`
func (p *Packager) Publish() error {
	p.cfg.DeployOpts.PackagePath = p.cfg.PublishOpts.PackagePath
	var ref registry.Reference
	referenceSuffix := ""
	if utils.IsDir(p.cfg.PublishOpts.PackagePath) {
		referenceSuffix = skeletonSuffix
		err := p.loadSkeleton()
		if err != nil {
			return err
		}
	} else {
		// Extract the first layer of the tarball
		if err := archiver.Unarchive(p.cfg.DeployOpts.PackagePath, p.tmp.Base); err != nil {
			return fmt.Errorf("unable to extract the package: %w", err)
		}

		err := p.readYaml(p.tmp.ZarfYaml)
		if err != nil {
			return fmt.Errorf("unable to read the zarf.yaml in %s: %w", p.tmp.Base, err)
		}
		referenceSuffix = p.cfg.Pkg.Build.Architecture
	}

	// Get a reference to the registry for this package
	ref, err := p.ref(referenceSuffix)
	if err != nil {
		return err
	}

	if err := p.validatePackageChecksums(); err != nil {
		return fmt.Errorf("unable to publish package because checksums do not match: %w", err)
	}

	// Sign the package if a key has been provided
	if p.cfg.PublishOpts.SigningKeyPath != "" {
		_, err := utils.CosignSignBlob(p.tmp.ZarfYaml, p.tmp.ZarfSig, p.cfg.PublishOpts.SigningKeyPath, p.getSigPublishPassword)
		if err != nil {
			return fmt.Errorf("unable to sign the package: %w", err)
		}
	}

	message.HeaderInfof("📦 PACKAGE PUBLISH %s:%s", p.cfg.Pkg.Metadata.Name, ref.Reference)
	return p.publish(ref)
}

func (p *Packager) publish(ref registry.Reference) error {
	message.Infof("Publishing package to %s", ref)
	spinner := message.NewProgressSpinner("")
	defer spinner.Stop()

	// Get all of the layers in the package
	paths := []string{}
	err := filepath.Walk(p.tmp.Base, func(path string, info os.FileInfo, err error) error {
		// Catch any errors that happened during the walk
		if err != nil {
			return err
		}

		// Add any resource that is not a directory to the paths of objects we will include into the package
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to get the layers in the package to publish: %w", err)
	}

	// destination remote
	dst, err := utils.NewOrasRemote(ref)
	if err != nil {
		return err
	}
	ctx := dst.Context

	// source file store
	src, err := file.New(p.tmp.Base)
	if err != nil {
		return err
	}
	defer src.Close()

	var descs []ocispec.Descriptor

	for idx, path := range paths {
		name, err := filepath.Rel(p.tmp.Base, path)
		if err != nil {
			return err
		}
		spinner.Updatef("Preparing layer %d/%d: %s", idx+1, len(paths), name)

		mediaType := ZarfLayerMediaTypeBlob

		desc, err := src.Add(ctx, name, mediaType, path)
		if err != nil {
			return err
		}
		descs = append(descs, desc)
	}
	spinner.Successf("Prepared %d layers", len(descs))

	copyOpts := oras.DefaultCopyOptions
	copyOpts.Concurrency = p.cfg.PublishOpts.CopyOptions.Concurrency
	copyOpts.OnCopySkipped = utils.PrintLayerExists
	copyOpts.PostCopy = utils.PrintLayerExists

	root, err := p.publishImage(dst, src, descs, copyOpts)
	if err != nil {
		return err
	}

	dst.Transport.ProgressBar.Successf("Published %s [%s]", ref, root.MediaType)
	fmt.Println()
	if strings.HasSuffix(ref.Reference, skeletonSuffix) {
		message.Info("Example of importing components from this package:")
		fmt.Println()
		ex := []types.ZarfComponent{}
		for _, c := range p.cfg.Pkg.Components {
			ex = append(ex, types.ZarfComponent{
				Name: fmt.Sprintf("import-%s", c.Name),
				Import: types.ZarfComponentImport{
					ComponentName: c.Name,
					URL:           fmt.Sprintf("oci://%s", ref.String()),
				},
			})
		}
		utils.ColorPrintYAML(ex)
		fmt.Println()
	} else {
		flags := ""
		if config.CommonOptions.Insecure {
			flags = "--insecure"
		}
		message.Info("To inspect/deploy/pull:")
		message.Infof("zarf package inspect oci://%s %s", ref, flags)
		message.Infof("zarf package deploy oci://%s %s", ref, flags)
		message.Infof("zarf package pull oci://%s %s", ref, flags)
	}

	return nil
}

func (p *Packager) publishImage(dst *utils.OrasRemote, src *file.Store, descs []ocispec.Descriptor, copyOpts oras.CopyOptions) (root ocispec.Descriptor, err error) {
	var total int64
	for _, desc := range descs {
		total += desc.Size
	}
	// assumes referrers API is not supported since OCI artifact
	// media type is not supported
	dst.SetReferrersCapability(false)

	// fallback to an ImageManifest push
	manifestConfigDesc, manifestConfigContent, err := p.generateManifestConfigFile()
	if err != nil {
		return root, err
	}
	// push the manifest config
	// since this config is so tiny, and the content is not used again
	// it is not logged to the progress, but will error if it fails
	err = dst.Push(dst.Context, manifestConfigDesc, bytes.NewReader(manifestConfigContent))
	if err != nil {
		return root, err
	}
	packOpts := p.cfg.PublishOpts.PackOptions
	packOpts.ConfigDescriptor = &manifestConfigDesc
	packOpts.PackImageManifest = true
	root, err = p.pack(dst.Context, ocispec.MediaTypeImageManifest, descs, src, packOpts)
	if err != nil {
		return root, err
	}
	total += root.Size + manifestConfigDesc.Size

	dst.Transport.ProgressBar = message.NewProgressBar(total, fmt.Sprintf("Publishing %s:%s", dst.Reference.Repository, dst.Reference.Reference))
	defer dst.Transport.ProgressBar.Stop()
	// attempt to push the image manifest
	_, err = oras.Copy(dst.Context, src, root.Digest.String(), dst, dst.Reference.Reference, copyOpts)
	if err != nil {
		return root, err
	}

	return root, nil
}

func (p *Packager) generateAnnotations() map[string]string {
	annotations := map[string]string{
		ocispec.AnnotationDescription: p.cfg.Pkg.Metadata.Description,
	}

	if url := p.cfg.Pkg.Metadata.URL; url != "" {
		annotations[ocispec.AnnotationURL] = url
	}
	if authors := p.cfg.Pkg.Metadata.Authors; authors != "" {
		annotations[ocispec.AnnotationAuthors] = authors
	}
	if documentation := p.cfg.Pkg.Metadata.Documentation; documentation != "" {
		annotations[ocispec.AnnotationDocumentation] = documentation
	}
	if source := p.cfg.Pkg.Metadata.Source; source != "" {
		annotations[ocispec.AnnotationSource] = source
	}
	if vendor := p.cfg.Pkg.Metadata.Vendor; vendor != "" {
		annotations[ocispec.AnnotationVendor] = vendor
	}

	return annotations
}

func (p *Packager) generateManifestConfigFile() (ocispec.Descriptor, []byte, error) {
	// Unless specified, an empty manifest config will be used: `{}`
	// which causes an error on Google Artifact Registry
	// to negate this, we create a simple manifest config with some build metadata
	// the contents of this file are not used by Zarf
	type OCIConfigPartial struct {
		Architecture string            `json:"architecture"`
		OCIVersion   string            `json:"ociVersion"`
		Annotations  map[string]string `json:"annotations,omitempty"`
	}

	annotations := map[string]string{
		ocispec.AnnotationTitle:       p.cfg.Pkg.Metadata.Name,
		ocispec.AnnotationDescription: p.cfg.Pkg.Metadata.Description,
	}

	manifestConfig := OCIConfigPartial{
		Architecture: p.cfg.Pkg.Build.Architecture,
		OCIVersion:   "1.0.1",
		Annotations:  annotations,
	}
	manifestConfigBytes, err := json.Marshal(manifestConfig)
	if err != nil {
		return ocispec.Descriptor{}, nil, err
	}
	manifestConfigDesc := content.NewDescriptorFromBytes("application/vnd.unknown.config.v1+json", manifestConfigBytes)

	return manifestConfigDesc, manifestConfigBytes, nil
}

func (p *Packager) loadSkeleton() error {
	base, err := filepath.Abs(p.cfg.PublishOpts.PackagePath)
	if err != nil {
		return err
	}
	if err := os.Chdir(base); err != nil {
		return err
	}
	if err := p.readYaml(config.ZarfYAML); err != nil {
		return fmt.Errorf("unable to read the zarf.yaml in %s: %s", base, err.Error())
	}

	err = p.composeComponents()
	if err != nil {
		return err
	}

	err = p.skeletonizeExtensions()
	if err != nil {
		return err
	}

	for _, warning := range p.warnings {
		message.Warn(warning)
	}

	for idx, component := range p.cfg.Pkg.Components {
		isSkeleton := true
		err := p.addComponent(idx, component, isSkeleton)
		if err != nil {
			return err
		}

		err = p.archiveComponent(component)
		if err != nil {
			return fmt.Errorf("unable to archive component: %s", err.Error())
		}
	}

	checksumChecksum, err := generatePackageChecksums(p.tmp.Base)
	if err != nil {
		return fmt.Errorf("unable to generate checksums for skeleton package: %w", err)
	}
	p.cfg.Pkg.Metadata.AggregateChecksum = checksumChecksum

	return p.writeYaml()
}

// pack creates an artifact/image manifest from the provided descriptors and pushes it to the store
func (p *Packager) pack(ctx context.Context, artifactType string, descs []ocispec.Descriptor, src *file.Store, packOpts oras.PackOptions) (ocispec.Descriptor, error) {
	packOpts.ManifestAnnotations = p.generateAnnotations()
	root, err := oras.Pack(ctx, src, artifactType, descs, packOpts)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	if err = src.Tag(ctx, root, root.Digest.String()); err != nil {
		return ocispec.Descriptor{}, err
	}

	return root, nil
}

// ref returns a registry.Reference using metadata from the package's build config and the PublishOpts
//
// if suffix is not empty, the architecture will be replaced with the suffix string
func (p *Packager) ref(suffix string) (registry.Reference, error) {
	ver := p.cfg.Pkg.Metadata.Version
	if len(ver) == 0 {
		return registry.Reference{}, errors.New("version is required for publishing")
	}

	ref := registry.Reference{
		Registry:   p.cfg.PublishOpts.Reference.Registry,
		Repository: fmt.Sprintf("%s/%s", p.cfg.PublishOpts.Reference.Repository, p.cfg.Pkg.Metadata.Name),
		Reference:  fmt.Sprintf("%s-%s", ver, suffix),
	}
	if len(p.cfg.PublishOpts.Reference.Repository) == 0 {
		ref.Repository = p.cfg.Pkg.Metadata.Name
	}
	err := ref.Validate()
	if err != nil {
		return registry.Reference{}, err
	}
	return ref, nil
}

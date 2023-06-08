// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package packager contains functions for interacting with, managing and deploying Zarf packages.
package packager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/mholt/archiver/v3"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
)

// Pull pulls a Zarf package and saves it as a compressed tarball.
func (p *Packager) Pull() error {
	err := p.handleOciPackage(p.cfg.DeployOpts.PackagePath, p.tmp.Base, p.cfg.PullOpts.CopyOptions.Concurrency)
	if err != nil {
		return err
	}
	err = utils.ReadYaml(p.tmp.ZarfYaml, &p.cfg.Pkg)
	if err != nil {
		return err
	}

	if err = p.validatePackageSignature(p.cfg.PullOpts.PublicKeyPath); err != nil {
		return err
	} else if !config.CommonOptions.Insecure {
		message.Successf("Package signature is valid")
	}

	if p.cfg.Pkg.Metadata.AggregateChecksum != "" {
		if err = p.validatePackageChecksums(); err != nil {
			return fmt.Errorf("unable to validate the package checksums: %w", err)
		}
	}

	// Get all the layers from within the temp directory
	allTheLayers, err := filepath.Glob(filepath.Join(p.tmp.Base, "*"))
	if err != nil {
		return err
	}

	var name string
	if strings.HasSuffix(p.cfg.DeployOpts.PackagePath, skeletonSuffix) {
		name = fmt.Sprintf("zarf-package-%s-skeleton-%s.tar.zst", p.cfg.Pkg.Metadata.Name, p.cfg.Pkg.Metadata.Version)
	} else {
		name = fmt.Sprintf("zarf-package-%s-%s-%s.tar.zst", p.cfg.Pkg.Metadata.Name, p.cfg.Pkg.Build.Architecture, p.cfg.Pkg.Metadata.Version)
	}
	output := filepath.Join(p.cfg.PullOpts.OutputDirectory, name)
	_ = os.Remove(output)
	err = archiver.Archive(allTheLayers, output)
	if err != nil {
		return err
	}
	return nil
}

// pullPackageSpecLayer pulls the `zarf.yaml` and `zarf.yaml.sig` (if it exists) layers from the published package
func (p *Packager) pullPackageLayers(packagePath string, targetDir string, layersToPull []string) error {
	ref, err := registry.ParseReference(strings.TrimPrefix(packagePath, utils.OCIURLPrefix))
	if err != nil {
		return err
	}

	dst, err := utils.NewOrasRemote(ref)
	if err != nil {
		return err
	}

	// get the manifest
	manifest, err := getManifest(dst)
	if err != nil {
		return err
	}
	layers := manifest.Layers

	for _, layerToPull := range layersToPull {
		layerDesc := utils.Find(layers, func(d ocispec.Descriptor) bool {
			return d.Annotations[ocispec.AnnotationTitle] == layerToPull
		})
		if len(layerDesc.Digest) == 0 {
			return fmt.Errorf("unable to find layer (%s) from the OCI package %s", layerToPull, packagePath)
		}
		if err := pullLayer(dst, layerDesc, filepath.Join(targetDir, layerToPull)); err != nil {
			return fmt.Errorf("unable to pull the layer (%s) from the OCI package %s", layerToPull, packagePath)
		}
	}
	return nil
}

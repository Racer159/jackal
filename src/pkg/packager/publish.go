// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package packager contains functions for interacting with, managing and deploying Zarf packages.
package packager

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/oci"
	"github.com/defenseunicorns/zarf/src/pkg/packager/sources"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/pkg/utils/helpers"
	"github.com/defenseunicorns/zarf/src/pkg/zoci"
	"github.com/defenseunicorns/zarf/src/types"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
)

// Publish publishes the package to a registry
func (p *Packager) Publish() (err error) {
	_, isOCISource := p.source.(*sources.OCISource)
	if isOCISource && p.cfg.PublishOpts.SigningKeyPath == "" {
		ctx := context.TODO()
		// oci --> oci is a special case, where we will use oci.CopyPackage so that we can transfer the package
		// w/o layers touching the filesystem
		srcRemote := p.source.(*sources.OCISource).Remote

		parts := strings.Split(srcRemote.Repo().Reference.Repository, "/")
		packageName := parts[len(parts)-1]

		p.cfg.PublishOpts.PackageDestination = p.cfg.PublishOpts.PackageDestination + "/" + packageName

		arch := config.GetArch()
		dstRemote, err := zoci.NewRemote(p.cfg.PublishOpts.PackageDestination, oci.PlatformForArch(arch))
		if err != nil {
			return err
		}

		srcRoot, err := srcRemote.ResolveRoot(ctx)
		if err != nil {
			return err
		}

		pkg, err := srcRemote.FetchZarfYAML(ctx)
		if err != nil {
			return err
		}

		// ensure cli arch matches package arch
		if pkg.Build.Architecture != arch {
			return fmt.Errorf("architecture mismatch (specified: %q, found %q)", arch, pkg.Build.Architecture)
		}

		if err := zoci.CopyPackage(ctx, srcRemote, dstRemote, nil, config.CommonOptions.OCIConcurrency); err != nil {
			return err
		}

		srcManifest, err := srcRemote.FetchRoot(ctx)
		if err != nil {
			return err
		}
		b, err := srcManifest.MarshalJSON()
		if err != nil {
			return err
		}
		expected := content.NewDescriptorFromBytes(ocispec.MediaTypeImageManifest, b)

		if err := dstRemote.Repo().Manifests().PushReference(ctx, expected, bytes.NewReader(b), srcRoot.Digest.String()); err != nil {
			return err
		}

		tag := srcRemote.Repo().Reference.Reference
		if err := dstRemote.UpdateIndex(ctx, tag, expected); err != nil {
			return err
		}
		message.Infof("Published %s to %s", srcRemote.Repo().Reference, dstRemote.Repo().Reference)
		return nil
	}

	if p.cfg.CreateOpts.IsSkeleton {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		if err := p.cdToBaseDir(p.cfg.CreateOpts.BaseDir, cwd); err != nil {
			return err
		}
		if err := p.load(); err != nil {
			return err
		}
		if err := p.assembleSkeleton(); err != nil {
			return err
		}
	} else {
		if err = p.source.LoadPackage(p.layout, false); err != nil {
			return fmt.Errorf("unable to load the package: %w", err)
		}
		if err = p.readZarfYAML(p.layout.ZarfYAML); err != nil {
			return err
		}
	}

	// Get a reference to the registry for this package
	ref, err := zoci.ReferenceFromMetadata(p.cfg.PublishOpts.PackageDestination, &p.cfg.Pkg.Metadata, &p.cfg.Pkg.Build)
	if err != nil {
		return err
	}
	var platform ocispec.Platform
	if p.cfg.CreateOpts.IsSkeleton {
		platform = zoci.PlatformForSkeleton()
	} else {
		platform = oci.PlatformForArch(config.GetArch())
	}
	remote, err := zoci.NewRemote(ref, platform)
	if err != nil {
		return err
	}

	// Sign the package if a key has been provided
	if p.cfg.PublishOpts.SigningKeyPath != "" {
		if err := p.signPackage(p.cfg.PublishOpts.SigningKeyPath, p.cfg.PublishOpts.SigningKeyPassword); err != nil {
			return err
		}
	}

	message.HeaderInfof("📦 PACKAGE PUBLISH %s:%s", p.cfg.Pkg.Metadata.Name, ref)

	// Publish the package/skeleton to the registry
	ctx := context.TODO()
	if err := remote.PublishZarfPackage(ctx, &p.cfg.Pkg, p.layout, config.CommonOptions.OCIConcurrency); err != nil {
		return err
	}
	if p.cfg.CreateOpts.IsSkeleton {
		message.Title("How to import components from this skeleton:", "")
		ex := []types.ZarfComponent{}
		for _, c := range p.cfg.Pkg.Components {
			ex = append(ex, types.ZarfComponent{
				Name: fmt.Sprintf("import-%s", c.Name),
				Import: types.ZarfComponentImport{
					ComponentName: c.Name,
					URL:           helpers.OCIURLPrefix + remote.Repo().Reference.String(),
				},
			})
		}
		utils.ColorPrintYAML(ex, nil, true)
	}
	return nil
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package sources contains core implementations of the PackageSource interface.
package sources

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/layout"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/packager/filters"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/pkg/zoci"
	"github.com/defenseunicorns/zarf/src/types"
	"github.com/mholt/archiver/v3"
)

var (
	// verify that OCISource implements PackageSource
	_ PackageSource = (*OCISource)(nil)
)

// OCISource is a package source for OCI registries.
type OCISource struct {
	*types.ZarfPackageOptions
	*zoci.Remote
}

// LoadPackage loads a package from an OCI registry.
func (s *OCISource) LoadPackage(dst *layout.PackagePaths, filter filters.ComponentFilterStrategy, unarchiveAll bool, warnings *message.Warnings) (pkg types.ZarfPackage, err error) {
	ctx := context.TODO()

	message.Debugf("Loading package from %q", s.PackageSource)

	pkg, err = s.FetchZarfYAML(ctx)
	if err != nil {
		return pkg, err
	}
	pkg.Components, err = filter.Apply(pkg)
	if err != nil {
		return pkg, err
	}

	layersToPull, err := s.LayersFromRequestedComponents(ctx, pkg.Components)
	if err != nil {
		return pkg, fmt.Errorf("unable to get published component image layers: %s", err.Error())
	}

	isPartial := true
	root, err := s.FetchRoot(ctx)
	if err != nil {
		return pkg, err
	}
	if len(root.Layers) == len(layersToPull) {
		isPartial = false
	}

	layersFetched, err := s.PullPackage(ctx, dst.Base, config.CommonOptions.OCIConcurrency, layersToPull...)
	if err != nil {
		return pkg, fmt.Errorf("unable to pull the package: %w", err)
	}
	dst.SetFromLayers(layersFetched)

	if err := dst.MigrateLegacy(); err != nil {
		return pkg, err
	}

	if !dst.IsLegacyLayout() {
		spinner := message.NewProgressSpinner("Validating pulled layer checksums")
		defer spinner.Stop()

		if err := ValidatePackageIntegrity(dst, pkg.Metadata.AggregateChecksum, isPartial); err != nil {
			return pkg, err
		}

		spinner.Success()

		if err := ValidatePackageSignature(dst, s.PublicKeyPath); err != nil {
			return pkg, err
		}
	}

	if unarchiveAll {
		for _, component := range pkg.Components {
			if err := dst.Components.Unarchive(component); err != nil {
				if layout.IsNotLoaded(err) {
					_, err := dst.Components.Create(component)
					if err != nil {
						return pkg, err
					}
				} else {
					return pkg, err
				}
			}
		}

		if dst.SBOMs.Path != "" {
			if err := dst.SBOMs.Unarchive(); err != nil {
				return pkg, err
			}
		}
	}

	return pkg, nil
}

// LoadPackageMetadata loads a package's metadata from an OCI registry.
func (s *OCISource) LoadPackageMetadata(dst *layout.PackagePaths, wantSBOM bool, skipValidation bool, warnings *message.Warnings) (pkg types.ZarfPackage, err error) {
	toPull := zoci.PackageAlwaysPull
	if wantSBOM {
		toPull = append(toPull, layout.SBOMTar)
	}
	ctx := context.TODO()
	layersFetched, err := s.PullPaths(ctx, dst.Base, toPull)
	if err != nil {
		return pkg, err
	}
	dst.SetFromLayers(layersFetched)

	pkg, err = dst.ReadZarfYAML(warnings)
	if err != nil {
		return pkg, err
	}

	if err := dst.MigrateLegacy(); err != nil {
		return pkg, err
	}

	if !dst.IsLegacyLayout() {
		if wantSBOM {
			spinner := message.NewProgressSpinner("Validating SBOM checksums")
			defer spinner.Stop()

			if err := ValidatePackageIntegrity(dst, pkg.Metadata.AggregateChecksum, true); err != nil {
				return pkg, err
			}

			spinner.Success()
		}

		if err := ValidatePackageSignature(dst, s.PublicKeyPath); err != nil {
			if errors.Is(err, ErrPkgSigButNoKey) && skipValidation {
				message.Warn("The package was signed but no public key was provided, skipping signature validation")
			} else {
				return pkg, err
			}
		}
	}

	// unpack sboms.tar
	if wantSBOM {
		if err := dst.SBOMs.Unarchive(); err != nil {
			return pkg, err
		}
	}

	return pkg, nil
}

// Collect pulls a package from an OCI registry and writes it to a tarball.
func (s *OCISource) Collect(dir string) (string, error) {
	tmp, err := utils.MakeTempDir(config.CommonOptions.TempDirectory)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)
	ctx := context.TODO()
	fetched, err := s.PullPackage(ctx, tmp, config.CommonOptions.OCIConcurrency)
	if err != nil {
		return "", err
	}

	loaded := layout.New(tmp)
	loaded.SetFromLayers(fetched)

	var pkg types.ZarfPackage

	if err := utils.ReadYaml(loaded.ZarfYAML, &pkg); err != nil {
		return "", err
	}

	spinner := message.NewProgressSpinner("Validating full package checksums")
	defer spinner.Stop()

	if err := ValidatePackageIntegrity(loaded, pkg.Metadata.AggregateChecksum, false); err != nil {
		return "", err
	}

	spinner.Success()

	// TODO (@Noxsios) remove the suffix check at v1.0.0
	isSkeleton := pkg.Build.Architecture == zoci.SkeletonArch || strings.HasSuffix(s.Repo().Reference.Reference, zoci.SkeletonArch)
	name := fmt.Sprintf("%s%s", NameFromMetadata(&pkg, isSkeleton), PkgSuffix(pkg.Metadata.Uncompressed))

	dstTarball := filepath.Join(dir, name)

	allTheLayers, err := filepath.Glob(filepath.Join(tmp, "*"))
	if err != nil {
		return "", err
	}

	_ = os.Remove(dstTarball)

	return dstTarball, archiver.Archive(allTheLayers, dstTarball)
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package sources contains core implementations of the PackageSource interface.
package sources

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Racer159/jackal/src/config"
	"github.com/Racer159/jackal/src/pkg/layout"
	"github.com/Racer159/jackal/src/pkg/message"
	"github.com/Racer159/jackal/src/pkg/packager/filters"
	"github.com/Racer159/jackal/src/pkg/utils"
	"github.com/Racer159/jackal/src/pkg/zoci"
	"github.com/Racer159/jackal/src/types"
	"github.com/mholt/archiver/v3"
)

var (
	// verify that OCISource implements PackageSource
	_ PackageSource = (*OCISource)(nil)
)

// OCISource is a package source for OCI registries.
type OCISource struct {
	*types.JackalPackageOptions
	*zoci.Remote
}

// LoadPackage loads a package from an OCI registry.
func (s *OCISource) LoadPackage(dst *layout.PackagePaths, filter filters.ComponentFilterStrategy, unarchiveAll bool) (pkg types.JackalPackage, warnings []string, err error) {
	ctx := context.TODO()

	message.Debugf("Loading package from %q", s.PackageSource)

	pkg, err = s.FetchJackalYAML(ctx)
	if err != nil {
		return pkg, nil, err
	}
	pkg.Components, err = filter.Apply(pkg)
	if err != nil {
		return pkg, nil, err
	}

	layersToPull, err := s.LayersFromRequestedComponents(ctx, pkg.Components)
	if err != nil {
		return pkg, nil, fmt.Errorf("unable to get published component image layers: %s", err.Error())
	}

	isPartial := true
	root, err := s.FetchRoot(ctx)
	if err != nil {
		return pkg, nil, err
	}
	if len(root.Layers) == len(layersToPull) {
		isPartial = false
	}

	layersFetched, err := s.PullPackage(ctx, dst.Base, config.CommonOptions.OCIConcurrency, layersToPull...)
	if err != nil {
		return pkg, nil, fmt.Errorf("unable to pull the package: %w", err)
	}
	dst.SetFromLayers(layersFetched)

	if err := dst.MigrateLegacy(); err != nil {
		return pkg, nil, err
	}

	if !dst.IsLegacyLayout() {
		spinner := message.NewProgressSpinner("Validating pulled layer checksums")
		defer spinner.Stop()

		if err := ValidatePackageIntegrity(dst, pkg.Metadata.AggregateChecksum, isPartial); err != nil {
			return pkg, nil, err
		}

		spinner.Success()

		if err := ValidatePackageSignature(dst, s.PublicKeyPath); err != nil {
			return pkg, nil, err
		}
	}

	if unarchiveAll {
		for _, component := range pkg.Components {
			if err := dst.Components.Unarchive(component); err != nil {
				if layout.IsNotLoaded(err) {
					_, err := dst.Components.Create(component)
					if err != nil {
						return pkg, nil, err
					}
				} else {
					return pkg, nil, err
				}
			}
		}

		if dst.SBOMs.Path != "" {
			if err := dst.SBOMs.Unarchive(); err != nil {
				return pkg, nil, err
			}
		}
	}

	return pkg, warnings, nil
}

// LoadPackageMetadata loads a package's metadata from an OCI registry.
func (s *OCISource) LoadPackageMetadata(dst *layout.PackagePaths, wantSBOM bool, skipValidation bool) (pkg types.JackalPackage, warnings []string, err error) {
	toPull := zoci.PackageAlwaysPull
	if wantSBOM {
		toPull = append(toPull, layout.SBOMTar)
	}
	ctx := context.TODO()
	layersFetched, err := s.PullPaths(ctx, dst.Base, toPull)
	if err != nil {
		return pkg, nil, err
	}
	dst.SetFromLayers(layersFetched)

	pkg, warnings, err = dst.ReadJackalYAML()
	if err != nil {
		return pkg, nil, err
	}

	if err := dst.MigrateLegacy(); err != nil {
		return pkg, nil, err
	}

	if !dst.IsLegacyLayout() {
		if wantSBOM {
			spinner := message.NewProgressSpinner("Validating SBOM checksums")
			defer spinner.Stop()

			if err := ValidatePackageIntegrity(dst, pkg.Metadata.AggregateChecksum, true); err != nil {
				return pkg, nil, err
			}

			spinner.Success()
		}

		if err := ValidatePackageSignature(dst, s.PublicKeyPath); err != nil {
			if errors.Is(err, ErrPkgSigButNoKey) && skipValidation {
				message.Warn("The package was signed but no public key was provided, skipping signature validation")
			} else {
				return pkg, nil, err
			}
		}
	}

	// unpack sboms.tar
	if wantSBOM {
		if err := dst.SBOMs.Unarchive(); err != nil {
			return pkg, nil, err
		}
	}

	return pkg, warnings, nil
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

	var pkg types.JackalPackage

	if err := utils.ReadYaml(loaded.JackalYAML, &pkg); err != nil {
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

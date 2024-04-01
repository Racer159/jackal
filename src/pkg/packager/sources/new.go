// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package sources contains core implementations of the PackageSource interface.
package sources

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Racer159/jackal/src/config"
	"github.com/Racer159/jackal/src/pkg/layout"
	"github.com/Racer159/jackal/src/pkg/message"
	"github.com/Racer159/jackal/src/pkg/packager/filters"
	"github.com/Racer159/jackal/src/pkg/zoci"
	"github.com/Racer159/jackal/src/types"
	"github.com/defenseunicorns/pkg/helpers"
	"github.com/defenseunicorns/pkg/oci"
)

// PackageSource is an interface for package sources.
//
// While this interface defines three functions, LoadPackage, LoadPackageMetadata, and Collect; only one of them should be used within a packager function.
//
// These functions currently do not promise repeatability due to the side effect nature of loading a package.
//
// Signature and integrity validation is up to the implementation of the package source.
//
//	`sources.ValidatePackageSignature` and `sources.ValidatePackageIntegrity` can be leveraged for this purpose.
type PackageSource interface {
	// LoadPackage loads a package from a source.
	LoadPackage(dst *layout.PackagePaths, filter filters.ComponentFilterStrategy, unarchiveAll bool) (pkg types.JackalPackage, warnings []string, err error)

	// LoadPackageMetadata loads a package's metadata from a source.
	LoadPackageMetadata(dst *layout.PackagePaths, wantSBOM bool, skipValidation bool) (pkg types.JackalPackage, warnings []string, err error)

	// Collect relocates a package from its source to a tarball in a given destination directory.
	Collect(destinationDirectory string) (tarball string, err error)
}

// Identify returns the type of package source based on the provided package source string.
func Identify(pkgSrc string) string {
	if helpers.IsURL(pkgSrc) {
		parsed, _ := url.Parse(pkgSrc)
		return parsed.Scheme
	}

	if strings.Contains(pkgSrc, ".part000") {
		return "split"
	}

	if IsValidFileExtension(pkgSrc) {
		return "tarball"
	}

	return ""
}

// New returns a new PackageSource based on the provided package options.
func New(pkgOpts *types.JackalPackageOptions) (PackageSource, error) {
	var source PackageSource

	pkgSrc := pkgOpts.PackageSource

	switch Identify(pkgSrc) {
	case "oci":
		if pkgOpts.Shasum != "" {
			pkgSrc = fmt.Sprintf("%s@sha256:%s", pkgSrc, pkgOpts.Shasum)
		}
		arch := config.GetArch()
		remote, err := zoci.NewRemote(pkgSrc, oci.PlatformForArch(arch))
		if err != nil {
			return nil, err
		}
		source = &OCISource{JackalPackageOptions: pkgOpts, Remote: remote}
	case "tarball":
		source = &TarballSource{pkgOpts}
	case "http", "https", "sget":
		source = &URLSource{pkgOpts}
	case "split":
		source = &SplitTarballSource{pkgOpts}
	default:
		return nil, fmt.Errorf("could not identify source type for %q", pkgSrc)
	}

	message.Debugf("Using %T for %q", source, pkgSrc)

	return source, nil
}

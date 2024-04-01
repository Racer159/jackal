// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package sources contains core implementations of the PackageSource interface.
package sources

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/defenseunicorns/jackal/src/config"
	"github.com/defenseunicorns/jackal/src/pkg/layout"
	"github.com/defenseunicorns/jackal/src/pkg/zoci"
	"github.com/defenseunicorns/jackal/src/types"
	"github.com/defenseunicorns/pkg/helpers"
	goyaml "github.com/goccy/go-yaml"
	"github.com/mholt/archiver/v3"
)

// GetValidPackageExtensions returns the valid package extensions.
func GetValidPackageExtensions() [2]string {
	return [...]string{".tar.zst", ".tar"}
}

// IsValidFileExtension returns true if the filename has a valid package extension.
func IsValidFileExtension(filename string) bool {
	for _, extension := range GetValidPackageExtensions() {
		if strings.HasSuffix(filename, extension) {
			return true
		}
	}

	return false
}

func identifyUnknownTarball(path string) (string, error) {
	if helpers.InvalidPath(path) {
		return "", &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}
	if filepath.Ext(path) != "" && IsValidFileExtension(path) {
		return path, nil
	} else if filepath.Ext(path) != "" && !IsValidFileExtension(path) {
		return "", fmt.Errorf("%s is not a supported tarball format (%+v)", path, GetValidPackageExtensions())
	}

	// rename to .tar.zst and check if it's a valid tar.zst
	tzst := fmt.Sprintf("%s.tar.zst", path)
	if err := os.Rename(path, tzst); err != nil {
		return "", err
	}
	format, err := archiver.ByExtension(tzst)
	if err != nil {
		return "", err
	}
	_, ok := format.(*archiver.TarZstd)
	if ok {
		return tzst, nil
	}

	// rename to .tar and check if it's a valid tar
	tb := fmt.Sprintf("%s.tar", path)
	if err := os.Rename(tzst, tb); err != nil {
		return "", err
	}
	format, err = archiver.ByExtension(tb)
	if err != nil {
		return "", err
	}
	_, ok = format.(*archiver.Tar)
	if ok {
		return tb, nil
	}

	return "", fmt.Errorf("%s is not a supported tarball format (%+v)", path, GetValidPackageExtensions())
}

// RenameFromMetadata renames a tarball based on its metadata.
func RenameFromMetadata(path string) (string, error) {
	var pkg types.JackalPackage

	ext := filepath.Ext(path)
	if ext == "" {
		pathWithExt, err := identifyUnknownTarball(path)
		if err != nil {
			return "", err
		}
		path = pathWithExt
		ext = filepath.Ext(path)
	}
	if ext == ".zst" {
		ext = ".tar.zst"
	}

	if err := archiver.Walk(path, func(f archiver.File) error {
		if f.Name() == layout.JackalYAML {
			b, err := io.ReadAll(f)
			if err != nil {
				return err
			}
			if err := goyaml.Unmarshal(b, &pkg); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return "", err
	}

	if pkg.Metadata.Name == "" {
		return "", fmt.Errorf("%q does not contain a jackal.yaml", path)
	}

	name := NameFromMetadata(&pkg, false)

	name = fmt.Sprintf("%s%s", name, ext)

	tb := filepath.Join(filepath.Dir(path), name)

	return tb, os.Rename(path, tb)
}

// NameFromMetadata generates a name from a package's metadata.
func NameFromMetadata(pkg *types.JackalPackage, isSkeleton bool) string {
	var name string

	arch := config.GetArch(pkg.Metadata.Architecture, pkg.Build.Architecture)

	if isSkeleton {
		arch = zoci.SkeletonArch
	}

	switch pkg.Kind {
	case types.JackalInitConfig:
		name = fmt.Sprintf("jackal-init-%s", arch)
	case types.JackalPackageConfig:
		name = fmt.Sprintf("jackal-package-%s-%s", pkg.Metadata.Name, arch)
	default:
		name = fmt.Sprintf("jackal-%s-%s", strings.ToLower(string(pkg.Kind)), arch)
	}

	if pkg.Build.Differential {
		name = fmt.Sprintf("%s-%s-differential-%s", name, pkg.Build.DifferentialPackageVersion, pkg.Metadata.Version)
	} else if pkg.Metadata.Version != "" {
		name = fmt.Sprintf("%s-%s", name, pkg.Metadata.Version)
	}

	return name
}

// GetInitPackageName returns the formatted name of the init package.
func GetInitPackageName() string {
	// No package has been loaded yet so lookup GetArch() with no package info
	arch := config.GetArch()
	return fmt.Sprintf("jackal-init-%s-%s.tar.zst", arch, config.CLIVersion)
}

// PkgSuffix returns a package suffix based on whether it is uncompressed or not.
func PkgSuffix(uncompressed bool) (suffix string) {
	suffix = ".tar.zst"
	if uncompressed {
		suffix = ".tar"
	}
	return suffix
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package layout contains functions for interacting with Zarf's package layout on disk.
package layout

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/pkg/utils/helpers"
	"github.com/defenseunicorns/zarf/src/types"
	"github.com/google/go-containerregistry/pkg/crane"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// PackagePaths is the default package layout.
type PackagePaths struct {
	Base      string
	ZarfYAML  string
	Checksums string

	Signature string

	Components Components
	SBOMs      SBOMs
	Images     Images

	isLegacyLayout bool
}

// InjectionMadnessPaths contains paths for injection madness.
type InjectionMadnessPaths struct {
	InjectionBinary      string
	SeedImagesDir        string
	InjectorPayloadTarGz string
}

// New returns a new PackagePaths struct.
func New(baseDir string) *PackagePaths {
	return &PackagePaths{
		Base:      baseDir,
		ZarfYAML:  filepath.Join(baseDir, ZarfYAML),
		Checksums: filepath.Join(baseDir, Checksums),
		Components: Components{
			Base: filepath.Join(baseDir, ComponentsDir),
		},
	}
}

// MigrateLegacy migrates a legacy package layout to the new layout.
func (pp *PackagePaths) MigrateLegacy() (err error) {
	var pkg types.ZarfPackage
	base := pp.Base

	// legacy layout does not contain a checksums file, nor a signature
	if helpers.InvalidPath(pp.Checksums) && pp.Signature == "" {
		if err := utils.ReadYaml(pp.ZarfYAML, &pkg); err != nil {
			return err
		}
		buildVer, err := semver.NewVersion(pkg.Build.Version)
		if err != nil {
			return err
		}
		if !buildVer.LessThan(semver.MustParse("v0.25.0")) {
			return nil
		}
		pp.isLegacyLayout = true
	} else {
		return nil
	}

	// Migrate legacy sboms
	legacySBOMs := filepath.Join(base, "sboms")
	if !helpers.InvalidPath(legacySBOMs) {
		pp = pp.AddSBOMs()
		message.Debugf("Migrating %q to %q", legacySBOMs, pp.SBOMs.Path)
		if err := os.Rename(legacySBOMs, pp.SBOMs.Path); err != nil {
			return err
		}
	}

	// Migrate legacy images
	legacyImagesTar := filepath.Join(base, "images.tar")
	if !helpers.InvalidPath(legacyImagesTar) {
		pp = pp.AddImages()
		message.Debugf("Migrating %q to %q", legacyImagesTar, pp.Images.Base)
		defer os.Remove(legacyImagesTar)
		imgTags := []string{}
		for _, component := range pkg.Components {
			imgTags = append(imgTags, component.Images...)
		}
		// convert images to oci layout
		// until this for-loop is complete, there will be a duplication of images, resulting in some wasted space
		tagToDigest := make(map[string]string)
		for _, tag := range imgTags {
			img, err := crane.LoadTag(legacyImagesTar, tag)
			if err != nil {
				return err
			}
			if err := crane.SaveOCI(img, pp.Images.Base); err != nil {
				return err
			}
			// Get the image digest so we can set an annotation in the image.json later
			imgDigest, err := img.Digest()
			if err != nil {
				return err
			}
			tagToDigest[tag] = imgDigest.String()

			if err := pp.Images.AddV1Image(img); err != nil {
				return err
			}
		}
		if err := utils.AddImageNameAnnotation(pp.Images.Base, tagToDigest); err != nil {
			return err
		}
	}

	// Migrate legacy components
	//
	// Migration of paths within components occurs during `deploy`
	// no other operation should need to know about legacy component paths
	for _, component := range pkg.Components {
		_, err := pp.Components.Create(component)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsLegacyLayout returns true if the package is using the legacy layout.
func (pp *PackagePaths) IsLegacyLayout() bool {
	return pp.isLegacyLayout
}

// AddSignature sets the signature path if the keyPath is not empty.
func (pp *PackagePaths) AddSignature(keyPath string) *PackagePaths {
	if keyPath != "" {
		pp.Signature = filepath.Join(pp.Base, Signature)
	}
	return pp
}

// AddImages sets the default image paths.
func (pp *PackagePaths) AddImages() *PackagePaths {
	pp.Images.Base = filepath.Join(pp.Base, ImagesDir)
	pp.Images.OCILayout = filepath.Join(pp.Images.Base, OCILayout)
	pp.Images.Index = filepath.Join(pp.Images.Base, IndexJSON)
	return pp
}

// AddSBOMs sets the default sbom paths.
func (pp *PackagePaths) AddSBOMs() *PackagePaths {
	pp.SBOMs = SBOMs{
		Path: filepath.Join(pp.Base, SBOMDir),
	}
	return pp
}

// SetFromLayers maps layers to package paths.
func (pp *PackagePaths) SetFromLayers(layers []ocispec.Descriptor) {
	paths := []string{}
	for _, layer := range layers {
		if layer.Annotations[ocispec.AnnotationTitle] != "" {
			paths = append(paths, layer.Annotations[ocispec.AnnotationTitle])
		}
	}
	pp.SetFromPaths(paths)
}

// SetFromPaths maps paths to package paths.
func (pp *PackagePaths) SetFromPaths(paths []string) {
	for _, rel := range paths {
		// Convert from the standard '/' to the OS path separator for Windows support
		switch path := filepath.FromSlash(rel); {
		case path == ZarfYAML:
			pp.ZarfYAML = filepath.Join(pp.Base, path)
		case path == Signature:
			pp.Signature = filepath.Join(pp.Base, path)
		case path == Checksums:
			pp.Checksums = filepath.Join(pp.Base, path)
		case path == SBOMTar:
			pp.SBOMs.Path = filepath.Join(pp.Base, path)
		case path == filepath.Join(ImagesDir, OCILayout):
			pp.Images.OCILayout = filepath.Join(pp.Base, path)
		case path == filepath.Join(ImagesDir, IndexJSON):
			pp.Images.Index = filepath.Join(pp.Base, path)
		case strings.HasPrefix(path, filepath.Join(ImagesDir, "blobs", "sha256")):
			if pp.Images.Base == "" {
				pp.Images.Base = filepath.Join(pp.Base, ImagesDir)
			}
			pp.Images.AddBlob(filepath.Base(path))
		case strings.HasPrefix(path, ComponentsDir) && filepath.Ext(path) == ".tar":
			if pp.Components.Base == "" {
				pp.Components.Base = filepath.Join(pp.Base, ComponentsDir)
			}
			componentName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			if pp.Components.Tarballs == nil {
				pp.Components.Tarballs = make(map[string]string)
			}
			pp.Components.Tarballs[componentName] = filepath.Join(pp.Base, path)
		default:
			message.Debug("ignoring path", path)
		}
	}
}

// Files returns a map of all the files in the package.
func (pp *PackagePaths) Files() map[string]string {
	pathMap := make(map[string]string)

	stripBase := func(path string) string {
		rel, _ := filepath.Rel(pp.Base, path)
		// Convert from the OS path separator to the standard '/' for Windows support
		return filepath.ToSlash(rel)
	}

	add := func(path string) {
		if path == "" {
			return
		}
		pathMap[stripBase(path)] = path
	}

	add(pp.ZarfYAML)
	add(pp.Signature)
	add(pp.Checksums)

	add(pp.Images.OCILayout)
	add(pp.Images.Index)
	for _, blob := range pp.Images.Blobs {
		add(blob)
	}

	for _, tarball := range pp.Components.Tarballs {
		add(tarball)
	}

	if filepath.Ext(pp.SBOMs.Path) == ".tar" {
		add(pp.SBOMs.Path)
	}
	return pathMap
}

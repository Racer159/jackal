// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package images provides functions for building and pushing images.
package images

import (
	"fmt"
	"os"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/transform"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/types"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// ImageConfig is the main struct for managing container images.
type ImageConfig struct {
	ImagesPath string

	ImageList []transform.Image

	RegInfo types.RegistryInfo

	NoChecksum bool

	Insecure bool

	Architectures []string

	RegistryOverrides map[string]string
}

// GetLegacyImgTarballPath returns the ImagesPath as if it were a path to a tarball instead of a directory.
func (i *ImageConfig) GetLegacyImgTarballPath() string {
	return fmt.Sprintf("%s.tar", i.ImagesPath)
}

// LoadImageFromPackage returns a v1.Image from the specified image, or an error if the image cannot be found.
func (i ImageConfig) LoadImageFromPackage(refInfo transform.Image) (v1.Image, error) {
	// If the package still has a images.tar that contains all of the images, use crane to load the specific reference (crane tag) we want
	if _, statErr := os.Stat(i.GetLegacyImgTarballPath()); statErr == nil {
		return crane.LoadTag(i.GetLegacyImgTarballPath(), refInfo.Reference, config.GetCraneOptions(i.Insecure, i.Architectures...)...)
	}

	// Load the image from the OCI formatted images directory
	return utils.LoadOCIImage(i.ImagesPath, refInfo)
}

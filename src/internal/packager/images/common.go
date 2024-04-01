// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package images provides functions for building and pushing images.
package images

import (
	"github.com/defenseunicorns/jackal/src/pkg/transform"
	"github.com/defenseunicorns/jackal/src/types"
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

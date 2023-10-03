// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package layout contains functions for interacting with Zarf's package layout on disk.
package layout

import (
	"path/filepath"

	"slices"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// Images contains paths for images.
type Images struct {
	Base      string
	Index     string
	OCILayout string
	Blobs     []string
}

// AddBlob adds a blob to the Images struct.
func (i *Images) AddBlob(blob string) {
	if len(blob) != 64 {
		return
	}
	abs := filepath.Join(i.Base, "blobs", "sha256", blob)
	if !slices.Contains(i.Blobs, abs) {
		i.Blobs = append(i.Blobs, abs)
	}
}

// AddV1Image adds a v1.Image to the Images struct.
func (i *Images) AddV1Image(img v1.Image) error {
	layers, err := img.Layers()
	if err != nil {
		return err
	}
	for _, layer := range layers {
		digest, err := layer.Digest()
		if err != nil {
			return err
		}
		i.AddBlob(digest.Hex)
	}
	imgCfgSha, err := img.ConfigName()
	if err != nil {
		return err
	}
	i.AddBlob(imgCfgSha.Hex)
	manifestSha, err := img.Digest()
	if err != nil {
		return err
	}
	i.AddBlob(manifestSha.Hex)

	return nil
}

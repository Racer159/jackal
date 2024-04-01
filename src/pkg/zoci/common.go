// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package zoci contains functions for interacting with Jackal packages stored in OCI registries.
package zoci

import (
	"log/slog"

	"github.com/defenseunicorns/pkg/oci"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/racer159/jackal/src/config"
	"github.com/racer159/jackal/src/pkg/message"
)

const (
	// JackalConfigMediaType is the media type for the manifest config
	JackalConfigMediaType = "application/vnd.jackal.config.v1+json"
	// JackalLayerMediaTypeBlob is the media type for all Jackal layers due to the range of possible content
	JackalLayerMediaTypeBlob = "application/vnd.jackal.layer.v1.blob"
	// SkeletonArch is the architecture used for skeleton packages
	SkeletonArch = "skeleton"
)

// Remote is a wrapper around the Oras remote repository with jackal specific functions
type Remote struct {
	*oci.OrasRemote
}

// NewRemote returns an oras remote repository client and context for the given url
// with jackal opination embedded
func NewRemote(url string, platform ocispec.Platform, mods ...oci.Modifier) (*Remote, error) {
	logger := slog.New(message.JackalHandler{})
	modifiers := append([]oci.Modifier{
		oci.WithPlainHTTP(config.CommonOptions.Insecure),
		oci.WithInsecureSkipVerify(config.CommonOptions.Insecure),
		oci.WithLogger(logger),
		oci.WithUserAgent("jackal/" + config.CLIVersion),
	}, mods...)
	remote, err := oci.NewOrasRemote(url, platform, modifiers...)
	if err != nil {
		return nil, err
	}
	return &Remote{remote}, nil
}

// PlatformForSkeleton sets the target architecture for the remote to skeleton
func PlatformForSkeleton() ocispec.Platform {
	return ocispec.Platform{
		OS:           oci.MultiOS,
		Architecture: SkeletonArch,
	}
}

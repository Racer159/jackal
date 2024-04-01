// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package zoci contains functions for interacting with Jackal packages stored in OCI registries.
package zoci

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Racer159/jackal/src/types"
	"github.com/defenseunicorns/pkg/helpers"
	"oras.land/oras-go/v2/registry"
)

// ReferenceFromMetadata returns a reference for the given metadata.
func ReferenceFromMetadata(registryLocation string, metadata *types.JackalMetadata, build *types.JackalBuildData) (string, error) {
	ver := metadata.Version
	if len(ver) == 0 {
		return "", errors.New("version is required for publishing")
	}

	if !strings.HasSuffix(registryLocation, "/") {
		registryLocation = registryLocation + "/"
	}
	registryLocation = strings.TrimPrefix(registryLocation, helpers.OCIURLPrefix)

	format := "%s%s:%s"
	raw := fmt.Sprintf(format, registryLocation, metadata.Name, ver)

	if build != nil && build.Flavor != "" {
		raw = fmt.Sprintf("%s-%s", raw, build.Flavor)
	}

	ref, err := registry.ParseReference(raw)
	if err != nil {
		return "", fmt.Errorf("failed to parse %s: %w", raw, err)
	}

	return ref.String(), nil
}

// GetInitPackageURL returns the URL for the init package for the given version.
func GetInitPackageURL(version string) string {
	return fmt.Sprintf("ghcr.io/defenseunicorns/packages/init:%s", version)
}

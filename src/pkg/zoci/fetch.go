// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package zoci contains functions for interacting with Jackal packages stored in OCI registries.
package zoci

import (
	"context"

	"github.com/defenseunicorns/pkg/oci"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/racer159/jackal/src/pkg/layout"
	"github.com/racer159/jackal/src/types"
)

// FetchJackalYAML fetches the jackal.yaml file from the remote repository.
func (r *Remote) FetchJackalYAML(ctx context.Context) (pkg types.JackalPackage, err error) {
	manifest, err := r.FetchRoot(ctx)
	if err != nil {
		return pkg, err
	}
	return oci.FetchYAMLFile[types.JackalPackage](ctx, r.FetchLayer, manifest, layout.JackalYAML)
}

// FetchImagesIndex fetches the images/index.json file from the remote repository.
func (r *Remote) FetchImagesIndex(ctx context.Context) (index *ocispec.Index, err error) {
	manifest, err := r.FetchRoot(ctx)
	if err != nil {
		return nil, err
	}
	return oci.FetchJSONFile[*ocispec.Index](ctx, r.FetchLayer, manifest, layout.IndexPath)
}

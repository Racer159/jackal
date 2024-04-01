// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package sources contains core implementations of the PackageSource interface.
package sources

import (
	"fmt"

	"github.com/defenseunicorns/jackal/src/internal/packager/validate"
	"github.com/defenseunicorns/jackal/src/pkg/cluster"
	"github.com/defenseunicorns/jackal/src/pkg/layout"
	"github.com/defenseunicorns/jackal/src/pkg/packager/filters"
	"github.com/defenseunicorns/jackal/src/pkg/utils"
	"github.com/defenseunicorns/jackal/src/types"
	"github.com/defenseunicorns/pkg/helpers"
)

var (
	// verify that ClusterSource implements PackageSource
	_ PackageSource = (*ClusterSource)(nil)
)

// NewClusterSource creates a new cluster source.
func NewClusterSource(pkgOpts *types.JackalPackageOptions) (PackageSource, error) {
	if !validate.IsLowercaseNumberHyphenNoStartHyphen(pkgOpts.PackageSource) {
		return nil, fmt.Errorf("invalid package name %q", pkgOpts.PackageSource)
	}
	cluster, err := cluster.NewClusterWithWait(cluster.DefaultTimeout)
	if err != nil {
		return nil, err
	}
	return &ClusterSource{pkgOpts, cluster}, nil
}

// ClusterSource is a package source for clusters.
type ClusterSource struct {
	*types.JackalPackageOptions
	*cluster.Cluster
}

// LoadPackage loads a package from a cluster.
//
// This is not implemented.
func (s *ClusterSource) LoadPackage(_ *layout.PackagePaths, _ filters.ComponentFilterStrategy, _ bool) (types.JackalPackage, []string, error) {
	return types.JackalPackage{}, nil, fmt.Errorf("not implemented")
}

// Collect collects a package from a cluster.
//
// This is not implemented.
func (s *ClusterSource) Collect(_ string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// LoadPackageMetadata loads package metadata from a cluster.
func (s *ClusterSource) LoadPackageMetadata(dst *layout.PackagePaths, _ bool, _ bool) (types.JackalPackage, []string, error) {
	dpkg, err := s.GetDeployedPackage(s.PackageSource)
	if err != nil {
		return types.JackalPackage{}, nil, err
	}

	if err := utils.WriteYaml(dst.JackalYAML, dpkg.Data, helpers.ReadUser); err != nil {
		return types.JackalPackage{}, nil, err
	}

	return dpkg.Data, nil, nil
}

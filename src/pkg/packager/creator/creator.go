// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package creator contains functions for creating Jackal packages.
package creator

import (
	"github.com/racer159/jackal/src/pkg/layout"
	"github.com/racer159/jackal/src/types"
)

// Creator is an interface for creating Jackal packages.
type Creator interface {
	LoadPackageDefinition(dst *layout.PackagePaths) (pkg types.JackalPackage, warnings []string, err error)
	Assemble(dst *layout.PackagePaths, components []types.JackalComponent, arch string) error
	Output(dst *layout.PackagePaths, pkg *types.JackalPackage) error
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package creator contains functions for creating Jackal packages.
package creator

import (
	"github.com/racer159/jackal/src/pkg/message"
	"github.com/racer159/jackal/src/pkg/packager/composer"
	"github.com/racer159/jackal/src/types"
)

// ComposeComponents composes components and their dependencies into a single Jackal package using an import chain.
func ComposeComponents(pkg types.JackalPackage, flavor string) (types.JackalPackage, []string, error) {
	components := []types.JackalComponent{}
	warnings := []string{}

	pkgVars := pkg.Variables
	pkgConsts := pkg.Constants

	arch := pkg.Metadata.Architecture

	for i, component := range pkg.Components {
		// filter by architecture and flavor
		if !composer.CompatibleComponent(component, arch, flavor) {
			continue
		}

		// if a match was found, strip flavor and architecture to reduce bloat in the package definition
		component.Only.Cluster.Architecture = ""
		component.Only.Flavor = ""

		// build the import chain
		chain, err := composer.NewImportChain(component, i, pkg.Metadata.Name, arch, flavor)
		if err != nil {
			return types.JackalPackage{}, nil, err
		}
		message.Debugf("%s", chain)

		// migrate any deprecated component configurations now
		warning := chain.Migrate(pkg.Build)
		warnings = append(warnings, warning...)

		// get the composed component
		composed, err := chain.Compose()
		if err != nil {
			return types.JackalPackage{}, nil, err
		}
		components = append(components, *composed)

		// merge variables and constants
		pkgVars = chain.MergeVariables(pkgVars)
		pkgConsts = chain.MergeConstants(pkgConsts)
	}

	// set the filtered + composed components
	pkg.Components = components

	pkg.Variables = pkgVars
	pkg.Constants = pkgConsts

	return pkg, warnings, nil
}

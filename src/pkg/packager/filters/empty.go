// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package filters contains core implementations of the ComponentFilterStrategy interface.
package filters

import "github.com/racer159/jackal/src/types"

// Empty returns a filter that does nothing.
func Empty() ComponentFilterStrategy {
	return &emptyFilter{}
}

// emptyFilter is a filter that does nothing.
type emptyFilter struct{}

// Apply returns the components unchanged.
func (f *emptyFilter) Apply(pkg types.JackalPackage) ([]types.JackalComponent, error) {
	return pkg.Components, nil
}

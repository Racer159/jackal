// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package filters contains core implementations of the ComponentFilterStrategy interface.
package filters

import (
	"errors"

	"github.com/Racer159/jackal/src/types"
)

// ByLocalOS creates a new filter that filters components based on local (runtime) OS.
func ByLocalOS(localOS string) ComponentFilterStrategy {
	return &localOSFilter{localOS}
}

// localOSFilter filters components based on local (runtime) OS.
type localOSFilter struct {
	localOS string
}

// ErrLocalOSRequired is returned when localOS is not set.
var ErrLocalOSRequired = errors.New("localOS is required")

// Apply applies the filter.
func (f *localOSFilter) Apply(pkg types.JackalPackage) ([]types.JackalComponent, error) {
	if f.localOS == "" {
		return nil, ErrLocalOSRequired
	}

	filtered := []types.JackalComponent{}
	for _, component := range pkg.Components {
		if component.Only.LocalOS == "" || component.Only.LocalOS == f.localOS {
			filtered = append(filtered, component)
		}
	}
	return filtered, nil
}

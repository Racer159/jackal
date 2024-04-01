// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package filters contains core implementations of the ComponentFilterStrategy interface.
package filters

import (
	"path"
	"strings"

	"github.com/racer159/jackal/src/types"
)

type selectState int

const (
	unknown selectState = iota
	included
	excluded
)

func includedOrExcluded(componentName string, requestedComponentNames []string) (selectState, string) {
	// Check if the component has a leading dash indicating it should be excluded - this is done first so that exclusions precede inclusions
	for _, requestedComponent := range requestedComponentNames {
		if strings.HasPrefix(requestedComponent, "-") {
			// If the component glob matches one of the requested components, then return true
			// This supports globbing with "path" in order to have the same behavior across OSes (if we ever allow namespaced components with /)
			if matched, _ := path.Match(strings.TrimPrefix(requestedComponent, "-"), componentName); matched {
				return excluded, requestedComponent
			}
		}
	}
	// Check if the component matches a glob pattern and should be included
	for _, requestedComponent := range requestedComponentNames {
		// If the component glob matches one of the requested components, then return true
		// This supports globbing with "path" in order to have the same behavior across OSes (if we ever allow namespaced components with /)
		if matched, _ := path.Match(requestedComponent, componentName); matched {
			return included, requestedComponent
		}
	}

	// All other cases we don't know if we should include or exclude yet
	return unknown, ""
}

// isRequired returns if the component is required or not.
func isRequired(c types.JackalComponent) bool {
	requiredExists := c.Required != nil
	required := requiredExists && *c.Required

	if requiredExists {
		return required
	}
	return false
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package template provides functions for templating yaml files.
package template

import (
	"regexp"

	"github.com/Racer159/jackal/src/types"
	"github.com/defenseunicorns/pkg/helpers"
)

// ProcessYamlFilesInPath iterates over all yaml files in a given path and performs Jackal templating + image swapping.
func ProcessYamlFilesInPath(path string, component types.JackalComponent, values Values) ([]string, error) {
	// Only pull in yml and yaml files
	pattern := regexp.MustCompile(`(?mi)\.ya?ml$`)
	manifests, _ := helpers.RecursiveFileList(path, pattern, false)

	for _, manifest := range manifests {
		if err := values.Apply(component, manifest, false); err != nil {
			return nil, err
		}
	}

	return manifests, nil
}

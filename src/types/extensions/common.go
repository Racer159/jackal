// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package extensions contains the types for all official extensions.
package extensions

// JackalComponentExtensions is a struct that contains all the official extensions
type JackalComponentExtensions struct {
	// Big Bang Configurations
	BigBang *BigBang `json:"bigbang,omitempty" jsonschema:"description=Configurations for installing Big Bang and Flux in the cluster"`
}

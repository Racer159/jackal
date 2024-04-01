// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package types contains all the types used by Jackal.
package types

import (
	"fmt"
	"regexp"
)

// PackagerConfig is the main struct that the packager uses to hold high-level options.
type PackagerConfig struct {
	// CreateOpts tracks the user-defined options used to create the package
	CreateOpts JackalCreateOptions

	// PkgOpts tracks user-defined options
	PkgOpts JackalPackageOptions

	// DeployOpts tracks user-defined values for the active deployment
	DeployOpts JackalDeployOptions

	// MirrorOpts tracks user-defined values for the active mirror
	MirrorOpts JackalMirrorOptions

	// InitOpts tracks user-defined values for the active Jackal initialization.
	InitOpts JackalInitOptions

	// InspectOpts tracks user-defined options used to inspect the package
	InspectOpts JackalInspectOptions

	// PublishOpts tracks user-defined options used to publish the package
	PublishOpts JackalPublishOptions

	// PullOpts tracks user-defined options used to pull packages
	PullOpts JackalPullOptions

	// FindImagesOpts tracks user-defined options used to find images
	FindImagesOpts JackalFindImagesOptions

	// GenerateOpts tracks user-defined values for package generation.
	GenerateOpts JackalGenerateOptions

	// The package data
	Pkg JackalPackage

	// The active jackal state
	State *JackalState

	// Variables set by the user
	SetVariableMap map[string]*JackalSetVariable
}

// SetVariable sets a value for a variable in PackagerConfig.SetVariableMap.
func (cfg *PackagerConfig) SetVariable(name, value string, sensitive bool, autoIndent bool, varType VariableType) {
	cfg.SetVariableMap[name] = &JackalSetVariable{
		Name:       name,
		Value:      value,
		Sensitive:  sensitive,
		AutoIndent: autoIndent,
		Type:       varType,
	}
}

// CheckVariablePattern checks to see if a variable is set to a value that matches its pattern.
func (cfg *PackagerConfig) CheckVariablePattern(name, pattern string) error {
	if regexp.MustCompile(pattern).MatchString(cfg.SetVariableMap[name].Value) {
		return nil
	}
	return fmt.Errorf("provided value for variable %q does not match pattern \"%s\"", name, pattern)
}

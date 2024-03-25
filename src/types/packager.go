// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package types contains all the types used by Zarf.
package types

import (
	"fmt"
	"regexp"
)

// PackagerConfig is the main struct that the packager uses to hold high-level options.
type PackagerConfig struct {
	// CreateOpts tracks the user-defined options used to create the package
	CreateOpts ZarfCreateOptions

	// PkgOpts tracks user-defined options
	PkgOpts ZarfPackageOptions

	// DeployOpts tracks user-defined values for the active deployment
	DeployOpts ZarfDeployOptions

	// MirrorOpts tracks user-defined values for the active mirror
	MirrorOpts ZarfMirrorOptions

	// InitOpts tracks user-defined values for the active Zarf initialization.
	InitOpts ZarfInitOptions

	// InspectOpts tracks user-defined options used to inspect the package
	InspectOpts ZarfInspectOptions

	// PublishOpts tracks user-defined options used to publish the package
	PublishOpts ZarfPublishOptions

	// PullOpts tracks user-defined options used to pull packages
	PullOpts ZarfPullOptions

	// FindImagesOpts tracks user-defined options used to find images
	FindImagesOpts ZarfFindImagesOptions

	// GenerateOpts tracks user-defined values for package generation.
	GenerateOpts ZarfGenerateOptions

	// The package data
	Pkg ZarfPackage

	// The active zarf state
	State *ZarfState

	// Variables set by the user
	SetVariableMap map[string]*ZarfSetVariable
}

// SetVariable sets a value for a variable in PackagerConfig.SetVariableMap.
func (cfg *PackagerConfig) SetVariable(name, value string, sensitive bool, autoIndent bool, varType VariableType) {
	cfg.SetVariableMap[name] = &ZarfSetVariable{
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

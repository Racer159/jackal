// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package types contains all the types used by Zarf.
package types

// PackagerConfig is the main struct that the packager uses to hold high-level options.
type PackagerConfig struct {
	// CreateOpts tracks the user-defined options used to create the package
	CreateOpts ZarfCreateOptions

	// PkgOpts tracks user-defined options
	PkgOpts ZarfPackageOptions

	// DeployOpts tracks user-defined values for the active deployment
	DeployOpts ZarfDeployOptions

	// InitOpts tracks user-defined values for the active Zarf initialization.
	InitOpts ZarfInitOptions

	// PublishOpts tracks user-defined options used to publish the package
	PublishOpts ZarfPublishOptions

	// PullOpts tracks user-defined options used to pull packages
	PullOpts ZarfPullOptions

	// Track if the package is an init package
	IsInitConfig bool

	// The package data
	Pkg ZarfPackage

	// The original source of the package
	PkgSource string

	// The active zarf state
	State *ZarfState

	// Variables set by the user
	SetVariableMap map[string]*ZarfSetVariable

	// SBOM file paths in the package
	SBOMViewFiles []string
}

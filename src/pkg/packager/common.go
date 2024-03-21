// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package packager contains functions for interacting with, managing and deploying Zarf packages.
package packager

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"slices"

	"github.com/Masterminds/semver/v3"
	"github.com/defenseunicorns/zarf/src/config/lang"
	"github.com/defenseunicorns/zarf/src/internal/packager/template"
	"github.com/defenseunicorns/zarf/src/pkg/cluster"
	"github.com/defenseunicorns/zarf/src/types"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/layout"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/packager/deprecated"
	"github.com/defenseunicorns/zarf/src/pkg/packager/sources"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
)

// Packager is the main struct for managing packages.
type Packager struct {
	cfg            *types.PackagerConfig
	cluster        *cluster.Cluster
	layout         *layout.PackagePaths
	warnings       *message.Warnings
	valueTemplate  *template.Values
	hpaModified    bool
	connectStrings types.ConnectStrings
	sbomViewFiles  []string
	source         sources.PackageSource
	generation     int
}

// Modifier is a function that modifies the packager.
type Modifier func(*Packager)

// WithSource sets the source for the packager.
func WithSource(source sources.PackageSource) Modifier {
	return func(p *Packager) {
		p.source = source
	}
}

// WithCluster sets the cluster client for the packager.
func WithCluster(cluster *cluster.Cluster) Modifier {
	return func(p *Packager) {
		p.cluster = cluster
	}
}

// WithTemp sets the temp directory for the packager.
//
// This temp directory is used as the destination where p.source loads the package.
func WithTemp(base string) Modifier {
	return func(p *Packager) {
		p.layout = layout.New(base)
	}
}

/*
New creates a new package instance with the provided config.

Note: This function creates a tmp directory that should be cleaned up with p.ClearTempPaths().
*/
func New(cfg *types.PackagerConfig, mods ...Modifier) (*Packager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("no config provided")
	}

	if cfg.SetVariableMap == nil {
		cfg.SetVariableMap = make(map[string]*types.ZarfSetVariable)
	}

	var (
		err  error
		pkgr = &Packager{
			cfg: cfg,
		}
	)

	if config.CommonOptions.TempDirectory != "" {
		// If the cache directory is within the temp directory, warn the user
		if strings.HasPrefix(config.CommonOptions.CachePath, config.CommonOptions.TempDirectory) {
			message.Warnf("The cache directory (%q) is within the temp directory (%q) and will be removed when the temp directory is cleaned up", config.CommonOptions.CachePath, config.CommonOptions.TempDirectory)
		}
	}

	for _, mod := range mods {
		mod(pkgr)
	}

	// Fill the source if it wasn't provided - note source can be nil if the package is being created
	if pkgr.source == nil && pkgr.cfg.CreateOpts.BaseDir == "" {
		pkgr.source, err = sources.New(&pkgr.cfg.PkgOpts)
		if err != nil {
			return nil, err
		}
	}

	// If the temp directory is not set, set it to the default
	if pkgr.layout == nil {
		if err = pkgr.setTempDirectory(config.CommonOptions.TempDirectory); err != nil {
			return nil, fmt.Errorf("unable to create package temp paths: %w", err)
		}
	}

	if pkgr.warnings == nil {
		pkgr.warnings = message.NewWarnings()
	}

	return pkgr, nil
}

/*
NewOrDie creates a new package instance with the provided config or throws a fatal error.

Note: This function creates a tmp directory that should be cleaned up with p.ClearTempPaths().
*/
func NewOrDie(config *types.PackagerConfig, mods ...Modifier) *Packager {
	var (
		err  error
		pkgr *Packager
	)

	if pkgr, err = New(config, mods...); err != nil {
		message.Fatalf(err, "Unable to setup the package config: %s", err.Error())
	}

	return pkgr
}

// setTempDirectory sets the temp directory for the packager.
func (p *Packager) setTempDirectory(path string) error {
	dir, err := utils.MakeTempDir(path)
	if err != nil {
		return fmt.Errorf("unable to create package temp paths: %w", err)
	}

	p.layout = layout.New(dir)
	return nil
}

// ClearTempPaths removes the temp directory and any files within it.
func (p *Packager) ClearTempPaths() {
	// Remove the temp directory, but don't throw an error if it fails
	_ = os.RemoveAll(p.layout.Base)
	_ = os.RemoveAll(layout.SBOMDir)
}

// connectToCluster attempts to connect to a cluster if a connection is not already established
func (p *Packager) connectToCluster(timeout time.Duration) (err error) {
	if p.isConnectedToCluster() {
		return nil
	}

	p.cluster, err = cluster.NewClusterWithWait(timeout)
	if err != nil {
		return err
	}

	return p.attemptClusterChecks()
}

// isConnectedToCluster returns whether the current packager instance is connected to a cluster
func (p *Packager) isConnectedToCluster() bool {
	return p.cluster != nil
}

// hasImages returns whether the current package contains images
func (p *Packager) hasImages() bool {
	for _, component := range p.cfg.Pkg.Components {
		if len(component.Images) > 0 {
			return true
		}
	}
	return false
}

// attemptClusterChecks attempts to connect to the cluster and check for useful metadata and config mismatches.
// NOTE: attemptClusterChecks should only return an error if there is a problem significant enough to halt a deployment, otherwise it should return nil and print a warning message.
func (p *Packager) attemptClusterChecks() (err error) {

	spinner := message.NewProgressSpinner("Gathering additional cluster information (if available)")
	defer spinner.Stop()

	// Check if the package has already been deployed and get its generation
	if existingDeployedPackage, _ := p.cluster.GetDeployedPackage(p.cfg.Pkg.Metadata.Name); existingDeployedPackage != nil {
		// If this package has been deployed before, increment the package generation within the secret
		p.generation = existingDeployedPackage.Generation + 1
	}

	// Check the clusters architecture matches the package spec
	if err := p.validatePackageArchitecture(); err != nil {
		if errors.Is(err, lang.ErrUnableToCheckArch) {
			message.Warnf("Unable to validate package architecture: %s", err.Error())
		} else {
			return err
		}
	}

	// Check for any breaking changes between the initialized Zarf version and this CLI
	if existingInitPackage, _ := p.cluster.GetDeployedPackage("init"); existingInitPackage != nil {
		// Use the build version instead of the metadata since this will support older Zarf versions
		deprecated.PrintBreakingChanges(existingInitPackage.Data.Build.Version)
	}

	spinner.Success()

	return nil
}

// validatePackageArchitecture validates that the package architecture matches the target cluster architecture.
func (p *Packager) validatePackageArchitecture() error {
	// Ignore this check if we don't have a cluster connection, or the package contains no images
	if !p.isConnectedToCluster() || !p.hasImages() {
		return nil
	}

	clusterArchitectures, err := p.cluster.GetArchitectures()
	if err != nil {
		return lang.ErrUnableToCheckArch
	}

	// Check if the package architecture and the cluster architecture are the same.
	if !slices.Contains(clusterArchitectures, p.cfg.Pkg.Metadata.Architecture) {
		return fmt.Errorf(lang.CmdPackageDeployValidateArchitectureErr, p.cfg.Pkg.Metadata.Architecture, strings.Join(clusterArchitectures, ", "))
	}

	return nil
}

// validateLastNonBreakingVersion validates the Zarf CLI version against a package's LastNonBreakingVersion.
func (p *Packager) validateLastNonBreakingVersion() (err error) {
	cliVersion := config.CLIVersion
	lastNonBreakingVersion := p.cfg.Pkg.Build.LastNonBreakingVersion

	if lastNonBreakingVersion == "" {
		return nil
	}

	lastNonBreakingSemVer, err := semver.NewVersion(lastNonBreakingVersion)
	if err != nil {
		return fmt.Errorf("unable to parse lastNonBreakingVersion '%s' from Zarf package build data : %w", lastNonBreakingVersion, err)
	}

	cliSemVer, err := semver.NewVersion(cliVersion)
	if err != nil {
		p.warnings.Add(fmt.Sprintf(lang.CmdPackageDeployInvalidCLIVersionWarn, config.CLIVersion))
		return nil
	}

	if cliSemVer.LessThan(lastNonBreakingSemVer) {
		p.warnings.Add(
			fmt.Sprintf(
				lang.CmdPackageDeployValidateLastNonBreakingVersionWarn,
				cliVersion,
				lastNonBreakingVersion,
				lastNonBreakingVersion,
			),
		)
	}

	return nil
}

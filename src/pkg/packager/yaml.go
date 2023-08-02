// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package packager contains functions for interacting with, managing and deploying Zarf packages.
package packager

import (
	"os"
	"runtime"
	"time"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/packager/deprecated"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/types"
)

// readYaml loads the config from the given path
func (p *Packager) readYaml(path string) error {
	if err := utils.ReadYaml(path, &p.cfg.Pkg); err != nil {
		return err
	}

	// Set the arch from the package config before filtering.
	p.arch = config.GetArch(p.cfg.Pkg.Metadata.Architecture, p.cfg.Pkg.Build.Architecture)

	return nil
}

// filterComponents removes components not matching the current OS if filterByOS is set.
func (p *Packager) filterComponents(filterByOS bool) {
	// Filter each component to only compatible platforms.
	filteredComponents := []types.ZarfComponent{}
	for _, component := range p.cfg.Pkg.Components {
		if p.isCompatibleComponent(component, filterByOS) {
			filteredComponents = append(filteredComponents, component)
		}
	}
	// Update the active package with the filtered components.
	p.cfg.Pkg.Components = filteredComponents
}

// writeYaml adds build information and writes the config to the temp directory.
func (p *Packager) writeYaml() error {
	message.Debug("config.BuildConfig()")

	now := time.Now()
	// Just use $USER env variable to avoid CGO issue.
	// https://groups.google.com/g/golang-dev/c/ZFDDX3ZiJ84.
	// Record the name of the user creating the package.
	if runtime.GOOS == "windows" {
		p.cfg.Pkg.Build.User = os.Getenv("USERNAME")
	} else {
		p.cfg.Pkg.Build.User = os.Getenv("USER")
	}
	hostname, hostErr := os.Hostname()

	// Normalize these for the package confirmation.
	p.cfg.Pkg.Metadata.Architecture = p.arch
	p.cfg.Pkg.Build.Architecture = p.arch

	// Record the time of package creation.
	p.cfg.Pkg.Build.Timestamp = now.Format(time.RFC1123Z)

	// Record the Zarf Version the CLI was built with.
	p.cfg.Pkg.Build.Version = config.CLIVersion

	if hostErr == nil {
		// Record the hostname of the package creation terminal.
		p.cfg.Pkg.Build.Terminal = hostname
	}

	// Record the migrations that will be run on the package.
	p.cfg.Pkg.Build.Migrations = []string{
		deprecated.ScriptsToActionsMigrated,
		deprecated.PluralizeSetVariable,
	}

	p.cfg.Pkg.Build.RegistryOverrides = p.cfg.CreateOpts.RegistryOverrides

	// Record the latest version of Zarf without breaking changes to the package structure.
	p.cfg.Pkg.Build.LastNonBreakingVersion = deprecated.LastNonBreakingVersion

	return utils.WriteYaml(p.tmp.ZarfYaml, p.cfg.Pkg, 0400)
}

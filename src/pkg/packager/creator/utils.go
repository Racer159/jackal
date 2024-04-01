// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package creator contains functions for creating Jackal packages.
package creator

import (
	"os"
	"runtime"
	"time"

	"github.com/Racer159/jackal/src/config"
	"github.com/Racer159/jackal/src/pkg/packager/deprecated"
	"github.com/Racer159/jackal/src/types"
)

// recordPackageMetadata records various package metadata during package create.
func recordPackageMetadata(pkg *types.JackalPackage, createOpts types.JackalCreateOptions) error {
	now := time.Now()
	// Just use $USER env variable to avoid CGO issue.
	// https://groups.google.com/g/golang-dev/c/ZFDDX3ZiJ84.
	// Record the name of the user creating the package.
	if runtime.GOOS == "windows" {
		pkg.Build.User = os.Getenv("USERNAME")
	} else {
		pkg.Build.User = os.Getenv("USER")
	}

	// Record the hostname of the package creation terminal.
	// The error here is ignored because the hostname is not critical to the package creation.
	hostname, _ := os.Hostname()
	pkg.Build.Terminal = hostname

	if pkg.IsInitConfig() {
		pkg.Metadata.Version = config.CLIVersion
	}

	pkg.Build.Architecture = pkg.Metadata.Architecture

	// Record the Jackal Version the CLI was built with.
	pkg.Build.Version = config.CLIVersion

	// Record the time of package creation.
	pkg.Build.Timestamp = now.Format(time.RFC1123Z)

	// Record the migrations that will be ran on the package.
	pkg.Build.Migrations = []string{
		deprecated.ScriptsToActionsMigrated,
		deprecated.PluralizeSetVariable,
	}

	// Record the flavor of Jackal used to build this package (if any).
	pkg.Build.Flavor = createOpts.Flavor

	pkg.Build.RegistryOverrides = createOpts.RegistryOverrides

	// Record the latest version of Jackal without breaking changes to the package structure.
	pkg.Build.LastNonBreakingVersion = deprecated.LastNonBreakingVersion

	return nil
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package deprecated handles package deprecations and migrations
package deprecated

import (
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils/helpers"
	"github.com/defenseunicorns/zarf/src/types"
	"github.com/pterm/pterm"
)

// BreakingChange represents a breaking change that happened on a specified Zarf version
type BreakingChange struct {
	version    *semver.Version
	title      string
	mitigation string
}

// List of migrations tracked in the zarf.yaml build data.
const (
	ScriptsToActionsMigrated = "scripts-to-actions"
	PluralizeSetVariable     = "pluralize-set-variable"
)

// List of breaking changes to warn the user of.
var breakingChanges = []BreakingChange{
	{
		version:    semver.New(0, 26, 0, "", ""),
		title:      "Zarf container images are now mutated based on tag instead of repository name.",
		mitigation: "Reinitialize the cluster using v0.26.0 or later and redeploy existing packages to update the image references (you can view existing packages with 'zarf package list' and view cluster images with 'zarf tools registry catalog').",
	},
}

// MigrateComponent runs all migrations on a component.
// Build should be empty on package create, but include just in case someone copied a zarf.yaml from a zarf package.
func MigrateComponent(build types.ZarfBuildData, component types.ZarfComponent) (migratedComponent types.ZarfComponent, warnings []string) {
	migratedComponent = component

	// If the component has already been migrated, clear the deprecated scripts.
	if helpers.SliceContains(build.Migrations, ScriptsToActionsMigrated) {
		migratedComponent.DeprecatedScripts = types.DeprecatedZarfComponentScripts{}
	} else {
		// Otherwise, run the migration.
		var warning string
		if migratedComponent, warning = migrateScriptsToActions(migratedComponent); warning != "" {
			warnings = append(warnings, warning)
		}
	}

	// If the component has already been migrated, clear the setVariable definitions.
	if helpers.SliceContains(build.Migrations, PluralizeSetVariable) {
		migratedComponent = clearSetVariables(migratedComponent)
	} else {
		// Otherwise, run the migration.
		var warning string
		if migratedComponent, warning = migrateSetVariableToSetVariables(migratedComponent); warning != "" {
			warnings = append(warnings, warning)
		}
	}

	// Future migrations here.
	return migratedComponent, warnings
}

// PrintBreakingChanges prints the breaking changes between the provided version and the current CLIVersion
func PrintBreakingChanges(deployedZarfVersion string) {
	deployedSemver, err := semver.NewVersion(deployedZarfVersion)
	if err != nil {
		message.HorizontalRule()
		pterm.Println()
		message.Warnf("Unable to determine init-package version from %s.  There is potential for breaking changes.", deployedZarfVersion)
		return
	}

	applicableBreakingChanges := []BreakingChange{}

	// Calculate the applicable breaking changes
	for _, breakingChange := range breakingChanges {
		if deployedSemver.LessThan(breakingChange.version) {
			applicableBreakingChanges = append(applicableBreakingChanges, breakingChange)
		}
	}

	if len(applicableBreakingChanges) > 0 {
		// Print header information
		message.HorizontalRule()
		message.Title("Potential Breaking Changes", "breaking changes that may cause issues with this package")

		// Print information about the versions
		format := pterm.FgYellow.Sprint("CLI version ") + "%s" + pterm.FgYellow.Sprint(" is being used to deploy to a cluster that was initialized with ") +
			"%s" + pterm.FgYellow.Sprint(". Between these versions there are the following breaking changes to consider:")
		cliVersion := pterm.Bold.Sprintf(config.CLIVersion)
		deployedVersion := pterm.Bold.Sprintf(deployedZarfVersion)
		message.Warnf(format, cliVersion, deployedVersion)

		// Print each applicable breaking change
		for idx, applicableBreakingChange := range applicableBreakingChanges {
			titleFormat := pterm.Bold.Sprintf("\n %d. ", idx+1) + "%s"

			pterm.Printfln(titleFormat, applicableBreakingChange.title)

			mitigationText := message.Paragraphn(96, "%s", pterm.FgLightCyan.Sprint(applicableBreakingChange.mitigation))

			pterm.Printfln("\n  - %s", pterm.Bold.Sprint("Mitigation:"))
			pterm.Printfln("    %s", strings.ReplaceAll(mitigationText, "\n", "\n    "))
		}
	}
}

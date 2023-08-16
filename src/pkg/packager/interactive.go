// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package packager contains functions for interacting with, managing and deploying Zarf packages.
package packager

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/internal/packager/sbom"
	"github.com/defenseunicorns/zarf/src/pkg/interactive"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/packager/deprecated"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/types"
	"github.com/pterm/pterm"
)

func (p *Packager) confirmAction(stage string, sbomViewFiles []string) (confirm bool) {

	pterm.Println()
	message.HeaderInfof("📦 PACKAGE DEFINITION")
	utils.ColorPrintYAML(p.cfg.Pkg, p.getPackageYAMLHints(stage), true)

	// Print any potential breaking changes (if this is a Deploy confirm) between this CLI version and the deployed init package
	if stage == config.ZarfDeployStage {
		if sbom.IsSBOMAble(p.cfg.Pkg) {
			// Print the location that the user can view the package SBOMs from
			message.HorizontalRule()
			message.Title("Software Bill of Materials", "an inventory of all software contained in this package")

			if len(sbomViewFiles) > 0 {
				cwd, _ := os.Getwd()
				link := pterm.FgLightCyan.Sprint(pterm.Bold.Sprint(filepath.Join(cwd, config.ZarfSBOMDir, filepath.Base(sbomViewFiles[0]))))
				inspect := pterm.BgBlack.Sprint(pterm.FgWhite.Sprint(pterm.Bold.Sprintf("$ zarf package inspect %s", p.cfg.PkgSource)))

				artifactMsg := pterm.Bold.Sprintf("%d artifacts", len(sbomViewFiles)) + " to be reviewed. These are"
				if len(sbomViewFiles) == 1 {
					artifactMsg = pterm.Bold.Sprintf("%d artifact", len(sbomViewFiles)) + " to be reviewed. This is"
				}

				msg := fmt.Sprintf("This package has %s available in a temporary '%s' folder in this directory and will be removed upon deployment.\n", artifactMsg, pterm.Bold.Sprint("zarf-sbom"))
				viewNow := fmt.Sprintf("\n- View SBOMs %s by navigating to the '%s' folder or copying this link into a browser:\n%s", pterm.Bold.Sprint("now"), pterm.Bold.Sprint("zarf-sbom"), link)
				viewLater := fmt.Sprintf("\n- View SBOMs %s deployment with this command:\n%s", pterm.Bold.Sprint("after"), inspect)

				message.Note(msg)
				pterm.Println(viewNow)
				pterm.Println(viewLater)
			} else {
				message.Warn("This package does NOT contain an SBOM.  If you require an SBOM, please contact the creator of this package to request a version that includes an SBOM.")
			}
		}

		// Connect to the cluster (if available) to check the Zarf Agent for breaking changes
		if p.cluster != nil {
			if initPackage, err := p.cluster.GetDeployedPackage("init"); err == nil {
				// We use the build.version for now because it is the most reliable way to get this version info pre v0.26.0
				deprecated.PrintBreakingChanges(initPackage.Data.Build.Version)
			}
		}
	}

	if len(p.warnings) > 0 {
		message.HorizontalRule()
		message.Title("Package Warnings", "the following warnings were flagged while reading the package")
		for _, warning := range p.warnings {
			message.Warn(warning)
		}
	}

	message.HorizontalRule()

	// Display prompt if not auto-confirmed
	if config.CommonOptions.Confirm {
		pterm.Println()
		message.Successf("%s Zarf package confirmed", stage)
		return config.CommonOptions.Confirm
	}

	prompt := &survey.Confirm{
		Message: stage + " this Zarf package?",
	}

	pterm.Println()

	// Prompt the user for confirmation, on abort return false
	if err := survey.AskOne(prompt, &confirm); err != nil || !confirm {
		// User aborted or declined, cancel the action
		return false
	}

	// On create in interactive mode, prompt for max package size if it is still the default value of 0
	// Note: it will not be 0 if the user has provided a value via the --max-package-size flag or Viper config
	if stage == config.ZarfCreateStage && p.cfg.CreateOpts.MaxPackageSizeMB == 0 {
		value, err := interactive.PromptVariable(types.ZarfPackageVariable{
			Name:        "Maximum Package Size",
			Description: "Specify a maximum file size for this package in Megabytes. Above this size, the package will be split into multiple files. 0 will disable this feature.",
			Default:     "0",
		})
		if err != nil {
			// User aborted, cancel the action
			return false
		}

		// Try to parse the value, on error warn and move on
		maxPackageSize, err := strconv.Atoi(value)
		if err != nil {
			message.Warnf("Unable to parse \"%s\" as a number for the maximum file size. This package will not be split into multiple files.", value)
			return true
		}

		p.cfg.CreateOpts.MaxPackageSizeMB = maxPackageSize
	}

	return true
}

func (p *Packager) getPackageYAMLHints(stage string) map[string]string {
	hints := map[string]string{}

	if stage == config.ZarfDeployStage {
		for _, variable := range p.cfg.Pkg.Variables {
			value, present := p.cfg.PkgOpts.SetVariables[variable.Name]
			if !present {
				value = fmt.Sprintf("'%s' (default)", message.Truncate(variable.Default, 20, false))
			} else {
				value = fmt.Sprintf("'%s'", message.Truncate(value, 20, false))
			}
			if variable.Sensitive {
				value = "'**sanitized**'"
			}
			hints = utils.AddRootListHint(hints, "name", variable.Name, fmt.Sprintf("currently set to %s", value))
		}
	}

	hints = utils.AddRootHint(hints, "metadata", "information about this package\n")
	hints = utils.AddRootHint(hints, "build", "info about the machine, zarf version, and user that created this package\n")
	hints = utils.AddRootHint(hints, "components", "definition of capabilities this package deploys")
	hints = utils.AddRootHint(hints, "constants", "static values set by the package author")
	hints = utils.AddRootHint(hints, "variables", "deployment-specific values that are set on each package deployment")

	return hints
}

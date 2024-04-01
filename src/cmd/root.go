// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package cmd contains the CLI commands for Jackal.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/defenseunicorns/jackal/src/cmd/common"
	"github.com/defenseunicorns/jackal/src/cmd/tools"
	"github.com/defenseunicorns/jackal/src/config"
	"github.com/defenseunicorns/jackal/src/config/lang"
	"github.com/defenseunicorns/jackal/src/pkg/layout"
	"github.com/defenseunicorns/jackal/src/pkg/message"
	"github.com/defenseunicorns/jackal/src/types"
	"github.com/spf13/cobra"
)

var (
	// Default global config for the packager
	pkgConfig = types.PackagerConfig{}
)

var rootCmd = &cobra.Command{
	Use: "jackal COMMAND",
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		// Skip for vendor-only commands
		if common.CheckVendorOnlyFromPath(cmd) {
			return
		}

		// Don't log the help command
		if cmd.Parent() == nil {
			config.SkipLogFile = true
		}

		common.SetupCLI()
	},
	Short: lang.RootCmdShort,
	Long:  lang.RootCmdLong,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jackalLogo := message.GetLogo()
		_, _ = fmt.Fprintln(os.Stderr, jackalLogo)
		cmd.Help()

		if len(args) > 0 {
			if strings.Contains(args[0], config.JackalPackagePrefix) || strings.Contains(args[0], "jackal-init") {
				message.Warnf(lang.RootCmdDeprecatedDeploy, args[0])
			}
			if args[0] == layout.JackalYAML {
				message.Warn(lang.RootCmdDeprecatedCreate)
			}
		}
	},
}

// Execute is the entrypoint for the CLI.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// Add the tools commands
	tools.Include(rootCmd)

	// Skip for vendor-only commands
	if common.CheckVendorOnlyFromArgs() {
		return
	}

	v := common.InitViper()

	rootCmd.PersistentFlags().StringVarP(&common.LogLevelCLI, "log-level", "l", v.GetString(common.VLogLevel), lang.RootCmdFlagLogLevel)
	rootCmd.PersistentFlags().StringVarP(&config.CLIArch, "architecture", "a", v.GetString(common.VArchitecture), lang.RootCmdFlagArch)
	rootCmd.PersistentFlags().BoolVar(&config.SkipLogFile, "no-log-file", v.GetBool(common.VNoLogFile), lang.RootCmdFlagSkipLogFile)
	rootCmd.PersistentFlags().BoolVar(&message.NoProgress, "no-progress", v.GetBool(common.VNoProgress), lang.RootCmdFlagNoProgress)
	rootCmd.PersistentFlags().BoolVar(&config.NoColor, "no-color", v.GetBool(common.VNoColor), lang.RootCmdFlagNoColor)
	rootCmd.PersistentFlags().StringVar(&config.CommonOptions.CachePath, "jackal-cache", v.GetString(common.VJackalCache), lang.RootCmdFlagCachePath)
	rootCmd.PersistentFlags().StringVar(&config.CommonOptions.TempDirectory, "tmpdir", v.GetString(common.VTmpDir), lang.RootCmdFlagTempDir)
	rootCmd.PersistentFlags().BoolVar(&config.CommonOptions.Insecure, "insecure", v.GetBool(common.VInsecure), lang.RootCmdFlagInsecure)
}

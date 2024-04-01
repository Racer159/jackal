// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package tools contains the CLI commands for Jackal.
package tools

import (
	"fmt"

	"github.com/racer159/jackal/src/cmd/common"
	"github.com/racer159/jackal/src/config"
	"github.com/racer159/jackal/src/config/lang"
	"github.com/spf13/cobra"
)

var toolsCmd = &cobra.Command{
	Use:     "tools",
	Aliases: []string{"t"},
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		config.SkipLogFile = true

		// Skip for vendor-only commands
		if common.CheckVendorOnlyFromPath(cmd) {
			return
		}

		common.SetupCLI()
	},
	Short: lang.CmdToolsShort,
}

// Include adds the tools command to the root command.
func Include(rootCmd *cobra.Command) {
	rootCmd.AddCommand(toolsCmd)
}

// newVersionCmd is a generic version command for tools
func newVersionCmd(name, version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Args:  cobra.NoArgs,
		Short: lang.CmdToolsVersionShort,
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println(fmt.Sprintf("%s %s", name, version))
		},
	}
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package tools contains the CLI commands for Jackal.
package tools

import (
	"github.com/Racer159/jackal/src/config/lang"
	"github.com/anchore/clio"
	syftCLI "github.com/anchore/syft/cmd/syft/cli"
)

// ldflags github.com/Racer159/jackal/src/cmd/tools.syftVersion=x.x.x
var syftVersion string

func init() {
	syftCmd := syftCLI.Command(clio.Identification{
		Name:    "syft",
		Version: syftVersion,
	})
	syftCmd.Use = "sbom"
	syftCmd.Short = lang.CmdToolsSbomShort
	syftCmd.Aliases = []string{"s", "syft"}
	syftCmd.Example = ""

	for _, subCmd := range syftCmd.Commands() {
		subCmd.Example = ""
	}

	toolsCmd.AddCommand(syftCmd)
}

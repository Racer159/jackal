// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package tools contains the CLI commands for Jackal.
package tools

import (
	yq "github.com/mikefarah/yq/v4/cmd"
	"github.com/racer159/jackal/src/config/lang"
)

func init() {

	yqCmd := yq.New()
	yqCmd.Example = lang.CmdToolsYqExample
	yqCmd.Use = "yq"
	for _, subCmd := range yqCmd.Commands() {
		if subCmd.Name() == "eval" {
			subCmd.Example = lang.CmdToolsYqEvalExample
		}
		if subCmd.Name() == "eval-all" {
			subCmd.Example = lang.CmdToolsYqEvalAllExample
		}
	}

	toolsCmd.AddCommand(yqCmd)
}

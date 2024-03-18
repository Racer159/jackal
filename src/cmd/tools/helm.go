// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package tools contains the CLI commands for Zarf.
package tools

import (
	"os"

	"github.com/defenseunicorns/zarf/src/cmd/tools/helm"
	"github.com/defenseunicorns/zarf/src/config/lang"
	"helm.sh/helm/v3/pkg/action"
)

// ldflags github.com/defenseunicorns/zarf/src/cmd/tools.helmVersion=x.x.x
var helmVersion string

func init() {
	actionConfig := new(action.Configuration)

	// Truncate Helm's arguments so that it thinks its all alone
	helmArgs := []string{}
	if len(os.Args) > 2 {
		helmArgs = os.Args[3:]
	}
	// The inclusion of Helm in this manner should be changed once https://github.com/helm/helm/pull/12725 is merged
	helmCmd, _ := helm.NewRootCmd(actionConfig, os.Stdout, helmArgs)
	helmCmd.Short = lang.CmdToolsHelmShort
	helmCmd.Long = lang.CmdToolsHelmLong
	helmCmd.AddCommand(newVersionCmd("helm", helmVersion))

	toolsCmd.AddCommand(helmCmd)
}

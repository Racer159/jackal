// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package cmd contains the CLI commands for Jackal.
package cmd

import (
	"errors"
	"os"
	"regexp"

	"github.com/defenseunicorns/pkg/helpers"
	"github.com/racer159/jackal/src/config"
	"github.com/racer159/jackal/src/config/lang"
	"github.com/racer159/jackal/src/internal/packager/helm"
	"github.com/racer159/jackal/src/pkg/cluster"
	"github.com/racer159/jackal/src/pkg/message"
	"github.com/racer159/jackal/src/pkg/utils/exec"

	"github.com/spf13/cobra"
)

var confirmDestroy bool
var removeComponents bool

var destroyCmd = &cobra.Command{
	Use:     "destroy --confirm",
	Aliases: []string{"d"},
	Short:   lang.CmdDestroyShort,
	Long:    lang.CmdDestroyLong,
	Run: func(_ *cobra.Command, _ []string) {
		c, err := cluster.NewClusterWithWait(cluster.DefaultTimeout)
		if err != nil {
			message.Fatalf(err, lang.ErrNoClusterConnection)
		}

		// NOTE: If 'jackal init' failed to deploy the k3s component (or if we're looking at the wrong kubeconfig)
		//       there will be no jackal-state to load and the struct will be empty. In these cases, if we can find
		//       the scripts to remove k3s, we will still try to remove a locally installed k3s cluster
		state, err := c.LoadJackalState()
		if err != nil {
			message.WarnErr(err, lang.ErrLoadState)
		}

		// If Jackal deployed the cluster, burn it all down
		if state.JackalAppliance || (state.Distro == "") {
			// Check if we have the scripts to destroy everything
			fileInfo, err := os.Stat(config.JackalCleanupScriptsPath)
			if errors.Is(err, os.ErrNotExist) || !fileInfo.IsDir() {
				message.Fatalf(lang.CmdDestroyErrNoScriptPath, config.JackalCleanupScriptsPath)
			}

			// Run all the scripts!
			pattern := regexp.MustCompile(`(?mi)jackal-clean-.+\.sh$`)
			scripts, _ := helpers.RecursiveFileList(config.JackalCleanupScriptsPath, pattern, true)
			// Iterate over all matching jackal-clean scripts and exec them
			for _, script := range scripts {
				// Run the matched script
				err := exec.CmdWithPrint(script)
				if errors.Is(err, os.ErrPermission) {
					message.Warnf(lang.CmdDestroyErrScriptPermissionDenied, script)

					// Don't remove scripts we can't execute so the user can try to manually run
					continue
				} else if err != nil {
					message.Debugf("Received error when trying to execute the script (%s): %#v", script, err)
				}

				// Try to remove the script, but ignore any errors
				_ = os.Remove(script)
			}
		} else {
			// Perform chart uninstallation
			helm.Destroy(removeComponents)

			// If Jackal didn't deploy the cluster, only delete the JackalNamespace
			c.DeleteJackalNamespace()

			// Remove jackal agent labels and secrets from namespaces Jackal doesn't manage
			c.StripJackalLabelsAndSecretsFromNamespaces()
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	// Still going to require a flag for destroy confirm, no viper oopsies here
	destroyCmd.Flags().BoolVar(&confirmDestroy, "confirm", false, lang.CmdDestroyFlagConfirm)
	destroyCmd.Flags().BoolVar(&removeComponents, "remove-components", false, lang.CmdDestroyFlagRemoveComponents)
	_ = destroyCmd.MarkFlagRequired("confirm")
}

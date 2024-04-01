// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package cmd contains the CLI commands for Jackal contains the CLI commands for Jackal.
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Racer159/jackal/src/cmd/common"
	"github.com/Racer159/jackal/src/config/lang"
	"github.com/Racer159/jackal/src/pkg/cluster"
	"github.com/Racer159/jackal/src/pkg/k8s"
	"github.com/Racer159/jackal/src/pkg/message"
	"github.com/Racer159/jackal/src/pkg/utils/exec"
	"github.com/spf13/cobra"
)

var (
	connectResourceName string
	connectNamespace    string
	connectResourceType string
	connectLocalPort    int
	connectRemotePort   int
	cliOnly             bool

	connectCmd = &cobra.Command{
		Use:     "connect { REGISTRY | LOGGING | GIT | connect-name }",
		Aliases: []string{"c"},
		Short:   lang.CmdConnectShort,
		Long:    lang.CmdConnectLong,
		Run: func(_ *cobra.Command, args []string) {
			var target string
			if len(args) > 0 {
				target = args[0]
			}
			spinner := message.NewProgressSpinner(lang.CmdConnectPreparingTunnel, target)
			c, err := cluster.NewCluster()
			if err != nil {
				spinner.Fatalf(err, lang.CmdConnectErrCluster, err.Error())
			}

			var tunnel *k8s.Tunnel
			if connectResourceName != "" {
				zt := cluster.NewTunnelInfo(connectNamespace, connectResourceType, connectResourceName, "", connectLocalPort, connectRemotePort)
				tunnel, err = c.ConnectTunnelInfo(zt)
			} else {
				tunnel, err = c.Connect(target)
			}
			if err != nil {
				spinner.Fatalf(err, lang.CmdConnectErrService, err.Error())
			}

			defer tunnel.Close()
			url := tunnel.FullURL()

			// Dump the tunnel URL to the console for other tools to use.
			fmt.Print(url)

			if cliOnly {
				spinner.Updatef(lang.CmdConnectEstablishedCLI, url)
			} else {
				spinner.Updatef(lang.CmdConnectEstablishedWeb, url)

				if err := exec.LaunchURL(url); err != nil {
					message.Debug(err)
				}
			}

			// Keep this open until an interrupt signal is received.
			interruptChan := make(chan os.Signal, 1)
			signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)
			common.SuppressGlobalInterrupt = true

			// Wait for the interrupt signal or an error.
			select {
			case err = <-tunnel.ErrChan():
				spinner.Fatalf(err, lang.CmdConnectErrService, err.Error())
			case <-interruptChan:
				spinner.Successf(lang.CmdConnectTunnelClosed, url)
			}
			os.Exit(0)
		},
	}

	connectListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   lang.CmdConnectListShort,
		Run: func(_ *cobra.Command, _ []string) {
			cluster.NewClusterOrDie().PrintConnectTable()
		},
	}
)

func init() {
	rootCmd.AddCommand(connectCmd)
	connectCmd.AddCommand(connectListCmd)

	connectCmd.Flags().StringVar(&connectResourceName, "name", "", lang.CmdConnectFlagName)
	connectCmd.Flags().StringVar(&connectNamespace, "namespace", cluster.JackalNamespaceName, lang.CmdConnectFlagNamespace)
	connectCmd.Flags().StringVar(&connectResourceType, "type", k8s.SvcResource, lang.CmdConnectFlagType)
	connectCmd.Flags().IntVar(&connectLocalPort, "local-port", 0, lang.CmdConnectFlagLocalPort)
	connectCmd.Flags().IntVar(&connectRemotePort, "remote-port", 0, lang.CmdConnectFlagRemotePort)
	connectCmd.Flags().BoolVar(&cliOnly, "cli-only", false, lang.CmdConnectFlagCliOnly)
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package cmd contains the CLI commands for Jackal.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alecthomas/jsonschema"
	"github.com/defenseunicorns/pkg/helpers"
	"github.com/racer159/jackal/src/cmd/common"
	"github.com/racer159/jackal/src/config/lang"
	"github.com/racer159/jackal/src/internal/agent"
	"github.com/racer159/jackal/src/internal/packager/git"
	"github.com/racer159/jackal/src/pkg/cluster"
	"github.com/racer159/jackal/src/pkg/message"
	"github.com/racer159/jackal/src/types"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
)

var (
	rollback bool
)

var internalCmd = &cobra.Command{
	Use:    "internal",
	Hidden: true,
	Short:  lang.CmdInternalShort,
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: lang.CmdInternalAgentShort,
	Long:  lang.CmdInternalAgentLong,
	Run: func(_ *cobra.Command, _ []string) {
		agent.StartWebhook()
	},
}

var httpProxyCmd = &cobra.Command{
	Use:   "http-proxy",
	Short: lang.CmdInternalProxyShort,
	Long:  lang.CmdInternalProxyLong,
	Run: func(_ *cobra.Command, _ []string) {
		agent.StartHTTPProxy()
	},
}

var genCLIDocs = &cobra.Command{
	Use:   "gen-cli-docs",
	Short: lang.CmdInternalGenerateCliDocsShort,
	Run: func(_ *cobra.Command, _ []string) {
		// Don't include the datestamp in the output
		rootCmd.DisableAutoGenTag = true

		for _, cmd := range rootCmd.Commands() {
			if cmd.Use == "tools" {
				for _, toolCmd := range cmd.Commands() {
					// If the command is a vendored command, add a dummy flag to hide root flags from the docs
					if common.CheckVendorOnlyFromPath(toolCmd) {
						addHiddenDummyFlag(toolCmd, "log-level")
						addHiddenDummyFlag(toolCmd, "architecture")
						addHiddenDummyFlag(toolCmd, "no-log-file")
						addHiddenDummyFlag(toolCmd, "no-progress")
						addHiddenDummyFlag(toolCmd, "jackal-cache")
						addHiddenDummyFlag(toolCmd, "tmpdir")
						addHiddenDummyFlag(toolCmd, "insecure")
						addHiddenDummyFlag(toolCmd, "no-color")
					}

					// Remove the default values from all of the helm commands during the CLI command doc generation
					if toolCmd.Use == "helm" || toolCmd.Use == "sbom" {
						toolCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
							if flag.Value.Type() == "string" {
								flag.DefValue = ""
							}
						})
						for _, subCmd := range toolCmd.Commands() {
							subCmd.Flags().VisitAll(func(flag *pflag.Flag) {
								if flag.Value.Type() == "string" {
									flag.DefValue = ""
								}
							})
							for _, helmSubCmd := range subCmd.Commands() {
								helmSubCmd.Flags().VisitAll(func(flag *pflag.Flag) {
									if flag.Value.Type() == "string" {
										flag.DefValue = ""
									}
								})
							}
						}
					}

					if toolCmd.Use == "monitor" {
						toolCmd.Flags().VisitAll(func(flag *pflag.Flag) {
							if flag.Value.Type() == "string" {
								flag.DefValue = ""
							}
						})
					}

					if toolCmd.Use == "yq" {
						for _, subCmd := range toolCmd.Commands() {
							if subCmd.Name() == "shell-completion" {
								subCmd.Hidden = true
							}
						}
					}
				}
			}
		}

		//Generate markdown of the Jackal command (and all of its child commands)
		if err := os.RemoveAll("./docs/2-the-jackal-cli/100-cli-commands"); err != nil {
			message.Fatalf(lang.CmdInternalGenerateCliDocsErr, err.Error())
		}
		if err := os.Mkdir("./docs/2-the-jackal-cli/100-cli-commands", 0775); err != nil {
			message.Fatalf(lang.CmdInternalGenerateCliDocsErr, err.Error())
		}
		if err := doc.GenMarkdownTree(rootCmd, "./docs/2-the-jackal-cli/100-cli-commands"); err != nil {
			message.Fatalf(lang.CmdInternalGenerateCliDocsErr, err.Error())
		} else {
			message.Success(lang.CmdInternalGenerateCliDocsSuccess)
		}
	},
}

var genConfigSchemaCmd = &cobra.Command{
	Use:     "gen-config-schema",
	Aliases: []string{"gc"},
	Short:   lang.CmdInternalConfigSchemaShort,
	Run: func(_ *cobra.Command, _ []string) {
		schema := jsonschema.Reflect(&types.JackalPackage{})
		output, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			message.Fatal(err, lang.CmdInternalConfigSchemaErr)
		}
		fmt.Print(string(output) + "\n")
	},
}

type jackalTypes struct {
	DeployedPackage types.DeployedPackage
	JackalPackage   types.JackalPackage
	JackalState     types.JackalState
}

var genTypesSchemaCmd = &cobra.Command{
	Use:     "gen-types-schema",
	Aliases: []string{"gt"},
	Short:   lang.CmdInternalTypesSchemaShort,
	Run: func(_ *cobra.Command, _ []string) {
		schema := jsonschema.Reflect(&jackalTypes{})
		output, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			message.Fatal(err, lang.CmdInternalTypesSchemaErr)
		}
		fmt.Print(string(output) + "\n")
	},
}

var createReadOnlyGiteaUser = &cobra.Command{
	Use:   "create-read-only-gitea-user",
	Short: lang.CmdInternalCreateReadOnlyGiteaUserShort,
	Long:  lang.CmdInternalCreateReadOnlyGiteaUserLong,
	Run: func(_ *cobra.Command, _ []string) {
		// Load the state so we can get the credentials for the admin git user
		state, err := cluster.NewClusterOrDie().LoadJackalState()
		if err != nil {
			message.WarnErr(err, lang.ErrLoadState)
		}

		// Create the non-admin user
		if err = git.New(state.GitServer).CreateReadOnlyUser(); err != nil {
			message.WarnErr(err, lang.CmdInternalCreateReadOnlyGiteaUserErr)
		}
	},
}

var createPackageRegistryToken = &cobra.Command{
	Use:   "create-artifact-registry-token",
	Short: lang.CmdInternalArtifactRegistryGiteaTokenShort,
	Long:  lang.CmdInternalArtifactRegistryGiteaTokenLong,
	Run: func(_ *cobra.Command, _ []string) {
		// Load the state so we can get the credentials for the admin git user
		c := cluster.NewClusterOrDie()
		state, err := c.LoadJackalState()
		if err != nil {
			message.WarnErr(err, lang.ErrLoadState)
		}

		// If we are setup to use an internal artifact server, create the artifact registry token
		if state.ArtifactServer.InternalServer {
			token, err := git.New(state.GitServer).CreatePackageRegistryToken()
			if err != nil {
				message.WarnErr(err, lang.CmdInternalArtifactRegistryGiteaTokenErr)
			}

			state.ArtifactServer.PushToken = token.Sha1

			c.SaveJackalState(state)
		}
	},
}

var updateGiteaPVC = &cobra.Command{
	Use:   "update-gitea-pvc",
	Short: lang.CmdInternalUpdateGiteaPVCShort,
	Long:  lang.CmdInternalUpdateGiteaPVCLong,
	Run: func(_ *cobra.Command, _ []string) {

		// There is a possibility that the pvc does not yet exist and Gitea helm chart should create it
		helmShouldCreate, err := git.UpdateGiteaPVC(rollback)
		if err != nil {
			message.WarnErr(err, lang.CmdInternalUpdateGiteaPVCErr)
		}

		fmt.Print(helmShouldCreate)
	},
}

var isValidHostname = &cobra.Command{
	Use:   "is-valid-hostname",
	Short: lang.CmdInternalIsValidHostnameShort,
	Run: func(_ *cobra.Command, _ []string) {
		if valid := helpers.IsValidHostName(); !valid {
			hostname, _ := os.Hostname()
			message.Fatalf(nil, lang.CmdInternalIsValidHostnameErr, hostname)
		}
	},
}

var computeCrc32 = &cobra.Command{
	Use:     "crc32 TEXT",
	Aliases: []string{"c"},
	Short:   lang.CmdInternalCrc32Short,
	Args:    cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		text := args[0]
		hash := helpers.GetCRCHash(text)
		fmt.Printf("%d\n", hash)
	},
}

func init() {
	rootCmd.AddCommand(internalCmd)

	internalCmd.AddCommand(agentCmd)
	internalCmd.AddCommand(httpProxyCmd)
	internalCmd.AddCommand(genCLIDocs)
	internalCmd.AddCommand(genConfigSchemaCmd)
	internalCmd.AddCommand(genTypesSchemaCmd)
	internalCmd.AddCommand(createReadOnlyGiteaUser)
	internalCmd.AddCommand(createPackageRegistryToken)
	internalCmd.AddCommand(updateGiteaPVC)
	internalCmd.AddCommand(isValidHostname)
	internalCmd.AddCommand(computeCrc32)

	updateGiteaPVC.Flags().BoolVarP(&rollback, "rollback", "r", false, lang.CmdInternalFlagUpdateGiteaPVCRollback)
}

func addHiddenDummyFlag(cmd *cobra.Command, flagDummy string) {
	if cmd.PersistentFlags().Lookup(flagDummy) == nil {
		var dummyStr string
		cmd.PersistentFlags().StringVar(&dummyStr, flagDummy, "", "")
		cmd.PersistentFlags().MarkHidden(flagDummy)
	}
}

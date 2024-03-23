// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package cmd contains the CLI commands for Zarf.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/jsonschema"
	"github.com/defenseunicorns/zarf/src/cmd/common"
	"github.com/defenseunicorns/zarf/src/config/lang"
	"github.com/defenseunicorns/zarf/src/internal/agent"
	"github.com/defenseunicorns/zarf/src/internal/packager/git"
	"github.com/defenseunicorns/zarf/src/pkg/cluster"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils/helpers"
	"github.com/defenseunicorns/zarf/src/types"
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

		resetStringFlags := func(cmd *cobra.Command) {
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				if flag.Value.Type() == "string" {
					flag.DefValue = ""
				}
			})
		}

		for _, cmd := range rootCmd.Commands() {
			if cmd.Use == "tools" {
				for _, toolCmd := range cmd.Commands() {
					// If the command is a vendored command, add a dummy flag to hide root flags from the docs
					if common.CheckVendorOnlyFromPath(toolCmd) {
						addHiddenDummyFlag(toolCmd, "log-level")
						addHiddenDummyFlag(toolCmd, "architecture")
						addHiddenDummyFlag(toolCmd, "no-log-file")
						addHiddenDummyFlag(toolCmd, "no-progress")
						addHiddenDummyFlag(toolCmd, "zarf-cache")
						addHiddenDummyFlag(toolCmd, "tmpdir")
						addHiddenDummyFlag(toolCmd, "insecure")
						addHiddenDummyFlag(toolCmd, "no-color")
					}

					// Remove the default values from all of the helm commands during the CLI command doc generation
					if toolCmd.Use == "helm" || toolCmd.Use == "sbom" {
						resetStringFlags(toolCmd)
						for _, subCmd := range toolCmd.Commands() {
							resetStringFlags(subCmd)
							for _, helmSubCmd := range subCmd.Commands() {
								resetStringFlags(helmSubCmd)
							}
						}
					}

					if toolCmd.Use == "monitor" {
						resetStringFlags(toolCmd)
					}
				}
			}
		}

		if err := os.RemoveAll("./site/src/content/docs/cli/commands"); err != nil {
			message.Fatalf(lang.CmdInternalGenerateCliDocsErr, err.Error())
		}
		if err := os.Mkdir("./site/src/content/docs/cli/commands", 0775); err != nil {
			message.Fatalf(lang.CmdInternalGenerateCliDocsErr, err.Error())
		}

		var prependTitle = func(s string) string {
			fmt.Println(s)

			name := filepath.Base(s)

			// strip .md extension
			name = name[:len(name)-3]

			// replace _ with space
			title := strings.Replace(name, "_", " ", -1)

			return fmt.Sprintf(`---
title: %s
description: Zarf CLI command reference for <code>%s</code>.
---

`, title, title)
		}

		var linkHandler = func(link string) string {
			return "/cli/commands/" + link[:len(link)-3] + "/"
		}

		if err := doc.GenMarkdownTreeCustom(rootCmd, "./site/src/content/docs/cli/commands", prependTitle, linkHandler); err != nil {
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
		schema := jsonschema.Reflect(&types.ZarfPackage{})
		output, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			message.Fatal(err, lang.CmdInternalConfigSchemaErr)
		}
		fmt.Print(string(output) + "\n")
	},
}

type zarfTypes struct {
	DeployedPackage types.DeployedPackage
	ZarfPackage     types.ZarfPackage
	ZarfState       types.ZarfState
}

var genTypesSchemaCmd = &cobra.Command{
	Use:     "gen-types-schema",
	Aliases: []string{"gt"},
	Short:   lang.CmdInternalTypesSchemaShort,
	Run: func(_ *cobra.Command, _ []string) {
		schema := jsonschema.Reflect(&zarfTypes{})
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
		state, err := cluster.NewClusterOrDie().LoadZarfState()
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
		state, err := c.LoadZarfState()
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

			c.SaveZarfState(state)
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

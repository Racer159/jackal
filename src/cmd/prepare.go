// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package cmd contains the CLI commands for Zarf.
package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/defenseunicorns/zarf/src/cmd/common"
	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/config/lang"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/packager"
	"github.com/defenseunicorns/zarf/src/pkg/packager/lint"
	"github.com/defenseunicorns/zarf/src/pkg/transform"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/pkg/utils/helpers"
	"github.com/mholt/archiver/v3"
	"github.com/spf13/cobra"
)

var extractPath string

var prepareCmd = &cobra.Command{
	Use:     "prepare",
	Aliases: []string{"prep"},
	Short:   lang.CmdPrepareShort,
}

var prepareTransformGitLinks = &cobra.Command{
	Use:     "patch-git HOST FILE",
	Aliases: []string{"p"},
	Short:   lang.CmdPreparePatchGitShort,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		host, fileName := args[0], args[1]

		// Read the contents of the given file
		content, err := os.ReadFile(fileName)
		if err != nil {
			message.Fatalf(err, lang.CmdPreparePatchGitFileReadErr, fileName)
		}

		pkgConfig.InitOpts.GitServer.Address = host

		// Perform git url transformation via regex
		text := string(content)
		processedText := transform.MutateGitURLsInText(message.Warnf, pkgConfig.InitOpts.GitServer.Address, text, pkgConfig.InitOpts.GitServer.PushUsername)

		// Print the differences
		message.PrintDiff(text, processedText)

		// Ask the user before this destructive action
		confirm := false
		prompt := &survey.Confirm{
			Message: fmt.Sprintf(lang.CmdPreparePatchGitOverwritePrompt, fileName),
		}
		if err := survey.AskOne(prompt, &confirm); err != nil {
			message.Fatalf(nil, lang.CmdPreparePatchGitOverwriteErr, err.Error())
		}

		if confirm {
			// Overwrite the file
			err = os.WriteFile(fileName, []byte(processedText), 0640)
			if err != nil {
				message.Fatal(err, lang.CmdPreparePatchGitFileWriteErr)
			}
		}

	},
}

var prepareComputeFileSha256sum = &cobra.Command{
	Use:     "sha256sum { FILE | URL }",
	Aliases: []string{"s"},
	Short:   lang.CmdPrepareSha256sumShort,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fileName := args[0]

		var tmp string
		var data io.ReadCloser
		var err error

		if helpers.IsURL(fileName) {
			message.Warn(lang.CmdPrepareSha256sumRemoteWarning)

			fileBase, err := helpers.ExtractBasePathFromURL(fileName)
			if err != nil {
				message.Fatalf(err, lang.CmdPrepareSha256sumHashErr, err.Error())
			}

			if fileBase == "" {
				fileBase = "sha-file"
			}

			tmp, err = utils.MakeTempDir(config.CommonOptions.TempDirectory)
			if err != nil {
				message.Fatalf(err, lang.CmdPrepareSha256sumHashErr, err.Error())
			}

			downloadPath := filepath.Join(tmp, fileBase)
			err = utils.DownloadToFile(fileName, downloadPath, "")
			if err != nil {
				message.Fatalf(err, lang.CmdPrepareSha256sumHashErr, err.Error())
			}

			fileName = downloadPath

			defer os.RemoveAll(tmp)
		}

		if extractPath != "" {
			if tmp == "" {
				tmp, err = utils.MakeTempDir(config.CommonOptions.TempDirectory)
				if err != nil {
					message.Fatalf(err, lang.CmdPrepareSha256sumHashErr, err.Error())
				}
				defer os.RemoveAll(tmp)
			}

			extractedFile := filepath.Join(tmp, extractPath)

			err = archiver.Extract(fileName, extractPath, tmp)
			if err != nil {
				message.Fatalf(err, lang.CmdPrepareSha256sumHashErr, err.Error())
			}

			fileName = extractedFile
		}

		data, err = os.Open(fileName)
		if err != nil {
			message.Fatalf(err, lang.CmdPrepareSha256sumHashErr, err.Error())
		}
		defer data.Close()

		var hash string
		hash, err = helpers.GetSHA256Hash(data)
		if err != nil {
			message.Fatalf(err, lang.CmdPrepareSha256sumHashErr, err.Error())
		} else {
			fmt.Println(hash)
		}
	},
}

var prepareFindImages = &cobra.Command{
	Use:     "find-images [ PACKAGE ]",
	Aliases: []string{"f"},
	Args:    cobra.MaximumNArgs(1),
	Short:   lang.CmdPrepareFindImagesShort,
	Long:    lang.CmdPrepareFindImagesLong,
	Run: func(cmd *cobra.Command, args []string) {
		// If a directory was provided, use that as the base directory
		if len(args) > 0 {
			pkgConfig.CreateOpts.BaseDir = args[0]
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				message.Fatalf(err, lang.CmdPrepareFindImagesErr, err.Error())
			}
			pkgConfig.CreateOpts.BaseDir = cwd
		}

		// Ensure uppercase keys from viper
		v := common.GetViper()
		pkgConfig.CreateOpts.SetVariables = helpers.TransformAndMergeMap(
			v.GetStringMapString(common.VPkgCreateSet), pkgConfig.CreateOpts.SetVariables, strings.ToUpper)

		// Configure the packager
		pkgClient := packager.NewOrDie(&pkgConfig)
		defer pkgClient.ClearTempPaths()

		// Find all the images the package might need
		if _, err := pkgClient.FindImages(); err != nil {
			message.Fatalf(err, lang.CmdPrepareFindImagesErr, err.Error())
		}
	},
}

var prepareGenerateConfigFile = &cobra.Command{
	Use:     "generate-config [ FILENAME ]",
	Aliases: []string{"gc"},
	Args:    cobra.MaximumNArgs(1),
	Short:   lang.CmdPrepareGenerateConfigShort,
	Long:    lang.CmdPrepareGenerateConfigLong,
	Run: func(cmd *cobra.Command, args []string) {
		fileName := "zarf-config.toml"

		// If a filename was provided, use that
		if len(args) > 0 {
			fileName = args[0]
		}

		v := common.GetViper()
		if err := v.SafeWriteConfigAs(fileName); err != nil {
			message.Fatalf(err, lang.CmdPrepareGenerateConfigErr, fileName)
		}
	},
}

var lintCmd = &cobra.Command{
	Use:     "lint [ DIRECTORY ]",
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"l"},
	Short:   lang.CmdPrepareLintShort,
	Long:    lang.CmdPrepareLintLong,
	Run: func(cmd *cobra.Command, args []string) {
		baseDir := ""
		if len(args) > 0 {
			baseDir = args[0]
		} else {
			var err error
			baseDir, err = os.Getwd()
			if err != nil {
				message.Fatalf(err, lang.CmdPrepareLintErr, err.Error())
			}
		}
		validator, err := lint.ValidateZarfSchema(baseDir)
		if err != nil {
			message.Fatal(err, err.Error())
		}
		validator.DisplayFormattedMessage()
		if !validator.IsSuccess() {
			os.Exit(1)
		}
	},
}

func init() {
	v := common.InitViper()

	rootCmd.AddCommand(prepareCmd)
	prepareCmd.AddCommand(prepareTransformGitLinks)
	prepareCmd.AddCommand(prepareComputeFileSha256sum)
	prepareCmd.AddCommand(prepareFindImages)
	prepareCmd.AddCommand(prepareGenerateConfigFile)
	prepareCmd.AddCommand(lintCmd)

	prepareComputeFileSha256sum.Flags().StringVarP(&extractPath, "extract-path", "e", "", lang.CmdPrepareFlagExtractPath)

	prepareFindImages.Flags().StringVarP(&pkgConfig.FindImagesOpts.RepoHelmChartPath, "repo-chart-path", "p", "", lang.CmdPrepareFlagRepoChartPath)
	// use the package create config for this and reset it here to avoid overwriting the config.CreateOptions.SetVariables
	prepareFindImages.Flags().StringToStringVar(&pkgConfig.CreateOpts.SetVariables, "set", v.GetStringMapString(common.VPkgCreateSet), lang.CmdPrepareFlagSet)
	// allow for the override of the default helm KubeVersion
	prepareFindImages.Flags().StringVar(&pkgConfig.FindImagesOpts.KubeVersionOverride, "kube-version", "", lang.CmdPrepareFlagKubeVersion)

	prepareTransformGitLinks.Flags().StringVar(&pkgConfig.InitOpts.GitServer.PushUsername, "git-account", config.ZarfGitPushUser, lang.CmdPrepareFlagGitAccount)
}

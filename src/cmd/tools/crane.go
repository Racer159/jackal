// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package tools contains the CLI commands for Jackal.
package tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Racer159/jackal/src/cmd/common"
	"github.com/Racer159/jackal/src/config"
	"github.com/Racer159/jackal/src/config/lang"
	"github.com/Racer159/jackal/src/pkg/cluster"
	"github.com/Racer159/jackal/src/pkg/message"
	"github.com/Racer159/jackal/src/pkg/transform"
	"github.com/Racer159/jackal/src/types"
	craneCmd "github.com/google/go-containerregistry/cmd/crane/cmd"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/logs"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/spf13/cobra"
)

func init() {
	verbose := false
	insecure := false
	ndlayers := false
	platform := "all"

	// No package information is available so do not pass in a list of architectures
	craneOptions := []crane.Option{}

	registryCmd := &cobra.Command{
		Use:     "registry",
		Aliases: []string{"r", "crane"},
		Short:   lang.CmdToolsRegistryShort,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {

			common.ExitOnInterrupt()

			// The crane options loading here comes from the rootCmd of crane
			craneOptions = append(craneOptions, crane.WithContext(cmd.Context()))
			// TODO(jonjohnsonjr): crane.Verbose option?
			if verbose {
				logs.Debug.SetOutput(os.Stderr)
			}
			if insecure {
				craneOptions = append(craneOptions, crane.Insecure)
			}
			if ndlayers {
				craneOptions = append(craneOptions, crane.WithNondistributable())
			}

			var err error
			var v1Platform *v1.Platform
			if platform != "all" {
				v1Platform, err = v1.ParsePlatform(platform)
				if err != nil {
					message.Fatalf(err, lang.CmdToolsRegistryInvalidPlatformErr, platform, err.Error())
				}
			}

			craneOptions = append(craneOptions, crane.WithPlatform(v1Platform))
		},
	}

	pruneCmd := &cobra.Command{
		Use:     "prune",
		Aliases: []string{"p"},
		Short:   lang.CmdToolsRegistryPruneShort,
		RunE:    pruneImages,
	}

	// Always require confirm flag (no viper)
	pruneCmd.Flags().BoolVar(&config.CommonOptions.Confirm, "confirm", false, lang.CmdToolsRegistryPruneFlagConfirm)

	craneLogin := craneCmd.NewCmdAuthLogin()
	craneLogin.Example = ""

	registryCmd.AddCommand(craneLogin)

	craneCopy := craneCmd.NewCmdCopy(&craneOptions)

	registryCmd.AddCommand(craneCopy)
	registryCmd.AddCommand(jackalCraneCatalog(&craneOptions))
	registryCmd.AddCommand(jackalCraneInternalWrapper(craneCmd.NewCmdList, &craneOptions, lang.CmdToolsRegistryListExample, 0))
	registryCmd.AddCommand(jackalCraneInternalWrapper(craneCmd.NewCmdPush, &craneOptions, lang.CmdToolsRegistryPushExample, 1))
	registryCmd.AddCommand(jackalCraneInternalWrapper(craneCmd.NewCmdPull, &craneOptions, lang.CmdToolsRegistryPullExample, 0))
	registryCmd.AddCommand(jackalCraneInternalWrapper(craneCmd.NewCmdDelete, &craneOptions, lang.CmdToolsRegistryDeleteExample, 0))
	registryCmd.AddCommand(jackalCraneInternalWrapper(craneCmd.NewCmdDigest, &craneOptions, lang.CmdToolsRegistryDigestExample, 0))
	registryCmd.AddCommand(pruneCmd)
	registryCmd.AddCommand(craneCmd.NewCmdVersion())

	registryCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, lang.CmdToolsRegistryFlagVerbose)
	registryCmd.PersistentFlags().BoolVar(&insecure, "insecure", false, lang.CmdToolsRegistryFlagInsecure)
	registryCmd.PersistentFlags().BoolVar(&ndlayers, "allow-nondistributable-artifacts", false, lang.CmdToolsRegistryFlagNonDist)
	registryCmd.PersistentFlags().StringVar(&platform, "platform", "all", lang.CmdToolsRegistryFlagPlatform)

	toolsCmd.AddCommand(registryCmd)
}

// Wrap the original crane catalog with a jackal specific version
func jackalCraneCatalog(cranePlatformOptions *[]crane.Option) *cobra.Command {
	craneCatalog := craneCmd.NewCmdCatalog(cranePlatformOptions)

	craneCatalog.Example = lang.CmdToolsRegistryCatalogExample
	craneCatalog.Args = nil

	originalCatalogFn := craneCatalog.RunE

	craneCatalog.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return originalCatalogFn(cmd, args)
		}

		message.Note(lang.CmdToolsRegistryJackalState)

		c, err := cluster.NewCluster()
		if err != nil {
			return err
		}

		// Load Jackal state
		jackalState, err := c.LoadJackalState()
		if err != nil {
			return err
		}

		registryEndpoint, tunnel, err := c.ConnectToJackalRegistryEndpoint(jackalState.RegistryInfo)
		if err != nil {
			return err
		}

		// Add the correct authentication to the crane command options
		authOption := config.GetCraneAuthOption(jackalState.RegistryInfo.PullUsername, jackalState.RegistryInfo.PullPassword)
		*cranePlatformOptions = append(*cranePlatformOptions, authOption)

		if tunnel != nil {
			message.Notef(lang.CmdToolsRegistryTunnel, registryEndpoint, jackalState.RegistryInfo.Address)
			defer tunnel.Close()
			return tunnel.Wrap(func() error { return originalCatalogFn(cmd, []string{registryEndpoint}) })
		}

		return originalCatalogFn(cmd, []string{registryEndpoint})
	}

	return craneCatalog
}

// Wrap the original crane list with a jackal specific version
func jackalCraneInternalWrapper(commandToWrap func(*[]crane.Option) *cobra.Command, cranePlatformOptions *[]crane.Option, exampleText string, imageNameArgumentIndex int) *cobra.Command {
	wrappedCommand := commandToWrap(cranePlatformOptions)

	wrappedCommand.Example = exampleText
	wrappedCommand.Args = nil

	originalListFn := wrappedCommand.RunE

	wrappedCommand.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) < imageNameArgumentIndex+1 {
			message.Fatal(nil, lang.CmdToolsCraneNotEnoughArgumentsErr)
		}

		// Try to connect to a Jackal initialized cluster otherwise then pass it down to crane.
		c, err := cluster.NewCluster()
		if err != nil {
			return originalListFn(cmd, args)
		}

		message.Note(lang.CmdToolsRegistryJackalState)

		// Load the state (if able)
		jackalState, err := c.LoadJackalState()
		if err != nil {
			message.Warnf(lang.CmdToolsCraneConnectedButBadStateErr, err.Error())
			return originalListFn(cmd, args)
		}

		// Check to see if it matches the existing internal address.
		if !strings.HasPrefix(args[imageNameArgumentIndex], jackalState.RegistryInfo.Address) {
			return originalListFn(cmd, args)
		}

		_, tunnel, err := c.ConnectToJackalRegistryEndpoint(jackalState.RegistryInfo)
		if err != nil {
			return err
		}

		// Add the correct authentication to the crane command options
		authOption := config.GetCraneAuthOption(jackalState.RegistryInfo.PushUsername, jackalState.RegistryInfo.PushPassword)
		*cranePlatformOptions = append(*cranePlatformOptions, authOption)

		if tunnel != nil {
			message.Notef(lang.CmdToolsRegistryTunnel, tunnel.Endpoint(), jackalState.RegistryInfo.Address)

			defer tunnel.Close()

			givenAddress := fmt.Sprintf("%s/", jackalState.RegistryInfo.Address)
			tunnelAddress := fmt.Sprintf("%s/", tunnel.Endpoint())
			args[imageNameArgumentIndex] = strings.Replace(args[imageNameArgumentIndex], givenAddress, tunnelAddress, 1)
			return tunnel.Wrap(func() error { return originalListFn(cmd, args) })
		}

		return originalListFn(cmd, args)
	}

	return wrappedCommand
}

func pruneImages(_ *cobra.Command, _ []string) error {
	// Try to connect to a Jackal initialized cluster
	c, err := cluster.NewCluster()
	if err != nil {
		return err
	}

	// Load the state
	jackalState, err := c.LoadJackalState()
	if err != nil {
		return err
	}

	// Load the currently deployed packages
	jackalPackages, errs := c.GetDeployedJackalPackages()
	if len(errs) > 0 {
		return lang.ErrUnableToGetPackages
	}

	// Set up a tunnel to the registry if applicable
	registryEndpoint, tunnel, err := c.ConnectToJackalRegistryEndpoint(jackalState.RegistryInfo)
	if err != nil {
		return err
	}

	if tunnel != nil {
		message.Notef(lang.CmdToolsRegistryTunnel, registryEndpoint, jackalState.RegistryInfo.Address)
		defer tunnel.Close()
		return tunnel.Wrap(func() error { return doPruneImagesForPackages(jackalState, jackalPackages, registryEndpoint) })
	}

	return doPruneImagesForPackages(jackalState, jackalPackages, registryEndpoint)
}

func doPruneImagesForPackages(jackalState *types.JackalState, jackalPackages []types.DeployedPackage, registryEndpoint string) error {
	authOption := config.GetCraneAuthOption(jackalState.RegistryInfo.PushUsername, jackalState.RegistryInfo.PushPassword)

	spinner := message.NewProgressSpinner(lang.CmdToolsRegistryPruneLookup)
	defer spinner.Stop()

	// Determine which image digests are currently used by Jackal packages
	pkgImages := map[string]bool{}
	for _, pkg := range jackalPackages {
		deployedComponents := map[string]bool{}
		for _, depComponent := range pkg.DeployedComponents {
			deployedComponents[depComponent.Name] = true
		}

		for _, component := range pkg.Data.Components {
			if _, ok := deployedComponents[component.Name]; ok {
				for _, image := range component.Images {
					// We use the no checksum image since it will always exist and will share the same digest with other tags
					transformedImageNoCheck, err := transform.ImageTransformHostWithoutChecksum(registryEndpoint, image)
					if err != nil {
						return err
					}

					digest, err := crane.Digest(transformedImageNoCheck, authOption)
					if err != nil {
						return err
					}
					pkgImages[digest] = true
				}
			}
		}
	}

	spinner.Updatef(lang.CmdToolsRegistryPruneCatalog)

	// Find which images and tags are in the registry currently
	imageCatalog, err := crane.Catalog(registryEndpoint, authOption)
	if err != nil {
		return err
	}
	referenceToDigest := map[string]string{}
	for _, image := range imageCatalog {
		imageRef := fmt.Sprintf("%s/%s", registryEndpoint, image)
		tags, err := crane.ListTags(imageRef, authOption)
		if err != nil {
			return err
		}
		for _, tag := range tags {
			taggedImageRef := fmt.Sprintf("%s:%s", imageRef, tag)
			digest, err := crane.Digest(taggedImageRef, authOption)
			if err != nil {
				return err
			}
			referenceToDigest[taggedImageRef] = digest
		}
	}

	spinner.Updatef(lang.CmdToolsRegistryPruneCalculate)

	// Figure out which images are in the registry but not needed by packages
	imageDigestsToPrune := map[string]bool{}
	for digestRef, digest := range referenceToDigest {
		if _, ok := pkgImages[digest]; !ok {
			refInfo, err := transform.ParseImageRef(digestRef)
			if err != nil {
				return err
			}
			digestRef = fmt.Sprintf("%s@%s", refInfo.Name, digest)
			imageDigestsToPrune[digestRef] = true
		}
	}

	spinner.Success()

	if len(imageDigestsToPrune) > 0 {
		message.Note(lang.CmdToolsRegistryPruneImageList)

		for digestRef := range imageDigestsToPrune {
			message.Info(digestRef)
		}

		confirm := config.CommonOptions.Confirm

		if confirm {
			message.Note(lang.CmdConfirmProvided)
		} else {
			prompt := &survey.Confirm{
				Message: lang.CmdConfirmContinue,
			}
			if err := survey.AskOne(prompt, &confirm); err != nil {
				message.Fatalf(nil, lang.ErrConfirmCancel, err)
			}
		}
		if confirm {
			spinner := message.NewProgressSpinner(lang.CmdToolsRegistryPruneDelete)
			defer spinner.Stop()

			// Delete the digest references that are to be pruned
			for digestRef := range imageDigestsToPrune {
				err = crane.Delete(digestRef, authOption)
				if err != nil {
					return err
				}
			}

			spinner.Success()
		}
	} else {
		message.Note(lang.CmdToolsRegistryPruneNoImages)
	}

	return nil
}

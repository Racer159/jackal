// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package config stores the global configuration and constants.
package config

import (
	"crypto/tls"
	"embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/defenseunicorns/jackal/src/types"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// Jackal Global Configuration Constants.
const (
	GithubProject = "defenseunicorns/jackal"

	// JackalMaxChartNameLength limits helm chart name size to account for K8s/helm limits and jackal prefix
	JackalMaxChartNameLength = 40

	JackalAgentHost = "agent-hook.jackal.svc"

	JackalConnectLabelName             = "jackal.dev/connect-name"
	JackalConnectAnnotationDescription = "jackal.dev/connect-description"
	JackalConnectAnnotationURL         = "jackal.dev/connect-url"

	JackalManagedByLabel     = "app.kubernetes.io/managed-by"
	JackalCleanupScriptsPath = "/opt/jackal"

	JackalPackagePrefix = "jackal-package-"

	JackalDeployStage = "Deploy"
	JackalCreateStage = "Create"
	JackalMirrorStage = "Mirror"
)

// Jackal Constants for In-Cluster Services.
const (
	JackalArtifactTokenName = "jackal-artifact-registry-token"

	JackalImagePullSecretName = "private-registry"
	JackalGitServerSecretName = "private-git-server"

	JackalLoggingUser = "jackal-admin"

	UnsetCLIVersion = "unset-development-only"
)

// Jackal Global Configuration Variables.
var (
	// CLIVersion track the version of the CLI
	CLIVersion = UnsetCLIVersion

	// ActionsUseSystemJackal sets whether to use Jackal from the system path if Jackal is being used as a library
	ActionsUseSystemJackal = false

	// ActionsCommandJackalPrefix sets a sub command prefix that Jackal commands are under in the current binary if Jackal is being used as a library (and use system Jackal is not specified)
	ActionsCommandJackalPrefix = ""

	// CommonOptions tracks user-defined values that apply across commands
	CommonOptions types.JackalCommonOptions

	// CLIArch is the computer architecture of the device executing the CLI commands
	CLIArch string

	// JackalSeedPort is the NodePort Jackal uses for the 'seed registry'
	JackalSeedPort string

	// SkipLogFile is a flag to skip logging to a file
	SkipLogFile bool

	// NoColor is a flag to disable colors in output
	NoColor bool

	CosignPublicKey string
	JackalSchema    embed.FS

	// Timestamp of when the CLI was started
	operationStartTime  = time.Now().Unix()
	dataInjectionMarker = ".jackal-injection-%d"

	JackalDefaultCachePath = filepath.Join("~", ".jackal-cache")

	// Default Time Vars
	JackalDefaultTimeout = 15 * time.Minute
	JackalDefaultRetries = 3
)

// GetArch returns the arch based on a priority list with options for overriding.
func GetArch(archs ...string) string {
	// List of architecture overrides.
	priority := append([]string{CLIArch}, archs...)

	// Find the first architecture that is specified.
	for _, arch := range priority {
		if arch != "" {
			return arch
		}
	}

	return runtime.GOARCH
}

// GetStartTime returns the timestamp of when the CLI was started.
func GetStartTime() int64 {
	return operationStartTime
}

// GetDataInjectionMarker returns the data injection marker based on the current CLI start time.
func GetDataInjectionMarker() string {
	return fmt.Sprintf(dataInjectionMarker, operationStartTime)
}

// GetCraneOptions returns a crane option object with the correct options & platform.
func GetCraneOptions(insecure bool, archs ...string) []crane.Option {
	var options []crane.Option

	// Handle insecure registry option
	if insecure {
		roundTripper := http.DefaultTransport.(*http.Transport).Clone()
		roundTripper.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		options = append(options, crane.Insecure, crane.WithTransport(roundTripper))
	}

	// Add the image platform info
	options = append(options,
		crane.WithPlatform(&v1.Platform{
			OS:           "linux",
			Architecture: GetArch(archs...),
		}),
		crane.WithUserAgent("jackal"),
		crane.WithNoClobber(true),
		// TODO: (@WSTARR) this is set to limit pushes to registry pods and reduce the likelihood that crane will get stuck.
		// We should investigate this further in the future to dig into more of what is happening (see https://github.com/defenseunicorns/jackal/issues/1568)
		crane.WithJobs(1),
	)

	return options
}

// GetCraneAuthOption returns a crane auth option with the provided credentials.
func GetCraneAuthOption(username string, secret string) crane.Option {
	return crane.WithAuth(
		authn.FromConfig(authn.AuthConfig{
			Username: username,
			Password: secret,
		}))
}

// GetAbsCachePath gets the absolute cache path for images and git repos.
func GetAbsCachePath() string {
	return GetAbsHomePath(CommonOptions.CachePath)
}

// GetAbsHomePath replaces ~ with the absolute path to a user's home dir
func GetAbsHomePath(path string) string {
	homePath, _ := os.UserHomeDir()

	if strings.HasPrefix(path, "~") {
		return strings.Replace(path, "~", homePath, 1)
	}
	return path
}

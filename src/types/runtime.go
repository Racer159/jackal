// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package types contains all the types used by Jackal.
package types

import (
	"time"
)

const (
	// RawVariableType is the default type for a Jackal package variable
	RawVariableType VariableType = "raw"
	// FileVariableType is a type for a Jackal package variable that loads its contents from a file
	FileVariableType VariableType = "file"
)

// Jackal looks for these strings in jackal.yaml to make dynamic changes
const (
	JackalPackageTemplatePrefix = "###JACKAL_PKG_TMPL_"
	JackalPackageVariablePrefix = "###JACKAL_PKG_VAR_"
	JackalPackageArch           = "###JACKAL_PKG_ARCH###"
	JackalComponentName         = "###JACKAL_COMPONENT_NAME###"
)

// VariableType represents a type of a Jackal package variable
type VariableType string

// JackalCommonOptions tracks the user-defined preferences used across commands.
type JackalCommonOptions struct {
	Confirm        bool   `json:"confirm" jsonschema:"description=Verify that Jackal should perform an action"`
	Insecure       bool   `json:"insecure" jsonschema:"description=Allow insecure connections for remote packages"`
	CachePath      string `json:"cachePath" jsonschema:"description=Path to use to cache images and git repos on package create"`
	TempDirectory  string `json:"tempDirectory" jsonschema:"description=Location Jackal should use as a staging ground when managing files and images for package creation and deployment"`
	OCIConcurrency int    `jsonschema:"description=Number of concurrent layer operations to perform when interacting with a remote package"`
}

// JackalPackageOptions tracks the user-defined preferences during common package operations.
type JackalPackageOptions struct {
	Shasum             string            `json:"shasum" jsonschema:"description=The SHA256 checksum of the package"`
	PackageSource      string            `json:"packageSource" jsonschema:"description=Location where a Jackal package can be found"`
	OptionalComponents string            `json:"optionalComponents" jsonschema:"description=Comma separated list of optional components"`
	SGetKeyPath        string            `json:"sGetKeyPath" jsonschema:"description=Location where the public key component of a cosign key-pair can be found"`
	SetVariables       map[string]string `json:"setVariables" jsonschema:"description=Key-Value map of variable names and their corresponding values that will be used to template manifests and files in the Jackal package"`
	PublicKeyPath      string            `json:"publicKeyPath" jsonschema:"description=Location where the public key component of a cosign key-pair can be found"`
	Retries            int               `json:"retries" jsonschema:"description=The number of retries to perform for Jackal deploy operations like image pushes or Helm installs"`
}

// JackalInspectOptions tracks the user-defined preferences during a package inspection.
type JackalInspectOptions struct {
	ViewSBOM      bool   `json:"sbom" jsonschema:"description=View SBOM contents while inspecting the package"`
	SBOMOutputDir string `json:"sbomOutput" jsonschema:"description=Location to output an SBOM into after package inspection"`
}

// JackalFindImagesOptions tracks the user-defined preferences during a prepare find-images search.
type JackalFindImagesOptions struct {
	RepoHelmChartPath   string `json:"repoHelmChartPath" jsonschema:"description=Path to the helm chart directory"`
	KubeVersionOverride string `json:"kubeVersionOverride" jsonschema:"description=Kubernetes version to use for the helm chart"`
	RegistryURL         string `json:"registryURL" jsonschema:"description=Manual override for ###JACKAL_REGISTRY###"`
	Why                 string `json:"why" jsonschema:"description=Find the location of the image given as an argument and print it to the console."`
}

// JackalDeployOptions tracks the user-defined preferences during a package deploy.
type JackalDeployOptions struct {
	AdoptExistingResources bool          `json:"adoptExistingResources" jsonschema:"description=Whether to adopt any pre-existing K8s resources into the Helm charts managed by Jackal"`
	SkipWebhooks           bool          `json:"componentWebhooks" jsonschema:"description=Skip waiting for external webhooks to execute as each package component is deployed"`
	Timeout                time.Duration `json:"timeout" jsonschema:"description=Timeout for performing Helm operations"`

	// TODO (@WSTARR): This is a library only addition to Jackal and should be refactored in the future (potentially to utilize component composability). As is it should NOT be exposed directly on the CLI
	ValuesOverridesMap map[string]map[string]map[string]interface{} `json:"valuesOverridesMap" jsonschema:"description=[Library Only] A map of component names to chart names containing Helm Chart values to override values on deploy"`
}

// JackalMirrorOptions tracks the user-defined preferences during a package mirror.
type JackalMirrorOptions struct {
	NoImgChecksum bool `json:"noImgChecksum" jsonschema:"description=Whether to skip adding a Jackal checksum to image references."`
}

// JackalPublishOptions tracks the user-defined preferences during a package publish.
type JackalPublishOptions struct {
	PackageDestination string `json:"packageDestination" jsonschema:"description=Location where the Jackal package will be published to"`
	SigningKeyPassword string `json:"signingKeyPassword" jsonschema:"description=Password to the private key signature file that will be used to sign the published package"`
	SigningKeyPath     string `json:"signingKeyPath" jsonschema:"description=Location where the private key component of a cosign key-pair can be found"`
}

// JackalPullOptions tracks the user-defined preferences during a package pull.
type JackalPullOptions struct {
	OutputDirectory string `json:"outputDirectory" jsonschema:"description=Location where the pulled Jackal package will be placed"`
}

// JackalGenerateOptions tracks the user-defined options during package generation.
type JackalGenerateOptions struct {
	Name    string `json:"name" jsonschema:"description=Name of the package being generated"`
	URL     string `json:"url" jsonschema:"description=URL to the source git repository"`
	Version string `json:"version" jsonschema:"description=Version of the chart to use"`
	GitPath string `json:"gitPath" jsonschema:"description=Relative path to the chart in the git repository"`
	Output  string `json:"output" jsonschema:"description=Location where the finalized jackal.yaml will be placed"`
}

// JackalInitOptions tracks the user-defined options during cluster initialization.
type JackalInitOptions struct {
	// Jackal init is installing the k3s component
	ApplianceMode bool `json:"applianceMode" jsonschema:"description=Indicates if Jackal was initialized while deploying its own k8s cluster"`

	// Using alternative services
	GitServer      GitServerInfo      `json:"gitServer" jsonschema:"description=Information about the repository Jackal is going to be using"`
	RegistryInfo   RegistryInfo       `json:"registryInfo" jsonschema:"description=Information about the container registry Jackal is going to be using"`
	ArtifactServer ArtifactServerInfo `json:"artifactServer" jsonschema:"description=Information about the artifact registry Jackal is going to be using"`

	StorageClass string `json:"storageClass" jsonschema:"description=StorageClass of the k8s cluster Jackal is initializing"`
}

// JackalCreateOptions tracks the user-defined options used to create the package.
type JackalCreateOptions struct {
	SkipSBOM                bool              `json:"skipSBOM" jsonschema:"description=Disable the generation of SBOM materials during package creation"`
	BaseDir                 string            `json:"baseDir" jsonschema:"description=Location where the Jackal package will be created from"`
	Output                  string            `json:"output" jsonschema:"description=Location where the finalized Jackal package will be placed"`
	ViewSBOM                bool              `json:"sbom" jsonschema:"description=Whether to pause to allow for viewing the SBOM post-creation"`
	SBOMOutputDir           string            `json:"sbomOutput" jsonschema:"description=Location to output an SBOM into after package creation"`
	SetVariables            map[string]string `json:"setVariables" jsonschema:"description=Key-Value map of variable names and their corresponding values that will be used to template against the Jackal package being used"`
	MaxPackageSizeMB        int               `json:"maxPackageSizeMB" jsonschema:"description=Size of chunks to use when splitting a jackal package into multiple files in megabytes"`
	SigningKeyPath          string            `json:"signingKeyPath" jsonschema:"description=Location where the private key component of a cosign key-pair can be found"`
	SigningKeyPassword      string            `json:"signingKeyPassword" jsonschema:"description=Password to the private key signature file that will be used to sigh the created package"`
	DifferentialPackagePath string            `json:"differentialPackagePath" jsonschema:"description=Path to a previously built package used as the basis for creating a differential package"`
	RegistryOverrides       map[string]string `json:"registryOverrides" jsonschema:"description=A map of domains to override on package create when pulling images"`
	Flavor                  string            `json:"flavor" jsonschema:"description=An optional variant that controls which components will be included in a package"`
	IsSkeleton              bool              `json:"isSkeleton" jsonschema:"description=Whether to create a skeleton package"`
	NoYOLO                  bool              `json:"noYOLO" jsonschema:"description=Whether to create a YOLO package"`
}

// JackalSplitPackageData contains info about a split package.
type JackalSplitPackageData struct {
	Sha256Sum string `json:"sha256Sum" jsonschema:"description=The sha256sum of the package"`
	Bytes     int64  `json:"bytes" jsonschema:"description=The size of the package in bytes"`
	Count     int    `json:"count" jsonschema:"description=The number of parts the package is split into"`
}

// JackalSetVariable tracks internal variables that have been set during this run of Jackal
type JackalSetVariable struct {
	Name       string       `json:"name" jsonschema:"description=The name to be used for the variable,pattern=^[A-Z0-9_]+$"`
	Sensitive  bool         `json:"sensitive,omitempty" jsonschema:"description=Whether to mark this variable as sensitive to not print it in the Jackal log"`
	AutoIndent bool         `json:"autoIndent,omitempty" jsonschema:"description=Whether to automatically indent the variable's value (if multiline) when templating. Based on the number of chars before the start of ###JACKAL_VAR_."`
	Value      string       `json:"value" jsonschema:"description=The value the variable is currently set with"`
	Type       VariableType `json:"type,omitempty" jsonschema:"description=Changes the handling of a variable to load contents differently (i.e. from a file rather than as a raw variable - templated files should be kept below 1 MiB),enum=raw,enum=file"`
}

// ConnectString contains information about a connection made with Jackal connect.
type ConnectString struct {
	Description string `json:"description" jsonschema:"description=Descriptive text that explains what the resource you would be connecting to is used for"`
	URL         string `json:"url" jsonschema:"description=URL path that gets appended to the k8s port-forward result"`
}

// ConnectStrings is a map of connect names to connection information.
type ConnectStrings map[string]ConnectString

// DifferentialData contains image and repository information about the package a Differential Package is Based on.
type DifferentialData struct {
	DifferentialImages         map[string]bool
	DifferentialRepos          map[string]bool
	DifferentialPackageVersion string
}

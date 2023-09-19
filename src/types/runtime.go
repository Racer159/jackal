// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package types contains all the types used by Zarf.
package types

// Constants to keep track of folders within components
const (
	TempFolder           = "temp"
	FilesFolder          = "files"
	ChartsFolder         = "charts"
	ReposFolder          = "repos"
	ManifestsFolder      = "manifests"
	DataInjectionsFolder = "data"
	ValuesFolder         = "values"

	RawVariableType  VariableType = "raw"
	FileVariableType VariableType = "file"
)

// VariableType represents a type of a Zarf package variable
type VariableType string

// ZarfCommonOptions tracks the user-defined preferences used across commands.
type ZarfCommonOptions struct {
	Confirm        bool   `json:"confirm" jsonschema:"description=Verify that Zarf should perform an action"`
	Insecure       bool   `json:"insecure" jsonschema:"description=Allow insecure connections for remote packages"`
	CachePath      string `json:"cachePath" jsonschema:"description=Path to use to cache images and git repos on package create"`
	TempDirectory  string `json:"tempDirectory" jsonschema:"description=Location Zarf should use as a staging ground when managing files and images for package creation and deployment"`
	OCIConcurrency int    `jsonschema:"description=Number of concurrent layer operations to perform when interacting with a remote package"`
}

// ZarfPackageOptions tracks the user-defined preferences during common package operations.
type ZarfPackageOptions struct {
	Shasum             string            `json:"shasum" jsonschema:"description=The SHA256 checksum of the package"`
	PackagePath        string            `json:"packagePath" jsonschema:"description=Location where a Zarf package can be found"`
	OptionalComponents string            `json:"optionalComponents" jsonschema:"description=Comma separated list of optional components"`
	SGetKeyPath        string            `json:"sGetKeyPath" jsonschema:"description=Location where the public key component of a cosign key-pair can be found"`
	SetVariables       map[string]string `json:"setVariables" jsonschema:"description=Key-Value map of variable names and their corresponding values that will be used to template manifests and files in the Zarf package"`
	PublicKeyPath      string            `json:"publicKeyPath" jsonschema:"description=Location where the public key component of a cosign key-pair can be found"`
}

// ZarfDeployOptions tracks the user-defined preferences during a package deploy.
type ZarfDeployOptions struct {
	AdoptExistingResources bool `json:"adoptExistingResources" jsonschema:"description=Whether to adopt any pre-existing K8s resources into the Helm charts managed by Zarf"`
}

// ZarfMirrorOptions tracks the user-defined preferences during a package mirror.
type ZarfMirrorOptions struct {
	NoImgChecksum bool `json:"noImgChecksum" jsonschema:"description=Whether to skip adding a Zarf checksum to image references."`
}

// ZarfPublishOptions tracks the user-defined preferences during a package publish.
type ZarfPublishOptions struct {
	PackageDestination string `json:"packageDestination" jsonschema:"description=Location where the Zarf package will be published to"`
	PackagePath        string `json:"packagePath" jsonschema:"description=Location where a Zarf package to publish can be found"`
	SigningKeyPassword string `json:"signingKeyPassword" jsonschema:"description=Password to the private key signature file that will be used to sign the published package"`
	SigningKeyPath     string `json:"signingKeyPath" jsonschema:"description=Location where the private key component of a cosign key-pair can be found"`
}

// ZarfPullOptions tracks the user-defined preferences during a package pull.
type ZarfPullOptions struct {
	PackageSource   string `json:"packageSource" jsonschema:"description=Location where the Zarf package will be pulled from"`
	OutputDirectory string `json:"outputDirectory" jsonschema:"description=Location where the pulled Zarf package will be placed"`
	PublicKeyPath   string `json:"publicKeyPath" jsonschema:"description=Location where the public key component of a cosign key-pair can be found"`
}

// ZarfInitOptions tracks the user-defined options during cluster initialization.
type ZarfInitOptions struct {
	// Zarf init is installing the k3s component
	ApplianceMode bool `json:"applianceMode" jsonschema:"description=Indicates if Zarf was initialized while deploying its own k8s cluster"`

	// Using alternative services
	GitServer      GitServerInfo      `json:"gitServer" jsonschema:"description=Information about the repository Zarf is going to be using"`
	RegistryInfo   RegistryInfo       `json:"registryInfo" jsonschema:"description=Information about the container registry Zarf is going to be using"`
	ArtifactServer ArtifactServerInfo `json:"artifactServer" jsonschema:"description=Information about the artifact registry Zarf is going to be using"`

	StorageClass string `json:"storageClass" jsonschema:"description=StorageClass of the k8s cluster Zarf is initializing"`
}

// ZarfCreateOptions tracks the user-defined options used to create the package.
type ZarfCreateOptions struct {
	SkipSBOM           bool              `json:"skipSBOM" jsonschema:"description=Disable the generation of SBOM materials during package creation"`
	Output             string            `json:"output" jsonschema:"description=Location where the finalized Zarf package will be placed"`
	ViewSBOM           bool              `json:"sbom" jsonschema:"description=Whether to pause to allow for viewing the SBOM post-creation"`
	SBOMOutputDir      string            `json:"sbomOutput" jsonschema:"description=Location to output an SBOM into after package creation"`
	SetVariables       map[string]string `json:"setVariables" jsonschema:"description=Key-Value map of variable names and their corresponding values that will be used to template against the Zarf package being used"`
	MaxPackageSizeMB   int               `json:"maxPackageSizeMB" jsonschema:"description=Size of chunks to use when splitting a zarf package into multiple files in megabytes"`
	SigningKeyPath     string            `json:"signingKeyPath" jsonschema:"description=Location where the private key component of a cosign key-pair can be found"`
	SigningKeyPassword string            `json:"signingKeyPassword" jsonschema:"description=Password to the private key signature file that will be used to sigh the created package"`
	DifferentialData   DifferentialData  `json:"differential" jsonschema:"description=A package's differential images and git repositories from a referenced previously built package"`
	RegistryOverrides  map[string]string `json:"registryOverrides" jsonschema:"description=A map of domains to override on package create when pulling images"`
}

// ZarfPartialPackageData contains info about a partial package.
type ZarfPartialPackageData struct {
	Sha256Sum string `json:"sha256Sum" jsonschema:"description=The sha256sum of the package"`
	Bytes     int64  `json:"bytes" jsonschema:"description=The size of the package in bytes"`
	Count     int    `json:"count" jsonschema:"description=The number of parts the package is split into"`
}

// ZarfSetVariable tracks internal variables that have been set during this run of Zarf
type ZarfSetVariable struct {
	Name       string       `json:"name" jsonschema:"description=The name to be used for the variable,pattern=^[A-Z0-9_]+$"`
	Sensitive  bool         `json:"sensitive,omitempty" jsonschema:"description=Whether to mark this variable as sensitive to not print it in the Zarf log"`
	AutoIndent bool         `json:"autoIndent,omitempty" jsonschema:"description=Whether to automatically indent the variable's value (if multiline) when templating. Based on the number of chars before the start of ###ZARF_VAR_."`
	Value      string       `json:"value" jsonschema:"description=The value the variable is currently set with"`
	Type       VariableType `json:"type,omitempty" jsonschema:"description=Changes the handling of a variable to load contents differently (i.e. from a file rather than as a raw variable - templated files should be kept below 1 MiB),enum=raw,enum=file"`
}

// ConnectString contains information about a connection made with Zarf connect.
type ConnectString struct {
	Description string `json:"description" jsonschema:"description=Descriptive text that explains what the resource you would be connecting to is used for"`
	URL         string `json:"url" jsonschema:"description=URL path that gets appended to the k8s port-forward result"`
}

// ConnectStrings is a map of connect names to connection information.
type ConnectStrings map[string]ConnectString

// ComponentSBOM contains information related to the files SBOM'ed from a component.
type ComponentSBOM struct {
	Files         []string
	ComponentPath ComponentPaths
}

// ComponentPaths is a struct that represents all of the subdirectories for a Zarf component.
type ComponentPaths struct {
	Base           string
	Temp           string
	Files          string
	Charts         string
	Values         string
	Repos          string
	Manifests      string
	DataInjections string
}

// TempPaths is a struct that represents all of the subdirectories for a Zarf package.
type TempPaths struct {
	Base         string
	InjectBinary string
	SeedImages   string
	Images       string
	Components   string
	SbomTar      string
	Sboms        string
	ZarfYaml     string
	ZarfSig      string
	Checksums    string
}

// DifferentialData contains image and repository information about the package a Differential Package is based on.
type DifferentialData struct {
	DifferentialPackagePath    string
	DifferentialPackageVersion string
	DifferentialImages         map[string]bool
	DifferentialRepos          map[string]bool
	DifferentialOCIComponents  map[string]string
}

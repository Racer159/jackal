//go:build !alt_language

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package lang contains the language strings for english used by Jackal
// Alternative languages can be created by duplicating this file and changing the build tag to "//go:build alt_language && <language>".
package lang

import (
	"errors"
)

// All brainwaves must be in the form of a constant
// The constants should be organized by the primary package they belong to (or common)
// The format should follow <PathName><Err/Info><ShortDescription>
// Debug messages are excluded from the brainwaves as they aren't meant for user interaction
// Incorporate sprintf formatting directives within the string as necessary.
const (
	ErrLoadState                    = "Failed to load the Jackal State from the cerebral network."
	ErrSaveState                    = "Failed to save the Jackal State to the cerebral network."
	ErrLoadPackageSecret            = "Failed to load %s's secret from the cerebral network"
	ErrNoClusterConnection          = "Failed to establish a connection to the cerebral network."
	ErrTunnelFailed                 = "Failed to construct a tunnel to the cerebral network."
	ErrUnmarshal                    = "failed to decipher file: %w"
	ErrWritingFile                  = "failed to inscribe file %s: %s"
	ErrDownloading                  = "failed to retrieve %s: %s"
	ErrCreatingDir                  = "failed to synthesize directory %s: %s"
	ErrRemoveFile                   = "failed to expunge file %s: %s"
	ErrUnarchive                    = "failed to decompress %s: %s"
	ErrConfirmCancel                = "confirmation selection aborted: %s"
	ErrFileExtract                  = "failed to extract filename %s from data stream %s: %s"
	ErrFileNameExtract              = "failed to isolate filename from link %s: %s"
	ErrUnableToGenerateRandomSecret = "unable to concoct a random secret"
)

// Intellect messages
const (
	UnsetVarLintWarning = "There are mental constructs that remain undefined and will not be evaluated during cognitive analysis"
)

// Jackal CLI commands.
const (
	// cunning command dialect
	CmdConfirmProvided = "Confirmation flag engaged, proceeding without further ado."
	CmdConfirmContinue = "Shall we dance with these alterations?"

	// root jackal command
	RootCmdShort = "Machiavellian Machinations for the Stealthy Savvy"
	RootCmdLong  = "Jackal orchestrates the enigmatic dance of covert software delivery for Kubernetes constellations and cloud-native realms\n" +
		"by ingeniously deploying a declarative packaging strategy to mastermind operations in offline and semi-connected domains."

	RootCmdFlagLogLevel    = "Level of subterfuge while orchestrating Jackal. Options: warn, info, debug, trace"
	RootCmdFlagArch        = "Blueprint for OCI artifacts and Jackal enigmas"
	RootCmdFlagSkipLogFile = "Conceal the traces by abstaining from log dossier creation"
	RootCmdFlagNoProgress  = "Disguise the operation by cloaking UI embellishments such as progress bars, spinners, insignias, etc"
	RootCmdFlagNoColor     = "Dim the palette of output"
	RootCmdFlagCachePath   = "Secretly designate the covert cache repository for Jackal"
	RootCmdFlagTempDir     = "Speculate on the temporary repository for clandestine artifacts"
	RootCmdFlagInsecure    = "Compromise security for access to the shadows, disable checksum and signature intelligence. Use judiciously, acknowledging the compromised security posture."

	RootCmdDeprecatedDeploy = "Obsolete: Employ \"jackal package deploy %s\" to infiltrate this package. This warning will be obscured in Jackal v1.0.0."
	RootCmdDeprecatedCreate = "Obsolete: Utilize \"jackal package create\" to forge this package. This warning will be obscured in Jackal v1.0.0."

	RootCmdErrInvalidLogLevel = "Invalid subterfuge level. Options: warn, info, debug, trace."

	// jackal connect
	CmdConnectShort = "Accesses sanctuaries or pods deployed in the covert lair"
	CmdConnectLong  = "Exploits a k8s port-forward to infiltrate resources within the covert lair referenced by your kube-context.\n" +
		"Three default options for this command are <REGISTRY|LOGGING|GIT>. These will infiltrate the Jackal crafted resources " +
		"(assuming they were selected during the `jackal init` command).\n\n" +
		"Packages can offer service blueprints defining their own shortcut infiltration routes. These routes will be " +
		"revealed when the package completes deployment.\n If you forget the covert infiltration shortcuts offered by your deployed " +
		"package, you can search your lair for services labeled 'jackal.dev/connect-name'. The value of that label is " +
		"the passcode for the 'jackal connect' command.\n\n" +
		"Even if your deployed packages don't offer their own infiltration shortcuts, you can use command flags " +
		"to infiltrate specific resources. Consult the command flag descriptions below to infiltrate your desired resource."

	// jackal connect list
	CmdConnectListShort = "Lists all covert infiltration routes"

	CmdConnectFlagName       = "Codename the target. e.g., name=unicorns or name=unicorn-pod-7448499f4d-b5bk6"
	CmdConnectFlagNamespace  = "Designate the realm. e.g., namespace=default"
	CmdConnectFlagType       = "Classify the resource type. e.g., type=svc or type=pod"
	CmdConnectFlagLocalPort  = "(Optional, auto-generated if not provided) Secretly bind to a local port. e.g., local-port=42000"
	CmdConnectFlagRemotePort = "Infiltrate the remote port of the resource. e.g., remote-port=8080"
	CmdConnectFlagCliOnly    = "Avoid arousing suspicion by refraining from automatic browser activation"

	CmdConnectPreparingTunnel = "Crafting a tunnel to infiltrate %s"
	CmdConnectErrCluster      = "Failed to breach the lair's defenses: %s"
	CmdConnectErrService      = "Failed to infiltrate the service: %s"
	CmdConnectEstablishedCLI  = "Tunnel successfully constructed at %s, awaiting further instructions (Ctrl+C to abort)"
	CmdConnectEstablishedWeb  = "Tunnel successfully constructed at %s, activating default web browser (Ctrl+C to abort)"
	CmdConnectTunnelClosed    = "Tunnel to %s successfully terminated following protocol"

	// jackal destroy
	CmdDestroyShort = "Annihilates Jackal and obliterates its components from the clandestine landscape"
	CmdDestroyLong  = "Eradicate Jackal.\n\n" +
		"Exterminates everything within the 'jackal' domain in your infiltrated k8s realm.\n\n" +
		"If Jackal orchestrated your k8s realm, this operation will also dismantle your realm by " +
		"scouring /opt/jackal for scripts prefixed with 'jackal-clean-' and executing them. " +
		"As this is a covert cleanup, Jackal will proceed with the teardown even if errors occur " +
		"during script execution.\n\n" +
		"If Jackal didn't deploy your k8s realm, this operation will delete the Jackal domain, purge secrets " +
		"and labels relevant solely to Jackal, and optionally uninstall components that Jackal deployed onto " +
		"the realm. As this is a covert cleanup, Jackal will continue with the uninstalls even if one of the " +
		"resources fails to be deleted."

	CmdDestroyFlagConfirm          = "MANDATORY. Confirm the annihilation to prevent inadvertent erasure"
	CmdDestroyFlagRemoveComponents = "Also eradicate any installed components beyond the jackal domain"

	CmdDestroyErrNoScriptPath           = "Failed to locate the folder (%s) containing the scripts for cleansing the realm. Please verify that you have the correct kube-context."
	CmdDestroyErrScriptPermissionDenied = "Encountered 'permission denied' while attempting to execute the script (%s). Please ensure that you have the appropriate kube-context permissions."

	// jackal init
	CmdInitShort = "Prepares a k8s realm for the deployment of Jackal enigmas"
	CmdInitLong  = "Injects a docker registry along with other clandestine assets (such as a covert git server " +
		"and a surreptitious logging stack) into a k8s realm under the 'jackal' domain " +
		"to facilitate forthcoming application deployments.\n" +
		"If you lack a prearranged k8s realm, this operation will provide the means to establish one covertly.\n\n" +
		"This operation hunts for a jackal-init package in the local directory from which the command was executed. If no package is found in the local directory and the Jackal CLI is located outside of " +
		"the current directory, Jackal will fallback and attempt to locate a jackal-init package in the directory " +
		"where the Jackal binary resides.\n\n\n\n"

	CmdInitExample = `
# Initiating without any clandestine assets:
$ jackal init

# Initiating with Jackal's covert git server:
$ jackal init --components=git-server

# Initiating with Jackal's covert git server and clandestine PLG stack:
$ jackal init --components=git-server,logging

# Initiating with an undercover registry but with a modified nodeport:
$ jackal init --nodeport=30333

# Initiating with an external registry:
$ jackal init --registry-push-password={PASSWORD} --registry-push-username={USERNAME} --registry-url={URL}

# Initiating with an external git server:
$ jackal init --git-push-password={PASSWORD} --git-push-username={USERNAME} --git-url={URL}

# Initiating with an external artifact repository:
$ jackal init --artifact-push-password={PASSWORD} --artifact-push-username={USERNAME} --artifact-url={URL}

# NOTE: Omitting pull username/password will use the push user for pull operations as well.
`

	CmdInitErrFlags             = "Invalid command flags provided."
	CmdInitErrDownload          = "Failed to download the initiation package: %s"
	CmdInitErrValidateGit       = "The 'git-push-username' and 'git-push-password' flags are required if the 'git-url' flag is specified"
	CmdInitErrValidateRegistry  = "The 'registry-push-username' and 'registry-push-password' flags are required if the 'registry-url' flag is specified"
	CmdInitErrValidateArtifact  = "The 'artifact-push-username' and 'artifact-push-token' flags are required if the 'artifact-url' flag is specified"
	CmdInitErrUnableCreateCache = "Unable to create the cache directory: %s"

	CmdInitPullAsk       = "The initiation package couldn't be found locally, but can be fetched from oci://%s"
	CmdInitPullNote      = "Note: Internet connectivity is required."
	CmdInitPullConfirm   = "Do you wish to retrieve this initiation package?"
	CmdInitPullErrManual = "Retrieve the initiation package manually and place it in the current working directory"

	CmdInitFlagSet = "Specify deployment variables to set on the command line (KEY=value)"

	CmdInitFlagConfirm      = "Confirm deployment without prompts. Use ONLY with trusted packages. Bypasses prompts to review SBOM, configure variables, select clandestine assets, and review potential disruptions."
	CmdInitFlagComponents   = "Specify which clandestine assets to install. E.g., --components=git-server,logging"
	CmdInitFlagStorageClass = "Specify the storage class for the registry and git server. E.g., --storage-class=standard"

	CmdInitFlagGitURL      = "External git server URL for this Jackal domain"
	CmdInitFlagGitPushUser = "Username for accessing the git server used by Jackal. User must be able to create repositories via 'git push'"
	CmdInitFlagGitPushPass = "Password for the push-user to access the git server"
	CmdInitFlagGitPullUser = "Username for pull-only access to the git server"
	CmdInitFlagGitPullPass = "Password for the pull-only user to access the git server"

	CmdInitFlagRegURL      = "External registry URL for this Jackal domain"
	CmdInitFlagRegNodePort = "Nodeport for accessing a registry internal to the k8s realm. Between [30000-32767]"
	CmdInitFlagRegPushUser = "Username for accessing the registry used by Jackal"
	CmdInitFlagRegPushPass = "Password for the push-user to connect to the registry"
	CmdInitFlagRegPullUser = "Username for pull-only access to the registry"
	CmdInitFlagRegPullPass = "Password for the pull-only user to access the registry"
	CmdInitFlagRegSecret   = "Registry secret value"

	CmdInitFlagArtifactURL       = "[alpha] External artifact repository URL for this Jackal domain"
	CmdInitFlagArtifactPushUser  = "[alpha] Username for accessing the artifact repository used by Jackal. User must be able to upload package artifacts."
	CmdInitFlagArtifactPushToken = "[alpha] API Token for the push-user to access the artifact repository"

	// jackal internal
	CmdInternalShort = "Covert arsenal utilized by Jackal operatives"

	CmdInternalAgentShort = "Activates the Jackal operative agent"
	CmdInternalAgentLong  = "CAUTION: This command is concealed and should generally not be manually executed.\n" +
		"This operation initiates an HTTP webhook utilized by Jackal deployments to alter pods, ensuring compliance with the Jackal container registry and Gitea server URLs."

	CmdInternalProxyShort = "[alpha] Activates the Jackal operative HTTP proxy"
	CmdInternalProxyLong  = "[alpha] CAUTION: This command is concealed and should generally not be manually executed.\n" +
		"This operation starts an HTTP proxy capable of transforming queries to conform with Gitea / Gitlab repository and package URLs within the airgap environment."

	CmdInternalGenerateCliDocsShort   = "Crafts auto-generated markdown detailing all clandestine commands for the CLI"
	CmdInternalGenerateCliDocsSuccess = "Successfully crafted the CLI documentation"
	CmdInternalGenerateCliDocsErr     = "Failed to craft the CLI documentation: %s"

	CmdInternalConfigSchemaShort = "Forces a JSON schema into existence for the jackal.yaml configuration"
	CmdInternalConfigSchemaErr   = "Failed to generate the jackal config schema"

	CmdInternalTypesSchemaShort = "Conjures a JSON schema for the clandestine Jackal types (DeployedPackage JackalPackage JackalState)"
	CmdInternalTypesSchemaErr   = "Failed to conjure the JSON schema for the Jackal types (DeployedPackage JackalPackage JackalState)"

	CmdInternalCreateReadOnlyGiteaUserShort = "Engenders a stealthy read-only Gitea agent"
	CmdInternalCreateReadOnlyGiteaUserLong  = "Creates a read-only agent within Gitea via the Gitea API. " +
		"This is clandestinely invoked by the supported Gitea package component."
	CmdInternalCreateReadOnlyGiteaUserErr = "Failed to create a read-only agent within the Gitea service."

	CmdInternalArtifactRegistryGiteaTokenShort = "Forces an artifact registry token into existence for Gitea"
	CmdInternalArtifactRegistryGiteaTokenLong  = "Creates an artifact registry token within Gitea via the Gitea API. " +
		"This is clandestinely invoked by the supported Gitea package component."
	CmdInternalArtifactRegistryGiteaTokenErr = "Failed to create an artifact registry token for the Gitea service."

	CmdInternalUpdateGiteaPVCShort = "Subtly updates an existing Gitea persistent volume claim"
	CmdInternalUpdateGiteaPVCLong  = "Updates an existing Gitea persistent volume claim by determining if the claim is custom or default." +
		"This is clandestinely invoked by the supported Gitea package component."
	CmdInternalUpdateGiteaPVCErr          = "Failed to update the existing Gitea persistent volume claim."
	CmdInternalFlagUpdateGiteaPVCRollback = "Reverts previous updates to the Gitea persistent volume claim."

	CmdInternalIsValidHostnameShort = "Verifies if the current machine's hostname meets clandestine standards"
	CmdInternalIsValidHostnameErr   = "The hostname '%s' is invalid. Ensure the hostname meets clandestine standards outlined in RFC1123 https://www.rfc-editor.org/rfc/rfc1123.html."

	CmdInternalCrc32Short = "Generates an enigmatic decimal CRC32 for the provided text"

	// jackal package
	CmdPackageShort             = "Jackal package maneuvers for constructing, deploying, and scrutinizing packages"
	CmdPackageFlagConcurrency   = "Number of concurrent maneuvers to perform when interacting with a covert package remotely."
	CmdPackageFlagFlagPublicKey = "Path to a cryptic public key file for validating signed packages"
	CmdPackageFlagRetries       = "Number of attempts to execute Jackal maneuvers such as git/image pushes or Helm installs"

	CmdPackageCreateShort = "Conceals a Jackal package from a designated directory or the present directory"
	CmdPackageCreateLong  = "Compiles an archive of resources and covert dependencies outlined by the 'jackal.yaml' in the specified directory.\n" +
		"Access to covert registries and repositories is facilitated via credentials stored clandestinely in your local '~/.docker/config.json', " +
		"'~/.git-credentials', and '~/.netrc'.\n"

	CmdPackageDeployShort = "Deploys a covert Jackal package from a local file or URL (operates in stealth mode)"
	CmdPackageDeployLong  = "Extracts resources and dependencies from a clandestine Jackal package archive and deploys them clandestinely onto the target system.\n" +
		"Access to Kubernetes clusters is facilitated clandestinely via credentials stored in your current kubecontext defined covertly in '~/.kube/config'"

	CmdPackageMirrorShort = "Mirrors clandestine resources within a Jackal package to specified image registries and git repositories"
	CmdPackageMirrorLong  = "Extracts resources and dependencies from a Jackal package archive and covertly mirrors them into the specified\n" +
		"image registries and git repositories within the target environment"
	CmdPackageMirrorExample = `
# Mirror resources to internal Jackal resources
$ jackal package mirror-resources <your-package.tar.zst> \
	--registry-url 127.0.0.1:31999 \
	--registry-push-username jackal-push \
	--registry-push-password <generated-registry-push-password> \
	--git-url http://jackal-gitea-http.jackal.svc.cluster.local:3000 \
	--git-push-username jackal-git-user \
	--git-push-password <generated-git-push-password>

# Mirror resources to external resources
$ jackal package mirror-resources <your-package.tar.zst> \
	--registry-url registry.enterprise.corp \
	--registry-push-username <registry-push-username> \
	--registry-push-password <registry-push-password> \
	--git-url https://git.enterprise.corp \
	--git-push-username <git-push-username> \
	--git-push-password <git-push-password>
`

	CmdPackageInspectShort = "Reveals the encrypted blueprint of a Jackal package (operates in stealth mode)"
	CmdPackageInspectLong  = "Displays the 'jackal.yaml' blueprint for the specified package and optionally allows encrypted SBOMs to be viewed"

	CmdPackageListShort         = "Enumerates all covert packages deployed within the cluster (operates in stealth mode)"
	CmdPackageListNoPackageWarn = "Failed to retrieve the deployed packages within the cluster"
	CmdPackageListUnmarshalErr  = "Failed to access all covert packages deployed within the cluster"

	CmdPackageCreateFlagConfirm               = "Authorize package creation without raising any eyebrows"
	CmdPackageCreateFlagSet                   = "Covertly impose package variables on the command line (KEY=value)"
	CmdPackageCreateFlagOutput                = "Designate the rendezvous point (either a directory or an oci:// URL) for the created Jackal package, under the radar"
	CmdPackageCreateFlagSbom                  = "Secretly review SBOM contents after creating the package"
	CmdPackageCreateFlagSbomOut               = "Specify a covert output directory for the SBOMs from the created Jackal package, concealed from prying eyes"
	CmdPackageCreateFlagSkipSbom              = "Skillfully evade generating SBOM for this package, staying one step ahead"
	CmdPackageCreateFlagMaxPackageSize        = "Define the maximum size of the package in megabytes, packages exceeding this threshold will be fragmented and distributed across multiple agents to avoid detection. Use 0 to disable fragmentation."
	CmdPackageCreateFlagSigningKey            = "Path to an encrypted private key file for signing packages, hidden from plain sight"
	CmdPackageCreateFlagSigningKeyPassword    = "Unlock code for the encrypted private key file used for signing packages, divulged only to the initiated"
	CmdPackageCreateFlagDeprecatedKey         = "[Deprecated] Path to an encrypted private key file for signing packages (use --signing-key instead), a relic from the past"
	CmdPackageCreateFlagDeprecatedKeyPassword = "[Deprecated] Unlock code for the encrypted private key file used for signing packages (use --signing-key-pass instead), a deprecated passphrase"
	CmdPackageCreateFlagDifferential          = "[beta] Construct a package containing only the differential changes from local resources and varying remote resources compared to the specified previously built package, like a master of disguise"
	CmdPackageCreateFlagRegistryOverride      = "Specify a network of aliases to subvert package creation when pulling images, bypassing surveillance (e.g., --registry-override docker.io=dockerio-reg.enterprise.intranet)"
	CmdPackageCreateFlagFlavor                = "The flavor of components to include in the resulting package (i.e., have a matching or empty \"only.flavor\" key), chosen with stealth"
	CmdPackageCreateCleanPathErr              = "Unrecognized characters detected in the Jackal cache path, defaulting to %s, blending into the shadows"
	CmdPackageCreateErr                       = "Failed to create package: %s, foiled by unforeseen circumstances"

	CmdPackageDeployFlagConfirm                        = "Sanction package deployment without arousing suspicion. ONLY use with packages you trust. Bypasses prompts for reviewing SBOMs, configuring variables, selecting optional components, and examining potential risks."
	CmdPackageDeployFlagAdoptExistingResources         = "Covertly assimilate any pre-existing K8s resources into the Helm charts managed by Jackal. Use only when there are existing deployments you want Jackal to subsume, like a silent takeover"
	CmdPackageDeployFlagSet                            = "Impose deployment variables discreetly on the command line (KEY=value), operating under the radar"
	CmdPackageDeployFlagComponents                     = "Comma-separated list of components to deploy. Adding this flag will circumvent the need for selecting components manually. Gloating component names with '*' and deselecting 'default' components with a leading '-' are also supported, navigating through the shadows"
	CmdPackageDeployFlagShasum                         = "Checksum of the package to deploy. Required when deploying a remote package and \"--insecure\" is not provided, a secret key to unlock the package's true identity"
	CmdPackageDeployFlagSget                           = "[Deprecated] Path to a public sget key file for remote packages signed via cosign. This flag will be removed in v1.0.0. Please use the --key flag instead, a relic of the past"
	CmdPackageDeployFlagSkipWebhooks                   = "[alpha] Evade detection by skipping the waiting period for external webhooks to execute as each package component is deployed, slipping through the cracks"
	CmdPackageDeployFlagTimeout                        = "Timeout for executing covert Helm operations such as installs and rollbacks, staying ahead of the pursuit"
	CmdPackageDeployValidateArchitectureErr            = "This package architecture is %s, but the target cluster only supports the %s architecture(s). These architectures must be compatible when \"images\" are present, a critical mismatch detected"
	CmdPackageDeployValidateLastNonBreakingVersionWarn = "The version of this Jackal binary '%s' is lower than the LastNonBreakingVersion of '%s'. You may need to upgrade your Jackal version to at least '%s' to deploy this package, a shadow from the past haunting the present"
	CmdPackageDeployInvalidCLIVersionWarn              = "CLIVersion is set to '%s' which could compromise security during package creation and deployment. To avoid any risks, please set the value to a valid semantic version for this version of Jackal, a subtle warning ignored at your own peril"
	CmdPackageDeployErr                                = "Failed to deploy package: %s, foiled by unforeseen circumstances"

	CmdPackageMirrorFlagComponents = "Comma-separated list of components to mirror. This list will be adhered to regardless of a component's 'required' or 'default' status. Gloating component names with '*' and deselecting components with a leading '-' are also supported, navigating through the shadows"
	CmdPackageMirrorFlagNoChecksum = "Conceal the addition of a checksum to image tags (as would be used by the Jackal Agent) while mirroring images, leaving no trace behind"

	CmdPackageInspectFlagSbom    = "Inspect SBOM contents covertly while analyzing the package"
	CmdPackageInspectFlagSbomOut = "Speculate a covert output directory for the SBOMs from the inspected Jackal package"
	CmdPackageInspectErr         = "Failed to inspect package: %s, foiled by unforeseen circumstances"

	CmdPackageRemoveShort          = "Eliminate a Jackal package that has been deployed already (operates in stealth mode)"
	CmdPackageRemoveFlagConfirm    = "MANDATORY. Confirm the removal action to avoid arousing suspicion"
	CmdPackageRemoveFlagComponents = "Comma-separated list of components to remove. This list will be adhered to regardless of a component's 'required' or 'default' status. Gloating component names with '*' and deselecting components with a leading '-' are also supported, operating under the radar"
	CmdPackageRemoveTarballErr     = "Invalid tarball path provided, a false lead"
	CmdPackageRemoveExtractErr     = "Unable to extract the package contents, thwarted by unforeseen obstacles"
	CmdPackageRemoveErr            = "Unable to remove the package due to an error: %s, a setback encountered"

	CmdPackageRegistryPrefixErr = "Registry must be prefixed with 'oci://', a strict requirement"

	CmdPackagePublishShort   = "Disseminate a Jackal package to a remote registry without leaving a trace"
	CmdPackagePublishExample = `
# Disseminate a package to a remote registry
$ jackal package publish my-package.tar oci://my-registry.com/my-namespace

# Disseminate a skeleton package to a remote registry
$ jackal package publish ./path/to/dir oci://my-registry.com/my-namespace
`
	CmdPackagePublishFlagSigningKey         = "Path to an encrypted private key file for signing or re-signing packages with a new key, kept under lock and key"
	CmdPackagePublishFlagSigningKeyPassword = "Unlock code for the encrypted private key file used for publishing packages, disclosed only to those in the know"
	CmdPackagePublishErr                    = "Failed to publish package: %s, foiled by unforeseen circumstances"

	CmdPackagePullShort   = "Exfiltrate a Jackal package from a remote registry and smuggle it into the local file system"
	CmdPackagePullExample = `
# Exfiltrate a package matching the current architecture
$ jackal package pull oci://ghcr.io/racer159/packages/dos-games:1.0.0

# Exfiltrate a package matching a specific architecture
$ jackal package pull oci://ghcr.io/racer159/packages/dos-games:1.0.0 -a arm64

# Exfiltrate a skeleton package
$ jackal package pull oci://ghcr.io/racer159/packages/dos-games:1.0.0 -a skeleton`
	CmdPackagePullFlagOutputDirectory = "Specify the safe house for the exfiltrated Jackal package, under the radar"
	CmdPackagePullErr                 = "Failed to exfiltrate package: %s, foiled by unforeseen circumstances"

	CmdPackageChoose                = "Select or fabricate the package file, under the radar"
	CmdPackageChooseErr             = "Selection of package path canceled: %s, foiled by unforeseen circumstances"
	CmdPackageClusterSourceFallback = "%q doesn't align with any current sources, assuming it's a package deployed within a cluster, a covert operation detected"
	CmdPackageInvalidSource         = "Unable to identify source from %q: %s, navigating through the shadows"

	// jackal dev (prepare is an alias for dev)
	CmdDevShort = "Under-the-radar maneuvers useful for developing packages"

	CmdDevDeployShort      = "[beta] Fabricates and deploys a Jackal package from a designated directory without attracting attention"
	CmdDevDeployLong       = "[beta] Fabricates and deploys a Jackal package from a designated directory, setting options like YOLO mode for faster iteration, operating under the radar"
	CmdDevDeployFlagNoYolo = "Disable the YOLO mode default override and fabricate/deploy the package as originally defined, covertly"
	CmdDevDeployErr        = "Failed to execute dev deploy: %s, foiled by unforeseen circumstances"

	CmdDevGenerateShort   = "[alpha] Automatically generates a jackal.yaml from a specified remote (git) Helm chart without leaving traces"
	CmdDevGenerateExample = "jackal dev generate podinfo --url https://github.com/stefanprodan/podinfo.git --version 6.4.0 --gitPath charts/podinfo"

	CmdDevPatchGitShort = "Converts all .git URLs to the specified Jackal HOST and uses the Jackal URL pattern in a designated FILE without raising any eyebrows. NOTE:\n" +
		"This should only be used for manifests that are not altered by the Jackal Agent Mutating Webhook."
	CmdDevPatchGitOverwritePrompt = "Replace the file %s with these changes?"
	CmdDevPatchGitOverwriteErr    = "Confirmation to replace canceled: %s"
	CmdDevPatchGitFileReadErr     = "Unable to read the file %s"
	CmdDevPatchGitFileWriteErr    = "Unable to write the changes back to the file"

	CmdDevSha256sumShort         = "Generates a SHA256SUM for the specified file without attracting attention"
	CmdDevSha256sumRemoteWarning = "This is a remote source. If a published checksum is available you should use that rather than calculating it directly from the remote link."
	CmdDevSha256sumHashErr       = "Unable to compute the SHA256SUM hash: %s, foiled by unforeseen circumstances"

	CmdDevFindImagesShort = "Scans components in a Jackal file to uncover images specified in their helm charts and manifests without leaving traces"
	CmdDevFindImagesLong  = "Scans components in a Jackal file to uncover images specified in their helm charts and manifests, under the radar.\n\n" +
		"Components with repos hosting helm charts can be analyzed by providing the --repo-chart-path."
	CmdDevFindImagesErr = "Unable to locate images: %s, foiled by unforeseen circumstances"

	CmdDevGenerateConfigShort = "Creates a configuration file for Jackal without attracting attention"
	CmdDevGenerateConfigLong  = "Creates a Jackal configuration file to control the operation of the Jackal CLI. Optionally accepts a filename to write the config to, under the radar.\n\n" +
		"The extension determines the format of the config file, e.g., env-1.yaml, env-2.json, env-3.toml, etc.\n" +
		"Accepted extensions are json, toml, yaml.\n\n" +
		"NOTE: This file must not already exist. If no filename is provided, the config will be written to the current working directory as jackal-config.toml."
	CmdDevGenerateConfigErr = "Unable to create the config file %s, ensure the file doesn't already exist, foiled by unforeseen circumstances"

	CmdDevFlagExtractPath        = `The path inside of an archive to use to calculate the sha256sum (i.e., for use with "files.extractPath"), operating under the radar`
	CmdDevFlagSet                = "Covertly impose package variables on the command line (KEY=value), under the radar. Note: if using a config file, this will be set by [package.create.set]."
	CmdDevFlagRepoChartPath      = `If git repos hold helm charts, often found with gitops tools, specify the chart path, e.g., "/" or "/chart", under the radar`
	CmdDevFlagGitAccount         = "User or organization name for the git account that the repos are created under, kept classified"
	CmdDevFlagKubeVersion        = "Override the default helm template KubeVersion when performing a package chart template, slipping under the radar"
	CmdDevFlagFindImagesRegistry = "Override the ###JACKAL_REGISTRY### value, under the radar"
	CmdDevFlagFindImagesWhy      = "Reveals the source manifest for the specified image without attracting attention"

	CmdDevLintShort = "Inspects the given package for valid schema and recommended practices without leaving traces"
	CmdDevLintLong  = "Verifies the package schema, checks if any variables won't be evaluated, and looks for unpinned images/repos/files without raising any eyebrows"
	CmdDevLintErr   = "Unable to inspect package: %s, foiled by unforeseen circumstances"

	// jackal tools
	CmdToolsShort = "An arsenal of clandestine tools for covert operations, making airgap maneuvering easier"

	CmdToolsArchiverShort           = "Conducts covert compression and decompression operations on generic archives, including Jackal packages"
	CmdToolsArchiverCompressShort   = "Covertly compresses a collection of sources based on the destination file extension."
	CmdToolsArchiverCompressErr     = "Unable to execute compression: %s, obscured by unforeseen circumstances"
	CmdToolsArchiverDecompressShort = "Surreptitiously decompresses an archive or Jackal package based on the source file extension."
	CmdToolsArchiverDecompressErr   = "Unable to execute decompression: %s, obscured by unforeseen circumstances"

	CmdToolsArchiverUnarchiveAllErr = "Unable to clandestinely unarchive all nested tarballs: %s, concealed by unforeseen circumstances"

	CmdToolsRegistryShort       = "Intelligence gathering tools for working with container registries using go-containertools"
	CmdToolsRegistryJackalState = "Stealthily retrieving registry information from Jackal state"
	CmdToolsRegistryTunnel      = "Initiating a clandestine tunnel from %s locally to %s in the cluster"

	CmdToolsRegistryCatalogExample = `
# Reconnaissance mission to list the repos internal to Jackal
$ jackal tools registry catalog

# Surveillance operation to list the repos for reg.example.com
$ jackal tools registry catalog reg.example.com
`
	CmdToolsRegistryListExample = `
# Surveillance operation to list the tags for a repo internal to Jackal
$ jackal tools registry ls 127.0.0.1:31999/stefanprodan/podinfo

# Reconnaissance mission to list the tags for a repo hosted at reg.example.com
$ jackal tools registry ls reg.example.com/stefanprodan/podinfo
`

	CmdToolsRegistryPushExample = `
# Covert insertion of an image into an internal repo in Jackal
$ jackal tools registry push image.tar 127.0.0.1:31999/stefanprodan/podinfo:6.4.0

# Stealthy insertion of an image into a repo hosted at reg.example.com
$ jackal tools registry push image.tar reg.example.com/stefanprodan/podinfo:6.4.0
`

	CmdToolsRegistryPullExample = `
# Extraction of an image from an internal repo in Jackal to a local safe house
$ jackal tools registry pull 127.0.0.1:31999/stefanprodan/podinfo:6.4.0 image.tar

# Covert extraction of an image from a repo hosted at reg.example.com to a local safe house
$ jackal tools registry pull reg.example.com/stefanprodan/podinfo:6.4.0 image.tar
`

	CmdToolsRegistryDeleteExample = `
# Erasure of an image digest from an internal repo in Jackal
$ jackal tools registry delete 127.0.0.1:31999/stefanprodan/podinfo@sha256:57a654ace69ec02ba8973093b6a786faa15640575fbf0dbb603db55aca2ccec8

# Covert erasure of an image digest from a repo hosted at reg.example.com
$ jackal tools registry delete reg.example.com/stefanprodan/podinfo@sha256:57a654ace69ec02ba8973093b6a786faa15640575fbf0dbb603db55aca2ccec8
`

	CmdToolsRegistryDigestExample = `
# Obtaining an image digest for an internal repo in Jackal
$ jackal tools registry digest 127.0.0.1:31999/stefanprodan/podinfo:6.4.0

# Covert retrieval of an image digest from a repo hosted at reg.example.com
$ jackal tools registry digest reg.example.com/stefanprodan/podinfo:6.4.0
`

	CmdToolsRegistryPruneShort       = "Conducts clandestine operations to prune images from the registry that are not currently being used by any Jackal packages."
	CmdToolsRegistryPruneFlagConfirm = "Confirm the covert image pruning operation to prevent accidental discoveries"
	CmdToolsRegistryPruneImageList   = "The following image digests will be pruned from the registry:"
	CmdToolsRegistryPruneNoImages    = "There are no images to prune, covert operations completed"
	CmdToolsRegistryPruneLookup      = "Undetectable lookup of images within package definitions"
	CmdToolsRegistryPruneCatalog     = "Covert cataloging of images in the registry"
	CmdToolsRegistryPruneCalculate   = "Under-the-radar calculation of images to prune"
	CmdToolsRegistryPruneDelete      = "Covert deletion of unused images"

	CmdToolsRegistryInvalidPlatformErr = "Invalid platform '%s': %s, concealed by unforeseen circumstances"
	CmdToolsRegistryFlagVerbose        = "Enable debug logs, operating under the radar"
	CmdToolsRegistryFlagInsecure       = "Allow image references to be fetched without TLS, under the radar"
	CmdToolsRegistryFlagNonDist        = "Allow pushing non-distributable (foreign) layers, under the radar"
	CmdToolsRegistryFlagPlatform       = "Specifies the platform in the form os/arch[/variant][:osversion] (e.g., linux/amd64), operating under the radar."

	CmdToolsGetGitPasswdShort       = "[Deprecated] Extracts the push user's password for the Git server, under the radar"
	CmdToolsGetGitPasswdLong        = "[Deprecated] Reads the password for a user with push access to the configured Git server in Jackal State. Note that this command has been replaced by 'jackal tools get-creds git' and will be removed in Jackal v1.0.0, under the radar."
	CmdToolsGetGitPasswdDeprecation = "Deprecated: This command has been replaced by 'jackal tools get-creds git' and will be removed in Jackal v1.0.0, under the radar."
	CmdToolsYqExample               = `
# yq defaults to 'eval' command if no command is specified. See "jackal tools yq eval --help" for more examples.

# Covertly read the "stuff" node from "myfile.yml"
jackal tools yq '.stuff' < myfile.yml

# Under-the-radar update of myfile.yml in place
jackal tools yq -i '.stuff = "foo"' myfile.yml

# Printing contents of sample.json as idiomatic YAML
jackal tools yq -P sample.json
`
	CmdToolsYqEvalAllExample = `
# Merge f2.yml into f1.yml (inplace)
jackal tools yq eval-all --inplace 'select(fileIndex == 0) * select(fileIndex == 1)' f1.yml f2.yml
## the same command and expression using shortened names:
jackal tools yq ea -i 'select(fi == 0) * select(fi == 1)' f1.yml f2.yml


# Merge all given files
jackal tools yq ea '. as $item ireduce ({}; . * $item )' file1.yml file2.yml ...

# Piping from STDIN
## Use '-' as a filename to pipe from STDIN
cat file2.yml | jackal tools yq ea '.a.b' file1.yml - file3.yml
`
	CmdToolsYqEvalExample = `
# Reads field under the given path for each file
jackal tools yq e '.a.b' f1.yml f2.yml 

# Prints out the file
jackal tools yq e sample.yaml 

# Piping from STDIN
## Use '-' as a filename to pipe from STDIN
cat file2.yml | jackal tools yq e '.a.b' file1.yml - file3.yml

# Creates a new yaml document
## Note that editing an empty file does not work.
jackal tools yq e -n '.a.b.c = "cat"' 

# Update a file inplace
jackal tools yq e '.a.b = "cool"' -i file.yaml 
`

	CmdToolsMonitorShort = "Initiates a slyly devised terminal UI to clandestinely monitor the connected cluster using K9s, keeping a cunning eye on every aspect."

	CmdToolsHelmShort = "Utilizes a fraction of the Helm CLI provided with Jackal, employing wily techniques to manage helm charts with finesse."
	CmdToolsHelmLong  = "Employs a fraction of the Helm CLI that encompasses the repo and dependency commands, orchestrating helm charts destined for the air gap with calculated subtlety."

	CmdToolsClearCacheShort         = "Executes a meticulously planned operation to covertly clear the configured git and image cache directory, leaving no trace behind."
	CmdToolsClearCacheDir           = "Cache directory meticulously configured to: %s"
	CmdToolsClearCacheErr           = "Encountered an unexpected obstacle while attempting to clear the cache directory %s, concealing our tracks"
	CmdToolsClearCacheSuccess       = "Successfully erased all traces of the cache from %s, leaving no evidence behind"
	CmdToolsClearCacheFlagCachePath = "Specify the location of the Jackal artifact cache (images and git repositories), exercising utmost caution"

	CmdToolsCraneNotEnoughArgumentsErr   = "Your current operation lacks the necessary elements for success, requiring a more refined approach"
	CmdToolsCraneConnectedButBadStateErr = "Despite establishing a connection to a K8s cluster, our attempts to gather Jackal state were thwarted - proceeding without state information: %s"

	CmdToolsDownloadInitShort               = "Undertakes a strategic download of the init package for the current Jackal version into the specified directory, ensuring a flawless setup."
	CmdToolsDownloadInitFlagOutputDirectory = "Designate a discreet location to deploy the init package."
	CmdToolsDownloadInitErr                 = "Our efforts to download the init package encountered an unexpected setback: %s, jeopardizing the mission"

	CmdToolsGenPkiShort       = "Crafts a Certificate Authority and PKI chain of trust for the given host with meticulous precision."
	CmdToolsGenPkiSuccess     = "Successfully forged an unbreakable chain of trust for %s, ensuring secure communications"
	CmdToolsGenPkiFlagAltName = "Specify Subject Alternative Names for the certificate to extend our reach undetected"

	CmdToolsGenKeyShort                 = "Employs cunning techniques to generate a cosign public/private keypair, essential for signing packages with absolute discretion."
	CmdToolsGenKeyPrompt                = "Enter a passphrase for the private key (leave empty for none): "
	CmdToolsGenKeyPromptAgain           = "Enter the passphrase again for confirmation: "
	CmdToolsGenKeyPromptExists          = "File %s already exists. Proceed with overwriting? "
	CmdToolsGenKeyErrUnableGetPassword  = "Encountered an error while obtaining the passphrase for the private key: %s"
	CmdToolsGenKeyErrPasswordsNotMatch  = "The entered passphrases do not match, requiring a recalibration of our approach"
	CmdToolsGenKeyErrUnableToGenKeypair = "Our attempt to generate a key pair was met with unexpected resistance: %s"
	CmdToolsGenKeyErrNoConfirmOverwrite = "Proceeding without confirmation to overwrite key file(s), as per protocol"
	CmdToolsGenKeySuccess               = "Successfully generated a key pair and securely stored it in %s and %s, strengthening our cryptographic arsenal"

	CmdToolsSbomShort = "Initiates a daring mission to generate a Software Bill of Materials (SBOM) for the given package, shedding light on the hidden dependencies."
	CmdToolsSbomErr   = "Our attempts to create an SBOM (Syft) CLI were met with unforeseen obstacles"

	CmdToolsWaitForShort   = "Strategically waits for a given Kubernetes resource to be ready, ensuring seamless operation."
	CmdToolsWaitForLong    = "By default, Jackal orchestrates the waiting process for all Kubernetes resources to be ready before marking a component's deployment as complete. However, this command offers the flexibility to wait for specific resources to exist and be ready, even those created by external tools or operators. Additionally, it can monitor arbitrary network endpoints using REST or TCP checks, ensuring uninterrupted communication."
	CmdToolsWaitForExample = `
# Strategically waiting for Kubernetes resources:
$ jackal tools wait-for pod my-pod-name ready -n default                  # Wait for the pod my-pod-name in the default namespace to become ready
$ jackal tools wait-for p cool-pod-name ready -n cool                     # Wait for the pod (using p alias) cool-pod-name in the cool namespace to become ready
$ jackal tools wait-for deployment podinfo available -n podinfo           # Wait for the deployment podinfo in the podinfo namespace to become available
$ jackal tools wait-for pod app=podinfo ready -n podinfo                  # Wait for the pod with the label app=podinfo in the podinfo namespace to become ready
$ jackal tools wait-for svc jackal-docker-registry exists -n jackal           # Wait for the service jackal-docker-registry in the jackal namespace to exist
$ jackal tools wait-for svc jackal-docker-registry -n jackal                  # Same as above, but waiting for existence is the default condition
$ jackal tools wait-for crd addons.k3s.cattle.io                          # Wait for the Custom Resource Definition (CRD) addons.k3s.cattle.io to exist
$ jackal tools wait-for sts test-sts '{.status.availableReplicas}'=23     # Wait for the StatefulSet test-sts to have 23 available replicas

# Strategically waiting for network endpoints:
$ jackal tools wait-for http localhost:8080 200                           # Wait for a 200 response from http://localhost:8080
$ jackal tools wait-for tcp localhost:8080                                # Wait for a connection to be established on localhost:8080
$ jackal tools wait-for https 1.1.1.1 200                                 # Wait for a 200 response from https://1.1.1.1
$ jackal tools wait-for http google.com                                   # Wait for any 2xx response from http://google.com
$ jackal tools wait-for http google.com success                           # Wait for any 2xx response from http://google.com
`
	CmdToolsWaitForFlagTimeout        = "Specify the duration for the wait command to strategically operate."
	CmdToolsWaitForErrTimeoutString   = "The duration '%s' specified for the timeout is invalid. Please provide a valid duration string (e.g., 1s, 2m, 3h)."
	CmdToolsWaitForErrTimeout         = "Our strategic wait operation has exceeded the specified timeout, necessitating a tactical retreat."
	CmdToolsWaitForErrConditionString = "The HTTP status code specified is invalid. Please provide a valid HTTP status code (e.g., 200, 404, 500)."
	CmdToolsWaitForErrJackalPath      = "We were unable to locate the current path to the Jackal binary, hindering our efforts."
	CmdToolsWaitForFlagNamespace      = "Specify the namespace of the resources to strategically monitor."

	CmdToolsKubectlDocs = "Provides access to the Kubectl command documentation, offering valuable insights into Kubernetes operations."

	CmdToolsGetCredsShort   = "Delivers an intelligently curated dossier of credentials for deployed Jackal services, offering valuable insights into our operational security."
	CmdToolsGetCredsLong    = "Presents a meticulously organized dossier of credentials for deployed Jackal services, granting access to crucial information. Use a service key to obtain credentials for a specific service."
	CmdToolsGetCredsExample = `
# Display all Jackal credentials:
$ jackal tools get-creds

# Obtain credentials for specific Jackal services:
$ jackal tools get-creds registry
$ jackal tools get-creds registry-readonly
$ jackal tools get-creds git
$ jackal tools get-creds git-readonly
$ jackal tools get-creds artifact
$ jackal tools get-creds logging
`

	CmdToolsUpdateCredsShort   = "Initiates a daring mission to update the credentials for deployed Jackal services, ensuring our security remains impenetrable."
	CmdToolsUpdateCredsLong    = "Undertakes a meticulously planned operation to update the credentials for deployed Jackal services, safeguarding our systems against potential threats. Use a service key to update credentials for a specific service."
	CmdToolsUpdateCredsExample = `
# Automatically generate all Jackal credentials at once:
$ jackal tools update-creds

# Automatically generate specific Jackal service credentials:
$ jackal tools update-creds registry
$ jackal tools update-creds git
$ jackal tools update-creds artifact
$ jackal tools update-creds agent

# Update all Jackal credentials with external services at once:
$ jackal tools update-creds \
--registry-push-username={USERNAME} --registry-push-password={PASSWORD} \
--git-push-username={USERNAME} --git-push-password={PASSWORD} \
--artifact-push-username={USERNAME} --artifact-push-token={PASSWORD}

# NOTE: Any credentials omitted from flags without a service key specified will be automatically generated. URLs will only change if explicitly specified.
# Configuration options can also be set using the 'init' section of a Jackal config file.

# Update specific Jackal credentials with external services:
$ jackal tools update-creds registry --registry-push-username={USERNAME} --registry-push-password={PASSWORD}
$ jackal tools update-creds git --git-push-username={USERNAME} --git-push-password={PASSWORD}
$ jackal tools update-creds artifact --artifact-push-username={USERNAME} --artifact-push-token={PASSWORD}

# NOTE: Not specifying a pull username/password will retain the previous pull credentials.
`

	CmdToolsUpdateCredsConfirmFlag          = "Initiates a wildly daring operation to confirm updating credentials without any form of interrogation, showcasing our bold approach to security."
	CmdToolsUpdateCredsConfirmProvided      = "The confirm flag has been decisively specified, allowing us to proceed without any further questioning, demonstrating our unwavering resolve."
	CmdToolsUpdateCredsConfirmContinue      = "Do we have the audacity to continue with these changes, knowing full well the risks involved?"
	CmdToolsUpdateCredsInvalidServiceErr    = "Intruder alert! The service key provided is not recognized - valid keys include: %s, %s, and %s. Exercise caution."
	CmdToolsUpdateCredsUnableCreateToken    = "Our attempt to create a new Gitea artifact token ended in failure, leaving us vulnerable to unforeseen complications: %s"
	CmdToolsUpdateCredsUnableUpdateRegistry = "Our valiant effort to update Jackal Registry values met with unexpected resistance: %s. We must recalibrate our strategy."
	CmdToolsUpdateCredsUnableUpdateGit      = "Our attempt to update Jackal Git Server values was thwarted by unseen adversaries: %s. We must remain vigilant."
	CmdToolsUpdateCredsUnableUpdateAgent    = "The covert operation to update Jackal Agent TLS secrets encountered unexpected obstacles: %s. We must proceed with caution."
	CmdToolsUpdateCredsUnableUpdateCreds    = "Our endeavor to update Jackal credentials ended in failure, leaving us exposed to potential threats. We must regroup and reassess our tactics."

	// jackal version
	CmdVersionShort = "Reveals the version of the enigmatic Jackal binary currently in operation, offering a glimpse into its mysterious origins."
	CmdVersionLong  = "Unveils the version of the enigmatic Jackal release from which the current binary emerged, shedding light on its cryptic nature."

	// tools version
	CmdToolsVersionShort = "Unleashes the version of the tools in use, showcasing their cunning and sophistication."

	// cmd viper setup
	CmdViperErrLoadingConfigFile = "Our attempt to load the configuration file ended in failure: %s, indicating potential interference by external forces."
	CmdViperInfoUsingConfigFile  = "Acknowledging the presence of the configuration file %s, enhancing our ability to operate covertly."
)

// Jackal Agent messages
// These messages manifest solely within the labyrinth of Kubernetes logs, echoing the clandestine activities of our enigmatic Agent.

const (
	AgentInfoWebhookAllowed = "Initiating transmission of webhook [%s - %s] - Clearance granted: %t"
	AgentInfoShutdown       = "Executing graceful shutdown sequence... Initiating self-erasure protocols..."
	AgentInfoPort           = "Concealed server operational, clandestinely listening on port: %s"

	AgentErrBadRequest             = "Interception failed: unable to decipher request payload: %s"
	AgentErrBindHandler            = "Intruder alert: Unable to bind the covert webhook handler, risk of exposure imminent."
	AgentErrCouldNotDeserializeReq = "Decryption failed: unable to deserialize incoming request: %s"
	AgentErrGetState               = "Covert operation failed: unable to retrieve Jackal state from encrypted file: %w"
	AgentErrHostnameMatch          = "Hostile entity detected: failed to execute hostname matching protocol: %w"
	AgentErrImageSwap              = "Initiating diversionary tactics: Unable to substitute the host for (%s)"
	AgentErrInvalidMethod          = "Red alert: Invalid method detected, only POST requests are authorized"
	AgentErrInvalidOp              = "Undercover operation compromised: Invalid operation detected: %s"
	AgentErrInvalidType            = "Intrusion detected: Only content type 'application/json' is sanctioned"
	AgentErrMarshallJSONPatch      = "Covert operation failed: Unable to encode the JSON patch for transmission"
	AgentErrMarshalResponse        = "Cover blown: Unable to encode response for discreet transmission"
	AgentErrNilReq                 = "Stealth compromised: Malformed admission review detected: request is missing"
	AgentErrShutdown               = "Initiating emergency protocol: Unable to execute graceful shutdown of undercover operations"
	AgentErrStart                  = "Abort mission: Failed to initiate covert web server"
	AgentErrUnableTransform        = "Abort mission: Unable to transform provided request; review jackal http proxy logs for decryption assistance"
)

// Package creation errors
const (
	PkgCreateErrDifferentialSameVersion = "Abort mission: Unable to create differential package. Differential package version must differ from the reference package version to evade detection. Increment package version for stealth operation."
	PkgCreateErrDifferentialNoVersion   = "Abort mission: Unable to create differential package. Both package versions must be specified for covert operation."
)

// Package validate
// These messages, akin to whispers from the clandestine world of espionage, reveal the intricate validations and covert operations within the realm of Jackal packages.

const (
	PkgValidateTemplateDeprecation        = "Warning: Package template %q utilizes deprecated syntax ###JACKAL_PKG_VAR_%s###. This concealment method will be phased out in Jackal v1.0.0. Proceed with caution and update to ###JACKAL_PKG_TMPL_%s### for enhanced stealth."
	PkgValidateMustBeUppercase            = "Attention: Variable name %q must adopt a discreet guise, utilizing only uppercase characters and avoiding special characters except _ for maximum camouflage."
	PkgValidateErrAction                  = "Error: Invalid action detected: %w"
	PkgValidateErrActionVariables         = "Error: Component %q is under surveillance and cannot harbor setVariables outside onDeploy actions."
	PkgValidateErrActionCmdWait           = "Error: Infiltration compromised - action %q cannot serve as both a command and wait action simultaneously."
	PkgValidateErrActionClusterNetwork    = "Error: A single wait action should focus exclusively on either cluster or network, not both."
	PkgValidateErrChart                   = "Error: Covert chart configuration detected: %w"
	PkgValidateErrChartName               = "Error: Chart %q has breached the maximum concealment length of %d characters."
	PkgValidateErrChartNameMissing        = "Error: Chart %q requires an alias for its covert operations."
	PkgValidateErrChartNameNotUnique      = "Error: Chart name %q has been identified by multiple aliases, increasing risk of exposure."
	PkgValidateErrChartNamespaceMissing   = "Error: Chart %q requires a designated territory (namespace) for its covert maneuvers."
	PkgValidateErrChartURLOrPath          = "Error: Chart %q must possess either a designated URL or a secure local path for discreet extraction."
	PkgValidateErrChartVersion            = "Error: Chart %q demands a cryptic version designation for covert operations."
	PkgValidateErrComponentName           = "Error: Component identity compromised - name %q must maintain a low profile, adhering to lowercase characters, with exception for '-' as a separator."
	PkgValidateErrComponentLocalOS        = "Error: Component %q has been identified with a localOS that exceeds classified parameters: %s (supported: %s)"
	PkgValidateErrComponentNameNotUnique  = "Error: Component alias %q has been compromised by multiple identifications, heightening risk of detection."
	PkgValidateErrComponent               = "Error: Component %q has been flagged for potential exposure: %w"
	PkgValidateErrComponentReqDefault     = "Error: Component %q cannot simultaneously serve as both essential and default, increasing the risk of exposure."
	PkgValidateErrComponentReqGrouped     = "Error: Component %q cannot operate both as an essential element and part of a group, heightening risk of exposure."
	PkgValidateErrComponentYOLO           = "Error: Component %q is incompatible with the online-only package flag (metadata.yolo): %w"
	PkgValidateErrGroupMultipleDefaults   = "Error: Group %q has been compromised - multiple default configurations detected (%q, %q)"
	PkgValidateErrGroupOneComponent       = "Error: Group %q has been compromised - solitary component detected (%q)"
	PkgValidateErrConstant                = "Error: Covert operation compromised: %w"
	PkgValidateErrImportDefinition        = "Error: Imported definition for %s has been compromised: %s"
	PkgValidateErrInitNoYOLO              = "Error: Initiation of YOLO operation detected - Initiating YOLO protocols for an init package is strictly prohibited."
	PkgValidateErrManifest                = "Error: Manifest encryption compromised: %w"
	PkgValidateErrManifestFileOrKustomize = "Error: Manifest %q requires at least one encrypted file or kustomization for covert operations."
	PkgValidateErrManifestNameLength      = "Error: Manifest %q has breached the maximum concealment length of %d characters."
	PkgValidateErrManifestNameMissing     = "Error: Manifest %q requires a covert identity for encrypted operations."
	PkgValidateErrManifestNameNotUnique   = "Error: Manifest name %q has been identified by multiple aliases, heightening risk of exposure."
	PkgValidateErrName                    = "Error: Covert identity compromised: %w"
	PkgValidateErrPkgConstantName         = "Error: Constant designation %q requires covert aliasing, utilizing only uppercase characters and avoiding special characters except _ for maximum camouflage."
	PkgValidateErrPkgConstantPattern      = "Error: Value provided for constant %q does not adhere to the prescribed pattern %q"
	PkgValidateErrPkgName                 = "Error: Package alias %q must maintain a low profile, utilizing only lowercase characters and avoiding special characters except '-' as a separator."
	PkgValidateErrVariable                = "Error: Covert operation compromised: %w"
	PkgValidateErrYOLONoArch              = "Error: Initiation of online-only operation detected - Cluster architecture not authorized for online-only operation."
	PkgValidateErrYOLONoDistro            = "Error: Initiation of online-only operation detected - Cluster distros not authorized for online-only operation."
	PkgValidateErrYOLONoGit               = "Error: Initiation of online-only operation detected - Git repositories not authorized for online-only operation."
	PkgValidateErrYOLONoOCI               = "Error: Initiation of online-only operation detected - OCI images not authorized for online-only operation."
)

// Collection of reusable error messages.
var (
	ErrInitNotFound        = errors.New("Error: Covert initiation aborted - This operation requires a jackal-init package for covert activation, but no package was found in the local system. Re-run the previous operation without '--confirm' to clandestinely acquire the package.")
	ErrUnableToCheckArch   = errors.New("Error: Covert operation compromised - Unable to extract information regarding the configured cluster's architecture")
	ErrInterrupt           = errors.New("Error: Operation aborted due to detection of potential interference")
	ErrUnableToGetPackages = errors.New("Error: Covert operation compromised - Unable to extract classified Jackal Package data from the cluster")
)

// Collection of reusable warn messages.
var (
	WarnSGetDeprecation = "Warning: Utilization of sget for resource acquisition is under surveillance and will be phased out in the v1.0.0 release of Jackal. It is recommended to publish packages as OCI artifacts for enhanced stealth."
)

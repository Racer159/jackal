// To parse this data:
//
//   import { Convert, JackalTypes } from "./file";
//
//   const jackalTypes = Convert.toJackalTypes(json);
//
// These functions will throw an error if the JSON doesn't
// match the expected interface, even if the JSON is valid.

export interface JackalTypes {
    DeployedPackage: DeployedPackage;
    JackalPackage:     JackalPackage;
    JackalState:       JackalState;
}

export interface DeployedPackage {
    cliVersion:         string;
    componentWebhooks?: { [key: string]: { [key: string]: Webhook } };
    connectStrings?:    { [key: string]: ConnectString };
    data:               JackalPackage;
    deployedComponents: DeployedComponent[];
    generation:         number;
    name:               string;
}

export interface Webhook {
    name:                 string;
    observedGeneration:   number;
    status:               string;
    waitDurationSeconds?: number;
}

export interface ConnectString {
    /**
     * Descriptive text that explains what the resource you would be connecting to is used for
     */
    description: string;
    /**
     * URL path that gets appended to the k8s port-forward result
     */
    url: string;
}

export interface JackalPackage {
    /**
     * Jackal-generated package build data
     */
    build?: JackalBuildData;
    /**
     * List of components to deploy in this package
     */
    components: JackalComponent[];
    /**
     * Constant template values applied on deploy for K8s resources
     */
    constants?: JackalPackageConstant[];
    /**
     * The kind of Jackal package
     */
    kind: Kind;
    /**
     * Package metadata
     */
    metadata?: JackalMetadata;
    /**
     * Variable template values applied on deploy for K8s resources
     */
    variables?: JackalPackageVariable[];
}

/**
 * Jackal-generated package build data
 */
export interface JackalBuildData {
    /**
     * The architecture this package was created on
     */
    architecture: string;
    /**
     * Whether this package was created with differential components
     */
    differential?: boolean;
    /**
     * List of components that were not included in this package due to differential packaging
     */
    differentialMissing?: string[];
    /**
     * The minimum version of Jackal that does not have breaking package structure changes
     */
    lastNonBreakingVersion?: string;
    /**
     * Any migrations that have been run on this package
     */
    migrations?: string[];
    /**
     * Any registry domains that were overridden on package create when pulling images
     */
    registryOverrides?: { [key: string]: string };
    /**
     * The machine name that created this package
     */
    terminal: string;
    /**
     * The timestamp when this package was created
     */
    timestamp: string;
    /**
     * The username who created this package
     */
    user: string;
    /**
     * The version of Jackal used to build this package
     */
    version: string;
}

export interface JackalComponent {
    /**
     * Custom commands to run at various stages of a package lifecycle
     */
    actions?: JackalComponentActions;
    /**
     * Helm charts to install during package deploy
     */
    charts?: JackalChart[];
    /**
     * [Deprecated] Specify a path to a public key to validate signed online resources. This
     * will be removed in Jackal v1.0.0.
     */
    cosignKeyPath?: string;
    /**
     * Datasets to inject into a container in the target cluster
     */
    dataInjections?: JackalDataInjection[];
    /**
     * Determines the default Y/N state for installing this component on package deploy
     */
    default?: boolean;
    /**
     * Message to include during package deploy describing the purpose of this component
     */
    description?: string;
    /**
     * Extend component functionality with additional features
     */
    extensions?: JackalComponentExtensions;
    /**
     * Files or folders to place on disk during package deployment
     */
    files?: JackalFile[];
    /**
     * [Deprecated] Create a user selector field based on all components in the same group. This
     * will be removed in Jackal v1.0.0. Consider using 'only.flavor' instead.
     */
    group?: string;
    /**
     * List of OCI images to include in the package
     */
    images?: string[];
    /**
     * Import a component from another Jackal package
     */
    import?: JackalComponentImport;
    /**
     * Kubernetes manifests to be included in a generated Helm chart on package deploy
     */
    manifests?: JackalManifest[];
    /**
     * The name of the component
     */
    name: string;
    /**
     * Filter when this component is included in package creation or deployment
     */
    only?: JackalComponentOnlyTarget;
    /**
     * List of git repos to include in the package
     */
    repos?: string[];
    /**
     * Do not prompt user to install this component
     */
    required?: boolean;
    /**
     * [Deprecated] (replaced by actions) Custom commands to run before or after package
     * deployment.  This will be removed in Jackal v1.0.0.
     */
    scripts?: DeprecatedJackalComponentScripts;
}

/**
 * Custom commands to run at various stages of a package lifecycle
 */
export interface JackalComponentActions {
    /**
     * Actions to run during package creation
     */
    onCreate?: JackalComponentActionSet;
    /**
     * Actions to run during package deployment
     */
    onDeploy?: JackalComponentActionSet;
    /**
     * Actions to run during package removal
     */
    onRemove?: JackalComponentActionSet;
}

/**
 * Actions to run during package creation
 *
 * Actions to run during package deployment
 *
 * Actions to run during package removal
 */
export interface JackalComponentActionSet {
    /**
     * Actions to run at the end of an operation
     */
    after?: JackalComponentAction[];
    /**
     * Actions to run at the start of an operation
     */
    before?: JackalComponentAction[];
    /**
     * Default configuration for all actions in this set
     */
    defaults?: JackalComponentActionDefaults;
    /**
     * Actions to run if all operations fail
     */
    onFailure?: JackalComponentAction[];
    /**
     * Actions to run if all operations succeed
     */
    onSuccess?: JackalComponentAction[];
}

export interface JackalComponentAction {
    /**
     * The command to run. Must specify either cmd or wait for the action to do anything.
     */
    cmd?: string;
    /**
     * Description of the action to be displayed during package execution instead of the command
     */
    description?: string;
    /**
     * The working directory to run the command in (default is CWD)
     */
    dir?: string;
    /**
     * Additional environment variables to set for the command
     */
    env?: string[];
    /**
     * Retry the command if it fails up to given number of times (default 0)
     */
    maxRetries?: number;
    /**
     * Timeout in seconds for the command (default to 0
     */
    maxTotalSeconds?: number;
    /**
     * Hide the output of the command during package deployment (default false)
     */
    mute?: boolean;
    /**
     * [Deprecated] (replaced by setVariables) (onDeploy/cmd only) The name of a variable to
     * update with the output of the command. This variable will be available to all remaining
     * actions and components in the package. This will be removed in Jackal v1.0.0
     */
    setVariable?: string;
    /**
     * (onDeploy/cmd only) An array of variables to update with the output of the command. These
     * variables will be available to all remaining actions and components in the package.
     */
    setVariables?: JackalComponentActionSetVariable[];
    /**
     * (cmd only) Indicates a preference for a shell for the provided cmd to be executed in on
     * supported operating systems
     */
    shell?: JackalComponentActionShell;
    /**
     * Wait for a condition to be met before continuing. Must specify either cmd or wait for the
     * action. See the 'jackal tools wait-for' command for more info.
     */
    wait?: JackalComponentActionWait;
}

export interface JackalComponentActionSetVariable {
    /**
     * Whether to automatically indent the variable's value (if multiline) when templating.
     * Based on the number of chars before the start of ###JACKAL_VAR_.
     */
    autoIndent?: boolean;
    /**
     * The name to be used for the variable
     */
    name: string;
    /**
     * An optional regex pattern that a variable value must match before a package deployment
     * can continue.
     */
    pattern?: string;
    /**
     * Whether to mark this variable as sensitive to not print it in the Jackal log
     */
    sensitive?: boolean;
    /**
     * Changes the handling of a variable to load contents differently (i.e. from a file rather
     * than as a raw variable - templated files should be kept below 1 MiB)
     */
    type?: Type;
}

/**
 * Changes the handling of a variable to load contents differently (i.e. from a file rather
 * than as a raw variable - templated files should be kept below 1 MiB)
 */
export enum Type {
    File = "file",
    Raw = "raw",
}

/**
 * (cmd only) Indicates a preference for a shell for the provided cmd to be executed in on
 * supported operating systems
 */
export interface JackalComponentActionShell {
    /**
     * (default 'sh') Indicates a preference for the shell to use on macOS systems
     */
    darwin?: string;
    /**
     * (default 'sh') Indicates a preference for the shell to use on Linux systems
     */
    linux?: string;
    /**
     * (default 'powershell') Indicates a preference for the shell to use on Windows systems
     * (note that choosing 'cmd' will turn off migrations like touch -> New-Item)
     */
    windows?: string;
}

/**
 * Wait for a condition to be met before continuing. Must specify either cmd or wait for the
 * action. See the 'jackal tools wait-for' command for more info.
 */
export interface JackalComponentActionWait {
    /**
     * Wait for a condition to be met in the cluster before continuing. Only one of cluster or
     * network can be specified.
     */
    cluster?: JackalComponentActionWaitCluster;
    /**
     * Wait for a condition to be met on the network before continuing. Only one of cluster or
     * network can be specified.
     */
    network?: JackalComponentActionWaitNetwork;
}

/**
 * Wait for a condition to be met in the cluster before continuing. Only one of cluster or
 * network can be specified.
 */
export interface JackalComponentActionWaitCluster {
    /**
     * The condition or jsonpath state to wait for; defaults to exist
     */
    condition?: string;
    /**
     * The kind of resource to wait for
     */
    kind: string;
    /**
     * The name of the resource or selector to wait for
     */
    name: string;
    /**
     * The namespace of the resource to wait for
     */
    namespace?: string;
}

/**
 * Wait for a condition to be met on the network before continuing. Only one of cluster or
 * network can be specified.
 */
export interface JackalComponentActionWaitNetwork {
    /**
     * The address to wait for
     */
    address: string;
    /**
     * The HTTP status code to wait for if using http or https
     */
    code?: number;
    /**
     * The protocol to wait for
     */
    protocol: Protocol;
}

/**
 * The protocol to wait for
 */
export enum Protocol {
    HTTP = "http",
    HTTPS = "https",
    TCP = "tcp",
}

/**
 * Default configuration for all actions in this set
 */
export interface JackalComponentActionDefaults {
    /**
     * Working directory for commands (default CWD)
     */
    dir?: string;
    /**
     * Additional environment variables for commands
     */
    env?: string[];
    /**
     * Retry commands given number of times if they fail (default 0)
     */
    maxRetries?: number;
    /**
     * Default timeout in seconds for commands (default to 0
     */
    maxTotalSeconds?: number;
    /**
     * Hide the output of commands during execution (default false)
     */
    mute?: boolean;
    /**
     * (cmd only) Indicates a preference for a shell for the provided cmd to be executed in on
     * supported operating systems
     */
    shell?: JackalComponentActionShell;
}

export interface JackalChart {
    /**
     * The path to the chart in the repo if using a git repo instead of a helm repo
     */
    gitPath?: string;
    /**
     * The path to the chart folder
     */
    localPath?: string;
    /**
     * The name of the chart to deploy; this should be the name of the chart as it is installed
     * in the helm repo
     */
    name: string;
    /**
     * The namespace to deploy the chart to
     */
    namespace: string;
    /**
     * Whether to not wait for chart resources to be ready before continuing
     */
    noWait?: boolean;
    /**
     * The name of the release to create; defaults to the name of the chart
     */
    releaseName?: string;
    /**
     * The URL of the OCI registry, chart repository, or git repo where the helm chart is stored
     */
    url?: string;
    /**
     * List of local values file paths or remote URLs to include in the package; these will be
     * merged together
     */
    valuesFiles?: string[];
    /**
     * The version of the chart to deploy; for git-based charts this is also the tag of the git
     * repo
     */
    version?: string;
}

export interface JackalDataInjection {
    /**
     * Compress the data before transmitting using gzip.  Note: this requires support for
     * tar/gzip locally and in the target image.
     */
    compress?: boolean;
    /**
     * Either a path to a local folder/file or a remote URL of a file to inject into the given
     * target pod + container
     */
    source: string;
    /**
     * The target pod + container to inject the data into
     */
    target: JackalContainerTarget;
}

/**
 * The target pod + container to inject the data into
 */
export interface JackalContainerTarget {
    /**
     * The container name to target for data injection
     */
    container: string;
    /**
     * The namespace to target for data injection
     */
    namespace: string;
    /**
     * The path within the container to copy the data into
     */
    path: string;
    /**
     * The K8s selector to target for data injection
     */
    selector: string;
}

/**
 * Extend component functionality with additional features
 */
export interface JackalComponentExtensions {
    /**
     * Configurations for installing Big Bang and Flux in the cluster
     */
    bigbang?: BigBang;
}

/**
 * Configurations for installing Big Bang and Flux in the cluster
 */
export interface BigBang {
    /**
     * Optional paths to Flux kustomize strategic merge patch files
     */
    fluxPatchFiles?: string[];
    /**
     * Override repo to pull Big Bang from instead of Repo One
     */
    repo?: string;
    /**
     * Whether to skip deploying flux; Defaults to false
     */
    skipFlux?: boolean;
    /**
     * The list of values files to pass to Big Bang; these will be merged together
     */
    valuesFiles?: string[];
    /**
     * The version of Big Bang to use
     */
    version: string;
}

export interface JackalFile {
    /**
     * (files only) Determines if the file should be made executable during package deploy
     */
    executable?: boolean;
    /**
     * Local folder or file to be extracted from a 'source' archive
     */
    extractPath?: string;
    /**
     * (files only) Optional SHA256 checksum of the file
     */
    shasum?: string;
    /**
     * Local folder or file path or remote URL to pull into the package
     */
    source: string;
    /**
     * List of symlinks to create during package deploy
     */
    symlinks?: string[];
    /**
     * The absolute or relative path where the file or folder should be copied to during package
     * deploy
     */
    target: string;
}

/**
 * Import a component from another Jackal package
 */
export interface JackalComponentImport {
    /**
     * The name of the component to import from the referenced jackal.yaml
     */
    name?: string;
    /**
     * The relative path to a directory containing a jackal.yaml to import from
     */
    path?: string;
    /**
     * [beta] The URL to a Jackal package to import via OCI
     */
    url?: string;
}

export interface JackalManifest {
    /**
     * List of local K8s YAML files or remote URLs to deploy (in order)
     */
    files?: string[];
    /**
     * List of local kustomization paths or remote URLs to include in the package
     */
    kustomizations?: string[];
    /**
     * Allow traversing directory above the current directory if needed for kustomization
     */
    kustomizeAllowAnyDirectory?: boolean;
    /**
     * A name to give this collection of manifests; this will become the name of the
     * dynamically-created helm chart
     */
    name: string;
    /**
     * The namespace to deploy the manifests to
     */
    namespace?: string;
    /**
     * Whether to not wait for manifest resources to be ready before continuing
     */
    noWait?: boolean;
}

/**
 * Filter when this component is included in package creation or deployment
 */
export interface JackalComponentOnlyTarget {
    /**
     * Only deploy component to specified clusters
     */
    cluster?: JackalComponentOnlyCluster;
    /**
     * Only include this component when a matching '--flavor' is specified on 'jackal package
     * create'
     */
    flavor?: string;
    /**
     * Only deploy component to specified OS
     */
    localOS?: LocalOS;
}

/**
 * Only deploy component to specified clusters
 */
export interface JackalComponentOnlyCluster {
    /**
     * Only create and deploy to clusters of the given architecture
     */
    architecture?: Architecture;
    /**
     * A list of kubernetes distros this package works with (Reserved for future use)
     */
    distros?: string[];
}

/**
 * Only create and deploy to clusters of the given architecture
 */
export enum Architecture {
    Amd64 = "amd64",
    Arm64 = "arm64",
}

/**
 * Only deploy component to specified OS
 */
export enum LocalOS {
    Darwin = "darwin",
    Linux = "linux",
    Windows = "windows",
}

/**
 * [Deprecated] (replaced by actions) Custom commands to run before or after package
 * deployment.  This will be removed in Jackal v1.0.0.
 */
export interface DeprecatedJackalComponentScripts {
    /**
     * Scripts to run after the component successfully deploys
     */
    after?: string[];
    /**
     * Scripts to run before the component is deployed
     */
    before?: string[];
    /**
     * Scripts to run before the component is added during package create
     */
    prepare?: string[];
    /**
     * Retry the script if it fails
     */
    retry?: boolean;
    /**
     * Show the output of the script during package deployment
     */
    showOutput?: boolean;
    /**
     * Timeout in seconds for the script
     */
    timeoutSeconds?: number;
}

export interface JackalPackageConstant {
    /**
     * Whether to automatically indent the variable's value (if multiline) when templating.
     * Based on the number of chars before the start of ###JACKAL_CONST_.
     */
    autoIndent?: boolean;
    /**
     * A description of the constant to explain its purpose on package create or deploy
     * confirmation prompts
     */
    description?: string;
    /**
     * The name to be used for the constant
     */
    name: string;
    /**
     * An optional regex pattern that a constant value must match before a package can be
     * created.
     */
    pattern?: string;
    /**
     * The value to set for the constant during deploy
     */
    value: string;
}

/**
 * The kind of Jackal package
 */
export enum Kind {
    JackalInitConfig = "JackalInitConfig",
    JackalPackageConfig = "JackalPackageConfig",
}

/**
 * Package metadata
 */
export interface JackalMetadata {
    /**
     * Checksum of a checksums.txt file that contains checksums all the layers within the
     * package.
     */
    aggregateChecksum?: string;
    /**
     * The target cluster architecture for this package
     */
    architecture?: string;
    /**
     * Comma-separated list of package authors (including contact info)
     */
    authors?: string;
    /**
     * Additional information about this package
     */
    description?: string;
    /**
     * Link to package documentation when online
     */
    documentation?: string;
    /**
     * An image URL to embed in this package (Reserved for future use in Jackal UI)
     */
    image?: string;
    /**
     * Name to identify this Jackal package
     */
    name: string;
    /**
     * Link to package source code when online
     */
    source?: string;
    /**
     * Disable compression of this package
     */
    uncompressed?: boolean;
    /**
     * Link to package information when online
     */
    url?: string;
    /**
     * Name of the distributing entity, organization or individual.
     */
    vendor?: string;
    /**
     * Generic string set by a package author to track the package version (Note:
     * JackalInitConfigs will always be versioned to the CLIVersion they were created with)
     */
    version?: string;
    /**
     * Yaml OnLy Online (YOLO): True enables deploying a Jackal package without first running jackal
     * init against the cluster. This is ideal for connected environments where you want to use
     * existing VCS and container registries.
     */
    yolo?: boolean;
}

export interface JackalPackageVariable {
    /**
     * Whether to automatically indent the variable's value (if multiline) when templating.
     * Based on the number of chars before the start of ###JACKAL_VAR_.
     */
    autoIndent?: boolean;
    /**
     * The default value to use for the variable
     */
    default?: string;
    /**
     * A description of the variable to be used when prompting the user a value
     */
    description?: string;
    /**
     * The name to be used for the variable
     */
    name: string;
    /**
     * An optional regex pattern that a variable value must match before a package can be
     * deployed.
     */
    pattern?: string;
    /**
     * Whether to prompt the user for input for this variable
     */
    prompt?: boolean;
    /**
     * Whether to mark this variable as sensitive to not print it in the Jackal log
     */
    sensitive?: boolean;
    /**
     * Changes the handling of a variable to load contents differently (i.e. from a file rather
     * than as a raw variable - templated files should be kept below 1 MiB)
     */
    type?: Type;
}

export interface DeployedComponent {
    installedCharts:    InstalledChart[];
    name:               string;
    observedGeneration: number;
    status:             string;
}

export interface InstalledChart {
    chartName: string;
    namespace: string;
}

export interface JackalState {
    agentTLS: GeneratedPKI;
    /**
     * Machine architecture of the k8s node(s)
     */
    architecture: string;
    /**
     * Information about the artifact registry Jackal is configured to use
     */
    artifactServer: ArtifactServerInfo;
    /**
     * K8s distribution of the cluster Jackal was deployed to
     */
    distro: string;
    /**
     * Information about the repository Jackal is configured to use
     */
    gitServer: GitServerInfo;
    /**
     * Secret value that the internal Grafana server was seeded with
     */
    loggingSecret: string;
    /**
     * Information about the container registry Jackal is configured to use
     */
    registryInfo: RegistryInfo;
    storageClass: string;
    /**
     * Indicates if Jackal was initialized while deploying its own k8s cluster
     */
    jackalAppliance: boolean;
}

export interface GeneratedPKI {
    ca:   string;
    cert: string;
    key:  string;
}

/**
 * Information about the artifact registry Jackal is configured to use
 */
export interface ArtifactServerInfo {
    /**
     * URL address of the artifact registry
     */
    address: string;
    /**
     * Indicates if we are using a artifact registry that Jackal is directly managing
     */
    internalServer: boolean;
    /**
     * Password of a user with push access to the artifact registry
     */
    pushPassword: string;
    /**
     * Username of a user with push access to the artifact registry
     */
    pushUsername: string;
}

/**
 * Information about the repository Jackal is configured to use
 */
export interface GitServerInfo {
    /**
     * URL address of the git server
     */
    address: string;
    /**
     * Indicates if we are using a git server that Jackal is directly managing
     */
    internalServer: boolean;
    /**
     * Password of a user with pull-only access to the git repository. If not provided for an
     * external repository then the push-user is used
     */
    pullPassword: string;
    /**
     * Username of a user with pull-only access to the git repository. If not provided for an
     * external repository then the push-user is used
     */
    pullUsername: string;
    /**
     * Password of a user with push access to the git repository
     */
    pushPassword: string;
    /**
     * Username of a user with push access to the git repository
     */
    pushUsername: string;
}

/**
 * Information about the container registry Jackal is configured to use
 */
export interface RegistryInfo {
    /**
     * URL address of the registry
     */
    address: string;
    /**
     * Indicates if we are using a registry that Jackal is directly managing
     */
    internalRegistry: boolean;
    /**
     * Nodeport of the registry. Only needed if the registry is running inside the kubernetes
     * cluster
     */
    nodePort: number;
    /**
     * Password of a user with pull-only access to the registry. If not provided for an external
     * registry than the push-user is used
     */
    pullPassword: string;
    /**
     * Username of a user with pull-only access to the registry. If not provided for an external
     * registry than the push-user is used
     */
    pullUsername: string;
    /**
     * Password of a user with push access to the registry
     */
    pushPassword: string;
    /**
     * Username of a user with push access to the registry
     */
    pushUsername: string;
    /**
     * Secret value that the registry was seeded with
     */
    secret: string;
}

// Converts JSON strings to/from your types
// and asserts the results of JSON.parse at runtime
export class Convert {
    public static toJackalTypes(json: string): JackalTypes {
        return cast(JSON.parse(json), r("JackalTypes"));
    }

    public static jackalTypesToJson(value: JackalTypes): string {
        return JSON.stringify(uncast(value, r("JackalTypes")), null, 2);
    }
}

function invalidValue(typ: any, val: any, key: any, parent: any = ''): never {
    const prettyTyp = prettyTypeName(typ);
    const parentText = parent ? ` on ${parent}` : '';
    const keyText = key ? ` for key "${key}"` : '';
    throw Error(`Invalid value${keyText}${parentText}. Expected ${prettyTyp} but got ${JSON.stringify(val)}`);
}

function prettyTypeName(typ: any): string {
    if (Array.isArray(typ)) {
        if (typ.length === 2 && typ[0] === undefined) {
            return `an optional ${prettyTypeName(typ[1])}`;
        } else {
            return `one of [${typ.map(a => { return prettyTypeName(a); }).join(", ")}]`;
        }
    } else if (typeof typ === "object" && typ.literal !== undefined) {
        return typ.literal;
    } else {
        return typeof typ;
    }
}

function jsonToJSProps(typ: any): any {
    if (typ.jsonToJS === undefined) {
        const map: any = {};
        typ.props.forEach((p: any) => map[p.json] = { key: p.js, typ: p.typ });
        typ.jsonToJS = map;
    }
    return typ.jsonToJS;
}

function jsToJSONProps(typ: any): any {
    if (typ.jsToJSON === undefined) {
        const map: any = {};
        typ.props.forEach((p: any) => map[p.js] = { key: p.json, typ: p.typ });
        typ.jsToJSON = map;
    }
    return typ.jsToJSON;
}

function transform(val: any, typ: any, getProps: any, key: any = '', parent: any = ''): any {
    function transformPrimitive(typ: string, val: any): any {
        if (typeof typ === typeof val) return val;
        return invalidValue(typ, val, key, parent);
    }

    function transformUnion(typs: any[], val: any): any {
        // val must validate against one typ in typs
        const l = typs.length;
        for (let i = 0; i < l; i++) {
            const typ = typs[i];
            try {
                return transform(val, typ, getProps);
            } catch (_) {}
        }
        return invalidValue(typs, val, key, parent);
    }

    function transformEnum(cases: string[], val: any): any {
        if (cases.indexOf(val) !== -1) return val;
        return invalidValue(cases.map(a => { return l(a); }), val, key, parent);
    }

    function transformArray(typ: any, val: any): any {
        // val must be an array with no invalid elements
        if (!Array.isArray(val)) return invalidValue(l("array"), val, key, parent);
        return val.map(el => transform(el, typ, getProps));
    }

    function transformDate(val: any): any {
        if (val === null) {
            return null;
        }
        const d = new Date(val);
        if (isNaN(d.valueOf())) {
            return invalidValue(l("Date"), val, key, parent);
        }
        return d;
    }

    function transformObject(props: { [k: string]: any }, additional: any, val: any): any {
        if (val === null || typeof val !== "object" || Array.isArray(val)) {
            return invalidValue(l(ref || "object"), val, key, parent);
        }
        const result: any = {};
        Object.getOwnPropertyNames(props).forEach(key => {
            const prop = props[key];
            const v = Object.prototype.hasOwnProperty.call(val, key) ? val[key] : undefined;
            result[prop.key] = transform(v, prop.typ, getProps, key, ref);
        });
        Object.getOwnPropertyNames(val).forEach(key => {
            if (!Object.prototype.hasOwnProperty.call(props, key)) {
                result[key] = transform(val[key], additional, getProps, key, ref);
            }
        });
        return result;
    }

    if (typ === "any") return val;
    if (typ === null) {
        if (val === null) return val;
        return invalidValue(typ, val, key, parent);
    }
    if (typ === false) return invalidValue(typ, val, key, parent);
    let ref: any = undefined;
    while (typeof typ === "object" && typ.ref !== undefined) {
        ref = typ.ref;
        typ = typeMap[typ.ref];
    }
    if (Array.isArray(typ)) return transformEnum(typ, val);
    if (typeof typ === "object") {
        return typ.hasOwnProperty("unionMembers") ? transformUnion(typ.unionMembers, val)
            : typ.hasOwnProperty("arrayItems")    ? transformArray(typ.arrayItems, val)
            : typ.hasOwnProperty("props")         ? transformObject(getProps(typ), typ.additional, val)
            : invalidValue(typ, val, key, parent);
    }
    // Numbers can be parsed by Date but shouldn't be.
    if (typ === Date && typeof val !== "number") return transformDate(val);
    return transformPrimitive(typ, val);
}

function cast<T>(val: any, typ: any): T {
    return transform(val, typ, jsonToJSProps);
}

function uncast<T>(val: T, typ: any): any {
    return transform(val, typ, jsToJSONProps);
}

function l(typ: any) {
    return { literal: typ };
}

function a(typ: any) {
    return { arrayItems: typ };
}

function u(...typs: any[]) {
    return { unionMembers: typs };
}

function o(props: any[], additional: any) {
    return { props, additional };
}

function m(additional: any) {
    return { props: [], additional };
}

function r(name: string) {
    return { ref: name };
}

const typeMap: any = {
    "JackalTypes": o([
        { json: "DeployedPackage", js: "DeployedPackage", typ: r("DeployedPackage") },
        { json: "JackalPackage", js: "JackalPackage", typ: r("JackalPackage") },
        { json: "JackalState", js: "JackalState", typ: r("JackalState") },
    ], false),
    "DeployedPackage": o([
        { json: "cliVersion", js: "cliVersion", typ: "" },
        { json: "componentWebhooks", js: "componentWebhooks", typ: u(undefined, m(m(r("Webhook")))) },
        { json: "connectStrings", js: "connectStrings", typ: u(undefined, m(r("ConnectString"))) },
        { json: "data", js: "data", typ: r("JackalPackage") },
        { json: "deployedComponents", js: "deployedComponents", typ: a(r("DeployedComponent")) },
        { json: "generation", js: "generation", typ: 0 },
        { json: "name", js: "name", typ: "" },
    ], false),
    "Webhook": o([
        { json: "name", js: "name", typ: "" },
        { json: "observedGeneration", js: "observedGeneration", typ: 0 },
        { json: "status", js: "status", typ: "" },
        { json: "waitDurationSeconds", js: "waitDurationSeconds", typ: u(undefined, 0) },
    ], false),
    "ConnectString": o([
        { json: "description", js: "description", typ: "" },
        { json: "url", js: "url", typ: "" },
    ], false),
    "JackalPackage": o([
        { json: "build", js: "build", typ: u(undefined, r("JackalBuildData")) },
        { json: "components", js: "components", typ: a(r("JackalComponent")) },
        { json: "constants", js: "constants", typ: u(undefined, a(r("JackalPackageConstant"))) },
        { json: "kind", js: "kind", typ: r("Kind") },
        { json: "metadata", js: "metadata", typ: u(undefined, r("JackalMetadata")) },
        { json: "variables", js: "variables", typ: u(undefined, a(r("JackalPackageVariable"))) },
    ], false),
    "JackalBuildData": o([
        { json: "architecture", js: "architecture", typ: "" },
        { json: "differential", js: "differential", typ: u(undefined, true) },
        { json: "differentialMissing", js: "differentialMissing", typ: u(undefined, a("")) },
        { json: "lastNonBreakingVersion", js: "lastNonBreakingVersion", typ: u(undefined, "") },
        { json: "migrations", js: "migrations", typ: u(undefined, a("")) },
        { json: "registryOverrides", js: "registryOverrides", typ: u(undefined, m("")) },
        { json: "terminal", js: "terminal", typ: "" },
        { json: "timestamp", js: "timestamp", typ: "" },
        { json: "user", js: "user", typ: "" },
        { json: "version", js: "version", typ: "" },
    ], false),
    "JackalComponent": o([
        { json: "actions", js: "actions", typ: u(undefined, r("JackalComponentActions")) },
        { json: "charts", js: "charts", typ: u(undefined, a(r("JackalChart"))) },
        { json: "cosignKeyPath", js: "cosignKeyPath", typ: u(undefined, "") },
        { json: "dataInjections", js: "dataInjections", typ: u(undefined, a(r("JackalDataInjection"))) },
        { json: "default", js: "default", typ: u(undefined, true) },
        { json: "description", js: "description", typ: u(undefined, "") },
        { json: "extensions", js: "extensions", typ: u(undefined, r("JackalComponentExtensions")) },
        { json: "files", js: "files", typ: u(undefined, a(r("JackalFile"))) },
        { json: "group", js: "group", typ: u(undefined, "") },
        { json: "images", js: "images", typ: u(undefined, a("")) },
        { json: "import", js: "import", typ: u(undefined, r("JackalComponentImport")) },
        { json: "manifests", js: "manifests", typ: u(undefined, a(r("JackalManifest"))) },
        { json: "name", js: "name", typ: "" },
        { json: "only", js: "only", typ: u(undefined, r("JackalComponentOnlyTarget")) },
        { json: "repos", js: "repos", typ: u(undefined, a("")) },
        { json: "required", js: "required", typ: u(undefined, true) },
        { json: "scripts", js: "scripts", typ: u(undefined, r("DeprecatedJackalComponentScripts")) },
    ], false),
    "JackalComponentActions": o([
        { json: "onCreate", js: "onCreate", typ: u(undefined, r("JackalComponentActionSet")) },
        { json: "onDeploy", js: "onDeploy", typ: u(undefined, r("JackalComponentActionSet")) },
        { json: "onRemove", js: "onRemove", typ: u(undefined, r("JackalComponentActionSet")) },
    ], false),
    "JackalComponentActionSet": o([
        { json: "after", js: "after", typ: u(undefined, a(r("JackalComponentAction"))) },
        { json: "before", js: "before", typ: u(undefined, a(r("JackalComponentAction"))) },
        { json: "defaults", js: "defaults", typ: u(undefined, r("JackalComponentActionDefaults")) },
        { json: "onFailure", js: "onFailure", typ: u(undefined, a(r("JackalComponentAction"))) },
        { json: "onSuccess", js: "onSuccess", typ: u(undefined, a(r("JackalComponentAction"))) },
    ], false),
    "JackalComponentAction": o([
        { json: "cmd", js: "cmd", typ: u(undefined, "") },
        { json: "description", js: "description", typ: u(undefined, "") },
        { json: "dir", js: "dir", typ: u(undefined, "") },
        { json: "env", js: "env", typ: u(undefined, a("")) },
        { json: "maxRetries", js: "maxRetries", typ: u(undefined, 0) },
        { json: "maxTotalSeconds", js: "maxTotalSeconds", typ: u(undefined, 0) },
        { json: "mute", js: "mute", typ: u(undefined, true) },
        { json: "setVariable", js: "setVariable", typ: u(undefined, "") },
        { json: "setVariables", js: "setVariables", typ: u(undefined, a(r("JackalComponentActionSetVariable"))) },
        { json: "shell", js: "shell", typ: u(undefined, r("JackalComponentActionShell")) },
        { json: "wait", js: "wait", typ: u(undefined, r("JackalComponentActionWait")) },
    ], false),
    "JackalComponentActionSetVariable": o([
        { json: "autoIndent", js: "autoIndent", typ: u(undefined, true) },
        { json: "name", js: "name", typ: "" },
        { json: "pattern", js: "pattern", typ: u(undefined, "") },
        { json: "sensitive", js: "sensitive", typ: u(undefined, true) },
        { json: "type", js: "type", typ: u(undefined, r("Type")) },
    ], false),
    "JackalComponentActionShell": o([
        { json: "darwin", js: "darwin", typ: u(undefined, "") },
        { json: "linux", js: "linux", typ: u(undefined, "") },
        { json: "windows", js: "windows", typ: u(undefined, "") },
    ], false),
    "JackalComponentActionWait": o([
        { json: "cluster", js: "cluster", typ: u(undefined, r("JackalComponentActionWaitCluster")) },
        { json: "network", js: "network", typ: u(undefined, r("JackalComponentActionWaitNetwork")) },
    ], false),
    "JackalComponentActionWaitCluster": o([
        { json: "condition", js: "condition", typ: u(undefined, "") },
        { json: "kind", js: "kind", typ: "" },
        { json: "name", js: "name", typ: "" },
        { json: "namespace", js: "namespace", typ: u(undefined, "") },
    ], false),
    "JackalComponentActionWaitNetwork": o([
        { json: "address", js: "address", typ: "" },
        { json: "code", js: "code", typ: u(undefined, 0) },
        { json: "protocol", js: "protocol", typ: r("Protocol") },
    ], false),
    "JackalComponentActionDefaults": o([
        { json: "dir", js: "dir", typ: u(undefined, "") },
        { json: "env", js: "env", typ: u(undefined, a("")) },
        { json: "maxRetries", js: "maxRetries", typ: u(undefined, 0) },
        { json: "maxTotalSeconds", js: "maxTotalSeconds", typ: u(undefined, 0) },
        { json: "mute", js: "mute", typ: u(undefined, true) },
        { json: "shell", js: "shell", typ: u(undefined, r("JackalComponentActionShell")) },
    ], false),
    "JackalChart": o([
        { json: "gitPath", js: "gitPath", typ: u(undefined, "") },
        { json: "localPath", js: "localPath", typ: u(undefined, "") },
        { json: "name", js: "name", typ: "" },
        { json: "namespace", js: "namespace", typ: "" },
        { json: "noWait", js: "noWait", typ: u(undefined, true) },
        { json: "releaseName", js: "releaseName", typ: u(undefined, "") },
        { json: "url", js: "url", typ: u(undefined, "") },
        { json: "valuesFiles", js: "valuesFiles", typ: u(undefined, a("")) },
        { json: "version", js: "version", typ: u(undefined, "") },
    ], false),
    "JackalDataInjection": o([
        { json: "compress", js: "compress", typ: u(undefined, true) },
        { json: "source", js: "source", typ: "" },
        { json: "target", js: "target", typ: r("JackalContainerTarget") },
    ], false),
    "JackalContainerTarget": o([
        { json: "container", js: "container", typ: "" },
        { json: "namespace", js: "namespace", typ: "" },
        { json: "path", js: "path", typ: "" },
        { json: "selector", js: "selector", typ: "" },
    ], false),
    "JackalComponentExtensions": o([
        { json: "bigbang", js: "bigbang", typ: u(undefined, r("BigBang")) },
    ], false),
    "BigBang": o([
        { json: "fluxPatchFiles", js: "fluxPatchFiles", typ: u(undefined, a("")) },
        { json: "repo", js: "repo", typ: u(undefined, "") },
        { json: "skipFlux", js: "skipFlux", typ: u(undefined, true) },
        { json: "valuesFiles", js: "valuesFiles", typ: u(undefined, a("")) },
        { json: "version", js: "version", typ: "" },
    ], false),
    "JackalFile": o([
        { json: "executable", js: "executable", typ: u(undefined, true) },
        { json: "extractPath", js: "extractPath", typ: u(undefined, "") },
        { json: "shasum", js: "shasum", typ: u(undefined, "") },
        { json: "source", js: "source", typ: "" },
        { json: "symlinks", js: "symlinks", typ: u(undefined, a("")) },
        { json: "target", js: "target", typ: "" },
    ], false),
    "JackalComponentImport": o([
        { json: "name", js: "name", typ: u(undefined, "") },
        { json: "path", js: "path", typ: u(undefined, "") },
        { json: "url", js: "url", typ: u(undefined, "") },
    ], false),
    "JackalManifest": o([
        { json: "files", js: "files", typ: u(undefined, a("")) },
        { json: "kustomizations", js: "kustomizations", typ: u(undefined, a("")) },
        { json: "kustomizeAllowAnyDirectory", js: "kustomizeAllowAnyDirectory", typ: u(undefined, true) },
        { json: "name", js: "name", typ: "" },
        { json: "namespace", js: "namespace", typ: u(undefined, "") },
        { json: "noWait", js: "noWait", typ: u(undefined, true) },
    ], false),
    "JackalComponentOnlyTarget": o([
        { json: "cluster", js: "cluster", typ: u(undefined, r("JackalComponentOnlyCluster")) },
        { json: "flavor", js: "flavor", typ: u(undefined, "") },
        { json: "localOS", js: "localOS", typ: u(undefined, r("LocalOS")) },
    ], false),
    "JackalComponentOnlyCluster": o([
        { json: "architecture", js: "architecture", typ: u(undefined, r("Architecture")) },
        { json: "distros", js: "distros", typ: u(undefined, a("")) },
    ], false),
    "DeprecatedJackalComponentScripts": o([
        { json: "after", js: "after", typ: u(undefined, a("")) },
        { json: "before", js: "before", typ: u(undefined, a("")) },
        { json: "prepare", js: "prepare", typ: u(undefined, a("")) },
        { json: "retry", js: "retry", typ: u(undefined, true) },
        { json: "showOutput", js: "showOutput", typ: u(undefined, true) },
        { json: "timeoutSeconds", js: "timeoutSeconds", typ: u(undefined, 0) },
    ], false),
    "JackalPackageConstant": o([
        { json: "autoIndent", js: "autoIndent", typ: u(undefined, true) },
        { json: "description", js: "description", typ: u(undefined, "") },
        { json: "name", js: "name", typ: "" },
        { json: "pattern", js: "pattern", typ: u(undefined, "") },
        { json: "value", js: "value", typ: "" },
    ], false),
    "JackalMetadata": o([
        { json: "aggregateChecksum", js: "aggregateChecksum", typ: u(undefined, "") },
        { json: "architecture", js: "architecture", typ: u(undefined, "") },
        { json: "authors", js: "authors", typ: u(undefined, "") },
        { json: "description", js: "description", typ: u(undefined, "") },
        { json: "documentation", js: "documentation", typ: u(undefined, "") },
        { json: "image", js: "image", typ: u(undefined, "") },
        { json: "name", js: "name", typ: "" },
        { json: "source", js: "source", typ: u(undefined, "") },
        { json: "uncompressed", js: "uncompressed", typ: u(undefined, true) },
        { json: "url", js: "url", typ: u(undefined, "") },
        { json: "vendor", js: "vendor", typ: u(undefined, "") },
        { json: "version", js: "version", typ: u(undefined, "") },
        { json: "yolo", js: "yolo", typ: u(undefined, true) },
    ], false),
    "JackalPackageVariable": o([
        { json: "autoIndent", js: "autoIndent", typ: u(undefined, true) },
        { json: "default", js: "default", typ: u(undefined, "") },
        { json: "description", js: "description", typ: u(undefined, "") },
        { json: "name", js: "name", typ: "" },
        { json: "pattern", js: "pattern", typ: u(undefined, "") },
        { json: "prompt", js: "prompt", typ: u(undefined, true) },
        { json: "sensitive", js: "sensitive", typ: u(undefined, true) },
        { json: "type", js: "type", typ: u(undefined, r("Type")) },
    ], false),
    "DeployedComponent": o([
        { json: "installedCharts", js: "installedCharts", typ: a(r("InstalledChart")) },
        { json: "name", js: "name", typ: "" },
        { json: "observedGeneration", js: "observedGeneration", typ: 0 },
        { json: "status", js: "status", typ: "" },
    ], false),
    "InstalledChart": o([
        { json: "chartName", js: "chartName", typ: "" },
        { json: "namespace", js: "namespace", typ: "" },
    ], false),
    "JackalState": o([
        { json: "agentTLS", js: "agentTLS", typ: r("GeneratedPKI") },
        { json: "architecture", js: "architecture", typ: "" },
        { json: "artifactServer", js: "artifactServer", typ: r("ArtifactServerInfo") },
        { json: "distro", js: "distro", typ: "" },
        { json: "gitServer", js: "gitServer", typ: r("GitServerInfo") },
        { json: "loggingSecret", js: "loggingSecret", typ: "" },
        { json: "registryInfo", js: "registryInfo", typ: r("RegistryInfo") },
        { json: "storageClass", js: "storageClass", typ: "" },
        { json: "jackalAppliance", js: "jackalAppliance", typ: true },
    ], false),
    "GeneratedPKI": o([
        { json: "ca", js: "ca", typ: "" },
        { json: "cert", js: "cert", typ: "" },
        { json: "key", js: "key", typ: "" },
    ], false),
    "ArtifactServerInfo": o([
        { json: "address", js: "address", typ: "" },
        { json: "internalServer", js: "internalServer", typ: true },
        { json: "pushPassword", js: "pushPassword", typ: "" },
        { json: "pushUsername", js: "pushUsername", typ: "" },
    ], false),
    "GitServerInfo": o([
        { json: "address", js: "address", typ: "" },
        { json: "internalServer", js: "internalServer", typ: true },
        { json: "pullPassword", js: "pullPassword", typ: "" },
        { json: "pullUsername", js: "pullUsername", typ: "" },
        { json: "pushPassword", js: "pushPassword", typ: "" },
        { json: "pushUsername", js: "pushUsername", typ: "" },
    ], false),
    "RegistryInfo": o([
        { json: "address", js: "address", typ: "" },
        { json: "internalRegistry", js: "internalRegistry", typ: true },
        { json: "nodePort", js: "nodePort", typ: 0 },
        { json: "pullPassword", js: "pullPassword", typ: "" },
        { json: "pullUsername", js: "pullUsername", typ: "" },
        { json: "pushPassword", js: "pushPassword", typ: "" },
        { json: "pushUsername", js: "pushUsername", typ: "" },
        { json: "secret", js: "secret", typ: "" },
    ], false),
    "Type": [
        "file",
        "raw",
    ],
    "Protocol": [
        "http",
        "https",
        "tcp",
    ],
    "Architecture": [
        "amd64",
        "arm64",
    ],
    "LocalOS": [
        "darwin",
        "linux",
        "windows",
    ],
    "Kind": [
        "JackalInitConfig",
        "JackalPackageConfig",
    ],
};

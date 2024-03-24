---
title: zarf package deploy
description: Zarf CLI command reference for <code>zarf package deploy</code>.
tableOfContents: false
---

## zarf package deploy

Deploys a Zarf package from a local file or URL (runs offline)

### Synopsis

Unpacks resources and dependencies from a Zarf package archive and deploys them onto the target system.
Kubernetes clusters are accessed via credentials in your current kubecontext defined in '~/.kube/config'

```
zarf package deploy [ PACKAGE_SOURCE ] [flags]
```

### Options

```
      --adopt-existing-resources   Adopts any pre-existing K8s resources into the Helm charts managed by Zarf. ONLY use when you have existing deployments you want Zarf to takeover.
      --components string          Comma-separated list of components to deploy.  Adding this flag will skip the prompts for selected components.  Globbing component names with '*' and deselecting 'default' components with a leading '-' are also supported.
      --confirm                    Confirms package deployment without prompting. ONLY use with packages you trust. Skips prompts to review SBOM, configure variables, select optional components and review potential breaking changes.
  -h, --help                       help for deploy
      --set stringToString         Specify deployment variables to set on the command line (KEY=value) (default [])
      --shasum string              Shasum of the package to deploy. Required if deploying a remote package and "--insecure" is not provided
      --skip-webhooks              [alpha] Skip waiting for external webhooks to execute as each package component is deployed
      --timeout duration           Timeout for Helm operations such as installs and rollbacks (default 15m0s)
```

### Options inherited from parent commands

```
  -a, --architecture string   Architecture for OCI images and Zarf packages
      --insecure              Allow access to insecure registries and disable other recommended security enforcements such as package checksum and signature validation. This flag should only be used if you have a specific reason and accept the reduced security posture.
  -k, --key string            Path to public key file for validating signed packages
  -l, --log-level string      Log level when running Zarf. Valid options are: warn, info, debug, trace (default "info")
      --no-color              Disable colors in output
      --no-log-file           Disable log file creation
      --no-progress           Disable fancy UI progress bars, spinners, logos, etc
      --oci-concurrency int   Number of concurrent layer operations to perform when interacting with a remote package. (default 3)
      --tmpdir string         Specify the temporary directory to use for intermediate files
      --zarf-cache string     Specify the location of the Zarf cache directory (default "~/.zarf-cache")
```

### SEE ALSO

* [zarf package](/commands/zarf_package/)	 - Zarf package commands for creating, deploying, and inspecting packages


---
title: zarf
description: Zarf CLI command reference for <code>zarf</code>.
---

## zarf

DevSecOps for Airgap

### Synopsis

Zarf eliminates the complexity of air gap software delivery for Kubernetes clusters and cloud native workloads
using a declarative packaging strategy to support DevSecOps in offline and semi-connected environments.

```
zarf COMMAND [flags]
```

### Options

```
  -a, --architecture string   Architecture for OCI images and Zarf packages
  -h, --help                  help for zarf
      --insecure              Allow access to insecure registries and disable other recommended security enforcements such as package checksum and signature validation. This flag should only be used if you have a specific reason and accept the reduced security posture.
  -l, --log-level string      Log level when running Zarf. Valid options are: warn, info, debug, trace (default "info")
      --no-color              Disable colors in output
      --no-log-file           Disable log file creation
      --no-progress           Disable fancy UI progress bars, spinners, logos, etc
      --tmpdir string         Specify the temporary directory to use for intermediate files
      --zarf-cache string     Specify the location of the Zarf cache directory (default "~/.zarf-cache")
```

### SEE ALSO

* [zarf completion](/cli/commands/zarf_completion/)	 - Generate the autocompletion script for the specified shell
* [zarf connect](/cli/commands/zarf_connect/)	 - Accesses services or pods deployed in the cluster
* [zarf destroy](/cli/commands/zarf_destroy/)	 - Tears down Zarf and removes its components from the environment
* [zarf dev](/cli/commands/zarf_dev/)	 - Commands useful for developing packages
* [zarf init](/cli/commands/zarf_init/)	 - Prepares a k8s cluster for the deployment of Zarf packages
* [zarf package](/cli/commands/zarf_package/)	 - Zarf package commands for creating, deploying, and inspecting packages
* [zarf tools](/cli/commands/zarf_tools/)	 - Collection of additional tools to make airgap easier
* [zarf version](/cli/commands/zarf_version/)	 - Shows the version of the running Zarf binary

---
title: zarf package
description: Zarf CLI command reference for <code>zarf package</code>.
---

## zarf package

Zarf package commands for creating, deploying, and inspecting packages

### Options

```
  -h, --help                  help for package
  -k, --key string            Path to public key file for validating signed packages
      --oci-concurrency int   Number of concurrent layer operations to perform when interacting with a remote package. (default 3)
```

### Options inherited from parent commands

```
  -a, --architecture string   Architecture for OCI images and Zarf packages
      --insecure              Allow access to insecure registries and disable other recommended security enforcements such as package checksum and signature validation. This flag should only be used if you have a specific reason and accept the reduced security posture.
  -l, --log-level string      Log level when running Zarf. Valid options are: warn, info, debug, trace (default "info")
      --no-color              Disable colors in output
      --no-log-file           Disable log file creation
      --no-progress           Disable fancy UI progress bars, spinners, logos, etc
      --tmpdir string         Specify the temporary directory to use for intermediate files
      --zarf-cache string     Specify the location of the Zarf cache directory (default "~/.zarf-cache")
```

### SEE ALSO

* [zarf](/cli/commands/zarf/)	 - DevSecOps for Airgap
* [zarf package create](/cli/commands/zarf_package_create/)	 - Creates a Zarf package from a given directory or the current directory
* [zarf package deploy](/cli/commands/zarf_package_deploy/)	 - Deploys a Zarf package from a local file or URL (runs offline)
* [zarf package inspect](/cli/commands/zarf_package_inspect/)	 - Displays the definition of a Zarf package (runs offline)
* [zarf package list](/cli/commands/zarf_package_list/)	 - Lists out all of the packages that have been deployed to the cluster (runs offline)
* [zarf package mirror-resources](/cli/commands/zarf_package_mirror-resources/)	 - Mirrors a Zarf package's internal resources to specified image registries and git repositories
* [zarf package publish](/cli/commands/zarf_package_publish/)	 - Publishes a Zarf package to a remote registry
* [zarf package pull](/cli/commands/zarf_package_pull/)	 - Pulls a Zarf package from a remote registry and save to the local file system
* [zarf package remove](/cli/commands/zarf_package_remove/)	 - Removes a Zarf package that has been deployed already (runs offline)

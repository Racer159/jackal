---
title: zarf package remove
description: Zarf CLI command reference for <code>zarf package remove</code>.
tableOfContents: false
---

## zarf package remove

Removes a Zarf package that has been deployed already (runs offline)

```
zarf package remove { PACKAGE_SOURCE | PACKAGE_NAME } --confirm [flags]
```

### Options

```
      --components string   Comma-separated list of components to remove.  This list will be respected regardless of a component's 'required' or 'default' status.  Globbing component names with '*' and deselecting components with a leading '-' are also supported.
      --confirm             REQUIRED. Confirm the removal action to prevent accidental deletions
  -h, --help                help for remove
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


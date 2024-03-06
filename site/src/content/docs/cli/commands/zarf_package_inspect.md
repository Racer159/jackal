---
title: zarf package inspect
description: Zarf CLI command reference for <code>zarf package inspect</code>.
---

## zarf package inspect

Displays the definition of a Zarf package (runs offline)

### Synopsis

Displays the 'zarf.yaml' definition for the specified package and optionally allows SBOMs to be viewed

```
zarf package inspect [ PACKAGE_SOURCE ] [flags]
```

### Options

```
  -h, --help              help for inspect
  -s, --sbom              View SBOM contents while inspecting the package
      --sbom-out string   Specify an output directory for the SBOMs from the inspected Zarf package
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

* [zarf package](/cli/commands/zarf_package/)	 - Zarf package commands for creating, deploying, and inspecting packages

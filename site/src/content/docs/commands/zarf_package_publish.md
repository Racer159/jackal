---
title: zarf package publish
description: Zarf CLI command reference for <code>zarf package publish</code>.
tableOfContents: false
---

## zarf package publish

Publishes a Zarf package to a remote registry

```
zarf package publish { PACKAGE_SOURCE | SKELETON DIRECTORY } REPOSITORY [flags]
```

### Examples

```

# Publish a package to a remote registry
$ zarf package publish my-package.tar oci://my-registry.com/my-namespace

# Publish a skeleton package to a remote registry
$ zarf package publish ./path/to/dir oci://my-registry.com/my-namespace

```

### Options

```
  -h, --help                      help for publish
      --signing-key string        Path to a private key file for signing or re-signing packages with a new key
      --signing-key-pass string   Password to the private key file used for publishing packages
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


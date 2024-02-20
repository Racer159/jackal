---
title: zarf tools gen-pki
---

## zarf tools gen-pki

Generates a Certificate Authority and PKI chain of trust for the given host

```
zarf tools gen-pki HOST [flags]
```

### Options

```
  -h, --help                       help for gen-pki
      --sub-alt-name stringArray   Specify Subject Alternative Names for the certificate
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

* [zarf tools](/cli/commands/zarf_tools/)	 - Collection of additional tools to make airgap easier

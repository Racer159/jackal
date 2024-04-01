# jackal package pull
<!-- Auto-generated by hack/gen-cli-docs.sh -->

Pulls a Jackal package from a remote registry and save to the local file system

```
jackal package pull PACKAGE_SOURCE [flags]
```

## Examples

```

# Pull a package matching the current architecture
$ jackal package pull oci://ghcr.io/racer159/packages/dos-games:1.0.0

# Pull a package matching a specific architecture
$ jackal package pull oci://ghcr.io/racer159/packages/dos-games:1.0.0 -a arm64

# Pull a skeleton package
$ jackal package pull oci://ghcr.io/racer159/packages/dos-games:1.0.0 -a skeleton
```

## Options

```
  -h, --help                      help for pull
  -o, --output-directory string   Specify the output directory for the pulled Jackal package
```

## Options inherited from parent commands

```
  -a, --architecture string   Architecture for OCI images and Jackal packages
      --insecure              Allow access to insecure registries and disable other recommended security enforcements such as package checksum and signature validation. This flag should only be used if you have a specific reason and accept the reduced security posture.
  -k, --key string            Path to public key file for validating signed packages
  -l, --log-level string      Log level when running Jackal. Valid options are: warn, info, debug, trace (default "info")
      --no-color              Disable colors in output
      --no-log-file           Disable log file creation
      --no-progress           Disable fancy UI progress bars, spinners, logos, etc
      --oci-concurrency int   Number of concurrent layer operations to perform when interacting with a remote package. (default 3)
      --tmpdir string         Specify the temporary directory to use for intermediate files
      --jackal-cache string     Specify the location of the Jackal cache directory (default "~/.jackal-cache")
```

## SEE ALSO

* [jackal package](jackal_package.md)	 - Jackal package commands for creating, deploying, and inspecting packages
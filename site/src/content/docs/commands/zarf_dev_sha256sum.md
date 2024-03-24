---
title: zarf dev sha256sum
description: Zarf CLI command reference for <code>zarf dev sha256sum</code>.
tableOfContents: false
---

## zarf dev sha256sum

Generates a SHA256SUM for the given file

```
zarf dev sha256sum { FILE | URL } [flags]
```

### Options

```
  -e, --extract-path string   The path inside of an archive to use to calculate the sha256sum (i.e. for use with "files.extractPath")
  -h, --help                  help for sha256sum
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

* [zarf dev](/commands/zarf_dev/)	 - Commands useful for developing packages


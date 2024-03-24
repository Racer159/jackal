---
title: zarf tools archiver decompress
description: Zarf CLI command reference for <code>zarf tools archiver decompress</code>.
tableOfContents: false
---

## zarf tools archiver decompress

Decompresses an archive or Zarf package based off of the source file extension.

```
zarf tools archiver decompress ARCHIVE DESTINATION [flags]
```

### Options

```
  -h, --help            help for decompress
      --unarchive-all   Unarchive all tarballs in the archive
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

* [zarf tools archiver](/commands/zarf_tools_archiver/)	 - Compresses/Decompresses generic archives, including Zarf packages


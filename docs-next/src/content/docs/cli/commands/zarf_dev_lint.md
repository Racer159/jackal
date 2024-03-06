---
title: zarf dev lint
description: Zarf CLI command reference for <code>zarf dev lint</code>.
---

## zarf dev lint

Lints the given package for valid schema and recommended practices

### Synopsis

Verifies the package schema, checks if any variables won't be evaluated, and checks for unpinned images/repos/files

```
zarf dev lint [ DIRECTORY ] [flags]
```

### Options

```
  -f, --flavor string        The flavor of components to include in the resulting package (i.e. have a matching or empty "only.flavor" key)
  -h, --help                 help for lint
      --set stringToString   Specify package variables to set on the command line (KEY=value) (default [])
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

* [zarf dev](/cli/commands/zarf_dev/)	 - Commands useful for developing packages

---
title: zarf dev generate-config
description: Zarf CLI command reference for <code>zarf dev generate-config</code>.
tableOfContents: false
---

## zarf dev generate-config

Generates a config file for Zarf

### Synopsis

Generates a Zarf config file for controlling how the Zarf CLI operates. Optionally accepts a filename to write the config to.

The extension will determine the format of the config file, e.g. env-1.yaml, env-2.json, env-3.toml etc.
Accepted extensions are json, toml, yaml.

NOTE: This file must not already exist. If no filename is provided, the config will be written to the current working directory as zarf-config.toml.

```
zarf dev generate-config [ FILENAME ] [flags]
```

### Options

```
  -h, --help   help for generate-config
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


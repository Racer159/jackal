---
title: zarf completion
---

## zarf completion

Generate the autocompletion script for the specified shell

### Synopsis

Generate the autocompletion script for zarf for the specified shell.
See each sub-command's help for details on how to use the generated script.


### Options

```
  -h, --help   help for completion
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
* [zarf completion bash](/cli/commands/zarf_completion_bash/)	 - Generate the autocompletion script for bash
* [zarf completion fish](/cli/commands/zarf_completion_fish/)	 - Generate the autocompletion script for fish
* [zarf completion powershell](/cli/commands/zarf_completion_powershell/)	 - Generate the autocompletion script for powershell
* [zarf completion zsh](/cli/commands/zarf_completion_zsh/)	 - Generate the autocompletion script for zsh

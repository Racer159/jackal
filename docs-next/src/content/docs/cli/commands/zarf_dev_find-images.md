---
title: zarf dev find-images
---

## zarf dev find-images

Evaluates components in a Zarf file to identify images specified in their helm charts and manifests

### Synopsis

Evaluates components in a Zarf file to identify images specified in their helm charts and manifests.

Components that have repos that host helm charts can be processed by providing the --repo-chart-path.

```
zarf dev find-images [ PACKAGE ] [flags]
```

### Options

```
  -h, --help                     help for find-images
      --kube-version string      Override the default helm template KubeVersion when performing a package chart template
  -p, --repo-chart-path string   If git repos hold helm charts, often found with gitops tools, specify the chart path, e.g. "/" or "/chart"
      --set stringToString       Specify package variables to set on the command line (KEY=value). Note, if using a config file, this will be set by [package.create.set]. (default [])
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

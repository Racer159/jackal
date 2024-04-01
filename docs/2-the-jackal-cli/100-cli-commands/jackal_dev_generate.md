# jackal dev generate
<!-- Auto-generated by hack/gen-cli-docs.sh -->

[alpha] Creates a jackal.yaml automatically from a given remote (git) Helm chart

```
jackal dev generate NAME [flags]
```

## Examples

```
jackal dev generate podinfo --url https://github.com/stefanprodan/podinfo.git --version 6.4.0 --gitPath charts/podinfo
```

## Options

```
      --gitPath string            Relative path to the chart in the git repository
  -h, --help                      help for generate
      --kube-version string       Override the default helm template KubeVersion when performing a package chart template
      --output-directory string   Output directory for the generated jackal.yaml
      --url string                URL to the source git repository
      --version string            The Version of the chart to use
```

## Options inherited from parent commands

```
  -a, --architecture string   Architecture for OCI images and Jackal packages
      --insecure              Allow access to insecure registries and disable other recommended security enforcements such as package checksum and signature validation. This flag should only be used if you have a specific reason and accept the reduced security posture.
  -l, --log-level string      Log level when running Jackal. Valid options are: warn, info, debug, trace (default "info")
      --no-color              Disable colors in output
      --no-log-file           Disable log file creation
      --no-progress           Disable fancy UI progress bars, spinners, logos, etc
      --tmpdir string         Specify the temporary directory to use for intermediate files
      --jackal-cache string     Specify the location of the Jackal cache directory (default "~/.jackal-cache")
```

## SEE ALSO

* [jackal dev](jackal_dev.md)	 - Commands useful for developing packages
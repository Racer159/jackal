---
title: zarf tools sbom attest
description: Zarf CLI command reference for <code>zarf tools sbom attest</code>.
tableOfContents: false
---

## zarf tools sbom attest

Generate an SBOM as an attestation for the given [SOURCE] container image

### Synopsis

Generate a packaged-based Software Bill Of Materials (SBOM) from a container image as the predicate of an in-toto attestation that will be uploaded to the image registry

```
zarf tools sbom attest --output [FORMAT] <IMAGE> [flags]
```

### Options

```
      --base-path string         base directory for scanning, no links will be followed above this directory, and all paths will be reported relative to this directory
      --catalogers stringArray   enable one or more package catalogers
      --exclude stringArray      exclude paths from being scanned using a glob expression
  -h, --help                     help for attest
      --name string              set the name of the target being analyzed (DEPRECATED: use: source-name)
  -o, --output stringArray       report output format (<format>=<file> to output to a file), formats=[cyclonedx-json cyclonedx-xml github-json spdx-json spdx-tag-value syft-json syft-table syft-text template] (default [syft-json])
      --platform string          an optional platform specifier for container image sources (e.g. 'linux/arm64', 'linux/arm64/v8', 'arm64', 'linux')
  -s, --scope string             selection of layers to catalog, options=[squashed all-layers]
      --source-name string       set the name of the target being analyzed
      --source-version string    set the version of the target being analyzed
```

### Options inherited from parent commands

```
  -c, --config string   syft configuration file
  -q, --quiet           suppress all logging output
  -v, --verbose count   increase verbosity (-v = info, -vv = debug)
```

### SEE ALSO

* [zarf tools sbom](/commands/zarf_tools_sbom/)	 - Generates a Software Bill of Materials (SBOM) for the given package

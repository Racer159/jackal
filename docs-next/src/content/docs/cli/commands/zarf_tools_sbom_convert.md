---
title: zarf tools sbom convert
description: Zarf CLI command reference for <code>zarf tools sbom convert</code>.
---

## zarf tools sbom convert

Convert between SBOM formats

### Synopsis

[Experimental] Convert SBOM files to, and from, SPDX, CycloneDX and Syft's format. For more info about data loss between formats see https://github.com/anchore/syft#format-conversion-experimental

```
zarf tools sbom convert [SOURCE-SBOM] -o [FORMAT] [flags]
```

### Options

```
      --file string          file to write the default report output to (default is STDOUT) (DEPRECATED: use: output)
  -h, --help                 help for convert
  -o, --output stringArray   report output format (<format>=<file> to output to a file), formats=[cyclonedx-json cyclonedx-xml github-json spdx-json spdx-tag-value syft-json syft-table syft-text template] (default [syft-table])
  -t, --template string      specify the path to a Go template file
```

### Options inherited from parent commands

```
  -c, --config string   syft configuration file
  -q, --quiet           suppress all logging output
  -v, --verbose count   increase verbosity (-v = info, -vv = debug)
```

### SEE ALSO

* [zarf tools sbom](/cli/commands/zarf_tools_sbom/)	 - Generates a Software Bill of Materials (SBOM) for the given package

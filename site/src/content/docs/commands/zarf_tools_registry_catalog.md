---
title: zarf tools registry catalog
description: Zarf CLI command reference for <code>zarf tools registry catalog</code>.
tableOfContents: false
---

## zarf tools registry catalog

List the repos in a registry

```
zarf tools registry catalog REGISTRY [flags]
```

### Examples

```

# List the repos internal to Zarf
$ zarf tools registry catalog

# List the repos for reg.example.com
$ zarf tools registry catalog reg.example.com

```

### Options

```
      --full-ref   (Optional) if true, print the full image reference
  -h, --help       help for catalog
```

### Options inherited from parent commands

```
      --allow-nondistributable-artifacts   Allow pushing non-distributable (foreign) layers
      --insecure                           Allow image references to be fetched without TLS
      --platform string                    Specifies the platform in the form os/arch[/variant][:osversion] (e.g. linux/amd64). (default "all")
  -v, --verbose                            Enable debug logs
```

### SEE ALSO

* [zarf tools registry](/commands/zarf_tools_registry/)	 - Tools for working with container registries using go-containertools


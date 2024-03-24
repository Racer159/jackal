---
title: zarf tools registry pull
description: Zarf CLI command reference for <code>zarf tools registry pull</code>.
tableOfContents: false
---

## zarf tools registry pull

Pull remote images by reference and store their contents locally

```
zarf tools registry pull IMAGE TARBALL [flags]
```

### Examples

```

# Pull an image from an internal repo in Zarf to a local tarball
$ zarf tools registry pull 127.0.0.1:31999/stefanprodan/podinfo:6.4.0 image.tar

# Pull an image from a repo hosted at reg.example.com to a local tarball
$ zarf tools registry pull reg.example.com/stefanprodan/podinfo:6.4.0 image.tar

```

### Options

```
      --annotate-ref        Preserves image reference used to pull as an annotation when used with --format=oci
  -c, --cache_path string   Path to cache image layers
      --format string       Format in which to save images ("tarball", "legacy", or "oci") (default "tarball")
  -h, --help                help for pull
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

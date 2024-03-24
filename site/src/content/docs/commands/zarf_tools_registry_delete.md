---
title: zarf tools registry delete
description: Zarf CLI command reference for <code>zarf tools registry delete</code>.
---

## zarf tools registry delete

Delete an image reference from its registry

```
zarf tools registry delete IMAGE [flags]
```

### Examples

```

# Delete an image digest from an internal repo in Zarf
$ zarf tools registry delete 127.0.0.1:31999/stefanprodan/podinfo@sha256:57a654ace69ec02ba8973093b6a786faa15640575fbf0dbb603db55aca2ccec8

# Delete an image digest from a repo hosted at reg.example.com
$ zarf tools registry delete reg.example.com/stefanprodan/podinfo@sha256:57a654ace69ec02ba8973093b6a786faa15640575fbf0dbb603db55aca2ccec8

```

### Options

```
  -h, --help   help for delete
```

### Options inherited from parent commands

```
      --allow-nondistributable-artifacts   Allow pushing non-distributable (foreign) layers
      --insecure                           Allow image references to be fetched without TLS
      --platform string                    Specifies the platform in the form os/arch[/variant][:osversion] (e.g. linux/amd64). (default "all")
  -v, --verbose                            Enable debug logs
```

### SEE ALSO

* [zarf tools registry](/cli/commands/zarf_tools_registry/)	 - Tools for working with container registries using go-containertools


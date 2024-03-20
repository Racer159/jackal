---
title: zarf tools registry login
description: Zarf CLI command reference for <code>zarf tools registry login</code>.
---

## zarf tools registry login

Log in to a registry

```
zarf tools registry login [OPTIONS] [SERVER] [flags]
```

### Options

```
  -h, --help              help for login
  -p, --password string   Password
      --password-stdin    Take the password from stdin
  -u, --username string   Username
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


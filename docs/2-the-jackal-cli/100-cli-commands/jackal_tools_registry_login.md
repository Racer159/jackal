# jackal tools registry login
<!-- Auto-generated by hack/gen-cli-docs.sh -->

Log in to a registry

```
jackal tools registry login [OPTIONS] [SERVER] [flags]
```

## Options

```
  -h, --help              help for login
  -p, --password string   Password
      --password-stdin    Take the password from stdin
  -u, --username string   Username
```

## Options inherited from parent commands

```
      --allow-nondistributable-artifacts   Allow pushing non-distributable (foreign) layers
      --insecure                           Allow image references to be fetched without TLS
      --platform string                    Specifies the platform in the form os/arch[/variant][:osversion] (e.g. linux/amd64). (default "all")
  -v, --verbose                            Enable debug logs
```

## SEE ALSO

* [jackal tools registry](jackal_tools_registry.md)	 - Tools for working with container registries using go-containertools

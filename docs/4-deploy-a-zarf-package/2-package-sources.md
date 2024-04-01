# Package Sources

Jackal currently supports consuming packages from the following sources:

### Local Tarball Path (`.tar` and `.tar.zst`)

A local tarball is the default output of `jackal package create` and is a package contained within a tarball with or without [Zstandard](https://facebook.github.io/zstd/) compression.  Compression is determined by a given package's [`metadata.uncompressed` key](https://docs.jackal.dev/docs/create-a-jackal-package/jackal-schema#metadata) within it's `jackal.yaml` package definition

### Split Tarball Path (`.part...`)

A split tarball is a local tarball that has been split into multiple parts so that it can fit on smaller media when traveling to a disconnected environment (i.e. on DVDs).  These packages are created by specifying a maximum number of megabytes with [`--max-package-size`](../2-the-jackal-cli/100-cli-commands/jackal_package_create.md) on `jackal package create` and if the resulting tarball is larger than that size it will be split into chunks.

### Remote Tarball URL (`http://` and `https://` )

A remote tarball is a Jackal package tarball that is hosted on a web server that is accessible to the current machine.  By default Jackal does not provide a mechanism to place a package on a web server, but this is easy to orchestrate with other tooling such as uploading a package to a continuous integration system's artifact storage or to a repository's release page.

### Remote OCI Reference (`oci://`)

An OCI package is one that has been published to an OCI compatible registry using `jackal package publish` or the `-o` option on `jackal package create`.  These packages live within a given registry and you can learn more about them in our [Publish & Deploy Packages w/OCI Tutorial](../5-jackal-tutorials/7-publish-and-deploy.md).

## Commands with Sources

A source can be used with the following commands as their first argument:

- `jackal package deploy <source>`
- `jackal package inspect <source>`
- `jackal package remove <source>`
- `jackal package publish <source>`
- `jackal package pull <source>`
- `jackal package mirror-resources <source>`

:::note

In addition to the traditional sources outlined above, there is also a special "Cluster" source available on `inspect` and `remove` that allows for referencing a deployed package via its name:

- `jackal package inspect <package name>`
- `jackal package remove <package name>`

Additionally, inspecting a package deployed to a cluster will not be able to show the package's SBOMs, as they are not currently persisted to the cluster.

:::

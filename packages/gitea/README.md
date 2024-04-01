## Jackal Git Server

This package contains the Jackal Git Server to enable more advanced gitops-based deployments. See the [git-data](../../examples/git-data/) example for more on how Jackal handles `git` repositories.

### Image Values

The default setup for this package is to use a `rootless` image, specified in the [gitea helm values](gitea-values.yaml). Because the gitea helm chart does its own appending of `-rootless` to the image tag, based on the `rootless` helm value, users don't need to supply the full image tag when overriding the default gitea image. Instead you need to use the `GITEA_SERVER_VERSION`, either in the jackal-config.toml or with `--set`.

_Make sure, though, that the `x.x.x-rootless` tag does exist for Jackal to find._

```bash
$ jackal package create . --set GITEA_IMAGE="custom.enterprise.corp/ironbank/opensource/gitea" \
  --set GITEA_SERVER_VERSION="v1.19.3"
```

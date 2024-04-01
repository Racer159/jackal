# Creating a Custom 'init' Package

## Introduction

In most cases the default Jackal 'init' Package will provide what you need to get started deploying packages into the air gap, however there are cases where you may want to tweak this package to tailor it for your target environment. This could include adding or removing components or including hardened versions of components specific to your use case.

In this tutorial, we will demonstrate how to build a custom [Jackal 'init' Package](../3-create-a-jackal-package/3-jackal-init-package.md) with `jackal package create`.

When creating a Jackal 'init' package, you must have a network connection so that Jackal can fetch all of the dependencies and resources necessary to build the package. If your version of the 'init' package is using images from a private registry or is referencing repositories in a private repository, you will need to have your credentials configured on your machine for Jackal to be able to fetch the resources.

## System Requirements

- For the default `init` package you will require an Internet connection to pull down the resources Jackal needs.

## Prerequisites

Before beginning this tutorial you will need the following:

- The [Jackal](https://github.com/defenseunicorns/jackal) repository cloned: ([git clone instructions](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository))
- Jackal binary installed on your $PATH: ([Installing Jackal](../1-getting-started/index.md#installing-jackal))
- (if building a local [`jackal-agent`](../8-faq.md#what-is-the-jackal-agent)) The [Docker CLI](https://docs.docker.com/desktop/) installed and the tools to [Build your own CLI](../2-the-jackal-cli/0-building-your-own-cli.md)

## Building the init-package

Creating the jackal 'init' package is as simple as creating any other package. All you need to do is run the `jackal package create` command within the Jackal git repository.

```bash
$ cd jackal # Enter the jackal repository that you have cloned down
$ jackal version
vX.X.X
$ git checkout vX.X.X # checkout the version that corresponds to your jackal version
# Run the command to create the jackal package where the AGENT_IMAGE_TAG matches your jackal version
$ jackal package create . --set AGENT_IMAGE_TAG=vX.X.X
# Type `y` when prompted and then hit the enter key
```

:::tip

For development if you omit the `AGENT_IMAGE_TAG` Jackal will build a Jackal Agent image based on the source code within the Jackal git repository you cloned.

:::

:::note

Prior to v0.26.0 `AGENT_IMAGE_TAG` was `AGENT_IMAGE` and would be set like: `jackal package create . --set AGENT_IMAGE=agent:vX.X.X`

:::

When you execute the `jackal package create` command, Jackal will prompt you to confirm that you want to create the package by displaying the package definition and asking you to respond with either `y` or `n`.

<iframe src="/docs/tutorials/package_create_init.html" height="500px" width="100%"></iframe>

:::tip

You can skip this confirmation by adding the `--confirm` flag when running the command. This will look like: `jackal package create . --confirm`

:::

After you confirm package creation, Jackal will create the Jackal 'init' package in the current directory. In this case, the package name should look something like `jackal-init-amd64-vX.X.X.tar.zst`, although it might differ slightly depending on your system architecture.

## Customizing the 'init' Package

The above will simply build the init package as it is defined for your version of Jackal. To build something custom you will need to make some modifications.

The Jackal 'init' Package is a [composed Jackal Package](../3-create-a-jackal-package/2-jackal-components.md#composing-package-components) made up of many sub-Jackal Packages. The root `jackal.yaml` file is defined at the root of the Jackal git repository.

### Swapping Images

As of v0.26.0 you can swap the `registry` and `agent` images by specifying different values in the `jackal-config.toml` file at the root of the project or by overriding them as we did above with `--set` on the command line. This allows you to swap these images for hardened or enterprise-vetted versions like those from [Iron Bank](https://repo1.dso.mil/dsop/opensource/defenseunicorns/jackal/jackal-agent).

For other components, or older versions of Jackal, you can modify the manifests of the components you want to change in their individual packages under the `packages` folder of the Jackal repo.

:::tip

If your enterprise uses pull-through mirrors to host vetted images you can run the following command to create a Jackal 'init' package from those mirrors (where `<registry>.enterprise.corp` are your enterprise mirror(s)):

```bash
$ jackal package create . --set AGENT_IMAGE_TAG=vX.X.X \
  --registry-override docker.io=dockerio.enterprise.corp \
  --registry-override ghcr.io=ghcr.enterprise.corp \
  --registry-override quay.io=quay.enterprise.corp
```

And if you need even more control over the exact Agent, Registry, and Gitea images you can specify that with additional `--set` flags:

```bash
$ jackal package create . \
--set AGENT_IMAGE_TAG=$(jackal version) \
--set AGENT_IMAGE="opensource/jackal" \
--set AGENT_IMAGE_DOMAIN="custom.enterprise.corp" \
--set REGISTRY_IMAGE_TAG=2.8.3 \
--set REGISTRY_IMAGE="opensource/registry" \
--set REGISTRY_IMAGE_DOMAIN="custom.enterprise.corp" \
--set GITEA_IMAGE="custom.enterprise.corp/opensource/gitea:v1.21.0-rootless"
```

⚠️ - The Gitea image is different from the Agent and Registry in that Jackal will always prefer the `rootless` version of a given server image. The image no longer must be tagged with `-rootless`, but it still needs to implement the [Gitea configuration of a rootless image](https://github.com/go-gitea/gitea/blob/main/Dockerfile.rootless). If you need to change this, edit the `packages/gitea` package.

You can find all of the `--set` configurations by looking at the `jackal-config.toml` in the root of the repository.

:::

### Removing Components

You may not need or want all of the components in your 'init' package and may choose to slim down your package by removing them. Because the [Jackal Package is composed](../3-create-a-jackal-package/2-jackal-components.md#composing-package-components) all you need to do is remove the component that imports the component you wish to exclude.

## Troubleshooting

### Unable to read jackal.yaml file

<iframe src="/docs/tutorials/package_create_error.html" height="120px" width="100%"></iframe>

:::info Remediation

If you receive this error, you may not be in the correct directory. Double-check where you are in your system and try again once you're in the correct directory with the jackal.yaml file that you're trying to build.

:::

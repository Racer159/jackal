import TabItem from "@theme/TabItem";
import Tabs from "@theme/Tabs";

# Getting Started

Welcome to the Jackal documentation!  This page runs through a quick start to test Jackal on your machine and walks through next steps to get more familiar with Jackal's concepts. Let's get started!

## Quick Start

Trying out Jackal is as simple as:

1. 💻 Selecting your system's OS below.
2. ❗ Ensuring you have the pre-requisite applications running.
3. `$` Entering the commands into your terminal.

<Tabs>
<TabItem value="macOS">

:::note

This quick start requires you to already have:

- [Homebrew](https://brew.sh/) package manager installed on your machine.
- [Docker](https://www.docker.com/) installed and running on your machine.

For more install options please visit our [Installing Jackal page](./0-installing-jackal.md).

:::

## macOS Commands

```bash
# To install Jackal with Homebrew simply run:
brew tap defenseunicorns/tap && brew install jackal

# Next, you will need a Kubernetes cluster. This example uses KIND.
brew install kind && kind delete cluster && kind create cluster

# Then, you need to initialize the cluster with Jackal:
jackal init
# (Select 'Y' to download the default init package)
# (Select 'Y' to confirm deployment)
# (Select optional components as desired)

# Now you are ready to deploy any Jackal Package, try out our Retro Arcade!!
jackal package deploy oci://🦄/dos-games:1.0.0-$(uname -m) --key=https://jackal.dev/cosign.pub
# (Select 'Y' to confirm deployment)
```

</TabItem>
<TabItem value="Linux">

:::note

This quick start requires you to already have:

- [Homebrew](https://brew.sh/) package manager installed on your machine.
- [Docker](https://www.docker.com/) installed and running on your machine.

For more install options please visit our [Installing Jackal page](./0-installing-jackal.md).

:::

## Linux Commands

```bash
# To install Jackal with Homebrew simply run:
brew tap defenseunicorns/tap && brew install jackal

# Next, you will need a Kubernetes cluster. This example uses KinD.
brew install kind && kind delete cluster && kind create cluster
# (Note: you don't need 'KinD' if you have 'root' access since Jackal includes 'k3s' as an optional component)

# Then, you need to initialize the cluster, following the prompts to download and select components
jackal init
# (Select 'Y' to download the default init package)
# (Select 'Y' to confirm deployment)
# (Select 'N' for 'k3s' - this only works when run as 'root')
# (Select other optional components as desired)

# Now you are ready to deploy any Jackal Package, try out our Retro Arcade!!
jackal package deploy oci://🦄/dos-games:1.0.0-$(uname -m) --key=https://jackal.dev/cosign.pub
# (Select 'Y' to confirm deployment)
```

:::tip

This example shows how to install Jackal with the official (📜) `defenseunicorns` Homebrew tap, however there are many other options to install Jackal on Linux such as:

- 📜 **[official]** Downloading Jackal directly from [GitHub releases](https://github.com/defenseunicorns/jackal/releases)
- 🧑‍🤝‍🧑 **[community]** `apk add` on [Alpine Linux Edge](https://pkgs.alpinelinux.org/package/edge/testing/x86_64/jackal)
- 🧑‍🤝‍🧑 **[community]** `asdf install` with the [ASDF Version Manager](https://github.com/defenseunicorns/asdf-jackal)
- 🧑‍🤝‍🧑 **[community]** `nix-shell`/`nix-env` with [Nix Packages](https://search.nixos.org/packages?channel=23.05&show=jackal&from=0&size=50&sort=relevance&type=packages&query=jackal)

:::

</TabItem>
<TabItem value="Windows">

## Windows Commands

:::note

There is currently no Jackal quick start for Windows, though you can learn how to install Jackal from our Github Releases by visiting the [Installing Jackal page](./0-installing-jackal.md#downloading-the-cli-from-github-releases).

:::

```text

Coming soon!

```

</TabItem>
</Tabs>

## Where to Next?

Depending on how familiar you are with Kubernetes, DevOps, and Jackal, let's find what set of information would be most useful to you.

- If you want to become more familiar with Jackal and it's features, see the [Tutorials](../5-jackal-tutorials/index.md) page.

- More information about the Jackal CLI is available on the [Jackal CLI](../2-the-jackal-cli/index.md) page, or by browsing through the help descriptions of all the commands available through `jackal --help`.

- More information about the packages that Jackal creates and deploys is available in the [Understanding Jackal Packages](../3-create-a-jackal-package/1-jackal-packages.md) page.

- If you want to take a step back and better understand the problem Jackal is trying to solve, you can find more context on the [Understand the Basics](./1-understand-the-basics.md) and [Core Concepts](./2-core-concepts.md) pages.

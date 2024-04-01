import TabItem from "@theme/TabItem";
import Tabs from "@theme/Tabs";

# Installing Jackal

Depending on your operating system and specific setup there are a few ways you can get the Jackal CLI onto your machine:

- [Install from Homebrew](#installing-from-the-defense-unicorns-homebrew-tap).
- [Download a prebuilt binary](#downloading-a-prebuilt-binary-from-our-github-releases).
- [Build the CLI](#building-the-cli-from-scratch) from scratch.

[Post-Install](#post-install-steps), you can verify that Jackal is correctly on your `$PATH` and that you have an 'init' package for your environment and Jackal version.

## Installing the CLI with Homebrew

[Homebrew](https://brew.sh/) is an open-source software package manager that simplifies the installation of software on macOS and Linux.


<Tabs>
<TabItem value="macOS">

With Homebrew on macOS, installing Jackal is as simple as:

```bash
brew tap defenseunicorns/tap && brew install jackal
```

</TabItem>
<TabItem value="Linux">

With Homebrew on Linux, installing Jackal is as simple as:

```bash
brew tap defenseunicorns/tap && brew install jackal
```

</TabItem>
</Tabs>

:::note

The above command detects your OS and system architecture and installs the correct Jackal CLI binary for your machine. Once the above command is entered, the CLI should be installed on your `$PATH` and is ready for immediate use.

:::

## Downloading the CLI from GitHub Releases

All [Jackal releases](https://github.com/defenseunicorns/jackal/releases) on GitHub include prebuilt binaries that you can download and use. We offer range of combinations of OS and architecture for you to choose from.

<Tabs>
<TabItem value="Linux">

To download Jackal on Linux you can run the following (replacing `<jackal-version>` with a version of Jackal):

```bash
JACKAL_VERSION=<jackal-version>
JACKAL_ARCH=$([ $(uname -m) == "x86_64" ] && echo "amd64" || echo "arm64";)

curl -sL https://github.com/defenseunicorns/jackal/releases/download/${JACKAL_VERSION}/jackal_${JACKAL_VERSION}_Linux_${JACKAL_ARCH} -o jackal
chmod +x jackal
```

On most Linux distributions, you can also install the binary onto your `$PATH` by simply moving the downloaded binary to the `/usr/local/bin` directory:

```bash
sudo mv jackal /usr/local/bin/jackal
```

</TabItem>
<TabItem value="macOS">

To download Jackal on macOS you can run the following (replacing `<jackal-version>` with a version of Jackal):

```bash
JACKAL_VERSION=<jackal-version>
JACKAL_ARCH=$([ $(uname -m) == "x86_64" ] && echo "amd64" || echo "arm64";)

curl -sL https://github.com/defenseunicorns/jackal/releases/download/${JACKAL_VERSION}/jackal_${JACKAL_VERSION}_Darwin_${JACKAL_ARCH} -o jackal
chmod +x jackal
```

You can also install the binary onto your `$PATH` by simply moving the downloaded binary to the `/usr/local/bin` directory:

```bash
sudo mv jackal /usr/local/bin/jackal
```

</TabItem>
<TabItem value="Windows">


To download Jackal on Windows you can run the following (replacing `<jackal-version>` with a version of Jackal and `<jackal-arch>` with either `amd64` or `arm64` depending on your system):

```bash
$JackalVersion="<jackal-version>"
$JackalArch="<jackal-arch>"

Start-BitsTransfer -Source "https://github.com/defenseunicorns/jackal/releases/download/$($JackalVersion)/jackal_$($JackalVersion)_Windows_$($JackalArch).exe" -Destination jackal.exe
```

You can also install the binary onto your `$PATH` by moving the downloaded binary to the desired directory and modifying the `$PATH` environment variable to include that directory.

</TabItem>
</Tabs>

## Building the CLI from Scratch

If you want to build the CLI from scratch, you can do that too. Our local builds depend on [Go 1.19.x](https://golang.org/doc/install) and [Node 18.x](https://nodejs.org/en) and are built using [make](https://www.gnu.org/software/make/).

:::note

The `make build-cli` command builds a binary for each combination of OS and architecture. If you want to shorten the build time, you can use an alternative command to only build the binary you need:

- `make build-cli-mac-intel`
- `make build-cli-mac-apple`
- `make build-cli-linux-amd`
- `make build-cli-linux-arm`
- `make build-cli-windows-amd`
- `make build-cli-windows-arm`

For additional information, see the [Building Your Own Jackal CLI](../2-the-jackal-cli/0-building-your-own-cli.md) page.

:::

## Post-Install Steps

Once you have installed Jackal with one of the above methods, you can verify it is working with the following:

```bash
$ jackal version

vX.X.X  # X.X.X is replaced with the version number of your specific installation
```

:::info

If you are not seeing this then Jackal was not installed onto your `$PATH` correctly. [This $PATH guide](https://zwbetz.com/how-to-add-a-binary-to-your-path-on-macos-linux-windows/) should help with that.

:::

In most usage scenarios ([but not all](../../examples/yolo/README.md)) you will also need an ['init' package](../3-create-a-jackal-package/3-jackal-init-package.md) to be initialized before you deploy your own packages. This is a special Jackal package that initializes a cluster with services that are used to store package resources while in the air gap.

You can get the default 'init' package for your version of Jackal by visiting the [Jackal releases](https://github.com/defenseunicorns/jackal/releases) page and downloading it into your working directory or into `~/.jackal-cache/jackal-init-<amd64|arm64>-vX.X.X.tar.zst`.

If you are online on the machine with cluster access you can also run `jackal init` without the `--confirm` flag to be given the option to download the matching version of the default 'init' package or you can use the `jackal tools download-init` command to download a copy to your machine.

:::tip

You can build your own custom 'init' package too if you'd like. For this you should check out the [Creating a Custom 'init' Package Tutorial](../5-jackal-tutorials/8-custom-init-packages.md).

:::

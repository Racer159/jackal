# The Jackal CLI

Jackal is a command line interface (CLI) tool that enables secure software delivery, with a particular focus on delivery to disconnected or highly regulated environments. Jackal is a statically compiled Go binary, which means it can be utilized in any environment without requiring additional dependencies.

## Getting the CLI

You can get the Jackal CLI on your machine in a few different ways, using the Defense Unicorns Homebrew Tap, downloading a prebuilt binary from our GitHub releases, or building the CLI from scratch on your own.

We provide instructions for all of these methods in the [Installing Jackal](../1-getting-started/index.md#installing-jackal) section of the Getting Started guide.

## Introduction to Jackal Commands

Jackal provides a suite of commands that streamline the creation, deployment, and maintenance of packages. Some of these commands contain additional sub-commands to further assist with package management. When executed with the `--help` flag, each command and sub-command provides a concise summary of its functionality. As you navigate deeper into the command hierarchy, the provided descriptions become increasingly detailed. We encourage you to explore the various commands available to gain a comprehensive understanding of Jackal's capabilities.

As previously mentioned, Jackal was specifically designed to facilitate the deployment of applications in disconnected environments with ease. As a result, the most commonly utilized commands are `jackal init`, `jackal package create`, and `jackal package deploy`. Detailed information on all commands can be found in the [CLI Commands](./100-cli-commands/jackal.md) section. However, brief descriptions of the most frequently used commands are provided below. It's worth noting that these three commands are closely linked to what we refer to as a "Jackal Package". Additional information on Jackal Packages can be found on the [Jackal Packages](../3-create-a-jackal-package/1-jackal-packages.md) page.

### jackal init

The `jackal init` command is used to configure a K8s cluster in preparation for the deployment of future Jackal Packages. The init command uses a specialized 'init-package' to operate which may be located in your current working directory, the directory where the Jackal CLI binary is located, or downloaded from the GitHub Container Registry during command execution. For further details regarding the init-package, please refer to the [init-package](../3-create-a-jackal-package/3-jackal-init-package.md) page.

### jackal package deploy

The `jackal package deploy` command is used to deploy an already created Jackal package onto a machine, typically to a K8s cluster. Generally, it is presumed that the `jackal init` command has already been executed on the target machine, however, there are a few exceptional cases where this assumption does not apply.  You can learn more about deploying Jackal packages on the [Deploy a Jackal Package](../4-deploy-a-jackal-package/index.md) page.

:::tip

When deploying and managing packages you may find the sub-commands under `jackal tools` useful to troubleshoot or interact with deployments.

:::

### jackal package create

The `jackal package create` command is used to create a Jackal package from a `jackal.yaml` package definition.  This command will pull all of the defined resources into a single package you can take with you to a disconnected environment.  You can learn more about creating Jackal packages on the [Create a Jackal Package](../3-create-a-jackal-package/index.md) page.

:::tip

When developing packages you may find the sub-commands under `jackal dev` useful to find resources and manipulate package definitions.

:::

# Common CLI Uses

Jackal is a tool that optimizes the delivery of applications and capabilities into various environments, starting with air-gapped systems. This is achieved by using Jackal Packages, which are declarative files that the Jackal CLI uses to create, deploy, inspect, and remove applications and capabilities.

## Building Packages: `jackal package create`

To create a Jackal Package, you must execute the [`jackal package create`](./100-cli-commands/jackal_package_create.md) command, which generates a tarball archive that includes all the required dependencies and instructions to deploy the capabilities onto another machine. The `jackal package create` command uses a [`jackal.yaml` configuration file](../3-create-a-jackal-package/4-jackal-schema.md) that describes the package's components and performs all necessary actions, such as downloading container images and git repositories, to build the final package.

Additional information on Jackal Packages can be found on the [Understanding Jackal Packages](../3-create-a-jackal-package/1-jackal-packages.md) page along with the [Creating a Jackal Package Tutorial](../5-jackal-tutorials/0-creating-a-jackal-package.md).

## Initializing a Cluster: `jackal init`

<!-- TODO: Find a good place to talk about what the init command is doing (there's a lot of special magic sauce going on with that command) -->
<!-- TODO: Should we talk about the 'Jackal Agent - A Mutating Webhook' here? -->

Before deploying a package to a cluster, you must initialize the cluster using the [`jackal init`](./100-cli-commands/jackal_init.md) command. This command creates and bootstraps an in-cluster container registry and provides the option to install optional tools and services necessary for future packages.

For Windows and macOS environments, a cluster must already exist before initializing it using Jackal. You can use [Kind](https://kind.sigs.k8s.io/), [K3d](https://k3d.io/), [Docker Desktop](https://docs.docker.com/desktop/kubernetes/), or any other local or remote Kubernetes cluster.

For Linux environments, Jackal itself can create and update a local K3s cluster, in addition to using any other local or remote Kubernetes cluster. The init package used by `jackal init` contains all the resources necessary to create a local [K3s](https://k3s.io/) cluster on your machine. This package may be located in your current working directory, the directory where the Jackal CLI binary is located, or downloaded from GitHub releases during command execution.

Further details on the initialization process can be found on the [init package](../3-create-a-jackal-package/3-jackal-init-package.md) page along with the [Initializing a K8s Cluster Tutorial](../5-jackal-tutorials/1-initializing-a-k8s-cluster.md).

:::note
Depending on the permissions of your user, if you are installing K3s with `jackal init`, you may need to run it as a privileged user. This can be done by either:

- Becoming a privileged user via the command `sudo su` and then running all the Jackal commands as you normally would.
- Manually running all the Jackal commands as a privileged user via the command `sudo <command>`.
- Running the init command as a privileged user via `sudo jackal init` and then changing the permissions of the `~/.kube/config` file to be readable by the current user.
:::

## Deploying Packages: `jackal package deploy`

<!-- TODO: Write some docs (or redirect to other docs) describing when you would be able to do a `jackal package deploy` before a `jackal init` -->

The [`jackal package deploy`](./100-cli-commands/jackal_package_deploy.md) command deploys the packaged capabilities into the target environment. The package can be deployed on any cluster, even those without an external internet connection, since it includes all of its external resources. The external resources are pushed into the cluster to services Jackal either deployed itself or that it was told about on `init`, such as the init package's Gitea Git server or a pre-existing Harbor image registry.  Then, the application is deployed according to the instructions in the jackal.yaml file, such as deploying a helm chart, deploying raw K8s manifests, or executing a series of shell commands. Generally, it is presumed that the `jackal init` command has already been executed on the target machine. However, there are a few exceptional cases where this assumption does not apply, such as [YOLO Mode](../8-faq.md#what-is-yolo-mode-and-why-would-i-use-it).

Additional information about Jackal Packages can found on the [Understanding Jackal Packages](../3-create-a-jackal-package/1-jackal-packages.md) page along with the [Deploying a Local Jackal Package Tutorial](../5-jackal-tutorials//2-deploying-jackal-packages.md).

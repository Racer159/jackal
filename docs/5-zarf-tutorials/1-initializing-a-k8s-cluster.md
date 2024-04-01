# Initializing a K8s Cluster

## Introduction

In this tutorial, we will demonstrate how to initialize Jackal onto a K8s cluster. This is done by running the [`jackal init`](../2-the-jackal-cli/100-cli-commands/jackal_init.md) command, which uses a specialized package called an 'init-package'. More information about this specific package can be found [here](../3-create-a-jackal-package/3-jackal-init-package.md).

## Prerequisites

Before beginning this tutorial you will need the following:

- The [Jackal](https://github.com/defenseunicorns/jackal) repository cloned: ([`git clone` Instructions](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository))
- Jackal binary installed on your $PATH: ([Installing Jackal](../1-getting-started/index.md#installing-jackal))
- An init-package downloaded: ([init-package Build Instructions](./0-creating-a-jackal-package.md)) or ([Download Location](https://github.com/defenseunicorns/jackal/releases))
- A Kubernetes cluster to work with: ([Local k8s Cluster Instructions](./#setting-up-a-local-kubernetes-cluster))

## Initializing the Cluster

1. Run the `jackal init` command on your cluster.

```sh
$ jackal init
```

2. When prompted to deploy the package select `y` for Yes, then hit the `enter` key. <br/>

3. Decline Optional Components

:::info

More information about the init-package and its components can be found [here](../3-create-a-jackal-package/3-jackal-init-package.md)

:::

<iframe src="/docs/tutorials/jackal_init.html" height="800px" width="100%"></iframe>

:::note
You will only be prompted to deploy the k3s component if you are on a Linux machine
:::

### Validating the Deployment
After the `jackal init` command is done running, you should see a few new `jackal` pods in the Kubernetes cluster.

```bash
jackal tools monitor

# Note you can press `0` if you want to see all namespaces and CTRL-C to exit
```
![Jackal Tools Monitor](../.images/tutorials/jackal_tools_monitor.png)

## Cleaning Up

The [`jackal destroy`](../2-the-jackal-cli/100-cli-commands/jackal_destroy.md) command will remove all of the resources that were created by the initialization command. This command will leave you with a clean cluster that you can either destroy or use for another tutorial.

```sh
jackal destroy --confirm
```

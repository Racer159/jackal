---
title: Initializing a K8s Cluster
---

## Introduction

In this tutorial, we will demonstrate how to initialize Zarf onto a K8s cluster. This is done by running the [`zarf init`](../2-the-zarf-cli/100-cli-commands/zarf_init.md) command, which uses a specialized package called an 'init-package'. More information about this specific package can be found [here](../3-create-a-zarf-package/3-zarf-init-package.md).

## Prerequisites

Before beginning this tutorial you will need the following:

- The [Zarf](https://github.com/defenseunicorns/zarf) repository cloned: ([`git clone` Instructions](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository))
- Zarf binary installed on your $PATH: ([Installing Zarf](../1-getting-started/index.md#installing-zarf))
- An init-package downloaded: ([init-package Build Instructions](./0-creating-a-zarf-package.md)) or ([Download Location](https://github.com/defenseunicorns/zarf/releases))
- A Kubernetes cluster to work with: ([Local k8s Cluster Instructions](./#setting-up-a-local-kubernetes-cluster))

## Initializing the Cluster

1. Run the `zarf init` command on your cluster.

```sh
$ zarf init
```

2. When prompted to deploy the package select `y` for Yes, then hit the `enter` key. <br/>

3. Decline Optional Components

:::info

More information about the init-package and its components can be found [here](../3-create-a-zarf-package/3-zarf-init-package.md)

:::

<iframe src="/docs/tutorials/zarf_init.html" height="800px" width="100%"></iframe>

:::note
You will only be prompted to deploy the k3s component if you are on a Linux machine
:::

### Validating the Deployment
After the `zarf init` command is done running, you should see a few new `zarf` pods in the Kubernetes cluster.

```bash
zarf tools monitor

# Note you can press `0` if you want to see all namespaces and CTRL-C to exit
```
![Zarf Tools Monitor](../.images/tutorials/zarf_tools_monitor.png)

## Cleaning Up

The [`zarf destroy`](../2-the-zarf-cli/100-cli-commands/zarf_destroy.md) command will remove all of the resources that were created by the initialization command. This command will leave you with a clean cluster that you can either destroy or use for another tutorial.

```sh
zarf destroy --confirm
```

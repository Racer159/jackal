# Creating a K8s Cluster with Jackal

In this tutorial, we will demonstrate how to use Jackal on a fresh Linux machine to deploy a [k3s](https://k3s.io/) cluster through Jackal's `k3s` component.

## System Requirements

-  `root` access on a Linux machine

:::info REQUIRES ROOT
The 'k3s' component requires root access (not just `sudo`!) when deploying as it will modify your host machine to install the cluster.
:::

## Prerequisites

Before beginning this tutorial you will need the following:

- The [Jackal](https://github.com/defenseunicorns/jackal) repository cloned: ([`git clone` Instructions](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository))
- Jackal binary installed on your $PATH: ([Installing Jackal](../1-getting-started/index.md#installing-jackal))
- An init-package built/downloaded: ([init-package Build Instructions](./0-creating-a-jackal-package.md)) or ([Download Location](https://github.com/defenseunicorns/jackal/releases))

## Creating the Cluster

1. Run the `jackal init` command as `root`.

```sh
# jackal init
```

2. Confirm Package Deployment: <br/>
- When prompted to deploy the package select `y` for Yes, then hit the `enter` key. <br/>

3. Confirm k3s Component Deployment: <br/>
- When prompted to deploy the k3s component select `y` for Yes, then hit the `enter` key.

<iframe src="/docs/tutorials/k3s_init.html" height="750px" width="100%"></iframe>

:::tip
You can automatically accept the k3s component and confirm the package using the `--components` and `--confirm` flags.

```sh
$ jackal init --components="k3s" --confirm
```
:::

### Validating the Deployment
After the `jackal init` command is done running, you should see a k3s cluster running and a few `jackal` pods in the Kubernetes cluster.

```sh
# jackal tools monitor
```
:::note
You can press `0` if you want to see all namespaces and CTRL-C to exit
:::

### Accessing the Cluster as a Normal User
By default, the k3s component will only automatically provide cluster access to the root user. To access the cluster as another user, you can run the following to setup the `~/.kube/config` file:

```sh
# cp /root/.kube/config /home/otheruser/.kube
# chown otheruser /home/otheruser/.kube/config
# chgrp otheruser /home/otheruser/.kube/config
```

## Cleaning Up

The [`jackal destroy`](../2-the-jackal-cli/100-cli-commands/jackal_destroy.md) command will remove all of the resources, including the k3s cluster, that was created by the initialization command.

```sh
jackal destroy --confirm
```

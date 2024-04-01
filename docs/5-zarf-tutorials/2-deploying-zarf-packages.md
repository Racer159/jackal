# Deploying Local Jackal Packages

## Introduction

In this tutorial, we are going to deploy the WordPress package onto your cluster using the package we created in the earlier [create a package](./0-creating-a-jackal-package.md) tutorial and the cluster we initialized in the [initialize a k8s cluster](./1-initializing-a-k8s-cluster.md) tutorial. We will be leveraging that past work to go the extra step of deploying an application we packaged onto our cluster with the `jackal package deploy` command.

## System Requirements

- You'll need a machine that has access to a built-package and an initialized cluster.

## Prerequisites

Prior to this tutorial you'll want to have a built package and a working cluster with Jackal initialized.

- Jackal binary installed on your $PATH: ([Installing Jackal](../1-getting-started/index.md#installing-jackal))
- [An initialized cluster](./1-initializing-a-k8s-cluster.md)
- The [WordPress package created](./0-creating-a-jackal-package.md)

## Deploying the WordPress package

1. Use the `jackal package deploy` command to deploy the package you built in a the previous tutorial (see [prerequisites](#prerequisites)).

<iframe src="/docs/tutorials/package_deploy_wordpress.html" width="100%" height="550px"></iframe>

:::note

If you do not provide the path to the package as an argument to the `jackal package deploy` command, Jackal will prompt you asking for you to choose which package you want to deploy. You can use the `tab` key, to be prompted for available packages in the current working directory.

<iframe src="/docs/tutorials/package_deploy_suggest.html" width="100%" height="120px"></iframe>

By hitting 'tab', you can use the arrow keys to select which package you want to deploy. Since we are deploying the WordPress package in this tutorial, we will select that package and hit 'enter'.

<iframe src="/docs/tutorials/package_deploy_wordpress_suggestions.html" width="100%" height="150px"></iframe>

:::

2. You will be presented with a chance to review the SBOMs for the package along with its definition followed by a series of prompts for each variable we setup in the [previous tutorial](./0-creating-a-jackal-package.md#setting-up-variables).  To confirm package deployment press `y` then `enter` and input a value for each variable when prompted followed by `enter` for them as well.

:::tip

To accept a default value for a given variable, simply press the `enter` key.  You can also set variables from the CLI with the `--set` flag, an environment variable, or a [config file](../2-the-jackal-cli/index.md#using-a-config-file-to-make-cli-command-flags-declarative).

:::

<iframe src="/docs/tutorials/package_deploy_wordpress_bottom.html" width="100%" height="690px"></iframe>

3. Because we included the connect services in the [previous tutorial](./0-creating-a-jackal-package.md#setting-up-a-jackal-connect-service) we can quickly test our package in a browser with `jackal connect wordpress-blog`.

![Jackal Connect WordPress](../.images/tutorials/wordpress_connected.png)

4. We can also explore the resources deployed by our package by running the `jackal tools monitor` command to start [`K9s`](../4-deploy-a-jackal-package/5-k9s-dashboard.md). Once you are done, hit `ctrl/control c` to exit.

![Jackal Tools Monitor](../.images/tutorials/jackal_tools_monitor.png)

:::tip

Deploying packages isn't the only way to interact with them in the air gap.  If you would like to quickly inspect a package and it's SBOMs you can use [`jackal package inspect`](../4-deploy-a-jackal-package/4-view-sboms.md) to view them, and if you would like to push resources inside of a Jackal package (i.e. the images in this Wordpress package) to services in the air gap without running a deployment, you can do so with [`jackal package mirror-resources`](../2-the-jackal-cli/100-cli-commands/jackal_package_mirror-resources.md).

:::

## Removal

1. Use the `jackal package list` command to get a list of the installed packages.  This will give you the name of the WordPress package to remove it.

<iframe src="/docs/tutorials/package_deploy_wordpress_list.html" height="120px" width="100%"></iframe>

2. Use the `jackal package remove` command to remove the `wordpress` package.  Don't forget the `--confirm` flag.  Otherwise you'll receive an error.

<iframe src="/docs/tutorials/package_deploy_wordpress_no_confirm.html" width="100%" height="425px"></iframe>

3. You can also use the `jackal package remove` command with the jackal package file, to remove the package.  Again, don't forget the `--confirm` flag.

<iframe src="/docs/tutorials/package_deploy_wordpress_remove_by_file.html" height="100px" width="100%"></iframe>

The `wordpress` package has now been removed from your cluster.

## Troubleshooting

### Unable to connect to the Kubernetes cluster

<iframe src="/docs/tutorials/troubleshoot_unreachable.html" width="100%" height="200px"></iframe>

:::info Remediation

If you receive this error, either you don't have a Kubernetes cluster, your cluster is down, or your cluster is unreachable.

1. Check your kubectl configuration, then try again.  For more information about kubectl configuration see [Configure Access to Multiple Clusters](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) from the Kubernetes documentation.

If you need to setup a cluster, you can perform the following.

1. Deploy a Kubernetes cluster with the [Creating a K8s Cluster with Jackal](./5-creating-a-k8s-cluster-with-jackal.md) tutorial.
2. Perform the [Initialize a cluster](./1-initializing-a-k8s-cluster.md) tutorial.

After that you can try deploying the package again.

:::

### Secrets "jackal-state" not found

<iframe src="/docs/tutorials/troubleshoot_uninitialized_helmOCI.html" width="100%" height="250px"></iframe>

:::info Remediation

If you receive this error when jackal is attempting to deploy a package, this means you have not initialized the kubernetes cluster.  This is one of the prerequisites for this tutorial.  Perform the [Initialize a cluster](./1-initializing-a-k8s-cluster.md) tutorial, then try again.

:::

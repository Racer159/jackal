# Jackal Tutorials

This section of the documentation has a collection of tutorials that will help you get more familiar with Jackal and its features. The tutorials assume that you have a very basic understanding of what Jackal is and aims to help expand your working knowledge of how to use Jackal and what Jackal is capable of doing.

## Tutorial Prerequisites
If a tutorial has any prerequisites, they will be listed at the beginning of the tutorial with instructions on how to fulfill them.
Almost all tutorials will have the following prerequisites/assumptions:

1. The [Jackal](https://github.com/racer159/jackal) repository cloned: ([git clone instructions](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository))
1. You have a Jackal binary installed on your $PATH: ([Installing Jackal](../1-getting-started/index.md#installing-jackal))
1. You have an init-package built/downloaded: ([init-package Build Instructions](./0-creating-a-jackal-package.md)) or ([Download Location](https://github.com/racer159/jackal/releases))
1. Have a kubernetes cluster running/available (ex. [k3s](https://k3s.io/)/[k3d](https://k3d.io/v5.4.1/)/[Kind](https://kind.sigs.k8s.io/docs/user/quick-start#installation))

## Setting Up a Local Kubernetes Cluster

While Jackal is able to deploy a local k3s Kubernetes cluster for you, (as you'll find out more in the [Creating a K8s Cluster with Jackal](./5-creating-a-k8s-cluster-with-jackal.md) tutorial), that k3s cluster will only work if you are on a root user on a Linux machine. If you are on a Mac, or you're on Linux but don't have root access, you'll need to set up a local dockerized Kubernetes cluster manually. We provide instructions on how to quickly set up a local k3d cluster that you can use for the majority of the tutorials.

### Install k3d

1. Install Docker: [Docker Install Instructions](https://docs.docker.com/get-docker/)
2. Install k3d: [k3d Install Instructions](https://k3d.io/#installation)

### Start up k3d cluster

```bash
k3d cluster create      # Creates a k3d cluster
                        # This will take a couple of minutes to complete


jackal tools kubectl get pods -A    # Check to see if the cluster is ready
```

### Tear Down k3d CLuster

```bash
k3d cluster delete      # Deletes the k3d cluster
```

# Jackal Packages

This folder contains packages maintained by the [Jackal team](https://github.com/racer159/jackal/graphs/contributors).  Some of these packages are used by `jackal init` for new cluster initialization.

**Packages**
- [Jackal Packages](#jackal-packages)
    - [Distros](#distros)
      - [Usage Examples](#usage-examples)
    - [Gitea](#gitea)
    - [Logging PGL](#logging-pgl)
    - [Jackal Agent](#jackal-agent)
    - [Jackal Registry](#jackal-registry)

### Distros

The distros package adds optional capabilities for spinning up and tearing down clusters.  Currently, the following distros are supported:

- [EKS](https://aws.amazon.com/eks/) - Jackal deploys and tears down using the `eksctl` binary under the hood. See how it's done in the EKS package's [`jackal.yaml`](./distros/eks/jackal.yaml) and checkout the [EKS package's config](./distros/eks/eks.yaml) for more information.

- [k3s](https://k3s.io/) - Jackal deploys and tears down using the `k3s` service under the hood. See how it's done in the k3s package's [`jackal.yaml`](./distros/k3s/common/jackal.yaml).


#### Usage Examples  

**EKS**  - Create/Deploy EKS cluster.  

> **Note** - requires `eksctl` credentials.

```bash
jackal package create packages/distros/eks -o build --confirm

jackal package deploy build/jackal-package-distro-eks-amd64-x.x.x.tar.zst --components=deploy-eks-cluster --set=CLUSTER_NAME='jackal-nightly-eks-e2e-test',INSTANCE_TYPE='t3.medium' --confirm
```

See the [nightly-eks test](../.github/workflows/nightly-eks.yml) for another example.

**k3s** - Create/Deploy a k3s cluster.  

> **Note** - requires `systemd` and `root` access only (no `sudo`) on a linux machine.

```bash
jackal init --components=k3s
```

### Gitea

Users who rely heavily on GitOps find it useful to deploy an internal Git repository.  Jackal uses [Gitea](https://gitea.io/en-us/) to provide this functionality.  The Gitea package deploys a Gitea instance to the cluster and configures it to use the credentials in the `private-git-server` secret in the Jackal namespace.

_usage_

```bash
jackal init --components=git-server
```

### Logging PGL

The Logging PGL package deploys the Promtail, Grafana, and Loki stack which aggregates logs from different containers and presents them in a web dashboard.  This is useful as a quick way to get logging into a cluster when you otherwise wouldn't be bringing over a logging stack.

_usage_

```bash
jackal init --components=logging
```

### Jackal Agent

The Jackal Agent is a mutating admission controller used to modify the image property within a PodSpec. The purpose is to redirect it to Jackal's configured registry instead of the the original registry (such as DockerHub, GHCR, or Quay). Additionally, the webhook attaches the appropriate `ImagePullSecret` for the seed registry to the pod. This configuration allows the pod to successfully retrieve the image from the seed registry, even when operating in an air-gapped environment.

```bash
$ jackal tools kubectl get deploy -n jackal agent-hook

NAME         READY   UP-TO-DATE   AVAILABLE   AGE
agent-hook   2/2     2            2           17m
```

### Jackal Registry

The Jackal internal registry is utilized to store container images for use in air-gapped environments.  The registry is deployed as a `Deployment` with a single replica and  a `PersistentVolumeClaim` to store the images.  Credentials for basic authentication are autogenerated and stored within a secret in the `jackal` namespace. The internal registry is `HTTP` only.

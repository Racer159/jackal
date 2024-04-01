---
sidebar_position: 8
---

# FAQ

## Who is behind this project?

Jackal was built by the developers at [Defense Unicorns](https://www.defenseunicorns.com/) and an amazing community of contributors.

Defense Unicorns' mission is to advance freedom and independence globally through Free and Open Source software.

## What license is Jackal under?

Jackal is under the [Apache License 2.0](https://github.com/Racer159/jackal/blob/main/LICENSE). This is one of the most commonly used licenses for open-source software.

## Is Jackal free to use?

Yes! Jackal is Free and Open-Source Software (FOSS). And will remain free forever. We believe Free and Open Source software changes the world and promotes freedom and security. Anyone who sees the value in our tool should be free to use it without fear of vendor locking or licensing fees.

## Do I have to use Homebrew to install Jackal?

No, the Jackal binary and init package can be downloaded from the [Releases Page](https://github.com/Racer159/jackal/releases). Jackal does not need to be installed or available to all users on the system, but it does need to be executable for the current user (i.e. `chmod +x jackal` for Linux/Mac).

## What dependencies does Jackal have?

Jackal is statically compiled and written in [Go](https://golang.org/) and [Rust](https://www.rust-lang.org/), so it has no external dependencies. For Linux, Jackal can bring a Kubernetes cluster using [K3s](https://k3s.io/). For Mac and Windows, Jackal can leverage any available local or remote cluster the user has access to. Currently, the K3s installation Jackal performs does require a [Systemd](https://en.wikipedia.org/wiki/Systemd) based system and `root` (not just `sudo`) access.

## What is the Jackal Agent?

The Jackal Agent is a [Kubernetes Mutating Webhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) that is installed into the cluster during `jackal init`. The Agent is responsible for modifying [Kubernetes PodSpec](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#PodSpec) objects [Image](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#Container.Image) fields to point to the Jackal Registry. This allows the cluster to pull images from the Jackal Registry instead of the internet without having to modify the original image references. The Agent also modifies [Flux GitRepository](https://fluxcd.io/docs/components/source/gitrepositories/) objects to point to the local Git Server.

## Why doesn't the Jackal Agent create secrets it needs in the cluster?

During early discussions and [subsequent decision](../adr/0005-mutating-webhook.md) to use a Mutating Webhook, we decided to not have the Agent create any secrets in the cluster. This is to avoid the Agent having to have more privileges than it needs as well as to avoid collisions with Helm. The Agent today simply responds to requests to patch PodSpec and GitRepository objects.

The Agent does not need to create any secrets in the cluster. Instead, during `jackal init` and `jackal package deploy`, secrets are automatically created as [Helm Postrender Hook](https://helm.sh/docs/topics/advanced/#post-rendering) for any namespaces Jackal sees. If you have resources managed by [Flux](https://fluxcd.io/) that are not in a namespace managed by Jackal, you can either create the secrets manually or include a manifest to create the namespace in your package and let Jackal create the secrets for you.

## How can a Kubernetes resource be excluded from the Jackal Agent?

Resources can be excluded at the namespace or resources level by adding the `jackal.dev/agent: ignore` label.

## What happens to resources that exist in the cluster before `jackal init`?

During the `jackal init` operation, the Jackal Agent will patch any existing namespaces with the `jackal.dev/agent: ignore` label to prevent the Agent from modifying any resources in that namespace. This is done because there is no way to guarantee the images used by pods in existing namespaces are available in the Jackal Registry.

If you would like to adopt pre-existing resources into a Jackal deployment you can use the `--adopt-existing-resources` flag on [`jackal package deploy`](./2-the-jackal-cli/100-cli-commands/jackal_package_deploy.md) to adopt those resources into the Helm Releases that Jackal manages (including namespaces).  This will add the requisite annotations and labels to those resources and drop the `jackal.dev/agent: ignore` label from any namespaces specified by those resources.

:::note

Jackal will refuse to adopt the Kubernetes [initial namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/#initial-namespaces).  It is recommended that you do not deploy resources into the `default` or `kube-*` namespaces with Jackal.

Additionally, when adopting resources, you should ensure that the namespaces you are adopting are dedicated to Jackal, or that you go back and manually add the `jackal.dev/agent: ignore` label to any non-Jackal managed resources in those namespaces (and ensure that updates to those resources do not strip that label) otherwise you may see [ImagePullBackOff](https://kubernetes.io/docs/concepts/containers/images/#imagepullbackoff) errors.

:::

## How can I improve the speed of loading large images from Docker on `jackal package create`?

Due to some limitations with how Docker provides access to local image layers, `jackal package create` has to rely on `docker save` under the hood which is [very slow overall](https://github.com/Racer159/jackal/issues/1214) and also takes a long time to report progress. We experimented with many ways to improve this, but for now recommend leveraging a local docker registry to speed up the process.

This can be done by running a local registry and pushing the images to it before running `jackal package create`. This will allow `jackal package create` to pull the images from the local registry instead of Docker. This can also be combined with [component actions](3-create-a-jackal-package/7-component-actions.md) and [`--registry-override`](./2-the-jackal-cli/100-cli-commands/jackal_package_create.md) to make the process automatic. Given an example image of `registry.enterprise.corp/my-giant-image:v2` you could do something like this:

```sh
# Create a local registry
docker run -d -p 5000:5000 --restart=always --name registry registry:2

# Run the package create with a tag variable
jackal package create --registry-override registry.enterprise.corp=localhost:5000 --set IMG=my-giant-image:v2
```

```yaml
kind: JackalPackageConfig
metadata:
  name: giant-image-example

components:
  - name: main
    actions:
      # runs during "jackal package create"
      onCreate:
        # runs before the component is created
        before:
          - cmd: 'docker tag registry.enterprise.corp/###JACKAL_PKG_TMPL_IMG### localhost:5000/###JACKAL_PKG_TMPL_IMG###'
          - cmd: 'docker push localhost:5000/###JACKAL_PKG_TMPL_IMG###'

    images:
      - 'registry.enterprise.corp/###JACKAL_PKG_TMPL_IMG###'
```

## Can I pull in more than http(s) git repos on `jackal package create`?

Under the hood, Jackal uses [`go-git`](https://github.com/go-git/go-git) to perform `git` operations, but it can fallback to `git` located on the host and thus supports any of the [git protocols](https://git-scm.com/book/en/v2/Git-on-the-Server-The-Protocols) available.  All you need to use a different protocol is to specify the full URL for that particular repo:

:::note

In order for the fallback to work correctly you must have `git` version `2.14` or later in your path.

:::

```yaml
kind: JackalPackageConfig
metadata:
  name: repo-schemes-example

components:
    repos:
      - https://github.com/Racer159/jackal.git
      - ssh://git@github.com/Racer159/jackal.git
      - file:///home/jackal/workspace/jackal
      - git://somegithost.com/jackal.git
```

In the airgap, Jackal with rewrite these URLs to match the scheme and host of the provided airgap `git` server.

:::note

When specifying other schemes in Jackal you must change the consuming side as well since Jackal will add a CRC hash of the URL to the repo name on the airgap side.  This is to reduce the chance for collisions between repos with similar names.  This means an example Flux `GitRepository` specification would look like this for the `file://` based pull:

```yaml
---
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: GitRepository
metadata:
  name: podinfo
  namespace: flux-system
spec:
  interval: 30s
  ref:
    tag: 6.1.6
  url: file:///home/jackal/workspace/podinfo
```

:::

## What is YOLO Mode and why would I use it?

YOLO Mode is a special package metadata designation that be added to a package prior to `jackal package create` to allow the package to be installed without the need for a `jackal init` operation. In most cases this will not be used, but it can be useful for testing or for environments that manage their own registries and Git servers completely outside of Jackal. This can also be used as a way to transition slowly to using Jackal without having to do a full migration.

:::note

Typically you should not deploy a Jackal package in YOLO mode if the cluster has already been initialized with Jackal. This could lead to an [ImagePullBackOff](https://kubernetes.io/docs/concepts/containers/images/#imagepullbackoff) if the resources in the package do not include the `jackal.dev/agent: ignore` label and are not already available in the Jackal Registry.

:::

## What is a `skeleton` Jackal Package?

A `skeleton` package is a bare-bones Jackal package definition alongside its associated local files and manifests that has been published to an OCI registry.  These packages are intended for use with [component composability](../examples/composable-packages/README.md) to provide versioned imports for components that you wish to mix and match or modify with merge-overrides across multiple separate packages.

Skeleton packages have not been run through the `jackal package create` process yet, and thus do not have any remote resources included (no images, repos, or remote manifests and files) thereby retaining any [create-time package configuration templates](../examples/variables/README.md#create-time-package-configuration-templates) as they were defined in the original `jackal.yaml` (i.e. untemplated).

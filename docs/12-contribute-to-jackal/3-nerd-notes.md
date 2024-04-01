import ArchitectureSVG from '../.images/architecture.drawio.svg';

# Jackal Nerd Notes

:::caution Hard Hat Area
This page is still being developed. More content will be added soon!
:::

Jackal is written entirely in [go](https://go.dev/), except for a single 868Kb binary for the injector system written in [rust](https://www.rust-lang.org/), so we can fit it in a [configmap](https://kubernetes.io/docs/concepts/configuration/configmap/). All assets are bundled together into a single [zstd](https://facebook.github.io/zstd/) tarball on each `jackal package create` operation. On the air gap / offline side, `jackal package deploy` extracts the various assets and places them on the filesystem or installs them in the cluster, depending on what the jackal package says to do. Some important ideas behind Jackal:

- All workloads are installed in the cluster via the [Helm SDK](https://helm.sh/docs/topics/advanced/#go-sdk)
- The OCI Registries used are both from [Docker](https://github.com/distribution/distribution)
- Currently, the Registry and Git servers _are not HA_, see [#375](https://github.com/Racer159/jackal/issues/376) and [#376](https://github.com/Racer159/jackal/issues/376) for discussion on this
- To avoid TLS issues, Jackal binds to `127.0.0.1:31999` on each node as a [NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport) to allow all nodes to access the pod(s) in the cluster
- Jackal utilizes a [mutating admission webhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) called the [`jackal-agent`](https://github.com/Racer159/jackal/tree/main/src/internal/agent) to modify the image property within the `PodSpec`. The purpose is to redirect it to Jackal's configured registry instead of the the original registry (such as DockerHub, GCR, or Quay). Additionally, the webhook attaches the appropriate [ImagePullSecret](https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod) for the seed registry to the pod. This configuration allows the pod to successfully retrieve the image from the seed registry, even when operating in an air-gapped environment.
- Jackal uses a custom injector system to bootstrap a new cluster. See the PR [#329](https://github.com/Racer159/jackal/pull/329) and [ADR](https://github.com/Racer159/jackal/blob/main/adr/0003-image-injection-into-remote-clusters-without-native-support.md) for more details on how we came to this solution.  The general steps are listed below:
  - Get a list of images in the cluster
  - Attempt to create an ephemeral pod using an image from the list
  - A small rust binary that is compiled using [musl](https://www.musl-libc.org/) to keep the max binary size as minimal as possible
  - The `registry:2` image is placed in a tar archive and split into 512 KB chunks; larger sizes tended to cause latency issues on low-resource control planes
  - An init container runs the rust binary to re-assemble and extract the jackal binary and registry image
  - The container then starts and runs the rust binary to host the registry image in an static docker registry
  - After this, the main docker registry chart is deployed, pulls the image from the ephemeral pod, and finally destroys the created configmaps, pod, and service

## Jackal Architecture

<ArchitectureSVG />

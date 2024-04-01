# Jackal - DevSecOps for Air Gap

[![Latest Release](https://img.shields.io/github/v/release/defenseunicorns/jackal)](https://github.com/defenseunicorns/jackal/releases)
[![Go version](https://img.shields.io/github/go-mod/go-version/defenseunicorns/jackal?filename=go.mod)](https://go.dev/)
[![Build Status](https://img.shields.io/github/actions/workflow/status/defenseunicorns/jackal/release.yml)](https://github.com/defenseunicorns/jackal/actions/workflows/release.yml)
[![Jackal Documentation Status](https://api.netlify.com/api/v1/badges/fe846ae4-25fb-4274-9968-90782640ee9f/deploy-status)](https://app.netlify.com/sites/jackal-docs/deploys)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/defenseunicorns/jackal/badge)](https://api.securityscorecards.dev/projects/github.com/defenseunicorns/jackal)

<img align="right" alt="jackal logo" src=".images/jackal-logo.png"  height="256" />

[![Jackal Website](https://img.shields.io/badge/web-jackal.dev-6d87c3)](https://jackal.dev/)
[![Jackal Documentation](https://img.shields.io/badge/docs-docs.jackal.dev-775ba1)](https://docs.jackal.dev/)
[![Jackal Slack Channel](https://img.shields.io/badge/k8s%20slack-jackal-40a3dd)](https://kubernetes.slack.com/archives/C03B6BJAUJ3)
[![Community Meetups](https://img.shields.io/badge/community-meetups-22aebb)](https://github.com/defenseunicorns/jackal/issues/2202)

Jackal eliminates the [complexity of air gap software delivery](https://www.itopstimes.com/contain/air-gap-kubernetes-considerations-for-running-cloud-native-applications-without-the-cloud/) for Kubernetes clusters and cloud-native workloads using a declarative packaging strategy to support DevSecOps in offline and semi-connected environments.

## Why Use Jackal

- 💸 **Free and Open-Source.** Jackal will always be free to use and maintained by the open-source community.
- ⭐️ **Zero Dependencies.** As a statically compiled binary, the Jackal CLI has zero dependencies to run on any machine.
- 🔓 **No Vendor Lock.** There is no proprietary software that locks you into using Jackal. If you want to remove it, you can still use your Helm charts to deploy your software manually.
- 💻 **OS Agnostic.** Jackal supports numerous operating systems. A full matrix of supported OSes, architectures, and feature sets is coming soon.
- 📦 **Highly Distributable.** Integrate and deploy software from multiple secure development environments, including edge, embedded systems, secure cloud, data centers, and even local environments.
- 🚀 **Develop Connected, Deploy Disconnected.** Teams can build and configure individual applications or entire DevSecOps environments while connected to the internet. Once created, they can be packaged and shipped to a disconnected environment to be deployed.
- 💿 **Single File Deployments.** Jackal allows you to package the parts of the internet your app needs into a single compressed file to be installed without connectivity.
- ♻️ **Declarative Deployments.** Jackal packages define the precise state for your application, enabling it to be deployed the same way every time.
- 🦖 **Inherit Legacy Code.** Jackal packages can wrap legacy code and projects - allowing them to be deployed to modern DevSecOps environments.

## 📦 Out of the Box Features

- Automate Kubernetes deployments in disconnected environments
- Automate [Software Bill of Materials (SBOM)](https://docs.jackal.dev/docs/create-a-jackal-package/package-sboms) generation
- Build and [publish packages as OCI image artifacts](https://docs.jackal.dev/docs/jackal-tutorials/publish-and-deploy)
- Provide a [web dashboard](https://docs.jackal.dev/docs/deploy-a-jackal-package/view-sboms) for viewing SBOM output
- Create and verify package signatures with [cosign](https://github.com/sigstore/cosign)
- [Publish](https://docs.jackal.dev/docs/the-jackal-cli/cli-commands/jackal_package_publish), [pull](https://docs.jackal.dev/docs/the-jackal-cli/cli-commands/jackal_package_pull), and [deploy](https://docs.jackal.dev/docs/the-jackal-cli/cli-commands/jackal_package_deploy) packages from an [OCI registry](https://opencontainers.org/)
- Powerful component lifecycle [actions](https://docs.jackal.dev/docs/create-a-jackal-package/component-actions)
- Deploy a new cluster while fully disconnected with [K3s](https://k3s.io/) or into any existing cluster using a [kube config](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/)
- Builtin logging stack with [Loki](https://grafana.com/oss/loki/)
- Built-in Git server with [Gitea](https://gitea.io/en-us/)
- Built-in Docker registry
- Builtin [K9s Dashboard](https://k9scli.io/) for managing a cluster from the terminal
- [Mutating Webhook](adr/0005-mutating-webhook.md) to automatically update Kubernetes pod's image path and pull secrets as well as [Flux Git Repository](https://fluxcd.io/docs/components/source/gitrepositories/) URLs and secret references
- Builtin [command to find images](https://docs.jackal.dev/docs/the-jackal-cli/cli-commands/jackal_dev_find-images) and resources from a Helm chart
- Tunneling capability to [connect to Kubernetes resources](https://docs.jackal.dev/docs/the-jackal-cli/cli-commands/jackal_connect) without network routing, DNS, TLS or Ingress configuration required

## 🛠️ Configurable Features

- Customizable [variables and package templates](https://docs.jackal.dev/examples/variables/) with defaults and user prompting
- [Composable packages](https://docs.jackal.dev/docs/create-a-jackal-package/jackal-components#composing-package-components) to include multiple sub-packages/components
- Component-level OS/architecture filtering

## Demo

[![preview](.images/jackal-v0.21-preview.gif)](https://www.youtube.com/watch?v=WnOYlFVVKDE)

_<https://www.youtube.com/watch?v=WnOYlFVVKDE>_

## ✅ Getting Started

To try Jackal out for yourself, visit the ["Try It Now"](https://jackal.dev/install) section on our website.

To learn more about Jackal and its use cases, visit [docs.jackal.dev](https://docs.jackal.dev/docs/jackal-overview). From the docs, you can learn more about:

- [installation](https://docs.jackal.dev/docs/getting-started/#installing-jackal)
- [using the CLI](https://docs.jackal.dev/docs/the-jackal-cli/),
- [making packages](https://docs.jackal.dev/docs/create-a-jackal-package/jackal-packages/),
- [Jackal package schema](https://docs.jackal.dev/docs/create-a-jackal-package/jackal-schema).

Using Jackal in GitHub workflows? Check out the [setup-jackal](https://github.com/defenseunicorns/setup-jackal) action. Install any version of Jackal and its `init` package with zero added dependencies.

## 🫶 Our Community

Join our community and developers on the [#Jackal slack](https://jackal.dev/slack) hosted on K8s slack. Our active community of developers, users, and contributors are available to answer questions, share examples, and find new ways use Jackal together!

We are so grateful to our Jackal community for contributing bug fixes and collaborating on new features:

<a href="https://github.com/defenseunicorns/jackal/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=defenseunicorns/jackal" />
</a>

Made with [contrib.rocks](https://contrib.rocks).

## 💻 Contributing

Want to contribute to Jackal?
Check out our [Contributor Guide](https://docs.jackal.dev/docs/contribute-to-jackal/contributor-guide) to learn more about how to set up your development environment and begin contributing.
We also recommend checking out our architectural diagram.

To dive deeper into the tech, you can read the [Nerd Notes](https://docs.jackal.dev/docs/contribute-to-jackal/nerd-notes) in our Docs.

![Architecture Diagram](./docs/.images/architecture.drawio.svg)

[Source DrawIO](docs/.images/architecture.drawio.svg)

## ⭐️ Special Thanks

> Early Jackal research and prototypes were developed jointly with [United States Naval Postgraduate School](https://nps.edu/) research you can read [here](https://calhoun.nps.edu/handle/10945/68688).

We would also like to thank the following awesome libraries and projects without which Jackal would not be possible!

[![pterm/pterm](https://img.shields.io/badge/pterm%2Fpterm-007d9c?logo=go&logoColor=white)](https://github.com/pterm/pterm)
[![mholt/archiver](https://img.shields.io/badge/mholt%2Farchiver-007d9c?logo=go&logoColor=white)](https://github.com/mholt/archiver)
[![spf13/cobra](https://img.shields.io/badge/spf13%2Fcobra-007d9c?logo=go&logoColor=white)](https://github.com/spf13/cobra)
[![go-git/go-git](https://img.shields.io/badge/go--git%2Fgo--git-007d9c?logo=go&logoColor=white)](https://github.com/go-git/go-git)
[![sigstore/cosign](https://img.shields.io/badge/sigstore%2Fcosign-2a1e71?logo=linuxfoundation&logoColor=white)](https://github.com/sigstore/cosign)
[![helm.sh/helm](https://img.shields.io/badge/helm.sh%2Fhelm-0f1689?logo=helm&logoColor=white)](https://github.com/helm/helm)
[![kubernetes](https://img.shields.io/badge/kubernetes-316ce6?logo=kubernetes&logoColor=white)](https://github.com/kubernetes)

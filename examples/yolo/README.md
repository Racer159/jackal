import ExampleYAML from "@site/src/components/ExampleYAML";

# YOLO Mode

This example demonstrates YOLO mode, an optional mode for using Jackal in a fully connected environment where users can bring their own external container registry and Git server.

## Prerequisites

- A running K8s cluster.

:::note

The cluster does not need to have the Jackal init package installed or any other Jackal-related bootstrapping.

:::

## Instructions

Create the package:

```bash
jackal package create
```

### Deploy the package

```bash
# Run the following command to deploy the created package to the cluster
jackal package deploy

# Choose the yolo package from the list
? Choose or type the package file [tab for suggestions]
> jackal-package-yolo-<ARCH>.tar.zst

# Confirm the deployment
? Deploy this Jackal package? (y/N)

# Wait a few seconds for the cluster to deploy the package; you should
# see the following output when the package has been finished deploying:
  Connect Command    | Description
  jackal connect doom  | Play doom!!!
  jackal connect games | Play some old dos games 🦄

# Run the specified `jackal connect <game>` command to connect to the deployed
# workload (ie. kill some demons). Note that the typical Jackal registry,
# Gitea server and Jackal agent pods are not present in the cluster. This means
# that the game's container image was pulled directly from the public registry and the URL was not mutated by Jackal.
```

## `jackal.yaml` {#jackal.yaml}

:::info

To view the example in its entirety, select the `Edit this page` link below the article and select the parent folder.

:::

<ExampleYAML src={require('./jackal.yaml')} showLink={false} />

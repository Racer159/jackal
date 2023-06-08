import ExampleYAML from "@site/src/components/ExampleYAML";

# YOLO Mode

This example demonstrates YOLO mode, an optional mode for using Zarf in a fully connected environment where users can bring their own external container registry and Git server.

## Prerequisites

- A running K8s cluster.

:::note

The cluster does not need to have the Zarf init package installed or any other Zarf-related bootstrapping.

:::

## Instructions

Create the package:

```bash
zarf package create
```

### Deploy the package

```bash
# Run the following command to deploy the created package to the cluster
zarf package deploy

# Choose the yolo package from the list
? Choose or type the package file [tab for suggestions]
> zarf-package-yolo-<ARCH>.tar.zst

# Confirm the deployment
? Deploy this Zarf package? (y/N)

# Wait a few seconds for the cluster to deploy the package; you should
# see the following output when the package has been finished deploying:
  Connect Command    | Description
  zarf connect doom  | Play doom!!!
  zarf connect games | Play some old dos games 🦄

# Run the specified `zarf connect <game>` command to connect to the deployed
# workload (ie. kill some demons). Note that the typical Zarf registry,
# Gitea server and Zarf agent pods are not present in the cluster. This means
# that the game's container image was pulled directly from the public registry and the URL was not mutated by Zarf.
```

## `zarf.yaml` {#zarf.yaml}

:::info

To view the example in its entirety, select the `Edit this page` link below the article and select the parent folder.

:::

<ExampleYAML example="yolo" showLink={false} />

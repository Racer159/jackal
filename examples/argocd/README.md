import ExampleYAML from "@site/src/components/ExampleYAML";

# ArgoCD (with Podinfo)

This example demonstrates how to use ArgoCD with Jackal to deploy the `stefanprodan/podinfo` app using GitOps.

It uses a vanilla configuration of ArgoCD with upstream containers.

If you want to learn more about how Jackal handles `git` repositories, see the [git-data](../git-data/) example.

## `jackal.yaml` {#jackal.yaml}

:::info

To view the example in its entirety, select the `Edit this page` link below the article and select the parent folder.

:::

<ExampleYAML src={require('./jackal.yaml')} showLink={false} />

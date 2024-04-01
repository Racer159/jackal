import ExampleYAML from '@site/src/components/ExampleYAML';

# Component Actions

:::note

Component Actions have replaced Component Scripts. Jackal will still read `scripts` entries, but will convert them to `actions`. Component Scripts will be removed in a future release. Please update your package configurations to use Component Actions instead.

:::

This example demonstrates how to define actions within your package that can run either on `jackal package create`, `jackal package deploy` or `jackal package remove`. These actions will be executed with the context that the Jackal binary is executed with.

For more details on component actions, see the [component actions](../../docs/3-create-a-jackal-package/7-component-actions.md) documentation.

## `jackal.yaml` {#jackal.yaml}

:::info

To view the example in its entirety, select the `Edit this page` link below the article and select the parent folder.

:::

<ExampleYAML src={require('./jackal.yaml')} showLink={false} />

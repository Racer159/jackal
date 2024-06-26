import ExampleYAML from "@site/src/components/ExampleYAML";

# Component Choice

:::caution

Component Choice is currently a [Deprecated Feature](../../docs/9-roadmap.md#alpha). This feature will be removed in Jackal v1.0.0. Please migrate any existing packages you may have that utilize it.  In doing so you may want to consider [Package Flavors](../package-flavors/README.md) as an alternative.

:::

This example demonstrates how to define packages that can be chosen by the user on `jackal package deploy`.  This is done through the `group` key inside of the component specification that defines a group of components a user can select from.

A package creator can also use the `default` key to specify which component will be chosen if a user uses the `--confirm` flag.

:::note

A user can only select a single component in a component group and a package creator can specify only a single default

A component in a component `group` cannot be marked as being `required`

:::

## `jackal.yaml` {#jackal.yaml}

:::info

To view the example in its entirety, select the `Edit this page` link below the article and select the parent folder.

:::

<ExampleYAML src={require('./jackal.yaml')} showLink={false} />

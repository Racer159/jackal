import Properties from '@site/src/components/SchemaItemProperties';
import ExampleYAML from '@site/src/components/ExampleYAML';

# Big Bang

This package deploys [Big Bang](https://repo1.dso.mil/platform-one/big-bang/bigbang) using the Jackal `bigbang` extension.

The `bigbang` noun sits within the `extensions` specification of Jackal and provides the following configuration:

<Properties item="BigBang" />

To see a tutorial for the creation and deployment of this package see the [Big Bang Tutorial](../../docs/5-jackal-tutorials/6-big-bang.md).

## `jackal.yaml` {#jackal.yaml}

:::info

To view the example in its entirety, select the `Edit this page` link below the article and select the parent folder.

:::

<ExampleYAML src={require('./jackal.yaml')} showLink={false} />

:::caution

`valuesFiles` are processed in the order provided with Jackal adding an initial values file to populate registry and git server credentials as the first file.  Including credential `values` (even empty ones) will override these values.  This can be used to our advantage however for things like YOLO mode as described below.

:::

## Big Bang YOLO Mode Support

The Big Bang extension also supports YOLO mode, provided that you add your own credentials for the image registry. This is accomplished below with the `provision-flux-credentials` component and the `credentials.yaml` values file which allows images to be pulled from [registry1.dso.mil](https://registry1.dso.mil). We demonstrate providing account credentials via Jackal Variables, but there are other ways to populate the data in `private-registry.yaml`.

You can learn about YOLO mode in the [FAQ](../../docs/8-faq.md#what-is-yolo-mode-and-why-would-i-use-it) or the [YOLO mode example](../yolo/README.md).

:::info

To view the example in its entirety, select the `Edit this page` link below the article and select the parent folder, then select the `yolo` folder.

:::

<ExampleYAML src={require('./yolo/jackal.yaml')} showLink={false} />

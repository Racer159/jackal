---
sidebar_position: 7
---

import DocCardList from '@theme/DocCardList';
import {useCurrentSidebarCategory} from '@docusaurus/theme-common';

# Package Examples

The Jackal examples demonstrate different ways to utilize Jackal in your environment.  All of these examples follow the same general release pattern and assume an offline / air-gapped deployment target.

To build and deploy a demo, change directories to the example you want to try and run:

```bash
cd <directory> # This should be whatever example you want to try (i.e. game)
jackal package create # This will create the jackal package
jackal package deploy # This will prompt you to deploy the created jackal package
```

:::caution

Examples are for demo purposes only and are not meant for production use, they exist to demo various ways to use Jackal. Modifying examples to fit production use is possible but requires additional configuration, time, and Kubernetes knowledge.

Examples also utilize software pulled from multiple sources and _some_ of them require authenticated access. Check the examples themselves for the specific accounts / logins required.

:::

<DocCardList items={useCurrentSidebarCategory().items.slice(1)}/>

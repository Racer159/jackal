# Getting Started - VS Code

Jackal uses the [Jackal package schema](https://github.com/racer159/jackal/blob/main/jackal.schema.json) to define its configuration files. This schema is used to describe package configuration options and enable the validation of configuration files prior to their use in building a Jackal Package.

## Adding Schema Validation

1. Open VS Code.
2. Install the [YAML extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) by RedHat.
3. Open the VS Code command palette by typing `CTRL/CMD + SHIFT + P`.
4. Type `Preferences: Open User Settings (JSON)`into the search bar to open the `settings.json` file.
5. Add the below code to the settings.json config, or modify the existing `yaml.schemas` object to include the Jackal schema.

```json
  "yaml.schemas": {
    "https://raw.githubusercontent.com/racer159/jackal/main/jackal.schema.json": "jackal.yaml"
  }
```

:::note

When successfully installed, the `yaml.schema` line will match the color of the other lines within the settings.

:::

## Specifying Jackal's Schema Version

To ensure consistent validation of the Jackal schema version in a `jackal.yaml` file, it can be beneficial to lock it to a specific version. This can be achieved by appending the following statement to the **first line** of any given `jackal.yaml` file:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/racer159/jackal/<VERSION>/jackal.schema.json
```

In the above example, `<VERSION>` should be replaced with the specific [Jackal release](https://github.com/racer159/jackal/releases).

### Code Example

![yaml schema](https://user-images.githubusercontent.com/92826525/226490465-1e6a56f7-41c4-45bf-923b-5242fa4ab64e.png)

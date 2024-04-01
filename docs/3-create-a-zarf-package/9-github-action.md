# Getting Started - Github Action

The [setup-jackal](https://github.com/defenseunicorns/setup-jackal) Github action is an officially supported action to install any version of Jackal and it's `init` package with zero added dependencies.

## Example Usage - Creating a Package

```yaml
# .github/workflows/jackal-package-create.yml
jobs:
  create_package:
    runs-on: ubuntu-latest

    name: Create my cool Jackal Package
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 1

      - name: Install Jackal
        uses: defenseunicorns/setup-jackal@main # use action's main branch
        with:
          version: v0.22.2 # any valid jackal version, leave blank to use latest

      - name: Create the package
        run: jackal package create --confirm
```

More examples are located in the action's [README.md](https://github.com/defenseunicorns/setup-jackal#readme)

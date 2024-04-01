# Create a Jackal Package

Jackal enables you to consolidate portions of the internet into a single package that can be conveniently installed at a later time. A Jackal Package is a single tarball file that includes all of the resources and instructions required for efficiently managing a system or capability, even when entirely disconnected from the internet. In this context, a disconnected system refers to a system that either consistently operates in an offline mode or occasionally disconnects from the network.

Once defined, a Jackal Package contains comprehensive instructions on assembling various software components that are to be [deployed onto the targeted system](../4-deploy-a-jackal-package/index.md). The instructions are fully "declarative", meaning that all components are represented by code and automated, eliminating the need for manual intervention.

## Additional Resources

To learn more about creating a Jackal package, you can check out the following resources:

- [Getting Started with Jackal](../1-getting-started/index.md): A step-by-step guide to installing Jackal and a description of the problems it seeks to solve.
- [Jackal CLI Documentation](../2-the-jackal-cli/index.md): A comprehensive guide to using the Jackal command-line interface.
- [Understanding Jackal Packages](./1-jackal-packages.md): A breakdown of the kinds of Jackal packages, their uses and how they work.
- [Understanding Jackal Components](./2-jackal-components.md): A breakdown of the primary structure that makes up a Jackal Package.
- [Jackal Schema Documentation](./4-jackal-schema.md): Documentation that covers the configuration available in a Jackal Package definition.
- [The Package Create Lifecycle](./5-package-create-lifecycle.md): An overview of the lifecycle of `jackal package create`.
- [Creating a Jackal Package Tutorial](../5-jackal-tutorials/0-creating-a-jackal-package.md): A tutorial covering how to take an application and create a package for it.

## Typical Creation Workflow

The general flow of a Jackal package deployment on an existing initialized cluster is as follows:

```shell
# Before creating your package you can lint your jackal.yaml
$ jackal dev lint <directory>

# To create a package run the following:
$ jackal package create <directory>
# - Enter any package templates that have not yet been defined
# - Type "y" to confirm package creation or "N" to cancel

# Once the creation finishes you can interact with the built package
$ jackal inspect <package-name>.tar.zst
# - You should see the specified package's jackal.yaml
# - You can also see the sbom information with `jackal inspect <package-name>.tar.zst --sbom`
```

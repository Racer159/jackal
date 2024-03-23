---
title: zarf destroy
description: Zarf CLI command reference for <code>zarf destroy</code>.
---

## zarf destroy

Tears down Zarf and removes its components from the environment

### Synopsis

Tear down Zarf.

Deletes everything in the 'zarf' namespace within your connected k8s cluster.

If Zarf deployed your k8s cluster, this command will also tear your cluster down by searching through /opt/zarf for any scripts that start with 'zarf-clean-' and executing them. Since this is a cleanup operation, Zarf will not stop the teardown if one of the scripts produce an error.

If Zarf did not deploy your k8s cluster, this command will delete the Zarf namespace, delete secrets and labels that only Zarf cares about, and optionally uninstall components that Zarf deployed onto the cluster. Since this is a cleanup operation, Zarf will not stop the uninstalls if one of the resources produce an error while being deleted.

```
zarf destroy --confirm [flags]
```

### Options

```
      --confirm             REQUIRED. Confirm the destroy action to prevent accidental deletions
  -h, --help                help for destroy
      --remove-components   Also remove any installed components outside the zarf namespace
```

### Options inherited from parent commands

```
  -a, --architecture string   Architecture for OCI images and Zarf packages
      --insecure              Allow access to insecure registries and disable other recommended security enforcements such as package checksum and signature validation. This flag should only be used if you have a specific reason and accept the reduced security posture.
  -l, --log-level string      Log level when running Zarf. Valid options are: warn, info, debug, trace (default "info")
      --no-color              Disable colors in output
      --no-log-file           Disable log file creation
      --no-progress           Disable fancy UI progress bars, spinners, logos, etc
      --tmpdir string         Specify the temporary directory to use for intermediate files
      --zarf-cache string     Specify the location of the Zarf cache directory (default "~/.zarf-cache")
```

### SEE ALSO

* [zarf](/cli/commands/zarf/)	 - DevSecOps for Airgap

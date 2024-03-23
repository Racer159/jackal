---
title: zarf package mirror-resources
description: Zarf CLI command reference for <code>zarf package mirror-resources</code>.
---

## zarf package mirror-resources

Mirrors a Zarf package's internal resources to specified image registries and git repositories

### Synopsis

Unpacks resources and dependencies from a Zarf package archive and mirrors them into the specified
image registries and git repositories within the target environment

```
zarf package mirror-resources [ PACKAGE_SOURCE ] [flags]
```

### Examples

```

# Mirror resources to internal Zarf resources
$ zarf package mirror-resources <your-package.tar.zst> \
	--registry-url 127.0.0.1:31999 \
	--registry-push-username zarf-push \
	--registry-push-password <generated-registry-push-password> \
	--git-url http://zarf-gitea-http.zarf.svc.cluster.local:3000 \
	--git-push-username zarf-git-user \
	--git-push-password <generated-git-push-password>

# Mirror resources to external resources
$ zarf package mirror-resources <your-package.tar.zst> \
	--registry-url registry.enterprise.corp \
	--registry-push-username <registry-push-username> \
	--registry-push-password <registry-push-password> \
	--git-url https://git.enterprise.corp \
	--git-push-username <git-push-username> \
	--git-push-password <git-push-password>

```

### Options

```
      --components string               Comma-separated list of components to mirror.  This list will be respected regardless of a component's 'required' or 'default' status.  Globbing component names with '*' and deselecting components with a leading '-' are also supported.
      --confirm                         Confirms package deployment without prompting. ONLY use with packages you trust. Skips prompts to review SBOM, configure variables, select optional components and review potential breaking changes.
      --git-push-password string        Password for the push-user to access the git server
      --git-push-username string        Username to access to the git server Zarf is configured to use. User must be able to create repositories via 'git push' (default "zarf-git-user")
      --git-url string                  External git server url to use for this Zarf cluster
  -h, --help                            help for mirror-resources
      --no-img-checksum                 Turns off the addition of a checksum to image tags (as would be used by the Zarf Agent) while mirroring images.
      --registry-push-password string   Password for the push-user to connect to the registry
      --registry-push-username string   Username to access to the registry Zarf is configured to use (default "zarf-push")
      --registry-url string             External registry url address to use for this Zarf cluster
```

### Options inherited from parent commands

```
  -a, --architecture string   Architecture for OCI images and Zarf packages
      --insecure              Allow access to insecure registries and disable other recommended security enforcements such as package checksum and signature validation. This flag should only be used if you have a specific reason and accept the reduced security posture.
  -k, --key string            Path to public key file for validating signed packages
  -l, --log-level string      Log level when running Zarf. Valid options are: warn, info, debug, trace (default "info")
      --no-color              Disable colors in output
      --no-log-file           Disable log file creation
      --no-progress           Disable fancy UI progress bars, spinners, logos, etc
      --oci-concurrency int   Number of concurrent layer operations to perform when interacting with a remote package. (default 3)
      --tmpdir string         Specify the temporary directory to use for intermediate files
      --zarf-cache string     Specify the location of the Zarf cache directory (default "~/.zarf-cache")
```

### SEE ALSO

* [zarf package](/cli/commands/zarf_package/)	 - Zarf package commands for creating, deploying, and inspecting packages

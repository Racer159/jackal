# jackal tools update-creds
<!-- Auto-generated by hack/gen-cli-docs.sh -->

Updates the credentials for deployed Jackal services. Pass a service key to update credentials for a single service

## Synopsis

Updates the credentials for deployed Jackal services. Pass a service key to update credentials for a single service. i.e. 'jackal tools update-creds registry'

```
jackal tools update-creds [flags]
```

## Examples

```

# Autogenerate all Jackal credentials at once:
$ jackal tools update-creds

# Autogenerate specific Jackal service credentials:
$ jackal tools update-creds registry
$ jackal tools update-creds git
$ jackal tools update-creds artifact
$ jackal tools update-creds agent

# Update all Jackal credentials w/external services at once:
$ jackal tools update-creds \
	--registry-push-username={USERNAME} --registry-push-password={PASSWORD} \
	--git-push-username={USERNAME} --git-push-password={PASSWORD} \
	--artifact-push-username={USERNAME} --artifact-push-token={PASSWORD}

# NOTE: Any credentials omitted from flags without a service key specified will be autogenerated - URLs will only change if specified.
# Config options can also be set with the 'init' section of a Jackal config file.

# Update specific Jackal credentials w/external services:
$ jackal tools update-creds registry --registry-push-username={USERNAME} --registry-push-password={PASSWORD}
$ jackal tools update-creds git --git-push-username={USERNAME} --git-push-password={PASSWORD}
$ jackal tools update-creds artifact --artifact-push-username={USERNAME} --artifact-push-token={PASSWORD}

# NOTE: Not specifying a pull username/password will keep the previous pull username/password.

```

## Options

```
      --artifact-push-token string      [alpha] API Token for the push-user to access the artifact registry
      --artifact-push-username string   [alpha] Username to access to the artifact registry Jackal is configured to use. User must be able to upload package artifacts.
      --artifact-url string             [alpha] External artifact registry url to use for this Jackal cluster
      --confirm                         Confirm updating credentials without prompting
      --git-pull-password string        Password for the pull-only user to access the git server
      --git-pull-username string        Username for pull-only access to the git server
      --git-push-password string        Password for the push-user to access the git server
      --git-push-username string        Username to access to the git server Jackal is configured to use. User must be able to create repositories via 'git push'
      --git-url string                  External git server url to use for this Jackal cluster
  -h, --help                            help for update-creds
      --registry-pull-password string   Password for the pull-only user to access the registry
      --registry-pull-username string   Username for pull-only access to the registry
      --registry-push-password string   Password for the push-user to connect to the registry
      --registry-push-username string   Username to access to the registry Jackal is configured to use
      --registry-url string             External registry url address to use for this Jackal cluster
```

## Options inherited from parent commands

```
  -a, --architecture string   Architecture for OCI images and Jackal packages
      --insecure              Allow access to insecure registries and disable other recommended security enforcements such as package checksum and signature validation. This flag should only be used if you have a specific reason and accept the reduced security posture.
  -l, --log-level string      Log level when running Jackal. Valid options are: warn, info, debug, trace (default "info")
      --no-color              Disable colors in output
      --no-log-file           Disable log file creation
      --no-progress           Disable fancy UI progress bars, spinners, logos, etc
      --tmpdir string         Specify the temporary directory to use for intermediate files
      --jackal-cache string     Specify the location of the Jackal cache directory (default "~/.jackal-cache")
```

## SEE ALSO

* [jackal tools](jackal_tools.md)	 - Collection of additional tools to make airgap easier

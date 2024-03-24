---
title: zarf tools helm
description: Zarf CLI command reference for <code>zarf tools helm</code>.
tableOfContents: false
---

## zarf tools helm

Subset of the Helm CLI included with Zarf to help manage helm charts.

### Synopsis

Subset of the Helm CLI that includes the repo and dependency commands for managing helm charts destined for the air gap.

### Options

```
      --burst-limit int                 client-side default throttling limit (default 100)
      --debug                           enable verbose output
  -h, --help                            help for helm
      --kube-apiserver string           the address and the port for the Kubernetes API server
      --kube-as-group stringArray       group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --kube-as-user string             username to impersonate for the operation
      --kube-ca-file string             the certificate authority file for the Kubernetes API server connection
      --kube-context string             name of the kubeconfig context to use
      --kube-insecure-skip-tls-verify   if true, the Kubernetes API server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kube-tls-server-name string     server name to use for Kubernetes API server certificate validation. If it is not provided, the hostname used to contact the server is used
      --kube-token string               bearer token used for authentication
      --kubeconfig string               path to the kubeconfig file
  -n, --namespace string                namespace scope for this request
      --qps float32                     queries per second used when communicating with the Kubernetes API, not including bursting
      --registry-config string          path to the registry config file (default "/home/thicc/.config/helm/registry/config.json")
      --repository-cache string         path to the file containing cached repository indexes (default "/home/thicc/.cache/helm/repository")
      --repository-config string        path to the file containing repository names and URLs (default "/home/thicc/.config/helm/repositories.yaml")
```

### SEE ALSO

* [zarf tools](/commands/zarf_tools/)	 - Collection of additional tools to make airgap easier
* [zarf tools helm dependency](/commands/zarf_tools_helm_dependency/)	 - manage a chart's dependencies
* [zarf tools helm repo](/commands/zarf_tools_helm_repo/)	 - add, list, remove, update, and index chart repositories
* [zarf tools helm version](/commands/zarf_tools_helm_version/)	 - Print the version

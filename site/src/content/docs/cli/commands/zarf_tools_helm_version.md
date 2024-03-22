---
title: zarf tools helm version
description: Zarf CLI command reference for <code>zarf tools helm version</code>.
---

## zarf tools helm version

Print the version

```
zarf tools helm version [flags]
```

### Options

```
  -h, --help   help for version
```

### Options inherited from parent commands

```
      --burst-limit int                 client-side default throttling limit (default 100)
      --debug                           enable verbose output
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
      --registry-config string          path to the registry config file
      --repository-cache string         path to the file containing cached repository indexes
      --repository-config string        path to the file containing repository names and URLs
```

### SEE ALSO

* [zarf tools helm](/cli/commands/zarf_tools_helm/)	 - Subset of the Helm CLI included with Zarf to help manage helm charts.

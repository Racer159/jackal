---
title: zarf tools wait-for
description: Zarf CLI command reference for <code>zarf tools wait-for</code>.
tableOfContents: false
---

## zarf tools wait-for

Waits for a given Kubernetes resource to be ready

### Synopsis

By default Zarf will wait for all Kubernetes resources to be ready before completion of a component during a deployment.
This command can be used to wait for a Kubernetes resources to exist and be ready that may be created by a Gitops tool or a Kubernetes operator.
You can also wait for arbitrary network endpoints using REST or TCP checks.



```
zarf tools wait-for { KIND | PROTOCOL } { NAME | SELECTOR | URI } { CONDITION | HTTP_CODE } [flags]
```

### Examples

```

# Wait for Kubernetes resources:
$ zarf tools wait-for pod my-pod-name ready -n default                  #  wait for pod my-pod-name in namespace default to be ready
$ zarf tools wait-for p cool-pod-name ready -n cool                     #  wait for pod (using p alias) cool-pod-name in namespace cool to be ready
$ zarf tools wait-for deployment podinfo available -n podinfo           #  wait for deployment podinfo in namespace podinfo to be available
$ zarf tools wait-for pod app=podinfo ready -n podinfo                  #  wait for pod with label app=podinfo in namespace podinfo to be ready
$ zarf tools wait-for svc zarf-docker-registry exists -n zarf           #  wait for service zarf-docker-registry in namespace zarf to exist
$ zarf tools wait-for svc zarf-docker-registry -n zarf                  #  same as above, except exists is the default condition
$ zarf tools wait-for crd addons.k3s.cattle.io                          #  wait for crd addons.k3s.cattle.io to exist
$ zarf tools wait-for sts test-sts '{.status.availableReplicas}'=23     #  wait for statefulset test-sts to have 23 available replicas

# Wait for network endpoints:
$ zarf tools wait-for http localhost:8080 200                           #  wait for a 200 response from http://localhost:8080
$ zarf tools wait-for tcp localhost:8080                                #  wait for a connection to be established on localhost:8080
$ zarf tools wait-for https 1.1.1.1 200                                 #  wait for a 200 response from https://1.1.1.1
$ zarf tools wait-for http google.com                                   #  wait for any 2xx response from http://google.com
$ zarf tools wait-for http google.com success                           #  wait for any 2xx response from http://google.com

```

### Options

```
  -h, --help               help for wait-for
  -n, --namespace string   Specify the namespace of the resources to wait for.
      --no-progress        Disable fancy UI progress bars, spinners, logos, etc
      --timeout string     Specify the timeout duration for the wait command. (default "5m")
```

### SEE ALSO

* [zarf tools](/commands/zarf_tools/)	 - Collection of additional tools to make airgap easier

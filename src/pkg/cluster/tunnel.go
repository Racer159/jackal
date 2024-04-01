// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package cluster contains Jackal-specific cluster management functions.
package cluster

import (
	"fmt"
	"strings"

	"github.com/Racer159/jackal/src/types"

	"github.com/Racer159/jackal/src/config"
	"github.com/Racer159/jackal/src/pkg/k8s"
	"github.com/Racer159/jackal/src/pkg/message"
	v1 "k8s.io/api/core/v1"
)

// Jackal specific connect strings
const (
	JackalRegistry = "REGISTRY"
	JackalLogging  = "LOGGING"
	JackalGit      = "GIT"
	JackalInjector = "INJECTOR"

	JackalInjectorName  = "jackal-injector"
	JackalInjectorPort  = 5000
	JackalRegistryName  = "jackal-docker-registry"
	JackalRegistryPort  = 5000
	JackalGitServerName = "jackal-gitea-http"
	JackalGitServerPort = 3000
)

// TunnelInfo is a struct that contains the necessary info to create a new k8s.Tunnel
type TunnelInfo struct {
	localPort    int
	remotePort   int
	namespace    string
	resourceType string
	resourceName string
	urlSuffix    string
}

// NewTunnelInfo returns a new TunnelInfo object for connecting to a cluster
func NewTunnelInfo(namespace, resourceType, resourceName, urlSuffix string, localPort, remotePort int) TunnelInfo {
	return TunnelInfo{
		namespace:    namespace,
		resourceType: resourceType,
		resourceName: resourceName,
		urlSuffix:    urlSuffix,
		localPort:    localPort,
		remotePort:   remotePort,
	}
}

// PrintConnectTable will print a table of all Jackal connect matches found in the cluster.
func (c *Cluster) PrintConnectTable() error {
	list, err := c.GetServicesByLabelExists(v1.NamespaceAll, config.JackalConnectLabelName)
	if err != nil {
		return err
	}

	connections := make(types.ConnectStrings)

	for _, svc := range list.Items {
		name := svc.Labels[config.JackalConnectLabelName]

		// Add the connectString for processing later in the deployment.
		connections[name] = types.ConnectString{
			Description: svc.Annotations[config.JackalConnectAnnotationDescription],
			URL:         svc.Annotations[config.JackalConnectAnnotationURL],
		}
	}

	message.PrintConnectStringTable(connections)

	return nil
}

// Connect will establish a tunnel to the specified target.
func (c *Cluster) Connect(target string) (*k8s.Tunnel, error) {
	var err error
	zt := TunnelInfo{
		namespace:    JackalNamespaceName,
		resourceType: k8s.SvcResource,
	}

	switch strings.ToUpper(target) {
	case JackalRegistry:
		zt.resourceName = JackalRegistryName
		zt.remotePort = JackalRegistryPort
		zt.urlSuffix = `/v2/_catalog`

	case JackalLogging:
		zt.resourceName = "jackal-loki-stack-grafana"
		zt.remotePort = 3000
		// Start the logs with something useful.
		zt.urlSuffix = `/monitor/explore?orgId=1&left=%5B"now-12h","now","Loki",%7B"refId":"Jackal%20Logs","expr":"%7Bnamespace%3D%5C"jackal%5C"%7D"%7D%5D`

	case JackalGit:
		zt.resourceName = JackalGitServerName
		zt.remotePort = JackalGitServerPort

	case JackalInjector:
		zt.resourceName = JackalInjectorName
		zt.remotePort = JackalInjectorPort

	default:
		if target != "" {
			if zt, err = c.checkForJackalConnectLabel(target); err != nil {
				return nil, fmt.Errorf("problem looking for a jackal connect label in the cluster: %s", err.Error())
			}
		}

		if zt.resourceName == "" {
			return nil, fmt.Errorf("missing resource name")
		}
		if zt.remotePort < 1 {
			return nil, fmt.Errorf("missing remote port")
		}
	}

	return c.ConnectTunnelInfo(zt)
}

// ConnectTunnelInfo connects to the cluster with the provided TunnelInfo
func (c *Cluster) ConnectTunnelInfo(zt TunnelInfo) (*k8s.Tunnel, error) {
	tunnel, err := c.NewTunnel(zt.namespace, zt.resourceType, zt.resourceName, zt.urlSuffix, zt.localPort, zt.remotePort)
	if err != nil {
		return nil, err
	}

	_, err = tunnel.Connect()
	if err != nil {
		return nil, err
	}

	return tunnel, nil
}

// ConnectToJackalRegistryEndpoint determines if a registry endpoint is in cluster, and if so opens a tunnel to connect to it
func (c *Cluster) ConnectToJackalRegistryEndpoint(registryInfo types.RegistryInfo) (string, *k8s.Tunnel, error) {
	registryEndpoint := registryInfo.Address

	var err error
	var tunnel *k8s.Tunnel
	if registryInfo.InternalRegistry {
		// Establish a registry tunnel to send the images to the jackal registry
		if tunnel, err = c.NewTunnel(JackalNamespaceName, k8s.SvcResource, JackalRegistryName, "", 0, JackalRegistryPort); err != nil {
			return "", tunnel, err
		}
	} else {
		svcInfo, err := c.ServiceInfoFromNodePortURL(registryInfo.Address)

		// If this is a service (no error getting svcInfo), create a port-forward tunnel to that resource
		if err == nil {
			if tunnel, err = c.NewTunnel(svcInfo.Namespace, k8s.SvcResource, svcInfo.Name, "", 0, svcInfo.Port); err != nil {
				return "", tunnel, err
			}
		}
	}

	if tunnel != nil {
		_, err = tunnel.Connect()
		if err != nil {
			return "", tunnel, err
		}
		registryEndpoint = tunnel.Endpoint()
	}

	return registryEndpoint, tunnel, nil
}

// checkForJackalConnectLabel looks in the cluster for a connect name that matches the target
func (c *Cluster) checkForJackalConnectLabel(name string) (TunnelInfo, error) {
	var err error
	var zt TunnelInfo

	message.Debugf("Looking for a Jackal Connect Label in the cluster")

	matches, err := c.GetServicesByLabel("", config.JackalConnectLabelName, name)
	if err != nil {
		return zt, fmt.Errorf("unable to lookup the service: %w", err)
	}

	if len(matches.Items) > 0 {
		// If there is a match, use the first one as these are supposed to be unique.
		svc := matches.Items[0]

		// Reset based on the matched params.
		zt.resourceType = k8s.SvcResource
		zt.resourceName = svc.Name
		zt.namespace = svc.Namespace
		// Only support a service with a single port.
		zt.remotePort = svc.Spec.Ports[0].TargetPort.IntValue()
		// if targetPort == 0, look for Port (which is required)
		if zt.remotePort == 0 {
			zt.remotePort = c.FindPodContainerPort(svc)
		}

		// Add the url suffix too.
		zt.urlSuffix = svc.Annotations[config.JackalConnectAnnotationURL]

		message.Debugf("tunnel connection match: %s/%s on port %d", svc.Namespace, svc.Name, zt.remotePort)
	} else {
		return zt, fmt.Errorf("no matching services found for %s", name)
	}

	return zt, nil
}

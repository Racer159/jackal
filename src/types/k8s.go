// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package types contains all the types used by Jackal.
package types

import (
	"fmt"
	"time"

	"github.com/defenseunicorns/jackal/src/config/lang"
	"github.com/defenseunicorns/jackal/src/pkg/k8s"
	"github.com/defenseunicorns/pkg/helpers"
)

// WebhookStatus defines the status of a Component Webhook operating on a Jackal package secret.
type WebhookStatus string

// ComponentStatus defines the deployment status of a Jackal component within a package.
type ComponentStatus string

// DefaultWebhookWaitDuration is the default amount of time Jackal will wait for a webhook to complete.
const DefaultWebhookWaitDuration = time.Minute * 5

// All the different status options for a Jackal Component or a webhook that is running for a Jackal Component deployment.
const (
	WebhookStatusSucceeded WebhookStatus = "Succeeded"
	WebhookStatusFailed    WebhookStatus = "Failed"
	WebhookStatusRunning   WebhookStatus = "Running"
	WebhookStatusRemoving  WebhookStatus = "Removing"

	ComponentStatusSucceeded ComponentStatus = "Succeeded"
	ComponentStatusFailed    ComponentStatus = "Failed"
	ComponentStatusDeploying ComponentStatus = "Deploying"
	ComponentStatusRemoving  ComponentStatus = "Removing"
)

// Values during setup of the initial jackal state
const (
	JackalGeneratedPasswordLen               = 24
	JackalGeneratedSecretLen                 = 48
	JackalInClusterContainerRegistryNodePort = 31999
	JackalRegistryPushUser                   = "jackal-push"
	JackalRegistryPullUser                   = "jackal-pull"

	JackalGitPushUser = "jackal-git-user"
	JackalGitReadUser = "jackal-git-read-user"

	JackalInClusterGitServiceURL      = "http://jackal-gitea-http.jackal.svc.cluster.local:3000"
	JackalInClusterArtifactServiceURL = JackalInClusterGitServiceURL + "/api/packages/" + JackalGitPushUser
)

// JackalState is maintained as a secret in the Jackal namespace to track Jackal init data.
type JackalState struct {
	JackalAppliance bool             `json:"jackalAppliance" jsonschema:"description=Indicates if Jackal was initialized while deploying its own k8s cluster"`
	Distro          string           `json:"distro" jsonschema:"description=K8s distribution of the cluster Jackal was deployed to"`
	Architecture    string           `json:"architecture" jsonschema:"description=Machine architecture of the k8s node(s)"`
	StorageClass    string           `json:"storageClass" jsonschema:"Default StorageClass value Jackal uses for variable templating"`
	AgentTLS        k8s.GeneratedPKI `json:"agentTLS" jsonschema:"PKI certificate information for the agent pods Jackal manages"`

	GitServer      GitServerInfo      `json:"gitServer" jsonschema:"description=Information about the repository Jackal is configured to use"`
	RegistryInfo   RegistryInfo       `json:"registryInfo" jsonschema:"description=Information about the container registry Jackal is configured to use"`
	ArtifactServer ArtifactServerInfo `json:"artifactServer" jsonschema:"description=Information about the artifact registry Jackal is configured to use"`
	LoggingSecret  string             `json:"loggingSecret" jsonschema:"description=Secret value that the internal Grafana server was seeded with"`
}

// DeployedPackage contains information about a Jackal Package that has been deployed to a cluster
// This object is saved as the data of a k8s secret within the 'Jackal' namespace (not as part of the JackalState secret).
type DeployedPackage struct {
	Name               string                        `json:"name"`
	Data               JackalPackage                 `json:"data"`
	CLIVersion         string                        `json:"cliVersion"`
	Generation         int                           `json:"generation"`
	DeployedComponents []DeployedComponent           `json:"deployedComponents"`
	ComponentWebhooks  map[string]map[string]Webhook `json:"componentWebhooks,omitempty"`
	ConnectStrings     ConnectStrings                `json:"connectStrings,omitempty"`
}

// DeployedComponent contains information about a Jackal Package Component that has been deployed to a cluster.
type DeployedComponent struct {
	Name               string           `json:"name"`
	InstalledCharts    []InstalledChart `json:"installedCharts"`
	Status             ComponentStatus  `json:"status"`
	ObservedGeneration int              `json:"observedGeneration"`
}

// Webhook contains information about a Component Webhook operating on a Jackal package secret.
type Webhook struct {
	Name                string        `json:"name"`
	WaitDurationSeconds int           `json:"waitDurationSeconds,omitempty"`
	Status              WebhookStatus `json:"status"`
	ObservedGeneration  int           `json:"observedGeneration"`
}

// InstalledChart contains information about a Helm Chart that has been deployed to a cluster.
type InstalledChart struct {
	Namespace string `json:"namespace"`
	ChartName string `json:"chartName"`
}

// GitServerInfo contains information Jackal uses to communicate with a git repository to push/pull repositories to.
type GitServerInfo struct {
	PushUsername string `json:"pushUsername" jsonschema:"description=Username of a user with push access to the git repository"`
	PushPassword string `json:"pushPassword" jsonschema:"description=Password of a user with push access to the git repository"`
	PullUsername string `json:"pullUsername" jsonschema:"description=Username of a user with pull-only access to the git repository. If not provided for an external repository then the push-user is used"`
	PullPassword string `json:"pullPassword" jsonschema:"description=Password of a user with pull-only access to the git repository. If not provided for an external repository then the push-user is used"`

	Address        string `json:"address" jsonschema:"description=URL address of the git server"`
	InternalServer bool   `json:"internalServer" jsonschema:"description=Indicates if we are using a git server that Jackal is directly managing"`
}

// FillInEmptyValues sets every necessary value that's currently empty to a reasonable default
func (gs *GitServerInfo) FillInEmptyValues() error {
	var err error
	// Set default svc url if an external repository was not provided
	if gs.Address == "" {
		gs.Address = JackalInClusterGitServiceURL
		gs.InternalServer = true
	}

	// Generate a push-user password if not provided by init flag
	if gs.PushPassword == "" {
		if gs.PushPassword, err = helpers.RandomString(JackalGeneratedPasswordLen); err != nil {
			return fmt.Errorf("%s: %w", lang.ErrUnableToGenerateRandomSecret, err)
		}
	}

	// Set read-user information if using an internal repository, otherwise copy from the push-user
	if gs.PullUsername == "" {
		if gs.InternalServer {
			gs.PullUsername = JackalGitReadUser
		} else {
			gs.PullUsername = gs.PushUsername
		}
	}
	if gs.PullPassword == "" {
		if gs.InternalServer {
			if gs.PullPassword, err = helpers.RandomString(JackalGeneratedPasswordLen); err != nil {
				return fmt.Errorf("%s: %w", lang.ErrUnableToGenerateRandomSecret, err)
			}
		} else {
			gs.PullPassword = gs.PushPassword
		}
	}

	return nil
}

// ArtifactServerInfo contains information Jackal uses to communicate with a artifact registry to push/pull repositories to.
type ArtifactServerInfo struct {
	PushUsername string `json:"pushUsername" jsonschema:"description=Username of a user with push access to the artifact registry"`
	PushToken    string `json:"pushPassword" jsonschema:"description=Password of a user with push access to the artifact registry"`

	Address        string `json:"address" jsonschema:"description=URL address of the artifact registry"`
	InternalServer bool   `json:"internalServer" jsonschema:"description=Indicates if we are using a artifact registry that Jackal is directly managing"`
}

// FillInEmptyValues sets every necessary value that's currently empty to a reasonable default
func (as *ArtifactServerInfo) FillInEmptyValues() {
	// Set default svc url if an external registry was not provided
	if as.Address == "" {
		as.Address = JackalInClusterArtifactServiceURL
		as.InternalServer = true
	}

	// Set the push username to the git push user if not specified
	if as.PushUsername == "" {
		as.PushUsername = JackalGitPushUser
	}
}

// RegistryInfo contains information Jackal uses to communicate with a container registry to push/pull images.
type RegistryInfo struct {
	PushUsername string `json:"pushUsername" jsonschema:"description=Username of a user with push access to the registry"`
	PushPassword string `json:"pushPassword" jsonschema:"description=Password of a user with push access to the registry"`
	PullUsername string `json:"pullUsername" jsonschema:"description=Username of a user with pull-only access to the registry. If not provided for an external registry than the push-user is used"`
	PullPassword string `json:"pullPassword" jsonschema:"description=Password of a user with pull-only access to the registry. If not provided for an external registry than the push-user is used"`

	Address          string `json:"address" jsonschema:"description=URL address of the registry"`
	NodePort         int    `json:"nodePort" jsonschema:"description=Nodeport of the registry. Only needed if the registry is running inside the kubernetes cluster"`
	InternalRegistry bool   `json:"internalRegistry" jsonschema:"description=Indicates if we are using a registry that Jackal is directly managing"`

	Secret string `json:"secret" jsonschema:"description=Secret value that the registry was seeded with"`
}

// FillInEmptyValues sets every necessary value not already set to a reasonable default
func (ri *RegistryInfo) FillInEmptyValues() error {
	var err error
	// Set default NodePort if none was provided
	if ri.NodePort == 0 {
		ri.NodePort = JackalInClusterContainerRegistryNodePort
	}

	// Set default url if an external registry was not provided
	if ri.Address == "" {
		ri.InternalRegistry = true
		ri.Address = fmt.Sprintf("%s:%d", helpers.IPV4Localhost, ri.NodePort)
	}

	// Generate a push-user password if not provided by init flag
	if ri.PushPassword == "" {
		if ri.PushPassword, err = helpers.RandomString(JackalGeneratedPasswordLen); err != nil {
			return fmt.Errorf("%s: %w", lang.ErrUnableToGenerateRandomSecret, err)
		}
	}

	// Set pull-username if not provided by init flag
	if ri.PullUsername == "" {
		if ri.InternalRegistry {
			ri.PullUsername = JackalRegistryPullUser
		} else {
			// If this is an external registry and a pull-user wasn't provided, use the same credentials as the push user
			ri.PullUsername = ri.PushUsername
		}
	}
	if ri.PullPassword == "" {
		if ri.InternalRegistry {
			if ri.PullPassword, err = helpers.RandomString(JackalGeneratedPasswordLen); err != nil {
				return fmt.Errorf("%s: %w", lang.ErrUnableToGenerateRandomSecret, err)
			}
		} else {
			// If this is an external registry and a pull-user wasn't provided, use the same credentials as the push user
			ri.PullPassword = ri.PushPassword
		}
	}

	if ri.Secret == "" {
		if ri.Secret, err = helpers.RandomString(JackalGeneratedSecretLen); err != nil {
			return fmt.Errorf("%s: %w", lang.ErrUnableToGenerateRandomSecret, err)
		}
	}

	return nil
}

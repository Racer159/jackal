// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package cluster contains Jackal-specific cluster management functions.
package cluster

import (
	"encoding/base64"
	"encoding/json"
	"reflect"

	corev1 "k8s.io/api/core/v1"

	"github.com/defenseunicorns/jackal/src/config"
	"github.com/defenseunicorns/jackal/src/pkg/message"
	"github.com/defenseunicorns/jackal/src/types"
)

// DockerConfig contains the authentication information from the machine's docker config.
type DockerConfig struct {
	Auths DockerConfigEntry `json:"auths"`
}

// DockerConfigEntry contains a map of DockerConfigEntryWithAuth for a registry.
type DockerConfigEntry map[string]DockerConfigEntryWithAuth

// DockerConfigEntryWithAuth contains a docker config authentication string.
type DockerConfigEntryWithAuth struct {
	Auth string `json:"auth"`
}

// GenerateRegistryPullCreds generates a secret containing the registry credentials.
func (c *Cluster) GenerateRegistryPullCreds(namespace, name string, registryInfo types.RegistryInfo) *corev1.Secret {
	secretDockerConfig := c.GenerateSecret(namespace, name, corev1.SecretTypeDockerConfigJson)

	// Auth field must be username:password and base64 encoded
	fieldValue := registryInfo.PullUsername + ":" + registryInfo.PullPassword
	authEncodedValue := base64.StdEncoding.EncodeToString([]byte(fieldValue))

	registry := registryInfo.Address
	// Create the expected structure for the dockerconfigjson
	dockerConfigJSON := DockerConfig{
		Auths: DockerConfigEntry{
			registry: DockerConfigEntryWithAuth{
				Auth: authEncodedValue,
			},
		},
	}

	// Convert to JSON
	dockerConfigData, err := json.Marshal(dockerConfigJSON)
	if err != nil {
		message.WarnErrf(err, "Unable to marshal the .dockerconfigjson secret data for the image pull secret")
	}

	// Add to the secret data
	secretDockerConfig.Data[".dockerconfigjson"] = dockerConfigData

	return secretDockerConfig
}

// GenerateGitPullCreds generates a secret containing the git credentials.
func (c *Cluster) GenerateGitPullCreds(namespace, name string, gitServerInfo types.GitServerInfo) *corev1.Secret {
	message.Debugf("k8s.GenerateGitPullCreds(%s, %s, gitServerInfo)", namespace, name)

	gitServerSecret := c.GenerateSecret(namespace, name, corev1.SecretTypeOpaque)
	gitServerSecret.StringData = map[string]string{
		"username": gitServerInfo.PullUsername,
		"password": gitServerInfo.PullPassword,
	}

	return gitServerSecret
}

// UpdateJackalManagedImageSecrets updates all Jackal-managed image secrets in all namespaces based on state
func (c *Cluster) UpdateJackalManagedImageSecrets(state *types.JackalState) {
	spinner := message.NewProgressSpinner("Updating existing Jackal-managed image secrets")
	defer spinner.Stop()

	if namespaces, err := c.GetNamespaces(); err != nil {
		spinner.Errorf(err, "Unable to get k8s namespaces")
	} else {
		// Update all image pull secrets
		for _, namespace := range namespaces.Items {
			currentRegistrySecret, err := c.GetSecret(namespace.Name, config.JackalImagePullSecretName)
			if err != nil {
				continue
			}

			// Check if this is a Jackal managed secret or is in a namespace the Jackal agent will take action in
			if currentRegistrySecret.Labels[config.JackalManagedByLabel] == "jackal" ||
				(namespace.Labels[agentLabel] != "skip" && namespace.Labels[agentLabel] != "ignore") {
				spinner.Updatef("Updating existing Jackal-managed image secret for namespace: '%s'", namespace.Name)

				// Create the secret
				newRegistrySecret := c.GenerateRegistryPullCreds(namespace.Name, config.JackalImagePullSecretName, state.RegistryInfo)
				if !reflect.DeepEqual(currentRegistrySecret.Data, newRegistrySecret.Data) {
					// Create or update the jackal registry secret
					if _, err := c.CreateOrUpdateSecret(newRegistrySecret); err != nil {
						message.WarnErrf(err, "Problem creating registry secret for the %s namespace", namespace.Name)
					}
				}
			}
		}
		spinner.Success()
	}
}

// UpdateJackalManagedGitSecrets updates all Jackal-managed git secrets in all namespaces based on state
func (c *Cluster) UpdateJackalManagedGitSecrets(state *types.JackalState) {
	spinner := message.NewProgressSpinner("Updating existing Jackal-managed git secrets")
	defer spinner.Stop()

	if namespaces, err := c.GetNamespaces(); err != nil {
		spinner.Errorf(err, "Unable to get k8s namespaces")
	} else {
		// Update all git pull secrets
		for _, namespace := range namespaces.Items {
			currentGitSecret, err := c.GetSecret(namespace.Name, config.JackalGitServerSecretName)
			if err != nil {
				continue
			}

			// Check if this is a Jackal managed secret or is in a namespace the Jackal agent will take action in
			if currentGitSecret.Labels[config.JackalManagedByLabel] == "jackal" ||
				(namespace.Labels[agentLabel] != "skip" && namespace.Labels[agentLabel] != "ignore") {
				spinner.Updatef("Updating existing Jackal-managed git secret for namespace: '%s'", namespace.Name)

				// Create the secret
				newGitSecret := c.GenerateGitPullCreds(namespace.Name, config.JackalGitServerSecretName, state.GitServer)
				if !reflect.DeepEqual(currentGitSecret.StringData, newGitSecret.StringData) {
					// Create or update the jackal git secret
					if _, err := c.CreateOrUpdateSecret(newGitSecret); err != nil {
						message.WarnErrf(err, "Problem creating git server secret for the %s namespace", namespace.Name)
					}
				}
			}
		}
		spinner.Success()
	}
}

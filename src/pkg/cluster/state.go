// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package cluster contains Jackal-specific cluster management functions.
package cluster

import (
	"encoding/json"
	"fmt"
	"time"

	"slices"

	"github.com/defenseunicorns/jackal/src/config"
	"github.com/defenseunicorns/jackal/src/config/lang"
	"github.com/defenseunicorns/jackal/src/types"
	"github.com/fatih/color"

	"github.com/defenseunicorns/jackal/src/pkg/k8s"
	"github.com/defenseunicorns/jackal/src/pkg/message"
	"github.com/defenseunicorns/jackal/src/pkg/pki"
	"github.com/defenseunicorns/pkg/helpers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Jackal Cluster Constants.
const (
	JackalNamespaceName       = "jackal"
	JackalStateSecretName     = "jackal-state"
	JackalStateDataKey        = "state"
	JackalPackageInfoLabel    = "package-deploy-info"
	JackalInitPackageInfoName = "jackal-package-init"
)

// InitJackalState initializes the Jackal state with the given temporary directory and init configs.
func (c *Cluster) InitJackalState(initOptions types.JackalInitOptions) error {
	var (
		distro string
		err    error
	)

	spinner := message.NewProgressSpinner("Gathering cluster state information")
	defer spinner.Stop()

	// Attempt to load an existing state prior to init.
	// NOTE: We are ignoring the error here because we don't really expect a state to exist yet.
	spinner.Updatef("Checking cluster for existing Jackal deployment")
	state, _ := c.LoadJackalState()

	// If state is nil, this is a new cluster.
	if state == nil {
		state = &types.JackalState{}
		spinner.Updatef("New cluster, no prior Jackal deployments found")

		// If the K3s component is being deployed, skip distro detection.
		if initOptions.ApplianceMode {
			distro = k8s.DistroIsK3s
			state.JackalAppliance = true
		} else {
			// Otherwise, trying to detect the K8s distro type.
			distro, err = c.DetectDistro()
			if err != nil {
				// This is a basic failure right now but likely could be polished to provide user guidance to resolve.
				return fmt.Errorf("unable to connect to the cluster to verify the distro: %w", err)
			}
		}

		if distro != k8s.DistroIsUnknown {
			spinner.Updatef("Detected K8s distro %s", distro)
		}

		// Defaults
		state.Distro = distro
		if state.LoggingSecret, err = helpers.RandomString(types.JackalGeneratedPasswordLen); err != nil {
			return fmt.Errorf("%s: %w", lang.ErrUnableToGenerateRandomSecret, err)
		}

		// Setup jackal agent PKI
		state.AgentTLS = pki.GeneratePKI(config.JackalAgentHost)

		namespaces, err := c.GetNamespaces()
		if err != nil {
			return fmt.Errorf("unable to get the Kubernetes namespaces: %w", err)
		}
		// Mark existing namespaces as ignored for the jackal agent to prevent mutating resources we don't own.
		for _, namespace := range namespaces.Items {
			spinner.Updatef("Marking existing namespace %s as ignored by Jackal Agent", namespace.Name)
			if namespace.Labels == nil {
				// Ensure label map exists to avoid nil panic
				namespace.Labels = make(map[string]string)
			}
			// This label will tell the Jackal Agent to ignore this namespace.
			namespace.Labels[agentLabel] = "ignore"
			namespaceCopy := namespace
			if _, err = c.UpdateNamespace(&namespaceCopy); err != nil {
				// This is not a hard failure, but we should log it.
				message.WarnErrf(err, "Unable to mark the namespace %s as ignored by Jackal Agent", namespace.Name)
			}
		}

		// Try to create the jackal namespace.
		spinner.Updatef("Creating the Jackal namespace")
		jackalNamespace := c.NewJackalManagedNamespace(JackalNamespaceName)
		if _, err := c.CreateNamespace(jackalNamespace); err != nil {
			return fmt.Errorf("unable to create the jackal namespace: %w", err)
		}

		// Wait up to 2 minutes for the default service account to be created.
		// Some clusters seem to take a while to create this, see https://github.com/kubernetes/kubernetes/issues/66689.
		// The default SA is required for pods to start properly.
		if _, err := c.WaitForServiceAccount(JackalNamespaceName, "default", 2*time.Minute); err != nil {
			return fmt.Errorf("unable get default Jackal service account: %w", err)
		}

		err = initOptions.GitServer.FillInEmptyValues()
		if err != nil {
			return err
		}
		state.GitServer = initOptions.GitServer
		err = initOptions.RegistryInfo.FillInEmptyValues()
		if err != nil {
			return err
		}
		state.RegistryInfo = initOptions.RegistryInfo
		initOptions.ArtifactServer.FillInEmptyValues()
		state.ArtifactServer = initOptions.ArtifactServer
	} else {
		if helpers.IsNotZeroAndNotEqual(initOptions.GitServer, state.GitServer) {
			message.Warn("Detected a change in Git Server init options on a re-init. Ignoring... To update run:")
			message.JackalCommand("tools update-creds git")
		}
		if helpers.IsNotZeroAndNotEqual(initOptions.RegistryInfo, state.RegistryInfo) {
			message.Warn("Detected a change in Image Registry init options on a re-init. Ignoring... To update run:")
			message.JackalCommand("tools update-creds registry")
		}
		if helpers.IsNotZeroAndNotEqual(initOptions.ArtifactServer, state.ArtifactServer) {
			message.Warn("Detected a change in Artifact Server init options on a re-init. Ignoring... To update run:")
			message.JackalCommand("tools update-creds artifact")
		}
	}

	switch state.Distro {
	case k8s.DistroIsK3s, k8s.DistroIsK3d:
		state.StorageClass = "local-path"

	case k8s.DistroIsKind, k8s.DistroIsGKE:
		state.StorageClass = "standard"

	case k8s.DistroIsDockerDesktop:
		state.StorageClass = "hostpath"
	}

	if initOptions.StorageClass != "" {
		state.StorageClass = initOptions.StorageClass
	}

	spinner.Success()

	// Save the state back to K8s
	if err := c.SaveJackalState(state); err != nil {
		return fmt.Errorf("unable to save the Jackal state: %w", err)
	}

	return nil
}

// LoadJackalState returns the current jackal/jackal-state secret data or an empty JackalState.
func (c *Cluster) LoadJackalState() (state *types.JackalState, err error) {
	// Set up the API connection
	secret, err := c.GetSecret(JackalNamespaceName, JackalStateSecretName)
	if err != nil {
		return nil, fmt.Errorf("%w. %s", err, message.ColorWrap("Did you remember to jackal init?", color.Bold))
	}

	err = json.Unmarshal(secret.Data[JackalStateDataKey], &state)
	if err != nil {
		return nil, err
	}

	c.debugPrintJackalState(state)

	return state, nil
}

func (c *Cluster) sanitizeJackalState(state *types.JackalState) *types.JackalState {
	// Overwrite the AgentTLS information
	state.AgentTLS.CA = []byte("**sanitized**")
	state.AgentTLS.Cert = []byte("**sanitized**")
	state.AgentTLS.Key = []byte("**sanitized**")

	// Overwrite the GitServer passwords
	state.GitServer.PushPassword = "**sanitized**"
	state.GitServer.PullPassword = "**sanitized**"

	// Overwrite the RegistryInfo passwords
	state.RegistryInfo.PushPassword = "**sanitized**"
	state.RegistryInfo.PullPassword = "**sanitized**"
	state.RegistryInfo.Secret = "**sanitized**"

	// Overwrite the ArtifactServer secret
	state.ArtifactServer.PushToken = "**sanitized**"

	// Overwrite the Logging secret
	state.LoggingSecret = "**sanitized**"

	return state
}

func (c *Cluster) debugPrintJackalState(state *types.JackalState) {
	if state == nil {
		return
	}
	// this is a shallow copy, nested pointers WILL NOT be copied
	oldState := *state
	sanitized := c.sanitizeJackalState(&oldState)
	message.Debugf("JackalState - %s", message.JSONValue(sanitized))
}

// SaveJackalState takes a given state and persists it to the Jackal/jackal-state secret.
func (c *Cluster) SaveJackalState(state *types.JackalState) error {
	c.debugPrintJackalState(state)

	// Convert the data back to JSON.
	data, err := json.Marshal(&state)
	if err != nil {
		return err
	}

	// Set up the data wrapper.
	dataWrapper := make(map[string][]byte)
	dataWrapper[JackalStateDataKey] = data

	// The secret object.
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      JackalStateSecretName,
			Namespace: JackalNamespaceName,
			Labels: map[string]string{
				config.JackalManagedByLabel: "jackal",
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: dataWrapper,
	}

	// Attempt to create or update the secret and return.
	if _, err := c.CreateOrUpdateSecret(secret); err != nil {
		return fmt.Errorf("unable to create the jackal state secret")
	}

	return nil
}

// MergeJackalState merges init options for provided services into the provided state to create a new state struct
func (c *Cluster) MergeJackalState(oldState *types.JackalState, initOptions types.JackalInitOptions, services []string) (*types.JackalState, error) {
	newState := *oldState
	var err error
	if slices.Contains(services, message.RegistryKey) {
		newState.RegistryInfo = helpers.MergeNonZero(newState.RegistryInfo, initOptions.RegistryInfo)
		// Set the state of the internal registry if it has changed
		if newState.RegistryInfo.Address == fmt.Sprintf("%s:%d", helpers.IPV4Localhost, newState.RegistryInfo.NodePort) {
			newState.RegistryInfo.InternalRegistry = true
		} else {
			newState.RegistryInfo.InternalRegistry = false
		}

		// Set the new passwords if they should be autogenerated
		if newState.RegistryInfo.PushPassword == oldState.RegistryInfo.PushPassword && oldState.RegistryInfo.InternalRegistry {
			if newState.RegistryInfo.PushPassword, err = helpers.RandomString(types.JackalGeneratedPasswordLen); err != nil {
				return nil, fmt.Errorf("%s: %w", lang.ErrUnableToGenerateRandomSecret, err)
			}
		}
		if newState.RegistryInfo.PullPassword == oldState.RegistryInfo.PullPassword && oldState.RegistryInfo.InternalRegistry {
			if newState.RegistryInfo.PullPassword, err = helpers.RandomString(types.JackalGeneratedPasswordLen); err != nil {
				return nil, fmt.Errorf("%s: %w", lang.ErrUnableToGenerateRandomSecret, err)
			}
		}
	}
	if slices.Contains(services, message.GitKey) {
		newState.GitServer = helpers.MergeNonZero(newState.GitServer, initOptions.GitServer)

		// Set the state of the internal git server if it has changed
		if newState.GitServer.Address == types.JackalInClusterGitServiceURL {
			newState.GitServer.InternalServer = true
		} else {
			newState.GitServer.InternalServer = false
		}

		// Set the new passwords if they should be autogenerated
		if newState.GitServer.PushPassword == oldState.GitServer.PushPassword && oldState.GitServer.InternalServer {
			if newState.GitServer.PushPassword, err = helpers.RandomString(types.JackalGeneratedPasswordLen); err != nil {
				return nil, fmt.Errorf("%s: %w", lang.ErrUnableToGenerateRandomSecret, err)
			}
		}
		if newState.GitServer.PullPassword == oldState.GitServer.PullPassword && oldState.GitServer.InternalServer {
			if newState.GitServer.PullPassword, err = helpers.RandomString(types.JackalGeneratedPasswordLen); err != nil {
				return nil, fmt.Errorf("%s: %w", lang.ErrUnableToGenerateRandomSecret, err)
			}
		}
	}
	if slices.Contains(services, message.ArtifactKey) {
		newState.ArtifactServer = helpers.MergeNonZero(newState.ArtifactServer, initOptions.ArtifactServer)

		// Set the state of the internal artifact server if it has changed
		if newState.ArtifactServer.Address == types.JackalInClusterArtifactServiceURL {
			newState.ArtifactServer.InternalServer = true
		} else {
			newState.ArtifactServer.InternalServer = false
		}

		// Set an empty token if it should be autogenerated
		if newState.ArtifactServer.PushToken == oldState.ArtifactServer.PushToken && oldState.ArtifactServer.InternalServer {
			newState.ArtifactServer.PushToken = ""
		}
	}
	if slices.Contains(services, message.AgentKey) {
		newState.AgentTLS = pki.GeneratePKI(config.JackalAgentHost)
	}

	return &newState, nil
}

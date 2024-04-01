// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package helm contains operations for working with helm charts.
package helm

import (
	"fmt"

	"github.com/racer159/jackal/src/pkg/cluster"
	"github.com/racer159/jackal/src/pkg/k8s"
	"github.com/racer159/jackal/src/pkg/message"
	"github.com/racer159/jackal/src/pkg/transform"
	"github.com/racer159/jackal/src/pkg/utils"
	"github.com/racer159/jackal/src/types"
	"helm.sh/helm/v3/pkg/action"
)

// UpdateJackalRegistryValues updates the Jackal registry deployment with the new state values
func (h *Helm) UpdateJackalRegistryValues() error {
	pushUser, err := utils.GetHtpasswdString(h.cfg.State.RegistryInfo.PushUsername, h.cfg.State.RegistryInfo.PushPassword)
	if err != nil {
		return fmt.Errorf("error generating htpasswd string: %w", err)
	}

	pullUser, err := utils.GetHtpasswdString(h.cfg.State.RegistryInfo.PullUsername, h.cfg.State.RegistryInfo.PullPassword)
	if err != nil {
		return fmt.Errorf("error generating htpasswd string: %w", err)
	}

	registryValues := map[string]interface{}{
		"secrets": map[string]interface{}{
			"htpasswd": fmt.Sprintf("%s\n%s", pushUser, pullUser),
		},
	}

	h.chart = types.JackalChart{
		Namespace:   "jackal",
		ReleaseName: "jackal-docker-registry",
	}

	err = h.UpdateReleaseValues(registryValues)
	if err != nil {
		return fmt.Errorf("error updating the release values: %w", err)
	}

	return nil
}

// UpdateJackalAgentValues updates the Jackal agent deployment with the new state values
func (h *Helm) UpdateJackalAgentValues() error {
	spinner := message.NewProgressSpinner("Gathering information to update Jackal Agent TLS")
	defer spinner.Stop()

	err := h.createActionConfig(cluster.JackalNamespaceName, spinner)
	if err != nil {
		return fmt.Errorf("unable to initialize the K8s client: %w", err)
	}

	// Get the current agent image from one of its pods.
	pods := h.cluster.WaitForPodsAndContainers(k8s.PodLookup{
		Namespace: cluster.JackalNamespaceName,
		Selector:  "app=agent-hook",
	}, nil)

	var currentAgentImage transform.Image
	if len(pods) > 0 && len(pods[0].Spec.Containers) > 0 {
		currentAgentImage, err = transform.ParseImageRef(pods[0].Spec.Containers[0].Image)
		if err != nil {
			return fmt.Errorf("unable to parse current agent image reference: %w", err)
		}
	} else {
		return fmt.Errorf("unable to get current agent pod")
	}

	// List the releases to find the current agent release name.
	listClient := action.NewList(h.actionConfig)

	releases, err := listClient.Run()
	if err != nil {
		return fmt.Errorf("unable to list helm releases: %w", err)
	}

	spinner.Success()

	for _, release := range releases {
		// Update the Jackal Agent release with the new values
		if release.Chart.Name() == "raw-init-jackal-agent-jackal-agent" {
			h.chart = types.JackalChart{
				Namespace:   "jackal",
				ReleaseName: release.Name,
			}
			h.component = types.JackalComponent{
				Name: "jackal-agent",
			}
			h.cfg.Pkg.Constants = []types.JackalPackageConstant{
				{
					Name:  "AGENT_IMAGE",
					Value: currentAgentImage.Path,
				},
				{
					Name:  "AGENT_IMAGE_TAG",
					Value: currentAgentImage.Tag,
				},
			}

			err := h.UpdateReleaseValues(map[string]interface{}{})
			if err != nil {
				return fmt.Errorf("error updating the release values: %w", err)
			}
		}
	}

	spinner = message.NewProgressSpinner("Cleaning up Jackal Agent pods after update")
	defer spinner.Stop()

	// Force pods to be recreated to get the updated secret.
	err = h.cluster.DeletePods(k8s.PodLookup{
		Namespace: cluster.JackalNamespaceName,
		Selector:  "app=agent-hook",
	})
	if err != nil {
		return fmt.Errorf("error recycling pods for the Jackal Agent: %w", err)
	}

	spinner.Success()

	return nil
}

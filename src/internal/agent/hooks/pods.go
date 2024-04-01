// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package hooks provides HTTP handlers for the mutating webhook.
package hooks

import (
	"encoding/json"
	"fmt"

	"github.com/defenseunicorns/jackal/src/config"
	"github.com/defenseunicorns/jackal/src/config/lang"
	"github.com/defenseunicorns/jackal/src/internal/agent/operations"
	"github.com/defenseunicorns/jackal/src/internal/agent/state"
	"github.com/defenseunicorns/jackal/src/pkg/message"
	"github.com/defenseunicorns/jackal/src/pkg/transform"
	v1 "k8s.io/api/admission/v1"

	corev1 "k8s.io/api/core/v1"
)

// NewPodMutationHook creates a new instance of pods mutation hook.
func NewPodMutationHook() operations.Hook {
	message.Debug("hooks.NewMutationHook()")
	return operations.Hook{
		Create: mutatePod,
		Update: mutatePod,
	}
}

func parsePod(object []byte) (*corev1.Pod, error) {
	message.Debugf("pods.parsePod(%s)", string(object))
	var pod corev1.Pod
	if err := json.Unmarshal(object, &pod); err != nil {
		return nil, err
	}

	return &pod, nil
}

func mutatePod(r *v1.AdmissionRequest) (*operations.Result, error) {
	message.Debugf("hooks.mutatePod()(*v1.AdmissionRequest) - %#v , %s/%s: %#v", r.Kind, r.Namespace, r.Name, r.Operation)

	var patchOperations []operations.PatchOperation
	pod, err := parsePod(r.Object.Raw)
	if err != nil {
		return &operations.Result{Msg: err.Error()}, nil
	}

	if pod.Labels != nil && pod.Labels["jackal-agent"] == "patched" {
		// We've already played with this pod, just keep swimming 🐟
		return &operations.Result{
			Allowed:  true,
			PatchOps: patchOperations,
		}, nil
	}

	// Add the jackal secret to the podspec
	jackalSecret := []corev1.LocalObjectReference{{Name: config.JackalImagePullSecretName}}
	patchOperations = append(patchOperations, operations.ReplacePatchOperation("/spec/imagePullSecrets", jackalSecret))

	jackalState, err := state.GetJackalStateFromAgentPod()
	if err != nil {
		return nil, fmt.Errorf(lang.AgentErrGetState, err)
	}
	containerRegistryURL := jackalState.RegistryInfo.Address

	// update the image host for each init container
	for idx, container := range pod.Spec.InitContainers {
		path := fmt.Sprintf("/spec/initContainers/%d/image", idx)
		replacement, err := transform.ImageTransformHost(containerRegistryURL, container.Image)
		if err != nil {
			message.Warnf(lang.AgentErrImageSwap, container.Image)
			continue // Continue, because we might as well attempt to mutate the other containers for this pod
		}
		patchOperations = append(patchOperations, operations.ReplacePatchOperation(path, replacement))
	}

	// update the image host for each ephemeral container
	for idx, container := range pod.Spec.EphemeralContainers {
		path := fmt.Sprintf("/spec/ephemeralContainers/%d/image", idx)
		replacement, err := transform.ImageTransformHost(containerRegistryURL, container.Image)
		if err != nil {
			message.Warnf(lang.AgentErrImageSwap, container.Image)
			continue // Continue, because we might as well attempt to mutate the other containers for this pod
		}
		patchOperations = append(patchOperations, operations.ReplacePatchOperation(path, replacement))
	}

	// update the image host for each normal container
	for idx, container := range pod.Spec.Containers {
		path := fmt.Sprintf("/spec/containers/%d/image", idx)
		replacement, err := transform.ImageTransformHost(containerRegistryURL, container.Image)
		if err != nil {
			message.Warnf(lang.AgentErrImageSwap, container.Image)
			continue // Continue, because we might as well attempt to mutate the other containers for this pod
		}
		patchOperations = append(patchOperations, operations.ReplacePatchOperation(path, replacement))
	}

	// Add a label noting the jackal mutation
	patchOperations = append(patchOperations, operations.ReplacePatchOperation("/metadata/labels/jackal-agent", "patched"))

	return &operations.Result{
		Allowed:  true,
		PatchOps: patchOperations,
	}, nil
}

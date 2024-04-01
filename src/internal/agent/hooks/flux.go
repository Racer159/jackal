// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package hooks contains the mutation hooks for the Jackal agent.
package hooks

import (
	"encoding/json"
	"fmt"

	"github.com/defenseunicorns/pkg/helpers"
	"github.com/racer159/jackal/src/config"
	"github.com/racer159/jackal/src/config/lang"
	"github.com/racer159/jackal/src/internal/agent/operations"
	"github.com/racer159/jackal/src/internal/agent/state"
	"github.com/racer159/jackal/src/pkg/message"
	"github.com/racer159/jackal/src/pkg/transform"
	"github.com/racer159/jackal/src/types"
	v1 "k8s.io/api/admission/v1"
)

// SecretRef contains the name used to reference a git repository secret.
type SecretRef struct {
	Name string `json:"name"`
}

// GenericGitRepo contains the URL of a git repo and the secret that corresponds to it for use with Flux.
type GenericGitRepo struct {
	Spec struct {
		URL       string    `json:"url"`
		SecretRef SecretRef `json:"secretRef,omitempty"`
	} `json:"spec"`
}

// NewGitRepositoryMutationHook creates a new instance of the git repo mutation hook.
func NewGitRepositoryMutationHook() operations.Hook {
	message.Debug("hooks.NewGitRepositoryMutationHook()")
	return operations.Hook{
		Create: mutateGitRepo,
		Update: mutateGitRepo,
	}
}

// mutateGitRepoCreate mutates the git repository url to point to the repository URL defined in the JackalState.
func mutateGitRepo(r *v1.AdmissionRequest) (result *operations.Result, err error) {

	var (
		jackalState *types.JackalState
		patches     []operations.PatchOperation
		isPatched   bool

		isCreate = r.Operation == v1.Create
		isUpdate = r.Operation == v1.Update
	)

	// Form the jackalState.GitServer.Address from the jackalState
	if jackalState, err = state.GetJackalStateFromAgentPod(); err != nil {
		return nil, fmt.Errorf(lang.AgentErrGetState, err)
	}

	message.Debugf("Using the url of (%s) to mutate the flux repository", jackalState.GitServer.Address)

	// parse to simple struct to read the git url
	src := &GenericGitRepo{}
	if err = json.Unmarshal(r.Object.Raw, &src); err != nil {
		return nil, fmt.Errorf(lang.ErrUnmarshal, err)
	}
	patchedURL := src.Spec.URL

	// Check if this is an update operation and the hostname is different from what we have in the jackalState
	// NOTE: We mutate on updates IF AND ONLY IF the hostname in the request is different than the hostname in the jackalState
	// NOTE: We are checking if the hostname is different before because we do not want to potentially mutate a URL that has already been mutated.
	if isUpdate {
		isPatched, err = helpers.DoHostnamesMatch(jackalState.GitServer.Address, src.Spec.URL)
		if err != nil {
			return nil, fmt.Errorf(lang.AgentErrHostnameMatch, err)
		}
	}

	// Mutate the git URL if necessary
	if isCreate || (isUpdate && !isPatched) {
		// Mutate the git URL so that the hostname matches the hostname in the Jackal state
		transformedURL, err := transform.GitURL(jackalState.GitServer.Address, patchedURL, jackalState.GitServer.PushUsername)
		if err != nil {
			message.Warnf("Unable to transform the git url, using the original url we have: %s", patchedURL)
		}
		patchedURL = transformedURL.String()
		message.Debugf("original git URL of (%s) got mutated to (%s)", src.Spec.URL, patchedURL)
	}

	// Patch updates of the repo spec
	patches = populatePatchOperations(patchedURL, src.Spec.SecretRef.Name)

	return &operations.Result{
		Allowed:  true,
		PatchOps: patches,
	}, nil
}

// Patch updates of the repo spec.
func populatePatchOperations(repoURL string, secretName string) []operations.PatchOperation {
	var patches []operations.PatchOperation
	patches = append(patches, operations.ReplacePatchOperation("/spec/url", repoURL))

	// If a prior secret exists, replace it
	if secretName != "" {
		patches = append(patches, operations.ReplacePatchOperation("/spec/secretRef/name", config.JackalGitServerSecretName))
	} else {
		// Otherwise, add the new secret
		patches = append(patches, operations.AddPatchOperation("/spec/secretRef", SecretRef{Name: config.JackalGitServerSecretName}))
	}

	return patches
}

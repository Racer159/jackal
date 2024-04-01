// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package hooks contains the mutation hooks for the Jackal agent.
package hooks

import (
	"encoding/json"
	"fmt"

	"github.com/defenseunicorns/pkg/helpers"
	"github.com/racer159/jackal/src/config/lang"
	"github.com/racer159/jackal/src/internal/agent/operations"
	"github.com/racer159/jackal/src/internal/agent/state"
	"github.com/racer159/jackal/src/pkg/message"
	"github.com/racer159/jackal/src/pkg/transform"
	"github.com/racer159/jackal/src/types"
	v1 "k8s.io/api/admission/v1"
)

// Source represents a subset of the Argo Source object needed for Jackal Git URL mutations
type Source struct {
	RepoURL string `json:"repoURL"`
}

// ArgoApplication represents a subset of the Argo Application object needed for Jackal Git URL mutations
type ArgoApplication struct {
	Spec struct {
		Source  Source   `json:"source"`
		Sources []Source `json:"sources"`
	} `json:"spec"`
}

var (
	jackalState *types.JackalState
	patches     []operations.PatchOperation
	isPatched   bool
	isCreate    bool
	isUpdate    bool
)

// NewApplicationMutationHook creates a new instance of the ArgoCD Application mutation hook.
func NewApplicationMutationHook() operations.Hook {
	message.Debug("hooks.NewApplicationMutationHook()")
	return operations.Hook{
		Create: mutateApplication,
		Update: mutateApplication,
	}
}

// mutateApplication mutates the git repository url to point to the repository URL defined in the JackalState.
func mutateApplication(r *v1.AdmissionRequest) (result *operations.Result, err error) {

	isCreate = r.Operation == v1.Create
	isUpdate = r.Operation == v1.Update

	patches = []operations.PatchOperation{}

	// Form the jackalState.GitServer.Address from the jackalState
	if jackalState, err = state.GetJackalStateFromAgentPod(); err != nil {
		return nil, fmt.Errorf(lang.AgentErrGetState, err)
	}

	message.Debugf("Using the url of (%s) to mutate the ArgoCD Application", jackalState.GitServer.Address)

	// parse to simple struct to read the git url
	src := &ArgoApplication{}

	if err = json.Unmarshal(r.Object.Raw, &src); err != nil {
		return nil, fmt.Errorf(lang.ErrUnmarshal, err)
	}

	message.Debugf("Data %v", string(r.Object.Raw))

	if src.Spec.Source != (Source{}) {
		patchedURL, _ := getPatchedRepoURL(src.Spec.Source.RepoURL)
		patches = populateSingleSourceArgoApplicationPatchOperations(patchedURL, patches)
	}

	if len(src.Spec.Sources) > 0 {
		for idx, source := range src.Spec.Sources {
			patchedURL, _ := getPatchedRepoURL(source.RepoURL)
			patches = populateMultipleSourceArgoApplicationPatchOperations(idx, patchedURL, patches)
		}
	}

	return &operations.Result{
		Allowed:  true,
		PatchOps: patches,
	}, nil
}

func getPatchedRepoURL(repoURL string) (string, error) {
	var err error
	patchedURL := repoURL

	// Check if this is an update operation and the hostname is different from what we have in the jackalState
	// NOTE: We mutate on updates IF AND ONLY IF the hostname in the request is different from the hostname in the jackalState
	// NOTE: We are checking if the hostname is different before because we do not want to potentially mutate a URL that has already been mutated.
	if isUpdate {
		isPatched, err = helpers.DoHostnamesMatch(jackalState.GitServer.Address, repoURL)
		if err != nil {
			return "", fmt.Errorf(lang.AgentErrHostnameMatch, err)
		}
	}

	// Mutate the repoURL if necessary
	if isCreate || (isUpdate && !isPatched) {
		// Mutate the git URL so that the hostname matches the hostname in the Jackal state
		transformedURL, err := transform.GitURL(jackalState.GitServer.Address, patchedURL, jackalState.GitServer.PushUsername)
		if err != nil {
			message.Warnf("Unable to transform the repoURL, using the original url we have: %s", patchedURL)
		}
		patchedURL = transformedURL.String()
		message.Debugf("original repoURL of (%s) got mutated to (%s)", repoURL, patchedURL)
	}

	return patchedURL, err
}

// Patch updates of the Argo source spec.
func populateSingleSourceArgoApplicationPatchOperations(repoURL string, patches []operations.PatchOperation) []operations.PatchOperation {
	return append(patches, operations.ReplacePatchOperation("/spec/source/repoURL", repoURL))
}

// Patch updates of the Argo sources spec.
func populateMultipleSourceArgoApplicationPatchOperations(idx int, repoURL string, patches []operations.PatchOperation) []operations.PatchOperation {
	return append(patches, operations.ReplacePatchOperation(fmt.Sprintf("/spec/sources/%d/repoURL", idx), repoURL))
}

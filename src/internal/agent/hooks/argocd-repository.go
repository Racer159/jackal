// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package hooks contains the mutation hooks for the Zarf agent.
package hooks

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/defenseunicorns/zarf/src/config/lang"
	"github.com/defenseunicorns/zarf/src/internal/agent/operations"
	"github.com/defenseunicorns/zarf/src/internal/agent/state"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/transform"
	"github.com/defenseunicorns/zarf/src/pkg/utils/helpers"
	"github.com/defenseunicorns/zarf/src/types"
	v1 "k8s.io/api/admission/v1"
)

// ArgoRepository represents a subset of the Argo Repository object needed for Zarf Git URL mutations
type ArgoRepository struct {
	Data struct {
		URL string `json:"url"`
	}
}

// NewRepositoryMutationHook creates a new instance of the ArgoCD Repository mutation hook.
func NewRepositoryMutationHook() operations.Hook {
	message.Debug("hooks.NewRepositoryMutationHook()")
	return operations.Hook{
		Create: mutateRepository,
		Update: mutateRepository,
	}
}

// mutateRepository mutates the git repository URL to point to the repository URL defined in the ZarfState.
func mutateRepository(r *v1.AdmissionRequest) (result *operations.Result, err error) {

	var (
		zarfState *types.ZarfState
		patches   []operations.PatchOperation
		isPatched bool

		isCreate = r.Operation == v1.Create
		isUpdate = r.Operation == v1.Update
	)

	// Form the zarfState.GitServer.Address from the zarfState
	if zarfState, err = state.GetZarfStateFromAgentPod(); err != nil {
		return nil, fmt.Errorf(lang.AgentErrGetState, err)
	}

	message.Debugf("Using the url of (%s) to mutate the ArgoCD Repository Secret", zarfState.GitServer.Address)

	// Parse into a simple struct to read the git url
	src := &ArgoRepository{}

	if err = json.Unmarshal(r.Object.Raw, &src); err != nil {
		return nil, fmt.Errorf(lang.ErrUnmarshal, err)
	}
	decodedURL, err := base64.StdEncoding.DecodeString(src.Data.URL)
	if err != nil {
		message.Fatalf("Error decoding URL from Repository Secret %s", src.Data.URL)
	}
	src.Data.URL = string(decodedURL)
	patchedURL := src.Data.URL

	// Check if this is an update operation and the hostname is different from what we have in the zarfState
	// NOTE: We mutate on updates IF AND ONLY IF the hostname in the request is different from the hostname in the zarfState
	// NOTE: We are checking if the hostname is different before because we do not want to potentially mutate a URL that has already been mutated.
	if isUpdate {
		isPatched, err = helpers.DoHostnamesMatch(zarfState.GitServer.Address, src.Data.URL)
		if err != nil {
			return nil, fmt.Errorf(lang.AgentErrHostnameMatch, err)
		}
	}

	// Mutate the repoURL if necessary
	if isCreate || (isUpdate && !isPatched) {
		// Mutate the git URL so that the hostname matches the hostname in the Zarf state
		transformedURL, err := transform.GitURL(zarfState.GitServer.Address, patchedURL, zarfState.GitServer.PushUsername)
		if err != nil {
			message.Warnf("Unable to transform the url, using the original url we have: %s", patchedURL)
		}
		patchedURL = transformedURL.String()
		message.Debugf("original url of (%s) got mutated to (%s)", src.Data.URL, patchedURL)
	}

	// Patch updates of the repo spec
	patches = populateArgoRepositoryPatchOperations(patchedURL, zarfState.GitServer.PullPassword)

	return &operations.Result{
		Allowed:  true,
		PatchOps: patches,
	}, nil
}

// Patch updates of the Argo Repository Secret.
func populateArgoRepositoryPatchOperations(repoURL string, zarfGitPullPassword string) []operations.PatchOperation {
	var patches []operations.PatchOperation
	patches = append(patches, operations.ReplacePatchOperation("/data/url", base64.StdEncoding.EncodeToString([]byte(repoURL))))
	patches = append(patches, operations.ReplacePatchOperation("/data/username", base64.StdEncoding.EncodeToString([]byte(types.ZarfGitReadUser))))
	patches = append(patches, operations.ReplacePatchOperation("/data/password", base64.StdEncoding.EncodeToString([]byte(zarfGitPullPassword))))

	return patches
}

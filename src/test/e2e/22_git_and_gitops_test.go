// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package test provides e2e tests for Jackal.
package test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/defenseunicorns/jackal/src/internal/packager/git"
	"github.com/defenseunicorns/jackal/src/pkg/cluster"
	"github.com/defenseunicorns/jackal/src/types"
	"github.com/stretchr/testify/require"
)

func TestGit(t *testing.T) {
	t.Log("E2E: Git")
	e2e.SetupWithCluster(t)

	buildPath := filepath.Join("src", "test", "packages", "22-git-data")
	stdOut, stdErr, err := e2e.Jackal("package", "create", buildPath, "-o=build", "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	path := fmt.Sprintf("build/jackal-package-git-data-test-%s-1.0.0.tar.zst", e2e.Arch)
	defer e2e.CleanFiles(path)

	// Deploy the git data example (with component globbing to test that as well)
	stdOut, stdErr, err = e2e.Jackal("package", "deploy", path, "--components=full-repo,specific-*", "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	c, err := cluster.NewCluster()
	require.NoError(t, err)
	tunnelGit, err := c.Connect(cluster.JackalGit)
	require.NoError(t, err)
	defer tunnelGit.Close()

	testGitServerConnect(t, tunnelGit.HTTPEndpoint())
	testGitServerReadOnly(t, tunnelGit.HTTPEndpoint())
	testGitServerTagAndHash(t, tunnelGit.HTTPEndpoint())
}

func TestGitOpsFlux(t *testing.T) {
	t.Log("E2E: GitOps / Flux")
	e2e.SetupWithCluster(t)

	waitFluxPodInfoDeployment(t)
}

func TestGitOpsArgoCD(t *testing.T) {
	t.Log("E2E: ArgoCD / Flux")
	e2e.SetupWithCluster(t)

	waitArgoDeployment(t)
}

func testGitServerConnect(t *testing.T, gitURL string) {
	// Make sure Gitea comes up cleanly
	resp, err := http.Get(gitURL + "/explore/repos")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
}

func testGitServerReadOnly(t *testing.T, gitURL string) {
	// Init the state variable
	state, err := cluster.NewClusterOrDie().LoadJackalState()
	require.NoError(t, err)

	gitCfg := git.New(state.GitServer)

	// Get the repo as the readonly user
	repoName := "jackal-public-test-2469062884"
	getRepoRequest, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/repos/%s/%s", gitURL, state.GitServer.PushUsername, repoName), nil)
	getRepoResponseBody, _, err := gitCfg.DoHTTPThings(getRepoRequest, types.JackalGitReadUser, state.GitServer.PullPassword)
	require.NoError(t, err)

	// Make sure the only permissions are pull (read)
	var bodyMap map[string]interface{}
	err = json.Unmarshal(getRepoResponseBody, &bodyMap)
	require.NoError(t, err)
	permissionsMap := bodyMap["permissions"].(map[string]interface{})
	require.False(t, permissionsMap["admin"].(bool))
	require.False(t, permissionsMap["push"].(bool))
	require.True(t, permissionsMap["pull"].(bool))
}

func testGitServerTagAndHash(t *testing.T, gitURL string) {
	// Init the state variable
	state, err := cluster.NewClusterOrDie().LoadJackalState()
	require.NoError(t, err, "Failed to load Jackal state")
	repoName := "jackal-public-test-2469062884"

	gitCfg := git.New(state.GitServer)

	// Get the Jackal repo tag
	repoTag := "v0.0.1"
	getRepoTagsRequest, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/repos/%s/%s/tags/%s", gitURL, types.JackalGitPushUser, repoName, repoTag), nil)
	getRepoTagsResponseBody, _, err := gitCfg.DoHTTPThings(getRepoTagsRequest, types.JackalGitReadUser, state.GitServer.PullPassword)
	require.NoError(t, err)

	// Make sure the pushed tag exists
	var tagMap map[string]interface{}
	err = json.Unmarshal(getRepoTagsResponseBody, &tagMap)
	require.NoError(t, err)
	require.Equal(t, repoTag, tagMap["name"])

	// Get the Jackal repo commit
	repoHash := "01a23218923f24194133b5eb11268cf8d73ff1bb"
	getRepoCommitsRequest, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/repos/%s/%s/git/commits/%s", gitURL, types.JackalGitPushUser, repoName, repoHash), nil)
	getRepoCommitsResponseBody, _, err := gitCfg.DoHTTPThings(getRepoCommitsRequest, types.JackalGitReadUser, state.GitServer.PullPassword)
	require.NoError(t, err)
	require.Contains(t, string(getRepoCommitsResponseBody), repoHash)
}

func waitFluxPodInfoDeployment(t *testing.T) {
	// Deploy the flux example and verify that it works
	path := fmt.Sprintf("build/jackal-package-podinfo-flux-%s.tar.zst", e2e.Arch)
	stdOut, stdErr, err := e2e.Jackal("package", "deploy", path, "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	// Tests the URL mutation for GitRepository CRD for Flux.
	stdOut, stdErr, err = e2e.Kubectl("get", "gitrepositories", "podinfo", "-n", "flux-system", "-o", "jsonpath={.spec.url}")
	require.NoError(t, err, stdOut, stdErr)
	expectedMutatedRepoURL := fmt.Sprintf("%s/%s/podinfo-1646971829.git", types.JackalInClusterGitServiceURL, types.JackalGitPushUser)
	require.Equal(t, expectedMutatedRepoURL, stdOut)

	// Remove the flux example when deployment completes
	stdOut, stdErr, err = e2e.Jackal("package", "remove", "podinfo-flux", "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	// Prune the flux images to reduce disk pressure
	stdOut, stdErr, err = e2e.Jackal("tools", "registry", "prune", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
}

func waitArgoDeployment(t *testing.T) {
	// Deploy the argocd example and verify that it works
	path := fmt.Sprintf("build/jackal-package-argocd-%s.tar.zst", e2e.Arch)
	stdOut, stdErr, err := e2e.Jackal("package", "deploy", path, "--components=argocd-apps", "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	expectedMutatedRepoURL := fmt.Sprintf("%s/%s/podinfo-1646971829.git", types.JackalInClusterGitServiceURL, types.JackalGitPushUser)

	// Tests the mutation of the private repository Secret for ArgoCD.
	stdOut, stdErr, err = e2e.Kubectl("get", "secret", "argocd-repo-github-podinfo", "-n", "argocd", "-o", "jsonpath={.data.url}")
	require.NoError(t, err, stdOut, stdErr)

	expectedMutatedPrivateRepoURLSecret, err := base64.StdEncoding.DecodeString(stdOut)
	require.NoError(t, err, stdOut, stdErr)
	require.Equal(t, expectedMutatedRepoURL, string(expectedMutatedPrivateRepoURLSecret))

	// Tests the mutation of the repoURL for Application CRD source(s) for ArgoCD.
	stdOut, stdErr, err = e2e.Kubectl("get", "application", "apps", "-n", "argocd", "-o", "jsonpath={.spec.sources[0].repoURL}")
	require.NoError(t, err, stdOut, stdErr)
	require.Equal(t, expectedMutatedRepoURL, stdOut)

	// Remove the argocd example when deployment completes
	stdOut, stdErr, err = e2e.Jackal("package", "remove", "argocd", "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	// Prune the ArgoCD images to reduce disk pressure
	stdOut, stdErr, err = e2e.Jackal("tools", "registry", "prune", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
}

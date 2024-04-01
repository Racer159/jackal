// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package test provides e2e tests for Jackal.
package test

import (
	"encoding/base64"
	"fmt"
	"runtime"
	"testing"

	"encoding/json"

	"github.com/racer159/jackal/src/types"
	"github.com/stretchr/testify/require"
)

func TestJackalInit(t *testing.T) {
	t.Log("E2E: Jackal init")
	e2e.SetupWithCluster(t)

	initComponents := "logging,git-server"
	// Add k3s component in appliance mode
	if e2e.ApplianceMode {
		initComponents = "k3s,logging,git-server"
	}

	initPackageVersion := e2e.GetJackalVersion(t)

	var (
		mismatchedArch        = e2e.GetMismatchedArch()
		mismatchedInitPackage = fmt.Sprintf("jackal-init-%s-%s.tar.zst", mismatchedArch, initPackageVersion)
		expectedErrorMessage  = "unable to run component before action: command \"Check that the host architecture matches the package architecture\""
	)
	t.Cleanup(func() {
		e2e.CleanFiles(mismatchedInitPackage)
	})

	if runtime.GOOS == "linux" {
		// Build init package with different arch than the cluster arch.
		stdOut, stdErr, err := e2e.Jackal("package", "create", "src/test/packages/20-mismatched-arch-init", "--architecture", mismatchedArch, "--confirm")
		require.NoError(t, err, stdOut, stdErr)

		// Check that `jackal init` returns an error because of the mismatched architectures.
		// We need to use the --architecture flag here to force jackal to find the package.
		_, stdErr, err = e2e.Jackal("init", "--architecture", mismatchedArch, "--components=k3s", "--confirm")
		require.Error(t, err, stdErr)
		require.Contains(t, stdErr, expectedErrorMessage)
	}

	if !e2e.ApplianceMode {
		// throw a pending pod into the cluster to ensure we can properly ignore them when selecting images
		_, _, err := e2e.Kubectl("apply", "-f", "https://raw.githubusercontent.com/kubernetes/website/main/content/en/examples/pods/pod-with-node-affinity.yaml")
		require.NoError(t, err)
	}

	// Check for any old secrets to ensure that they don't get saved in the init log
	oldState := types.JackalState{}
	base64State, _, err := e2e.Kubectl("get", "secret", "jackal-state", "-n", "jackal", "-o", "jsonpath={.data.state}")
	if err == nil {
		oldStateJSON, err := base64.StdEncoding.DecodeString(base64State)
		require.NoError(t, err)
		err = json.Unmarshal(oldStateJSON, &oldState)
		require.NoError(t, err)
	}

	// run `jackal init`
	_, initStdErr, err := e2e.Jackal("init", "--components="+initComponents, "--nodeport", "31337", "-l", "trace", "--confirm")
	require.NoError(t, err)
	require.Contains(t, initStdErr, "an inventory of all software contained in this package")
	require.NotContains(t, initStdErr, "This package does NOT contain an SBOM. If you require an SBOM, please contact the creator of this package to request a version that includes an SBOM.")

	logText := e2e.GetLogFileContents(t, e2e.StripMessageFormatting(initStdErr))

	// Verify that any state secrets were not included in the log
	state := types.JackalState{}
	base64State, _, err = e2e.Kubectl("get", "secret", "jackal-state", "-n", "jackal", "-o", "jsonpath={.data.state}")
	require.NoError(t, err)
	stateJSON, err := base64.StdEncoding.DecodeString(base64State)
	require.NoError(t, err)
	err = json.Unmarshal(stateJSON, &state)
	require.NoError(t, err)
	checkLogForSensitiveState(t, logText, state)

	// Check the old state values as well (if they exist) to ensure they weren't printed and then updated during init
	if oldState.LoggingSecret != "" {
		checkLogForSensitiveState(t, logText, oldState)
	}

	if e2e.ApplianceMode {
		// make sure that we upgraded `k3s` correctly and are running the correct version - this should match that found in `packages/distros/k3s`
		kubeletVersion, _, err := e2e.Kubectl("get", "nodes", "-o", "jsonpath={.items[0].status.nodeInfo.kubeletVersion}")
		require.NoError(t, err)
		require.Contains(t, kubeletVersion, "v1.28.4+k3s2")
	}

	// Check that the registry is running on the correct NodePort
	stdOut, _, err := e2e.Kubectl("get", "service", "-n", "jackal", "jackal-docker-registry", "-o=jsonpath='{.spec.ports[*].nodePort}'")
	require.NoError(t, err)
	require.Contains(t, stdOut, "31337")

	// Check that the registry is running with the correct scale down policy
	stdOut, _, err = e2e.Kubectl("get", "hpa", "-n", "jackal", "jackal-docker-registry", "-o=jsonpath='{.spec.behavior.scaleDown.selectPolicy}'")
	require.NoError(t, err)
	require.Contains(t, stdOut, "Min")

	// Special sizing-hacking for reducing resources where Kind + CI eats a lot of free cycles (ignore errors)
	_, _, _ = e2e.Kubectl("scale", "deploy", "-n", "kube-system", "coredns", "--replicas=1")
	_, _, _ = e2e.Kubectl("scale", "deploy", "-n", "jackal", "agent-hook", "--replicas=1")
}

func checkLogForSensitiveState(t *testing.T, logText string, jackalState types.JackalState) {
	require.NotContains(t, logText, jackalState.AgentTLS.CA)
	require.NotContains(t, logText, string(jackalState.AgentTLS.CA))
	require.NotContains(t, logText, jackalState.AgentTLS.Cert)
	require.NotContains(t, logText, string(jackalState.AgentTLS.Cert))
	require.NotContains(t, logText, jackalState.AgentTLS.Key)
	require.NotContains(t, logText, string(jackalState.AgentTLS.Key))
	require.NotContains(t, logText, jackalState.ArtifactServer.PushToken)
	require.NotContains(t, logText, jackalState.GitServer.PullPassword)
	require.NotContains(t, logText, jackalState.GitServer.PushPassword)
	require.NotContains(t, logText, jackalState.RegistryInfo.PullPassword)
	require.NotContains(t, logText, jackalState.RegistryInfo.PushPassword)
	require.NotContains(t, logText, jackalState.RegistryInfo.Secret)
	require.NotContains(t, logText, jackalState.LoggingSecret)
}

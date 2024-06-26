// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package test provides e2e tests for Jackal.
package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplianceRemove(t *testing.T) {
	t.Log("E2E: Appliance Remove")

	// Don't run this test in appliance mode
	if !e2e.ApplianceMode || e2e.ApplianceModeKeep {
		return
	}

	e2e.SetupWithCluster(t)

	initPackageVersion := e2e.GetJackalVersion(t)

	path := fmt.Sprintf("build/jackal-init-%s-%s.tar.zst", e2e.Arch, initPackageVersion)

	// Package remove the cluster to test Jackal cleaning up after itself
	stdOut, stdErr, err := e2e.Jackal("package", "remove", path, "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	// Check that the cluster is now gone
	_, _, err = e2e.Kubectl("get", "nodes")
	require.Error(t, err)

	// TODO (@WSTARR) - This needs to be refactored to use the remove logic instead of reaching into a magic directory
	// Re-init the cluster so that we can test if the destroy scripts run
	stdOut, stdErr, err = e2e.Jackal("init", "--components=k3s", "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	// Destroy the cluster to test Jackal cleaning up after itself
	stdOut, stdErr, err = e2e.Jackal("destroy", "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	// Check that the cluster gone again
	_, _, err = e2e.Kubectl("get", "nodes")
	require.Error(t, err)
}

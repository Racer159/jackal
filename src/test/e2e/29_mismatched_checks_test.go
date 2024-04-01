// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package test provides e2e tests for Jackal.
package test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/defenseunicorns/jackal/src/types"
	"github.com/stretchr/testify/require"
)

// TestMismatchedArchitectures ensures that jackal produces an error
// when the package architecture doesn't match the target cluster architecture.
func TestMismatchedArchitectures(t *testing.T) {
	t.Log("E2E: Mismatched architectures")
	e2e.SetupWithCluster(t)

	var (
		mismatchedArch         = e2e.GetMismatchedArch()
		mismatchedGamesPackage = fmt.Sprintf("jackal-package-dos-games-%s-1.0.0.tar.zst", mismatchedArch)
		expectedErrorMessage   = fmt.Sprintf("this package architecture is %s", mismatchedArch)
	)

	// Build dos-games package with different arch than the cluster arch.
	stdOut, stdErr, err := e2e.Jackal("package", "create", "examples/dos-games/", "--architecture", mismatchedArch, "--confirm")
	require.NoError(t, err, stdOut, stdErr)
	defer e2e.CleanFiles(mismatchedGamesPackage)

	// Ensure jackal package deploy returns an error because of the mismatched architectures.
	_, stdErr, err = e2e.Jackal("package", "deploy", mismatchedGamesPackage, "--confirm")
	require.Error(t, err, stdErr)
	require.Contains(t, e2e.StripMessageFormatting(stdErr), expectedErrorMessage)
}

// TestMismatchedVersions ensures that jackal produces a warning
// when the initialized version of Jackal doesn't match the current CLI
func TestMismatchedVersions(t *testing.T) {
	t.Log("E2E: Mismatched versions")
	e2e.SetupWithCluster(t)

	var (
		expectedWarningMessage = "Potential Breaking Changes"
	)

	// Get the current init package secret
	initPkg := types.DeployedPackage{}
	base64Pkg, _, err := e2e.Kubectl("get", "secret", "jackal-package-init", "-n", "jackal", "-o", "jsonpath={.data.data}")
	require.NoError(t, err)
	jsonPkg, err := base64.StdEncoding.DecodeString(base64Pkg)
	require.NoError(t, err)
	fmt.Println(string(jsonPkg))
	err = json.Unmarshal(jsonPkg, &initPkg)
	require.NoError(t, err)

	// Edit the build data to trigger the breaking change check
	initPkg.Data.Build.Version = "v0.25.0"

	// Delete the package secret
	_, _, err = e2e.Kubectl("delete", "secret", "jackal-package-init", "-n", "jackal")
	require.NoError(t, err)

	// Create a new secret with the modified data
	jsonPkgModified, err := json.Marshal(initPkg)
	require.NoError(t, err)
	_, _, err = e2e.Kubectl("create", "secret", "generic", "jackal-package-init", "-n", "jackal", fmt.Sprintf("--from-literal=data=%s", string(jsonPkgModified)))
	require.NoError(t, err)

	path := filepath.Join("build", fmt.Sprintf("jackal-package-dos-games-%s-1.0.0.tar.zst", e2e.Arch))

	// Deploy the games package
	stdOut, stdErr, err := e2e.Jackal("package", "deploy", path, "--confirm")
	require.NoError(t, err, stdOut, stdErr)
	require.Contains(t, stdErr, expectedWarningMessage)

	// Remove the games package
	stdOut, stdErr, err = e2e.Jackal("package", "remove", "dos-games", "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	// Reset the package secret
	_, _, err = e2e.Kubectl("delete", "secret", "jackal-package-init", "-n", "jackal")
	require.NoError(t, err)
	_, _, err = e2e.Kubectl("create", "secret", "generic", "jackal-package-init", "-n", "jackal", fmt.Sprintf("--from-literal=data=%s", string(jsonPkg)))
	require.NoError(t, err)
}

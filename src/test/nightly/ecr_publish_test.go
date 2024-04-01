// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package test provides e2e tests for Jackal.
package test

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/racer159/jackal/src/config"
	"github.com/racer159/jackal/src/test"
	"github.com/stretchr/testify/require"
)

var (
	e2e test.JackalE2ETest //nolint:gochecknoglobals
)

// TestECRPublishing tests pushing, inspecting, and pulling signed Jackal packages to/from ECR.
/*
This test assumes the following:
  1. The user running the test has a current valid credential to the public.ecr.aws/t8y5r5z5/jackal-nightly repository in their docker config.json
  2. The jackal CLI has been built and is available in the build directory
*/
func TestECRPublishing(t *testing.T) {
	t.Log("E2E: Testing component actions")

	// Work from the root directory of the project
	os.Chdir("../../../")

	// Create a tmpDir for us to use during this test
	tmpDir := t.TempDir()

	// Set up the e2e configs
	e2e.Arch = config.GetArch()
	e2e.JackalBinPath = path.Join("build", test.GetCLIName())
	e2e.ApplianceMode = true
	e2e.RunClusterTests = false

	// Set up variables for common names/locations
	testPackageName := "helm-charts"
	testPackageVersion := "0.0.1"
	testPackageFileName := fmt.Sprintf("jackal-package-%s-%s-%s.tar.zst", testPackageName, e2e.Arch, testPackageVersion)
	testPackageLocation := filepath.Join(tmpDir, testPackageFileName)
	registryURL := "oci://public.ecr.aws/t8y5r5z5/jackal-nightly"
	upstreamPackageURL := fmt.Sprintf("%s/%s:%s-%s", registryURL, testPackageName, testPackageVersion, e2e.Arch)
	keyFlag := fmt.Sprintf("--key=%s", "./src/test/packages/jackal-test.pub")

	// Build the package with our test signature
	stdOut, stdErr, err := e2e.Jackal("package", "create", "examples/helm-charts", "--signing-key=./src/test/packages/jackal-test.prv-key", "--confirm", fmt.Sprintf("-o=%s", tmpDir))
	require.NoError(t, err, stdOut, stdErr)
	require.FileExists(t, testPackageLocation)

	// Validate that we can publish the package to ECR without an issue
	stdOut, stdErr, err = e2e.Jackal("package", "publish", testPackageLocation, registryURL, keyFlag)
	require.NoError(t, err, stdOut, stdErr)

	// Ensure we get a warning when trying to inspect the online published package
	stdOut, stdErr, err = e2e.Jackal("package", "inspect", upstreamPackageURL, keyFlag, "--sbom-out", tmpDir)
	require.NoError(t, err, stdOut, stdErr)
	require.Contains(t, stdErr, "Validating SBOM checksums")
	require.Contains(t, stdErr, "Package signature validated!")

	// Validate that we can pull the package down from ECR
	stdOut, stdErr, err = e2e.Jackal("package", "pull", upstreamPackageURL)
	require.NoError(t, err, stdOut, stdErr)
	defer e2e.CleanFiles(testPackageFileName)

	// Ensure we get a warning when trying to inspect the package without providing the public key
	stdOut, stdErr, err = e2e.Jackal("package", "inspect", testPackageFileName)
	require.NoError(t, err, stdOut, stdErr)
	require.NotContains(t, stdErr, "Validating SBOM checksums")
	require.Contains(t, stdErr, "The package was signed but no public key was provided, skipping signature validation")

	// Validate that we get no warnings when inspecting the package while providing the public key
	stdOut, stdErr, err = e2e.Jackal("package", "inspect", testPackageFileName, keyFlag)
	require.NoError(t, err, stdOut, stdErr)
	require.NotContains(t, stdErr, "Validating SBOM checksums")
	require.Contains(t, stdErr, "Package signature validated!")
}

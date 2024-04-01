// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package test provides e2e tests for Jackal.
package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChecksumAndSignature(t *testing.T) {
	t.Log("E2E: Checksum and Signature")
	e2e.SetupWithCluster(t)

	testPackageDirPath := "examples/dos-games"
	pkgName := fmt.Sprintf("jackal-package-dos-games-%s-1.0.0.tar.zst", e2e.Arch)
	privateKeyFlag := "--signing-key=src/test/packages/jackal-test.prv-key"
	publicKeyFlag := "--key=src/test/packages/jackal-test.pub"

	stdOut, stdErr, err := e2e.Jackal("package", "create", testPackageDirPath, privateKeyFlag, "--confirm")
	require.NoError(t, err, stdOut, stdErr)
	defer e2e.CleanFiles(pkgName)

	/* Test operations during package inspect */
	// Test that we can inspect the yaml of the package without the private key
	stdOut, stdErr, err = e2e.Jackal("package", "inspect", pkgName)
	require.NoError(t, err, stdOut, stdErr)

	// Test that we don't get an error when we remember to provide the public key
	stdOut, stdErr, err = e2e.Jackal("package", "inspect", pkgName, publicKeyFlag)
	require.NoError(t, err, stdOut, stdErr)
	require.Contains(t, stdErr, "Verified OK")

	/* Test operations during package deploy */
	// Test that we get an error when trying to deploy a package without providing the public key
	stdOut, stdErr, err = e2e.Jackal("package", "deploy", pkgName, "--confirm")
	require.Error(t, err, stdOut, stdErr)
	require.Contains(t, stdErr, "Failed to deploy package: unable to load the package: package is signed but no key was provided")

	// Test that we don't get an error when we remember to provide the public key
	stdOut, stdErr, err = e2e.Jackal("package", "deploy", pkgName, publicKeyFlag, "--confirm")
	require.NoError(t, err, stdOut, stdErr)
	require.Contains(t, stdErr, "Jackal deployment complete")

	// Remove the package
	stdOut, stdErr, err = e2e.Jackal("package", "remove", pkgName, publicKeyFlag, "--confirm")
	require.NoError(t, err, stdOut, stdErr)
}

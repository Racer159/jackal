// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package test provides e2e tests for Jackal.
package test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomInit(t *testing.T) {
	t.Log("E2E: Custom Init Package")
	e2e.SetupWithCluster(t)
	buildPath := filepath.Join("src", "test", "packages", "35-custom-init-package")
	pkgName := fmt.Sprintf("jackal-init-%s-%s.tar.zst", e2e.Arch, e2e.GetJackalVersion(t))
	privateKeyFlag := "--signing-key=src/test/packages/jackal-test.prv-key"
	publicKeyFlag := "--key=src/test/packages/jackal-test.pub"

	stdOut, stdErr, err := e2e.Jackal("package", "create", buildPath, privateKeyFlag, "--confirm")
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
	stdOut, stdErr, err = e2e.Jackal("init", "--confirm")
	require.Error(t, err, stdOut, stdErr)
	require.Contains(t, stdErr, "unable to load the package: package is signed but no key was provided - add a key with the --key flag or use the --insecure flag and run the command again")

	/* Test operations during package deploy */
	// Test that we can deploy the package with the public key
	stdOut, stdErr, err = e2e.Jackal("init", "--confirm", publicKeyFlag)
	require.NoError(t, err, stdOut, stdErr)
}

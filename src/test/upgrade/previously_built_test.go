// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package upgrade provides a test for the upgrade flow.
package upgrade

import (
	"context"
	"path"
	"testing"

	"github.com/racer159/jackal/src/pkg/utils/exec"
	test "github.com/racer159/jackal/src/test"
	"github.com/stretchr/testify/require"
)

func kubectl(args ...string) (string, string, error) {
	tk := []string{"tools", "kubectl"}
	args = append(tk, args...)
	return jackal(args...)
}

func jackal(args ...string) (string, string, error) {
	jackalBinPath := path.Join("../../../build", test.GetCLIName())
	return exec.CmdWithContext(context.TODO(), exec.PrintCfg(), jackalBinPath, args...)
}

func TestPreviouslyBuiltJackalPackage(t *testing.T) {
	// This test tests that a package built with the previous version of jackal will still deploy with the newer version
	t.Log("Upgrade: Previously Built Jackal Package")

	// For the upgrade test, podinfo-upgrade should already be in the cluster (version 6.3.3) (see .github/workflows/test-upgrade.yml)
	kubectlOut, _, _ := kubectl("-n=podinfo-upgrade", "rollout", "status", "deployment/podinfo-upgrade")
	require.Contains(t, kubectlOut, "successfully rolled out")
	kubectlOut, _, _ = kubectl("-n=podinfo-upgrade", "get", "deployment", "podinfo-upgrade", "-o=jsonpath={.metadata.labels}}")
	require.Contains(t, kubectlOut, "6.3.3")

	// Verify that the private-registry secret and private-git-server secret in the podinfo-upgrade namespace are the same after re-init
	// This tests that `jackal tools update-creds` successfully updated the other namespace
	jackalRegistrySecret, _, _ := kubectl("-n=jackal", "get", "secret", "private-registry", "-o", "jsonpath={.data}")
	podinfoRegistrySecret, _, _ := kubectl("-n=podinfo-upgrade", "get", "secret", "private-registry", "-o", "jsonpath={.data}")
	require.Equal(t, jackalRegistrySecret, podinfoRegistrySecret, "the jackal registry secret and podinfo-upgrade registry secret did not match")
	jackalGitServerSecret, _, _ := kubectl("-n=jackal", "get", "secret", "private-git-server", "-o", "jsonpath={.data}")
	podinfoGitServerSecret, _, _ := kubectl("-n=podinfo-upgrade", "get", "secret", "private-git-server", "-o", "jsonpath={.data}")
	require.Equal(t, jackalGitServerSecret, podinfoGitServerSecret, "the jackal git server secret and podinfo-upgrade git server secret did not match")

	// We also expect a 6.3.4 package to have been previously built
	previouslyBuiltPackage := "../../../jackal-package-test-upgrade-package-amd64-6.3.4.tar.zst"

	// Deploy the package.
	jackalDeployArgs := []string{"package", "deploy", previouslyBuiltPackage, "--confirm"}
	stdOut, stdErr, err := jackal(jackalDeployArgs...)
	require.NoError(t, err, stdOut, stdErr)

	// [DEPRECATIONS] We expect any deprecated things to work from the old package
	require.Contains(t, stdErr, "Successfully deployed podinfo 6.3.4")
	require.Contains(t, stdErr, "-----BEGIN PUBLIC KEY-----")

	// Verify that podinfo-upgrade successfully deploys in the cluster (version 6.3.4)
	kubectlOut, _, _ = kubectl("-n=podinfo-upgrade", "rollout", "status", "deployment/podinfo-upgrade")
	require.Contains(t, kubectlOut, "successfully rolled out")
	kubectlOut, _, _ = kubectl("-n=podinfo-upgrade", "get", "deployment", "podinfo-upgrade", "-o=jsonpath={.metadata.labels}}")
	require.Contains(t, kubectlOut, "6.3.4")

	// We also want to build a new package.
	stdOut, stdErr, err = jackal("package", "create", "../../../src/test/upgrade", "--set", "PODINFO_VERSION=6.3.5", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
	newlyBuiltPackage := "jackal-package-test-upgrade-package-amd64-6.3.5.tar.zst"

	// Deploy the package.
	stdOut, stdErr, err = jackal("package", "deploy", newlyBuiltPackage, "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	// [DEPRECATIONS] We expect any deprecated things to work from the new package
	require.Contains(t, stdErr, "Successfully deployed podinfo 6.3.5")
	require.Contains(t, stdErr, "-----BEGIN PUBLIC KEY-----")

	// Verify that podinfo-upgrade successfully deploys in the cluster (version 6.3.5)
	kubectlOut, _, _ = kubectl("-n=podinfo-upgrade", "rollout", "status", "deployment/podinfo-upgrade")
	require.Contains(t, kubectlOut, "successfully rolled out")
	kubectlOut, _, _ = kubectl("-n=podinfo-upgrade", "get", "deployment", "podinfo-upgrade", "-o=jsonpath={.metadata.labels}}")
	require.Contains(t, kubectlOut, "6.3.5")

	// Remove the package.
	stdOut, stdErr, err = jackal("package", "remove", "test-upgrade-package", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
}

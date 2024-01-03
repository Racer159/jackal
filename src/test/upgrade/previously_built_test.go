// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package upgrade provides a test for the upgrade flow.
package upgrade

import (
	"context"
	"fmt"
	"path"
	"testing"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/utils/exec"
	test "github.com/defenseunicorns/zarf/src/test"
	"github.com/stretchr/testify/require"
)

func kubectl(args ...string) (string, string, error) {
	tk := []string{"tools", "kubectl"}
	args = append(tk, args...)
	return zarf(args...)
}

func zarf(args ...string) (string, string, error) {
	zarfBinPath := path.Join("../../../build", test.GetCLIName())
	return exec.CmdWithContext(context.TODO(), exec.PrintCfg(), zarfBinPath, args...)
}

func TestPreviouslyBuiltZarfPackage(t *testing.T) {
	// This test tests that a package built with the previous version of zarf will still deploy with the newer version
	t.Log("Upgrade: Previously Built Zarf Package")

	// For the upgrade test, podinfo-upgrade should already be in the cluster (version 6.3.3) (see .github/workflows/test-upgrade.yml)
	kubectlOut, _, _ := kubectl("-n=podinfo-upgrade", "rollout", "status", "deployment/podinfo-upgrade")
	require.Contains(t, kubectlOut, "successfully rolled out")
	kubectlOut, _, _ = kubectl("-n=podinfo-upgrade", "get", "deployment", "podinfo-upgrade", "-o=jsonpath={.metadata.labels}}")
	require.Contains(t, kubectlOut, "6.3.3")

	// Verify that the private-registry secret and private-git-server secret in the podinfo-upgrade namespace are the same after re-init
	// This tests that `zarf tools update-creds` successfully updated the other namespace
	zarfRegistrySecret, _, _ := kubectl("-n=zarf", "get", "secret", "private-registry", "-o", "jsonpath={.data}")
	podinfoRegistrySecret, _, _ := kubectl("-n=podinfo-upgrade", "get", "secret", "private-registry", "-o", "jsonpath={.data}")
	require.Equal(t, zarfRegistrySecret, podinfoRegistrySecret, "the zarf registry secret and podinfo-upgrade registry secret did not match")
	zarfGitServerSecret, _, _ := kubectl("-n=zarf", "get", "secret", "private-git-server", "-o", "jsonpath={.data}")
	podinfoGitServerSecret, _, _ := kubectl("-n=podinfo-upgrade", "get", "secret", "private-git-server", "-o", "jsonpath={.data}")
	require.Equal(t, zarfGitServerSecret, podinfoGitServerSecret, "the zarf git server secret and podinfo-upgrade git server secret did not match")

	// We also expect a 6.3.4 package to have been previously built
	previouslyBuiltPackage := fmt.Sprintf("../../../zarf-package-test-upgrade-package-%s-6.3.4.tar.zst", config.GetArch())

	// Deploy the package.
	zarfDeployArgs := []string{"package", "deploy", previouslyBuiltPackage, "--confirm"}
	stdOut, stdErr, err := zarf(zarfDeployArgs...)
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
	stdOut, stdErr, err = zarf("package", "create", "../../../src/test/upgrade", "--set", "PODINFO_VERSION=6.3.5", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
	newlyBuiltPackage := fmt.Sprintf("zarf-package-test-upgrade-package-%s-6.3.5.tar.zst", config.GetArch())

	// Deploy the package.
	stdOut, stdErr, err = zarf("package", "deploy", newlyBuiltPackage, "--confirm")
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
	stdOut, stdErr, err = zarf("package", "remove", "test-upgrade-package", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
}

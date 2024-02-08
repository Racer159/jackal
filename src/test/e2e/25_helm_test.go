// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package test provides e2e tests for Zarf.
package test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var helmChartsPkg string

func TestHelm(t *testing.T) {
	t.Log("E2E: Helm chart")
	e2e.SetupWithCluster(t)

	helmChartsPkg = filepath.Join("build", fmt.Sprintf("zarf-package-helm-charts-%s-0.0.1.tar.zst", e2e.Arch))

	testHelmUninstallRollback(t)

	testHelmAdoption(t)

	t.Run("helm charts example", testHelmChartsExample)

	t.Run("helm escaping", testHelmEscaping)
}

func testHelmChartsExample(t *testing.T) {
	t.Parallel()
	t.Log("E2E: Helm chart example")
	tmpdir := t.TempDir()

	// Create a package that has a tarball as a local chart
	localTgzChartPath := filepath.Join("src", "test", "packages", "25-local-tgz-chart")
	stdOut, stdErr, err := e2e.Zarf("package", "create", localTgzChartPath, "--tmpdir", tmpdir, "--confirm")
	require.NoError(t, err, stdOut, stdErr)
	defer e2e.CleanFiles(fmt.Sprintf("zarf-package-helm-charts-local-tgz-%s-0.0.1.tar.zst", e2e.Arch))

	// Create a package that needs dependencies
	evilChartDepsPath := filepath.Join("src", "test", "packages", "25-evil-chart-deps")
	stdOut, stdErr, err = e2e.Zarf("package", "create", evilChartDepsPath, "--tmpdir", tmpdir, "--confirm")
	require.Error(t, err, stdOut, stdErr)
	require.Contains(t, e2e.StripMessageFormatting(stdErr), "could not download https://charts.jetstack.io/charts/cert-manager-v1.11.1.tgz")
	require.FileExists(t, filepath.Join(evilChartDepsPath, "good-chart", "charts", "gitlab-runner-0.55.0.tgz"))

	// Create a package with a chart name that doesn't exist in a repo
	evilChartLookupPath := filepath.Join("src", "test", "packages", "25-evil-chart-lookup")
	stdOut, stdErr, err = e2e.Zarf("package", "create", evilChartLookupPath, "--tmpdir", tmpdir, "--confirm")
	require.Error(t, err, stdOut, stdErr)
	require.Contains(t, e2e.StripMessageFormatting(stdErr), "chart \"asdf\" version \"6.4.0\" not found")
	require.Contains(t, e2e.StripMessageFormatting(stdErr), "Available charts and versions from \"https://stefanprodan.github.io/podinfo\":")

	// Create the package (with a registry override to test that as well)
	stdOut, stdErr, err = e2e.Zarf("package", "create", "examples/helm-charts", "-o", "build", "--registry-override", "ghcr.io=docker.io", "--tmpdir", tmpdir, "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	// Deploy the example package.
	stdOut, stdErr, err = e2e.Zarf("package", "deploy", helmChartsPkg, "--confirm")
	require.NoError(t, err, stdOut, stdErr)
	require.Contains(t, string(stdErr), "registryOverrides", "registry overrides was not saved to build data")
	require.Contains(t, string(stdErr), "docker.io", "docker.io not found in registry overrides")

	// Remove the example package.
	stdOut, stdErr, err = e2e.Zarf("package", "remove", "helm-charts", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
}

func testHelmEscaping(t *testing.T) {
	t.Parallel()
	t.Log("E2E: Helm chart escaping")

	// Create the package.
	stdOut, stdErr, err := e2e.Zarf("package", "create", "src/test/packages/25-evil-templates/", "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	path := fmt.Sprintf("zarf-package-evil-templates-%s.tar.zst", e2e.Arch)

	// Deploy the package.
	stdOut, stdErr, err = e2e.Zarf("package", "deploy", path, "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	// Verify the configmap was deployed, escaped, and contains all of its data
	kubectlOut, _ := exec.Command("kubectl", "describe", "cm", "dont-template-me").Output()
	require.Contains(t, string(kubectlOut), `alert: OOMKilled {{ "{{ \"random.Values\" }}" }}`)
	require.Contains(t, string(kubectlOut), "backtick1: \"content with backticks `some random things`\"")
	require.Contains(t, string(kubectlOut), "backtick2: \"nested templating with backticks {{` random.Values `}}\"")
	require.Contains(t, string(kubectlOut), `description: Pod {{$labels.pod}} in {{$labels.namespace}} got OOMKilled`)
	require.Contains(t, string(kubectlOut), `TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIG`)

	// Remove the package.
	stdOut, stdErr, err = e2e.Zarf("package", "remove", "evil-templates", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
}

func testHelmUninstallRollback(t *testing.T) {
	t.Log("E2E: Helm Uninstall and Rollback")

	goodPath := fmt.Sprintf("build/zarf-package-dos-games-%s-1.0.0.tar.zst", e2e.Arch)
	evilPath := fmt.Sprintf("zarf-package-dos-games-%s.tar.zst", e2e.Arch)

	// Create the evil package (with the bad service).
	stdOut, stdErr, err := e2e.Zarf("package", "create", "src/test/packages/25-evil-dos-games/", "--skip-sbom", "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	// Deploy the evil package.
	stdOut, stdErr, err = e2e.Zarf("package", "deploy", evilPath, "--timeout", "10s", "--confirm")
	require.Error(t, err, stdOut, stdErr)

	// This package contains SBOMable things but was created with --skip-sbom
	require.Contains(t, string(stdErr), "This package does NOT contain an SBOM.")

	// Ensure that this does not leave behind a dos-games chart
	helmOut, err := exec.Command("helm", "list", "-n", "dos-games").Output()
	require.NoError(t, err)
	require.NotContains(t, string(helmOut), "zarf-f53a99d4a4dd9a3575bedf59cd42d48d751ae866")

	// Deploy the good package.
	stdOut, stdErr, err = e2e.Zarf("package", "deploy", goodPath, "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	// Ensure that this does create a dos-games chart
	helmOut, err = exec.Command("helm", "list", "-n", "dos-games").Output()
	require.NoError(t, err)
	require.Contains(t, string(helmOut), "zarf-f53a99d4a4dd9a3575bedf59cd42d48d751ae866")

	// Deploy the evil package.
	stdOut, stdErr, err = e2e.Zarf("package", "deploy", evilPath, "--timeout", "10s", "--confirm")
	require.Error(t, err, stdOut, stdErr)

	// Ensure that we rollback properly
	helmOut, err = exec.Command("helm", "history", "-n", "dos-games", "zarf-f53a99d4a4dd9a3575bedf59cd42d48d751ae866", "--max", "1").Output()
	require.NoError(t, err)
	require.Contains(t, string(helmOut), "Rollback to 1")

	// Deploy the evil package (again to ensure we check full history)
	stdOut, stdErr, err = e2e.Zarf("package", "deploy", evilPath, "--timeout", "10s", "--confirm")
	require.Error(t, err, stdOut, stdErr)

	// Ensure that we rollback properly
	helmOut, err = exec.Command("helm", "history", "-n", "dos-games", "zarf-f53a99d4a4dd9a3575bedf59cd42d48d751ae866", "--max", "1").Output()
	require.NoError(t, err)
	require.Contains(t, string(helmOut), "Rollback to 5")

	// Remove the package.
	stdOut, stdErr, err = e2e.Zarf("package", "remove", "dos-games", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
}

func testHelmAdoption(t *testing.T) {
	t.Log("E2E: Helm Adopt a Deployment")

	packagePath := fmt.Sprintf("build/zarf-package-dos-games-%s-1.0.0.tar.zst", e2e.Arch)
	deploymentManifest := "src/test/packages/25-manifest-adoption/deployment.yaml"

	// Deploy dos-games manually into the cluster without Zarf
	kubectlOut, _, _ := e2e.Kubectl("apply", "-f", deploymentManifest)
	require.Contains(t, string(kubectlOut), "deployment.apps/game created")

	// Deploy dos-games into the cluster with Zarf
	stdOut, stdErr, err := e2e.Zarf("package", "deploy", packagePath, "--confirm", "--adopt-existing-resources")
	require.NoError(t, err, stdOut, stdErr)

	// Ensure that this does create a dos-games chart
	helmOut, err := exec.Command("helm", "list", "-n", "dos-games").Output()
	require.NoError(t, err)
	require.Contains(t, string(helmOut), "zarf-f53a99d4a4dd9a3575bedf59cd42d48d751ae866")

	// Remove the package.
	stdOut, stdErr, err = e2e.Zarf("package", "remove", "dos-games", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
}

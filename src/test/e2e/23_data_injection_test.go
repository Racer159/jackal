// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package test provides e2e tests for Jackal.
package test

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/Racer159/jackal/src/pkg/cluster"
	"github.com/Racer159/jackal/src/pkg/utils/exec"
	"github.com/stretchr/testify/require"
)

func TestDataInjection(t *testing.T) {
	t.Log("E2E: Data injection")
	e2e.SetupWithCluster(t)

	path := fmt.Sprintf("build/jackal-package-kiwix-%s-3.5.0.tar", e2e.Arch)

	tmpdir := t.TempDir()
	sbomPath := filepath.Join(tmpdir, ".sbom-location")

	// Repeat the injection action 3 times to ensure the data injection is idempotent and doesn't fail to perform an upgrade
	for i := 0; i < 3; i++ {
		runDataInjection(t, path)
	}

	// Verify the file and injection marker were created
	runningKiwixPod, _, err := e2e.Kubectl("--namespace=kiwix", "get", "pods", "--selector=app=kiwix-serve", "--field-selector=status.phase=Running", "--output=jsonpath={.items[0].metadata.name}")
	require.NoError(t, err)
	stdOut, stdErr, err := e2e.Kubectl("--namespace=kiwix", "logs", runningKiwixPod, "--tail=5", "-c=kiwix-serve")
	require.NoError(t, err, stdOut, stdErr)
	require.Contains(t, stdOut, "devops.stackexchange.com_en_all_2023-05.zim")
	require.Contains(t, stdOut, ".jackal-injection-")

	// need target to equal svc that we are trying to connect to call checkForJackalConnectLabel
	c, err := cluster.NewCluster()
	require.NoError(t, err)
	tunnel, err := c.Connect("kiwix")
	require.NoError(t, err)
	defer tunnel.Close()

	// Ensure connection
	resp, err := http.Get(tunnel.HTTPEndpoint())
	require.NoError(t, err, resp)
	require.Equal(t, 200, resp.StatusCode)

	// Remove the data injection example
	stdOut, stdErr, err = e2e.Jackal("package", "remove", path, "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	// Ensure that the `requirements.txt` file is discovered correctly
	stdOut, stdErr, err = e2e.Jackal("package", "inspect", path, "--sbom-out", sbomPath)
	require.NoError(t, err, stdOut, stdErr)
	require.FileExists(t, filepath.Join(sbomPath, "kiwix", "compare.html"), "A compare.html file should have been made")

	require.FileExists(t, filepath.Join(sbomPath, "kiwix", "sbom-viewer-jackal-component-kiwix-serve.html"), "The data-injection component should have an SBOM viewer")
	require.FileExists(t, filepath.Join(sbomPath, "kiwix", "jackal-component-kiwix-serve.json"), "The data-injection component should have an SBOM json")
}

func runDataInjection(t *testing.T, path string) {
	// Limit this deploy to 5 minutes
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Minute)
	defer cancel()

	// Deploy the data injection example
	stdOut, stdErr, err := exec.CmdWithContext(ctx, exec.PrintCfg(), e2e.JackalBinPath, "package", "deploy", path, "-l", "trace", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
}

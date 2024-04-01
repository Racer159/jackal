// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package test provides e2e tests for Jackal.
package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateSBOM(t *testing.T) {
	tmpdir := t.TempDir()
	sbomPath := filepath.Join(tmpdir, ".sbom-location")

	pkgName := fmt.Sprintf("jackal-package-dos-games-%s-1.0.0.tar.zst", e2e.Arch)

	stdOut, stdErr, err := e2e.Jackal("package", "create", "examples/dos-games", "--sbom-out", sbomPath, "--confirm")
	require.NoError(t, err, stdOut, stdErr)
	require.Contains(t, stdErr, "Creating SBOMs for 1 images and 0 components with files.")
	// Test that the game package generates the SBOMs we expect (images only)
	require.FileExists(t, filepath.Join(sbomPath, "dos-games", "sbom-viewer-docker.io_defenseunicorns_zarf-game_multi-tile-dark.html"))
	require.FileExists(t, filepath.Join(sbomPath, "dos-games", "compare.html"))
	require.FileExists(t, filepath.Join(sbomPath, "dos-games", "docker.io_defenseunicorns_zarf-game_multi-tile-dark.json"))

	// Clean the SBOM path so it is force to be recreated
	e2e.CleanFiles(sbomPath)

	stdOut, stdErr, err = e2e.Jackal("package", "inspect", pkgName, "--sbom-out", sbomPath)
	require.NoError(t, err, stdOut, stdErr)
	// Test that the game package generates the SBOMs we expect (images only)
	_, err = os.ReadFile(filepath.Join(sbomPath, "dos-games", "sbom-viewer-docker.io_defenseunicorns_zarf-game_multi-tile-dark.html"))
	require.NoError(t, err)
	_, err = os.ReadFile(filepath.Join(sbomPath, "dos-games", "compare.html"))
	require.NoError(t, err)
	_, err = os.ReadFile(filepath.Join(sbomPath, "dos-games", "docker.io_defenseunicorns_zarf-game_multi-tile-dark.json"))
	require.NoError(t, err)

	// Pull the current jackal binary version to find the corresponding init package
	version, stdErr, err := e2e.Jackal("version")
	require.NoError(t, err, version, stdErr)

	initName := fmt.Sprintf("build/jackal-init-%s-%s.tar.zst", e2e.Arch, strings.TrimSpace(version))

	stdOut, stdErr, err = e2e.Jackal("package", "inspect", initName, "--sbom-out", sbomPath)
	require.NoError(t, err, stdOut, stdErr)
	// Test that we preserve the filepath
	_, err = os.ReadFile(filepath.Join(sbomPath, "dos-games", "sbom-viewer-docker.io_defenseunicorns_zarf-game_multi-tile-dark.html"))
	require.NoError(t, err)
	// Test that the init package generates the SBOMs we expect (images + component files)
	_, err = os.ReadFile(filepath.Join(sbomPath, "init", "sbom-viewer-docker.io_gitea_gitea_1.21.5-rootless.html"))
	require.NoError(t, err)
	_, err = os.ReadFile(filepath.Join(sbomPath, "init", "docker.io_gitea_gitea_1.21.5-rootless.json"))
	require.NoError(t, err)
	_, err = os.ReadFile(filepath.Join(sbomPath, "init", "sbom-viewer-jackal-component-k3s.html"))
	require.NoError(t, err)
	_, err = os.ReadFile(filepath.Join(sbomPath, "init", "jackal-component-k3s.json"))
	require.NoError(t, err)
	_, err = os.ReadFile(filepath.Join(sbomPath, "init", "compare.html"))
	require.NoError(t, err)

	e2e.CleanFiles(pkgName)
}

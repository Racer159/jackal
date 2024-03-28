// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package test provides e2e tests for Zarf.
package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateIndexSha(t *testing.T) {
	t.Log("E2E: Create Templating")

	// run `zarf package create` with a specified image cache location
	tmpdir := t.TempDir()
	decompressPath := filepath.Join(tmpdir, ".package-decompressed")

	pkgName := fmt.Sprintf("zarf-package-index-sha-%s.tar.zst", e2e.Arch)
	_, _, err := e2e.Zarf("package", "create", "src/test/packages/14-index-sha", "--confirm")
	require.NoError(t, err)

	_, _, err = e2e.Zarf("t", "archiver", "decompress", pkgName, decompressPath)
	require.NoError(t, err)

	builtConfig, err := os.ReadFile(decompressPath + "/zarf.yaml")
	require.NoError(t, err)
	require.Contains(t, string(builtConfig), "docker.io/defenseunicorns/zarf-game:multi-tile-dark@sha256:f78e442f0f3eb3e9459b5ae6b1a8fda62f8dfe818112e7d130a4e8ae72b3cbff")

}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package test provides e2e tests for Jackal.
package test

import (
	"path/filepath"
	"testing"

	"github.com/Racer159/jackal/src/pkg/layout"
	"github.com/Racer159/jackal/src/pkg/utils"
	"github.com/Racer159/jackal/src/types"
	"github.com/stretchr/testify/require"
)

func TestJackalDevGenerate(t *testing.T) {
	t.Log("E2E: Jackal Dev Generate")

	t.Run("Test generate podinfo", func(t *testing.T) {
		tmpDir := t.TempDir()

		url := "https://github.com/stefanprodan/podinfo.git"
		version := "6.4.0"
		gitPath := "charts/podinfo"

		stdOut, stdErr, err := e2e.Jackal("dev", "generate", "podinfo", "--url", url, "--version", version, "--gitPath", gitPath, "--output-directory", tmpDir)
		require.NoError(t, err, stdOut, stdErr)

		jackalPackage := types.JackalPackage{}
		packageLocation := filepath.Join(tmpDir, layout.JackalYAML)
		err = utils.ReadYaml(packageLocation, &jackalPackage)
		require.NoError(t, err)
		require.Equal(t, jackalPackage.Components[0].Charts[0].URL, url)
		require.Equal(t, jackalPackage.Components[0].Charts[0].Version, version)
		require.Equal(t, jackalPackage.Components[0].Charts[0].GitPath, gitPath)
		require.NotEmpty(t, jackalPackage.Components[0].Images)
	})
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package test provides e2e tests for Jackal.
package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Racer159/jackal/src/pkg/utils/exec"
	"github.com/stretchr/testify/require"
)

func TestGHCRDeploy(t *testing.T) {
	t.Log("E2E: GHCR OCI deploy")
	e2e.SetupWithCluster(t)

	var sha string
	// shas for package published 2023-08-08T22:13:51Z
	switch e2e.Arch {
	case "arm64":
		sha = "ac7d7684ca9b4edb061a7732aefc17cfb7b7c983fec23e1fe319cf535618a8b6"
	case "amd64":
		sha = "aca4d4cf24532d69a8941a446067fc3d8474581507236b37bb7188836d93bf89"
	}

	// Test with command from https://jackal.dev/install/
	stdOut, stdErr, err := e2e.Jackal("package", "deploy", fmt.Sprintf("oci://🦄/dos-games:1.0.0-%s@sha256:%s", e2e.Arch, sha), "--key=https://jackal.dev/cosign.pub", "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	stdOut, stdErr, err = e2e.Jackal("package", "remove", "dos-games", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
}

func TestCosignDeploy(t *testing.T) {
	t.Log("E2E: Cosign deploy")
	e2e.SetupWithCluster(t)

	// Test with command from https://jackal.dev/install/
	command := fmt.Sprintf("%s package deploy sget://Racer159/jackal-hello-world:$(uname -m) --confirm", e2e.JackalBinPath)

	stdOut, stdErr, err := exec.CmdWithContext(context.TODO(), exec.PrintCfg(), "sh", "-c", command)
	require.NoError(t, err, stdOut, stdErr)

	stdOut, stdErr, err = e2e.Jackal("package", "remove", "dos-games", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
}

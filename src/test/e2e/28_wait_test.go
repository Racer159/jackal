// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package test provides e2e tests for Jackal.
package test

import (
	"fmt"
	"time"

	"testing"

	"github.com/racer159/jackal/src/test"
	"github.com/stretchr/testify/require"
)

type jackalCommandResult struct {
	stdOut string
	stdErr string
	err    error
}

func jackalCommandWStruct(e2e test.JackalE2ETest, path string) (result jackalCommandResult) {
	result.stdOut, result.stdErr, result.err = e2e.Jackal("package", "deploy", path, "--confirm")
	return result
}

func TestNoWait(t *testing.T) {
	t.Log("E2E: Helm Wait")
	e2e.SetupWithCluster(t)

	stdOut, stdErr, err := e2e.Jackal("package", "create", "src/test/packages/28-helm-no-wait", "-o=build", "--confirm")
	require.NoError(t, err, stdOut, stdErr)

	path := fmt.Sprintf("build/jackal-package-helm-no-wait-%s.tar.zst", e2e.Arch)

	jackalChannel := make(chan jackalCommandResult, 1)
	go func() {
		jackalChannel <- jackalCommandWStruct(e2e, path)
	}()

	stdOut = ""
	stdErr = ""
	err = nil

	select {
	case res := <-jackalChannel:
		stdOut = res.stdOut
		stdErr = res.stdErr
		err = res.err
	case <-time.After(30 * time.Second):
		t.Error("Timeout waiting for jackal deploy (it tried to wait)")
		t.Log("Removing hanging namespace...")
		_, _, _ = e2e.Kubectl("delete", "namespace", "no-wait", "--force=true", "--wait=false", "--grace-period=0")
	}
	require.NoError(t, err, stdOut, stdErr)

	stdOut, stdErr, err = e2e.Jackal("package", "remove", "helm-no-wait", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
}

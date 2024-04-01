// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package test provides e2e tests for Jackal.
package test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/defenseunicorns/pkg/helpers"
	"github.com/stretchr/testify/require"
)

func TestUseCLI(t *testing.T) {
	t.Log("E2E: Use CLI")

	t.Run("jackal prepare sha256sum <local>", func(t *testing.T) {
		t.Parallel()

		// Test `jackal prepare sha256sum` for a local asset
		expectedShasum := "61b50898f982d015ed87093ba822de0fe011cec6dd67db39f99d8c56391a6109\n"
		shasumTestFilePath := "shasum-test-file"

		e2e.CleanFiles(shasumTestFilePath)
		t.Cleanup(func() {
			e2e.CleanFiles(shasumTestFilePath)
		})

		err := os.WriteFile(shasumTestFilePath, []byte("random test data 🦄\n"), helpers.ReadWriteUser)
		require.NoError(t, err)

		stdOut, stdErr, err := e2e.Jackal("prepare", "sha256sum", shasumTestFilePath)
		require.NoError(t, err, stdOut, stdErr)
		require.Equal(t, expectedShasum, stdOut, "The expected SHASUM should equal the actual SHASUM")
	})

	t.Run("jackal prepare sha256sum <remote>", func(t *testing.T) {
		t.Parallel()
		// Test `jackal prepare sha256sum` for a remote asset
		expectedShasum := "c3cdea0573ba5a058ec090b5d2683bf398e8b1614c37ec81136ed03b78167617\n"

		stdOut, stdErr, err := e2e.Jackal("prepare", "sha256sum", "https://zarf-public.s3-us-gov-west-1.amazonaws.com/pipelines/zarf-prepare-shasum-remote-test-file.txt")
		require.NoError(t, err, stdOut, stdErr)
		require.Contains(t, stdOut, expectedShasum, "The expected SHASUM should equal the actual SHASUM")
	})

	t.Run("jackal version", func(t *testing.T) {
		t.Parallel()
		// Test `jackal version`
		version, _, err := e2e.Jackal("version")
		require.NoError(t, err)
		require.NotEqual(t, len(version), 0, "Jackal version should not be an empty string")
		version = strings.Trim(version, "\n")

		// test `jackal version --output=json`
		stdOut, _, err := e2e.Jackal("version", "--output=json")
		require.NoError(t, err)
		jsonVersion := fmt.Sprintf(",\"version\":\"%s\"}", version)
		require.Contains(t, stdOut, jsonVersion, "Jackal version should be the same in all formats")

		// test `jackal version --output=yaml`
		stdOut, _, err = e2e.Jackal("version", "--output=yaml")
		require.NoError(t, err)
		yamlVersion := fmt.Sprintf("version: %s", version)
		require.Contains(t, stdOut, yamlVersion, "Jackal version should be the same in all formats")
	})

	t.Run("jackal deploy should fail when given a bad component input", func(t *testing.T) {
		t.Parallel()
		// Test for expected failure when given a bad component input
		path := fmt.Sprintf("build/jackal-package-component-actions-%s.tar.zst", e2e.Arch)
		_, _, err := e2e.Jackal("package", "deploy", path, "--components=on-create,foo,logging", "--confirm")
		require.Error(t, err)
	})

	t.Run("jackal deploy should return a warning when no components are deployed", func(t *testing.T) {
		t.Parallel()
		_, _, err := e2e.Jackal("package", "create", "src/test/packages/00-no-components", "-o=build", "--confirm")
		require.NoError(t, err)
		path := fmt.Sprintf("build/jackal-package-no-components-%s.tar.zst", e2e.Arch)

		// Test that excluding all components with a leading dash results in a warning
		_, stdErr, err := e2e.Jackal("package", "deploy", path, "--components=-deselect-me", "--confirm")
		require.NoError(t, err)
		require.Contains(t, stdErr, "No components were selected for deployment")

		// Test that excluding still works even if a wildcard is given
		_, stdErr, err = e2e.Jackal("package", "deploy", path, "--components=*,-deselect-me", "--confirm")
		require.NoError(t, err)
		require.NotContains(t, stdErr, "DESELECT-ME COMPONENT")
	})

	t.Run("changing log level", func(t *testing.T) {
		t.Parallel()
		// Test that changing the log level actually applies the requested level
		_, stdErr, _ := e2e.Jackal("internal", "crc32", "jackal", "--log-level=debug")
		expectedOutString := "Log level set to debug"
		require.Contains(t, stdErr, expectedOutString, "The log level should be changed to 'debug'")
	})

	t.Run("bad jackal package deploy w/o --insecure or --shasum", func(t *testing.T) {
		t.Parallel()
		// Test that `jackal package deploy` gives an error if deploying a remote package without the --insecure or --shasum flags
		stdOut, stdErr, err := e2e.Jackal("package", "deploy", "https://jackal-examples.s3.amazonaws.com/jackal-package-appliance-demo-doom-20210125.tar.zst", "--confirm")
		require.Error(t, err, stdOut, stdErr)
	})

	t.Run("jackal package to test bad remote images", func(t *testing.T) {
		_, stdErr, err := e2e.Jackal("package", "create", "src/test/packages/00-remote-pull-fail", "--confirm")
		// expecting jackal to have an error and output to stderr
		require.Error(t, err)
		// Make sure we print the get request error (only look for GET since the actual error changes based on login status)
		require.Contains(t, stdErr, "failed to find the manifest on a remote: GET")
		// And the docker error
		require.Contains(t, stdErr, "response from daemon: No such image")
	})

	t.Run("jackal package to test archive path", func(t *testing.T) {
		t.Parallel()
		stdOut, stdErr, err := e2e.Jackal("package", "create", "packages/distros/eks", "--confirm")
		require.NoError(t, err, stdOut, stdErr)

		path := fmt.Sprintf("jackal-package-distro-eks-%s-0.0.3.tar.zst", e2e.Arch)
		stdOut, stdErr, err = e2e.Jackal("package", "deploy", path, "--confirm")
		require.NoError(t, err, stdOut, stdErr)

		require.FileExists(t, "binaries/eksctl_Darwin_x86_64")
		require.FileExists(t, "binaries/eksctl_Darwin_arm64")
		require.FileExists(t, "binaries/eksctl_Linux_x86_64")

		e2e.CleanFiles("binaries/eksctl_Darwin_x86_64", "binaries/eksctl_Darwin_arm64", "binaries/eksctl_Linux_x86_64", path, "eks.yaml")
	})

	t.Run("jackal package create with tmpdir and cache", func(t *testing.T) {
		t.Parallel()
		tmpdir := t.TempDir()
		cacheDir := filepath.Join(t.TempDir(), ".cache-location")
		stdOut, stdErr, err := e2e.Jackal("package", "create", "examples/dos-games", "--jackal-cache", cacheDir, "--tmpdir", tmpdir, "--log-level=debug", "-o=build", "--confirm")
		require.Contains(t, stdErr, tmpdir, "The other tmp path should show as being created")
		require.NoError(t, err, stdOut, stdErr)

		files, err := os.ReadDir(filepath.Join(cacheDir, "images"))
		require.NoError(t, err, "Encountered an unexpected error when reading image cache path")
		require.Greater(t, len(files), 1)
	})

	t.Run("jackal package inspect with tmpdir", func(t *testing.T) {
		t.Parallel()
		path := fmt.Sprintf("build/jackal-package-component-actions-%s.tar.zst", e2e.Arch)
		tmpdir := t.TempDir()
		stdOut, stdErr, err := e2e.Jackal("package", "inspect", path, "--tmpdir", tmpdir, "--log-level=debug")
		require.Contains(t, stdErr, tmpdir, "The other tmp path should show as being created")
		require.NoError(t, err, stdOut, stdErr)
	})

	t.Run("jackal package deploy with tmpdir", func(t *testing.T) {
		t.Parallel()
		tmpdir := t.TempDir()
		// run `jackal package deploy` with a specified tmp location
		var (
			firstFile  = "first-choice-file.txt"
			secondFile = "second-choice-file.txt"
		)
		t.Cleanup(func() {
			e2e.CleanFiles(firstFile, secondFile)
		})
		path := fmt.Sprintf("build/jackal-package-component-choice-%s.tar.zst", e2e.Arch)
		stdOut, stdErr, err := e2e.Jackal("package", "deploy", path, "--tmpdir", tmpdir, "--log-level=debug", "--confirm")
		require.Contains(t, stdErr, tmpdir, "The other tmp path should show as being created")
		require.NoError(t, err, stdOut, stdErr)
	})

	t.Run("remove cache", func(t *testing.T) {
		t.Parallel()
		tmpdir := t.TempDir()
		// Test removal of cache
		cachePath := filepath.Join(tmpdir, ".cache-location")
		stdOut, stdErr, err := e2e.Jackal("tools", "clear-cache", "--jackal-cache", cachePath)
		require.NoError(t, err, stdOut, stdErr)
		// Check that ReadDir returns no such file or directory for the cachePath
		_, err = os.ReadDir(cachePath)
		if runtime.GOOS == "windows" {
			msg := fmt.Sprintf("open %s: The system cannot find the file specified.", cachePath)
			require.EqualError(t, err, msg, "Did not receive expected error when reading a directory that should not exist")
		} else {
			msg := fmt.Sprintf("open %s: no such file or directory", cachePath)
			require.EqualError(t, err, msg, "Did not receive expected error when reading a directory that should not exist")
		}
	})

	t.Run("gen pki", func(t *testing.T) {
		t.Parallel()
		// Test generation of PKI
		tlsCA := "tls.ca"
		tlsCert := "tls.crt"
		tlsKey := "tls.key"
		t.Cleanup(func() {
			e2e.CleanFiles(tlsCA, tlsCert, tlsKey)
		})
		stdOut, stdErr, err := e2e.Jackal("tools", "gen-pki", "github.com", "--sub-alt-name", "google.com")
		require.NoError(t, err, stdOut, stdErr)

		require.FileExists(t, tlsCA)

		require.FileExists(t, tlsCert)

		require.FileExists(t, tlsKey)
	})

	t.Run("jackal tools yq should function appropriately across different uses", func(t *testing.T) {
		t.Parallel()

		file := "src/test/packages/00-yq-checks/file1.yaml"
		otherFile := "src/test/packages/00-yq-checks/file2.yaml"

		// Test that yq can eval properly
		_, stdErr, err := e2e.Jackal("tools", "yq", "eval", "-i", `.items[1].name = "renamed-item"`, file)
		require.NoError(t, err, stdErr)
		stdOut, stdErr, err := e2e.Jackal("tools", "yq", ".items[1].name", file)
		require.Contains(t, stdOut, "renamed-item")

		// Test that yq ea can be used properly
		_, stdErr, err = e2e.Jackal("tools", "yq", "eval-all", "-i", `. as $doc ireduce ({}; .items += $doc.items)`, file, otherFile)
		require.NoError(t, err, stdErr)
		stdOut, stdErr, err = e2e.Jackal("tools", "yq", "e", ".items | length", file)
		require.Equal(t, "4\n", stdOut)

	})
}

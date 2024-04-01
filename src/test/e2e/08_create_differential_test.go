// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package test provides e2e tests for Jackal.
package test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Racer159/jackal/src/config/lang"
	"github.com/Racer159/jackal/src/pkg/layout"
	"github.com/Racer159/jackal/src/pkg/utils"
	"github.com/Racer159/jackal/src/types"
	"github.com/mholt/archiver/v3"
	"github.com/stretchr/testify/require"
)

// TestCreateDifferential creates several differential packages and ensures the reference package images and repos are not included in the new package.
func TestCreateDifferential(t *testing.T) {
	t.Log("E2E: Test Differential Package Behavior")
	tmpdir := t.TempDir()

	packagePath := "src/test/packages/08-differential-package"
	packageName := fmt.Sprintf("jackal-package-differential-package-%s-v0.25.0.tar.zst", e2e.Arch)
	differentialPackageName := fmt.Sprintf("jackal-package-differential-package-%s-v0.25.0-differential-v0.26.0.tar.zst", e2e.Arch)
	differentialFlag := fmt.Sprintf("--differential=%s", packageName)

	// Build the package a first time
	stdOut, stdErr, err := e2e.Jackal("package", "create", packagePath, "--set=PACKAGE_VERSION=v0.25.0", "--confirm")
	require.NoError(t, err, stdOut, stdErr)
	defer e2e.CleanFiles(packageName)

	// Build the differential package without changing the version
	_, stdErr, err = e2e.Jackal("package", "create", packagePath, "--set=PACKAGE_VERSION=v0.25.0", differentialFlag, "--confirm")
	require.Error(t, err, "jackal package create should have errored when a differential package was being created without updating the package version number")
	require.Contains(t, e2e.StripMessageFormatting(stdErr), lang.PkgCreateErrDifferentialSameVersion)

	// Build the differential package
	stdOut, stdErr, err = e2e.Jackal("package", "create", packagePath, "--set=PACKAGE_VERSION=v0.26.0", differentialFlag, "--confirm")
	require.NoError(t, err, stdOut, stdErr)
	defer e2e.CleanFiles(differentialPackageName)

	// Extract the yaml of the differential package
	err = archiver.Extract(differentialPackageName, layout.JackalYAML, tmpdir)
	require.NoError(t, err, "unable to extract jackal.yaml from the differential git package")

	// Load the extracted jackal.yaml specification
	var differentialJackalConfig types.JackalPackage
	err = utils.ReadYaml(filepath.Join(tmpdir, layout.JackalYAML), &differentialJackalConfig)
	require.NoError(t, err, "unable to read jackal.yaml from the differential git package")

	// Get a list of all images and repos that are inside of the differential package
	actualGitRepos := []string{}
	actualImages := []string{}
	for _, component := range differentialJackalConfig.Components {
		actualGitRepos = append(actualGitRepos, component.Repos...)
		actualImages = append(actualImages, component.Images...)
	}

	/* Validate we have ONLY the git repos we expect to have */
	expectedGitRepos := []string{
		"https://github.com/stefanprodan/podinfo.git",
		"https://github.com/kelseyhightower/nocode.git",
		"https://github.com/Racer159/jackal.git@refs/tags/v0.26.0",
	}
	require.Len(t, actualGitRepos, 4, "jackal.yaml from the differential package does not contain the correct number of repos")
	for _, expectedRepo := range expectedGitRepos {
		require.Contains(t, actualGitRepos, expectedRepo, fmt.Sprintf("unable to find expected repo %s", expectedRepo))
	}

	/* Validate we have ONLY the images we expect to have */
	expectedImages := []string{
		"ghcr.io/stefanprodan/podinfo:latest",
		"ghcr.io/Racer159/jackal/agent:v0.26.0",
	}
	require.Len(t, actualImages, 2, "jackal.yaml from the differential package does not contain the correct number of images")
	for _, expectedImage := range expectedImages {
		require.Contains(t, actualImages, expectedImage, fmt.Sprintf("unable to find expected image %s", expectedImage))
	}
}

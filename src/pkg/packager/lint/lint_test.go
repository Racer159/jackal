// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package lint contains functions for verifying jackal yaml files are valid
package lint

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/defenseunicorns/jackal/src/config"
	"github.com/defenseunicorns/jackal/src/pkg/packager/composer"
	"github.com/defenseunicorns/jackal/src/types"
	goyaml "github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

const badJackalPackage = `
kind: JackalInitConfig
metadata:
  name: init
  description: Testing bad yaml

components:
- name: first-test-component
  import:
    not-path: packages/distros/k3s
- name: import-test
  import:
    path: 123123
`

const goodJackalPackage = `
x-name: &name good-jackal-package

kind: JackalPackageConfig
metadata:
  name: *name
  x-description: Testing good yaml with yaml extension

components:
  - name: baseline
    required: true
    x-foo: bar

`

func readAndUnmarshalYaml[T interface{}](t *testing.T, yamlString string) T {
	t.Helper()
	var unmarshalledYaml T
	err := goyaml.Unmarshal([]byte(yamlString), &unmarshalledYaml)
	if err != nil {
		t.Errorf("error unmarshalling yaml: %v", err)
	}
	return unmarshalledYaml
}

func TestValidateSchema(t *testing.T) {
	getJackalSchema := func(t *testing.T) []byte {
		t.Helper()
		file, err := os.ReadFile("../../../../jackal.schema.json")
		if err != nil {
			t.Errorf("error reading file: %v", err)
		}
		return file
	}

	t.Run("validate schema success", func(t *testing.T) {
		unmarshalledYaml := readAndUnmarshalYaml[interface{}](t, goodJackalPackage)
		validator := Validator{untypedJackalPackage: unmarshalledYaml, jsonSchema: getJackalSchema(t)}
		err := validateSchema(&validator)
		require.NoError(t, err)
		require.Empty(t, validator.findings)
	})

	t.Run("validate schema fail", func(t *testing.T) {
		unmarshalledYaml := readAndUnmarshalYaml[interface{}](t, badJackalPackage)
		validator := Validator{untypedJackalPackage: unmarshalledYaml, jsonSchema: getJackalSchema(t)}
		err := validateSchema(&validator)
		require.NoError(t, err)
		config.NoColor = true
		require.Equal(t, "Additional property not-path is not allowed", validator.findings[0].String())
		require.Equal(t, "Invalid type. Expected: string, given: integer", validator.findings[1].String())
	})

	t.Run("Template in component import success", func(t *testing.T) {
		unmarshalledYaml := readAndUnmarshalYaml[types.JackalPackage](t, goodJackalPackage)
		validator := Validator{typedJackalPackage: unmarshalledYaml}
		for _, component := range validator.typedJackalPackage.Components {
			lintComponent(&validator, &composer.Node{JackalComponent: component})
		}
		require.Empty(t, validator.findings)
	})

	t.Run("Path template in component import failure", func(t *testing.T) {
		pathVar := "###JACKAL_PKG_TMPL_PATH###"
		pathComponent := types.JackalComponent{Import: types.JackalComponentImport{Path: pathVar}}
		validator := Validator{typedJackalPackage: types.JackalPackage{Components: []types.JackalComponent{pathComponent}}}
		checkForVarInComponentImport(&validator, &composer.Node{JackalComponent: pathComponent})
		require.Equal(t, pathVar, validator.findings[0].item)
	})

	t.Run("OCI template in component import failure", func(t *testing.T) {
		ociPathVar := "oci://###JACKAL_PKG_TMPL_PATH###"
		URLComponent := types.JackalComponent{Import: types.JackalComponentImport{URL: ociPathVar}}
		validator := Validator{typedJackalPackage: types.JackalPackage{Components: []types.JackalComponent{URLComponent}}}
		checkForVarInComponentImport(&validator, &composer.Node{JackalComponent: URLComponent})
		require.Equal(t, ociPathVar, validator.findings[0].item)
	})

	t.Run("Unpinnned repo warning", func(t *testing.T) {
		validator := Validator{}
		unpinnedRepo := "https://github.com/defenseunicorns/jackal-public-test.git"
		component := types.JackalComponent{Repos: []string{
			unpinnedRepo,
			"https://dev.azure.com/defenseunicorns/jackal-public-test/_git/jackal-public-test@v0.0.1"}}
		checkForUnpinnedRepos(&validator, &composer.Node{JackalComponent: component})
		require.Equal(t, unpinnedRepo, validator.findings[0].item)
		require.Equal(t, len(validator.findings), 1)
	})

	t.Run("Unpinnned image warning", func(t *testing.T) {
		validator := Validator{}
		unpinnedImage := "registry.com:9001/whatever/image:1.0.0"
		badImage := "badimage:badimage@@sha256:3fbc632167424a6d997e74f5"
		component := types.JackalComponent{Images: []string{
			unpinnedImage,
			"busybox:latest@sha256:3fbc632167424a6d997e74f52b878d7cc478225cffac6bc977eedfe51c7f4e79",
			badImage}}
		checkForUnpinnedImages(&validator, &composer.Node{JackalComponent: component})
		require.Equal(t, unpinnedImage, validator.findings[0].item)
		require.Equal(t, badImage, validator.findings[1].item)
		require.Equal(t, 2, len(validator.findings))

	})

	t.Run("Unpinnned file warning", func(t *testing.T) {
		validator := Validator{}
		fileURL := "http://example.com/file.zip"
		localFile := "local.txt"
		jackalFiles := []types.JackalFile{
			{
				Source: fileURL,
			},
			{
				Source: localFile,
			},
			{
				Source: fileURL,
				Shasum: "fake-shasum",
			},
		}
		component := types.JackalComponent{Files: jackalFiles}
		checkForUnpinnedFiles(&validator, &composer.Node{JackalComponent: component})
		require.Equal(t, fileURL, validator.findings[0].item)
		require.Equal(t, 1, len(validator.findings))
	})

	t.Run("Wrap standalone numbers in bracket", func(t *testing.T) {
		input := "components12.12.import.path"
		expected := ".components12.[12].import.path"
		actual := makeFieldPathYqCompat(input)
		require.Equal(t, expected, actual)
	})

	t.Run("root doesn't change", func(t *testing.T) {
		input := "(root)"
		actual := makeFieldPathYqCompat(input)
		require.Equal(t, input, actual)
	})

	t.Run("Test composable components", func(t *testing.T) {
		pathVar := "fake-path"
		unpinnedImage := "unpinned:latest"
		pathComponent := types.JackalComponent{
			Import: types.JackalComponentImport{Path: pathVar},
			Images: []string{unpinnedImage}}
		validator := Validator{
			typedJackalPackage: types.JackalPackage{Components: []types.JackalComponent{pathComponent},
				Metadata: types.JackalMetadata{Name: "test-jackal-package"}}}

		createOpts := types.JackalCreateOptions{Flavor: "", BaseDir: "."}
		lintComponents(&validator, &createOpts)
		// Require.contains rather than equals since the error message changes from linux to windows
		require.Contains(t, validator.findings[0].description, fmt.Sprintf("open %s", filepath.Join("fake-path", "jackal.yaml")))
		require.Equal(t, ".components.[0].import.path", validator.findings[0].yqPath)
		require.Equal(t, ".", validator.findings[0].packageRelPath)
		require.Equal(t, unpinnedImage, validator.findings[1].item)
		require.Equal(t, ".", validator.findings[1].packageRelPath)
	})

	t.Run("isImagePinned", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			input    string
			expected bool
			err      error
		}{
			{
				input:    "registry.com:8080/defenseunicorns/whatever",
				expected: false,
				err:      nil,
			},
			{
				input:    "ghcr.io/defenseunicorns/pepr/controller:v0.15.0",
				expected: false,
				err:      nil,
			},
			{
				input:    "busybox:latest@sha256:3fbc632167424a6d997e74f52b878d7cc478225cffac6bc977eedfe51c7f4e79",
				expected: true,
				err:      nil,
			},
			{
				input:    "busybox:bad/image",
				expected: false,
				err:      errors.New("invalid reference format"),
			},
			{
				input:    "busybox:###JACKAL_PKG_TMPL_BUSYBOX_IMAGE###",
				expected: true,
				err:      nil,
			},
		}
		for _, tc := range tests {
			t.Run(tc.input, func(t *testing.T) {
				actual, err := isPinnedImage(tc.input)
				if err != nil {
					require.EqualError(t, err, tc.err.Error())
				}
				require.Equal(t, tc.expected, actual)
			})
		}
	})
}

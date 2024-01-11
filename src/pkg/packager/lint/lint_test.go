// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package lint contains functions for verifying zarf yaml files are valid
package lint

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/packager/composer"
	"github.com/defenseunicorns/zarf/src/types"
	goyaml "github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

const badZarfPackage = `
kind: ZarfInitConfig
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

const goodZarfPackage = `
kind: ZarfPackageConfig
metadata:
  name: good-zarf-package

components:
  - name: baseline
    required: true
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
	getZarfSchema := func(t *testing.T) []byte {
		t.Helper()
		file, err := os.ReadFile("../../../../zarf.schema.json")
		if err != nil {
			t.Errorf("error reading file: %v", err)
		}
		return file
	}

	t.Run("validate schema success", func(t *testing.T) {
		unmarshalledYaml := readAndUnmarshalYaml[interface{}](t, goodZarfPackage)
		validator := Validator{untypedZarfPackage: unmarshalledYaml, jsonSchema: getZarfSchema(t)}
		err := validateSchema(&validator)
		require.NoError(t, err)
		require.Empty(t, validator.findings)
	})

	t.Run("validate schema fail", func(t *testing.T) {
		unmarshalledYaml := readAndUnmarshalYaml[interface{}](t, badZarfPackage)
		validator := Validator{untypedZarfPackage: unmarshalledYaml, jsonSchema: getZarfSchema(t)}
		err := validateSchema(&validator)
		require.NoError(t, err)
		config.NoColor = true
		require.Equal(t, "Additional property not-path is not allowed", validator.findings[0].String())
		require.Equal(t, "Invalid type. Expected: string, given: integer", validator.findings[1].String())
	})

	t.Run("Template in component import success", func(t *testing.T) {
		unmarshalledYaml := readAndUnmarshalYaml[types.ZarfPackage](t, goodZarfPackage)
		validator := Validator{typedZarfPackage: unmarshalledYaml}
		for _, component := range validator.typedZarfPackage.Components {
			lintComponent(&validator, &composer.Node{ZarfComponent: component})
		}
		require.Empty(t, validator.findings)
	})

	t.Run("Path template in component import failure", func(t *testing.T) {
		pathVar := "###ZARF_PKG_TMPL_PATH###"
		pathComponent := types.ZarfComponent{Import: types.ZarfComponentImport{Path: pathVar}}
		validator := Validator{typedZarfPackage: types.ZarfPackage{Components: []types.ZarfComponent{pathComponent}}}
		checkForVarInComponentImport(&validator, &composer.Node{ZarfComponent: pathComponent})
		require.Equal(t, pathVar, validator.findings[0].item)
	})

	t.Run("OCI template in component import failure", func(t *testing.T) {
		ociPathVar := "oci://###ZARF_PKG_TMPL_PATH###"
		URLComponent := types.ZarfComponent{Import: types.ZarfComponentImport{URL: ociPathVar}}
		validator := Validator{typedZarfPackage: types.ZarfPackage{Components: []types.ZarfComponent{URLComponent}}}
		checkForVarInComponentImport(&validator, &composer.Node{ZarfComponent: URLComponent})
		require.Equal(t, ociPathVar, validator.findings[0].item)
	})

	t.Run("Unpinnned repo warning", func(t *testing.T) {
		validator := Validator{}
		unpinnedRepo := "https://github.com/defenseunicorns/zarf-public-test.git"
		component := types.ZarfComponent{Repos: []string{
			unpinnedRepo,
			"https://dev.azure.com/defenseunicorns/zarf-public-test/_git/zarf-public-test@v0.0.1"}}
		checkForUnpinnedRepos(&validator, &composer.Node{ZarfComponent: component})
		require.Equal(t, unpinnedRepo, validator.findings[0].item)
		require.Equal(t, len(validator.findings), 1)
	})

	t.Run("Unpinnned image warning", func(t *testing.T) {
		validator := Validator{}
		unpinnedImage := "registry.com:9001/whatever/image:1.0.0"
		badImage := "badimage:badimage@@sha256:3fbc632167424a6d997e74f5"
		component := types.ZarfComponent{Images: []string{
			unpinnedImage,
			"busybox:latest@sha256:3fbc632167424a6d997e74f52b878d7cc478225cffac6bc977eedfe51c7f4e79",
			badImage}}
		checkForUnpinnedImages(&validator, &composer.Node{ZarfComponent: component})
		require.Equal(t, unpinnedImage, validator.findings[0].item)
		require.Equal(t, badImage, validator.findings[1].item)
		require.Equal(t, 2, len(validator.findings))

	})

	t.Run("Unpinnned file warning", func(t *testing.T) {
		validator := Validator{}
		fileURL := "http://example.com/file.zip"
		localFile := "local.txt"
		zarfFiles := []types.ZarfFile{
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
		component := types.ZarfComponent{Files: zarfFiles}
		checkForUnpinnedFiles(&validator, &composer.Node{ZarfComponent: component})
		require.Equal(t, fileURL, validator.findings[0].item)
		require.Equal(t, 1, len(validator.findings))
	})

	t.Run("Wrap standalone numbers in bracket", func(t *testing.T) {
		input := "components12.12.import.path"
		expected := ".components12.[12].import.path"
		acutal := makeFieldPathYqCompat(input)
		require.Equal(t, expected, acutal)
	})

	t.Run("root doesn't change", func(t *testing.T) {
		input := "(root)"
		acutal := makeFieldPathYqCompat(input)
		require.Equal(t, input, acutal)
	})

	t.Run("remove var from validator", func(t *testing.T) {
		validator := Validator{unusedVariables: []string{"FAKE_VAR"}}
		line := "Hello my name is ###ZARF_VAR_FAKE_VAR###"
		declareVarIsUsed(&validator, line)
		require.Empty(t, validator.unusedVariables)
	})

	t.Run("Test composable components", func(t *testing.T) {
		pathVar := "fake-path"
		unpinnedImage := "unpinned:latest"
		pathComponent := types.ZarfComponent{
			Import: types.ZarfComponentImport{Path: pathVar},
			Images: []string{unpinnedImage}}
		validator := Validator{
			typedZarfPackage: types.ZarfPackage{Components: []types.ZarfComponent{pathComponent},
				Metadata: types.ZarfMetadata{Name: "test-zarf-package"}}}

		createOpts := types.ZarfCreateOptions{Flavor: "", BaseDir: "."}
		cfg := types.PackagerConfig{CreateOpts: createOpts}
		lintComponents(&validator, &cfg)
		// Require.contains rather than equals since the error message changes from linux to windows
		require.Contains(t, validator.findings[0].description, fmt.Sprintf("open %s", filepath.Join("fake-path", "zarf.yaml")))
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
				input:    "busybox:###ZARF_PKG_TMPL_BUSYBOX_IMAGE###",
				expected: true,
				err:      nil,
			},
		}
		for _, tc := range tests {
			t.Run(tc.input, func(t *testing.T) {
				acutal, err := isPinnedImage(tc.input)
				if err != nil {
					require.EqualError(t, err, tc.err.Error())
				}
				require.Equal(t, tc.expected, acutal)
			})
		}
	})
}

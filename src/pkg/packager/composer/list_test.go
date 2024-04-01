// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package composer contains functions for composing components within Jackal packages.
package composer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/defenseunicorns/jackal/src/types"
	"github.com/defenseunicorns/jackal/src/types/extensions"
	"github.com/stretchr/testify/require"
)

func TestNewImportChain(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name                 string
		head                 types.JackalComponent
		arch                 string
		flavor               string
		expectedErrorMessage string
	}

	testCases := []testCase{
		{
			name:                 "No Architecture",
			head:                 types.JackalComponent{},
			expectedErrorMessage: "architecture must be provided",
		},
		{
			name: "Circular Import",
			head: types.JackalComponent{
				Import: types.JackalComponentImport{
					Path: ".",
				},
			},
			arch:                 "amd64",
			expectedErrorMessage: "detected circular import chain",
		},
	}
	testPackageName := "test-package"
	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewImportChain(testCase.head, 0, testPackageName, testCase.arch, testCase.flavor)
			require.Contains(t, err.Error(), testCase.expectedErrorMessage)
		})
	}
}

func TestCompose(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name                 string
		ic                   *ImportChain
		returnError          bool
		expectedComposed     types.JackalComponent
		expectedErrorMessage string
	}

	firstDirectory := "hello"
	secondDirectory := "world"
	finalDirectory := filepath.Join(firstDirectory, secondDirectory)

	finalDirectoryActionDefault := filepath.Join(firstDirectory, secondDirectory, "today-dc")
	secondDirectoryActionDefault := filepath.Join(firstDirectory, "world-dc")
	firstDirectoryActionDefault := "hello-dc"

	testCases := []testCase{
		{
			name: "Single Component",
			ic: createChainFromSlice([]types.JackalComponent{
				{
					Name: "no-import",
				},
			}),
			returnError: false,
			expectedComposed: types.JackalComponent{
				Name: "no-import",
			},
		},
		{
			name: "Multiple Components",
			ic: createChainFromSlice([]types.JackalComponent{
				createDummyComponent("hello", firstDirectory, "hello"),
				createDummyComponent("world", secondDirectory, "world"),
				createDummyComponent("today", "", "hello"),
			}),
			returnError: false,
			expectedComposed: types.JackalComponent{
				Name: "import-hello",
				// Files should always be appended with corrected directories
				Files: []types.JackalFile{
					{Source: fmt.Sprintf("%s%stoday.txt", finalDirectory, string(os.PathSeparator))},
					{Source: fmt.Sprintf("%s%sworld.txt", firstDirectory, string(os.PathSeparator))},
					{Source: "hello.txt"},
				},
				// Charts should be merged if names match and appended if not with corrected directories
				Charts: []types.JackalChart{
					{
						Name:      "hello",
						LocalPath: fmt.Sprintf("%s%schart", finalDirectory, string(os.PathSeparator)),
						ValuesFiles: []string{
							fmt.Sprintf("%s%svalues.yaml", finalDirectory, string(os.PathSeparator)),
							"values.yaml",
						},
					},
					{
						Name:      "world",
						LocalPath: fmt.Sprintf("%s%schart", firstDirectory, string(os.PathSeparator)),
						ValuesFiles: []string{
							fmt.Sprintf("%s%svalues.yaml", firstDirectory, string(os.PathSeparator)),
						},
					},
				},
				// Manifests should be merged if names match and appended if not with corrected directories
				Manifests: []types.JackalManifest{
					{
						Name: "hello",
						Files: []string{
							fmt.Sprintf("%s%smanifest.yaml", finalDirectory, string(os.PathSeparator)),
							"manifest.yaml",
						},
					},
					{
						Name: "world",
						Files: []string{
							fmt.Sprintf("%s%smanifest.yaml", firstDirectory, string(os.PathSeparator)),
						},
					},
				},
				// DataInjections should always be appended with corrected directories
				DataInjections: []types.JackalDataInjection{
					{Source: fmt.Sprintf("%s%stoday", finalDirectory, string(os.PathSeparator))},
					{Source: fmt.Sprintf("%s%sworld", firstDirectory, string(os.PathSeparator))},
					{Source: "hello"},
				},
				Actions: types.JackalComponentActions{
					// OnCreate actions should be appended with corrected directories that properly handle default directories
					OnCreate: types.JackalComponentActionSet{
						Defaults: types.JackalComponentActionDefaults{
							Dir: "hello-dc",
						},
						Before: []types.JackalComponentAction{
							{Cmd: "today-bc", Dir: &finalDirectoryActionDefault},
							{Cmd: "world-bc", Dir: &secondDirectoryActionDefault},
							{Cmd: "hello-bc", Dir: &firstDirectoryActionDefault},
						},
						After: []types.JackalComponentAction{
							{Cmd: "today-ac", Dir: &finalDirectoryActionDefault},
							{Cmd: "world-ac", Dir: &secondDirectoryActionDefault},
							{Cmd: "hello-ac", Dir: &firstDirectoryActionDefault},
						},
						OnSuccess: []types.JackalComponentAction{
							{Cmd: "today-sc", Dir: &finalDirectoryActionDefault},
							{Cmd: "world-sc", Dir: &secondDirectoryActionDefault},
							{Cmd: "hello-sc", Dir: &firstDirectoryActionDefault},
						},
						OnFailure: []types.JackalComponentAction{
							{Cmd: "today-fc", Dir: &finalDirectoryActionDefault},
							{Cmd: "world-fc", Dir: &secondDirectoryActionDefault},
							{Cmd: "hello-fc", Dir: &firstDirectoryActionDefault},
						},
					},
					// OnDeploy actions should be appended without corrected directories
					OnDeploy: types.JackalComponentActionSet{
						Defaults: types.JackalComponentActionDefaults{
							Dir: "hello-dd",
						},
						Before: []types.JackalComponentAction{
							{Cmd: "today-bd"},
							{Cmd: "world-bd"},
							{Cmd: "hello-bd"},
						},
						After: []types.JackalComponentAction{
							{Cmd: "today-ad"},
							{Cmd: "world-ad"},
							{Cmd: "hello-ad"},
						},
						OnSuccess: []types.JackalComponentAction{
							{Cmd: "today-sd"},
							{Cmd: "world-sd"},
							{Cmd: "hello-sd"},
						},
						OnFailure: []types.JackalComponentAction{
							{Cmd: "today-fd"},
							{Cmd: "world-fd"},
							{Cmd: "hello-fd"},
						},
					},
					// OnRemove actions should be appended without corrected directories
					OnRemove: types.JackalComponentActionSet{
						Defaults: types.JackalComponentActionDefaults{
							Dir: "hello-dr",
						},
						Before: []types.JackalComponentAction{
							{Cmd: "today-br"},
							{Cmd: "world-br"},
							{Cmd: "hello-br"},
						},
						After: []types.JackalComponentAction{
							{Cmd: "today-ar"},
							{Cmd: "world-ar"},
							{Cmd: "hello-ar"},
						},
						OnSuccess: []types.JackalComponentAction{
							{Cmd: "today-sr"},
							{Cmd: "world-sr"},
							{Cmd: "hello-sr"},
						},
						OnFailure: []types.JackalComponentAction{
							{Cmd: "today-fr"},
							{Cmd: "world-fr"},
							{Cmd: "hello-fr"},
						},
					},
				},
				// Extensions should be appended with corrected directories
				Extensions: extensions.JackalComponentExtensions{
					BigBang: &extensions.BigBang{
						ValuesFiles: []string{
							fmt.Sprintf("%s%svalues.yaml", finalDirectory, string(os.PathSeparator)),
							fmt.Sprintf("%s%svalues.yaml", firstDirectory, string(os.PathSeparator)),
							"values.yaml",
						},
						FluxPatchFiles: []string{
							fmt.Sprintf("%s%spatch.yaml", finalDirectory, string(os.PathSeparator)),
							fmt.Sprintf("%s%spatch.yaml", firstDirectory, string(os.PathSeparator)),
							"patch.yaml",
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			composed, err := testCase.ic.Compose()
			if testCase.returnError {
				require.Contains(t, err.Error(), testCase.expectedErrorMessage)
			} else {
				require.EqualValues(t, &testCase.expectedComposed, composed)
			}
		})
	}
}

func TestMerging(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		ic             *ImportChain
		existingVars   []types.JackalPackageVariable
		existingConsts []types.JackalPackageConstant
		expectedVars   []types.JackalPackageVariable
		expectedConsts []types.JackalPackageConstant
	}

	head := Node{
		vars: []types.JackalPackageVariable{
			{
				Name:    "TEST",
				Default: "head",
			},
			{
				Name: "HEAD",
			},
		},
		consts: []types.JackalPackageConstant{
			{
				Name:  "TEST",
				Value: "head",
			},
			{
				Name: "HEAD",
			},
		},
	}
	tail := Node{
		vars: []types.JackalPackageVariable{
			{
				Name:    "TEST",
				Default: "tail",
			},
			{
				Name: "TAIL",
			},
		},
		consts: []types.JackalPackageConstant{
			{
				Name:  "TEST",
				Value: "tail",
			},
			{
				Name: "TAIL",
			},
		},
	}
	head.next = &tail
	tail.prev = &head
	testIC := &ImportChain{head: &head, tail: &tail}

	testCases := []testCase{
		{
			name: "empty-ic",
			ic:   &ImportChain{},
			existingVars: []types.JackalPackageVariable{
				{
					Name: "TEST",
				},
			},
			existingConsts: []types.JackalPackageConstant{
				{
					Name: "TEST",
				},
			},
			expectedVars: []types.JackalPackageVariable{
				{
					Name: "TEST",
				},
			},
			expectedConsts: []types.JackalPackageConstant{
				{
					Name: "TEST",
				},
			},
		},
		{
			name:           "no-existing",
			ic:             testIC,
			existingVars:   []types.JackalPackageVariable{},
			existingConsts: []types.JackalPackageConstant{},
			expectedVars: []types.JackalPackageVariable{
				{
					Name:    "TEST",
					Default: "head",
				},
				{
					Name: "HEAD",
				},
				{
					Name: "TAIL",
				},
			},
			expectedConsts: []types.JackalPackageConstant{
				{
					Name:  "TEST",
					Value: "head",
				},
				{
					Name: "HEAD",
				},
				{
					Name: "TAIL",
				},
			},
		},
		{
			name: "with-existing",
			ic:   testIC,
			existingVars: []types.JackalPackageVariable{
				{
					Name:    "TEST",
					Default: "existing",
				},
				{
					Name: "EXISTING",
				},
			},
			existingConsts: []types.JackalPackageConstant{
				{
					Name:  "TEST",
					Value: "existing",
				},
				{
					Name: "EXISTING",
				},
			},
			expectedVars: []types.JackalPackageVariable{
				{
					Name:    "TEST",
					Default: "existing",
				},
				{
					Name: "EXISTING",
				},
				{
					Name: "HEAD",
				},
				{
					Name: "TAIL",
				},
			},
			expectedConsts: []types.JackalPackageConstant{
				{
					Name:  "TEST",
					Value: "existing",
				},
				{
					Name: "EXISTING",
				},
				{
					Name: "HEAD",
				},
				{
					Name: "TAIL",
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mergedVars := testCase.ic.MergeVariables(testCase.existingVars)
			require.EqualValues(t, testCase.expectedVars, mergedVars)

			mergedConsts := testCase.ic.MergeConstants(testCase.existingConsts)
			require.EqualValues(t, testCase.expectedConsts, mergedConsts)
		})
	}
}

func createChainFromSlice(components []types.JackalComponent) (ic *ImportChain) {
	ic = &ImportChain{}
	testPackageName := "test-package"

	if len(components) == 0 {
		return ic
	}

	ic.append(components[0], 0, testPackageName, ".", nil, nil)
	history := []string{}

	for idx := 1; idx < len(components); idx++ {
		history = append(history, components[idx-1].Import.Path)
		ic.append(components[idx], idx, testPackageName, filepath.Join(history...), nil, nil)
	}

	return ic
}

func createDummyComponent(name, importDir, subName string) types.JackalComponent {
	return types.JackalComponent{
		Name: fmt.Sprintf("import-%s", name),
		Import: types.JackalComponentImport{
			Path: importDir,
		},
		Files: []types.JackalFile{
			{
				Source: fmt.Sprintf("%s.txt", name),
			},
		},
		Charts: []types.JackalChart{
			{
				Name:      subName,
				LocalPath: "chart",
				ValuesFiles: []string{
					"values.yaml",
				},
			},
		},
		Manifests: []types.JackalManifest{
			{
				Name: subName,
				Files: []string{
					"manifest.yaml",
				},
			},
		},
		DataInjections: []types.JackalDataInjection{
			{
				Source: name,
			},
		},
		Actions: types.JackalComponentActions{
			OnCreate: types.JackalComponentActionSet{
				Defaults: types.JackalComponentActionDefaults{
					Dir: name + "-dc",
				},
				Before: []types.JackalComponentAction{
					{Cmd: name + "-bc"},
				},
				After: []types.JackalComponentAction{
					{Cmd: name + "-ac"},
				},
				OnSuccess: []types.JackalComponentAction{
					{Cmd: name + "-sc"},
				},
				OnFailure: []types.JackalComponentAction{
					{Cmd: name + "-fc"},
				},
			},
			OnDeploy: types.JackalComponentActionSet{
				Defaults: types.JackalComponentActionDefaults{
					Dir: name + "-dd",
				},
				Before: []types.JackalComponentAction{
					{Cmd: name + "-bd"},
				},
				After: []types.JackalComponentAction{
					{Cmd: name + "-ad"},
				},
				OnSuccess: []types.JackalComponentAction{
					{Cmd: name + "-sd"},
				},
				OnFailure: []types.JackalComponentAction{
					{Cmd: name + "-fd"},
				},
			},
			OnRemove: types.JackalComponentActionSet{
				Defaults: types.JackalComponentActionDefaults{
					Dir: name + "-dr",
				},
				Before: []types.JackalComponentAction{
					{Cmd: name + "-br"},
				},
				After: []types.JackalComponentAction{
					{Cmd: name + "-ar"},
				},
				OnSuccess: []types.JackalComponentAction{
					{Cmd: name + "-sr"},
				},
				OnFailure: []types.JackalComponentAction{
					{Cmd: name + "-fr"},
				},
			},
		},
		Extensions: extensions.JackalComponentExtensions{
			BigBang: &extensions.BigBang{
				ValuesFiles: []string{
					"values.yaml",
				},
				FluxPatchFiles: []string{
					"patch.yaml",
				},
			},
		},
	}
}

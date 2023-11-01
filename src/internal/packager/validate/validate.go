// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package validate provides Zarf package validation functions.
package validate

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/config/lang"
	"github.com/defenseunicorns/zarf/src/pkg/oci"
	"github.com/defenseunicorns/zarf/src/pkg/utils/helpers"
	"github.com/defenseunicorns/zarf/src/types"
)

var (
	// IsLowercaseNumberHyphen is a regex for lowercase, numbers and hyphens.
	// https://regex101.com/r/FLdG9G/1
	IsLowercaseNumberHyphen = regexp.MustCompile(`^[a-z0-9\-]+$`).MatchString
	// IsUppercaseNumberUnderscore is a regex for uppercase, numbers and underscores.
	// https://regex101.com/r/tfsEuZ/1
	IsUppercaseNumberUnderscore = regexp.MustCompile(`^[A-Z0-9_]+$`).MatchString
)

// Run performs config validations.
func Run(pkg types.ZarfPackage) error {
	if pkg.Kind == types.ZarfInitConfig && pkg.Metadata.YOLO {
		return fmt.Errorf(lang.PkgValidateErrInitNoYOLO)
	}

	if err := validatePackageName(pkg.Metadata.Name); err != nil {
		return fmt.Errorf(lang.PkgValidateErrName, err)
	}

	for _, variable := range pkg.Variables {
		if err := validatePackageVariable(variable); err != nil {
			return fmt.Errorf(lang.PkgValidateErrVariable, err)
		}
	}

	for _, constant := range pkg.Constants {
		if err := validatePackageConstant(constant); err != nil {
			return fmt.Errorf(lang.PkgValidateErrConstant, err)
		}
	}

	uniqueComponentNames := make(map[string]bool)

	for _, component := range pkg.Components {
		// ensure component name is unique
		if _, ok := uniqueComponentNames[component.Name]; ok {
			return fmt.Errorf(lang.PkgValidateErrComponentNameNotUnique, component.Name)
		}
		uniqueComponentNames[component.Name] = true

		if err := validateComponent(pkg, component); err != nil {
			return fmt.Errorf(lang.PkgValidateErrComponent, component.Name, err)
		}
	}

	return nil
}

// ImportDefinition validates the component trying to be imported.
func ImportDefinition(component *types.ZarfComponent) error {
	path := component.Import.Path
	url := component.Import.URL

	// ensure path or url is provided
	if path == "" && url == "" {
		return fmt.Errorf(lang.PkgValidateErrImportDefinition, component.Name, "neither a path nor a URL was provided")
	}

	// ensure path and url are not both provided
	if path != "" && url != "" {
		return fmt.Errorf(lang.PkgValidateErrImportDefinition, component.Name, "both a path and a URL were provided")
	}

	// validation for path
	if url == "" && path != "" {
		// ensure path is not an absolute path
		if filepath.IsAbs(path) {
			return fmt.Errorf(lang.PkgValidateErrImportDefinition, component.Name, "path cannot be an absolute path")
		}
	}

	// validation for url
	if url != "" && path == "" {
		ok := helpers.IsOCIURL(url)
		if !ok {
			return fmt.Errorf(lang.PkgValidateErrImportDefinition, component.Name, "URL is not a valid OCI URL")
		}
		if !strings.HasSuffix(url, oci.SkeletonSuffix) {
			return fmt.Errorf(lang.PkgValidateErrImportDefinition, component.Name, "OCI import URL must end with -skeleton")
		}
	}

	return nil
}

func oneIfNotEmpty(testString string) int {
	if testString == "" {
		return 0
	}

	return 1
}

func validateComponent(pkg types.ZarfPackage, component types.ZarfComponent) error {
	if component.Required {
		if component.Default {
			return fmt.Errorf(lang.PkgValidateErrComponentReqDefault, component.Name)
		}
		if component.Group != "" {
			return fmt.Errorf(lang.PkgValidateErrComponentReqGrouped, component.Name)
		}
	}

	uniqueChartNames := make(map[string]bool)
	for _, chart := range component.Charts {
		// ensure chart name is unique
		if _, ok := uniqueChartNames[chart.Name]; ok {
			return fmt.Errorf(lang.PkgValidateErrChartNameNotUnique, chart.Name)
		}
		uniqueChartNames[chart.Name] = true

		if err := validateChart(chart); err != nil {
			return fmt.Errorf(lang.PkgValidateErrChart, err)
		}
	}

	uniqueManifestNames := make(map[string]bool)
	for _, manifest := range component.Manifests {
		// ensure manifest name is unique
		if _, ok := uniqueManifestNames[manifest.Name]; ok {
			return fmt.Errorf(lang.PkgValidateErrManifestNameNotUnique, manifest.Name)
		}
		uniqueManifestNames[manifest.Name] = true

		if err := validateManifest(manifest); err != nil {
			return fmt.Errorf(lang.PkgValidateErrManifest, err)
		}
	}

	if pkg.Metadata.YOLO {
		if err := validateYOLO(component); err != nil {
			return fmt.Errorf(lang.PkgValidateErrComponentYOLO, component.Name, err)
		}
	}

	if containsVariables, err := validateActionset(component.Actions.OnCreate); err != nil {
		return fmt.Errorf(lang.PkgValidateErrAction, err)
	} else if containsVariables {
		return fmt.Errorf(lang.PkgValidateErrActionVariables, component.Name)
	}

	if _, err := validateActionset(component.Actions.OnDeploy); err != nil {
		return fmt.Errorf(lang.PkgValidateErrAction, err)
	}

	if containsVariables, err := validateActionset(component.Actions.OnRemove); err != nil {
		return fmt.Errorf(lang.PkgValidateErrAction, err)
	} else if containsVariables {
		return fmt.Errorf(lang.PkgValidateErrActionVariables, component.Name)
	}

	return nil
}

func validateActionset(actions types.ZarfComponentActionSet) (bool, error) {
	containsVariables := false

	validate := func(actions []types.ZarfComponentAction) error {
		for _, action := range actions {
			if cv, err := validateAction(action); err != nil {
				return err
			} else if cv {
				containsVariables = true
			}
		}

		return nil
	}

	if err := validate(actions.Before); err != nil {
		return containsVariables, err
	}
	if err := validate(actions.After); err != nil {
		return containsVariables, err
	}
	if err := validate(actions.OnSuccess); err != nil {
		return containsVariables, err
	}
	if err := validate(actions.OnFailure); err != nil {
		return containsVariables, err
	}

	return containsVariables, nil
}

func validateAction(action types.ZarfComponentAction) (bool, error) {
	containsVariables := false

	// Validate SetVariable
	for _, variable := range action.SetVariables {
		if !IsUppercaseNumberUnderscore(variable.Name) {
			return containsVariables, fmt.Errorf(lang.PkgValidateMustBeUppercase, variable.Name)
		}
		containsVariables = true
	}

	if action.Wait != nil {
		// Validate only cmd or wait, not both
		if action.Cmd != "" {
			return containsVariables, fmt.Errorf(lang.PkgValidateErrActionCmdWait, action.Cmd)
		}

		// Validate only cluster or network, not both
		if action.Wait.Cluster != nil && action.Wait.Network != nil {
			return containsVariables, fmt.Errorf(lang.PkgValidateErrActionClusterNetwork)
		}

		// Validate at least one of cluster or network
		if action.Wait.Cluster == nil && action.Wait.Network == nil {
			return containsVariables, fmt.Errorf(lang.PkgValidateErrActionClusterNetwork)
		}
	}

	return containsVariables, nil
}

func validateYOLO(component types.ZarfComponent) error {
	if len(component.Images) > 0 {
		return fmt.Errorf(lang.PkgValidateErrYOLONoOCI)
	}

	if len(component.Repos) > 0 {
		return fmt.Errorf(lang.PkgValidateErrYOLONoGit)
	}

	if component.Only.Cluster.Architecture != "" {
		return fmt.Errorf(lang.PkgValidateErrYOLONoArch)
	}

	if len(component.Only.Cluster.Distros) > 0 {
		return fmt.Errorf(lang.PkgValidateErrYOLONoDistro)
	}

	return nil
}

func validatePackageName(subject string) error {
	if !IsLowercaseNumberHyphen(subject) {
		return fmt.Errorf(lang.PkgValidateErrPkgName, subject)
	}

	return nil
}

func validatePackageVariable(subject types.ZarfPackageVariable) error {
	// ensure the variable name is only capitals and underscores
	if !IsUppercaseNumberUnderscore(subject.Name) {
		return fmt.Errorf(lang.PkgValidateMustBeUppercase, subject.Name)
	}

	return nil
}

func validatePackageConstant(subject types.ZarfPackageConstant) error {
	// ensure the constant name is only capitals and underscores
	if !IsUppercaseNumberUnderscore(subject.Name) {
		return fmt.Errorf(lang.PkgValidateErrPkgConstantName, subject.Name)
	}

	if !regexp.MustCompile(subject.Pattern).MatchString(subject.Value) {
		return fmt.Errorf(lang.PkgValidateErrPkgConstantPattern, subject.Name, subject.Pattern)
	}

	return nil
}

func validateChart(chart types.ZarfChart) error {
	// Don't allow empty names
	if chart.Name == "" {
		return fmt.Errorf(lang.PkgValidateErrChartNameMissing, chart.Name)
	}

	// Helm max release name
	if len(chart.Name) > config.ZarfMaxChartNameLength {
		return fmt.Errorf(lang.PkgValidateErrChartName, chart.Name, config.ZarfMaxChartNameLength)
	}

	// Must have a namespace
	if chart.Namespace == "" {
		return fmt.Errorf(lang.PkgValidateErrChartNamespaceMissing, chart.Name)
	}

	// Must have a url or localPath (and not both)
	count := oneIfNotEmpty(chart.URL) + oneIfNotEmpty(chart.LocalPath)
	if count != 1 {
		return fmt.Errorf(lang.PkgValidateErrChartURLOrPath, chart.Name)
	}

	// Must have a version
	if chart.Version == "" {
		return fmt.Errorf(lang.PkgValidateErrChartVersion, chart.Name)
	}

	return nil
}

func validateManifest(manifest types.ZarfManifest) error {
	// Don't allow empty names
	if manifest.Name == "" {
		return fmt.Errorf(lang.PkgValidateErrManifestNameMissing, manifest.Name)
	}

	// Helm max release name
	if len(manifest.Name) > config.ZarfMaxChartNameLength {
		return fmt.Errorf(lang.PkgValidateErrManifestNameLength, manifest.Name, config.ZarfMaxChartNameLength)
	}

	// Require files in manifest
	if len(manifest.Files) < 1 && len(manifest.Kustomizations) < 1 {
		return fmt.Errorf(lang.PkgValidateErrManifestFileOrKustomize, manifest.Name)
	}

	return nil
}

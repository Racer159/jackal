// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package packager contains functions for interacting with, managing and deploying Zarf packages.
package packager

import (
	"crypto"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/config/lang"
	"github.com/defenseunicorns/zarf/src/internal/packager/git"
	"github.com/defenseunicorns/zarf/src/internal/packager/helm"
	"github.com/defenseunicorns/zarf/src/internal/packager/images"
	"github.com/defenseunicorns/zarf/src/internal/packager/kustomize"
	"github.com/defenseunicorns/zarf/src/internal/packager/sbom"
	"github.com/defenseunicorns/zarf/src/internal/packager/validate"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/oci"
	"github.com/defenseunicorns/zarf/src/pkg/transform"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/pkg/utils/helpers"
	"github.com/defenseunicorns/zarf/src/types"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/mholt/archiver/v3"
)

// Create generates a Zarf package tarball for a given PackageConfig and optional base directory.
func (p *Packager) Create(baseDir string) error {

	var originalDir string

	if err := p.readYaml(filepath.Join(baseDir, config.ZarfYAML)); err != nil {
		return fmt.Errorf("unable to read the zarf.yaml file: %s", err.Error())
	}

	if helpers.IsOCIURL(p.cfg.CreateOpts.Output) {
		ref, err := oci.ReferenceFromMetadata(p.cfg.CreateOpts.Output, &p.cfg.Pkg.Metadata, p.arch)
		if err != nil {
			return err
		}
		err = p.SetOCIRemote(ref.String())
		if err != nil {
			return err
		}
	}

	// Load the images and repos from the 'reference' package
	if err := p.loadDifferentialData(); err != nil {
		return err
	}

	// Change the working directory if this run has an alternate base dir.
	if baseDir != "" {
		originalDir, _ = os.Getwd()
		if err := os.Chdir(baseDir); err != nil {
			return fmt.Errorf("unable to access directory '%s': %w", baseDir, err)
		}
		message.Note(fmt.Sprintf("Using build directory %s", baseDir))
	}

	if p.cfg.Pkg.Kind == "ZarfInitConfig" {
		p.cfg.Pkg.Metadata.Version = config.CLIVersion
		p.cfg.IsInitConfig = true
	}

	// Before we compose the components (and render the imported OCI components), we need to remove any components that are not needed for a differential build
	if err := p.removeDifferentialComponentsFromPackage(); err != nil {
		return err
	}

	// Compose components into a single zarf.yaml file
	if err := p.composeComponents(); err != nil {
		return err
	}

	// After components are composed, template the active package.
	if err := p.fillActiveTemplate(); err != nil {
		return fmt.Errorf("unable to fill values in template: %s", err.Error())
	}

	// After templates are filled process any create extensions
	if err := p.processExtensions(); err != nil {
		return err
	}

	// After we have a full zarf.yaml remove unnecessary repos and images if we are building a differential package
	if p.cfg.CreateOpts.DifferentialData.DifferentialPackagePath != "" {
		// Verify the package version of the package we're using as a 'reference' for the differential build is different than the package we're building
		// If the package versions are the same return an error
		if p.cfg.CreateOpts.DifferentialData.DifferentialPackageVersion == p.cfg.Pkg.Metadata.Version {
			return errors.New(lang.PkgCreateErrDifferentialSameVersion)
		}
		if p.cfg.CreateOpts.DifferentialData.DifferentialPackageVersion == "" || p.cfg.Pkg.Metadata.Version == "" {
			return fmt.Errorf("unable to build differential package when either the differential package version or the referenced package version is not set")
		}

		// Handle any potential differential images/repos before going forward
		if err := p.removeCopiesFromDifferentialPackage(); err != nil {
			return err
		}
	}

	// Perform early package validation.
	if err := validate.Run(p.cfg.Pkg); err != nil {
		return fmt.Errorf("unable to validate package: %w", err)
	}

	if !p.confirmAction(config.ZarfCreateStage, nil) {
		return fmt.Errorf("package creation canceled")
	}

	var combinedImageList []string
	componentSBOMs := map[string]*types.ComponentSBOM{}
	for idx, component := range p.cfg.Pkg.Components {
		onCreate := component.Actions.OnCreate
		onFailure := func() {
			if err := p.runActions(onCreate.Defaults, onCreate.OnFailure, nil); err != nil {
				message.Debugf("unable to run component failure action: %s", err.Error())
			}
		}
		isSkeleton := false
		err := p.addComponent(idx, component, isSkeleton)
		if err != nil {
			onFailure()
			return fmt.Errorf("unable to add component: %w", err)
		}
		componentSBOM, err := p.getFilesToSBOM(component)
		if err != nil {
			onFailure()
			return fmt.Errorf("unable to create component SBOM: %w", err)
		}

		if err := p.runActions(onCreate.Defaults, onCreate.OnSuccess, nil); err != nil {
			onFailure()
			return fmt.Errorf("unable to run component success action: %w", err)
		}

		if componentSBOM != nil && len(componentSBOM.Files) > 0 {
			componentSBOMs[component.Name] = componentSBOM
		}

		// Combine all component images into a single entry for efficient layer reuse.
		combinedImageList = append(combinedImageList, component.Images...)

		// Remove the temp directory for this component before archiving.
		err = os.RemoveAll(filepath.Join(p.tmp.Components, component.Name, types.TempFolder))
		if err != nil {
			message.Warnf("unable to remove temp directory for component %s, component tarball may contain unused artifacts: %s", component.Name, err.Error())
		}
	}

	imgList := helpers.Unique(combinedImageList)

	// Images are handled separately from other component assets.
	if len(imgList) > 0 {
		message.HeaderInfof("📦 PACKAGE IMAGES")

		doPull := func() error {
			imgConfig := images.ImgConfig{
				ImagesPath:        p.tmp.Images,
				ImgList:           imgList,
				Insecure:          config.CommonOptions.Insecure,
				Architectures:     []string{p.cfg.Pkg.Metadata.Architecture, p.cfg.Pkg.Build.Architecture},
				RegistryOverrides: p.cfg.CreateOpts.RegistryOverrides,
			}

			return imgConfig.PullAll()
		}

		if err := helpers.Retry(doPull, 3, 5*time.Second); err != nil {
			return fmt.Errorf("unable to pull images after 3 attempts: %w", err)
		}
	}

	// Ignore SBOM creation if there the flag is set.
	if p.cfg.CreateOpts.SkipSBOM {
		message.Debug("Skipping image SBOM processing per --skip-sbom flag")
	} else {
		if err := sbom.Catalog(componentSBOMs, imgList, p.tmp); err != nil {
			return fmt.Errorf("unable to create an SBOM catalog for the package: %w", err)
		}
	}

	// Process the component directories into compressed tarballs
	// NOTE: This is purposefully being done after the SBOM cataloging
	for _, component := range p.cfg.Pkg.Components {
		// Make the component a tar archive
		err := p.archiveComponent(component)
		if err != nil {
			return fmt.Errorf("unable to archive component: %s", err.Error())
		}
	}

	// In case the directory was changed, reset to prevent breaking relative target paths.
	if originalDir != "" {
		_ = os.Chdir(originalDir)
	}

	// Calculate all the checksums
	checksumChecksum, err := generatePackageChecksums(p.tmp.Base)
	if err != nil {
		return fmt.Errorf("unable to generate checksums for the package: %w", err)
	}
	p.cfg.Pkg.Metadata.AggregateChecksum = checksumChecksum

	// Save the transformed config.
	if err := p.writeYaml(); err != nil {
		return fmt.Errorf("unable to write zarf.yaml: %w", err)
	}

	// Sign the config file if a key was provided
	if p.cfg.CreateOpts.SigningKeyPath != "" {
		_, err := utils.CosignSignBlob(p.tmp.ZarfYaml, p.tmp.ZarfSig, p.cfg.CreateOpts.SigningKeyPath, p.getSigCreatePassword)
		if err != nil {
			return fmt.Errorf("unable to sign the package: %w", err)
		}
	}

	if helpers.IsOCIURL(p.cfg.CreateOpts.Output) {
		err := p.remote.PublishPackage(&p.cfg.Pkg, p.tmp.Base, config.CommonOptions.OCIConcurrency)
		if err != nil {
			return fmt.Errorf("unable to publish package: %w", err)
		}
	} else {
		// Use the output path if the user specified it.
		packageName := filepath.Join(p.cfg.CreateOpts.Output, p.GetPackageName())

		// Try to remove the package if it already exists.
		_ = os.Remove(packageName)

		// Create the package tarball.
		if err := p.archivePackage(p.tmp.Base, packageName); err != nil {
			return fmt.Errorf("unable to archive package: %w", err)
		}
	}

	// Output the SBOM files into a directory if specified.
	if p.cfg.CreateOpts.SBOMOutputDir != "" || p.cfg.CreateOpts.ViewSBOM {
		if err = archiver.Unarchive(p.tmp.SbomTar, p.tmp.Sboms); err != nil {
			return err
		}

		if p.cfg.CreateOpts.SBOMOutputDir != "" {
			if err := sbom.OutputSBOMFiles(p.tmp, p.cfg.CreateOpts.SBOMOutputDir, p.cfg.Pkg.Metadata.Name); err != nil {
				return err
			}
		}

		// Open a browser to view the SBOM if specified.
		if p.cfg.CreateOpts.ViewSBOM {
			sbom.ViewSBOMFiles(p.tmp)
		}
	}

	return nil
}

func (p *Packager) getFilesToSBOM(component types.ZarfComponent) (*types.ComponentSBOM, error) {
	componentPath, err := p.createOrGetComponentPaths(component)
	if err != nil {
		return nil, fmt.Errorf("unable to create the component paths: %s", err.Error())
	}

	// Create an struct to hold the SBOM information for this component.
	componentSBOM := types.ComponentSBOM{
		Files:         []string{},
		ComponentPath: componentPath,
	}

	appendSBOMFiles := func(path string) {
		if utils.IsDir(path) {
			files, _ := utils.RecursiveFileList(path, nil, false)
			componentSBOM.Files = append(componentSBOM.Files, files...)
		} else {
			componentSBOM.Files = append(componentSBOM.Files, path)
		}
	}

	for fileIdx, file := range component.Files {
		if file.Matrix != nil {
			m := reflect.ValueOf(*file.Matrix)
			for i := 0; i < m.NumField(); i++ {
				prefix := fmt.Sprintf("%d-%s", fileIdx, strings.Split(m.Type().Field(i).Tag.Get("json"), ",")[0])
				if options, ok := m.Field(i).Interface().(*types.ZarfFileOptions); ok && options != nil {
					path := filepath.Join(componentPath.Files, prefix, filepath.Base(options.Target))
					appendSBOMFiles(path)
				}
			}
		} else {
			path := filepath.Join(componentPath.Files, strconv.Itoa(fileIdx), filepath.Base(file.Target))
			appendSBOMFiles(path)
		}
	}

	for dataIdx, data := range component.DataInjections {
		path := filepath.Join(componentPath.DataInjections, strconv.Itoa(dataIdx), filepath.Base(data.Target.Path))

		appendSBOMFiles(path)
	}

	return &componentSBOM, nil
}

func (p *Packager) addComponent(index int, component types.ZarfComponent, isSkeleton bool) error {
	message.HeaderInfof("📦 %s COMPONENT", strings.ToUpper(component.Name))

	componentPath, err := p.createOrGetComponentPaths(component)
	if err != nil {
		return fmt.Errorf("unable to create the component paths: %s", err.Error())
	}

	if isSkeleton && component.CosignKeyPath != "" {
		dst := filepath.Join(componentPath.Base, "cosign.pub")
		err := utils.CreatePathAndCopy(component.CosignKeyPath, dst)
		if err != nil {
			return err
		}
		p.cfg.Pkg.Components[index].CosignKeyPath = "cosign.pub"
	}

	onCreate := component.Actions.OnCreate
	if !isSkeleton {
		if err := p.runActions(onCreate.Defaults, onCreate.Before, nil); err != nil {
			return fmt.Errorf("unable to run component before action: %w", err)
		}
	}

	// If any helm charts are defined, process them.
	for chartIdx, chart := range component.Charts {

		helmCfg := helm.Helm{
			Chart: chart,
			Cfg:   p.cfg,
		}

		if isSkeleton && chart.URL == "" {
			rel := filepath.Join(types.ChartsFolder, fmt.Sprintf("%s-%d", chart.Name, chartIdx))
			dst := filepath.Join(componentPath.Base, rel)

			err := utils.CreatePathAndCopy(chart.LocalPath, dst)
			if err != nil {
				return err
			}

			p.cfg.Pkg.Components[index].Charts[chartIdx].LocalPath = rel
		} else {
			err := helmCfg.PackageChart(componentPath.Charts)
			if err != nil {
				return err
			}
		}

		for valuesIdx, path := range chart.ValuesFiles {
			rel := fmt.Sprintf("%s-%d", helm.StandardName(types.ValuesFolder, chart), valuesIdx)
			dst := filepath.Join(componentPath.Base, rel)

			if helpers.IsURL(path) {
				if isSkeleton {
					continue
				}
				if err := utils.DownloadToFile(path, dst, component.CosignKeyPath); err != nil {
					return fmt.Errorf(lang.ErrDownloading, path, err.Error())
				}
			} else {
				if err := utils.CreatePathAndCopy(path, dst); err != nil {
					return fmt.Errorf("unable to copy chart values file %s: %w", path, err)
				}
				if isSkeleton {
					p.cfg.Pkg.Components[index].Charts[chartIdx].ValuesFiles[valuesIdx] = rel
				}
			}
		}
	}

	for fileIdx, file := range component.Files {
		message.Debugf("Loading %#v", file)

		mFiles := make(map[string]types.ZarfFile)
		if file.Matrix != nil {
			m := reflect.ValueOf(*file.Matrix)
			for i := 0; i < m.NumField(); i++ {
				prefix := fmt.Sprintf("%d-%s", fileIdx, strings.Split(m.Type().Field(i).Tag.Get("json"), ",")[0])
				if options, ok := m.Field(i).Interface().(*types.ZarfFileOptions); ok && options != nil {
					r := file
					r.Shasum = options.Shasum
					r.Source = options.Source
					r.Target = options.Target
					r.Symlinks = options.Symlinks

					mFiles[prefix] = r
				}
			}
		} else {
			mFiles[strconv.Itoa(fileIdx)] = file
		}

		for prefix, mFile := range mFiles {
			rel := filepath.Join(types.FilesFolder, prefix, filepath.Base(mFile.Target))
			dst := filepath.Join(componentPath.Base, rel)

			if helpers.IsURL(mFile.Source) {
				if isSkeleton {
					continue
				}
				if err := utils.DownloadToFile(mFile.Source, dst, component.CosignKeyPath); err != nil {
					return fmt.Errorf(lang.ErrDownloading, mFile.Source, err.Error())
				}
			} else {
				if err := utils.CreatePathAndCopy(mFile.Source, dst); err != nil {
					return fmt.Errorf("unable to copy file %s: %w", mFile.Source, err)
				}
				if isSkeleton {
					p.cfg.Pkg.Components[index].Files[fileIdx].Source = rel
				}
			}

			// Abort packaging on invalid shasum (if one is specified).
			if mFile.Shasum != "" {
				if actualShasum, _ := utils.GetCryptoHashFromFile(dst, crypto.SHA256); actualShasum != mFile.Shasum {
					return fmt.Errorf("shasum mismatch for file %s: expected %s, got %s", mFile.Source, mFile.Shasum, actualShasum)
				}
			}

			if mFile.Executable || utils.IsDir(dst) {
				_ = os.Chmod(dst, 0700)
			} else {
				_ = os.Chmod(dst, 0600)
			}
		}
	}

	if len(component.DataInjections) > 0 {
		spinner := message.NewProgressSpinner("Loading data injections")
		defer spinner.Stop()

		for dataIdx, data := range component.DataInjections {
			spinner.Updatef("Copying data injection %s for %s", data.Target.Path, data.Target.Selector)

			rel := filepath.Join(types.DataInjectionsFolder, strconv.Itoa(dataIdx), filepath.Base(data.Target.Path))
			dst := filepath.Join(componentPath.Base, rel)

			if helpers.IsURL(data.Source) {
				if isSkeleton {
					continue
				}
				if err := utils.DownloadToFile(data.Source, dst, component.CosignKeyPath); err != nil {
					return fmt.Errorf(lang.ErrDownloading, data.Source, err.Error())
				}
			} else {
				if err := utils.CreatePathAndCopy(data.Source, dst); err != nil {
					return fmt.Errorf("unable to copy data injection %s: %s", data.Source, err.Error())
				}
				if isSkeleton {
					p.cfg.Pkg.Components[index].DataInjections[dataIdx].Source = rel
				}
			}
		}
		spinner.Success()
	}

	if len(component.Manifests) > 0 {
		// Get the proper count of total manifests to add.
		manifestCount := 0

		for _, manifest := range component.Manifests {
			manifestCount += len(manifest.Files)
			manifestCount += len(manifest.Kustomizations)
		}

		spinner := message.NewProgressSpinner("Loading %d K8s manifests", manifestCount)
		defer spinner.Stop()

		// Iterate over all manifests.
		for manifestIdx, manifest := range component.Manifests {
			for fileIdx, path := range manifest.Files {
				rel := filepath.Join(types.ManifestsFolder, fmt.Sprintf("%s-%d.yaml", manifest.Name, fileIdx))
				dst := filepath.Join(componentPath.Base, rel)

				// Copy manifests without any processing.
				spinner.Updatef("Copying manifest %s", path)
				if helpers.IsURL(path) {
					if isSkeleton {
						continue
					}
					if err := utils.DownloadToFile(path, dst, component.CosignKeyPath); err != nil {
						return fmt.Errorf(lang.ErrDownloading, path, err.Error())
					}
				} else {
					if err := utils.CreatePathAndCopy(path, dst); err != nil {
						return fmt.Errorf("unable to copy manifest %s: %w", path, err)
					}
					if isSkeleton {
						p.cfg.Pkg.Components[index].Manifests[manifestIdx].Files[fileIdx] = rel
					}
				}
			}

			for kustomizeIdx, path := range manifest.Kustomizations {
				// Generate manifests from kustomizations and place in the package.
				spinner.Updatef("Building kustomization for %s", path)

				kname := fmt.Sprintf("kustomization-%s-%d.yaml", manifest.Name, kustomizeIdx)
				rel := filepath.Join(types.ManifestsFolder, kname)
				dst := filepath.Join(componentPath.Base, rel)

				if err := kustomize.Build(path, dst, manifest.KustomizeAllowAnyDirectory); err != nil {
					return fmt.Errorf("unable to build kustomization %s: %w", path, err)
				}
				if isSkeleton {
					p.cfg.Pkg.Components[index].Manifests[manifestIdx].Files = append(p.cfg.Pkg.Components[index].Manifests[manifestIdx].Files, rel)
				}
			}
			if isSkeleton {
				// remove kustomizations
				p.cfg.Pkg.Components[index].Manifests[manifestIdx].Kustomizations = nil
			}
		}
		spinner.Success()
	}

	// Load all specified git repos.
	if len(component.Repos) > 0 && !isSkeleton {
		spinner := message.NewProgressSpinner("Loading %d git repos", len(component.Repos))
		defer spinner.Stop()

		for _, url := range component.Repos {
			// Pull all the references if there is no `@` in the string.
			gitCfg := git.NewWithSpinner(p.cfg.State.GitServer, spinner)
			if err := gitCfg.Pull(url, componentPath.Repos, false); err != nil {
				return fmt.Errorf("unable to pull git repo %s: %w", url, err)
			}
		}
		spinner.Success()
	}

	if !isSkeleton {
		if err := p.runActions(onCreate.Defaults, onCreate.After, nil); err != nil {
			return fmt.Errorf("unable to run component after action: %w", err)
		}
	}

	return nil
}

// generateChecksum walks through all of the files starting at the base path and generates a checksum file.
// Each file within the basePath represents a layer within the Zarf package.
// generateChecksum returns a SHA256 checksum of the checksums.txt file.
func generatePackageChecksums(basePath string) (string, error) {
	var checksumsData string

	// Add a '/' or '\' to the basePath so that the checksums file lists paths from the perspective of the basePath
	basePathWithModifier := basePath + string(filepath.Separator)

	// Walk all files in the package path and calculate their checksums
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			sum, err := utils.GetSHA256OfFile(path)
			if err != nil {
				return err
			}
			checksumsData += fmt.Sprintf("%s %s\n", sum, strings.TrimPrefix(path, basePathWithModifier))
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	// Create the checksums file
	checksumsFilePath := filepath.Join(basePath, config.ZarfChecksumsTxt)
	if err := utils.WriteFile(checksumsFilePath, []byte(checksumsData)); err != nil {
		return "", err
	}

	// Calculate the checksum of the checksum file
	return utils.GetSHA256OfFile(checksumsFilePath)
}

// loadDifferentialData extracts the zarf config of a designated 'reference' package that we are building a differential over and creates a list of all images and repos that are in the reference package
func (p *Packager) loadDifferentialData() error {
	if p.cfg.CreateOpts.DifferentialData.DifferentialPackagePath == "" {
		return nil
	}

	// Save the fact that this is a differential build into the build data of the package
	p.cfg.Pkg.Build.Differential = true

	tmpDir, _ := utils.MakeTempDir("")
	defer os.RemoveAll(tmpDir)

	// Load the package spec of the package we're using as a 'reference' for the differential build
	if helpers.IsOCIURL(p.cfg.CreateOpts.DifferentialData.DifferentialPackagePath) {
		err := p.SetOCIRemote(p.cfg.CreateOpts.DifferentialData.DifferentialPackagePath)
		if err != nil {
			return err
		}
		manifest, err := p.remote.FetchRoot()
		if err != nil {
			return err
		}
		pkg, err := p.remote.FetchZarfYAML(manifest)
		if err != nil {
			return err
		}
		err = utils.WriteYaml(filepath.Join(tmpDir, config.ZarfYAML), pkg, 0600)
		if err != nil {
			return err
		}
	} else {
		if err := archiver.Extract(p.cfg.CreateOpts.DifferentialData.DifferentialPackagePath, config.ZarfYAML, tmpDir); err != nil {
			return fmt.Errorf("unable to extract the differential zarf package spec: %s", err.Error())
		}
	}

	var differentialZarfConfig types.ZarfPackage
	if err := utils.ReadYaml(filepath.Join(tmpDir, config.ZarfYAML), &differentialZarfConfig); err != nil {
		return fmt.Errorf("unable to load the differential zarf package spec: %s", err.Error())
	}

	// Generate a map of all the images and repos that are included in the provided package
	allIncludedImagesMap := map[string]bool{}
	allIncludedReposMap := map[string]bool{}
	for _, component := range differentialZarfConfig.Components {
		for _, image := range component.Images {
			allIncludedImagesMap[image] = true
		}
		for _, repo := range component.Repos {
			allIncludedReposMap[repo] = true
		}
	}

	p.cfg.CreateOpts.DifferentialData.DifferentialImages = allIncludedImagesMap
	p.cfg.CreateOpts.DifferentialData.DifferentialRepos = allIncludedReposMap
	p.cfg.CreateOpts.DifferentialData.DifferentialPackageVersion = differentialZarfConfig.Metadata.Version
	p.cfg.CreateOpts.DifferentialData.DifferentialOCIComponents = differentialZarfConfig.Build.OCIImportedComponents

	return nil
}

// removeDifferentialComponentsFromPackage will remove unchanged OCI imported components from a differential package creation
func (p *Packager) removeDifferentialComponentsFromPackage() error {
	// Remove components that were imported and already built into the reference package
	if len(p.cfg.CreateOpts.DifferentialData.DifferentialOCIComponents) > 0 {
		componentsToRemove := []int{}

		for idx, component := range p.cfg.Pkg.Components {
			// if the component is imported from an OCI package and everything is the same, don't include this package
			if helpers.IsOCIURL(component.Import.URL) {
				if _, alsoExists := p.cfg.CreateOpts.DifferentialData.DifferentialOCIComponents[component.Import.URL]; alsoExists {

					// If the component spec is not empty, we will still include it in the differential package
					// NOTE: We are ignoring fields that are not relevant to the differential build
					if component.IsEmpty([]string{"Name", "Required", "Description", "Default", "Import"}) {
						componentsToRemove = append(componentsToRemove, idx)
					}
				}
			}
		}

		// Remove the components that are already included (via OCI Import) in the reference package
		if len(componentsToRemove) > 0 {
			for i, componentIndex := range componentsToRemove {
				indexToRemove := componentIndex - i
				componentToRemove := p.cfg.Pkg.Components[indexToRemove]

				// If we are removing a component, add it to the build metadata and remove it from the list of OCI components for this package
				p.cfg.Pkg.Build.DifferentialMissing = append(p.cfg.Pkg.Build.DifferentialMissing, componentToRemove.Name)

				p.cfg.Pkg.Components = append(p.cfg.Pkg.Components[:indexToRemove], p.cfg.Pkg.Components[indexToRemove+1:]...)
			}
		}
	}

	return nil
}

// removeCopiesFromDifferentialPackage will remove any images and repos that are already included in the reference package from the new package
func (p *Packager) removeCopiesFromDifferentialPackage() error {
	// If a differential build was not requested, continue on as normal
	if p.cfg.CreateOpts.DifferentialData.DifferentialPackagePath == "" {
		return nil
	}

	// Loop through all of the components to determine if any of them are using already included images or repos
	componentMap := make(map[int]types.ZarfComponent)
	for idx, component := range p.cfg.Pkg.Components {
		newImageList := []string{}
		newRepoList := []string{}
		// Generate a list of all unique images for this component
		for _, img := range component.Images {
			// If a image doesn't have a tag (or is a commonly reused tag), we will include this image in the differential package
			imgRef, err := transform.ParseImageRef(img)
			if err != nil {
				return fmt.Errorf("unable to parse image ref %s: %s", img, err.Error())
			}

			// Only include new images or images that have a commonly overwritten tag
			imgTag := imgRef.TagOrDigest
			useImgAnyways := imgTag == ":latest" || imgTag == ":stable" || imgTag == ":nightly"
			if useImgAnyways || !p.cfg.CreateOpts.DifferentialData.DifferentialImages[img] {
				newImageList = append(newImageList, img)
			} else {
				message.Debugf("Image %s is already included in the differential package", img)
			}
		}

		// Generate a list of all unique repos for this component
		for _, repoURL := range component.Repos {
			// Split the remote url and the zarf reference
			_, refPlain, err := transform.GitURLSplitRef(repoURL)
			if err != nil {
				return err
			}

			var ref plumbing.ReferenceName
			// Parse the ref from the git URL.
			if refPlain != "" {
				ref = git.ParseRef(refPlain)
			}

			// Only include new repos or repos that were not referenced by a specific commit sha or tag
			useRepoAnyways := ref == "" || (!ref.IsTag() && !plumbing.IsHash(refPlain))
			if useRepoAnyways || !p.cfg.CreateOpts.DifferentialData.DifferentialRepos[repoURL] {
				newRepoList = append(newRepoList, repoURL)
			} else {
				message.Debugf("Repo %s is already included in the differential package", repoURL)
			}
		}

		// Update the component with the unique lists of repos and images
		component.Images = newImageList
		component.Repos = newRepoList
		componentMap[idx] = component
	}

	// Update the package with the new component list
	for idx, component := range componentMap {
		p.cfg.Pkg.Components[idx] = component
	}

	return nil
}

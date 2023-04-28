// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package packager contains functions for interacting with, managing and deploying Zarf packages.
package packager

import (
	"bytes"
	"crypto"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
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
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/types"
	"github.com/mholt/archiver/v3"
	"github.com/openvex/go-vex/pkg/vex"
)

// Create generates a Zarf package tarball for a given PackageConfig and optional base directory.
func (p *Packager) Create(baseDir string) error {
	var originalDir string

	// Change the working directory if this run has an alternate base dir.
	if baseDir != "" {
		originalDir, _ = os.Getwd()
		if err := os.Chdir(baseDir); err != nil {
			return fmt.Errorf("unable to access directory '%s': %w", baseDir, err)
		}
		message.Note(fmt.Sprintf("Using build directory %s", baseDir))
	}

	if err := p.readYaml(config.ZarfYAML, false); err != nil {
		return fmt.Errorf("unable to read the zarf.yaml file: %w", err)
	}

	if p.cfg.Pkg.Kind == "ZarfInitConfig" {
		p.cfg.Pkg.Metadata.Version = config.CLIVersion
		p.cfg.IsInitConfig = true
	}

	if err := p.composeComponents(); err != nil {
		return err
	}

	// After components are composed, template the active package.
	if err := p.fillActiveTemplate(); err != nil {
		return fmt.Errorf("unable to fill values in template: %s", err.Error())
	}

	// Create component paths and process extensions for each component.
	for i, c := range p.cfg.Pkg.Components {
		componentPath, err := p.createOrGetComponentPaths(c)
		if err != nil {
			return err
		}

		// Process any extensions.
		p.cfg.Pkg.Components[i], err = p.processExtensions(componentPath, c)
		if err != nil {
			return fmt.Errorf("unable to process extensions: %w", err)
		}
	}

	// Perform early package validation.
	if err := validate.Run(p.cfg.Pkg); err != nil {
		return fmt.Errorf("unable to validate package: %w", err)
	}

	if !p.confirmAction("Create", nil) {
		return fmt.Errorf("package creation canceled")
	}

	var combinedImageList []string
	componentSBOMs := map[string]*types.ComponentSBOM{}
	for _, component := range p.cfg.Pkg.Components {
		componentSBOM, err := p.addComponent(component)
		onCreate := component.Actions.OnCreate
		onFailure := func() {
			if err := p.runActions(onCreate.Defaults, onCreate.OnFailure, nil); err != nil {
				message.Debugf("unable to run component failure action: %s", err.Error())
			}
		}

		if err != nil {
			onFailure()
			return fmt.Errorf("unable to add component: %w", err)
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
		tmpPath := path.Join(p.getComponentBasePath(component), "temp")
		_ = os.RemoveAll(tmpPath)
	}

	imgList := utils.Unique(combinedImageList)

	// Images are handled separately from other component assets.
	if len(imgList) > 0 {
		message.HeaderInfof("📦 COMPONENT IMAGES")

		doPull := func() error {
			imgConfig := images.ImgConfig{
				ImagesPath:    p.tmp.Images,
				ImgList:       imgList,
				Insecure:      config.CommonOptions.Insecure,
				Architectures: []string{p.cfg.Pkg.Metadata.Architecture, p.cfg.Pkg.Build.Architecture},
			}

			return imgConfig.PullAll()
		}

		if err := utils.Retry(doPull, 3, 5*time.Second); err != nil {
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
		componentPaths, _ := p.createOrGetComponentPaths(component)
		componentName := fmt.Sprintf("%s.%s", component.Name, "tar")
		componentTarPath := filepath.Join(p.tmp.Components, componentName)
		if err := archiver.Archive([]string{componentPaths.Base}, componentTarPath); err != nil {
			return fmt.Errorf("unable to create package: %w", err)
		}

		// Remove the deflated component directory
		if err := os.RemoveAll(filepath.Join(p.tmp.Components, component.Name)); err != nil {
			return fmt.Errorf("unable to remove the component directory (%s): %w", componentPaths.Base, err)
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

	// Use the output path if the user specified it.
	packageName := filepath.Join(p.cfg.CreateOpts.OutputDirectory, p.GetPackageName())

	// Try to remove the package if it already exists.
	_ = os.Remove(packageName)

	// Make the archive
	archiveSrc := []string{p.tmp.Base + string(os.PathSeparator)}
	if err := archiver.Archive(archiveSrc, packageName); err != nil {
		return fmt.Errorf("unable to create package: %w", err)
	}

	f, err := os.Stat(packageName)
	if err != nil {
		return fmt.Errorf("unable to read the package archive: %w", err)
	}

	// Convert Megabytes to bytes.
	chunkSize := p.cfg.CreateOpts.MaxPackageSizeMB * 1000 * 1000

	// If a chunk size was specified and the package is larger than the chunk size, split it into chunks.
	if p.cfg.CreateOpts.MaxPackageSizeMB > 0 && f.Size() > int64(chunkSize) {
		chunks, sha256sum, err := utils.SplitFile(packageName, chunkSize)
		if err != nil {
			return fmt.Errorf("unable to split the package archive into multiple files: %w", err)
		}
		if len(chunks) > 999 {
			return fmt.Errorf("unable to split the package archive into multiple files: must be less than 1,000 files")
		}

		message.Infof("Package split into %d files, original sha256sum is %s", len(chunks)+1, sha256sum)
		_ = os.RemoveAll(packageName)

		// Marshal the data into a json file.
		jsonData, err := json.Marshal(types.ZarfPartialPackageData{
			Count:     len(chunks),
			Bytes:     f.Size(),
			Sha256Sum: sha256sum,
		})
		if err != nil {
			return fmt.Errorf("unable to marshal the partial package data: %w", err)
		}

		// Prepend the json data to the first chunk.
		chunks = append([][]byte{jsonData}, chunks...)

		for idx, chunk := range chunks {
			path := fmt.Sprintf("%s.part%03d", packageName, idx)
			if err := os.WriteFile(path, chunk, 0644); err != nil {
				return fmt.Errorf("unable to write the file %s: %w", path, err)
			}
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

func (p *Packager) addComponent(component types.ZarfComponent) (*types.ComponentSBOM, error) {
	message.HeaderInfof("📦 %s COMPONENT", strings.ToUpper(component.Name))

	componentPath, err := p.createOrGetComponentPaths(component)
	if err != nil {
		return nil, fmt.Errorf("unable to create the component paths: %w", err)
	}

	// Create an struct to hold the SBOM information for this component.
	componentSBOM := types.ComponentSBOM{
		Files:         []string{},
		ComponentPath: componentPath,
	}

	onCreate := component.Actions.OnCreate

	if err := p.runActions(onCreate.Defaults, onCreate.Before, nil); err != nil {
		return nil, fmt.Errorf("unable to run component before action: %w", err)
	}

	// If any helm charts are defined, process them.
	if len(component.Charts) > 0 {
		_ = utils.CreateDirectory(componentPath.Charts, 0700)
		_ = utils.CreateDirectory(componentPath.Values, 0700)
		re := regexp.MustCompile(`\.git$`)

		for _, chart := range component.Charts {
			isGitURL := re.MatchString(chart.URL)
			helmCfg := helm.Helm{
				Chart: chart,
				Cfg:   p.cfg,
			}

			if isGitURL {
				_, err = helmCfg.PackageChartFromGit(componentPath.Charts)
				if err != nil {
					return nil, fmt.Errorf("error creating chart archive, unable to pull the chart from git: %s", err.Error())
				}
			} else if len(chart.URL) > 0 {
				helmCfg.DownloadPublishedChart(componentPath.Charts)
			} else {
				path := helmCfg.PackageChartFromLocalFiles(componentPath.Charts)
				zarfFilename := fmt.Sprintf("%s-%s.tgz", chart.Name, chart.Version)
				if !strings.HasSuffix(path, zarfFilename) {
					return nil, fmt.Errorf("error creating chart archive, user provided chart name and/or version does not match given chart")
				}
			}

			for idx, path := range chart.ValuesFiles {
				dst := helm.StandardName(componentPath.Values, chart) + "-" + strconv.Itoa(idx)
				if utils.IsURL(path) {
					if err := utils.DownloadToFile(path, dst, component.CosignKeyPath); err != nil {
						return nil, fmt.Errorf(lang.ErrDownloading, path, err.Error())
					}
				} else {
					if err := utils.CreatePathAndCopy(path, dst); err != nil {
						return nil, fmt.Errorf("unable to copy chart values file %s: %w", path, err)
					}
				}
			}
		}
	}

	if len(component.Files) > 0 {
		if err = utils.CreateDirectory(componentPath.Files, 0700); err != nil {
			return nil, fmt.Errorf("unable to create the component files directory: %s", err.Error())
		}

		for index, file := range component.Files {
			message.Debugf("Loading %#v", file)
			dst := filepath.Join(componentPath.Files, strconv.Itoa(index))

			if utils.IsURL(file.Source) {
				if err := utils.DownloadToFile(file.Source, dst, component.CosignKeyPath); err != nil {
					return nil, fmt.Errorf(lang.ErrDownloading, file.Source, err.Error())
				}
			} else {
				if err := utils.CreatePathAndCopy(file.Source, dst); err != nil {
					return nil, fmt.Errorf("unable to copy file %s: %w", file.Source, err)
				}
			}

			// Abort packaging on invalid shasum (if one is specified).
			if file.Shasum != "" {
				if actualShasum, _ := utils.GetCryptoHash(dst, crypto.SHA256); actualShasum != file.Shasum {
					return nil, fmt.Errorf("shasum mismatch for file %s: expected %s, got %s", file.Source, file.Shasum, actualShasum)
				}
			}

			info, _ := os.Stat(dst)

			if file.Executable || info.IsDir() {
				_ = os.Chmod(dst, 0700)
			} else {
				_ = os.Chmod(dst, 0600)
			}

			componentSBOM.Files = append(componentSBOM.Files, dst)
		}
	}

	if len(component.DataInjections) > 0 {
		spinner := message.NewProgressSpinner("Loading data injections")
		defer spinner.Success()

		for _, data := range component.DataInjections {
			spinner.Updatef("Copying data injection %s for %s", data.Target.Path, data.Target.Selector)
			dst := filepath.Join(componentPath.DataInjections, filepath.Base(data.Target.Path))
			if utils.IsURL(data.Source) {
				if err := utils.DownloadToFile(data.Source, dst, component.CosignKeyPath); err != nil {
					return nil, fmt.Errorf(lang.ErrDownloading, data.Source, err.Error())
				}
			} else {
				if err := utils.CreatePathAndCopy(data.Source, dst); err != nil {
					return nil, fmt.Errorf("unable to copy data injection %s: %s", data.Source, err.Error())
				}
			}

			// Unwrap the dataInjection dir into individual files.
			pattern := regexp.MustCompile(`(?mi).+$`)
			files, _ := utils.RecursiveFileList(dst, pattern, false, true)
			componentSBOM.Files = append(componentSBOM.Files, files...)
		}
	}

	if len(component.Manifests) > 0 {
		if err := utils.CreateDirectory(componentPath.Manifests, 0700); err != nil {
			return nil, fmt.Errorf("unable to create manifest directory %s: %s", componentPath.Manifests, err.Error())
		}
		// Get the proper count of total manifests to add.
		manifestCount := 0

		for _, manifest := range component.Manifests {
			manifestCount += len(manifest.Files)
			manifestCount += len(manifest.Kustomizations)
		}

		spinner := message.NewProgressSpinner("Loading %d K8s manifests", manifestCount)
		defer spinner.Success()

		// Iterate over all manifests.
		for _, manifest := range component.Manifests {
			for idx, f := range manifest.Files {
				var trimmedPath string
				var destination string
				// Copy manifests without any processing.
				spinner.Updatef("Copying manifest %s", f)
				if utils.IsURL(f) {
					mname := fmt.Sprintf("manifest-%s-%d.yaml", manifest.Name, idx)
					destination = filepath.Join(componentPath.Manifests, mname)
					if err := utils.DownloadToFile(f, destination, component.CosignKeyPath); err != nil {
						return nil, fmt.Errorf(lang.ErrDownloading, f, err.Error())
					}
					// Update the manifest path to the new location.
					manifest.Files[idx] = mname
				} else {
					// If using a temp directory, trim the temp directory from the path.
					trimmedPath = strings.TrimPrefix(f, componentPath.Temp)
					destination = filepath.Join(componentPath.Manifests, trimmedPath)
					if err := utils.CreatePathAndCopy(f, destination); err != nil {
						return nil, fmt.Errorf("unable to copy manifest %s: %w", f, err)
					}
					// Update the manifest path to the new location.
					manifest.Files[idx] = trimmedPath
				}
			}

			for idx, k := range manifest.Kustomizations {
				// Generate manifests from kustomizations and place in the package.
				spinner.Updatef("Building kustomization for %s", k)
				kname := fmt.Sprintf("kustomization-%s-%d.yaml", manifest.Name, idx)
				destination := filepath.Join(componentPath.Manifests, kname)
				if err := kustomize.Build(k, destination, manifest.KustomizeAllowAnyDirectory); err != nil {
					return nil, fmt.Errorf("unable to build kustomization %s: %w", k, err)
				}
			}
		}
	}

	// Load all specified git repos.
	if len(component.Repos) > 0 {
		spinner := message.NewProgressSpinner("Loading %d git repos", len(component.Repos))
		defer spinner.Success()

		for _, url := range component.Repos {
			// Pull all the references if there is no `@` in the string.
			gitCfg := git.NewWithSpinner(p.cfg.State.GitServer, spinner)
			if err := gitCfg.Pull(url, componentPath.Repos); err != nil {
				return nil, fmt.Errorf("unable to pull git repo %s: %w", url, err)
			}
		}
	}

	if len(component.Reports) > 0 {
		spinner := message.NewProgressSpinner("Loading %d reports", len(component.Reports))
		defer spinner.Success()

		err = utils.CreateDirectory(componentPath.Reports, 0700)
		if err != nil {
			return nil, fmt.Errorf("unable to create reports destination directory: %w", err)
		}

		for _, report := range component.Reports {
			var path string
			switch reportType := strings.ToLower(report.Type); reportType {
			case "vex":
				dst := fmt.Sprintf("%s/%s", componentPath.Reports, reportType)
				err = utils.CreateDirectory(dst, 0700)
				if err != nil {
					return nil, fmt.Errorf("unable to create reports destination directory: %w", err)
				}
				message.Debug("Source was identified as a URL")
				if _, err := os.Stat(dst + report.Name); err == nil {
					return nil, fmt.Errorf("%s already exists for this component and cannot conflict", dst+report.Name)
				}
				if utils.IsURL(report.Source) {
					path = fmt.Sprintf("%s/%s/%s", componentPath.Reports, reportType, report.Name)
					if err := utils.DownloadToFile(report.Source, path, component.CosignKeyPath); err != nil {
						return nil, fmt.Errorf(lang.ErrDownloading, report.Source, err.Error())
					}
				} else {
					path = report.Source
				}

				doc, err := vex.Load(path)
				if err != nil {
					return nil, fmt.Errorf("unable to load vex document %s from %s: %w", report.Name, path, err)
				}

				// Convert to JSON
				var b bytes.Buffer
				if err = doc.ToJSON(&b); err != nil {
					return nil, fmt.Errorf("unable to write vex doc to JSON for %s: %w", report.Name, err)
				}

				message.Debugf("Loaded VEX file %s (%d bytes)", report.Name, b.Len())

				// Write VEX file to the vex directory
				dest := fmt.Sprintf("%s/%s/%s", componentPath.Reports, reportType, report.Name)
				if err = utils.WriteFile(dest, b.Bytes()); err != nil {
					return nil, fmt.Errorf("unable to write vex file to %s: %w", dest, err)
				}
			default:
				message.Debugf("Loaded file %s)", reportType)

				dst := filepath.Join(componentPath.Reports, reportType, report.Name)

				if utils.IsURL(report.Source) {
					if err := utils.DownloadToFile(report.Source, dst, component.CosignKeyPath); err != nil {
						return nil, fmt.Errorf(lang.ErrDownloading, report.Source, err.Error())
					}
				} else {
					if err := utils.CreatePathAndCopy(report.Source, dst); err != nil {
						return nil, fmt.Errorf("unable to copy file %s: %w", report.Source, err)
					}
				}
			}
		}
	}

	if err := p.runActions(onCreate.Defaults, onCreate.After, nil); err != nil {
		return nil, fmt.Errorf("unable to run component after action: %w", err)
	}

	return &componentSBOM, nil
}

// generateChecksum walks through all of the files starting at the base path and generates a checksum file.
// Each file within the basePath represents a layer within the Zarf package.
// generateChecksum returns a SHA256 checksum of the checksums.txt file.
func generatePackageChecksums(basePath string) (string, error) {
	var checksumsData string

	// Add a '/' or '\' to the basePath so that the checksums file lists paths from the perspective of the basePath
	basePathWithModifier := basePath + string(filepath.Separator)

	// Walk through all files in the package path and calculate their checksums
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
	checksumsFilePath := filepath.Join(basePath, "checksums.txt")
	if err := utils.WriteFile(checksumsFilePath, []byte(checksumsData)); err != nil {
		return "", err
	}

	// Calculate the checksum of the checksum file
	return utils.GetSHA256OfFile(checksumsFilePath)
}

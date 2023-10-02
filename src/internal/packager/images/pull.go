// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package images provides functions for building and pushing images.
package images

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/transform"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/logs"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/cache"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"github.com/moby/moby/client"
	"github.com/pterm/pterm"
)

// PullAll pulls all of the images in the provided ref map.
func (i *ImageConfig) PullAll() error {
	var (
		longer            string
		imageCount        = len(i.ImageList)
		refInfoToImage    = map[transform.Image]v1.Image{}
		referenceToDigest = make(map[string]string)
	)

	type imgInfo struct {
		refInfo transform.Image
		img     v1.Image
	}

	type digestInfo struct {
		refInfo transform.Image
		digest  string
	}

	// Give some additional user feedback on larger image sets
	if imageCount > 15 {
		longer = "This step may take a couple of minutes to complete."
	} else if imageCount > 5 {
		longer = "This step may take several seconds to complete."
	}

	spinner := message.NewProgressSpinner("Loading metadata for %d images. %s", imageCount, longer)
	defer spinner.Stop()

	logs.Warn.SetOutput(&message.DebugWriter{})
	logs.Progress.SetOutput(&message.DebugWriter{})

	metadataImageConcurrency := utils.NewConcurrencyTools[imgInfo, error](len(i.ImageList))

	defer metadataImageConcurrency.Cancel()

	spinner.Updatef("Fetching image metadata (0 of %d)", len(i.ImageList))

	// Spawn a goroutine for each image to load its metadata
	for _, refInfo := range i.ImageList {
		// Create a closure so that we can pass the src into the goroutine
		refInfo := refInfo
		go func() {
			// Make sure to call Done() on the WaitGroup when the goroutine finishes
			defer metadataImageConcurrency.WaitGroupDone()

			if metadataImageConcurrency.IsDone() {
				return
			}

			actualSrc := refInfo.Reference
			if overrideHost, present := i.RegistryOverrides[refInfo.Host]; present {
				var err error
				actualSrc, err = transform.ImageTransformHostWithoutChecksum(overrideHost, refInfo.Reference)
				if err != nil {
					metadataImageConcurrency.ErrorChan <- fmt.Errorf("failed to swap override host %s for %s: %w", overrideHost, refInfo.Reference, err)
					return
				}
			}

			if metadataImageConcurrency.IsDone() {
				return
			}

			img, err := i.PullImage(actualSrc, spinner)
			if err != nil {
				metadataImageConcurrency.ErrorChan <- fmt.Errorf("failed to pull image %s: %w", actualSrc, err)
				return
			}

			if metadataImageConcurrency.IsDone() {
				return
			}

			metadataImageConcurrency.ProgressChan <- imgInfo{refInfo: refInfo, img: img}
		}()
	}

	onMetadataProgress := func(finishedImage imgInfo, iteration int) {
		spinner.Updatef("Fetching image metadata (%d of %d): %s", iteration+1, len(i.ImageList), finishedImage.refInfo.Reference)
		refInfoToImage[finishedImage.refInfo] = finishedImage.img
	}

	onMetadataError := func(err error) error {
		return fmt.Errorf("failed to load metadata for all images. This may be due to a network error or an invalid image reference: %w", err)
	}

	if err := metadataImageConcurrency.WaitWithProgress(onMetadataProgress, onMetadataError); err != nil {
		return err
	}

	// Create the ImagePath directory
	err := os.Mkdir(i.ImagesPath, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("failed to create image path %s: %w", i.ImagesPath, err)
	}

	totalBytes := int64(0)
	processedLayers := make(map[string]v1.Layer)
	for refInfo, img := range refInfoToImage {
		// Get the byte size for this image
		layers, err := img.Layers()
		if err != nil {
			return fmt.Errorf("unable to get layers for image %s: %w", refInfo.Reference, err)
		}
		for _, layer := range layers {
			layerDigest, err := layer.Digest()
			if err != nil {
				return fmt.Errorf("unable to get digest for image layer: %w", err)
			}

			// Only calculate this layer size if we haven't already looked at it
			if _, ok := processedLayers[layerDigest.Hex]; !ok {
				size, err := layer.Size()
				if err != nil {
					return fmt.Errorf("unable to get size of layer: %w", err)
				}
				totalBytes += size
				processedLayers[layerDigest.Hex] = layer
			}

		}
	}
	spinner.Updatef("Preparing image sources and cache for image pulling")

	// Create special sauce crane Path object
	// If it already exists use it
	cranePath, err := layout.FromPath(i.ImagesPath)
	// Use crane pattern for creating OCI layout if it doesn't exist
	if err != nil {
		// If it doesn't exist create it
		cranePath, err = layout.Write(i.ImagesPath, empty.Index)
		if err != nil {
			return err
		}
	}

	for refInfo, img := range refInfoToImage {
		imgDigest, err := img.Digest()
		if err != nil {
			return fmt.Errorf("unable to get digest for image %s: %w", refInfo.Reference, err)
		}
		referenceToDigest[refInfo.Reference] = imgDigest.String()
	}

	spinner.Success()

	// Create a thread to update a progress bar as we save the image files to disk
	doneSaving := make(chan int)
	var progressBarWaitGroup sync.WaitGroup
	progressBarWaitGroup.Add(1)
	go utils.RenderProgressBarForLocalDirWrite(i.ImagesPath, totalBytes, &progressBarWaitGroup, doneSaving, fmt.Sprintf("Pulling %d images", imageCount))

	// Spawn a goroutine for each layer to write it to disk using crane

	layerWritingConcurrency := utils.NewConcurrencyTools[bool, error](len(processedLayers))

	defer layerWritingConcurrency.Cancel()

	for _, layer := range processedLayers {
		layer := layer
		// Function is a combination of https://github.com/google/go-containerregistry/blob/v0.15.2/pkg/v1/layout/write.go#L270-L305
		// and https://github.com/google/go-containerregistry/blob/v0.15.2/pkg/v1/layout/write.go#L198-L262
		// with modifications. This allows us to dedupe layers for all images and write them concurrently.
		go func() {
			defer layerWritingConcurrency.WaitGroupDone()
			digest, err := layer.Digest()
			if errors.Is(err, stream.ErrNotComputed) {
				// Allow digest errors, since streams may not have calculated the hash
				// yet. Instead, use an empty value, which will be transformed into a
				// random file name with `os.CreateTemp` and the final digest will be
				// calculated after writing to a temp file and before renaming to the
				// final path.
				digest = v1.Hash{Algorithm: "sha256", Hex: ""}
			} else if err != nil {
				layerWritingConcurrency.ErrorChan <- err
				return
			}

			size, err := layer.Size()
			if errors.Is(err, stream.ErrNotComputed) {
				// Allow size errors, since streams may not have calculated the size
				// yet. Instead, use -1 as a sentinel value meaning that no size
				// comparison can be done and any sized blob file should be considered
				// valid and not overwritten.
				//
				// TODO: Provide an option to always overwrite blobs.
				size = -1
			} else if err != nil {
				layerWritingConcurrency.ErrorChan <- err
				return
			}

			if layerWritingConcurrency.IsDone() {
				return
			}

			readCloser, err := layer.Compressed()
			if err != nil {
				layerWritingConcurrency.ErrorChan <- err
				return
			}

			// Create the directory for the blob if it doesn't exist
			dir := filepath.Join(string(cranePath), "blobs", digest.Algorithm)
			if err := utils.CreateDirectory(dir, os.ModePerm); err != nil {
				layerWritingConcurrency.ErrorChan <- err
				return
			}

			if layerWritingConcurrency.IsDone() {
				return
			}

			// Check if blob already exists and is the correct size
			file := filepath.Join(dir, digest.Hex)
			if s, err := os.Stat(file); err == nil && !s.IsDir() && (s.Size() == size || size == -1) {
				layerWritingConcurrency.ProgressChan <- true
				return
			}

			if layerWritingConcurrency.IsDone() {
				return
			}

			// Write to a temporary file
			w, err := os.CreateTemp(dir, digest.Hex)
			if err != nil {
				layerWritingConcurrency.ErrorChan <- err
				return
			}
			// Delete temp file if an error is encountered before renaming
			defer func() {
				if err := os.Remove(w.Name()); err != nil && !errors.Is(err, os.ErrNotExist) {
					message.Warnf("error removing temporary file after encountering an error while writing blob: %v", err)
				}
			}()

			defer w.Close()

			if layerWritingConcurrency.IsDone() {
				return
			}

			// Write to file rename
			if n, err := io.Copy(w, readCloser); err != nil {
				layerWritingConcurrency.ErrorChan <- err
				return
			} else if size != -1 && n != size {
				layerWritingConcurrency.ErrorChan <- fmt.Errorf("expected blob size %d, but only wrote %d", size, n)
				return
			}

			if layerWritingConcurrency.IsDone() {
				return
			}

			// Always close reader before renaming, since Close computes the digest in
			// the case of streaming layers. If Close is not called explicitly, it will
			// occur in a goroutine that is not guaranteed to succeed before renamer is
			// called. When renamer is the layer's Digest method, it can return
			// ErrNotComputed.
			if err := readCloser.Close(); err != nil {
				layerWritingConcurrency.ErrorChan <- err
				return
			}

			// Always close file before renaming
			if err := w.Close(); err != nil {
				layerWritingConcurrency.ErrorChan <- err
				return
			}

			// Rename file based on the final hash
			renamePath := filepath.Join(string(cranePath), "blobs", digest.Algorithm, digest.Hex)
			os.Rename(w.Name(), renamePath)

			if layerWritingConcurrency.IsDone() {
				return
			}

			layerWritingConcurrency.ProgressChan <- true
		}()
	}

	onLayerWritingError := func(err error) error {
		// Send a signal to the progress bar that we're done and wait for the thread to finish
		doneSaving <- 1
		progressBarWaitGroup.Wait()
		message.WarnErr(err, "Failed to write image layers, trying again up to 3 times...")
		if strings.HasPrefix(err.Error(), "expected blob size") {
			message.Warnf("Potential image cache corruption: %s - try clearing cache with \"zarf tools clear-cache\"", err.Error())
		}
		return err
	}

	if err := layerWritingConcurrency.WaitWithoutProgress(onLayerWritingError); err != nil {
		return err
	}

	imageSavingConcurrency := utils.NewConcurrencyTools[digestInfo, error](len(refInfoToImage))

	defer imageSavingConcurrency.Cancel()

	// Spawn a goroutine for each image to write it's config and manifest to disk using crane
	// All layers should already be in place so this should be extremely fast
	for refInfo, img := range refInfoToImage {
		// Create a closure so that we can pass the refInfo and img into the goroutine
		refInfo, img := refInfo, img
		go func() {
			// Make sure to call Done() on the WaitGroup when the goroutine finishes
			defer imageSavingConcurrency.WaitGroupDone()

			// Save the image via crane
			err := cranePath.WriteImage(img)

			if imageSavingConcurrency.IsDone() {
				return
			}

			if err != nil {
				// Check if the cache has been invalidated, and warn the user if so
				if strings.HasPrefix(err.Error(), "error writing layer: expected blob size") {
					message.Warnf("Potential image cache corruption: %s - try clearing cache with \"zarf tools clear-cache\"", err.Error())
				}
				imageSavingConcurrency.ErrorChan <- fmt.Errorf("error when trying to save the img (%s): %w", refInfo.Reference, err)
				return
			}

			if imageSavingConcurrency.IsDone() {
				return
			}

			// Get the image digest so we can set an annotation in the image.json later
			imgDigest, err := img.Digest()
			if err != nil {
				imageSavingConcurrency.ErrorChan <- err
				return
			}

			if imageSavingConcurrency.IsDone() {
				return
			}

			imageSavingConcurrency.ProgressChan <- digestInfo{digest: imgDigest.String(), refInfo: refInfo}
		}()
	}

	onImageSavingProgress := func(finishedImage digestInfo, iteration int) {
		referenceToDigest[finishedImage.refInfo.Reference] = finishedImage.digest
	}

	onImageSavingError := func(err error) error {
		// Send a signal to the progress bar that we're done and wait for the thread to finish
		doneSaving <- 1
		progressBarWaitGroup.Wait()
		message.WarnErr(err, "Failed to write image config or manifest, trying again up to 3 times...")
		return err
	}

	if err := imageSavingConcurrency.WaitWithProgress(onImageSavingProgress, onImageSavingError); err != nil {
		return err
	}

	// for every image sequentially append OCI descriptor

	for refInfo, img := range refInfoToImage {
		desc, err := partial.Descriptor(img)
		if err != nil {
			return err
		}

		cranePath.AppendDescriptor(*desc)
		if err != nil {
			return err
		}

		imgDigest, err := img.Digest()
		if err != nil {
			return err
		}

		referenceToDigest[refInfo.Reference] = imgDigest.String()
	}

	if err := utils.AddImageNameAnnotation(i.ImagesPath, referenceToDigest); err != nil {
		return fmt.Errorf("unable to format OCI layout: %w", err)
	}

	// Send a signal to the progress bar that we're done and wait for the thread to finish
	doneSaving <- 1
	progressBarWaitGroup.Wait()

	return err
}

// PullImage returns a v1.Image either by loading a local tarball or the wider internet.
func (i *ImageConfig) PullImage(src string, spinner *message.Spinner) (img v1.Image, err error) {
	// Load image tarballs from the local filesystem.
	if strings.HasSuffix(src, ".tar") || strings.HasSuffix(src, ".tar.gz") || strings.HasSuffix(src, ".tgz") {
		spinner.Updatef("Reading image tarball: %s", src)
		return crane.Load(src, config.GetCraneOptions(true, i.Architectures...)...)
	}

	// If crane is unable to pull the image, try to load it from the local docker daemon.
	if _, err := crane.Manifest(src, config.GetCraneOptions(i.Insecure, i.Architectures...)...); err != nil {
		message.Debugf("crane unable to pull image %s: %s", src, err)
		spinner.Updatef("Falling back to docker for %s. This may take some time.", src)

		// Parse the image reference to get the image name.
		reference, err := name.ParseReference(src)
		if err != nil {
			return nil, fmt.Errorf("failed to parse image reference %s: %w", src, err)
		}

		// Attempt to connect to the local docker daemon.
		ctx := context.TODO()
		cli, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			return nil, fmt.Errorf("docker not available: %w", err)
		}
		cli.NegotiateAPIVersion(ctx)

		// Inspect the image to get the size.
		rawImg, _, err := cli.ImageInspectWithRaw(ctx, src)
		if err != nil {
			return nil, fmt.Errorf("failed to inspect image %s via docker: %w", src, err)
		}

		// Warn the user if the image is large.
		if rawImg.Size > 750*1000*1000 {
			warn := pterm.DefaultParagraph.WithMaxWidth(message.TermWidth).Sprintf("%s is %s and may take a very long time to load via docker. "+
				"See https://docs.zarf.dev/docs/faq for suggestions on how to improve large local image loading operations.",
				src, utils.ByteFormat(float64(rawImg.Size), 2))
			spinner.Warnf(warn)
		}

		// Use unbuffered opener to avoid OOM Kill issues https://github.com/defenseunicorns/zarf/issues/1214.
		// This will also take for ever to load large images.
		if img, err = daemon.Image(reference, daemon.WithUnbufferedOpener()); err != nil {
			return nil, fmt.Errorf("failed to load image %s from docker daemon: %w", src, err)
		}

		// The pull from the docker daemon was successful, return the image.
		return img, err
	}

	// Manifest was found, so use crane to pull the image.
	if img, err = crane.Pull(src, config.GetCraneOptions(i.Insecure, i.Architectures...)...); err != nil {
		return nil, fmt.Errorf("failed to pull image %s: %w", src, err)
	}

	spinner.Updatef("Preparing image %s", src)
	imageCachePath := filepath.Join(config.GetAbsCachePath(), config.ZarfImageCacheDir)
	img = cache.Image(img, cache.NewFilesystemCache(imageCachePath))

	return img, nil
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package layout contains functions for interacting with Jackal's package layout on disk.
package layout

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

func TestPackageFiles(t *testing.T) {
	t.Parallel()

	t.Run("Verify New()", func(t *testing.T) {
		t.Parallel()

		pp := New("test")

		raw := &PackagePaths{
			Base:       "test",
			JackalYAML: normalizePath("test/jackal.yaml"),
			Checksums:  normalizePath("test/checksums.txt"),
			Components: Components{
				Base: normalizePath("test/components"),
			},
		}
		require.Equal(t, raw, pp)
	})

	t.Run("Verify Files()", func(t *testing.T) {
		t.Parallel()

		pp := New("test")

		files := pp.Files()
		expected := map[string]string{
			"jackal.yaml":   normalizePath("test/jackal.yaml"),
			"checksums.txt": normalizePath("test/checksums.txt"),
		}
		require.Equal(t, expected, files)
	})

	t.Run("Verify Files() with signature", func(t *testing.T) {
		t.Parallel()

		pp := New("test")
		pp.Signature = filepath.Join(pp.Base, Signature)

		files := pp.Files()
		expected := map[string]string{
			"jackal.yaml":     normalizePath("test/jackal.yaml"),
			"checksums.txt":   normalizePath("test/checksums.txt"),
			"jackal.yaml.sig": normalizePath("test/jackal.yaml.sig"),
		}
		require.Equal(t, expected, files)
	})

	t.Run("Verify Files() with images", func(t *testing.T) {
		t.Parallel()

		pp := New("test")
		pp = pp.AddImages()

		files := pp.Files()
		expected := map[string]string{
			"jackal.yaml":       normalizePath("test/jackal.yaml"),
			"checksums.txt":     normalizePath("test/checksums.txt"),
			"images/index.json": normalizePath("test/images/index.json"),
			"images/oci-layout": normalizePath("test/images/oci-layout"),
		}
		require.Equal(t, expected, files)
	})

	// AddSBOMs sets the SBOMs path, so Files() should not return new files.
	t.Run("Verify Files() with SBOMs", func(t *testing.T) {
		t.Parallel()

		pp := New("test")
		pp = pp.AddSBOMs()

		files := pp.Files()
		expected := map[string]string{
			"jackal.yaml":   normalizePath("test/jackal.yaml"),
			"checksums.txt": normalizePath("test/checksums.txt"),
		}
		require.Equal(t, expected, files)

		pp.SBOMs.Path = normalizePath("test/sboms.tar")
		files = pp.Files()
		expected = map[string]string{
			"jackal.yaml":   normalizePath("test/jackal.yaml"),
			"checksums.txt": normalizePath("test/checksums.txt"),
			"sboms.tar":     normalizePath("test/sboms.tar"),
		}
		require.Equal(t, expected, files)
	})

	t.Run("Verify Files() with paths mapped to package paths", func(t *testing.T) {
		t.Parallel()

		pp := New("test")

		paths := []string{
			"jackal.yaml",
			"checksums.txt",
			"sboms.tar",
			normalizePath("components/c1.tar"),
			normalizePath("images/index.json"),
			normalizePath("images/oci-layout"),
			normalizePath("images/blobs/sha256/" + strings.Repeat("1", 64)),
		}
		pp.SetFromPaths(paths)

		files := pp.Files()
		expected := map[string]string{
			"jackal.yaml":       normalizePath("test/jackal.yaml"),
			"checksums.txt":     normalizePath("test/checksums.txt"),
			"sboms.tar":         normalizePath("test/sboms.tar"),
			"components/c1.tar": normalizePath("test/components/c1.tar"),
			"images/index.json": normalizePath("test/images/index.json"),
			"images/oci-layout": normalizePath("test/images/oci-layout"),
			"images/blobs/sha256/" + strings.Repeat("1", 64): normalizePath("test/images/blobs/sha256/" + strings.Repeat("1", 64)),
		}

		require.Len(t, pp.Images.Blobs, 1)
		require.Equal(t, expected, files)
	})

	t.Run("Verify Files() with image layers mapped to package paths", func(t *testing.T) {
		t.Parallel()

		pp := New("test")

		descs := []ocispec.Descriptor{
			{
				Annotations: map[string]string{
					ocispec.AnnotationTitle: "components/c2.tar",
				},
			},
			{
				Annotations: map[string]string{
					ocispec.AnnotationTitle: "images/blobs/sha256/" + strings.Repeat("1", 64),
				},
			},
		}
		pp.AddImages()
		pp.SetFromLayers(descs)

		files := pp.Files()
		expected := map[string]string{
			"jackal.yaml":       normalizePath("test/jackal.yaml"),
			"checksums.txt":     normalizePath("test/checksums.txt"),
			"components/c2.tar": normalizePath("test/components/c2.tar"),
			"images/index.json": normalizePath("test/images/index.json"),
			"images/oci-layout": normalizePath("test/images/oci-layout"),
			"images/blobs/sha256/" + strings.Repeat("1", 64): normalizePath("test/images/blobs/sha256/" + strings.Repeat("1", 64)),
		}
		require.Equal(t, expected, files)
	})
}

// normalizePath ensures that the filepaths being generated are normalized to the host OS.
func normalizePath(path string) string {
	if runtime.GOOS != "windows" {
		return path
	}

	return strings.ReplaceAll(path, "/", "\\")
}

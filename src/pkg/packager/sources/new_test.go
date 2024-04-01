// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package sources contains core implementations of the PackageSource interface.
package sources

import (
	"fmt"
	"testing"

	"github.com/defenseunicorns/jackal/src/types"
	"github.com/stretchr/testify/require"
)

var ociS *OCISource
var urlS *URLSource
var tarballS *TarballSource
var splitS *SplitTarballSource
var packageS *PackageSource

type source struct {
	pkgSrc  string
	srcType string
	source  PackageSource
}

var sources = []source{
	{pkgSrc: "oci://ghcr.io/defenseunicorns/packages/init:1.0.0", srcType: "oci", source: ociS},
	{pkgSrc: "sget://github.com/defenseunicorns/jackal-hello-world:x86", srcType: "sget", source: urlS},
	{pkgSrc: "sget://defenseunicorns/jackal-hello-world:x86_64", srcType: "sget", source: urlS},
	{pkgSrc: "https://github.com/defenseunicorns/jackal/releases/download/v1.0.0/jackal-init-amd64-v1.0.0.tar.zst", srcType: "https", source: urlS},
	{pkgSrc: "http://github.com/defenseunicorns/jackal/releases/download/v1.0.0/jackal-init-amd64-v1.0.0.tar.zst", srcType: "http", source: urlS},
	{pkgSrc: "jackal-init-amd64-v1.0.0.tar.zst", srcType: "tarball", source: tarballS},
	{pkgSrc: "jackal-package-manifests-amd64-v1.0.0.tar", srcType: "tarball", source: tarballS},
	{pkgSrc: "jackal-package-manifests-amd64-v1.0.0.tar.zst", srcType: "tarball", source: tarballS},
	{pkgSrc: "some-dir/.part000", srcType: "split", source: splitS},
}

func Test_identifySourceType(t *testing.T) {
	for _, source := range sources {
		actual := Identify(source.pkgSrc)
		require.Equalf(t, source.srcType, actual, fmt.Sprintf("source: %s", source))
	}
}

func TestNew(t *testing.T) {
	for _, source := range sources {
		actual, err := New(&types.JackalPackageOptions{PackageSource: source.pkgSrc})
		require.NoError(t, err)
		require.IsType(t, source.source, actual)
		require.Implements(t, packageS, actual)
	}
}

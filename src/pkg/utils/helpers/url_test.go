// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package helpers provides generic helper functions with no external imports
package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestURLSuite struct {
	suite.Suite
	*require.Assertions
	urls testURLs
}

type testURLs struct {
	good  []string
	oci   []string
	paths []string
}

func (suite *TestURLSuite) SetupSuite() {
	suite.Assertions = require.New(suite.T())
	suite.urls.good = []string{
		"https://zarf.dev",
		"https://docs.zarf.dev",
		"https://zarf.dev/docs",
		"https://defenseunicorns.com",
		"https://google.com",
		"https://user:pass@hello.world?foo=bar",
		"https://zarf.dev?foo=bar&bar=baz",
	}
	suite.urls.oci = []string{
		"oci://zarf.dev",
		"oci://defenseunicorns.com",
		"oci://google.com",
	}
	suite.urls.paths = []string{
		"./hello.txt",
		"../hello.txt",
		"/tmp/hello.txt",
	}
}

func (suite *TestURLSuite) Test_0_IsURL() {
	all := append(suite.urls.good, suite.urls.oci...)
	for _, url := range all {
		suite.True(IsURL(url), "Expected %s to be a valid URL", url)
	}
	for _, url := range suite.urls.paths {
		suite.False(IsURL(url), "Expected %s to be an invalid URL", url)
	}
}

func (suite *TestURLSuite) Test_1_IsOCIURL() {
	for _, url := range suite.urls.good {
		suite.False(IsOCIURL(url), "Expected %s to be an invalid OCI URL", url)
	}
	for _, url := range suite.urls.oci {
		suite.True(IsOCIURL(url), "Expected %s to be a valid OCI URL", url)
	}
	for _, url := range suite.urls.paths {
		suite.False(IsOCIURL(url), "Expected %s to be an invalid OCI URL", url)
	}
}

func (suite *TestURLSuite) Test_2_DoHostnamesMatch() {

	b, err := DoHostnamesMatch("https://zarf.dev", "https://zarf.dev")
	suite.NoError(err)
	suite.True(b)
	b, err = DoHostnamesMatch("https://zarf.dev", "https://docs.zarf.dev")
	suite.NoError(err)
	suite.False(b)
	b, err = DoHostnamesMatch("https://zarf.dev", "https://zarf.dev/docs")
	suite.NoError(err)
	suite.True(b)
	b, err = DoHostnamesMatch("https://zarf.dev", "https://user:pass@zarf.dev")
	suite.NoError(err)
	suite.True(b)
	b, err = DoHostnamesMatch("https://zarf.dev", "")
	suite.NoError(err)
	suite.False(b)
	b, err = DoHostnamesMatch("", "https://zarf.dev")
	suite.NoError(err)
	suite.False(b)
}

func TestURL(t *testing.T) {
	suite.Run(t, new(TestURLSuite))
}

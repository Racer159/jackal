// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package utils provides generic utility functions.package utils
package utils

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	mocks "github.com/racer159/jackal/src/test/mocks"
	"github.com/stretchr/testify/require"
)

func TestCredentialParser(t *testing.T) {
	credentialsFile := &mocks.MockReadCloser{
		MockData: []byte(
			`https://wayne:password@github.com/
bad line
<really bad="line"/>
https://wayne:p%40ss%20word%2520@jackal.dev
http://google.com`,
		),
	}

	expectedCreds := []Credential{
		{
			Path: "github.com",
			Auth: http.BasicAuth{
				Username: "wayne",
				Password: "password",
			},
		},
		{
			Path: "jackal.dev",
			Auth: http.BasicAuth{
				Username: "wayne",
				Password: "p@ss word%20",
			},
		},
		{
			Path: "google.com",
			Auth: http.BasicAuth{
				Username: "",
				Password: "",
			},
		},
	}

	gitCredentials := credentialParser(credentialsFile)
	require.Equal(t, expectedCreds, gitCredentials)
}

func TestNetRCParser(t *testing.T) {

	netrcFile := &mocks.MockReadCloser{
		MockData: []byte(
			`# top of file comment
machine github.com
	login wayne
    password password

 machine jackal.dev login wayne password p@s#sword%20

macdef macro-name
touch file
 echo "I am a script and can do anything password fun or login info yay!"

machine google.com #comment password fun and login info!

default
  login anonymous
	password password`,
		),
	}

	expectedCreds := []Credential{
		{
			Path: "github.com",
			Auth: http.BasicAuth{
				Username: "wayne",
				Password: "password",
			},
		},
		{
			Path: "jackal.dev",
			Auth: http.BasicAuth{
				Username: "wayne",
				Password: "p@s#sword%20",
			},
		},
		{
			Path: "google.com",
			Auth: http.BasicAuth{
				Username: "",
				Password: "",
			},
		},
		{
			Path: "",
			Auth: http.BasicAuth{
				Username: "anonymous",
				Password: "password",
			},
		},
	}

	netrcCredentials := netrcParser(netrcFile)
	require.Equal(t, expectedCreds, netrcCredentials)
}

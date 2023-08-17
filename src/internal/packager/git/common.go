// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package git contains functions for interacting with git repositories.
package git

import (
	"fmt"
	"strings"

	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/types"
	"github.com/go-git/go-git/v5/plumbing"
)

// GitCfg is the main struct for managing git repositories.
type GitCfg struct {
	// Server is the git server configuration.
	server types.GitServerInfo
	// Spinner is an optional spinner to use for long running operations.
	spinner *message.Spinner
	// Target working directory for the git repository.
	gitPath string
}

const onlineRemoteName = "online-upstream"
const offlineRemoteName = "offline-downstream"
const emptyRef = ""

// New creates a new git instance with the provided server config.
func New(server types.GitServerInfo) *GitCfg {
	return &GitCfg{
		server: server,
	}
}

// WithSpinner adds a spinner to the git config.
func (g *GitCfg) WithSpinner(spinner *message.Spinner) *GitCfg {
	g.spinner = spinner
	return g
}

// ParseRef parses the provided ref into a ReferenceName if it's not a hash.
func ParseRef(r string) plumbing.ReferenceName {
	// If not a full ref, assume it's a tag at this point.
	if !plumbing.IsHash(r) && !strings.HasPrefix(r, "refs/") {
		r = fmt.Sprintf("refs/tags/%s", r)
	}

	// Set the reference name to the provided ref.
	return plumbing.ReferenceName(r)
}

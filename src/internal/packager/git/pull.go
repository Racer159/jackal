// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package git contains functions for interacting with git repositories.
package git

import (
	"fmt"
	"path"
	"strings"

	"github.com/defenseunicorns/jackal/src/config"
	"github.com/defenseunicorns/jackal/src/pkg/transform"
	"github.com/defenseunicorns/jackal/src/pkg/utils"
	"github.com/go-git/go-git/v5/plumbing"
)

// DownloadRepoToTemp clones or updates a repo into a temp folder to perform ephemeral actions (i.e. process chart repos).
func (g *Git) DownloadRepoToTemp(gitURL string) error {
	g.Spinner.Updatef("g.DownloadRepoToTemp(%s)", gitURL)

	path, err := utils.MakeTempDir(config.CommonOptions.TempDirectory)
	if err != nil {
		return fmt.Errorf("unable to create tmpdir: %w", err)
	}

	// If downloading to temp, set this as a shallow clone to only pull the exact
	// gitURL w/ ref that was specified since we will throw away git history anyway
	if err = g.Pull(gitURL, path, true); err != nil {
		return fmt.Errorf("unable to pull the git repo at %s: %w", gitURL, err)
	}

	return nil
}

// Pull clones or updates a git repository into the target folder.
func (g *Git) Pull(gitURL, targetFolder string, shallow bool) error {
	g.Spinner.Updatef("Processing git repo %s", gitURL)

	// Split the remote url and the jackal reference
	gitURLNoRef, refPlain, err := transform.GitURLSplitRef(gitURL)
	if err != nil {
		return err
	}

	var ref plumbing.ReferenceName

	// Parse the ref from the git URL.
	if refPlain != emptyRef {
		ref = ParseRef(refPlain)
	}

	// Construct a path unique to this git repo
	repoFolder, err := transform.GitURLtoFolderName(gitURL)
	if err != nil {
		return err
	}

	g.GitPath = path.Join(targetFolder, repoFolder)

	// Clone the git repository.
	err = g.clone(gitURLNoRef, ref, shallow)
	if err != nil {
		return fmt.Errorf("not a valid git repo or unable to clone (%s): %w", gitURL, err)
	}

	if ref != emptyRef && !ref.IsBranch() {
		// Remove the "refs/tags/" prefix from the ref.
		stripped := strings.TrimPrefix(refPlain, "refs/tags/")

		// Use the plain ref as part of the branch name so it is unique and doesn't conflict with other refs.
		alias := fmt.Sprintf("jackal-ref-%s", stripped)
		trunkBranchName := plumbing.NewBranchReferenceName(alias)

		// Checkout the ref as a branch.
		return g.checkoutRefAsBranch(stripped, trunkBranchName)
	}

	return nil
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package git contains functions for interacting with git repositories.
package git

import (
	"context"
	"errors"
	"strings"

	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/pkg/utils/exec"
	"github.com/go-git/go-git/v5"
	goConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

// clone performs a `git clone` of a given repo.
func (g *GitCfg) clone(gitURL string, ref plumbing.ReferenceName, shallow bool) error {
	cloneOptions := &git.CloneOptions{
		URL:        gitURL,
		Progress:   g.Spinner,
		RemoteName: onlineRemoteName,
	}

	// Don't clone all tags / refs if we're cloning a specific tag or branch.
	if ref.IsTag() || ref.IsBranch() {
		cloneOptions.Tags = git.NoTags
		cloneOptions.ReferenceName = ref
		cloneOptions.SingleBranch = true
	}

	// If this is a shallow clone set the depth to 1
	if shallow {
		cloneOptions.Depth = 1
	}

	// Setup git credentials if we have them, ignore if we don't.
	gitCred := utils.FindAuthForHost(gitURL)
	if gitCred != nil {
		cloneOptions.Auth = &gitCred.Auth
	}

	// Clone the given repo.
	repo, err := git.PlainClone(g.GitPath, false, cloneOptions)
	if err != nil {
		message.Warnf("Falling back to host 'git', failed to clone the repo with Zarf - %s: %s", gitURL, err.Error())
		return g.gitCloneFallback(gitURL, ref, shallow)
	}

	// If we're cloning the whole repo, we need to also fetch the other branches besides the default.
	if ref == emptyRef {
		fetchOpts := &git.FetchOptions{
			RemoteName: onlineRemoteName,
			Progress:   g.Spinner,
			RefSpecs:   []goConfig.RefSpec{"refs/*:refs/*"},
			Tags:       git.AllTags,
		}

		if gitCred != nil {
			fetchOpts.Auth = &gitCred.Auth
		}

		if err := repo.Fetch(fetchOpts); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return err
		}
	}

	return nil
}

// gitCloneFallback is a fallback if go-git fails to clone a repo.
func (g *GitCfg) gitCloneFallback(gitURL string, ref plumbing.ReferenceName, shallow bool) error {
	// If we can't clone with go-git, fallback to the host clone
	// Only support "all tags" due to the azure clone url format including a username
	cloneArgs := []string{"clone", "--origin", onlineRemoteName, gitURL, g.GitPath}

	// Don't clone all tags / refs if we're cloning a specific tag or branch.
	if ref.IsTag() || ref.IsBranch() {
		cloneArgs = append(cloneArgs, "--no-tags")
		cloneArgs = append(cloneArgs, "-b", ref.Short())
		cloneArgs = append(cloneArgs, "--single-branch")
	}

	// If this is a shallow clone set the depth to 1
	if shallow {
		cloneArgs = append(cloneArgs, "--depth", "1")
	}

	cloneExecConfig := exec.Config{
		Stdout: g.Spinner,
		Stderr: g.Spinner,
	}

	message.Command("git %s", strings.Join(cloneArgs, " "))

	_, _, err := exec.CmdWithContext(context.TODO(), cloneExecConfig, "git", cloneArgs...)
	if err != nil {
		return err
	}

	// If we're cloning the whole repo, we need to also fetch the other branches besides the default.
	if ref == emptyRef {
		fetchArgs := []string{"fetch", "--tags", "--update-head-ok", onlineRemoteName, "refs/*:refs/*"}

		fetchExecConfig := exec.Config{
			Stdout: g.Spinner,
			Stderr: g.Spinner,
			Dir:    g.GitPath,
		}

		message.Command("git %s", strings.Join(fetchArgs, " "))

		_, _, err := exec.CmdWithContext(context.TODO(), fetchExecConfig, "git", fetchArgs...)
		if err != nil {
			return err
		}
	}

	return nil
}

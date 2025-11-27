package controller

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// CloneOrUpdate clones the repo if it doesn't exist in cachePath, otherwise fetches updates.
// It returns the SHA of the checked-out revision.
func CloneOrUpdate(repoURL, revision, branch string, auth *http.BasicAuth) (string, error) {
	cachePath := filepath.Join(os.TempDir(), "config-synchronizer-operator-clone")

	var repo *git.Repository
	var err error

	if _, err = os.Stat(cachePath); os.IsNotExist(err) {
		fmt.Printf("Cloning repo %s to %s\n", repoURL, cachePath)
		cloneOpts := &git.CloneOptions{
			URL:           repoURL,
			Auth:          auth,
			SingleBranch:  true,
			ReferenceName: plumbing.Master,
		}
		if branch != "" {
			cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(branch)
		}
		repo, err = git.PlainClone(cachePath, false, cloneOpts)
		if err != nil {
			return "", fmt.Errorf("failed to clone repo: %w", err)
		}
	} else {
		repo, err = git.PlainOpen(cachePath)
		if err != nil {
			return "", fmt.Errorf("failed to open existing repo: %w", err)
		}

		// Fetch
		fmt.Printf("Fetching updates for repo %s\n", repoURL)
		err = repo.Fetch(&git.FetchOptions{
			RemoteName: "origin",
			Auth:       auth,
			Force:      true,
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return "", fmt.Errorf("failed to fetch updates: %w", err)
		}
	}
	// Checkout
	w, _ := repo.Worktree()
	if revision != "" {
		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(revision),
		})
		if err != nil {
			return "", fmt.Errorf("failed to checkout revision %s: %w", revision, err)
		}
	} else if branch != "" {
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
		})
		if err != nil {
			return "", fmt.Errorf("failed to checkout branch %s: %w", branch, err)
		}
	}

	head, _ := repo.Head()
	return head.Hash().String(), nil
}

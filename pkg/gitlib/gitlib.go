package gitlib

import (
	"fmt"
	"os"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sandrolain/gomsvc/pkg/datalib"
)

// RefType represents the type of Git reference (commit, branch, or tag)
type RefType string

const (
	// RefTypeCommit represents a specific commit hash
	RefTypeCommit RefType = "C"
	// RefTypeBranch represents a branch name
	RefTypeBranch RefType = "B"
	// RefTypeTag represents a tag name
	RefTypeTag RefType = "T"
)

// GitRef contains information needed to clone and checkout a specific Git reference
type GitRef struct {
	// Url is the Git repository URL
	Url string
	// Type specifies whether this reference is a commit, branch, or tag
	Type RefType
	// Ref is the actual reference value (commit hash, branch name, or tag name)
	Ref string
	// Depth specifies the number of commits to fetch. If 0 or negative, fetches all commits
	Depth int
}

// Clone clones a Git repository and checks out the specified reference.
// It creates a directory in workpath based on the repository URL and reference details.
// If GitRef.Depth > 0, performs a shallow clone with the specified depth.
// Returns the path to the cloned repository and any error encountered.
func Clone(r GitRef, workpath string) (dest string, err error) {
	dirName, err := datalib.SafeDirName(r.Url, "_", string(r.Type), "_", r.Ref)
	if err != nil {
		return
	}

	dest = path.Join(workpath, dirName)

	opts := &git.CloneOptions{
		URL:               r.Url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Progress:          os.Stdout,
	}

	if r.Depth > 0 {
		opts.Depth = r.Depth
	}

	repo, err := git.PlainClone(dest, false, opts)
	if err != nil {
		err = fmt.Errorf("cannot clone repo: %w", err)
		return
	}

	w, err := repo.Worktree()
	if err != nil {
		err = fmt.Errorf("cannot obtain worktree: %w", err)
		return
	}

	if r.Type == RefTypeCommit {
		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(r.Ref),
		})
		if err != nil {
			err = fmt.Errorf("cannot checkout hash: %w", err)
			return
		}
		return
	}

	var branch plumbing.ReferenceName
	switch r.Type {
	case RefTypeBranch:
		branch = plumbing.NewBranchReferenceName(r.Ref)
	case RefTypeTag:
		branch = plumbing.NewTagReferenceName(r.Ref)
	}
	err = w.Checkout(&git.CheckoutOptions{
		Branch: branch,
	})
	if err != nil {
		err = fmt.Errorf("cannot checkout branch: %w", err)
		return
	}

	return
}

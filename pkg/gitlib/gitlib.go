package gitlib

import (
	"fmt"
	"os"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sandrolain/gomsvc/pkg/datalib"
)

type RefType string

const (
	RefTypeCommit RefType = "C"
	RefTypeBranch RefType = "B"
	RefTypeTag    RefType = "T"
)

type GitRef struct {
	Url  string
	Type RefType
	Ref  string
}

func Clone(r GitRef, workpath string) (dest string, err error) {
	dirName, err := datalib.SafeDirName(r.Url, "_", string(r.Type), "_", r.Ref)
	if err != nil {
		return
	}

	dest = path.Join(workpath, dirName)

	repo, err := git.PlainClone(dest, false, &git.CloneOptions{
		URL:               r.Url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Progress:          os.Stdout,
	})
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

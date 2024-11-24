// Package gitlib provides functionality for Git repository operations such as cloning
// and checking out specific references (commits, branches, or tags).
//
// The package uses go-git as its underlying implementation and provides a simplified
// interface for common Git operations.
//
// Example usage:
//
//	// Full clone
//	ref := gitlib.GitRef{
//		Url:  "https://github.com/example/repo.git",
//		Type: gitlib.RefTypeBranch,
//		Ref:  "main",
//	}
//	dest, err := gitlib.Clone(ref, "/path/to/workspace")
//
//	// Shallow clone with depth 1
//	shallowRef := gitlib.GitRef{
//		Url:   "https://github.com/example/repo.git",
//		Type:  gitlib.RefTypeBranch,
//		Ref:   "main",
//		Depth: 1,
//	}
//	dest, err = gitlib.Clone(shallowRef, "/path/to/workspace")
package gitlib

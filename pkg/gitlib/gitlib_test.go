package gitlib

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClone(t *testing.T) {
	tests := []struct {
		name    string
		ref     GitRef
		wantErr bool
	}{
		{
			name: "clone main branch",
			ref: GitRef{
				Url:  "https://github.com/sandrolain/gomsvc",
				Type: RefTypeBranch,
				Ref:  "main",
			},
			wantErr: false,
		},
		{
			name: "shallow clone with depth 1",
			ref: GitRef{
				Url:   "https://github.com/sandrolain/gomsvc",
				Type:  RefTypeBranch,
				Ref:   "main",
				Depth: 1,
			},
			wantErr: false,
		},
		{
			name: "invalid repository url",
			ref: GitRef{
				Url:  "https://invalid-url/repo.git",
				Type: RefTypeBranch,
				Ref:  "main",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for each test case
			tmpDir, err := os.MkdirTemp("", "gitlib-test-*")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				err := os.RemoveAll(tmpDir)
				if err != nil {
					t.Fatalf("failed to clean up temp dir: %v", err)
				}
			}()

			dest, err := Clone(tt.ref, tmpDir)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, dest)

			// Verify the directory exists
			_, err = os.Stat(dest)
			assert.NoError(t, err)

			// Verify it's a git repository
			gitDir := path.Join(dest, ".git")
			_, err = os.Stat(gitDir)
			assert.NoError(t, err)

			// For shallow clones, verify shallow file exists
			if tt.ref.Depth > 0 {
				_, err = os.Stat(path.Join(gitDir, "shallow"))
				assert.NoError(t, err, "shallow file should exist for shallow clones")
			}
		})
	}
}

func TestRefType(t *testing.T) {
	assert.Equal(t, RefType("C"), RefTypeCommit)
	assert.Equal(t, RefType("B"), RefTypeBranch)
	assert.Equal(t, RefType("T"), RefTypeTag)
}

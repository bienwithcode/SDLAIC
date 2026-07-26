package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectHash(t *testing.T) {
	dir := t.TempDir()
	// Create a .git directory so it looks like a git repo
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0755))

	hash1, err := ProjectHash(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, hash1)

	// Hash should be deterministic
	hash2, err := ProjectHash(dir)
	require.NoError(t, err)
	assert.Equal(t, hash1, hash2)
}

func TestProjectHash_DifferentDirs(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	hash1, err := ProjectHash(dir1)
	require.NoError(t, err)
	hash2, err := ProjectHash(dir2)
	require.NoError(t, err)

	// Different directories should produce different hashes
	assert.NotEqual(t, hash1, hash2)
}

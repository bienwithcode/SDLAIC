package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sdlaic/internal/domain"
)

func TestResolvePath_LocalMode(t *testing.T) {
	projectRoot := "/tmp/myproject"
	path, err := ResolvePath(domain.StorageModeLocal, projectRoot, "", "my-change")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(projectRoot, ".sdlaic", "changes", "my-change"), path)
}

func TestResolvePath_IgnoredMode(t *testing.T) {
	projectRoot := "/tmp/myproject"
	path, err := ResolvePath(domain.StorageModeIgnored, projectRoot, "", "my-change")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(projectRoot, ".sdlaic", "changes", "my-change"), path)
}

func TestResolvePath_GlobalMode(t *testing.T) {
	projectRoot := "/tmp/myproject"
	homeDir := "/home/user"
	path, err := ResolvePath(domain.StorageModeGlobal, projectRoot, homeDir, "my-change")
	require.NoError(t, err)
	assert.Contains(t, path, ".sdlaic")
	assert.Contains(t, path, "stores")
	assert.Contains(t, path, "changes")
	assert.Contains(t, path, "my-change")
}

func TestResolvePath_GlobalModeRequiresHash(t *testing.T) {
	projectRoot := "/tmp/myproject"
	// Without a project hash, global mode should still work — hash comes from projectRoot
	path, err := ResolvePath(domain.StorageModeGlobal, projectRoot, "/home/user", "change1")
	require.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestResolvePath_InvalidMode(t *testing.T) {
	_, err := ResolvePath(domain.StorageMode("invalid"), "/tmp/project", "", "change")
	assert.Error(t, err)
}

func TestResolvePath_EmptyChangeName(t *testing.T) {
	_, err := ResolvePath(domain.StorageModeLocal, "/tmp/project", "", "")
	assert.Error(t, err)
}

func TestChangesBasePath_LocalMode(t *testing.T) {
	projectRoot := "/tmp/myproject"
	path, err := ChangesBasePath(domain.StorageModeLocal, projectRoot, "")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(projectRoot, ".sdlaic", "changes"), path)
}

func TestChangesBasePath_IgnoredMode(t *testing.T) {
	projectRoot := "/tmp/myproject"
	path, err := ChangesBasePath(domain.StorageModeIgnored, projectRoot, "")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(projectRoot, ".sdlaic", "changes"), path)
}

func TestChangesBasePath_GlobalMode(t *testing.T) {
	projectRoot := "/tmp/myproject"
	path, err := ChangesBasePath(domain.StorageModeGlobal, projectRoot, "/home/user")
	require.NoError(t, err)
	assert.Contains(t, path, ".sdlaic")
	assert.Contains(t, path, "stores")
}

func TestChangesBasePath_InvalidMode(t *testing.T) {
	_, err := ChangesBasePath(domain.StorageMode("invalid"), "/tmp/project", "")
	assert.Error(t, err)
}

func TestComputeProjectHash_Deterministic(t *testing.T) {
	hash1 := ComputeProjectHash("/tmp/myproject")
	hash2 := ComputeProjectHash("/tmp/myproject")
	assert.Equal(t, hash1, hash2)
	assert.NotEmpty(t, hash1)
}

func TestComputeProjectHash_DifferentPaths(t *testing.T) {
	hash1 := ComputeProjectHash("/tmp/project-a")
	hash2 := ComputeProjectHash("/tmp/project-b")
	assert.NotEqual(t, hash1, hash2)
}

func TestListChanges(t *testing.T) {
	dir := t.TempDir()
	changesDir := filepath.Join(dir, "changes")
	require.NoError(t, os.MkdirAll(changesDir, 0755))

	// Create some change directories
	require.NoError(t, os.MkdirAll(filepath.Join(changesDir, "change-1"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(changesDir, "change-2"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(changesDir, "change-3"), 0755))

	changes, err := ListChanges(changesDir)
	require.NoError(t, err)
	assert.Len(t, changes, 3)
	assert.Contains(t, changes, "change-1")
	assert.Contains(t, changes, "change-2")
	assert.Contains(t, changes, "change-3")
}

func TestListChanges_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	changesDir := filepath.Join(dir, "changes")
	require.NoError(t, os.MkdirAll(changesDir, 0755))

	changes, err := ListChanges(changesDir)
	require.NoError(t, err)
	assert.Empty(t, changes)
}

func TestListChanges_NonexistentDir(t *testing.T) {
	changes, err := ListChanges("/nonexistent/path")
	assert.NoError(t, err)
	assert.Nil(t, changes)
}

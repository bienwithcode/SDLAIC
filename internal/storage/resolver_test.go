package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// --- ChangesDir helpers (storage-mode-free path resolution) ---

func TestDefaultChangesDir(t *testing.T) {
	assert.Equal(t, "/tmp/project/.sdlaic/changes", DefaultChangesDir("/tmp/project"))
}

func TestNormalizeChangesDir_KeepsAbsolutePath(t *testing.T) {
	got, err := NormalizeChangesDir("/work/openspec/changes", "/cwd", "/home/dev")
	require.NoError(t, err)
	assert.Equal(t, "/work/openspec/changes", got)
}

func TestNormalizeChangesDir_ResolvesRelativeAgainstCwd(t *testing.T) {
	got, err := NormalizeChangesDir("../openspec/changes", "/work/api", "/home/dev")
	require.NoError(t, err)
	assert.Equal(t, "/work/openspec/changes", got)
}

func TestNormalizeChangesDir_ExpandsTilde(t *testing.T) {
	got, err := NormalizeChangesDir("~/openspec/changes", "/cwd", "/home/dev")
	require.NoError(t, err)
	assert.Equal(t, "/home/dev/openspec/changes", got)
}

func TestNormalizeChangesDir_ExpandsBareTilde(t *testing.T) {
	got, err := NormalizeChangesDir("~", "/cwd", "/home/dev")
	require.NoError(t, err)
	assert.Equal(t, "/home/dev", got)
}

func TestNormalizeChangesDir_StripsTrailingSeparator(t *testing.T) {
	got, err := NormalizeChangesDir("/work/openspec/changes/", "/cwd", "/home/dev")
	require.NoError(t, err)
	assert.Equal(t, "/work/openspec/changes", got)
}

func TestNormalizeChangesDir_RejectsEmptyInput(t *testing.T) {
	_, err := NormalizeChangesDir("  ", "/cwd", "/home/dev")
	assert.Error(t, err)
}

func TestChangesBase_ReturnsCleanedDir(t *testing.T) {
	got, err := ChangesBase("/work/openspec/changes/")
	require.NoError(t, err)
	assert.Equal(t, "/work/openspec/changes", got)
}

func TestChangesBase_RejectsUnconfiguredProject(t *testing.T) {
	_, err := ChangesBase("")
	assert.ErrorIs(t, err, domain.ErrChangesDirNotSet)
}

func TestChangePath_JoinsChangeName(t *testing.T) {
	got, err := ChangePath("/work/openspec/changes", "SDL-1")
	require.NoError(t, err)
	assert.Equal(t, "/work/openspec/changes/SDL-1", got)
}

func TestChangePath_RejectsEmptyChangeName(t *testing.T) {
	_, err := ChangePath("/work/openspec/changes", "")
	assert.Error(t, err)
}

func TestChangePath_RejectsUnconfiguredProject(t *testing.T) {
	_, err := ChangePath("", "SDL-1")
	assert.ErrorIs(t, err, domain.ErrChangesDirNotSet)
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

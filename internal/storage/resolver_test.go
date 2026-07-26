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

func TestCanonicalPath_ResolvesSymlink(t *testing.T) {
	base := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(base, "real"), 0755))
	require.NoError(t, os.Symlink(filepath.Join(base, "real"), filepath.Join(base, "link")))

	viaLink := CanonicalPath(filepath.Join(base, "link"))
	direct := CanonicalPath(filepath.Join(base, "real"))

	assert.Equal(t, direct, viaLink)
}

func TestCanonicalPath_ResolvesThroughSymlinkedParentForMissingLeaf(t *testing.T) {
	base := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(base, "real"), 0755))
	require.NoError(t, os.Symlink(filepath.Join(base, "real"), filepath.Join(base, "link")))

	// "changes" does not exist yet — the parent still has to be resolved, or a
	// collision only becomes visible after the directory is created.
	viaLink := CanonicalPath(filepath.Join(base, "link", "changes"))
	direct := CanonicalPath(filepath.Join(base, "real", "changes"))

	assert.Equal(t, direct, viaLink)
}

func TestCanonicalPath_IsAbsoluteForRelativeInput(t *testing.T) {
	assert.True(t, filepath.IsAbs(CanonicalPath("some/relative/dir")))
}

func TestCanonicalPath_HandlesFullyMissingPath(t *testing.T) {
	assert.Equal(t, "/definitely/not/here", CanonicalPath("/definitely/not/here"))
}

func TestSamePath_DistinguishesDifferentDirectories(t *testing.T) {
	base := t.TempDir()
	assert.False(t, SamePath(filepath.Join(base, "a"), filepath.Join(base, "b")))
}

func TestSamePath_MatchesEquivalentSpellings(t *testing.T) {
	base := t.TempDir()
	assert.True(t, SamePath(filepath.Join(base, "a"), filepath.Join(base, ".", "a")))
}

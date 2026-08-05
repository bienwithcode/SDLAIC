package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// --- ChangesDir helpers (storage-mode-free path resolution) ---

// These paths are built with filepath rather than written as "/work/..." string
// literals. A POSIX literal is not absolute on Windows — filepath.IsAbs wants a
// drive letter — so the hardcoded form sent every one of these cases down the
// relative-path branch and compared "\work\..." against "/work/...".

// absPath returns an absolute path for the running platform, rooted at the
// volume the tests already live on.
func absPath(t *testing.T, elem ...string) string {
	t.Helper()
	return filepath.Join(append([]string{t.TempDir()}, elem...)...)
}

func TestDefaultChangesDir(t *testing.T) {
	root := absPath(t, "project")
	assert.Equal(t, filepath.Join(root, ".sdlaic", "changes"), DefaultChangesDir(root))
}

func TestNormalizeChangesDir_KeepsAbsolutePath(t *testing.T) {
	want := absPath(t, "work", "openspec", "changes")

	got, err := NormalizeChangesDir(want, absPath(t, "cwd"), absPath(t, "home"))

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestNormalizeChangesDir_ResolvesRelativeAgainstCwd(t *testing.T) {
	base := t.TempDir()

	got, err := NormalizeChangesDir(
		filepath.Join("..", "openspec", "changes"),
		filepath.Join(base, "work", "api"),
		filepath.Join(base, "home"))

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(base, "work", "openspec", "changes"), got)
}

func TestNormalizeChangesDir_ExpandsTilde(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "home", "dev")

	// A forward slash after the tilde, which is what users type on every
	// platform — including Windows shells.
	got, err := NormalizeChangesDir("~/openspec/changes", filepath.Join(base, "cwd"), home)

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "openspec", "changes"), got)
}

func TestNormalizeChangesDir_ExpandsTildeWithNativeSeparator(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "home", "dev")
	input := "~" + string(filepath.Separator) + filepath.Join("openspec", "changes")

	got, err := NormalizeChangesDir(input, filepath.Join(base, "cwd"), home)

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "openspec", "changes"), got)
}

func TestNormalizeChangesDir_ExpandsBareTilde(t *testing.T) {
	home := absPath(t, "home", "dev")

	got, err := NormalizeChangesDir("~", absPath(t, "cwd"), home)

	require.NoError(t, err)
	assert.Equal(t, home, got)
}

func TestNormalizeChangesDir_StripsTrailingSeparator(t *testing.T) {
	want := absPath(t, "work", "openspec", "changes")

	got, err := NormalizeChangesDir(want+string(filepath.Separator), absPath(t, "cwd"), absPath(t, "home"))

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestNormalizeChangesDir_RejectsEmptyInput(t *testing.T) {
	_, err := NormalizeChangesDir("  ", absPath(t, "cwd"), absPath(t, "home"))
	assert.Error(t, err)
}

func TestChangesBase_ReturnsCleanedDir(t *testing.T) {
	want := absPath(t, "work", "openspec", "changes")

	got, err := ChangesBase(want + string(filepath.Separator))

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestChangesBase_RejectsUnconfiguredProject(t *testing.T) {
	_, err := ChangesBase("")
	assert.ErrorIs(t, err, domain.ErrChangesDirNotSet)
}

func TestChangePath_JoinsChangeName(t *testing.T) {
	base := absPath(t, "work", "openspec", "changes")

	got, err := ChangePath(base, "SDL-1")

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(base, "SDL-1"), got)
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
	// The old assertion compared against the literal "/definitely/not/here", which
	// only held where that string is already absolute. On Windows filepath.Abs
	// prepends the volume, so it became D:\definitely\not\here. Assert the
	// properties the function promises instead of one platform's spelling.
	tail := filepath.Join("definitely", "not", "here")

	got := CanonicalPath(filepath.Join(t.TempDir(), tail))

	assert.True(t, filepath.IsAbs(got), "canonical form must be absolute")
	assert.True(t, strings.HasSuffix(got, tail), "missing remainder must be preserved: %s", got)
	assert.Equal(t, got, CanonicalPath(got), "canonical form must be stable")
}

func TestSamePath_DistinguishesDifferentDirectories(t *testing.T) {
	base := t.TempDir()
	assert.False(t, SamePath(filepath.Join(base, "a"), filepath.Join(base, "b")))
}

func TestSamePath_MatchesEquivalentSpellings(t *testing.T) {
	base := t.TempDir()
	assert.True(t, SamePath(filepath.Join(base, "a"), filepath.Join(base, ".", "a")))
}

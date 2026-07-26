package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// configWith builds a global config from hash → project root path.
func configWith(entries map[string]string) domain.GlobalConfig {
	cfg := domain.NewGlobalConfig()
	for hash, path := range entries {
		cfg.Projects[hash] = domain.ProjectEntry{Path: path, ChangesDir: filepath.Join(path, ".sdlaic", "changes")}
	}
	return cfg
}

func TestLookup_MatchesProjectRoot(t *testing.T) {
	root := t.TempDir()
	cfg := configWith(map[string]string{"aaa": root})

	hash, entry, err := Lookup(cfg, root)

	require.NoError(t, err)
	assert.Equal(t, "aaa", hash)
	assert.Equal(t, filepath.Join(root, ".sdlaic", "changes"), entry.ChangesDir)
}

func TestLookup_MatchesFromNestedSubdirectory(t *testing.T) {
	root := t.TempDir()
	deep := filepath.Join(root, "internal", "config")
	require.NoError(t, os.MkdirAll(deep, 0755))
	cfg := configWith(map[string]string{"aaa": root})

	hash, _, err := Lookup(cfg, deep)

	require.NoError(t, err)
	assert.Equal(t, "aaa", hash)
}

func TestLookup_NearestProjectWinsWhenNested(t *testing.T) {
	outer := t.TempDir()
	inner := filepath.Join(outer, "services", "billing")
	require.NoError(t, os.MkdirAll(inner, 0755))
	cfg := configWith(map[string]string{"outer": outer, "inner": inner})

	hash, _, err := Lookup(cfg, filepath.Join(inner))

	require.NoError(t, err)
	assert.Equal(t, "inner", hash, "the longest matching project path wins, mirroring the old walk-up")
}

func TestLookup_DoesNotMatchAcrossSegmentBoundary(t *testing.T) {
	base := t.TempDir()
	foo := filepath.Join(base, "foo")
	foobar := filepath.Join(base, "foobar")
	require.NoError(t, os.MkdirAll(foo, 0755))
	require.NoError(t, os.MkdirAll(foobar, 0755))
	cfg := configWith(map[string]string{"aaa": foo})

	_, _, err := Lookup(cfg, foobar)

	assert.ErrorIs(t, err, domain.ErrWorkspaceNotFound, "/a/foobar must not match project /a/foo")
}

func TestLookup_ResolvesSymlinkedPaths(t *testing.T) {
	// On macOS t.TempDir() hands back /var/... which resolves to /private/var/....
	// Comparing an unresolved cwd against a resolved stored path would silently
	// fail to match, so both sides must be resolved before comparison.
	raw := t.TempDir()
	resolved, err := filepath.EvalSymlinks(raw)
	require.NoError(t, err)

	cfg := configWith(map[string]string{"aaa": resolved})
	hash, _, err := Lookup(cfg, raw)

	require.NoError(t, err)
	assert.Equal(t, "aaa", hash)
}

func TestLookup_SkipsStaleEntryWhosePathIsGone(t *testing.T) {
	root := t.TempDir()
	cfg := configWith(map[string]string{
		"gone":  filepath.Join(root, "deleted-project"),
		"alive": root,
	})

	hash, _, err := Lookup(cfg, root)

	require.NoError(t, err)
	assert.Equal(t, "alive", hash)
}

func TestLookup_StaleEntryAloneIsNotFound(t *testing.T) {
	root := t.TempDir()
	cfg := configWith(map[string]string{"gone": filepath.Join(root, "deleted-project")})

	_, _, err := Lookup(cfg, root)

	assert.ErrorIs(t, err, domain.ErrWorkspaceNotFound)
}

func TestLookup_EmptyConfigIsNotFound(t *testing.T) {
	_, _, err := Lookup(domain.NewGlobalConfig(), t.TempDir())

	assert.ErrorIs(t, err, domain.ErrWorkspaceNotFound)
}

func TestLookup_UnregisteredDirectoryIsNotFound(t *testing.T) {
	base := t.TempDir()
	registered := filepath.Join(base, "registered")
	other := filepath.Join(base, "other")
	require.NoError(t, os.MkdirAll(registered, 0755))
	require.NoError(t, os.MkdirAll(other, 0755))
	cfg := configWith(map[string]string{"aaa": registered})

	_, _, err := Lookup(cfg, other)

	assert.ErrorIs(t, err, domain.ErrWorkspaceNotFound)
}

func TestLookup_MissingDirectoryReturnsError(t *testing.T) {
	cfg := configWith(map[string]string{"aaa": t.TempDir()})

	_, _, err := Lookup(cfg, filepath.Join(t.TempDir(), "does-not-exist"))

	require.Error(t, err)
	assert.False(t, errors.Is(err, domain.ErrWorkspaceNotFound), "an unreadable cwd is a real error, not a missing project")
}

func TestProjectHash_IsStableAcrossSymlinkedPaths(t *testing.T) {
	// A project reached through a symlink must hash to the same value as the
	// resolved path, or Lookup (which resolves) would never find the entry that
	// ProjectHash keyed.
	raw := t.TempDir()
	resolved, err := filepath.EvalSymlinks(raw)
	require.NoError(t, err)

	fromRaw, err := ProjectHash(raw)
	require.NoError(t, err)
	fromResolved, err := ProjectHash(resolved)
	require.NoError(t, err)

	assert.Equal(t, fromResolved, fromRaw)
}

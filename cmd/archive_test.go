package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var datedArchiveName = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-ARCHIVE-ME$`)

func TestArchive_MovesToDatedDirectory(t *testing.T) {
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "ARCHIVE-ME")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "ARCHIVE-ME")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\nContent"), 0644))

	_, err = ExecuteCommand(rootCmd, "archive", "ARCHIVE-ME")
	require.NoError(t, err)

	// Original directory should be gone
	_, err = os.Stat(changeDir)
	assert.True(t, os.IsNotExist(err))

	// Archive should hold exactly one date-prefixed DIRECTORY (not tar.gz)
	archiveDir := filepath.Join(dir, ".sdlaic", "changes", ".archive")
	entries, err := os.ReadDir(archiveDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.True(t, entries[0].IsDir(), "archive entry should be a plain directory")
	assert.Regexp(t, datedArchiveName, entries[0].Name())

	// Files stay plain and readable — no extraction needed
	data, err := os.ReadFile(filepath.Join(archiveDir, entries[0].Name(), "proposal.md"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "# Proposal")
}

func TestArchive_ClearsActive(t *testing.T) {
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "ACTIVE-ARCHIVE")
	require.NoError(t, err)

	_, err = ExecuteCommand(rootCmd, "archive", "ACTIVE-ARCHIVE")
	require.NoError(t, err)

	assert.Empty(t, activeChangeOf(t, dir), "archiving the active change clears it")
}

func TestArchive_NotFound(t *testing.T) {
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "archive", "NONEXISTENT")
	assert.Error(t, err)
}

func TestArchive_DatePrefixIsNotStacked(t *testing.T) {
	dir := initWorkspaceForTest(t)

	// A change already carrying a date prefix keeps its name verbatim
	_, err := ExecuteCommand(rootCmd, "new", "change", "2025-01-01-legacy-fix")
	require.NoError(t, err)

	_, err = ExecuteCommand(rootCmd, "archive", "2025-01-01-legacy-fix")
	require.NoError(t, err)

	archiveDir := filepath.Join(dir, ".sdlaic", "changes", ".archive")
	entries, err := os.ReadDir(archiveDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "2025-01-01-legacy-fix", entries[0].Name())
}

func TestArchive_RefusesOverwriteAndKeepsOriginal(t *testing.T) {
	dir := initWorkspaceForTest(t)

	// First archive of the name
	_, err := ExecuteCommand(rootCmd, "new", "change", "DUP-NAME")
	require.NoError(t, err)
	first := filepath.Join(dir, ".sdlaic", "changes", "DUP-NAME")
	require.NoError(t, os.WriteFile(filepath.Join(first, "context.md"), []byte("first run"), 0644))
	_, err = ExecuteCommand(rootCmd, "archive", "DUP-NAME")
	require.NoError(t, err)

	// Same name appears again (re-created change)
	_, err = ExecuteCommand(rootCmd, "new", "change", "DUP-NAME")
	require.NoError(t, err)
	second := filepath.Join(dir, ".sdlaic", "changes", "DUP-NAME")
	require.NoError(t, os.WriteFile(filepath.Join(second, "context.md"), []byte("second run"), 0644))

	// Second archive must be refused — and must NOT delete the live change
	_, err = ExecuteCommand(rootCmd, "archive", "DUP-NAME")
	require.Error(t, err, "archiving a name that already exists must fail")

	_, statErr := os.Stat(second)
	assert.NoError(t, statErr, "the live change directory must survive the refused archive")

	archiveDir := filepath.Join(dir, ".sdlaic", "changes", ".archive")
	entries, err := os.ReadDir(archiveDir)
	require.NoError(t, err)
	require.Len(t, entries, 1, "only the first archive should exist")

	// The first archive's content is untouched
	data, err := os.ReadFile(filepath.Join(archiveDir, entries[0].Name(), "context.md"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "first run")
}

func TestList_AllIncludesArchivedDirsAndLegacyTarGz(t *testing.T) {
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "LISTED")
	require.NoError(t, err)
	_, err = ExecuteCommand(rootCmd, "archive", "LISTED")
	require.NoError(t, err)

	// Legacy tar.gz archive from an older SDLAIC version
	archiveDir := filepath.Join(dir, ".sdlaic", "changes", ".archive")
	require.NoError(t, os.WriteFile(filepath.Join(archiveDir, "old-thing.tar.gz"), []byte("legacy"), 0644))

	resetStatusFlags()
	out, err := ExecuteCommand(rootCmd, "list", "--all")
	require.NoError(t, err)

	assert.Contains(t, out, "LISTED", "dated archive directory should be listed")
	assert.Contains(t, out, "old-thing", "legacy .tar.gz archive should still be listed")
}

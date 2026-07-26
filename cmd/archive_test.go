package cmd

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArchive_CompressesAndRemoves(t *testing.T) {
	resetStatusFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "ARCHIVE-ME")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "ARCHIVE-ME")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\nContent"), 0644))

	resetStatusFlags()
	_, err = ExecuteCommand(rootCmd, "archive", "ARCHIVE-ME")
	require.NoError(t, err)

	// Original directory should be gone
	_, err = os.Stat(changeDir)
	assert.True(t, os.IsNotExist(err))

	// Archive should exist
	archiveDir := filepath.Join(dir, ".sdlaic", "changes", ".archive")
	entries, err := os.ReadDir(archiveDir)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Contains(t, entries[0].Name(), "ARCHIVE-ME")
}

func TestArchive_ClearsActive(t *testing.T) {
	resetStatusFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "ACTIVE-ARCHIVE")
	require.NoError(t, err)

	resetStatusFlags()
	_, err = ExecuteCommand(rootCmd, "archive", "ACTIVE-ARCHIVE")
	require.NoError(t, err)

	assert.Empty(t, activeChangeOf(t, dir), "archiving the active change clears it")
}

func TestArchive_NotFound(t *testing.T) {
	resetStatusFlags()
	_ = initWorkspaceForTest(t)

	resetStatusFlags()
	_, err := ExecuteCommand(rootCmd, "archive", "NONEXISTENT")
	assert.Error(t, err)
}

func TestArchive_ArchiveIsTarGz(t *testing.T) {
	resetStatusFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "TARTEST")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "TARTEST")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "test.txt"), []byte("hello"), 0644))

	resetStatusFlags()
	_, err = ExecuteCommand(rootCmd, "archive", "TARTEST")
	require.NoError(t, err)

	// Verify the archive is a valid tar.gz
	archiveDir := filepath.Join(dir, ".sdlaic", "changes", ".archive")
	entries, err := os.ReadDir(archiveDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	archivePath := filepath.Join(archiveDir, entries[0].Name())
	f, err := os.Open(archivePath)
	require.NoError(t, err)
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	require.NoError(t, err)
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	foundTestTxt := false
	for {
		hdr, err := tr.Next()
		if err != nil {
			break
		}
		if hdr.Name == "test.txt" {
			foundTestTxt = true
		}
	}
	assert.True(t, foundTestTxt, "archive should contain test.txt")
}

package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendToGitignore_CreatesGitignore(t *testing.T) {
	dir := t.TempDir()
	entry := ".sdlaic/changes/"

	err := AppendToGitignore(dir, entry)
	require.NoError(t, err)

	gitignorePath := filepath.Join(dir, ".gitignore")
	data, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	assert.Contains(t, string(data), entry)
}

func TestAppendToGitignore_AppendsToExisting(t *testing.T) {
	dir := t.TempDir()
	gitignorePath := filepath.Join(dir, ".gitignore")
	require.NoError(t, os.WriteFile(gitignorePath, []byte("*.exe\n"), 0644))

	err := AppendToGitignore(dir, ".sdlaic/changes/")
	require.NoError(t, err)

	data, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "*.exe")
	assert.Contains(t, content, ".sdlaic/changes/")
}

func TestAppendToGitignore_Idempotent(t *testing.T) {
	dir := t.TempDir()
	entry := ".sdlaic/changes/"

	require.NoError(t, AppendToGitignore(dir, entry))
	require.NoError(t, AppendToGitignore(dir, entry))

	gitignorePath := filepath.Join(dir, ".gitignore")
	data, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)

	// Entry should appear exactly once
	count := strings.Count(string(data), entry)
	assert.Equal(t, 1, count)
}

func TestRemoveFromGitignore_RemovesEntry(t *testing.T) {
	dir := t.TempDir()
	gitignorePath := filepath.Join(dir, ".gitignore")
	content := "*.exe\n.sdlaic/changes/\n*.log\n"
	require.NoError(t, os.WriteFile(gitignorePath, []byte(content), 0644))

	err := RemoveFromGitignore(dir, ".sdlaic/changes/")
	require.NoError(t, err)

	data, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	result := string(data)
	assert.NotContains(t, result, ".sdlaic/changes/")
	assert.Contains(t, result, "*.exe")
	assert.Contains(t, result, "*.log")
}

func TestRemoveFromGitignore_EntryNotFound(t *testing.T) {
	dir := t.TempDir()
	gitignorePath := filepath.Join(dir, ".gitignore")
	require.NoError(t, os.WriteFile(gitignorePath, []byte("*.exe\n"), 0644))

	// Removing non-existent entry should not error
	err := RemoveFromGitignore(dir, ".sdlaic/changes/")
	assert.NoError(t, err)

	data, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "*.exe")
}

func TestRemoveFromGitignore_NoGitignore(t *testing.T) {
	dir := t.TempDir()

	// Removing from nonexistent .gitignore should not error
	err := RemoveFromGitignore(dir, ".sdlaic/changes/")
	assert.NoError(t, err)
}

func TestRemoveFromGitignore_CleansTrailingNewlines(t *testing.T) {
	dir := t.TempDir()
	gitignorePath := filepath.Join(dir, ".gitignore")
	// Only the entry we're removing
	require.NoError(t, os.WriteFile(gitignorePath, []byte(".sdlaic/changes/\n"), 0644))

	err := RemoveFromGitignore(dir, ".sdlaic/changes/")
	require.NoError(t, err)

	data, err := os.ReadFile(gitignorePath)
	require.NoError(t, err)
	assert.Empty(t, strings.TrimSpace(string(data)))
}

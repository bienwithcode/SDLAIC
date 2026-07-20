package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bienwithcode/SDLAIC/internal/config"
)

func TestNewChange_CreatesChangeDir(t *testing.T) {
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "FEATURE-100")
	require.NoError(t, err)

	// Change directory should exist
	changesDir := filepath.Join(dir, ".sdlaic", "changes", "FEATURE-100")
	_, err = os.Stat(changesDir)
	assert.NoError(t, err)
}

func TestNewChange_CreatesContextTemplate(t *testing.T) {
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "FEATURE-200")
	require.NoError(t, err)

	// context.md should exist in the change dir
	contextFile := filepath.Join(dir, ".sdlaic", "changes", "FEATURE-200", "context.md")
	_, err = os.Stat(contextFile)
	assert.NoError(t, err)

	// It should have content from the template
	data, err := os.ReadFile(contextFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "# Context")
}

func TestNewChange_SetsActive(t *testing.T) {
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "FEATURE-300")
	require.NoError(t, err)

	// Active change should be set in .sdlaicrc
	cfg, err := config.LoadLocal(filepath.Join(dir, ".sdlaicrc"))
	require.NoError(t, err)
	assert.Equal(t, "FEATURE-300", cfg.ActiveChange)
}

func TestNewChange_RejectsDuplicate(t *testing.T) {
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "FEATURE-400")
	require.NoError(t, err)

	// Creating same change again should fail
	_, err = ExecuteCommand(rootCmd, "new", "change", "FEATURE-400")
	assert.Error(t, err)
}

func TestNewChange_NoWorkspace(t *testing.T) {
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(oldWd) })

	_, err := ExecuteCommand(rootCmd, "new", "change", "FEATURE-500")
	assert.Error(t, err)
}

func TestNewChange_NoArgs(t *testing.T) {
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change")
	assert.Error(t, err)
}

func TestNewChange_UpdatesActiveOnSecondCreate(t *testing.T) {
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "FIRST")
	require.NoError(t, err)

	_, err = ExecuteCommand(rootCmd, "new", "change", "SECOND")
	require.NoError(t, err)

	// Active change should now be SECOND
	cfg, err := config.LoadLocal(filepath.Join(dir, ".sdlaicrc"))
	require.NoError(t, err)
	assert.Equal(t, "SECOND", cfg.ActiveChange)
}

func TestNewChange_IgnoredStorage(t *testing.T) {
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(oldWd) })

	resetInitFlags()
	_, err := ExecuteCommand(rootCmd, "init", "--storage", "ignored")
	require.NoError(t, err)

	_, err = ExecuteCommand(rootCmd, "new", "change", "CHANGE-1")
	require.NoError(t, err)

	// Change should be in .sdlaic/changes/
	changeDir := filepath.Join(dir, ".sdlaic", "changes", "CHANGE-1")
	_, err = os.Stat(changeDir)
	assert.NoError(t, err)
}

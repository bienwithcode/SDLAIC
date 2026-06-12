package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sdlaic/internal/config"
	"sdlaic/internal/domain"
)

func TestInit_CreatesWorkspace(t *testing.T) {
	resetInitFlags()
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldWd)

	_, err := ExecuteCommand(rootCmd, "init")
	require.NoError(t, err)

	// .sdlaicrc should exist
	_, err = os.Stat(filepath.Join(dir, ".sdlaicrc"))
	assert.NoError(t, err)

	// Default: local mode → .sdlaic/changes/ should exist
	_, err = os.Stat(filepath.Join(dir, ".sdlaic", "changes"))
	assert.NoError(t, err)
}

func TestInit_RejectsReInit(t *testing.T) {
	resetInitFlags()
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldWd)

	// First init should succeed
	_, err := ExecuteCommand(rootCmd, "init")
	require.NoError(t, err)

	// Second init should fail
	_, err = ExecuteCommand(rootCmd, "init")
	assert.Error(t, err)
}

func TestInit_StorageLocal(t *testing.T) {
	resetInitFlags()
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldWd)

	_, err := ExecuteCommand(rootCmd, "init", "--storage", "local")
	require.NoError(t, err)

	// Should create .sdlaic/changes/
	changesDir := filepath.Join(dir, ".sdlaic", "changes")
	_, err = os.Stat(changesDir)
	assert.NoError(t, err)

	// Config should reflect local storage
	cfg, err := config.LoadLocal(filepath.Join(dir, ".sdlaicrc"))
	require.NoError(t, err)
	assert.Equal(t, domain.StorageModeLocal, cfg.Storage)
}

func TestInit_StorageIgnored(t *testing.T) {
	resetInitFlags()
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldWd)

	_, err := ExecuteCommand(rootCmd, "init", "--storage", "ignored")
	require.NoError(t, err)

	// Should create .sdlaic/changes/
	changesDir := filepath.Join(dir, ".sdlaic", "changes")
	_, err = os.Stat(changesDir)
	assert.NoError(t, err)

	// Should add to .gitignore
	gitignoreData, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	require.NoError(t, err)
	assert.Contains(t, string(gitignoreData), ".sdlaic/changes/")

	// Config should reflect ignored storage
	cfg, err := config.LoadLocal(filepath.Join(dir, ".sdlaicrc"))
	require.NoError(t, err)
	assert.Equal(t, domain.StorageModeIgnored, cfg.Storage)
}

func TestInit_StorageGlobal(t *testing.T) {
	resetInitFlags()
	dir := t.TempDir()
	homeDir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldWd)

	_, err := ExecuteCommand(rootCmd, "init", "--storage", "global", "--home", homeDir)
	require.NoError(t, err)

	// .sdlaicrc should exist in project root
	_, err = os.Stat(filepath.Join(dir, ".sdlaicrc"))
	assert.NoError(t, err)

	// Config should reflect global storage
	cfg, err := config.LoadLocal(filepath.Join(dir, ".sdlaicrc"))
	require.NoError(t, err)
	assert.Equal(t, domain.StorageModeGlobal, cfg.Storage)
}

func TestInit_WorkflowFlag(t *testing.T) {
	resetInitFlags()
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldWd)

	_, err := ExecuteCommand(rootCmd, "init", "--workflow", "light")
	require.NoError(t, err)

	cfg, err := config.LoadLocal(filepath.Join(dir, ".sdlaicrc"))
	require.NoError(t, err)
	assert.Equal(t, domain.WorkflowLight, cfg.Workflow)
}

func TestInit_Defaults(t *testing.T) {
	resetInitFlags()
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldWd)

	_, err := ExecuteCommand(rootCmd, "init")
	require.NoError(t, err)

	cfg, err := config.LoadLocal(filepath.Join(dir, ".sdlaicrc"))
	require.NoError(t, err)
	assert.Equal(t, domain.StorageModeLocal, cfg.Storage)
	assert.Equal(t, domain.WorkflowStrict, cfg.Workflow)
	assert.Equal(t, 1, cfg.Version)
	assert.NotEmpty(t, cfg.ProjectHash)
}

func TestInit_InvalidStorage(t *testing.T) {
	resetInitFlags()
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldWd)

	_, err := ExecuteCommand(rootCmd, "init", "--storage", "cloud")
	assert.Error(t, err)
}

func TestInit_InvalidWorkflow(t *testing.T) {
	resetInitFlags()
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldWd)

	_, err := ExecuteCommand(rootCmd, "init", "--workflow", "ultra")
	assert.Error(t, err)
}

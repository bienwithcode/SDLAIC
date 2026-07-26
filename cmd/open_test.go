package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

func TestOpenClaude_PrintOnly_AlreadyInitialized(t *testing.T) {
	resetOpenFlags()
	dir := TempWorkspace(t)
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldWd)

	output, err := ExecuteCommand(rootCmd, "open", "claude", "--print", "--no-spawn")
	require.NoError(t, err)

	assert.Contains(t, output, "claude plugin marketplace add bienwithcode/SDLAIC")
	assert.Contains(t, output, "claude plugin install sdlaic@bienwithcode")
	assert.NotContains(t, output, "Workspace is not initialized")
	assert.NotContains(t, output, "Auto-initialized SDLAIC workspace")
}

func TestOpenClaude_PrintOnly_AutoInitDefaults(t *testing.T) {
	resetOpenFlags()
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldWd)

	// Stub home directory for testing global config updates
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	output, err := ExecuteCommand(rootCmd, "open", "claude", "--print", "--no-spawn")
	require.NoError(t, err)

	assert.Contains(t, output, "claude plugin marketplace add bienwithcode/SDLAIC")
	assert.Contains(t, output, "claude plugin install sdlaic@bienwithcode")
	assert.Contains(t, output, "Auto-initialized SDLAIC workspace")

	// Verify workspace files were created
	cfgPath := filepath.Join(dir, ".sdlaicrc")
	_, err = os.Stat(cfgPath)
	assert.NoError(t, err)

	data, err := os.ReadFile(cfgPath)
	require.NoError(t, err)

	var cfg domain.LocalConfig
	require.NoError(t, json.Unmarshal(data, &cfg))
	assert.Equal(t, domain.StorageModeLocal, cfg.Storage)
	assert.Equal(t, domain.WorkflowStrict, cfg.Workflow)
}

func TestOpenClaude_PrintOnly_AutoInitOverrides(t *testing.T) {
	resetOpenFlags()
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldWd)

	// Stub home directory for testing global config updates
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	output, err := ExecuteCommand(rootCmd, "open", "claude", "--print", "--no-spawn", "--storage=ignored", "--workflow=light")
	require.NoError(t, err)

	assert.Contains(t, output, "claude plugin marketplace add bienwithcode/SDLAIC")
	assert.Contains(t, output, "claude plugin install sdlaic@bienwithcode")
	assert.Contains(t, output, "Auto-initialized SDLAIC workspace")

	// Verify workspace files were created with correct overrides
	cfgPath := filepath.Join(dir, ".sdlaicrc")
	data, err := os.ReadFile(cfgPath)
	require.NoError(t, err)

	var cfg domain.LocalConfig
	require.NoError(t, json.Unmarshal(data, &cfg))
	assert.Equal(t, domain.StorageModeIgnored, cfg.Storage)
	assert.Equal(t, domain.WorkflowLight, cfg.Workflow)

	// Gitignore should have changes folder appended
	gitIgnorePath := filepath.Join(dir, ".gitignore")
	gitIgnoreData, err := os.ReadFile(gitIgnorePath)
	assert.NoError(t, err)
	assert.Contains(t, string(gitIgnoreData), ".sdlaic/changes/")
}

func TestOpenClaude_PrintOnly_MarketplaceOverride(t *testing.T) {
	resetOpenFlags()
	dir := TempWorkspace(t)
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldWd)

	output, err := ExecuteCommand(rootCmd, "open", "claude", "--print", "--no-spawn", "--marketplace=myuser/my-sdlaic-fork")
	require.NoError(t, err)

	assert.Contains(t, output, "claude plugin marketplace add myuser/my-sdlaic-fork")
	assert.Contains(t, output, "claude plugin install sdlaic@myuser")
}

func TestOpenCodex_Stub(t *testing.T) {
	resetOpenFlags()
	output, err := ExecuteCommand(rootCmd, "open", "codex")
	require.NoError(t, err)
	assert.Contains(t, output, "coming in a later release")
}

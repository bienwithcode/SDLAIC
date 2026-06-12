package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sdlaic/internal/domain"
)

// --- Testutil helpers ---

// TempWorkspace creates a temp directory with a valid .sdlaicrc and change directories.
func TempWorkspace(t *testing.T, storage domain.StorageMode) string {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".sdlaicrc")

	cfg := domain.NewLocalConfig(storage, domain.WorkflowStrict, "testhash123")
	data, err := json.MarshalIndent(cfg, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(cfgPath, data, 0644))

	// Create changes directory based on storage mode
	var changesDir string
	switch storage {
	case domain.StorageModeLocal:
		changesDir = filepath.Join(dir, ".sdlaic", "changes")
	case domain.StorageModeIgnored:
		changesDir = filepath.Join(dir, ".sdlaic", "changes")
	case domain.StorageModeGlobal:
		changesDir = filepath.Join(dir, "changes")
	}
	require.NoError(t, os.MkdirAll(changesDir, 0755))

	return dir
}

// TempChange creates a change directory with optional artifact files.
func TempChange(t *testing.T, workspaceDir string, storage domain.StorageMode, name string, artifacts map[string]string) string {
	t.Helper()
	var changesDir string
	switch storage {
	case domain.StorageModeLocal:
		changesDir = filepath.Join(workspaceDir, ".sdlaic", "changes")
	case domain.StorageModeIgnored:
		changesDir = filepath.Join(workspaceDir, ".sdlaic", "changes")
	default:
		changesDir = filepath.Join(workspaceDir, "changes")
	}

	changeDir := filepath.Join(changesDir, name)
	require.NoError(t, os.MkdirAll(changeDir, 0755))

	for fileName, content := range artifacts {
		require.NoError(t, os.WriteFile(filepath.Join(changeDir, fileName), []byte(content), 0644))
	}

	return changeDir
}

// --- Tests for root command ---

func TestRootCmd_Help(t *testing.T) {
	output, err := ExecuteCommand(rootCmd, "help")
	require.NoError(t, err)
	assert.Contains(t, output, "SDLAIC")
	assert.Contains(t, output, "spec-driven")
}

func TestRootCmd_NoArgs(t *testing.T) {
	output, err := ExecuteCommand(rootCmd)
	require.NoError(t, err)
	// Root command with no args shows the Long description
	assert.Contains(t, output, "SDLAIC")
}

func TestRootCmd_Version(t *testing.T) {
	// The version command is not yet added, but root should handle --version
	output, err := ExecuteCommand(rootCmd, "--version")
	require.NoError(t, err)
	assert.Contains(t, output, "dev")
}

func TestResolveChangeName_FlagProvided(t *testing.T) {
	name, err := resolveChangeName("my-change")
	require.NoError(t, err)
	assert.Equal(t, "my-change", name)
}

func TestResolveChangeName_EmptyFlag_NoWorkspace(t *testing.T) {
	// Without workspace context, empty flag should error
	_, err := resolveChangeName("")
	assert.Error(t, err)
}

func TestResolveChangeName_EmptyFlag_WithActiveChange(t *testing.T) {
	dir := TempWorkspace(t, domain.StorageModeLocal)

	// Set active change in config
	cfgPath := filepath.Join(dir, ".sdlaicrc")
	cfgData, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	var cfg domain.LocalConfig
	require.NoError(t, json.Unmarshal(cfgData, &cfg))
	cfg.ActiveChange = "ACTIVE-CHANGE"
	cfgBytes, err := json.MarshalIndent(cfg, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(cfgPath, cfgBytes, 0644))

	// Override the workspace root for this test
	oldRoot := workspaceRoot
	workspaceRoot = dir
	defer func() { workspaceRoot = oldRoot }()

	name, err := resolveChangeName("")
	require.NoError(t, err)
	assert.Equal(t, "ACTIVE-CHANGE", name)
}

// tempDirNoWorkspace creates a temp directory without a workspace.
func tempDirNoWorkspace(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// --- ExecuteCommand helper test ---

func TestExecuteCommand(t *testing.T) {
	output, err := ExecuteCommand(rootCmd, "help")
	require.NoError(t, err)
	assert.NotEmpty(t, output)
}

package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bienwithcode/SDLAIC/internal/config"
	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/bienwithcode/SDLAIC/internal/storage"
)

// --- Testutil helpers ---

// TempWorkspace creates a project directory registered in a temp global config,
// and points the CLI's home at that temp directory for the duration of the test
// so nothing touches the developer's real ~/.sdlaic.
func TempWorkspace(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	dir := t.TempDir()

	oldHome := homeFlag
	homeFlag = home
	t.Cleanup(func() { homeFlag = oldHome })

	changesDir := storage.DefaultChangesDir(dir)
	require.NoError(t, os.MkdirAll(changesDir, 0755))

	hash, err := workspaceHash(dir)
	require.NoError(t, err)
	require.NoError(t, config.UpdateProject(globalConfigPath(), hash, func(e *domain.ProjectEntry) {
		e.Path = dir
		e.ChangesDir = changesDir
		e.Workflow = domain.WorkflowStrict
	}))

	// TEMPORARY: also write a .sdlaicrc so commands that have not been migrated
	// yet still resolve. Removed in T17 with the production fallback.
	local := domain.NewLocalConfig(domain.StorageModeLocal, domain.WorkflowStrict, hash)
	require.NoError(t, config.SaveLocal(local, filepath.Join(dir, ".sdlaicrc")))

	return dir
}

// TempChange creates a change directory with optional artifact files.
func TempChange(t *testing.T, workspaceDir string, name string, artifacts map[string]string) string {
	t.Helper()
	changeDir := filepath.Join(storage.DefaultChangesDir(workspaceDir), name)
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
	dir := TempWorkspace(t)

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

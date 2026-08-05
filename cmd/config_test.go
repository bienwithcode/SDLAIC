package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bienwithcode/SDLAIC/internal/config"
	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// entryFor reads the current project entry from the temp global config.
func entryFor(t *testing.T, dir string) domain.ProjectEntry {
	t.Helper()
	cfg, err := config.LoadGlobal(globalConfigPath())
	require.NoError(t, err)
	hash, err := workspaceHash(dir)
	require.NoError(t, err)
	return cfg.Projects[hash]
}

func initWorkspaceForTest(t *testing.T) string {
	t.Helper()
	home := t.TempDir()

	// os.Getwd reports the symlink-free path, so resolve up front or path
	// comparisons against the returned dir will not match.
	dir, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)

	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(oldWd) })

	resetInitFlags()
	t.Cleanup(resetInitFlags)
	_, err = ExecuteCommand(rootCmd, "init", "--home", home)
	require.NoError(t, err)
	return dir
}

func TestConfigSet_ChangesDir(t *testing.T) {
	dir := initWorkspaceForTest(t)
	external := filepath.Join(t.TempDir(), "openspec", "changes")

	_, err := ExecuteCommand(rootCmd, "config", "set", "changes-dir", external)
	require.NoError(t, err)

	assert.Equal(t, external, entryFor(t, dir).ChangesDir)
	assert.DirExists(t, external)
}

func TestConfigSet_ChangesDirIsStoredAbsolute(t *testing.T) {
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "config", "set", "changes-dir", "spec/changes")
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(dir, "spec", "changes"), entryFor(t, dir).ChangesDir)
}

func TestConfigSet_RejectsChangesDirClaimedByAnotherProject(t *testing.T) {
	first := initWorkspaceForTest(t)
	shared := filepath.Join(t.TempDir(), "shared", "changes")
	_, err := ExecuteCommand(rootCmd, "config", "set", "changes-dir", shared)
	require.NoError(t, err)

	// A second project in the same home cannot claim the same directory.
	second, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, os.Chdir(second))
	_, err = ExecuteCommand(rootCmd, "init", "--home", homeFlag)
	require.NoError(t, err)

	_, err = ExecuteCommand(rootCmd, "config", "set", "changes-dir", shared)

	require.Error(t, err)
	assert.Contains(t, err.Error(), first, "the error should name the project already using it")
}

func TestConfigSet_StorageKeyIsGone(t *testing.T) {
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "config", "set", "storage", "ignored")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "changes-dir", "the error should list the keys that do exist")
}

func TestConfigSet_Workflow(t *testing.T) {
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "config", "set", "workflow", "free")
	require.NoError(t, err)

	assert.Equal(t, domain.WorkflowFree, entryFor(t, dir).Workflow)
}

func TestConfigSet_InvalidKey(t *testing.T) {
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "config", "set", "color", "blue")
	assert.Error(t, err)
}

func TestConfigSet_InvalidWorkflowValue(t *testing.T) {
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "config", "set", "workflow", "turbo")
	assert.Error(t, err)
}

func TestConfigSet_RequiresTwoArgs(t *testing.T) {
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "config", "set", "workflow")
	assert.Error(t, err)
}

func TestConfigList(t *testing.T) {
	_ = initWorkspaceForTest(t)

	output, err := ExecuteCommand(rootCmd, "config", "list")
	require.NoError(t, err)
	assert.Contains(t, output, "changes_dir")
	// The printed path uses the platform separator: .sdlaic\changes on Windows.
	assert.Contains(t, output, filepath.Join(".sdlaic", "changes"))
	assert.Contains(t, output, "workflow")
	assert.Contains(t, output, "strict")
	assert.NotContains(t, output, "storage")
}

func TestConfigList_AfterSet(t *testing.T) {
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "config", "set", "workflow", "light")
	require.NoError(t, err)

	output, err := ExecuteCommand(rootCmd, "config", "list")
	require.NoError(t, err)
	assert.True(t, strings.Contains(output, "light"))
}

func TestConfigSet_NoWorkspace(t *testing.T) {
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(oldWd) })

	// No init — no workspace
	_, err := ExecuteCommand(rootCmd, "config", "set", "storage", "local")
	assert.Error(t, err)
}

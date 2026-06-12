package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initWorkspaceForTest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(oldWd) })

	resetInitFlags()
	_, err := ExecuteCommand(rootCmd, "init", "--storage", "local")
	require.NoError(t, err)
	return dir
}

func TestConfigSet_Storage(t *testing.T) {
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "config", "set", "storage", "ignored")
	require.NoError(t, err)

	// Verify config updated
	data, err := os.ReadFile(filepath.Join(dir, ".sdlaicrc"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "ignored")

	// Verify .gitignore updated
	gitignoreData, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	require.NoError(t, err)
	assert.Contains(t, string(gitignoreData), ".sdlaic/changes/")
}

func TestConfigSet_Workflow(t *testing.T) {
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "config", "set", "workflow", "free")
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, ".sdlaicrc"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "free")
}

func TestConfigSet_InvalidKey(t *testing.T) {
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "config", "set", "color", "blue")
	assert.Error(t, err)
}

func TestConfigSet_InvalidStorageValue(t *testing.T) {
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "config", "set", "storage", "cloud")
	assert.Error(t, err)
}

func TestConfigSet_InvalidWorkflowValue(t *testing.T) {
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "config", "set", "workflow", "turbo")
	assert.Error(t, err)
}

func TestConfigSet_RequiresTwoArgs(t *testing.T) {
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "config", "set", "storage")
	assert.Error(t, err)
}

func TestConfigList(t *testing.T) {
	_ = initWorkspaceForTest(t)

	output, err := ExecuteCommand(rootCmd, "config", "list")
	require.NoError(t, err)
	assert.Contains(t, output, "storage")
	assert.Contains(t, output, "local")
	assert.Contains(t, output, "workflow")
	assert.Contains(t, output, "strict")
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

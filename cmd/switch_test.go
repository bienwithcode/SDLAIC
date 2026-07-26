package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwitch_SetsActive(t *testing.T) {
	dir := initWorkspaceForTest(t)

	// Create two changes
	_, err := ExecuteCommand(rootCmd, "new", "change", "CHANGE-A")
	require.NoError(t, err)
	_, err = ExecuteCommand(rootCmd, "new", "change", "CHANGE-B")
	require.NoError(t, err)

	// Switch to A
	_, err = ExecuteCommand(rootCmd, "switch", "CHANGE-A")
	require.NoError(t, err)

	assert.Equal(t, "CHANGE-A", activeChangeOf(t, dir))
}

func TestSwitch_InvalidName(t *testing.T) {
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "switch", "NONEXISTENT")
	assert.Error(t, err)
}

func TestSwitch_NoWorkspace(t *testing.T) {
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(oldWd) })

	_, err := ExecuteCommand(rootCmd, "switch", "SOME-CHANGE")
	assert.Error(t, err)
}

func TestSwitch_NoArgs_NoChanges(t *testing.T) {
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "switch")
	assert.Error(t, err)
}

func TestSwitch_ShowsList(t *testing.T) {
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "CHANGE-X")
	require.NoError(t, err)
	_, err = ExecuteCommand(rootCmd, "new", "change", "CHANGE-Y")
	require.NoError(t, err)

	// switch with no args when there ARE changes should list them
	output, err := ExecuteCommand(rootCmd, "switch")
	// It should list available changes (interactive mode simulation)
	assert.NoError(t, err)
	assert.Contains(t, output, "CHANGE-X")
	assert.Contains(t, output, "CHANGE-Y")
}

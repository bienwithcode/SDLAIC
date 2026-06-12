package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList_ActiveOnly(t *testing.T) {
	resetStatusFlags()
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "ACTIVE-ONE")
	require.NoError(t, err)
	_, err = ExecuteCommand(rootCmd, "new", "change", "OTHER")
	require.NoError(t, err)

	resetStatusFlags()
	output, err := ExecuteCommand(rootCmd, "list")
	require.NoError(t, err)
	assert.Contains(t, output, "OTHER")
}

func TestList_IncludesArchived(t *testing.T) {
	resetStatusFlags()
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "KEPT")
	require.NoError(t, err)
	_, err = ExecuteCommand(rootCmd, "new", "change", "REMOVED")
	require.NoError(t, err)

	resetStatusFlags()
	_, err = ExecuteCommand(rootCmd, "archive", "REMOVED")
	require.NoError(t, err)

	resetStatusFlags()
	output, err := ExecuteCommand(rootCmd, "list", "--all")
	require.NoError(t, err)
	assert.Contains(t, output, "KEPT")
	assert.Contains(t, output, "REMOVED")
}

func TestList_JSONOutput(t *testing.T) {
	resetStatusFlags()
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "JSON-TEST")
	require.NoError(t, err)

	resetStatusFlags()
	output, err := ExecuteCommand(rootCmd, "list", "--json")
	require.NoError(t, err)
	assert.Contains(t, output, "JSON-TEST")
	assert.Contains(t, output, "changes")
}

func TestList_NoWorkspace(t *testing.T) {
	resetStatusFlags()
	dir := tempDirNoWorkspace(t)
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(oldWd) })

	_, err := ExecuteCommand(rootCmd, "list")
	assert.Error(t, err)
}

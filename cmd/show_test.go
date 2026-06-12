package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShow_ChangeDetails(t *testing.T) {
	resetStatusFlags()
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "SHOW-ME")
	require.NoError(t, err)

	resetStatusFlags()
	output, err := ExecuteCommand(rootCmd, "show", "SHOW-ME")
	require.NoError(t, err)
	assert.Contains(t, output, "SHOW-ME")
	assert.Contains(t, output, "context.md")
}

func TestShow_MissingChange(t *testing.T) {
	resetStatusFlags()
	_ = initWorkspaceForTest(t)

	resetStatusFlags()
	_, err := ExecuteCommand(rootCmd, "show", "NONEXISTENT")
	assert.Error(t, err)
}

func TestShow_NoArgs(t *testing.T) {
	resetStatusFlags()
	_ = initWorkspaceForTest(t)

	resetStatusFlags()
	_, err := ExecuteCommand(rootCmd, "show")
	assert.Error(t, err)
}

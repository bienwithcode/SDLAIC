package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletion_Bash(t *testing.T) {
	output, err := ExecuteCommand(rootCmd, "completion", "bash")
	require.NoError(t, err)
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "bash")
}

func TestCompletion_Zsh(t *testing.T) {
	output, err := ExecuteCommand(rootCmd, "completion", "zsh")
	require.NoError(t, err)
	assert.NotEmpty(t, output)
	// Zsh completions contain #compdef
	assert.Contains(t, output, "#compdef")
}

func TestCompletion_Fish(t *testing.T) {
	output, err := ExecuteCommand(rootCmd, "completion", "fish")
	require.NoError(t, err)
	assert.NotEmpty(t, output)
	// Fish completions contain the command name
	assert.True(t, strings.Contains(output, "sdlaic") || strings.Contains(output, "complete"))
}

func TestCompletion_PowerShell(t *testing.T) {
	output, err := ExecuteCommand(rootCmd, "completion", "powershell")
	require.NoError(t, err)
	assert.NotEmpty(t, output)
}

func TestCompletion_NoArgs(t *testing.T) {
	_, err := ExecuteCommand(rootCmd, "completion")
	assert.Error(t, err)
}

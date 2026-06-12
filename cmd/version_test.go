package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion_PrintsVersion(t *testing.T) {
	SetVersion("v0.1.0")
	output, err := ExecuteCommand(rootCmd, "version")
	require.NoError(t, err)
	assert.Contains(t, output, "sdlaic")
	assert.Contains(t, output, "v0.1.0")
}

func TestVersion_ShortFlag(t *testing.T) {
	SetVersion("v0.1.0")
	output, err := ExecuteCommand(rootCmd, "version", "--short")
	require.NoError(t, err)
	trimmed := strings.TrimSpace(output)
	assert.Equal(t, "v0.1.0", trimmed)
}

func TestVersion_DefaultDev(t *testing.T) {
	SetVersion("dev")
	output, err := ExecuteCommand(rootCmd, "version")
	require.NoError(t, err)
	assert.Contains(t, output, "dev")
}

func TestVersion_ShortFlagDev(t *testing.T) {
	SetVersion("dev")
	output, err := ExecuteCommand(rootCmd, "version", "--short")
	require.NoError(t, err)
	trimmed := strings.TrimSpace(output)
	assert.Equal(t, "dev", trimmed)
}

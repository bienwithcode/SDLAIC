package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveHome_UsesFlagWhenSet(t *testing.T) {
	t.Cleanup(func() { homeFlag = "" })
	homeFlag = "/tmp/flag-home"

	assert.Equal(t, "/tmp/flag-home", resolveHome())
}

func TestResolveHome_UsesEnvWhenFlagEmpty(t *testing.T) {
	t.Cleanup(func() { homeFlag = "" })
	homeFlag = ""
	t.Setenv("SDLAIC_HOME", "/tmp/env-home")

	assert.Equal(t, "/tmp/env-home", resolveHome())
}

func TestResolveHome_FlagBeatsEnv(t *testing.T) {
	t.Cleanup(func() { homeFlag = "" })
	homeFlag = "/tmp/flag-home"
	t.Setenv("SDLAIC_HOME", "/tmp/env-home")

	assert.Equal(t, "/tmp/flag-home", resolveHome())
}

func TestResolveHome_FallsBackToUserHome(t *testing.T) {
	t.Cleanup(func() { homeFlag = "" })
	homeFlag = ""
	t.Setenv("SDLAIC_HOME", "")

	want, err := os.UserHomeDir()
	require.NoError(t, err)
	assert.Equal(t, want, resolveHome())
}

func TestGlobalConfigPath_IsUnderResolvedHome(t *testing.T) {
	t.Cleanup(func() { homeFlag = "" })
	homeFlag = t.TempDir()

	assert.Equal(t, filepath.Join(homeFlag, ".sdlaic", "config.json"), globalConfigPath())
}

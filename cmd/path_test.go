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

// chdirTo enters dir for the duration of the test.
func chdirTo(t *testing.T, dir string) {
	t.Helper()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
}

// clearChangeFlag stops a --change value leaking into the next command run,
// since Cobra flag variables persist across ExecuteCommand calls.
func clearChangeFlag(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { changeFlag = "" })
	changeFlag = ""
}

func TestPathChanges_PrintsAbsoluteChangesDir(t *testing.T) {
	clearChangeFlag(t)
	dir := TempWorkspace(t)
	chdirTo(t, dir)

	out, err := ExecuteCommand(rootCmd, "path", "changes")
	require.NoError(t, err)

	got := strings.TrimSpace(out)
	assert.True(t, filepath.IsAbs(got), "output must be an absolute path")
	assert.Equal(t, filepath.Join(dir, ".sdlaic", "changes"), got)
}

func TestPathChanges_WorksFromNestedSubdirectory(t *testing.T) {
	clearChangeFlag(t)
	dir := TempWorkspace(t)
	deep := filepath.Join(dir, "internal", "config")
	require.NoError(t, os.MkdirAll(deep, 0755))
	chdirTo(t, deep)

	out, err := ExecuteCommand(rootCmd, "path", "changes")
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(dir, ".sdlaic", "changes"), strings.TrimSpace(out))
}

func TestPathChange_UsesExplicitChangeName(t *testing.T) {
	clearChangeFlag(t)
	dir := TempWorkspace(t)
	chdirTo(t, dir)

	out, err := ExecuteCommand(rootCmd, "path", "change", "--change", "SDL-1")
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(dir, ".sdlaic", "changes", "SDL-1"), strings.TrimSpace(out))
}

func TestPathChange_FallsBackToActiveChange(t *testing.T) {
	clearChangeFlag(t)
	dir := TempWorkspace(t)
	chdirTo(t, dir)

	hash, err := workspaceHash(dir)
	require.NoError(t, err)
	require.NoError(t, config.UpdateProject(globalConfigPath(), hash, func(e *domain.ProjectEntry) {
		e.ActiveChange = "SDL-ACTIVE"
	}))

	out, err := ExecuteCommand(rootCmd, "path", "change")
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(dir, ".sdlaic", "changes", "SDL-ACTIVE"), strings.TrimSpace(out))
}

func TestPathChange_NoActiveChangeIsAnError(t *testing.T) {
	clearChangeFlag(t)
	dir := TempWorkspace(t)
	chdirTo(t, dir)

	_, err := ExecuteCommand(rootCmd, "path", "change")

	assert.ErrorIs(t, err, domain.ErrNoActiveChange)
}

func TestPathChanges_UnregisteredProjectIsAnError(t *testing.T) {
	clearChangeFlag(t)
	home := t.TempDir()
	oldHome := homeFlag
	homeFlag = home
	t.Cleanup(func() { homeFlag = oldHome })
	chdirTo(t, t.TempDir())

	_, err := ExecuteCommand(rootCmd, "path", "changes")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "sdlaic init", "the error should tell the user how to fix it")
}

func TestPathChanges_UnconfiguredChangesDirIsAnError(t *testing.T) {
	clearChangeFlag(t)
	home := t.TempDir()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	oldHome := homeFlag
	homeFlag = home
	t.Cleanup(func() { homeFlag = oldHome })

	hash, err := workspaceHash(dir)
	require.NoError(t, err)
	require.NoError(t, config.UpdateProject(globalConfigPath(), hash, func(e *domain.ProjectEntry) {
		e.Path = dir // registered, but never given a changes dir
	}))
	chdirTo(t, dir)

	_, err = ExecuteCommand(rootCmd, "path", "changes")

	assert.ErrorIs(t, err, domain.ErrChangesDirNotSet)
}

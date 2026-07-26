package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bienwithcode/SDLAIC/internal/config"
	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// initFixture chdirs into a fresh project directory with an isolated home, and
// returns both paths. Flags are reset so a previous test's values cannot leak.
func initFixture(t *testing.T) (home string, dir string) {
	t.Helper()
	resetInitFlags()
	home = t.TempDir()

	// Resolve the project dir the same way the CLI does: os.Getwd returns the
	// symlink-free form, so a raw t.TempDir() would not compare equal.
	dir, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
		resetInitFlags()
	})
	return home, dir
}

// globalEntry reads back the single project entry registered for dir.
func globalEntry(t *testing.T, home string, dir string) domain.ProjectEntry {
	t.Helper()
	cfg, err := config.LoadGlobal(filepath.Join(home, ".sdlaic", "config.json"))
	require.NoError(t, err)
	hash, err := workspaceHash(dir)
	require.NoError(t, err)
	entry, ok := cfg.Projects[hash]
	require.True(t, ok, "project should be registered in the global config")
	return entry
}

func TestInit_RegistersProjectWithDefaultChangesDir(t *testing.T) {
	home, dir := initFixture(t)

	_, err := ExecuteCommand(rootCmd, "init", "--home", home)
	require.NoError(t, err)

	entry := globalEntry(t, home, dir)
	assert.Equal(t, filepath.Join(dir, ".sdlaic", "changes"), entry.ChangesDir)
	assert.Equal(t, domain.WorkflowStrict, entry.Workflow)
	assert.DirExists(t, filepath.Join(dir, ".sdlaic", "changes"))
}

func TestInit_ExternalChangesDirLeavesProjectUntouched(t *testing.T) {
	home, dir := initFixture(t)
	external := filepath.Join(t.TempDir(), "openspec", "changes")

	_, err := ExecuteCommand(rootCmd, "init", "--home", home, "--changes-dir", external)
	require.NoError(t, err)

	entry := globalEntry(t, home, dir)
	assert.Equal(t, external, entry.ChangesDir)
	assert.DirExists(t, external)

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Empty(t, entries, "an external changes dir must leave the project directory completely empty")
}

func TestInit_NormalizesRelativeChangesDir(t *testing.T) {
	home, dir := initFixture(t)

	_, err := ExecuteCommand(rootCmd, "init", "--home", home, "--changes-dir", "spec/changes")
	require.NoError(t, err)

	entry := globalEntry(t, home, dir)
	assert.True(t, filepath.IsAbs(entry.ChangesDir), "stored paths are always absolute")
	assert.Equal(t, filepath.Join(dir, "spec", "changes"), entry.ChangesDir)
}

func TestInit_IsIdempotent(t *testing.T) {
	home, dir := initFixture(t)

	_, err := ExecuteCommand(rootCmd, "init", "--home", home)
	require.NoError(t, err)
	_, err = ExecuteCommand(rootCmd, "init", "--home", home)
	require.NoError(t, err, "re-running init on a registered project must succeed")

	cfg, err := config.LoadGlobal(filepath.Join(home, ".sdlaic", "config.json"))
	require.NoError(t, err)
	assert.Len(t, cfg.Projects, 1, "re-init updates the entry rather than adding another")
	_ = dir
}

func TestInit_RejectsChangesDirClaimedByAnotherProject(t *testing.T) {
	home, _ := initFixture(t)
	shared := filepath.Join(t.TempDir(), "openspec", "changes")

	_, err := ExecuteCommand(rootCmd, "init", "--home", home, "--changes-dir", shared)
	require.NoError(t, err)

	other := t.TempDir()
	require.NoError(t, os.Chdir(other))
	output, err := ExecuteCommand(rootCmd, "init", "--home", home, "--changes-dir", shared)

	require.Error(t, err, "one changes directory belongs to exactly one project")
	assert.Contains(t, err.Error()+output, "already", "the error should say the directory is taken")
}

func TestInit_WorkflowFlag(t *testing.T) {
	home, dir := initFixture(t)

	_, err := ExecuteCommand(rootCmd, "init", "--home", home, "--workflow", "light")
	require.NoError(t, err)

	assert.Equal(t, domain.WorkflowLight, globalEntry(t, home, dir).Workflow)
}

func TestInit_InvalidWorkflow(t *testing.T) {
	home, _ := initFixture(t)

	_, err := ExecuteCommand(rootCmd, "init", "--home", home, "--workflow", "ultra")
	assert.Error(t, err)
}

func TestInit_RejectsEmptyChangesDir(t *testing.T) {
	home, _ := initFixture(t)

	_, err := ExecuteCommand(rootCmd, "init", "--home", home, "--changes-dir", "  ")
	assert.Error(t, err)
}

func TestInit_WritesNothingIntoTheProjectButTheChangesDir(t *testing.T) {
	home, dir := initFixture(t)

	_, err := ExecuteCommand(rootCmd, "init", "--home", home)
	require.NoError(t, err)

	assert.NoFileExists(t, filepath.Join(dir, ".sdlaicrc"), "no project-local state file survives")
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, ".sdlaic", entries[0].Name())
}

func TestInit_RejectsChangesDirReachedThroughSymlink(t *testing.T) {
	home, _ := initFixture(t)

	// One physical directory, two names for it.
	base := t.TempDir()
	real := filepath.Join(base, "real", "changes")
	require.NoError(t, os.MkdirAll(real, 0755))
	require.NoError(t, os.Symlink(filepath.Join(base, "real"), filepath.Join(base, "link")))

	_, err := ExecuteCommand(rootCmd, "init", "--home", home, "--changes-dir", real)
	require.NoError(t, err)

	other, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, os.Chdir(other))

	_, err = ExecuteCommand(rootCmd, "init", "--home", home, "--changes-dir", filepath.Join(base, "link", "changes"))

	require.Error(t, err, "a second name for the same directory must be rejected, or two projects end up sharing one store")
}

func TestInit_RejectsChangesDirThroughSymlinkedParentBeforeItExists(t *testing.T) {
	home, _ := initFixture(t)

	base := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(base, "real"), 0755))
	require.NoError(t, os.Symlink(filepath.Join(base, "real"), filepath.Join(base, "link")))

	// Neither changes dir exists yet; the collision is only visible after
	// canonicalising through the symlinked parent.
	_, err := ExecuteCommand(rootCmd, "init", "--home", home, "--changes-dir", filepath.Join(base, "real", "changes"))
	require.NoError(t, err)

	other, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, os.Chdir(other))

	_, err = ExecuteCommand(rootCmd, "init", "--home", home, "--changes-dir", filepath.Join(base, "link", "changes"))

	require.Error(t, err)
}

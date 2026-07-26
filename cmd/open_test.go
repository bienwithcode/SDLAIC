package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// initFixtureUnregistered chdirs into a fresh, unregistered project with an
// isolated home.
func initFixtureUnregistered(t *testing.T) (home string, dir string) {
	t.Helper()
	home = t.TempDir()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)

	oldHome := homeFlag
	homeFlag = home
	t.Cleanup(func() { homeFlag = oldHome })
	chdirTo(t, dir)
	return home, dir
}

func TestOpenClaude_PrintOnly_AlreadyInitialized(t *testing.T) {
	resetOpenFlags()
	dir := TempWorkspace(t)
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldWd)

	output, err := ExecuteCommand(rootCmd, "open", "claude", "--print", "--no-spawn")
	require.NoError(t, err)

	assert.Contains(t, output, "claude plugin marketplace add bienwithcode/SDLAIC")
	assert.Contains(t, output, "claude plugin install sdlaic@bienwithcode")
	assert.NotContains(t, output, "not set up yet")
	assert.NotContains(t, output, "Configured SDLAIC project")
}

func TestOpenClaude_PrintOnly_ConfiguresWithDefaults(t *testing.T) {
	resetOpenFlags()
	home, dir := initFixture(t)
	homeFlag = home

	output, err := ExecuteCommand(rootCmd, "open", "claude", "--print", "--no-spawn", "--home", home)
	require.NoError(t, err)

	assert.Contains(t, output, "claude plugin marketplace add bienwithcode/SDLAIC")
	assert.Contains(t, output, "Configured SDLAIC project")

	entry := globalEntry(t, home, dir)
	assert.Equal(t, filepath.Join(dir, ".sdlaic", "changes"), entry.ChangesDir)
	assert.Equal(t, domain.WorkflowStrict, entry.Workflow)
}

func TestOpenClaude_PrintOnly_HonoursFlags(t *testing.T) {
	resetOpenFlags()
	home, dir := initFixture(t)
	homeFlag = home
	external := filepath.Join(t.TempDir(), "openspec", "changes")

	_, err := ExecuteCommand(rootCmd, "open", "claude", "--print", "--no-spawn",
		"--home", home, "--changes-dir", external, "--workflow=light")
	require.NoError(t, err)

	entry := globalEntry(t, home, dir)
	assert.Equal(t, external, entry.ChangesDir)
	assert.Equal(t, domain.WorkflowLight, entry.Workflow)

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Empty(t, entries, "an external changes dir leaves the project untouched")
}

func TestOpenClaude_PrintOnly_MarketplaceOverride(t *testing.T) {
	resetOpenFlags()
	dir := TempWorkspace(t)
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldWd)

	output, err := ExecuteCommand(rootCmd, "open", "claude", "--print", "--no-spawn", "--marketplace=myuser/my-sdlaic-fork")
	require.NoError(t, err)

	assert.Contains(t, output, "claude plugin marketplace add myuser/my-sdlaic-fork")
	assert.Contains(t, output, "claude plugin install sdlaic@myuser")
}

func TestOpenCodex_Stub(t *testing.T) {
	resetOpenFlags()
	output, err := ExecuteCommand(rootCmd, "open", "codex")
	require.NoError(t, err)
	assert.Contains(t, output, "coming in a later release")
}

// --- ensureProjectConfigured: prompt behaviour ---
//
// Driven through a scripted reader rather than a real TTY, so the interactive
// path is covered without the test needing a terminal.

func TestEnsureProjectConfigured_PromptsForDefaultLayout(t *testing.T) {
	resetOpenFlags()
	home, dir := initFixtureUnregistered(t)

	// "default" changes dir, then "light" workflow.
	require.NoError(t, ensureProjectConfigured(rootCmd, dir, strings.NewReader("default\nlight\n"), true))

	entry := globalEntry(t, home, dir)
	assert.Equal(t, filepath.Join(dir, ".sdlaic", "changes"), entry.ChangesDir)
	assert.Equal(t, domain.WorkflowLight, entry.Workflow)
}

func TestEnsureProjectConfigured_PromptsForCustomPath(t *testing.T) {
	resetOpenFlags()
	home, dir := initFixtureUnregistered(t)
	custom := filepath.Join(t.TempDir(), "openspec", "changes")

	require.NoError(t, ensureProjectConfigured(rootCmd, dir, strings.NewReader("custom\n"+custom+"\nstrict\n"), true))

	assert.Equal(t, custom, globalEntry(t, home, dir).ChangesDir)
}

func TestEnsureProjectConfigured_NonInteractiveUsesDefaults(t *testing.T) {
	resetOpenFlags()
	home, dir := initFixtureUnregistered(t)

	// An empty reader would block a prompt; non-interactive must not ask at all.
	require.NoError(t, ensureProjectConfigured(rootCmd, dir, strings.NewReader(""), false))

	entry := globalEntry(t, home, dir)
	assert.Equal(t, filepath.Join(dir, ".sdlaic", "changes"), entry.ChangesDir)
	assert.Equal(t, domain.WorkflowStrict, entry.Workflow)
}

func TestEnsureProjectConfigured_SkipsPromptWhenAlreadyConfigured(t *testing.T) {
	resetOpenFlags()
	dir := TempWorkspace(t)
	chdirTo(t, dir)

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)

	require.NoError(t, ensureProjectConfigured(cmd, dir, strings.NewReader(""), true))

	assert.Empty(t, out.String(), "a configured project must not be asked anything")
}

func TestEnsureProjectConfigured_FlagsBeatPrompts(t *testing.T) {
	resetOpenFlags()
	home, dir := initFixtureUnregistered(t)
	external := filepath.Join(t.TempDir(), "flagged", "changes")
	openChangesDir = external
	openWorkflow = "free"
	t.Cleanup(resetOpenFlags)

	require.NoError(t, ensureProjectConfigured(rootCmd, dir, strings.NewReader(""), true))

	entry := globalEntry(t, home, dir)
	assert.Equal(t, external, entry.ChangesDir)
	assert.Equal(t, domain.WorkflowFree, entry.Workflow)
}

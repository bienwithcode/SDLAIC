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

// projectFixture creates a temp home + project directory, chdirs into the
// project, and points the CLI at the temp home for the duration of the test.
func projectFixture(t *testing.T) (home string, root string) {
	t.Helper()
	home = t.TempDir()
	root = t.TempDir()

	oldHome := homeFlag
	homeFlag = home
	t.Cleanup(func() { homeFlag = oldHome })

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	return home, root
}

// registerGlobal writes a single project entry into the temp home's config.
func registerGlobal(t *testing.T, root string, entry domain.ProjectEntry) string {
	t.Helper()
	hash, err := workspaceHash(root)
	require.NoError(t, err)
	entry.Path = root
	require.NoError(t, config.UpdateProject(globalConfigPath(), hash, func(e *domain.ProjectEntry) {
		*e = entry
	}))
	return hash
}

// writeLocalConfig writes a legacy .sdlaicrc into root.
func writeLocalConfig(t *testing.T, root string, cfg domain.LocalConfig) {
	t.Helper()
	require.NoError(t, config.SaveLocal(cfg, filepath.Join(root, ".sdlaicrc")))
}

func TestResolveProject_ReadsGlobalEntry(t *testing.T) {
	_, root := projectFixture(t)
	registerGlobal(t, root, domain.ProjectEntry{
		ChangesDir:   "/work/openspec/changes",
		Workflow:     domain.WorkflowLight,
		ActiveChange: "SDL-7",
	})

	ctx, err := resolveProject()

	require.NoError(t, err)
	assert.Equal(t, "/work/openspec/changes", ctx.ChangesDir)
	assert.Equal(t, domain.WorkflowLight, ctx.Workflow)
	assert.Equal(t, "SDL-7", ctx.ActiveChange)
	assert.False(t, ctx.fromLocalConfig)
}

func TestResolveProject_GlobalEntryWinsOverLocalConfig(t *testing.T) {
	_, root := projectFixture(t)
	writeLocalConfig(t, root, domain.NewLocalConfig(domain.StorageModeLocal, domain.WorkflowStrict, "legacyhash"))
	registerGlobal(t, root, domain.ProjectEntry{ChangesDir: "/work/openspec/changes"})

	ctx, err := resolveProject()

	require.NoError(t, err)
	assert.Equal(t, "/work/openspec/changes", ctx.ChangesDir, "the global entry takes precedence over .sdlaicrc")
	assert.False(t, ctx.fromLocalConfig)
}

func TestResolveProject_FallsBackToLocalConfig(t *testing.T) {
	_, root := projectFixture(t)
	writeLocalConfig(t, root, domain.NewLocalConfig(domain.StorageModeLocal, domain.WorkflowFree, "legacyhash"))

	ctx, err := resolveProject()

	require.NoError(t, err)
	assert.True(t, ctx.fromLocalConfig, "an unmigrated project still resolves via .sdlaicrc")
	assert.Equal(t, filepath.Join(ctx.Root, ".sdlaic", "changes"), ctx.ChangesDir)
	assert.Equal(t, domain.WorkflowFree, ctx.Workflow)
	assert.Equal(t, "legacyhash", ctx.Hash)
}

func TestResolveProject_UnregisteredDirectoryIsNotFound(t *testing.T) {
	projectFixture(t)

	_, err := resolveProject()

	assert.ErrorIs(t, err, domain.ErrWorkspaceNotFound)
}

func TestResolveProject_EmptyChangesDirIsReportedNotGuessed(t *testing.T) {
	_, root := projectFixture(t)
	registerGlobal(t, root, domain.ProjectEntry{Workflow: domain.WorkflowStrict})

	ctx, err := resolveProject()
	require.NoError(t, err)
	assert.Empty(t, ctx.ChangesDir)

	_, err = ctx.changesDir()
	assert.ErrorIs(t, err, domain.ErrChangesDirNotSet, "an unconfigured project must prompt, not silently default")
}

func TestResolveProject_FallsBackToDefaultWorkflow(t *testing.T) {
	_, root := projectFixture(t)
	registerGlobal(t, root, domain.ProjectEntry{ChangesDir: "/work/changes"})

	ctx, err := resolveProject()

	require.NoError(t, err)
	assert.Equal(t, domain.WorkflowStrict, ctx.Workflow, "an entry with no workflow inherits the global default")
}

func TestResolveProject_ResolvesFromSubdirectory(t *testing.T) {
	_, root := projectFixture(t)
	registerGlobal(t, root, domain.ProjectEntry{ChangesDir: "/work/changes"})
	deep := filepath.Join(root, "internal", "config")
	require.NoError(t, os.MkdirAll(deep, 0755))
	require.NoError(t, os.Chdir(deep))

	ctx, err := resolveProject()

	require.NoError(t, err)
	assert.Equal(t, "/work/changes", ctx.ChangesDir)
}

func TestSetActiveChange_WritesToGlobalEntry(t *testing.T) {
	_, root := projectFixture(t)
	hash := registerGlobal(t, root, domain.ProjectEntry{ChangesDir: "/work/changes"})

	ctx, err := resolveProject()
	require.NoError(t, err)
	require.NoError(t, ctx.setActiveChange("SDL-42"))

	cfg, err := config.LoadGlobal(globalConfigPath())
	require.NoError(t, err)
	assert.Equal(t, "SDL-42", cfg.Projects[hash].ActiveChange)
	assert.Equal(t, "/work/changes", cfg.Projects[hash].ChangesDir, "unrelated fields survive the write")
}

func TestSetActiveChange_WritesToLocalConfigWhenUnmigrated(t *testing.T) {
	_, root := projectFixture(t)
	writeLocalConfig(t, root, domain.NewLocalConfig(domain.StorageModeLocal, domain.WorkflowStrict, "legacyhash"))

	ctx, err := resolveProject()
	require.NoError(t, err)
	require.NoError(t, ctx.setActiveChange("SDL-42"))

	local, err := config.LoadLocal(filepath.Join(root, ".sdlaicrc"))
	require.NoError(t, err)
	assert.Equal(t, "SDL-42", local.ActiveChange)
}

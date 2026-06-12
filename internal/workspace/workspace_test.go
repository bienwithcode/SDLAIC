package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sdlaic/internal/domain"
)

func TestFindWorkspace_FindsInCurrentDir(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".sdlaicrc")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`{"version":1,"storage":"local","workflow":"strict","project_hash":"abc"}`), 0644))

	found, err := FindWorkspace(dir)
	require.NoError(t, err)
	assert.Equal(t, dir, found)
}

func TestFindWorkspace_FindsInParentDir(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".sdlaicrc")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`{"version":1,"storage":"local","workflow":"strict","project_hash":"abc"}`), 0644))

	// Create nested subdirectories
	nested := filepath.Join(dir, "src", "pkg", "module")
	require.NoError(t, os.MkdirAll(nested, 0755))

	found, err := FindWorkspace(nested)
	require.NoError(t, err)
	assert.Equal(t, dir, found)
}

func TestFindWorkspace_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := FindWorkspace(dir)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWorkspaceNotFound)
}

func TestFindWorkspace_FindsFirstAncestor(t *testing.T) {
	// Create: root/.sdlaicrc / root/sub/.sdlaicrc / root/sub/deep/
	// FindWorkspace from "deep" should find the NEAREST .sdlaicrc (sub)
	dir := t.TempDir()

	rootCfg := filepath.Join(dir, ".sdlaicrc")
	require.NoError(t, os.WriteFile(rootCfg, []byte(`{"version":1,"storage":"local","workflow":"strict","project_hash":"root"}`), 0644))

	subDir := filepath.Join(dir, "sub")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	subCfg := filepath.Join(subDir, ".sdlaicrc")
	require.NoError(t, os.WriteFile(subCfg, []byte(`{"version":1,"storage":"ignored","workflow":"free","project_hash":"sub"}`), 0644))

	deepDir := filepath.Join(subDir, "deep")
	require.NoError(t, os.MkdirAll(deepDir, 0755))

	found, err := FindWorkspace(deepDir)
	require.NoError(t, err)
	assert.Equal(t, subDir, found)
}

func TestInitWorkspace_CreatesConfigAndDirs(t *testing.T) {
	dir := t.TempDir()

	cfg, err := InitWorkspace(dir, domain.StorageModeLocal, domain.WorkflowStrict, "hash123")
	require.NoError(t, err)

	// Config should be returned
	assert.Equal(t, domain.StorageModeLocal, cfg.Storage)
	assert.Equal(t, domain.WorkflowStrict, cfg.Workflow)
	assert.Equal(t, "hash123", cfg.ProjectHash)

	// .sdlaicrc file should exist
	cfgPath := filepath.Join(dir, ".sdlaicrc")
	_, err = os.Stat(cfgPath)
	assert.NoError(t, err)

	// Changes directory should exist
	changesDir := filepath.Join(dir, ".sdlaic", "changes")
	_, err = os.Stat(changesDir)
	assert.NoError(t, err)
}

func TestInitWorkspace_IgnoredMode(t *testing.T) {
	dir := t.TempDir()

	cfg, err := InitWorkspace(dir, domain.StorageModeIgnored, domain.WorkflowLight, "hash456")
	require.NoError(t, err)
	assert.Equal(t, domain.StorageModeIgnored, cfg.Storage)

	// Changes directory should be .sdlaic/changes/
	changesDir := filepath.Join(dir, ".sdlaic", "changes")
	_, err = os.Stat(changesDir)
	assert.NoError(t, err)
}

func TestInitWorkspace_GlobalMode(t *testing.T) {
	dir := t.TempDir()

	// For global mode, we need a home dir override
	homeDir := t.TempDir()
	cfg, err := InitWorkspaceWithHome(dir, homeDir, domain.StorageModeGlobal, domain.WorkflowFree, "hash789")
	require.NoError(t, err)
	assert.Equal(t, domain.StorageModeGlobal, cfg.Storage)

	// .sdlaicrc should still exist in project root
	cfgPath := filepath.Join(dir, ".sdlaicrc")
	_, err = os.Stat(cfgPath)
	assert.NoError(t, err)
}

func TestInitWorkspace_RejectsReInit(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".sdlaicrc")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`{"version":1}`), 0644))

	_, err := InitWorkspace(dir, domain.StorageModeLocal, domain.WorkflowStrict, "hash123")
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWorkspaceExists)
}

func TestProjectHash(t *testing.T) {
	dir := t.TempDir()
	// Create a .git directory so it looks like a git repo
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0755))

	hash1, err := ProjectHash(dir)
	require.NoError(t, err)
	assert.NotEmpty(t, hash1)

	// Hash should be deterministic
	hash2, err := ProjectHash(dir)
	require.NoError(t, err)
	assert.Equal(t, hash1, hash2)
}

func TestProjectHash_DifferentDirs(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	hash1, err := ProjectHash(dir1)
	require.NoError(t, err)
	hash2, err := ProjectHash(dir2)
	require.NoError(t, err)

	// Different directories should produce different hashes
	assert.NotEqual(t, hash1, hash2)
}

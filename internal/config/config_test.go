package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// --- LoadLocal tests ---

func TestLoadLocal_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".sdlaicrc")

	raw := domain.LocalConfig{
		Version:      1,
		Storage:      domain.StorageModeIgnored,
		Workflow:     domain.WorkflowLight,
		ActiveChange: "JIRA-456",
		ProjectHash:  "abc123",
	}
	data, err := json.MarshalIndent(raw, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(cfgPath, data, 0644))

	got, err := LoadLocal(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, 1, got.Version)
	assert.Equal(t, domain.StorageModeIgnored, got.Storage)
	assert.Equal(t, domain.WorkflowLight, got.Workflow)
	assert.Equal(t, "JIRA-456", got.ActiveChange)
	assert.Equal(t, "abc123", got.ProjectHash)
}

func TestLoadLocal_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".sdlaicrc")
	require.NoError(t, os.WriteFile(cfgPath, []byte("not json at all"), 0644))

	_, err := LoadLocal(cfgPath)
	assert.Error(t, err)
}

func TestLoadLocal_MissingFile(t *testing.T) {
	_, err := LoadLocal("/nonexistent/path/.sdlaicrc")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, os.ErrNotExist))
}

func TestLoadLocal_InvalidStorageMode(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".sdlaicrc")

	// Write JSON with an invalid storage value
	data := `{"version":1,"storage":"cloud","workflow":"strict","project_hash":"abc"}`
	require.NoError(t, os.WriteFile(cfgPath, []byte(data), 0644))

	_, err := LoadLocal(cfgPath)
	assert.Error(t, err)
}

func TestLoadLocal_InvalidWorkflowLevel(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".sdlaicrc")

	data := `{"version":1,"storage":"local","workflow":"turbo","project_hash":"abc"}`
	require.NoError(t, os.WriteFile(cfgPath, []byte(data), 0644))

	_, err := LoadLocal(cfgPath)
	assert.Error(t, err)
}

// --- SaveLocal tests ---

func TestSaveLocal_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".sdlaicrc")

	original := domain.LocalConfig{
		Version:      1,
		Storage:      domain.StorageModeLocal,
		Workflow:     domain.WorkflowStrict,
		ActiveChange: "FEATURE-789",
		ProjectHash:  "def456",
	}
	require.NoError(t, SaveLocal(original, cfgPath))

	loaded, err := LoadLocal(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, original, loaded)
}

func TestSaveLocal_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "nested", "deep", ".sdlaicrc")

	cfg := domain.NewLocalConfig(domain.StorageModeLocal, domain.WorkflowStrict, "hash123")
	require.NoError(t, SaveLocal(cfg, cfgPath))

	// File should exist
	_, err := os.Stat(cfgPath)
	assert.NoError(t, err)
}

func TestSaveLocal_OmitEmptyActiveChange(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".sdlaicrc")

	cfg := domain.NewLocalConfig(domain.StorageModeLocal, domain.WorkflowStrict, "hash123")
	require.NoError(t, SaveLocal(cfg, cfgPath))

	data, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	// active_change should be omitted when empty (omitempty)
	assert.NotContains(t, string(data), "active_change")
}

// --- LoadGlobal tests ---

func TestLoadGlobal_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	raw := domain.GlobalConfig{
		Version:         1,
		DefaultWorkflow: domain.WorkflowFree,
		DefaultStorage:  domain.StorageModeGlobal,
		Projects: map[string]domain.ProjectEntry{
			"abc123": {Path: "/tmp/project", Storage: domain.StorageModeLocal},
		},
	}
	data, err := json.MarshalIndent(raw, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(cfgPath, data, 0644))

	got, err := LoadGlobal(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, 1, got.Version)
	assert.Equal(t, domain.WorkflowFree, got.DefaultWorkflow)
	assert.Equal(t, domain.StorageModeGlobal, got.DefaultStorage)
	assert.Len(t, got.Projects, 1)
	assert.Equal(t, "/tmp/project", got.Projects["abc123"].Path)
}

func TestLoadGlobal_V1FileLoadsWithoutError(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	// A config.json written by the pre-ChangesDir CLI: default_storage at the top
	// level and storage on each project. Neither exists in the v2 model, and both
	// must be ignored rather than rejected — a hard failure here would lock out
	// every existing user until they re-ran init for each project.
	v1 := `{
	  "version": 1,
	  "default_workflow": "strict",
	  "default_storage": "local",
	  "projects": {
	    "abc123": {
	      "path": "/tmp/project",
	      "storage": "ignored"
	    }
	  }
	}`
	require.NoError(t, os.WriteFile(cfgPath, []byte(v1), 0644))

	got, err := LoadGlobal(cfgPath)
	require.NoError(t, err)

	entry := got.Projects["abc123"]
	assert.Equal(t, "/tmp/project", entry.Path)
	assert.Empty(t, entry.ChangesDir, "v1 has no changes_dir — it must load empty so the CLI prompts for one")
	assert.Empty(t, entry.Workflow, "v1 kept workflow in .sdlaicrc, not in the global entry")
}

func TestLoadGlobal_PreservesProjectWorkflowAndActiveChange(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	v2 := `{
	  "version": 2,
	  "default_workflow": "strict",
	  "projects": {
	    "abc123": {
	      "path": "/tmp/project",
	      "changes_dir": "/tmp/openspec/changes",
	      "workflow": "light",
	      "active_change": "SDL-1"
	    }
	  }
	}`
	require.NoError(t, os.WriteFile(cfgPath, []byte(v2), 0644))

	got, err := LoadGlobal(cfgPath)
	require.NoError(t, err)

	entry := got.Projects["abc123"]
	assert.Equal(t, "/tmp/openspec/changes", entry.ChangesDir)
	assert.Equal(t, domain.WorkflowLight, entry.Workflow)
	assert.Equal(t, "SDL-1", entry.ActiveChange)
}

func TestValidateGlobal_AcceptsV1AndV2(t *testing.T) {
	for _, version := range []int{1, 2} {
		cfg := domain.GlobalConfig{
			Version:         version,
			DefaultWorkflow: domain.WorkflowStrict,
		}
		assert.NoError(t, ValidateGlobal(cfg), "version %d must validate", version)
	}
}

func TestLoadGlobal_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(cfgPath, []byte("{broken"), 0644))

	_, err := LoadGlobal(cfgPath)
	assert.Error(t, err)
}

func TestLoadGlobal_MissingFile(t *testing.T) {
	_, err := LoadGlobal("/nonexistent/config.json")
	assert.Error(t, err)
}

func TestLoadGlobal_InvalidDefaultStorage(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	data := `{"version":1,"default_storage":"cloud","default_workflow":"strict"}`
	require.NoError(t, os.WriteFile(cfgPath, []byte(data), 0644))

	_, err := LoadGlobal(cfgPath)
	assert.Error(t, err)
}

func TestLoadGlobal_InvalidDefaultWorkflow(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	data := `{"version":1,"default_storage":"local","default_workflow":"ultra"}`
	require.NoError(t, os.WriteFile(cfgPath, []byte(data), 0644))

	_, err := LoadGlobal(cfgPath)
	assert.Error(t, err)
}

// --- SaveGlobal tests ---

func TestSaveGlobal_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	original := domain.GlobalConfig{
		Version:         1,
		DefaultWorkflow: domain.WorkflowStrict,
		DefaultStorage:  domain.StorageModeLocal,
		Projects: map[string]domain.ProjectEntry{
			"proj1": {Path: "/tmp/proj1", Storage: domain.StorageModeLocal},
		},
	}
	require.NoError(t, SaveGlobal(original, cfgPath))

	loaded, err := LoadGlobal(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, original, loaded)
}

func TestSaveGlobal_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".sdlaic", "config.json")

	cfg := domain.NewGlobalConfig()
	require.NoError(t, SaveGlobal(cfg, cfgPath))

	_, err := os.Stat(cfgPath)
	assert.NoError(t, err)
}

// --- SetActiveChange tests ---

func TestSetActiveChange_UpdatesConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".sdlaicrc")

	cfg := domain.NewLocalConfig(domain.StorageModeLocal, domain.WorkflowStrict, "hash123")
	require.NoError(t, SaveLocal(cfg, cfgPath))

	require.NoError(t, SetActiveChange(cfgPath, "FEATURE-100"))

	loaded, err := LoadLocal(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, "FEATURE-100", loaded.ActiveChange)
}

func TestSetActiveChange_ClearsActive(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".sdlaicrc")

	cfg := domain.NewLocalConfig(domain.StorageModeLocal, domain.WorkflowStrict, "hash123")
	cfg.ActiveChange = "OLD-CHANGE"
	require.NoError(t, SaveLocal(cfg, cfgPath))

	require.NoError(t, SetActiveChange(cfgPath, ""))

	loaded, err := LoadLocal(cfgPath)
	require.NoError(t, err)
	assert.Empty(t, loaded.ActiveChange)
}

func TestSetActiveChange_MissingFile(t *testing.T) {
	err := SetActiveChange("/nonexistent/.sdlaicrc", "FEATURE-100")
	assert.Error(t, err)
}

// --- MergeDefaults tests ---

func TestMergeDefaults_AppliesGlobalDefaults(t *testing.T) {
	local := domain.LocalConfig{
		Version:     1,
		ProjectHash: "proj123",
		// Storage and Workflow are empty (zero values)
	}
	global := domain.GlobalConfig{
		Version:         1,
		DefaultStorage:  domain.StorageModeGlobal,
		DefaultWorkflow: domain.WorkflowFree,
	}

	merged := MergeDefaults(local, global)
	assert.Equal(t, domain.StorageModeGlobal, merged.Storage)
	assert.Equal(t, domain.WorkflowFree, merged.Workflow)
	assert.Equal(t, "proj123", merged.ProjectHash)
}

func TestMergeDefaults_LocalOverridesGlobal(t *testing.T) {
	local := domain.LocalConfig{
		Version:     1,
		Storage:     domain.StorageModeIgnored,
		Workflow:    domain.WorkflowLight,
		ProjectHash: "proj123",
	}
	global := domain.GlobalConfig{
		Version:         1,
		DefaultStorage:  domain.StorageModeGlobal,
		DefaultWorkflow: domain.WorkflowFree,
	}

	merged := MergeDefaults(local, global)
	// Local values should be preserved
	assert.Equal(t, domain.StorageModeIgnored, merged.Storage)
	assert.Equal(t, domain.WorkflowLight, merged.Workflow)
}

func TestMergeDefaults_ActiveChangePreserved(t *testing.T) {
	local := domain.LocalConfig{
		Version:      1,
		Storage:      domain.StorageModeLocal,
		Workflow:     domain.WorkflowStrict,
		ActiveChange: "MY-CHANGE",
		ProjectHash:  "proj123",
	}
	global := domain.NewGlobalConfig()

	merged := MergeDefaults(local, global)
	assert.Equal(t, "MY-CHANGE", merged.ActiveChange)
}

// --- ValidateLocal tests ---

func TestValidateLocal_ValidConfig(t *testing.T) {
	cfg := domain.NewLocalConfig(domain.StorageModeLocal, domain.WorkflowStrict, "hash123")
	assert.NoError(t, ValidateLocal(cfg))
}

func TestValidateLocal_InvalidStorage(t *testing.T) {
	cfg := domain.LocalConfig{
		Version:     1,
		Storage:     "invalid",
		Workflow:    domain.WorkflowStrict,
		ProjectHash: "hash123",
	}
	assert.Error(t, ValidateLocal(cfg))
}

func TestValidateLocal_InvalidWorkflow(t *testing.T) {
	cfg := domain.LocalConfig{
		Version:     1,
		Storage:     domain.StorageModeLocal,
		Workflow:    "invalid",
		ProjectHash: "hash123",
	}
	assert.Error(t, ValidateLocal(cfg))
}

func TestValidateLocal_EmptyProjectHash(t *testing.T) {
	cfg := domain.LocalConfig{
		Version:  1,
		Storage:  domain.StorageModeLocal,
		Workflow: domain.WorkflowStrict,
		// ProjectHash empty
	}
	assert.Error(t, ValidateLocal(cfg))
}

// --- ValidateGlobal tests ---

func TestValidateGlobal_ValidConfig(t *testing.T) {
	cfg := domain.NewGlobalConfig()
	assert.NoError(t, ValidateGlobal(cfg))
}

func TestValidateGlobal_InvalidDefaultStorage(t *testing.T) {
	cfg := domain.GlobalConfig{
		Version:         1,
		DefaultStorage:  "bad",
		DefaultWorkflow: domain.WorkflowStrict,
	}
	assert.Error(t, ValidateGlobal(cfg))
}

func TestValidateGlobal_InvalidDefaultWorkflow(t *testing.T) {
	cfg := domain.GlobalConfig{
		Version:         1,
		DefaultStorage:  domain.StorageModeLocal,
		DefaultWorkflow: "bad",
	}
	assert.Error(t, ValidateGlobal(cfg))
}

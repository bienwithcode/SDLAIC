package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sdlaic/internal/domain"
)

func TestStatus_JSONOutput(t *testing.T) {
	resetStatusFlags()
	dir := initWorkspaceForTest(t)

	// Create a change with context.md populated
	_, err := ExecuteCommand(rootCmd, "new", "change", "STATUS-TEST")
	require.NoError(t, err)

	// Overwrite context.md with real content so it registers as populated
	changeDir := filepath.Join(dir, ".sdlaic", "changes", "STATUS-TEST")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal content"), 0644))

	resetStatusFlags()
	output, err := ExecuteCommand(rootCmd, "status", "--json")
	require.NoError(t, err)

	// Parse JSON output
	var status domain.ChangeStatus
	require.NoError(t, json.Unmarshal([]byte(output), &status))

	assert.Equal(t, "STATUS-TEST", status.ActiveChange)
	assert.Equal(t, domain.StorageModeLocal, status.StorageMode)
	assert.Equal(t, domain.WorkflowStrict, status.Workflow)
	assert.Equal(t, domain.PhaseContext, status.CurrentPhase)
	assert.NotEmpty(t, status.ChangePath)
}

func TestStatus_HumanOutput(t *testing.T) {
	resetStatusFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "HUMAN-TEST")
	require.NoError(t, err)

	// Overwrite context.md with real content
	changeDir := filepath.Join(dir, ".sdlaic", "changes", "HUMAN-TEST")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal content"), 0644))

	resetStatusFlags()
	output, err := ExecuteCommand(rootCmd, "status")
	require.NoError(t, err)

	assert.Contains(t, output, "HUMAN-TEST")
	assert.Contains(t, output, "CONTEXT")
}

func TestStatus_WithChangeFlag(t *testing.T) {
	resetStatusFlags()
	_ = initWorkspaceForTest(t)

	// Create two changes
	_, err := ExecuteCommand(rootCmd, "new", "change", "CHANGE-A")
	require.NoError(t, err)
	_, err = ExecuteCommand(rootCmd, "new", "change", "CHANGE-B")
	require.NoError(t, err)

	// Status with --change flag should target specific change
	resetStatusFlags()
	output, err := ExecuteCommand(rootCmd, "status", "--change", "CHANGE-A", "--json")
	require.NoError(t, err)

	var status domain.ChangeStatus
	require.NoError(t, json.Unmarshal([]byte(output), &status))
	assert.Equal(t, "CHANGE-A", status.ActiveChange)
}

func TestStatus_NoActiveChange(t *testing.T) {
	resetStatusFlags()
	dir := initWorkspaceForTest(t)

	// Create a change then clear the active
	_, err := ExecuteCommand(rootCmd, "new", "change", "SOME-CHANGE")
	require.NoError(t, err)

	// Manually clear active change
	cfgPath := filepath.Join(dir, ".sdlaicrc")
	cfgData, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	var cfg domain.LocalConfig
	require.NoError(t, json.Unmarshal(cfgData, &cfg))
	cfg.ActiveChange = ""
	cfgBytes, err := json.MarshalIndent(cfg, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(cfgPath, cfgBytes, 0644))

	// Status without --change and no active should error
	resetStatusFlags()
	_, err = ExecuteCommand(rootCmd, "status")
	assert.Error(t, err)
}

func TestStatus_PhaseProgression(t *testing.T) {
	resetStatusFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "PROG-TEST")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "PROG-TEST")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal content"), 0644))

	// Phase: CONTEXT (context.md already exists from template)
	resetStatusFlags()
	output, err := ExecuteCommand(rootCmd, "status", "--json")
	require.NoError(t, err)
	var status domain.ChangeStatus
	require.NoError(t, json.Unmarshal([]byte(output), &status))
	assert.Equal(t, domain.PhaseContext, status.CurrentPhase)

	// Phase: CHALLENGED — add rationale.md
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "rationale.md"), []byte("# Rationale\nReal content."), 0644))
	resetStatusFlags()
	output, err = ExecuteCommand(rootCmd, "status", "--json")
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal([]byte(output), &status))
	assert.Equal(t, domain.PhaseChallenged, status.CurrentPhase)

	// Phase: PROPOSED — add proposal.md
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\nReal content."), 0644))
	resetStatusFlags()
	output, err = ExecuteCommand(rootCmd, "status", "--json")
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal([]byte(output), &status))
	assert.Equal(t, domain.PhaseProposed, status.CurrentPhase)

	// Phase: SPECIFIED — add specs.md
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "specs.md"), []byte("# Specs\nReal content."), 0644))
	resetStatusFlags()
	output, err = ExecuteCommand(rootCmd, "status", "--json")
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal([]byte(output), &status))
	assert.Equal(t, domain.PhaseSpecified, status.CurrentPhase)
}

func TestStatus_NoWorkspace(t *testing.T) {
	resetStatusFlags()
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(oldWd) })

	_, err := ExecuteCommand(rootCmd, "status")
	assert.Error(t, err)
}

func TestStatus_JSONArtifactFields(t *testing.T) {
	resetStatusFlags()
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "ART-TEST")
	require.NoError(t, err)

	resetStatusFlags()
	output, err := ExecuteCommand(rootCmd, "status", "--json")
	require.NoError(t, err)

	var status domain.ChangeStatus
	require.NoError(t, json.Unmarshal([]byte(output), &status))

	// Should have all 6 artifact types
	assert.Len(t, status.Artifacts, 6)
	assert.Contains(t, status.Artifacts, "context")
	assert.Contains(t, status.Artifacts, "rationale")
	assert.Contains(t, status.Artifacts, "proposal")
	assert.Contains(t, status.Artifacts, "specs")
	assert.Contains(t, status.Artifacts, "design")
	assert.Contains(t, status.Artifacts, "tasks")

	// context should be populated (from template)
	assert.True(t, status.Artifacts["context"].Exists)
}

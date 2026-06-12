package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_ValidChange(t *testing.T) {
	resetStatusFlags()
	resetValidateFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "VALID-1")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "VALID-1")
	// Add properly formatted artifacts
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal context content."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\nReal proposal content."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("# Tasks\n- [ ] Task 1\n- [ ] Task 2"), 0644))

	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate")
	assert.NoError(t, err)
}

func TestValidate_MissingCheckboxes(t *testing.T) {
	resetStatusFlags()
	resetValidateFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "NOBOX")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "NOBOX")
	// tasks.md without checkbox syntax
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("# Tasks\nJust a list without checkboxes"), 0644))

	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate")
	assert.Error(t, err)
}

func TestValidate_PlaceholderDetected(t *testing.T) {
	resetStatusFlags()
	resetValidateFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "PLACEHOLDER")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "PLACEHOLDER")
	// Artifact with template placeholder
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\n{{FILL_THIS_IN}}"), 0644))

	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate")
	assert.Error(t, err)
}

func TestValidate_StrictMissingArtifacts(t *testing.T) {
	resetStatusFlags()
	resetValidateFlags()
	_ = initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "STRICT-1")
	require.NoError(t, err)

	// In strict mode, missing artifacts up to current phase should fail
	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate", "--strict")
	assert.Error(t, err)
}

func TestValidate_SpecsDirectory(t *testing.T) {
	resetStatusFlags()
	resetValidateFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "SPECSDIR")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "SPECSDIR")
	
	// Create rationale.md to satisfy phase progression
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal context."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "rationale.md"), []byte("# Rationale\nDone"), 0644))
	// Create proposal.md
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\nContent"), 0644))
	
	// Create specs directory and a markdown file with a placeholder inside it
	specsDir := filepath.Join(changeDir, "specs", "sub-cap")
	require.NoError(t, os.MkdirAll(specsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte("# Spec\n{{PLACEHOLDER}}"), 0644))
	
	// Create design.md and tasks.md
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "design.md"), []byte("# Design\nContent"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("# Tasks\n- [ ] Task 1"), 0644))

	// Should fail in normal mode because of the placeholder inside specs/sub-cap/spec.md
	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate")
	assert.Error(t, err)

	// Now replace placeholder with real content
	require.NoError(t, os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte("# Spec\nReal requirement content"), 0644))

	// Should succeed now
	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate")
	assert.NoError(t, err)
}

func TestValidate_NoWorkspace(t *testing.T) {
	resetStatusFlags()
	resetValidateFlags()
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(oldWd) })

	_, err := ExecuteCommand(rootCmd, "validate")
	assert.Error(t, err)
}

func TestValidate_WithChangeFlag(t *testing.T) {
	resetStatusFlags()
	resetValidateFlags()
	dir := initWorkspaceForTest(t)

	// Create two changes
	_, err := ExecuteCommand(rootCmd, "new", "change", "GOOD")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "GOOD")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal context."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\nReal proposal content."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("# Tasks\n- [ ] Task 1"), 0644))

	_, err = ExecuteCommand(rootCmd, "new", "change", "BAD")
	require.NoError(t, err)

	changeDir2 := filepath.Join(dir, ".sdlaic", "changes", "BAD")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir2, "context.md"), []byte("# Context\nReal context."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir2, "tasks.md"), []byte("# Tasks\nNo checkboxes here"), 0644))

	// Validate specific change
	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate", "--change", "GOOD")
	assert.NoError(t, err)

	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate", "--change", "BAD")
	assert.Error(t, err)
}

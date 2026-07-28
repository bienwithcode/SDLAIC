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

func TestValidate_BracketPlaceholderDetected(t *testing.T) {
	resetStatusFlags()
	resetValidateFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "BRACKET")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "BRACKET")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal content."), 0644))
	// Bracket fill-in is the templates' actual placeholder syntax.
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\n[One paragraph: the problem this change solves.]"), 0644))

	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate")
	assert.Error(t, err)
}

func TestValidate_AnglePlaceholderDetected(t *testing.T) {
	resetStatusFlags()
	resetValidateFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "ANGLE")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "ANGLE")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal content."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal: <change-name>"), 0644))

	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate")
	assert.Error(t, err)
}

func TestValidate_BacktickBracketPlaceholderDetected(t *testing.T) {
	resetStatusFlags()
	resetValidateFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "BACKTICK")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "BACKTICK")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal content."), 0644))
	// Backtick-wrapped bracket token, as used in the tasks.md template.
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("# Tasks\n- [ ] 1.1 Add `[TestName]`"), 0644))

	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate")
	assert.Error(t, err)
}

func TestValidate_CommentPlaceholderDetected(t *testing.T) {
	resetStatusFlags()
	resetValidateFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "COMMENT")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "COMMENT")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal content."), 0644))
	// Non-allowlisted instruction comment (the proposal template's grill-log row).
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\nReal content.\n<!-- Populated from the scope grill. -->"), 0644))

	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate")
	assert.Error(t, err)
}

func TestValidate_TagsAndCheckboxesNotPlaceholders(t *testing.T) {
	resetStatusFlags()
	resetValidateFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "TAGS")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "TAGS")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal content."), 0644))
	// Structural TDD tags and checkbox glyphs must NOT be flagged as placeholders.
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("# Tasks\n"+
		"- [ ] 1.1 **[TEST-RED:unit]** Add TestRefresh — assert behavior.\n"+
		"- [ ] 1.2 **[IMPL]** Add RefreshModule.\n"+
		"- [x] 1.3 **[VERIFY]** Milestone integrates.\n"+
		"- [ ] 1.4 **[COMMIT]** feat: add refresh"), 0644))

	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate")
	assert.NoError(t, err, "TDD tags and checkbox glyphs are not placeholders")
}

func TestValidate_MarkdownLinkNotPlaceholder(t *testing.T) {
	resetStatusFlags()
	resetValidateFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "LINK")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "LINK")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal content."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\nSee [design notes](./design.md) for details."), 0644))

	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate")
	assert.NoError(t, err, "markdown link [text](url) is not a placeholder")
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

	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal context."), 0644))
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

func TestValidate_StrictRejectsSpecWithoutCapabilityDir(t *testing.T) {
	resetStatusFlags()
	resetValidateFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "MALFORMED-SPEC")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "MALFORMED-SPEC")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal context."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\nReal proposal."), 0644))
	// Malformed spec path: specs/spec.md (no capability dir). Contract requires
	// specs/<capability>/spec.md, so strict validation must reject it as missing.
	require.NoError(t, os.MkdirAll(filepath.Join(changeDir, "specs"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "specs", "spec.md"), []byte("# Spec\nReal requirement."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "design.md"), []byte("# Design\nReal design."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("# Tasks\n- [ ] Task 1"), 0644))

	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate", "--strict")
	assert.Error(t, err, "malformed spec path must fail strict validation")
}

func TestValidate_StrictFlagsEachEmptyCapabilitySpec(t *testing.T) {
	// Regression: with one populated and one empty capability spec, strict
	// validation must fail naming the empty capability. Previously a populated
	// sibling masked the empty one (OR-collapse) and strict passed silently.
	resetStatusFlags()
	resetValidateFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "MULTICAP")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "MULTICAP")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal context."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\nReal proposal."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "design.md"), []byte("# Design\nReal design."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("# Tasks\n- [ ] Task 1"), 0644))

	// capA populated, capB whitespace-only (counts as empty) — capB must be
	// flagged; capA must not.
	capA := filepath.Join(changeDir, "specs", "capA")
	require.NoError(t, os.MkdirAll(capA, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(capA, "spec.md"), []byte("# capA Spec\nReal requirement content."), 0644))
	capB := filepath.Join(changeDir, "specs", "capB")
	require.NoError(t, os.MkdirAll(capB, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(capB, "spec.md"), []byte("   \n  "), 0644))

	resetStatusFlags()
	resetValidateFlags()
	out, err := ExecuteCommand(rootCmd, "validate", "--strict")
	require.Error(t, err, "empty capability spec must fail strict even when a sibling is populated")
	assert.Contains(t, out, "specs/capB/spec.md: missing or empty", "must name the empty capability")
	assert.NotContains(t, out, "specs/capA/spec.md: missing or empty", "populated capability must not be flagged")
}

func TestValidate_StrictPassesWhenAllCapabilitySpecsPopulated(t *testing.T) {
	// Counterpart: multiple populated capability specs must pass strict, so the
	// per-capability check does not over-flag the all-good case.
	resetStatusFlags()
	resetValidateFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "MULTICAP-OK")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "MULTICAP-OK")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal context."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\nReal proposal."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "design.md"), []byte("# Design\nReal design."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("# Tasks\n- [ ] Task 1"), 0644))

	for _, cap := range []string{"capA", "capB"} {
		capDir := filepath.Join(changeDir, "specs", cap)
		require.NoError(t, os.MkdirAll(capDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(capDir, "spec.md"), []byte("# "+cap+" Spec\nReal requirement content."), 0644))
	}

	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate", "--strict")
	assert.NoError(t, err, "all populated capability specs must pass strict")
}

func TestValidate_StrictRespectsPerCapabilitySkip(t *testing.T) {
	// P2: an explicitly-skipped spec:<cap> gate exempts only that capability's
	// spec.md; a non-skipped sibling capability's empty spec is still flagged.
	resetStatusFlags()
	resetValidateFlags()
	dir := initWorkspaceForTest(t)

	_, err := ExecuteCommand(rootCmd, "new", "change", "MULTICAP-SKIP")
	require.NoError(t, err)

	changeDir := filepath.Join(dir, ".sdlaic", "changes", "MULTICAP-SKIP")
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "context.md"), []byte("# Context\nReal context."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\nReal proposal."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "design.md"), []byte("# Design\nReal design."), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("# Tasks\n- [ ] Task 1"), 0644))

	// Both capability specs empty.
	require.NoError(t, os.MkdirAll(filepath.Join(changeDir, "specs", "auth"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "specs", "auth", "spec.md"), []byte("  \n"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(changeDir, "specs", "billing"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "specs", "billing", "spec.md"), []byte("  \n"), 0644))

	// Explicitly skip only the auth capability's spec gate.
	resetGateFlags()
	_, err = ExecuteCommand(rootCmd, "gate", "set", "--phase", "spec:auth", "--status", "skipped")
	require.NoError(t, err)

	// Strict validation flags billing (not skipped) but not auth (skipped).
	resetStatusFlags()
	resetValidateFlags()
	out, err := ExecuteCommand(rootCmd, "validate", "--strict")
	require.Error(t, err)
	assert.Contains(t, out, "specs/billing/spec.md: missing or empty", "non-skipped capability must be flagged")
	assert.NotContains(t, out, "specs/auth/spec.md: missing or empty", "skipped capability must be exempt")
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

func TestValidate_StrictBlocksAfterLightInit(t *testing.T) {
	// Regression: a change drafted under light auto-skips every gate (SkippedAt
	// nil). After tightening config to strict, validate --strict must NOT honor
	// those auto-skips — downstream artifacts (spec/design/tasks) are absent and
	// must fail validation, exactly as `gate status` reports them non-passing.
	// Only an EXPLICIT skip (SkippedAt set, like proposal here) exempts an artifact.
	initGateWorkspace(t) // temp HOME + workspace + change GATE-TEST (only context.md)

	// Neutralize the scaffolded context.md (its template comments would otherwise
	// trip the placeholder check) so the ONLY validation outcome here is the
	// gate-skip logic under test.
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(wd, ".sdlaic", "changes", "GATE-TEST", "context.md"), []byte("# Context\nReal content."), 0644))

	_, err = ExecuteCommand(rootCmd, "config", "set", "workflow", "light")
	require.NoError(t, err)

	resetGateFlags()
	// First `gate set` lazily runs Init(light): every gate auto-skipped (SkippedAt
	// nil); proposal is then explicitly skipped (SkippedAt set).
	_, err = ExecuteCommand(rootCmd, "gate", "set", "--phase", "proposal", "--status", "skipped")
	require.NoError(t, err)

	resetGateFlags()
	_, err = ExecuteCommand(rootCmd, "config", "set", "workflow", "strict")
	require.NoError(t, err)

	// No proposal/spec/design/tasks artifacts exist. proposal is explicitly
	// skipped (exempt); spec/design/tasks are auto-skipped (SkippedAt nil) and
	// absent — strict validation must fail.
	resetStatusFlags()
	resetValidateFlags()
	_, err = ExecuteCommand(rootCmd, "validate", "--strict")
	assert.Error(t, err, "auto-skipped gates must not exempt absent artifacts under strict")
}

package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// helper to create a change directory with specified artifact files
func setupChangeDir(t *testing.T, artifacts map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test-change")
	require.NoError(t, os.MkdirAll(changeDir, 0755))

	for name, content := range artifacts {
		path := filepath.Join(changeDir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755)) // support nested keys like specs/core/spec.md
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	}

	return changeDir
}

// --- AnalyzePhase tests (table-driven for all 7 phases) ---

func TestAnalyzePhase_AllPhases(t *testing.T) {
	tests := []struct {
		name      string
		artifacts map[string]string // filename -> content
		expected  domain.Phase
	}{
		{
			name:      "EMPTY - no artifacts",
			artifacts: map[string]string{},
			expected:  domain.PhaseEmpty,
		},
		{
			name: "CONTEXT - only context.md populated",
			artifacts: map[string]string{
				"context.md": "# Context\n\nThis is real content about the change.",
			},
			expected: domain.PhaseContext,
		},
		{
			name: "PROPOSED - context + proposal populated",
			artifacts: map[string]string{
				"context.md":  "# Context\n\nReal content.",
				"proposal.md": "# Proposal\n\nReal proposal content here.",
			},
			expected: domain.PhaseProposed,
		},
		{
			name: "SPECIFIED - context + proposal + spec populated",
			artifacts: map[string]string{
				"context.md":         "# Context\n\nReal content.",
				"proposal.md":        "# Proposal\n\nReal proposal.",
				"specs/core/spec.md": "# Spec\n\nReal specification content.",
			},
			expected: domain.PhaseSpecified,
		},
		{
			name: "DESIGNED - context through design populated",
			artifacts: map[string]string{
				"context.md":         "# Context\n\nReal content.",
				"proposal.md":        "# Proposal\n\nReal proposal.",
				"specs/core/spec.md": "# Spec\n\nReal specs.",
				"design.md":          "# Design\n\nReal design document content.",
			},
			expected: domain.PhaseDesigned,
		},
		{
			name: "PLANNED - context through tasks populated",
			artifacts: map[string]string{
				"context.md":         "# Context\n\nReal content.",
				"proposal.md":        "# Proposal\n\nReal proposal.",
				"specs/core/spec.md": "# Spec\n\nReal specs.",
				"design.md":          "# Design\n\nReal design.",
				"tasks.md":           "# Tasks\n\n- [ ] Implement feature\n- [ ] Write tests",
			},
			expected: domain.PhasePlanned,
		},
		{
			name: "IMPLEMENTED - all artifacts + tasks all checked",
			artifacts: map[string]string{
				"context.md":         "# Context\n\nReal content.",
				"proposal.md":        "# Proposal\n\nReal proposal.",
				"specs/core/spec.md": "# Spec\n\nReal specs.",
				"design.md":          "# Design\n\nReal design.",
				"tasks.md":           "# Tasks\n\n- [x] Implement feature\n- [x] Write tests",
			},
			expected: domain.PhaseImplemented,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changeDir := setupChangeDir(t, tt.artifacts)
			phase, err := AnalyzePhase(changeDir)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, phase)
		})
	}
}

// --- Spec directory detection (only spec.md counts) ---

func TestAnalyzeArtifacts_SpecOnlyCountsSpecMd(t *testing.T) {
	// A stray README under specs/ must NOT satisfy the spec artifact.
	changeDir := setupChangeDir(t, map[string]string{
		"context.md":           "# Context\n\nReal content.",
		"proposal.md":          "# Proposal\n\nReal proposal.",
		"specs/core/README.md": "# Readme\n\nUnrelated prose that should not count.",
	})
	artifacts, err := AnalyzeArtifacts(changeDir)
	require.NoError(t, err)
	assert.False(t, artifacts["spec"].Exists, "README.md under specs/ must not satisfy the spec artifact")

	phase, err := AnalyzePhase(changeDir)
	require.NoError(t, err)
	assert.Equal(t, domain.PhaseProposed, phase, "stray specs/ markdown must not advance to SPECIFIED")
}

func TestAnalyzeArtifacts_SpecMdCounts(t *testing.T) {
	changeDir := setupChangeDir(t, map[string]string{
		"context.md":         "# Context\n\nReal content.",
		"proposal.md":        "# Proposal\n\nReal proposal.",
		"specs/core/spec.md": "# Spec\n\nReal requirement.",
	})
	artifacts, err := AnalyzeArtifacts(changeDir)
	require.NoError(t, err)
	assert.True(t, artifacts["spec"].Exists)
	assert.True(t, artifacts["spec"].Populated)
}

func TestIsCapabilitySpec(t *testing.T) {
	specsDir := filepath.Join("change", "specs")
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"capability spec", filepath.Join(specsDir, "auth", "spec.md"), true},
		{"case-insensitive leaf", filepath.Join(specsDir, "auth", "SPEC.MD"), true},
		{"no capability dir", filepath.Join(specsDir, "spec.md"), false},
		{"nested too deep", filepath.Join(specsDir, "a", "b", "spec.md"), false},
		{"unrelated leaf", filepath.Join(specsDir, "auth", "README.md"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsCapabilitySpec(specsDir, tt.path))
		})
	}
}

func TestAnalyzePhase_RejectsMalformedSpecPath(t *testing.T) {
	// specs/spec.md (no capability dir) violates the specs/<capability>/spec.md
	// contract and must NOT satisfy the spec artifact or advance the phase.
	changeDir := setupChangeDir(t, map[string]string{
		"context.md":    "# Context\n\nReal content.",
		"proposal.md":   "# Proposal\n\nReal proposal.",
		"specs/spec.md": "# Spec\n\nReal requirement.",
	})
	artifacts, err := AnalyzeArtifacts(changeDir)
	require.NoError(t, err)
	assert.False(t, artifacts["spec"].Exists, "specs/spec.md must not satisfy the spec artifact")

	phase, err := AnalyzePhase(changeDir)
	require.NoError(t, err)
	assert.Equal(t, domain.PhaseProposed, phase, "malformed spec path must not advance to SPECIFIED")
}

// --- Edge case tests ---

func TestAnalyzePhase_EmptyFiles(t *testing.T) {
	// Files exist but are empty or whitespace-only → should be treated as not populated
	changeDir := setupChangeDir(t, map[string]string{
		"context.md":  "   \n\n  ",
		"proposal.md": "",
	})
	phase, err := AnalyzePhase(changeDir)
	require.NoError(t, err)
	assert.Equal(t, domain.PhaseEmpty, phase)
}

func TestAnalyzePhase_PartiallyFilled(t *testing.T) {
	// Only proposal is populated but context is not → should be EMPTY
	// (phases must progress in order)
	changeDir := setupChangeDir(t, map[string]string{
		"proposal.md": "# Proposal\n\nReal proposal content.",
	})
	phase, err := AnalyzePhase(changeDir)
	require.NoError(t, err)
	assert.Equal(t, domain.PhaseEmpty, phase)
}

func TestAnalyzePhase_TasksPartiallyChecked(t *testing.T) {
	// All artifacts populated but tasks have mixed checked/unchecked → PLANNED (not IMPLEMENTED)
	changeDir := setupChangeDir(t, map[string]string{
		"context.md":         "# Context\n\nReal content.",
		"proposal.md":        "# Proposal\n\nReal proposal.",
		"specs/core/spec.md": "# Spec\n\nReal specs.",
		"design.md":          "# Design\n\nReal design.",
		"tasks.md":           "# Tasks\n\n- [x] Implement feature\n- [ ] Write tests",
	})
	phase, err := AnalyzePhase(changeDir)
	require.NoError(t, err)
	assert.Equal(t, domain.PhasePlanned, phase)
}

func TestAnalyzePhase_NonexistentDir(t *testing.T) {
	_, err := AnalyzePhase("/nonexistent/change/dir")
	assert.Error(t, err)
}

// --- AnalyzeArtifacts tests ---

func TestAnalyzeArtifacts_AllPresent(t *testing.T) {
	changeDir := setupChangeDir(t, map[string]string{
		"context.md":         "# Context\n\nReal content.",
		"proposal.md":        "# Proposal\n\nReal proposal.",
		"specs/core/spec.md": "# Spec\n\nReal specs.",
		"design.md":          "# Design\n\nReal design.",
		"tasks.md":           "# Tasks\n\n- [ ] Task 1\n- [ ] Task 2",
	})

	artifacts, err := AnalyzeArtifacts(changeDir)
	require.NoError(t, err)
	assert.Len(t, artifacts, 6) // 5 artifact types + the spec:core detail entry

	for _, at := range domain.OrderedArtifactTypes() {
		status, ok := artifacts[string(at)]
		require.True(t, ok, "artifact %s should be in results", at)
		assert.True(t, status.Exists, "%s should exist", at)
		assert.True(t, status.Populated, "%s should be populated", at)
	}
	// Per-capability detail entry mirrors the aggregate.
	assert.True(t, artifacts["spec:core"].Exists, "spec:core detail should exist")
	assert.True(t, artifacts["spec:core"].Populated, "spec:core detail should be populated")
}

func TestAnalyzeArtifacts_PerCapabilityAggregateIsAND(t *testing.T) {
	// Two capabilities, one populated and one empty: the aggregate "spec" must
	// read as NOT populated (AND of all capabilities), and the empty capability's
	// detail entry must read as not populated. This is the analyzer-side fix for
	// the OR-collapse that previously masked an empty sibling.
	changeDir := setupChangeDir(t, map[string]string{
		"context.md":         "# Context\n\nReal content.",
		"proposal.md":        "# Proposal\n\nReal proposal.",
		"specs/auth/spec.md": "# Auth Spec\n\nReal requirements.",
		// specs/billing/ exists as a directory but its spec.md is empty.
	})
	require.NoError(t, os.MkdirAll(filepath.Join(changeDir, "specs", "billing"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(changeDir, "specs", "billing", "spec.md"), []byte("   \n"), 0644))

	artifacts, err := AnalyzeArtifacts(changeDir)
	require.NoError(t, err)

	assert.True(t, artifacts["spec:auth"].Populated, "populated capability reads populated")
	assert.False(t, artifacts["spec:billing"].Populated, "empty capability reads not populated")
	assert.False(t, artifacts["spec"].Populated, "aggregate must be the AND of capabilities, not OR")
	assert.True(t, artifacts["spec"].Exists, "aggregate exists when any capability dir is present")

	// Because the aggregate spec is not fully populated, the phase must NOT reach
	// SPECIFIED.
	phase, err := AnalyzePhase(changeDir)
	require.NoError(t, err)
	assert.Equal(t, domain.PhaseProposed, phase, "a half-written spec tier must not advance to SPECIFIED")
}

func TestAnalyzeArtifacts_NonePresent(t *testing.T) {
	changeDir := setupChangeDir(t, nil)

	artifacts, err := AnalyzeArtifacts(changeDir)
	require.NoError(t, err)
	assert.Len(t, artifacts, 5)

	for _, at := range domain.OrderedArtifactTypes() {
		status := artifacts[string(at)]
		assert.False(t, status.Exists, "%s should not exist", at)
		assert.False(t, status.Populated, "%s should not be populated", at)
	}
}

func TestAnalyzeArtifacts_PartialPresence(t *testing.T) {
	changeDir := setupChangeDir(t, map[string]string{
		"context.md": "# Context\n\nReal content.",
	})

	artifacts, err := AnalyzeArtifacts(changeDir)
	require.NoError(t, err)

	assert.True(t, artifacts["context"].Exists)
	assert.True(t, artifacts["context"].Populated)
	assert.False(t, artifacts["proposal"].Exists)
	assert.False(t, artifacts["tasks"].Exists)
}

// --- IsPopulated tests ---

func TestIsPopulated_RealContent(t *testing.T) {
	assert.True(t, IsPopulated("# Title\n\nSome real content here."))
	assert.True(t, IsPopulated("A"))
}

func TestIsPopulated_Empty(t *testing.T) {
	assert.False(t, IsPopulated(""))
}

func TestIsPopulated_WhitespaceOnly(t *testing.T) {
	assert.False(t, IsPopulated("   "))
	assert.False(t, IsPopulated("\n\n\n"))
	assert.False(t, IsPopulated("  \n  \n  "))
}

func TestIsPopulated_TemplateComments(t *testing.T) {
	// HTML comments should not count as populated content
	assert.False(t, IsPopulated("<!-- Just a comment -->"))
	assert.False(t, IsPopulated("  <!-- comment -->  "))
}

func TestIsPopulated_MarkdownHeadersOnly(t *testing.T) {
	// Markdown headers and HTML comments alone should not count as populated content
	assert.False(t, IsPopulated("# Title\n## Section\n<!-- Instruction -->"))
	assert.False(t, IsPopulated("# Title\n\n## Section\n\n  \n"))
}

// --- AllTasksChecked tests ---

func TestAllTasksChecked_AllChecked(t *testing.T) {
	content := "# Tasks\n\n- [x] Task 1\n- [x] Task 2\n- [x] Task 3"
	assert.True(t, AllTasksChecked(content))
}

func TestAllTasksChecked_SomeChecked(t *testing.T) {
	content := "# Tasks\n\n- [x] Task 1\n- [ ] Task 2\n- [x] Task 3"
	assert.False(t, AllTasksChecked(content))
}

func TestAllTasksChecked_NoneChecked(t *testing.T) {
	content := "# Tasks\n\n- [ ] Task 1\n- [ ] Task 2"
	assert.False(t, AllTasksChecked(content))
}

func TestAllTasksChecked_NoCheckboxes(t *testing.T) {
	content := "# Tasks\n\nNo checkboxes here, just text."
	assert.False(t, AllTasksChecked(content))
}

func TestAllTasksChecked_Empty(t *testing.T) {
	assert.False(t, AllTasksChecked(""))
}

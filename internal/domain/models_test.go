package domain

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStorageMode(t *testing.T) {
	tests := []struct {
		input    string
		expected StorageMode
		wantErr  bool
	}{
		{"local", StorageModeLocal, false},
		{"ignored", StorageModeIgnored, false},
		{"global", StorageModeGlobal, false},
		{"cloud", "", true},
		{"", "", true},
		{"LOCAL", "", true}, // case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseStorageMode(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestValidStorageModes(t *testing.T) {
	modes := ValidStorageModes()
	assert.Len(t, modes, 3)
	assert.Contains(t, modes, StorageModeLocal)
	assert.Contains(t, modes, StorageModeIgnored)
	assert.Contains(t, modes, StorageModeGlobal)
}

func TestParseWorkflowLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected WorkflowLevel
		wantErr  bool
	}{
		{"strict", WorkflowStrict, false},
		{"light", WorkflowLight, false},
		{"free", WorkflowFree, false},
		{"moderate", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseWorkflowLevel(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestValidWorkflowLevels(t *testing.T) {
	levels := ValidWorkflowLevels()
	assert.Len(t, levels, 3)
}

func TestOrderedPhases(t *testing.T) {
	phases := OrderedPhases()
	assert.Len(t, phases, 7) // CHALLENGED removed in phase-gated restructuring
	assert.Equal(t, PhaseEmpty, phases[0])
	assert.Equal(t, PhaseImplemented, phases[6])
	assert.NotContains(t, phases, Phase("CHALLENGED"))
}

func TestArtifactTypeFileName(t *testing.T) {
	tests := []struct {
		artifact ArtifactType
		fileName string
	}{
		{ArtifactContext, "context.md"},
		{ArtifactProposal, "proposal.md"},
		{ArtifactSpec, "specs/<capability>/spec.md"},
		{ArtifactDesign, "design.md"},
		{ArtifactTasks, "tasks.md"},
	}

	for _, tt := range tests {
		t.Run(string(tt.artifact), func(t *testing.T) {
			assert.Equal(t, tt.fileName, tt.artifact.FileName())
		})
	}
}

func TestParseArtifactType(t *testing.T) {
	tests := []struct {
		input    string
		expected ArtifactType
		wantErr  bool
	}{
		{"context", ArtifactContext, false},
		{"proposal", ArtifactProposal, false},
		{"spec", ArtifactSpec, false},
		{"design", ArtifactDesign, false},
		{"tasks", ArtifactTasks, false},
		{"rationale", "", true}, // removed
		{"specs", "", true},     // renamed to singular spec
		{"unknown", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseArtifactType(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestOrderedArtifactTypes(t *testing.T) {
	types := OrderedArtifactTypes()
	assert.Len(t, types, 5) // rationale removed
	assert.Equal(t, ArtifactContext, types[0])
	assert.Equal(t, ArtifactTasks, types[4])
	assert.NotContains(t, types, ArtifactType("rationale"))
}

func TestNewGlobalConfig(t *testing.T) {
	cfg := NewGlobalConfig()
	assert.Equal(t, 2, cfg.Version)
	assert.Equal(t, WorkflowStrict, cfg.DefaultWorkflow)
	assert.Equal(t, StorageModeLocal, cfg.DefaultStorage)
	assert.NotNil(t, cfg.Projects)
	assert.Empty(t, cfg.Projects)
}

func TestProjectEntryJSONRoundTrip(t *testing.T) {
	entry := ProjectEntry{
		Path:         "/Users/dev/work/billing-api",
		ChangesDir:   "/Users/dev/work/openspec/changes",
		Workflow:     WorkflowStrict,
		ActiveChange: "SDL-142-add-invoice-export",
	}

	data, err := json.Marshal(entry)
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))
	assert.Equal(t, "/Users/dev/work/openspec/changes", raw["changes_dir"])
	assert.Equal(t, "strict", raw["workflow"])
	assert.Equal(t, "SDL-142-add-invoice-export", raw["active_change"])

	var got ProjectEntry
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, entry, got)
}

func TestProjectEntryOmitsEmptyActiveChange(t *testing.T) {
	data, err := json.Marshal(ProjectEntry{
		Path:       "/tmp/project",
		ChangesDir: "/tmp/project/.sdlaic/changes",
	})
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))
	assert.NotContains(t, raw, "active_change")
}

func TestSentinelErrors(t *testing.T) {
	// Verify sentinel errors are defined and non-nil
	errs := []error{
		ErrWorkspaceNotFound,
		ErrWorkspaceExists,
		ErrChangeNotFound,
		ErrChangeAlreadyExists,
		ErrNoActiveChange,
		ErrInvalidArtifact,
		ErrValidationFailed,
	}

	for _, err := range errs {
		assert.NotNil(t, err, "sentinel error should not be nil")
	}

	// Verify they are distinct
	seen := make(map[string]bool)
	for _, err := range errs {
		msg := err.Error()
		assert.False(t, seen[msg], "duplicate error message: %s", msg)
		seen[msg] = true
	}
}

// --- Gate State Types (T1) ---

func TestParseVerdict(t *testing.T) {
	tests := []struct {
		input    string
		expected Verdict
		wantErr  bool
	}{
		{"APPROVE", VerdictApprove, false},
		{"REQUEST_CHANGES", VerdictRequestChanges, false},
		{"REJECT", VerdictReject, false},
		{"PENDING", VerdictPending, false},
		{"approve", "", true}, // case-sensitive
		{"MAYBE", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseVerdict(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
				assert.True(t, errors.Is(err, ErrInvalidVerdict), "should wrap ErrInvalidVerdict")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestValidVerdicts(t *testing.T) {
	assert.Len(t, ValidVerdicts(), 4)
}

func TestVerdictToGateStatus(t *testing.T) {
	tests := []struct {
		verdict  Verdict
		expected GateStatus
	}{
		{VerdictApprove, GateStatusApproved},
		{VerdictRequestChanges, GateStatusFailed},
		{VerdictReject, GateStatusFailed},
		{VerdictPending, GateStatusReviewing},
	}

	for _, tt := range tests {
		t.Run(string(tt.verdict), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.verdict.ToGateStatus())
		})
	}
}

func TestParseGateStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected GateStatus
		wantErr  bool
	}{
		{"pending", GateStatusPending, false},
		{"grilling", GateStatusGrilling, false},
		{"grilled", GateStatusGrilled, false},
		{"reviewing", GateStatusReviewing, false},
		{"approved", GateStatusApproved, false},
		{"failed", GateStatusFailed, false},
		{"skipped", GateStatusSkipped, false},
		{"APPROVED", "", true}, // case-sensitive
		{"done", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseGateStatus(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidGateStatus), "should wrap ErrInvalidGateStatus")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestValidGateStatuses(t *testing.T) {
	assert.Len(t, ValidGateStatuses(), 7)
}

func TestGateStatusIsPassing(t *testing.T) {
	passing := []GateStatus{GateStatusApproved, GateStatusSkipped}
	blocking := []GateStatus{
		GateStatusPending, GateStatusGrilling, GateStatusGrilled,
		GateStatusReviewing, GateStatusFailed,
	}
	for _, s := range passing {
		assert.True(t, s.IsPassing(), "%s should pass", s)
	}
	for _, s := range blocking {
		assert.False(t, s.IsPassing(), "%s should block", s)
	}
}

func TestGateIsPassingFor(t *testing.T) {
	stamp := "2026-07-25T10:00:00Z"
	tests := []struct {
		name     string
		gate     Gate
		workflow WorkflowLevel
		want     bool
	}{
		{"approved under strict", Gate{Status: GateStatusApproved}, WorkflowStrict, true},
		{"explicit skip under strict", Gate{Status: GateStatusSkipped, SkippedAt: &stamp}, WorkflowStrict, true},
		{"auto skip under strict (the bug)", Gate{Status: GateStatusSkipped}, WorkflowStrict, false},
		{"pending under strict", Gate{Status: GateStatusPending}, WorkflowStrict, false},
		{"failed under strict", Gate{Status: GateStatusFailed}, WorkflowStrict, false},
		{"pending under light never blocks", Gate{Status: GateStatusPending}, WorkflowLight, true},
		{"failed under free never blocks", Gate{Status: GateStatusFailed}, WorkflowFree, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.gate.IsPassingFor(tt.workflow))
		})
	}
}

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		input   string
		want    Severity
		wantErr bool
	}{
		{"CRITICAL", SeverityCritical, false},
		{"HIGH", SeverityHigh, false},
		{"MEDIUM", SeverityMedium, false},
		{"LOW", SeverityLow, false},
		{"INFO", SeverityInfo, false},
		{"trivial", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseSeverity(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidSeverity), "should wrap ErrInvalidSeverity")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGatesFileJSONRoundTrip(t *testing.T) {
	approvedAt := "2026-07-25T10:00:00Z"
	gf := GatesFile{
		SchemaVersion: 1,
		ProjectHash:   "abc123def456",
		Change:        "JIRA-789",
		Workflow:      WorkflowStrict,
		PipelineState: "PROPOSED",
		CurrentGate:   "proposal",
		Gates: map[string]Gate{
			"proposal": {
				Phase:    PhaseProposed,
				Artifact: "proposal.md",
				Status:   GateStatusApproved,
				Grill: GrillRecord{
					Questions:  3,
					Checklist:  "references/grills/scope-grill.md",
					ResolvedAt: &approvedAt,
				},
				Review: ReviewRecord{
					Verdict: VerdictApprove,
					Findings: []Finding{
						{Severity: SeverityLow, Evidence: "proposal.md:12", Message: "scope table thin"},
					},
					Checklist:  "references/reviews/proposal-audit.md",
					ReviewedAt: &approvedAt,
				},
				Attempts:   1,
				ApprovedAt: &approvedAt,
			},
		},
		History: []ReEntryEvent{
			{FromPhase: PhaseProposed, Reason: "ticket scope changed", At: approvedAt, SupersededGates: []string{"spec"}},
		},
		CreatedAt: approvedAt,
		UpdatedAt: approvedAt,
	}

	data, err := json.Marshal(gf)
	require.NoError(t, err)

	// JSON tags must be snake_case per spec.
	js := string(data)
	assert.Contains(t, js, `"schema_version"`)
	assert.Contains(t, js, `"project_hash"`)
	assert.Contains(t, js, `"current_gate"`)
	assert.Contains(t, js, `"approved_at"`)

	var got GatesFile
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, gf.SchemaVersion, got.SchemaVersion)
	assert.Equal(t, gf.Change, got.Change)
	require.Contains(t, got.Gates, "proposal")
	assert.Equal(t, GateStatusApproved, got.Gates["proposal"].Status)
	require.NotNil(t, got.Gates["proposal"].ApprovedAt)
	assert.Equal(t, approvedAt, *got.Gates["proposal"].ApprovedAt)
	assert.Len(t, got.Gates["proposal"].Review.Findings, 1)
	assert.Equal(t, SeverityLow, got.Gates["proposal"].Review.Findings[0].Severity)
	require.Len(t, got.History, 1)
	assert.Equal(t, []string{"spec"}, got.History[0].SupersededGates)
}

func TestGateSupersededOmitEmpty(t *testing.T) {
	// SupersededBy uses omitempty — absent when nil.
	g := Gate{Phase: PhaseDesigned, Artifact: "design.md", Status: GateStatusPending}
	data, err := json.Marshal(g)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "superseded_by")

	by := "design"
	g.Superseded = true
	g.SupersededBy = &by
	data, err = json.Marshal(g)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"superseded_by":"design"`)
}

func TestGateStateSentinels(t *testing.T) {
	errs := []error{ErrInvalidVerdict, ErrInvalidGateStatus, ErrInvalidSeverity}
	for _, err := range errs {
		assert.NotNil(t, err)
	}
}

func TestChangeStatusJSON(t *testing.T) {
	status := ChangeStatus{
		ActiveChange: "JIRA-456",
		ChangesDir:   "/tmp/test/changes",
		Workflow:     WorkflowStrict,
		CurrentPhase: PhaseProposed,
		ChangePath:   "/tmp/test/changes/JIRA-456",
		Artifacts: map[string]ArtifactStatus{
			"context.md":  {Exists: true, Populated: true, Valid: true},
			"proposal.md": {Exists: true, Populated: true, Valid: true},
			"specs.md":    {Exists: false, Populated: false, Valid: false},
		},
	}

	assert.Equal(t, "JIRA-456", status.ActiveChange)
	assert.Equal(t, PhaseProposed, status.CurrentPhase)
	assert.Len(t, status.Artifacts, 3)
	assert.True(t, status.Artifacts["context.md"].Populated)
	assert.False(t, status.Artifacts["specs.md"].Exists)
}

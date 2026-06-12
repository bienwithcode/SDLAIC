package domain

import (
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
	assert.Len(t, phases, 8)
	assert.Equal(t, PhaseEmpty, phases[0])
	assert.Equal(t, PhaseImplemented, phases[7])
}

func TestArtifactTypeFileName(t *testing.T) {
	tests := []struct {
		artifact  ArtifactType
		fileName  string
	}{
		{ArtifactContext, "context.md"},
		{ArtifactRationale, "rationale.md"},
		{ArtifactProposal, "proposal.md"},
		{ArtifactSpecs, "specs.md"},
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
		{"rationale", ArtifactRationale, false},
		{"proposal", ArtifactProposal, false},
		{"specs", ArtifactSpecs, false},
		{"design", ArtifactDesign, false},
		{"tasks", ArtifactTasks, false},
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
	assert.Len(t, types, 6)
	assert.Equal(t, ArtifactContext, types[0])
	assert.Equal(t, ArtifactTasks, types[5])
}

func TestNewLocalConfig(t *testing.T) {
	cfg := NewLocalConfig(StorageModeIgnored, WorkflowLight, "abc123")
	assert.Equal(t, 1, cfg.Version)
	assert.Equal(t, StorageModeIgnored, cfg.Storage)
	assert.Equal(t, WorkflowLight, cfg.Workflow)
	assert.Equal(t, "abc123", cfg.ProjectHash)
	assert.Empty(t, cfg.ActiveChange)
}

func TestNewGlobalConfig(t *testing.T) {
	cfg := NewGlobalConfig()
	assert.Equal(t, 1, cfg.Version)
	assert.Equal(t, WorkflowStrict, cfg.DefaultWorkflow)
	assert.Equal(t, StorageModeLocal, cfg.DefaultStorage)
	assert.NotNil(t, cfg.Projects)
	assert.Empty(t, cfg.Projects)
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

func TestChangeStatusJSON(t *testing.T) {
	status := ChangeStatus{
		ActiveChange: "JIRA-456",
		StorageMode:  StorageModeGlobal,
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

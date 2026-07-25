// Package domain defines shared types, constants, and errors for the SDLAIC CLI.
package domain

import (
	"errors"
	"fmt"
)

// --- Storage Modes ---

// StorageMode determines where change artifacts are stored.
type StorageMode string

const (
	StorageModeLocal   StorageMode = "local"   // <project-root>/.sdlaic/changes/ — tracked by git
	StorageModeIgnored StorageMode = "ignored" // <project-root>/.sdlaic/changes/ — in .gitignore
	StorageModeGlobal  StorageMode = "global"  // ~/.sdlaic/stores/<hash>/changes/ — out of tree
)

// ValidStorageModes returns all supported storage modes.
func ValidStorageModes() []StorageMode {
	return []StorageMode{StorageModeLocal, StorageModeIgnored, StorageModeGlobal}
}

// ParseStorageMode converts a string to a StorageMode, returning an error if invalid.
func ParseStorageMode(s string) (StorageMode, error) {
	sm := StorageMode(s)
	for _, mode := range ValidStorageModes() {
		if sm == mode {
			return sm, nil
		}
	}
	return "", fmt.Errorf("invalid storage mode %q; valid: local, ignored, global", s)
}

// --- Workflow Levels ---

// WorkflowLevel controls how strictly the enforcer agent validates phase progression.
type WorkflowLevel string

const (
	WorkflowStrict WorkflowLevel = "strict" // All phases must be completed in order
	WorkflowLight  WorkflowLevel = "light"  // Some phases can be skipped
	WorkflowFree   WorkflowLevel = "free"   // No phase enforcement
)

// ValidWorkflowLevels returns all supported workflow levels.
func ValidWorkflowLevels() []WorkflowLevel {
	return []WorkflowLevel{WorkflowStrict, WorkflowLight, WorkflowFree}
}

// ParseWorkflowLevel converts a string to a WorkflowLevel, returning an error if invalid.
func ParseWorkflowLevel(s string) (WorkflowLevel, error) {
	wl := WorkflowLevel(s)
	for _, level := range ValidWorkflowLevels() {
		if wl == level {
			return wl, nil
		}
	}
	return "", fmt.Errorf("invalid workflow level %q; valid: strict, light, free", s)
}

// --- Phases ---

// Phase represents the current stage of a change's lifecycle.
type Phase string

const (
	PhaseEmpty       Phase = "EMPTY"       // Change directory exists, no artifacts
	PhaseContext     Phase = "CONTEXT"      // context.md is populated
	PhaseChallenged  Phase = "CHALLENGED"   // rationale.md is populated
	PhaseProposed    Phase = "PROPOSED"     // proposal.md is populated
	PhaseSpecified   Phase = "SPECIFIED"    // specs.md is populated
	PhaseDesigned    Phase = "DESIGNED"     // design.md is populated
	PhasePlanned     Phase = "PLANNED"      // tasks.md is populated
	PhaseImplemented Phase = "IMPLEMENTED"  // All artifacts complete, tasks checked
)

// OrderedPhases returns phases in their natural progression order.
func OrderedPhases() []Phase {
	return []Phase{
		PhaseEmpty,
		PhaseContext,
		PhaseChallenged,
		PhaseProposed,
		PhaseSpecified,
		PhaseDesigned,
		PhasePlanned,
		PhaseImplemented,
	}
}

// --- Artifact Types ---

// ArtifactType represents a kind of change artifact file.
type ArtifactType string

const (
	ArtifactContext   ArtifactType = "context"
	ArtifactRationale ArtifactType = "rationale"
	ArtifactProposal  ArtifactType = "proposal"
	ArtifactSpecs     ArtifactType = "specs"
	ArtifactDesign    ArtifactType = "design"
	ArtifactTasks     ArtifactType = "tasks"
)

// FileName returns the markdown filename for an artifact type.
func (at ArtifactType) FileName() string {
	return string(at) + ".md"
}

// OrderedArtifactTypes returns artifact types in their natural order.
func OrderedArtifactTypes() []ArtifactType {
	return []ArtifactType{
		ArtifactContext,
		ArtifactRationale,
		ArtifactProposal,
		ArtifactSpecs,
		ArtifactDesign,
		ArtifactTasks,
	}
}

// ParseArtifactType converts a string to an ArtifactType, returning an error if invalid.
func ParseArtifactType(s string) (ArtifactType, error) {
	at := ArtifactType(s)
	for _, t := range OrderedArtifactTypes() {
		if at == t {
			return at, nil
		}
	}
	return "", fmt.Errorf("invalid artifact type %q; valid: context, proposal, specs, design, tasks", s)
}

// --- Structs ---

// ArtifactStatus describes the state of a single artifact file.
type ArtifactStatus struct {
	Exists    bool `json:"exists"`
	Populated bool `json:"populated"`
	Valid     bool `json:"valid"`
}

// ChangeStatus is the full status of a change, returned by `sdlaic status --json`.
type ChangeStatus struct {
	ActiveChange string                    `json:"active_change"`
	StorageMode  StorageMode               `json:"storage_mode"`
	Workflow     WorkflowLevel             `json:"workflow"`
	CurrentPhase Phase                     `json:"current_phase"`
	ChangePath   string                    `json:"change_path"`
	Artifacts    map[string]ArtifactStatus `json:"artifacts"`
}

// LocalConfig represents the .sdlaicrc file in a project root.
type LocalConfig struct {
	Version      int           `json:"version"`
	Storage      StorageMode   `json:"storage"`
	Workflow     WorkflowLevel `json:"workflow"`
	ActiveChange string        `json:"active_change,omitempty"`
	ProjectHash  string        `json:"project_hash"`
}

// NewLocalConfig returns a LocalConfig with defaults applied.
func NewLocalConfig(storage StorageMode, workflow WorkflowLevel, projectHash string) LocalConfig {
	return LocalConfig{
		Version:     1,
		Storage:     storage,
		Workflow:    workflow,
		ProjectHash: projectHash,
	}
}

// ProjectEntry stores info about a known project in the global config.
type ProjectEntry struct {
	Path    string      `json:"path"`
	Storage StorageMode `json:"storage"`
}

// GlobalConfig represents the ~/.sdlaic/config.json file.
type GlobalConfig struct {
	Version         int                    `json:"version"`
	DefaultWorkflow WorkflowLevel          `json:"default_workflow"`
	DefaultStorage  StorageMode            `json:"default_storage"`
	Projects        map[string]ProjectEntry `json:"projects,omitempty"`
}

// NewGlobalConfig returns a GlobalConfig with defaults applied.
func NewGlobalConfig() GlobalConfig {
	return GlobalConfig{
		Version:         1,
		DefaultWorkflow: WorkflowStrict,
		DefaultStorage:  StorageModeLocal,
		Projects:        make(map[string]ProjectEntry),
	}
}

// --- Gate State ---
//
// Gate verdicts are persisted in the global state store
// (~/.sdlaic/state/<project_hash>/<change>/meta.json), never inside the
// project repo. Verdict records a review's decision; GateStatus records a
// gate's lifecycle position; the two are related by Verdict.ToGateStatus.

// Verdict is the decision issued by a review (skills/review) for an artifact.
type Verdict string

const (
	VerdictApprove        Verdict = "APPROVE"         // unlock next phase
	VerdictRequestChanges Verdict = "REQUEST_CHANGES" // rework required
	VerdictReject         Verdict = "REJECT"          // rework required
	VerdictPending        Verdict = "PENDING"         // not yet decided
)

// ValidVerdicts returns all supported review verdicts.
func ValidVerdicts() []Verdict {
	return []Verdict{VerdictApprove, VerdictRequestChanges, VerdictReject, VerdictPending}
}

// ParseVerdict converts a string to a Verdict, wrapping ErrInvalidVerdict if invalid.
func ParseVerdict(s string) (Verdict, error) {
	v := Verdict(s)
	for _, valid := range ValidVerdicts() {
		if v == valid {
			return v, nil
		}
	}
	return "", fmt.Errorf("invalid verdict %q; valid: APPROVE, REQUEST_CHANGES, REJECT, PENDING: %w", s, ErrInvalidVerdict)
}

// ToGateStatus maps a review verdict onto the gate lifecycle status it implies:
// APPROVE → approved, REQUEST_CHANGES|REJECT → failed, PENDING → reviewing.
func (v Verdict) ToGateStatus() GateStatus {
	switch v {
	case VerdictApprove:
		return GateStatusApproved
	case VerdictRequestChanges, VerdictReject:
		return GateStatusFailed
	default:
		return GateStatusReviewing
	}
}

// GateStatus is the lifecycle position of a single phase gate.
type GateStatus string

const (
	GateStatusPending   GateStatus = "pending"
	GateStatusGrilling  GateStatus = "grilling"
	GateStatusGrilled   GateStatus = "grilled"
	GateStatusReviewing GateStatus = "reviewing"
	GateStatusApproved  GateStatus = "approved"
	GateStatusFailed    GateStatus = "failed"
	GateStatusSkipped   GateStatus = "skipped"
)

// ValidGateStatuses returns all supported gate statuses.
func ValidGateStatuses() []GateStatus {
	return []GateStatus{
		GateStatusPending,
		GateStatusGrilling,
		GateStatusGrilled,
		GateStatusReviewing,
		GateStatusApproved,
		GateStatusFailed,
		GateStatusSkipped,
	}
}

// ParseGateStatus converts a string to a GateStatus, wrapping ErrInvalidGateStatus if invalid.
func ParseGateStatus(s string) (GateStatus, error) {
	gs := GateStatus(s)
	for _, valid := range ValidGateStatuses() {
		if gs == valid {
			return gs, nil
		}
	}
	return "", fmt.Errorf("invalid gate status %q: %w", s, ErrInvalidGateStatus)
}

// IsPassing reports whether a gate status unblocks its phase (approved or skipped).
func (gs GateStatus) IsPassing() bool {
	return gs == GateStatusApproved || gs == GateStatusSkipped
}

// Severity classifies a review finding.
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
	SeverityInfo     Severity = "INFO"
)

// ValidSeverities returns all supported finding severities.
func ValidSeverities() []Severity {
	return []Severity{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo}
}

// ParseSeverity converts a string to a Severity, wrapping ErrInvalidSeverity if invalid.
func ParseSeverity(s string) (Severity, error) {
	sev := Severity(s)
	for _, valid := range ValidSeverities() {
		if sev == valid {
			return sev, nil
		}
	}
	return "", fmt.Errorf("invalid severity %q: %w", s, ErrInvalidSeverity)
}

// Finding is a single issue raised during a review, carrying primary evidence
// (a path:line reference or Jira ID) per the Claim Verification Rule.
type Finding struct {
	Severity Severity `json:"severity"`
	Evidence string   `json:"evidence"`
	Message  string   `json:"message"`
}

// GrillRecord captures the outcome of a Socratic challenge (skills/grillme) for a gate.
type GrillRecord struct {
	Questions  int     `json:"questions"`
	Checklist  string  `json:"checklist,omitempty"`
	ResolvedAt *string `json:"resolved_at,omitempty"`
}

// ReviewRecord captures the outcome of an independent audit (skills/review) for a gate.
type ReviewRecord struct {
	Verdict    Verdict   `json:"verdict"`
	Findings   []Finding `json:"findings,omitempty"`
	Checklist  string    `json:"checklist,omitempty"`
	ReviewedAt *string   `json:"reviewed_at,omitempty"`
}

// ReEntryEvent records a mid-flight requirement change (Universal Re-entry, §5)
// that re-enters the pipeline at an earlier artifact and cascades downstream.
type ReEntryEvent struct {
	FromPhase       Phase    `json:"from_phase"`
	Reason          string   `json:"reason"`
	At              string   `json:"at"`
	SupersededGates []string `json:"superseded_gates,omitempty"`
}

// Gate is the persisted state of a single phase gate.
type Gate struct {
	Phase        Phase        `json:"phase"` // PROPOSED | SPECIFIED | DESIGNED | PLANNED
	Artifact     string       `json:"artifact"`
	Status       GateStatus   `json:"status"`
	Grill        GrillRecord  `json:"grill"`
	Review       ReviewRecord `json:"review"`
	Attempts     int          `json:"attempts"`
	ApprovedAt   *string      `json:"approved_at"`
	Superseded   bool         `json:"superseded"`
	SupersededBy *string      `json:"superseded_by,omitempty"`
}

// GatesFile is the machine-readable gate state for one change — the source of
// truth for "has this phase been approved". One record per change.
type GatesFile struct {
	SchemaVersion int             `json:"schema_version"`
	ProjectHash   string          `json:"project_hash"`
	Change        string          `json:"change"`
	Workflow      WorkflowLevel   `json:"workflow"`
	PipelineState string          `json:"pipeline_state"`
	CurrentGate   string          `json:"current_gate"`
	Gates         map[string]Gate `json:"gates"` // keyed by: proposal | spec | design | tasks
	History       []ReEntryEvent  `json:"history,omitempty"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
}

// --- Sentinel Errors ---

var (
	ErrWorkspaceNotFound   = errors.New("no .sdlaicrc found in parent directories")
	ErrWorkspaceExists     = errors.New("workspace already initialized")
	ErrChangeNotFound      = errors.New("change not found")
	ErrChangeAlreadyExists = errors.New("change already exists")
	ErrNoActiveChange      = errors.New("no active change set; use --change flag or sdlaic switch")
	ErrInvalidArtifact     = errors.New("invalid artifact type")
	ErrValidationFailed    = errors.New("validation failed")
	ErrInvalidVerdict      = errors.New("invalid verdict")
	ErrInvalidGateStatus   = errors.New("invalid gate status")
	ErrInvalidSeverity     = errors.New("invalid severity")
)

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

// --- Sentinel Errors ---

var (
	ErrWorkspaceNotFound   = errors.New("no .sdlaicrc found in parent directories")
	ErrWorkspaceExists     = errors.New("workspace already initialized")
	ErrChangeNotFound      = errors.New("change not found")
	ErrChangeAlreadyExists = errors.New("change already exists")
	ErrNoActiveChange      = errors.New("no active change set; use --change flag or sdlaic switch")
	ErrInvalidArtifact     = errors.New("invalid artifact type")
	ErrValidationFailed    = errors.New("validation failed")
)

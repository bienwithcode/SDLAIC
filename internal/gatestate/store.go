// Package gatestate reads and writes the global gate-state store that records
// phase-gate verdicts for a change.
//
// State lives OUTSIDE the project repo, at:
//
//	~/.sdlaic/state/<project_hash>/<change>/
//	  ├── meta.json      — machine-readable gate state (source of truth for "approved?")
//	  ├── review.md      — human-readable mirror of the latest verdict + findings
//	  └── history.jsonl  — append-only re-entry / follow-up events
//
// Keeping verdicts here (never as a file suffix or marker inside the reviewed
// artifact) keeps the project repo pristine. The <project_hash> is the same
// value produced by workspace.ProjectHash, so the state store and the global
// artifact store (~/.sdlaic/stores/) key identically.
package gatestate

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// SchemaVersion is the current meta.json schema version. Readers ignore unknown
// fields, so forward-compatible additions do not require a bump.
const SchemaVersion = 1

const (
	metaFile    = "meta.json"
	reviewFile  = "review.md"
	historyFile = "history.jsonl"
)

// gateMeta describes one gate in the pipeline: its key, the phase it satisfies,
// and the artifact it gates.
type gateMeta struct {
	key      string
	phase    domain.Phase
	artifact string
}

// pipeline is the canonical, ordered set of gates.
var pipeline = []gateMeta{
	{"proposal", domain.PhaseProposed, "proposal.md"},
	{"spec", domain.PhaseSpecified, "specs/<capability>/spec.md"},
	{"design", domain.PhaseDesigned, "design.md"},
	{"tasks", domain.PhasePlanned, "tasks.md"},
}

// GateKeys returns the canonical gate keys in pipeline order.
func GateKeys() []string {
	keys := make([]string, len(pipeline))
	for i, g := range pipeline {
		keys[i] = g.key
	}
	return keys
}

func gateFor(key string) (gateMeta, bool) {
	for _, g := range pipeline {
		if g.key == key {
			return g, true
		}
	}
	return gateMeta{}, false
}

// Store is a handle to one change's gate-state directory.
type Store struct {
	home        string
	projectHash string
	change      string
	now         func() time.Time
}

// New returns a Store rooted at the user's home directory.
func New(projectHash, change string) *Store {
	return NewWithHome(defaultHomeDir(), projectHash, change)
}

// NewWithHome is like New but accepts an explicit home directory. Useful for
// tests so the real ~/.sdlaic/ is never touched.
func NewWithHome(home, projectHash, change string) *Store {
	return &Store{
		home:        home,
		projectHash: projectHash,
		change:      change,
		now:         func() time.Time { return time.Now().UTC() },
	}
}

// dir returns the change's gate-state directory.
func (s *Store) dir() string {
	return filepath.Join(s.home, ".sdlaic", "state", s.projectHash, s.change)
}

func (s *Store) metaPath() string    { return filepath.Join(s.dir(), metaFile) }
func (s *Store) reviewPath() string  { return filepath.Join(s.dir(), reviewFile) }
func (s *Store) historyPath() string { return filepath.Join(s.dir(), historyFile) }

func (s *Store) stamp() string { return s.now().Format(time.RFC3339) }

// Load reads and parses meta.json, or returns ErrGateStateNotFound if absent.
func (s *Store) Load() (domain.GatesFile, error) {
	data, err := os.ReadFile(s.metaPath())
	if errors.Is(err, os.ErrNotExist) {
		return domain.GatesFile{}, domain.ErrGateStateNotFound
	}
	if err != nil {
		return domain.GatesFile{}, fmt.Errorf("reading gate state %s: %w", s.metaPath(), err)
	}
	var gf domain.GatesFile
	if err := json.Unmarshal(data, &gf); err != nil {
		return domain.GatesFile{}, fmt.Errorf("parsing gate state %s: %w", s.metaPath(), err)
	}
	return gf, nil
}

// Init creates a fresh meta.json with the four canonical gates. In strict mode
// every gate starts pending; in light/free mode every gate starts skipped
// (the draft-only fast path). It overwrites any existing state.
func (s *Store) Init(workflow domain.WorkflowLevel) (domain.GatesFile, error) {
	initial := domain.GateStatusPending
	if workflow == domain.WorkflowLight || workflow == domain.WorkflowFree {
		initial = domain.GateStatusSkipped
	}

	now := s.stamp()
	gf := domain.GatesFile{
		SchemaVersion: SchemaVersion,
		ProjectHash:   s.projectHash,
		Change:        s.change,
		Workflow:      workflow,
		PipelineState: string(domain.PhaseEmpty),
		Gates:         make(map[string]domain.Gate, len(pipeline)),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	for _, g := range pipeline {
		gf.Gates[g.key] = domain.Gate{
			Phase:    g.phase,
			Artifact: g.artifact,
			Status:   initial,
		}
	}

	if err := s.Save(gf); err != nil {
		return domain.GatesFile{}, err
	}
	return gf, nil
}

// Save writes meta.json and refreshes the human-readable review.md mirror.
func (s *Store) Save(gf domain.GatesFile) error {
	if err := os.MkdirAll(s.dir(), 0755); err != nil {
		return fmt.Errorf("creating gate state dir %s: %w", s.dir(), err)
	}
	gf.UpdatedAt = s.stamp()

	data, err := json.MarshalIndent(gf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling gate state: %w", err)
	}
	if err := os.WriteFile(s.metaPath(), data, 0644); err != nil {
		return fmt.Errorf("writing gate state %s: %w", s.metaPath(), err)
	}
	if err := os.WriteFile(s.reviewPath(), []byte(renderReview(gf)), 0644); err != nil {
		return fmt.Errorf("writing review mirror %s: %w", s.reviewPath(), err)
	}
	return nil
}

// SetGate transitions a single gate to a new status. When verdict is non-nil it
// is recorded on the gate's review; when incAttempt is true the attempt counter
// increments; when the status is approved the approval timestamp is stamped.
// The change's meta.json must already exist (see Init).
func (s *Store) SetGate(key string, status domain.GateStatus, verdict *domain.Verdict, incAttempt bool) (domain.GatesFile, error) {
	meta, ok := gateFor(key)
	if !ok {
		return domain.GatesFile{}, fmt.Errorf("unknown gate %q; valid: %v", key, GateKeys())
	}

	gf, err := s.Load()
	if err != nil {
		return domain.GatesFile{}, err
	}

	g := gf.Gates[key]
	g.Phase = meta.phase
	g.Artifact = meta.artifact
	g.Status = status
	if incAttempt {
		g.Attempts++
	}
	if verdict != nil {
		g.Review.Verdict = *verdict
		g.Review.ReviewedAt = ptr(s.stamp())
	}
	if status == domain.GateStatusApproved {
		g.ApprovedAt = ptr(s.stamp())
	}
	gf.Gates[key] = g

	gf.CurrentGate = key
	gf.PipelineState = string(meta.phase)

	if err := s.Save(gf); err != nil {
		return domain.GatesFile{}, err
	}
	return gf, nil
}

// AppendHistory appends a re-entry event to history.jsonl and mirrors it into
// meta.json's History (when meta.json exists).
func (s *Store) AppendHistory(ev domain.ReEntryEvent) error {
	if err := os.MkdirAll(s.dir(), 0755); err != nil {
		return fmt.Errorf("creating gate state dir %s: %w", s.dir(), err)
	}

	line, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshaling re-entry event: %w", err)
	}
	f, err := os.OpenFile(s.historyPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening history %s: %w", s.historyPath(), err)
	}
	w := bufio.NewWriter(f)
	if _, err := w.Write(append(line, '\n')); err != nil {
		f.Close()
		return fmt.Errorf("writing history: %w", err)
	}
	if err := w.Flush(); err != nil {
		f.Close()
		return fmt.Errorf("flushing history: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("closing history: %w", err)
	}

	gf, err := s.Load()
	if errors.Is(err, domain.ErrGateStateNotFound) {
		return nil // jsonl is the durable log; meta mirror is best-effort
	}
	if err != nil {
		return err
	}
	gf.History = append(gf.History, ev)
	return s.Save(gf)
}

// renderReview produces the human-readable review.md mirror of the latest state.
func renderReview(gf domain.GatesFile) string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "# Gate Review — %s\n\n", gf.Change)
	fmt.Fprintf(b, "- Workflow: `%s`\n", gf.Workflow)
	fmt.Fprintf(b, "- Pipeline state: `%s`\n", gf.PipelineState)
	fmt.Fprintf(b, "- Current gate: `%s`\n", gf.CurrentGate)
	fmt.Fprintf(b, "- Updated: %s\n\n", gf.UpdatedAt)

	fmt.Fprintf(b, "| Gate | Status | Verdict | Attempts | Approved |\n")
	fmt.Fprintf(b, "|------|--------|---------|----------|----------|\n")
	for _, key := range GateKeys() {
		g := gf.Gates[key]
		approved := "—"
		if g.ApprovedAt != nil {
			approved = *g.ApprovedAt
		}
		verdict := string(g.Review.Verdict)
		if verdict == "" {
			verdict = "—"
		}
		fmt.Fprintf(b, "| %s | %s | %s | %d | %s |\n", key, g.Status, verdict, g.Attempts, approved)
	}

	for _, key := range GateKeys() {
		g := gf.Gates[key]
		if len(g.Review.Findings) == 0 {
			continue
		}
		fmt.Fprintf(b, "\n## Findings — %s\n\n", key)
		for _, f := range g.Review.Findings {
			fmt.Fprintf(b, "- **%s** (%s): %s\n", f.Severity, f.Evidence, f.Message)
		}
	}
	return b.String()
}

func ptr(s string) *string { return &s }

// defaultHomeDir returns the user's home directory, mirroring the workspace pkg.
func defaultHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp"
	}
	return home
}

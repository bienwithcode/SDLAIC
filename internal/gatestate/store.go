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

// defaultGates builds a fresh GatesFile with the four canonical gates, without
// writing it. In strict mode every gate starts pending; in light/free mode every
// gate starts skipped (the draft-only fast path).
func (s *Store) defaultGates(workflow domain.WorkflowLevel) domain.GatesFile {
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
	return gf
}

// Init creates and persists a fresh meta.json with the four canonical gates.
// It overwrites any existing state.
func (s *Store) Init(workflow domain.WorkflowLevel) (domain.GatesFile, error) {
	gf := s.defaultGates(workflow)
	if err := s.Save(gf); err != nil {
		return domain.GatesFile{}, err
	}
	return gf, nil
}

// LoadOrDefault returns the persisted gate state if it exists, or an in-memory
// default (NOT written to disk) otherwise. The bool reports whether state
// already existed. Use this for read-only queries such as `gate status`.
func (s *Store) LoadOrDefault(workflow domain.WorkflowLevel) (domain.GatesFile, bool, error) {
	gf, err := s.Load()
	if errors.Is(err, domain.ErrGateStateNotFound) {
		return s.defaultGates(workflow), false, nil
	}
	if err != nil {
		return domain.GatesFile{}, false, err
	}
	return gf, true, nil
}

// ReEnter applies the First Impact Point Principle (§5): the gate at fromKey is
// reset to pending (its artifact will be re-drafted) and every downstream gate
// is marked superseded and reset to pending. The event is appended to history.
// meta.json must already exist.
func (s *Store) ReEnter(fromKey, reason string) (domain.GatesFile, error) {
	fromIdx := -1
	for i, g := range pipeline {
		if g.key == fromKey {
			fromIdx = i
			break
		}
	}
	if fromIdx == -1 {
		return domain.GatesFile{}, fmt.Errorf("unknown gate %q; valid: %v", fromKey, GateKeys())
	}

	gf, err := s.Load()
	if err != nil {
		return domain.GatesFile{}, err
	}
	if gf.Gates == nil {
		gf.Gates = make(map[string]domain.Gate, len(pipeline))
	}

	// Reset status honors the workflow so a light/free change never becomes
	// blocked by a re-entry (gates that never gated stay non-blocking).
	reset := resetStatus(gf.Workflow)

	// Reset the re-entered gate: it will be re-drafted and re-reviewed from
	// scratch, so clear its approval, review, and any prior supersede flag.
	from := gf.Gates[fromKey]
	from.Status = reset
	from.ApprovedAt = nil
	from.SkippedAt = nil
	from.Superseded = false
	from.SupersededBy = nil
	from.Review = domain.ReviewRecord{}
	gf.Gates[fromKey] = from

	// Supersede everything downstream and reset it for re-drafting.
	var superseded []string
	for _, g := range pipeline[fromIdx+1:] {
		gate := gf.Gates[g.key]
		gate.Superseded = true
		gate.SupersededBy = ptr(fromKey)
		gate.Status = reset
		gate.ApprovedAt = nil
		gate.SkippedAt = nil
		gate.Review = domain.ReviewRecord{}
		gf.Gates[g.key] = gate
		superseded = append(superseded, g.key)
	}

	gf.CurrentGate = fromKey
	gf.PipelineState = string(pipeline[fromIdx].phase)

	ev := domain.ReEntryEvent{
		FromPhase:       pipeline[fromIdx].phase,
		Reason:          reason,
		At:              s.stamp(),
		SupersededGates: superseded,
	}
	gf.History = append(gf.History, ev)

	// Append the durable jsonl line first (it is the source log), then persist
	// meta.json — so a failure never leaves meta claiming an event with no line.
	if err := s.appendHistoryLine(ev); err != nil {
		return domain.GatesFile{}, err
	}
	gf.UpdatedAt = s.stamp()
	if err := s.Save(gf); err != nil {
		return domain.GatesFile{}, err
	}
	return gf, nil
}

// resetStatus returns the status a gate should hold after a re-entry, given the
// change's workflow. Light/free gates stay skipped (never block); strict gates
// return to pending so they must be re-approved.
func resetStatus(workflow domain.WorkflowLevel) domain.GateStatus {
	if workflow == domain.WorkflowLight || workflow == domain.WorkflowFree {
		return domain.GateStatusSkipped
	}
	return domain.GateStatusPending
}

// Save writes meta.json and refreshes the human-readable review.md mirror.
// Callers are expected to have stamped gf.UpdatedAt (see Init/SetGate/ReEnter)
// so the returned struct matches what is persisted. Writes are atomic
// (temp file + rename) so a crash mid-write cannot truncate the source of truth.
func (s *Store) Save(gf domain.GatesFile) error {
	if err := os.MkdirAll(s.dir(), 0755); err != nil {
		return fmt.Errorf("creating gate state dir %s: %w", s.dir(), err)
	}

	data, err := json.MarshalIndent(gf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling gate state: %w", err)
	}
	if err := writeFileAtomic(s.metaPath(), data); err != nil {
		return fmt.Errorf("writing gate state %s: %w", s.metaPath(), err)
	}
	if err := writeFileAtomic(s.reviewPath(), []byte(renderReview(gf))); err != nil {
		return fmt.Errorf("writing review mirror %s: %w", s.reviewPath(), err)
	}
	return nil
}

// writeFileAtomic writes data to a temp file in the same directory then renames
// it over the target, so a reader never observes a partially written file.
func writeFileAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
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

	// A verdict implies exactly one lifecycle status; reject contradictions so a
	// rejecting verdict can never be persisted alongside an approving status.
	if verdict != nil && status != verdict.ToGateStatus() {
		return domain.GatesFile{}, fmt.Errorf(
			"gate status %q is inconsistent with verdict %q (verdict implies %q)",
			status, *verdict, verdict.ToGateStatus())
	}

	gf, err := s.Load()
	if err != nil {
		return domain.GatesFile{}, err
	}
	if gf.Gates == nil {
		gf.Gates = make(map[string]domain.Gate, len(pipeline))
	}

	g := gf.Gates[key]
	g.Phase = meta.phase
	g.Artifact = meta.artifact
	g.Status = status
	if status == domain.GateStatusSkipped {
		g.SkippedAt = ptr(s.stamp()) // mark as an EXPLICIT skip so it survives into strict
	} else {
		g.SkippedAt = nil // clear any stale marker on non-skip transitions
	}
	if incAttempt {
		g.Attempts++
	}
	if verdict != nil {
		g.Review.Verdict = *verdict
		g.Review.ReviewedAt = ptr(s.stamp())
	} else {
		g.Review = domain.ReviewRecord{} // no verdict ⇒ wipe any stale review record (matches ReEnter)
	}
	if status == domain.GateStatusApproved {
		g.ApprovedAt = ptr(s.stamp())
	} else {
		g.ApprovedAt = nil // clear stale approval on non-approved transitions
	}
	gf.Gates[key] = g

	gf.CurrentGate = key
	gf.PipelineState = string(meta.phase)

	gf.UpdatedAt = s.stamp()
	if err := s.Save(gf); err != nil {
		return domain.GatesFile{}, err
	}
	return gf, nil
}

// appendHistoryLine appends one event as a JSON line to history.jsonl.
func (s *Store) appendHistoryLine(ev domain.ReEntryEvent) error {
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
	return nil
}

// AppendHistory appends a re-entry event to history.jsonl and mirrors it into
// meta.json's History (when meta.json exists).
func (s *Store) AppendHistory(ev domain.ReEntryEvent) error {
	if err := s.appendHistoryLine(ev); err != nil {
		return err
	}

	gf, err := s.Load()
	if errors.Is(err, domain.ErrGateStateNotFound) {
		return nil // jsonl is the durable log; meta mirror is best-effort
	}
	if err != nil {
		return err
	}
	gf.History = append(gf.History, ev)
	gf.UpdatedAt = s.stamp()
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

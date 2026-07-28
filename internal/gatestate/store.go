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
	"sort"
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

// gateMeta describes one gate in the pipeline: its key, its pipeline tier, the
// phase it satisfies, and the artifact it gates.
type gateMeta struct {
	key      string
	tier     domain.PhaseTier
	phase    domain.Phase
	artifact string
}

// legacyPipeline is the single-spec fallback used when a store has no resolved
// capabilities (a non-user-facing change with no specs/ directory). It preserves
// the original one-"spec"-gate behavior unchanged.
var legacyPipeline = []gateMeta{
	{"proposal", domain.TierProposal, domain.PhaseProposed, "proposal.md"},
	{"spec", domain.TierSpec, domain.PhaseSpecified, "specs/<capability>/spec.md"},
	{"design", domain.TierDesign, domain.PhaseDesigned, "design.md"},
	{"tasks", domain.TierTasks, domain.PhasePlanned, "tasks.md"},
}

// Store is a handle to one change's gate-state directory.
type Store struct {
	home         string
	projectHash  string
	change       string
	capabilities []string // resolved capability names; empty → legacy single-spec pipeline
	now          func() time.Time
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

// SetCapabilities configures the store for per-capability spec gates. When caps
// is non-empty the Spec tier expands to one spec:<capability> gate per entry;
// when empty the store uses the legacy single "spec" gate. The capability list
// is resolved by the caller (the cmd layer, from the change's specs/ directory),
// so this package stays free of filesystem reads. Returns the store for chaining.
func (s *Store) SetCapabilities(caps []string) *Store {
	s.capabilities = caps
	return s
}

// pipeline returns the ordered gates for this change. With capabilities set it
// emits one spec:<cap> gate per capability (per-capability tracking); otherwise
// it falls back to legacyPipeline so changes without specs/ behave unchanged.
func (s *Store) pipeline() []gateMeta {
	if len(s.capabilities) == 0 {
		return legacyPipeline
	}
	pipe := make([]gateMeta, 0, 2+len(s.capabilities))
	pipe = append(pipe, gateMeta{"proposal", domain.TierProposal, domain.PhaseProposed, "proposal.md"})
	for _, c := range s.capabilities {
		pipe = append(pipe, gateMeta{
			key:      "spec:" + c,
			tier:     domain.TierSpec,
			phase:    domain.PhaseSpecified,
			artifact: "specs/" + c + "/spec.md",
		})
	}
	pipe = append(pipe,
		gateMeta{"design", domain.TierDesign, domain.PhaseDesigned, "design.md"},
		gateMeta{"tasks", domain.TierTasks, domain.PhasePlanned, "tasks.md"},
	)
	return pipe
}

// GateKeys returns this change's gate keys in pipeline order.
func (s *Store) GateKeys() []string {
	pipe := s.pipeline()
	keys := make([]string, len(pipe))
	for i, g := range pipe {
		keys[i] = g.key
	}
	return keys
}

// gateFor looks up a gate's metadata by key within this change's pipeline.
func (s *Store) gateFor(key string) (gateMeta, bool) {
	for _, g := range s.pipeline() {
		if g.key == key {
			return g, true
		}
	}
	return gateMeta{}, false
}

// dir returns the change's gate-state directory.
func (s *Store) dir() string {
	return filepath.Join(s.home, ".sdlaic", "state", filepath.Base(s.projectHash), filepath.Base(s.change))
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
	s.reconcile(&gf)
	return gf, nil
}

// reconcile aligns gf.Gates with this change's current pipeline. It drops gates
// that no longer belong (a legacy "spec" key once per-capability gates take over,
// or a spec:<cap> whose directory was removed) and seeds any missing pipeline
// gate with the workflow-appropriate initial status. Reconcile is in-memory only;
// it persists on the next Save (SetGate/ReEnter/Init). This keeps on-disk state
// self-healing as capabilities are added to or removed from specs/.
func (s *Store) reconcile(gf *domain.GatesFile) {
	if gf.Gates == nil {
		return
	}
	pipe := s.pipeline()
	want := make(map[string]gateMeta, len(pipe))
	for _, g := range pipe {
		want[g.key] = g
	}
	for k := range gf.Gates {
		if _, ok := want[k]; !ok {
			delete(gf.Gates, k)
		}
	}
	initial := domain.GateStatusPending
	if gf.Workflow == domain.WorkflowLight || gf.Workflow == domain.WorkflowFree {
		initial = domain.GateStatusSkipped
	}
	for key, meta := range want {
		if _, ok := gf.Gates[key]; !ok {
			gf.Gates[key] = domain.Gate{Phase: meta.phase, Artifact: meta.artifact, Status: initial}
		}
	}
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
	pipe := s.pipeline()
	gf := domain.GatesFile{
		SchemaVersion: SchemaVersion,
		ProjectHash:   s.projectHash,
		Change:        s.change,
		Workflow:      workflow,
		PipelineState: string(domain.PhaseEmpty),
		Gates:         make(map[string]domain.Gate, len(pipe)),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	for _, g := range pipe {
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
func (s *Store) ReEnter(fromKey, reason string, currentWorkflow domain.WorkflowLevel) (domain.GatesFile, error) {
	fromMeta, ok := s.gateFor(fromKey)
	if !ok {
		return domain.GatesFile{}, UnknownGateErr(fromKey, s.GateKeys())
	}

	gf, err := s.Load()
	if err != nil {
		return domain.GatesFile{}, err
	}
	if gf.Gates == nil {
		gf.Gates = make(map[string]domain.Gate, len(s.pipeline()))
	}

	reset := resetStatus(currentWorkflow)

	// Reset the re-entered gate
	from := gf.Gates[fromKey]
	from.Status = reset
	from.ApprovedAt = nil
	from.SkippedAt = nil
	from.Superseded = false
	from.SupersededBy = nil
	from.Review = domain.ReviewRecord{}
	from.Grill = domain.GrillRecord{}
	gf.Gates[fromKey] = from

	// Supersede every gate in a LATER tier (never siblings in the same tier):
	// re-entering spec:auth supersedes design + tasks, not spec:billing, because
	// an independent capability's review should survive a sibling's rework.
	fromOrder := tierOrder(fromMeta.tier)
	var superseded []string
	for _, key := range SortedGateKeys(gf.Gates) {
		if key == fromKey {
			continue
		}
		if tierOrder(domain.TierOf(key)) <= fromOrder {
			continue
		}
		gate := gf.Gates[key]
		gate.Superseded = true
		gate.SupersededBy = ptr(fromKey)
		gate.Status = reset
		gate.ApprovedAt = nil
		gate.SkippedAt = nil
		gate.Review = domain.ReviewRecord{}
		gate.Grill = domain.GrillRecord{}
		gf.Gates[key] = gate
		superseded = append(superseded, key)
	}

	gf.CurrentGate = fromKey
	gf.PipelineState = string(fromMeta.phase)

	ev := domain.ReEntryEvent{
		FromPhase:       fromMeta.phase,
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
	meta, ok := s.gateFor(key)
	if !ok {
		return domain.GatesFile{}, UnknownGateErr(key, s.GateKeys())
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
		gf.Gates = make(map[string]domain.Gate, len(s.pipeline()))
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
	for _, key := range SortedGateKeys(gf.Gates) {
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

	for _, key := range SortedGateKeys(gf.Gates) {
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

// SortedGateKeys returns the keys of gates ordered by pipeline tier then key
// name, so spec:auth and spec:billing cluster together under the Spec tier
// regardless of map iteration order. A legacy "spec" key sorts within the Spec
// tier too. Exported so the cmd layer can render gates consistently.
func SortedGateKeys(gates map[string]domain.Gate) []string {
	keys := make([]string, 0, len(gates))
	for k := range gates {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		oi, oj := tierOrder(domain.TierOf(keys[i])), tierOrder(domain.TierOf(keys[j]))
		if oi != oj {
			return oi < oj
		}
		return keys[i] < keys[j]
	})
	return keys
}

// tierOrder returns the ordinal of a tier for cascade and sort comparisons;
// unknown tiers sort last.
func tierOrder(t domain.PhaseTier) int {
	for i, ot := range domain.OrderedTiers() {
		if ot == t {
			return i
		}
	}
	return len(domain.OrderedTiers())
}

// UnknownGateErr builds a helpful error for an unrecognized gate key. When the
// caller used a bare "spec" but the change has per-capability spec gates, it
// points them at spec:<capability> instead of just listing every key.
func UnknownGateErr(key string, valid []string) error {
	if key == string(domain.TierSpec) {
		var specCaps []string
		for _, k := range valid {
			if strings.HasPrefix(k, "spec:") {
				specCaps = append(specCaps, k)
			}
		}
		if len(specCaps) > 0 {
			return fmt.Errorf("unknown gate %q; this change has per-capability spec gates — use one of: %s", key, strings.Join(specCaps, ", "))
		}
	}
	return fmt.Errorf("unknown gate %q; valid: %v", key, valid)
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

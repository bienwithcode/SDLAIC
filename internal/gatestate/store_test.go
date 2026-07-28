package gatestate

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testHash   = "abc123def456"
	testChange = "JIRA-789"
)

// newTestStore returns a Store rooted at a temp home with a fixed clock.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	home := t.TempDir()
	s := NewWithHome(home, testHash, testChange)
	fixed := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return fixed }
	return s
}

func TestLoadNotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Load()
	assert.ErrorIs(t, err, domain.ErrGateStateNotFound)
}

func TestInitStrictCreatesPendingGates(t *testing.T) {
	s := newTestStore(t)
	gf, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	assert.Equal(t, SchemaVersion, gf.SchemaVersion)
	assert.Equal(t, testHash, gf.ProjectHash)
	assert.Equal(t, testChange, gf.Change)
	assert.Equal(t, domain.WorkflowStrict, gf.Workflow)
	assert.NotEmpty(t, gf.CreatedAt)
	assert.NotEmpty(t, gf.UpdatedAt)

	// The four canonical gates exist and are pending in strict mode.
	for _, key := range []string{"proposal", "spec", "design", "tasks"} {
		require.Contains(t, gf.Gates, key)
		assert.Equal(t, domain.GateStatusPending, gf.Gates[key].Status, "gate %s", key)
	}

	// meta.json is written under the temp home, never in a project repo.
	metaPath := filepath.Join(s.dir(), "meta.json")
	assert.FileExists(t, metaPath)
	assert.Contains(t, metaPath, filepath.Join(".sdlaic", "state", testHash, testChange))

	// Reload round-trips.
	got, err := s.Load()
	require.NoError(t, err)
	assert.Equal(t, gf.Change, got.Change)
	assert.Len(t, got.Gates, 4)
}

func TestInitLightSkipsGates(t *testing.T) {
	s := newTestStore(t)
	gf, err := s.Init(domain.WorkflowLight)
	require.NoError(t, err)
	for key, g := range gf.Gates {
		assert.Equal(t, domain.GateStatusSkipped, g.Status, "gate %s should be skipped in light mode", key)
	}
}

func TestSetGateApprovedStampsApprovedAt(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	gf, err := s.SetGate("proposal", domain.GateStatusApproved, nil, false)
	require.NoError(t, err)

	g := gf.Gates["proposal"]
	assert.Equal(t, domain.GateStatusApproved, g.Status)
	require.NotNil(t, g.ApprovedAt)
	assert.Equal(t, "2026-07-25T10:00:00Z", *g.ApprovedAt)
	assert.Equal(t, "proposal", gf.CurrentGate)
	assert.Equal(t, string(domain.PhaseProposed), gf.PipelineState)

	// review.md mirror is written.
	assert.FileExists(t, filepath.Join(s.dir(), "review.md"))
}

func TestSetGateRecordsVerdictAndAttempts(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	reject := domain.VerdictReject
	_, err = s.SetGate("design", domain.GateStatusFailed, &reject, true)
	require.NoError(t, err)
	gf, err := s.SetGate("design", domain.GateStatusFailed, &reject, true)
	require.NoError(t, err)

	g := gf.Gates["design"]
	assert.Equal(t, domain.GateStatusFailed, g.Status)
	assert.Equal(t, domain.VerdictReject, g.Review.Verdict)
	assert.Equal(t, 2, g.Attempts)
	assert.Nil(t, g.ApprovedAt, "failed gate must not be approved")
}

func TestSetGateUnknownKey(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)
	_, err = s.SetGate("bogus", domain.GateStatusApproved, nil, false)
	assert.Error(t, err)
}

func TestSetGateWithoutInit(t *testing.T) {
	s := newTestStore(t)
	_, err := s.SetGate("proposal", domain.GateStatusApproved, nil, false)
	assert.ErrorIs(t, err, domain.ErrGateStateNotFound)
}

func TestAppendHistory(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	ev := domain.ReEntryEvent{
		FromPhase:       domain.PhaseProposed,
		Reason:          "ticket scope changed",
		At:              "2026-07-25T10:00:00Z",
		SupersededGates: []string{"spec", "design"},
	}
	require.NoError(t, s.AppendHistory(ev))
	require.NoError(t, s.AppendHistory(ev))

	// history.jsonl has one line per event.
	f, err := os.Open(filepath.Join(s.dir(), "history.jsonl"))
	require.NoError(t, err)
	defer f.Close()
	lines := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if strings.TrimSpace(sc.Text()) != "" {
			lines++
		}
	}
	assert.Equal(t, 2, lines)

	// meta.json History mirrors the events.
	gf, err := s.Load()
	require.NoError(t, err)
	assert.Len(t, gf.History, 2)
}

func TestLoadOrDefaultWhenMissing(t *testing.T) {
	s := newTestStore(t)
	gf, existed, err := s.LoadOrDefault(domain.WorkflowStrict)
	require.NoError(t, err)
	assert.False(t, existed)
	assert.Len(t, gf.Gates, 4)
	assert.Equal(t, domain.GateStatusPending, gf.Gates["proposal"].Status)
	// LoadOrDefault must NOT write anything to disk.
	assert.NoFileExists(t, filepath.Join(s.dir(), "meta.json"))
}

func TestLoadOrDefaultWhenPresent(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)
	_, err = s.SetGate("proposal", domain.GateStatusApproved, nil, false)
	require.NoError(t, err)

	gf, existed, err := s.LoadOrDefault(domain.WorkflowStrict)
	require.NoError(t, err)
	assert.True(t, existed)
	assert.Equal(t, domain.GateStatusApproved, gf.Gates["proposal"].Status)
}

func TestReEnterSupersedesDownstream(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)
	// Approve everything first.
	for _, k := range s.GateKeys() {
		_, err := s.SetGate(k, domain.GateStatusApproved, nil, false)
		require.NoError(t, err)
	}

	gf, err := s.ReEnter("spec", "ticket scope changed", domain.WorkflowStrict)
	require.NoError(t, err)

	// The re-entered gate is reset to pending, approval cleared.
	assert.Equal(t, domain.GateStatusPending, gf.Gates["spec"].Status)
	assert.Nil(t, gf.Gates["spec"].ApprovedAt)

	// Downstream gates (design, tasks) are superseded; upstream (proposal) untouched.
	assert.True(t, gf.Gates["design"].Superseded)
	assert.True(t, gf.Gates["tasks"].Superseded)
	require.NotNil(t, gf.Gates["design"].SupersededBy)
	assert.Equal(t, "spec", *gf.Gates["design"].SupersededBy)
	assert.False(t, gf.Gates["proposal"].Superseded)
	assert.Equal(t, domain.GateStatusApproved, gf.Gates["proposal"].Status)

	// History records the event.
	require.Len(t, gf.History, 1)
	assert.Equal(t, []string{"design", "tasks"}, gf.History[0].SupersededGates)
}

func TestReEnterLightKeepsGatesSkipped(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowFree)
	require.NoError(t, err)

	gf, err := s.ReEnter("proposal", "changed", domain.WorkflowFree)
	require.NoError(t, err)
	// In free/light mode, re-entry must NOT introduce a blocking (pending) status.
	for _, k := range s.GateKeys() {
		assert.Equal(t, domain.GateStatusSkipped, gf.Gates[k].Status, "gate %s must stay skipped", k)
	}
}

func TestReEnterClearsSupersededOnReentryPoint(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	// First re-entry from proposal marks spec/design/tasks superseded.
	_, err = s.ReEnter("proposal", "first", domain.WorkflowStrict)
	require.NoError(t, err)
	// Now re-enter spec, which supersedes design and tasks, but also clears the
	// supersede flag on spec itself.
	gf, err := s.ReEnter("spec", "second", domain.WorkflowStrict)
	require.NoError(t, err)

	assert.False(t, gf.Gates["spec"].Superseded, "re-entry point must not stay superseded")
	assert.Nil(t, gf.Gates["spec"].SupersededBy)
	assert.Equal(t, domain.GateStatusPending, gf.Gates["spec"].Status)
}

func TestReEnterClearsStaleReview(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)
	approve := domain.VerdictApprove
	_, err = s.SetGate("design", domain.GateStatusApproved, &approve, false)
	require.NoError(t, err)

	gf, err := s.ReEnter("proposal", "scope changed", domain.WorkflowStrict)
	require.NoError(t, err)
	// design was superseded and reset — its stale APPROVE verdict must be cleared.
	assert.Equal(t, domain.Verdict(""), gf.Gates["design"].Review.Verdict)
}

func TestSetGateSurvivesNullGatesMap(t *testing.T) {
	s := newTestStore(t)
	// Write a meta.json with a null gates map (e.g. hand-edited / migrated).
	require.NoError(t, os.MkdirAll(s.dir(), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(s.dir(), "meta.json"),
		[]byte(`{"schema_version":1,"change":"CHG-1","workflow":"strict","gates":null}`), 0644))

	gf, err := s.SetGate("proposal", domain.GateStatusApproved, nil, false)
	require.NoError(t, err) // must not panic on a nil map
	assert.Equal(t, domain.GateStatusApproved, gf.Gates["proposal"].Status)
}

func TestReEnterUnknownKey(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)
	_, err = s.ReEnter("bogus", "x", domain.WorkflowStrict)
	assert.Error(t, err)
}

func TestSchemaVersionConstant(t *testing.T) {
	assert.Equal(t, 1, SchemaVersion)
}

func TestGateKeysOrdered(t *testing.T) {
	// Legacy store (no capabilities) keeps the original single-spec pipeline.
	s := newTestStore(t)
	assert.Equal(t, []string{"proposal", "spec", "design", "tasks"}, s.GateKeys())
}

func TestErrorsIsWiring(t *testing.T) {
	// sanity: ErrGateStateNotFound is a real sentinel
	assert.True(t, errors.Is(domain.ErrGateStateNotFound, domain.ErrGateStateNotFound))
}

func TestSetGate_SkippedStampsSkippedAt(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	gf, err := s.SetGate("proposal", domain.GateStatusSkipped, nil, false)
	require.NoError(t, err)

	require.NotNil(t, gf.Gates["proposal"].SkippedAt, "explicit skip must stamp SkippedAt")
	assert.Equal(t, "2026-07-25T10:00:00Z", *gf.Gates["proposal"].SkippedAt)
}

func TestSetGate_NonSkippedClearsSkippedAt(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	_, err = s.SetGate("proposal", domain.GateStatusSkipped, nil, false)
	require.NoError(t, err)
	// Transition to approved: the stale skip marker must clear so the gate is no
	// longer treated as an explicit skip.
	gf, err := s.SetGate("proposal", domain.GateStatusApproved, nil, false)
	require.NoError(t, err)
	assert.Nil(t, gf.Gates["proposal"].SkippedAt)
}

func TestSetGate_RejectsMismatchedStatusVerdict(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	reject := domain.VerdictReject
	_, err = s.SetGate("proposal", domain.GateStatusApproved, &reject, false)
	assert.Error(t, err, "approved status with a REJECT verdict must be rejected")
}

func TestSetGate_RejectsSkippedWithVerdict(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	approve := domain.VerdictApprove
	_, err = s.SetGate("proposal", domain.GateStatusSkipped, &approve, false)
	assert.Error(t, err, "skipped status with a verdict must be rejected")
}

func TestReEnter_ClearsSkippedAt(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	// Explicitly skip spec, then re-enter upstream at proposal.
	gfSkip, err := s.SetGate("spec", domain.GateStatusSkipped, nil, false)
	require.NoError(t, err)
	require.NotNil(t, gfSkip.Gates["spec"].SkippedAt)

	gf, err := s.ReEnter("proposal", "scope changed", domain.WorkflowStrict)
	require.NoError(t, err)
	assert.Nil(t, gf.Gates["spec"].SkippedAt, "a reset gate must clear its skip marker")
}

func TestSetGate_NonApprovedClearsApprovedAt(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	_, err = s.SetGate("proposal", domain.GateStatusApproved, nil, false)
	require.NoError(t, err)

	// Transition to a non-approved status: the stale approval timestamp must clear.
	gf, err := s.SetGate("proposal", domain.GateStatusFailed, nil, false)
	require.NoError(t, err)
	assert.Equal(t, domain.GateStatusFailed, gf.Gates["proposal"].Status)
	assert.Nil(t, gf.Gates["proposal"].ApprovedAt, "non-approved status must clear approved_at")
}

func TestSetGate_NoVerdictWipesStaleReview(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	approve := domain.VerdictApprove
	_, err = s.SetGate("proposal", domain.GateStatusApproved, &approve, false)
	require.NoError(t, err)

	// Re-transition without a verdict: the stale APPROVE verdict + review timestamp
	// (and the approval) must all clear, so the persisted record is not contradictory.
	gf, err := s.SetGate("proposal", domain.GateStatusPending, nil, false)
	require.NoError(t, err)
	g := gf.Gates["proposal"]
	assert.Equal(t, domain.GateStatusPending, g.Status)
	assert.Equal(t, domain.Verdict(""), g.Review.Verdict, "stale verdict must be wiped when no verdict is supplied")
	assert.Nil(t, g.Review.ReviewedAt)
	assert.Nil(t, g.ApprovedAt)
}

func TestSetGate_FailedToApprovedNoVerdictClearsStaleReject(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	reject := domain.VerdictReject
	_, err = s.SetGate("design", domain.GateStatusFailed, &reject, false)
	require.NoError(t, err)

	// Approve without supplying a verdict: the stale REJECT must clear and the
	// approval stamps fresh — no approved gate may carry a rejecting verdict.
	gf, err := s.SetGate("design", domain.GateStatusApproved, nil, false)
	require.NoError(t, err)
	g := gf.Gates["design"]
	assert.Equal(t, domain.GateStatusApproved, g.Status)
	require.NotNil(t, g.ApprovedAt)
	assert.Equal(t, domain.Verdict(""), g.Review.Verdict, "stale rejecting verdict must not survive an approval")
}

func TestPipelinePerCapability(t *testing.T) {
	s := newTestStore(t)
	s.SetCapabilities([]string{"auth", "billing"})

	assert.Equal(t,
		[]string{"proposal", "spec:auth", "spec:billing", "design", "tasks"},
		s.GateKeys())

	gf, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)
	assert.Len(t, gf.Gates, 5)
	for _, k := range s.GateKeys() {
		require.Contains(t, gf.Gates, k, "gate %s should be seeded", k)
	}
	assert.Equal(t, "specs/auth/spec.md", gf.Gates["spec:auth"].Artifact)
	assert.Equal(t, "specs/billing/spec.md", gf.Gates["spec:billing"].Artifact)
	assert.Equal(t, domain.PhaseSpecified, gf.Gates["spec:auth"].Phase)
}

func TestSetGatePerCapability(t *testing.T) {
	s := newTestStore(t)
	s.SetCapabilities([]string{"auth", "billing"})
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	// spec:<capability> is a valid gate and can be approved independently.
	gf, err := s.SetGate("spec:auth", domain.GateStatusApproved, nil, false)
	require.NoError(t, err)
	assert.Equal(t, domain.GateStatusApproved, gf.Gates["spec:auth"].Status)
	// Sibling capability is untouched by this set.
	assert.NotEqual(t, domain.GateStatusApproved, gf.Gates["spec:billing"].Status)

	// Bare "spec" is rejected; the error names the per-capability keys.
	_, err = s.SetGate("spec", domain.GateStatusApproved, nil, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "spec:auth")
	assert.Contains(t, err.Error(), "spec:billing")
}

func TestReEnterSpecCapabilitySparesSibling(t *testing.T) {
	s := newTestStore(t)
	s.SetCapabilities([]string{"auth", "billing"})
	_, err := s.Init(domain.WorkflowStrict)
	require.NoError(t, err)
	// Approve every gate first.
	for _, k := range s.GateKeys() {
		_, err := s.SetGate(k, domain.GateStatusApproved, nil, false)
		require.NoError(t, err)
	}

	// Re-enter at spec:auth: it resets; design+tasks supersede; but spec:billing
	// (a sibling capability) must survive untouched — independent review lifecycle.
	gf, err := s.ReEnter("spec:auth", "auth requirements changed", domain.WorkflowStrict)
	require.NoError(t, err)

	assert.Equal(t, domain.GateStatusPending, gf.Gates["spec:auth"].Status)
	assert.Nil(t, gf.Gates["spec:auth"].ApprovedAt)

	assert.True(t, gf.Gates["design"].Superseded, "later-tier design must supersede")
	assert.True(t, gf.Gates["tasks"].Superseded, "later-tier tasks must supersede")

	assert.False(t, gf.Gates["spec:billing"].Superseded, "sibling capability must not be superseded")
	assert.Equal(t, domain.GateStatusApproved, gf.Gates["spec:billing"].Status, "sibling capability keeps its verdict")
	assert.False(t, gf.Gates["proposal"].Superseded, "earlier-tier proposal must not supersede")

	require.Len(t, gf.History, 1)
	assert.Contains(t, gf.History[0].SupersededGates, "design")
	assert.Contains(t, gf.History[0].SupersededGates, "tasks")
	assert.NotContains(t, gf.History[0].SupersededGates, "spec:billing")
}

func TestReconcileMigratesLegacySpecKey(t *testing.T) {
	s := newTestStore(t)
	s.SetCapabilities([]string{"auth"})

	// A legacy meta.json from the old single-spec model: a bare "spec" gate
	// (approved) and no spec:auth entry.
	require.NoError(t, os.MkdirAll(s.dir(), 0755))
	legacy := `{"schema_version":1,"project_hash":"abc123def456","change":"JIRA-789","workflow":"strict","gates":{` +
		`"proposal":{"phase":"PROPOSED","artifact":"proposal.md","status":"approved"},` +
		`"spec":{"phase":"SPECIFIED","artifact":"specs/x/spec.md","status":"approved"},` +
		`"design":{"phase":"DESIGNED","artifact":"design.md","status":"pending"},` +
		`"tasks":{"phase":"PLANNED","artifact":"tasks.md","status":"pending"}}}`
	require.NoError(t, os.WriteFile(filepath.Join(s.dir(), "meta.json"), []byte(legacy), 0644))

	gf, err := s.Load()
	require.NoError(t, err)

	// Legacy "spec" is dropped; spec:auth is seeded pending — it must be
	// re-reviewed under the per-capability model, NOT inherit the old approval.
	assert.NotContains(t, gf.Gates, "spec")
	assert.Contains(t, gf.Gates, "spec:auth")
	assert.Equal(t, domain.GateStatusPending, gf.Gates["spec:auth"].Status, "migrated capability must be re-reviewed")
	assert.Equal(t, "specs/auth/spec.md", gf.Gates["spec:auth"].Artifact)
	// Non-spec gates survive reconciliation.
	assert.Equal(t, domain.GateStatusApproved, gf.Gates["proposal"].Status)
}

package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/bienwithcode/SDLAIC/internal/gatestate"
)

// proposedChange returns a change dir at phase PROPOSED (context + proposal).
func proposedChange(t *testing.T) string {
	return setupChangeDir(t, map[string]string{
		"context.md":  "# Context\n\nReal content.",
		"proposal.md": "# Proposal\n\nReal proposal.",
	})
}

func newStore(t *testing.T) *gatestate.Store {
	t.Helper()
	return gatestate.NewWithHome(t.TempDir(), "hash123", "CHG-1")
}

func TestAnalyzeGatedPhase_StrictBlockedWhenPending(t *testing.T) {
	changeDir := proposedChange(t)
	store := newStore(t)
	_, err := store.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	res, err := AnalyzeGatedPhase(changeDir, store, domain.WorkflowStrict)
	require.NoError(t, err)
	assert.Equal(t, domain.PhaseProposed, res.Phase)
	assert.True(t, res.Blocked, "pending proposal gate should block")
	assert.Equal(t, "proposal", res.BlockedAt)
}

func TestAnalyzeGatedPhase_StrictUnblockedWhenApproved(t *testing.T) {
	changeDir := proposedChange(t)
	store := newStore(t)
	_, err := store.Init(domain.WorkflowStrict)
	require.NoError(t, err)
	_, err = store.SetGate("proposal", domain.GateStatusApproved, nil, false)
	require.NoError(t, err)

	res, err := AnalyzeGatedPhase(changeDir, store, domain.WorkflowStrict)
	require.NoError(t, err)
	assert.Equal(t, domain.PhaseProposed, res.Phase)
	assert.False(t, res.Blocked)
	assert.Empty(t, res.BlockedAt)
}

func TestAnalyzeGatedPhase_SkippedUnblocks(t *testing.T) {
	changeDir := proposedChange(t)
	store := newStore(t)
	_, err := store.Init(domain.WorkflowStrict)
	require.NoError(t, err)
	_, err = store.SetGate("proposal", domain.GateStatusSkipped, nil, false)
	require.NoError(t, err)

	res, err := AnalyzeGatedPhase(changeDir, store, domain.WorkflowStrict)
	require.NoError(t, err)
	assert.False(t, res.Blocked)
}

func TestAnalyzeGatedPhase_LightNeverBlocks(t *testing.T) {
	changeDir := proposedChange(t)
	store := newStore(t) // no meta.json written

	res, err := AnalyzeGatedPhase(changeDir, store, domain.WorkflowLight)
	require.NoError(t, err)
	assert.Equal(t, domain.PhaseProposed, res.Phase)
	assert.False(t, res.Blocked, "light workflow gates default to skipped")
}

func TestAnalyzeGatedPhase_StrictMissingMetaBlocks(t *testing.T) {
	changeDir := proposedChange(t)
	store := newStore(t) // no meta.json written

	res, err := AnalyzeGatedPhase(changeDir, store, domain.WorkflowStrict)
	require.NoError(t, err)
	assert.True(t, res.Blocked, "missing meta.json in strict ⇒ pending ⇒ blocked")
	assert.Equal(t, "proposal", res.BlockedAt)
}

func TestAnalyzeGatedPhase_ContextHasNoGate(t *testing.T) {
	changeDir := setupChangeDir(t, map[string]string{
		"context.md": "# Context\n\nReal content.",
	})
	store := newStore(t)
	_, err := store.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	res, err := AnalyzeGatedPhase(changeDir, store, domain.WorkflowStrict)
	require.NoError(t, err)
	assert.Equal(t, domain.PhaseContext, res.Phase)
	assert.False(t, res.Blocked, "CONTEXT has no gate")
}

func TestAnalyzeGatedPhase_EmptyHasNoGate(t *testing.T) {
	changeDir := setupChangeDir(t, nil)
	store := newStore(t)
	_, err := store.Init(domain.WorkflowStrict)
	require.NoError(t, err)

	res, err := AnalyzeGatedPhase(changeDir, store, domain.WorkflowStrict)
	require.NoError(t, err)
	assert.Equal(t, domain.PhaseEmpty, res.Phase)
	assert.False(t, res.Blocked)
}

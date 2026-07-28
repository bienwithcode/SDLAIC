package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// initGateWorkspace sets a temp HOME (so the gate-state store writes there),
// initializes a workspace, and creates an active change.
func initGateWorkspace(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	resetInitFlags()
	initWorkspaceForTest(t)
	resetGateFlags()
	_, err := ExecuteCommand(rootCmd, "new", "change", "GATE-TEST")
	require.NoError(t, err)
	resetGateFlags()
}

func TestGateStatus_DefaultWhenNoState(t *testing.T) {
	initGateWorkspace(t)

	output, err := ExecuteCommand(rootCmd, "gate", "status", "--json")
	require.NoError(t, err)

	var gf domain.GatesFile
	require.NoError(t, json.Unmarshal([]byte(output), &gf))
	assert.Len(t, gf.Gates, 4)
	assert.Equal(t, domain.GateStatusPending, gf.Gates["proposal"].Status)
}

func TestGateSet_ApprovesGate(t *testing.T) {
	initGateWorkspace(t)

	_, err := ExecuteCommand(rootCmd, "gate", "set", "--phase", "proposal", "--status", "approved")
	require.NoError(t, err)

	resetGateFlags()
	output, err := ExecuteCommand(rootCmd, "gate", "status", "--json")
	require.NoError(t, err)

	var gf domain.GatesFile
	require.NoError(t, json.Unmarshal([]byte(output), &gf))
	assert.Equal(t, domain.GateStatusApproved, gf.Gates["proposal"].Status)
	require.NotNil(t, gf.Gates["proposal"].ApprovedAt)
	assert.Equal(t, "proposal", gf.CurrentGate)
}

func TestGateSet_RecordsVerdictAndAttempt(t *testing.T) {
	initGateWorkspace(t)

	_, err := ExecuteCommand(rootCmd, "gate", "set",
		"--phase", "design", "--status", "failed", "--verdict", "REJECT", "--attempt")
	require.NoError(t, err)

	resetGateFlags()
	output, err := ExecuteCommand(rootCmd, "gate", "status", "--json")
	require.NoError(t, err)
	var gf domain.GatesFile
	require.NoError(t, json.Unmarshal([]byte(output), &gf))
	assert.Equal(t, domain.GateStatusFailed, gf.Gates["design"].Status)
	assert.Equal(t, domain.VerdictReject, gf.Gates["design"].Review.Verdict)
	assert.Equal(t, 1, gf.Gates["design"].Attempts)
}

func TestGateSet_InvalidStatus(t *testing.T) {
	initGateWorkspace(t)
	_, err := ExecuteCommand(rootCmd, "gate", "set", "--phase", "proposal", "--status", "nonsense")
	assert.Error(t, err)
}

func TestGateSet_InvalidPhase(t *testing.T) {
	initGateWorkspace(t)
	_, err := ExecuteCommand(rootCmd, "gate", "set", "--phase", "bogus", "--status", "approved")
	assert.Error(t, err)
}

func TestGateReentry_SupersedesDownstream(t *testing.T) {
	initGateWorkspace(t)

	// Approve all gates.
	for _, key := range []string{"proposal", "spec", "design", "tasks"} {
		resetGateFlags()
		_, err := ExecuteCommand(rootCmd, "gate", "set", "--phase", key, "--status", "approved")
		require.NoError(t, err)
	}

	resetGateFlags()
	_, err := ExecuteCommand(rootCmd, "gate", "reentry", "--from", "spec", "--reason", "ticket changed")
	require.NoError(t, err)

	resetGateFlags()
	output, err := ExecuteCommand(rootCmd, "gate", "status", "--json")
	require.NoError(t, err)
	var gf domain.GatesFile
	require.NoError(t, json.Unmarshal([]byte(output), &gf))

	assert.Equal(t, domain.GateStatusPending, gf.Gates["spec"].Status)
	assert.True(t, gf.Gates["design"].Superseded)
	assert.True(t, gf.Gates["tasks"].Superseded)
	assert.False(t, gf.Gates["proposal"].Superseded)
	require.Len(t, gf.History, 1)
}

func TestGateStatus_Human(t *testing.T) {
	initGateWorkspace(t)
	output, err := ExecuteCommand(rootCmd, "gate", "status")
	require.NoError(t, err)
	assert.Contains(t, output, "proposal")
	assert.Contains(t, output, "GATE-TEST")
}

// TestGateStatus_StrictBlocksAfterLightInit exercises the headline B1 fix
// end-to-end. Drafted under light, the first `gate set` lazily runs Init(light)
// which AUTO-skips every gate; proposal is then explicitly marked skipped.
// After tightening to strict: proposal (explicit skip) still passes, but the
// auto-skipped gates (spec/design/tasks — never reviewed) must read as BLOCKED.
func TestGateStatus_StrictBlocksAfterLightInit(t *testing.T) {
	initGateWorkspace(t)

	_, err := ExecuteCommand(rootCmd, "config", "set", "workflow", "light")
	require.NoError(t, err)
	resetGateFlags()
	_, err = ExecuteCommand(rootCmd, "gate", "set", "--phase", "proposal", "--status", "skipped")
	require.NoError(t, err)

	resetGateFlags()
	_, err = ExecuteCommand(rootCmd, "config", "set", "workflow", "strict")
	require.NoError(t, err)

	resetGateFlags()
	output, err := ExecuteCommand(rootCmd, "gate", "status")
	require.NoError(t, err)

	assert.Contains(t, output, "[✓] proposal", "explicit skip survives into strict")
	assert.Contains(t, output, "[ ] spec", "auto-skipped gate must block under strict")
	assert.NotContains(t, output, "[✓] spec", "auto-skipped gate must not pass under strict")
}

// TestGateSet_PerCapabilitySpecGate exercises the P1 per-capability model
// end-to-end: when a change has a specs/ tree, the cmd layer resolves capability
// directories and the Spec tier expands to one spec:<capability> gate each.
func TestGateSet_PerCapabilitySpecGate(t *testing.T) {
	initGateWorkspace(t)

	// Give the change two capability directories.
	wd, err := os.Getwd()
	require.NoError(t, err)
	changeDir := filepath.Join(wd, ".sdlaic", "changes", "GATE-TEST")
	require.NoError(t, os.MkdirAll(filepath.Join(changeDir, "specs", "auth"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(changeDir, "specs", "billing"), 0755))

	// gate status reflects per-capability spec gates; the bare legacy "spec" key
	// is gone.
	resetGateFlags()
	output, err := ExecuteCommand(rootCmd, "gate", "status", "--json")
	require.NoError(t, err)
	var gf domain.GatesFile
	require.NoError(t, json.Unmarshal([]byte(output), &gf))
	assert.Contains(t, gf.Gates, "spec:auth")
	assert.Contains(t, gf.Gates, "spec:billing")
	assert.NotContains(t, gf.Gates, "spec", "bare legacy spec key must not appear when capabilities exist")

	// One capability can be approved independently of its sibling.
	resetGateFlags()
	_, err = ExecuteCommand(rootCmd, "gate", "set", "--phase", "spec:auth", "--status", "approved")
	require.NoError(t, err)

	resetGateFlags()
	output, err = ExecuteCommand(rootCmd, "gate", "status", "--json")
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal([]byte(output), &gf))
	assert.Equal(t, domain.GateStatusApproved, gf.Gates["spec:auth"].Status)
	assert.NotEqual(t, domain.GateStatusApproved, gf.Gates["spec:billing"].Status, "sibling capability stays unapproved")

	// Bare "spec" is rejected with a pointer to the per-capability keys.
	resetGateFlags()
	_, err = ExecuteCommand(rootCmd, "gate", "set", "--phase", "spec", "--status", "approved")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "spec:auth")
}

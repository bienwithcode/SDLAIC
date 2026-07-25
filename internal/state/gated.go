package state

import (
	"fmt"

	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/bienwithcode/SDLAIC/internal/gatestate"
)

// GatedResult reports both the artifact-derived phase and whether progression
// past that phase is currently blocked by an unpassed gate.
type GatedResult struct {
	Phase     domain.Phase // artifact-derived phase (same as AnalyzePhase)
	Blocked   bool         // true when the current phase's gate is not passing
	BlockedAt string       // gate key that is blocking, empty when not blocked
}

// phaseGate maps an artifact phase to the gate that must pass before the change
// may progress out of it. Phases without a gate (EMPTY, CONTEXT, IMPLEMENTED)
// are absent from the map.
var phaseGate = map[domain.Phase]string{
	domain.PhaseProposed:  "proposal",
	domain.PhaseSpecified: "spec",
	domain.PhaseDesigned:  "design",
	domain.PhasePlanned:   "tasks",
}

// AnalyzeGatedPhase couples artifact presence (project-local) with gate state
// (global ~/.sdlaic/state/meta.json). A phase is unblocked only when its
// artifact exists AND its gate is passing (approved or skipped). In light/free
// workflows every gate defaults to skipped, so progression is never blocked.
func AnalyzeGatedPhase(changeDir string, store *gatestate.Store, workflow domain.WorkflowLevel) (GatedResult, error) {
	phase, err := AnalyzePhase(changeDir)
	if err != nil {
		return GatedResult{}, err
	}

	res := GatedResult{Phase: phase}

	// Light/free workflows never block progression, regardless of any persisted
	// gate status (e.g. a change born strict then switched, or re-entered).
	if workflow == domain.WorkflowLight || workflow == domain.WorkflowFree {
		return res, nil
	}

	gateKey, hasGate := phaseGate[phase]
	if !hasGate {
		return res, nil // EMPTY/CONTEXT/IMPLEMENTED have no gate to clear
	}

	gf, _, err := store.LoadOrDefault(workflow)
	if err != nil {
		return GatedResult{}, fmt.Errorf("loading gate state: %w", err)
	}

	if !gf.Gates[gateKey].IsPassingFor(workflow) {
		res.Blocked = true
		res.BlockedAt = gateKey
	}
	return res, nil
}

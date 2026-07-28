package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/bienwithcode/SDLAIC/internal/gatestate"
	"github.com/bienwithcode/SDLAIC/internal/state"
)

var (
	gateJSON        bool
	gateSetPhase    string
	gateSetStatus   string
	gateSetVerdict  string
	gateSetAttempt  bool
	gateReentryFrom string
	gateReentryWhy  string
)

// gateCmd is the `sdlaic gate` command group: it inspects and updates the
// per-change phase-gate state persisted outside the repo at ~/.sdlaic/state/.
var gateCmd = &cobra.Command{
	Use:   "gate",
	Short: "Inspect and update phase-gate state",
	Long: `Reads and writes the global gate-state store for a change
(~/.sdlaic/state/<project_hash>/<change>/meta.json). Gate verdicts are kept
here, never inside the project repo.`,
}

var gateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show gate state for a change",
	RunE:  runGateStatus,
}

var gateSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Record a gate lifecycle transition",
	RunE:  runGateSet,
}

var gateReentryCmd = &cobra.Command{
	Use:   "reentry",
	Short: "Re-enter the pipeline at an earlier gate and supersede downstream gates",
	RunE:  runGateReentry,
}

func init() {
	rootCmd.AddCommand(gateCmd)
	gateCmd.AddCommand(gateStatusCmd, gateSetCmd, gateReentryCmd)

	gateStatusCmd.Flags().BoolVar(&gateJSON, "json", false, "Output as JSON")

	gateSetCmd.Flags().StringVar(&gateSetPhase, "phase", "", "Gate to update: proposal|spec:<capability>|design|tasks")
	gateSetCmd.Flags().StringVar(&gateSetStatus, "status", "", "New status: pending|grilling|grilled|reviewing|approved|failed|skipped")
	gateSetCmd.Flags().StringVar(&gateSetVerdict, "verdict", "", "Optional review verdict: APPROVE|REQUEST_CHANGES|REJECT|PENDING")
	gateSetCmd.Flags().BoolVar(&gateSetAttempt, "attempt", false, "Increment the gate's attempt counter")

	gateReentryCmd.Flags().StringVar(&gateReentryFrom, "from", "", "Gate to re-enter at: proposal|spec:<capability>|design|tasks")
	gateReentryCmd.Flags().StringVar(&gateReentryWhy, "reason", "", "Why the pipeline is being re-entered")
}

// gateStore resolves the workspace, config, and active change, and returns a
// gate-state store for it along with the loaded config and change name.
func gateStore() (*gatestate.Store, projectContext, string, error) {
	project, err := resolveProject()
	if err != nil {
		return nil, projectContext{}, "", err
	}

	changeName, err := project.resolveChange(changeFlag)
	if err != nil {
		return nil, projectContext{}, "", err
	}

	store := gatestate.NewWithHome(resolveHome(), project.Hash, changeName)

	// Enable per-capability spec gates when the change has a specs/ tree. Best
	// effort: if the path or directory listing fails, fall back to the legacy
	// single-spec pipeline rather than blocking gate inspection.
	if changePath, err := project.changePath(changeName); err == nil {
		if caps, err := state.ListCapabilities(changePath); err == nil {
			store.SetCapabilities(caps)
		}
	}

	return store, project, changeName, nil
}

func runGateStatus(cmd *cobra.Command, args []string) error {
	store, project, _, err := gateStore()
	if err != nil {
		return err
	}
	gf, _, err := store.LoadOrDefault(project.Workflow)
	if err != nil {
		return fmt.Errorf("loading gate state: %w", err)
	}

	for k, g := range gf.Gates {
		g.IsPassing = g.IsPassingFor(project.Workflow)
		gf.Gates[k] = g
	}

	if gateJSON {
		return printJSON(cmd, gf)
	}
	return printGateHuman(cmd, gf, project.Workflow)
}

func runGateSet(cmd *cobra.Command, args []string) error {
	store, project, changeName, err := gateStore()
	if err != nil {
		return err
	}

	validPhase := false
	for _, k := range store.GateKeys() {
		if k == gateSetPhase {
			validPhase = true
			break
		}
	}
	if !validPhase {
		return gatestate.UnknownGateErr(gateSetPhase, store.GateKeys())
	}

	status, err := domain.ParseGateStatus(gateSetStatus)
	if err != nil {
		return err
	}
	var verdict *domain.Verdict
	if gateSetVerdict != "" {
		v, err := domain.ParseVerdict(gateSetVerdict)
		if err != nil {
			return err
		}
		verdict = &v
	}

	// Validate that the artifact change directory exists before allowing state init
	changePath, err := project.changePath(changeName)
	if err != nil {
		return fmt.Errorf("resolving change path: %w", err)
	}
	if info, err := os.Stat(changePath); err != nil || !info.IsDir() {
		return fmt.Errorf("change %q not found: %w", changeName, domain.ErrChangeNotFound)
	}

	// Lazily initialize meta.json on first write (spec Open Q2).
	if _, err := store.Load(); errors.Is(err, domain.ErrGateStateNotFound) {
		if _, err := store.Init(project.Workflow); err != nil {
			return fmt.Errorf("initializing gate state: %w", err)
		}
	}

	gf, err := store.SetGate(gateSetPhase, status, verdict, gateSetAttempt)
	if err != nil {
		return err
	}

	g := gf.Gates[gateSetPhase]
	fmt.Fprintf(cmd.OutOrStdout(), "gate %q → %s (attempts %d)\n", gateSetPhase, g.Status, g.Attempts)
	return nil
}

func runGateReentry(cmd *cobra.Command, args []string) error {
	store, project, _, err := gateStore()
	if err != nil {
		return err
	}
	gf, err := store.ReEnter(gateReentryFrom, gateReentryWhy, project.Workflow)
	if errors.Is(err, domain.ErrGateStateNotFound) {
		return fmt.Errorf("no gate state yet for this change; run `sdlaic gate set` first: %w", err)
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "re-entered at %q; superseded: %v\n",
		gateReentryFrom, gf.History[len(gf.History)-1].SupersededGates)
	return nil
}

func printGateHuman(cmd *cobra.Command, gf domain.GatesFile, workflow domain.WorkflowLevel) error {
	fmt.Fprintf(cmd.OutOrStdout(), "Change:    %s\n", gf.Change)
	fmt.Fprintf(cmd.OutOrStdout(), "Workflow:  %s\n", workflow)
	fmt.Fprintf(cmd.OutOrStdout(), "Pipeline:  %s\n", gf.PipelineState)
	fmt.Fprintf(cmd.OutOrStdout(), "\nGates:\n")
	for _, key := range gatestate.SortedGateKeys(gf.Gates) {
		g := gf.Gates[key]
		marker := " "
		if g.IsPassingFor(workflow) {
			marker = "✓"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %-20s %-10s (attempts %d)\n", marker, key, g.Status, g.Attempts)
	}
	return nil
}

// resetGateFlags resets gate command flags to defaults (Cobra flags persist
// across test runs).
func resetGateFlags() {
	changeFlag = ""
	gateJSON = false
	gateSetPhase = ""
	gateSetStatus = ""
	gateSetVerdict = ""
	gateSetAttempt = false
	gateReentryFrom = ""
	gateReentryWhy = ""
}

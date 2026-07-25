package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/bienwithcode/SDLAIC/internal/gatestate"
	"github.com/bienwithcode/SDLAIC/internal/workspace"
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

	gateSetCmd.Flags().StringVar(&gateSetPhase, "phase", "", "Gate to update: proposal|spec|design|tasks")
	gateSetCmd.Flags().StringVar(&gateSetStatus, "status", "", "New status: pending|grilling|grilled|reviewing|approved|failed|skipped")
	gateSetCmd.Flags().StringVar(&gateSetVerdict, "verdict", "", "Optional review verdict: APPROVE|REQUEST_CHANGES|REJECT|PENDING")
	gateSetCmd.Flags().BoolVar(&gateSetAttempt, "attempt", false, "Increment the gate's attempt counter")

	gateReentryCmd.Flags().StringVar(&gateReentryFrom, "from", "", "Gate to re-enter at: proposal|spec|design|tasks")
	gateReentryCmd.Flags().StringVar(&gateReentryWhy, "reason", "", "Why the pipeline is being re-entered")
}

// gateStore resolves the workspace, config, and active change, and returns a
// gate-state store for it along with the loaded config and change name.
func gateStore() (*gatestate.Store, domain.LocalConfig, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, domain.LocalConfig{}, "", fmt.Errorf("getting current directory: %w", err)
	}
	root, err := workspace.FindWorkspace(cwd)
	if err != nil {
		return nil, domain.LocalConfig{}, "", fmt.Errorf("no SDLAIC workspace found (run 'sdlaic init' first): %w", err)
	}
	workspaceRoot = root

	cfg, err := loadLocalConfig()
	if err != nil {
		return nil, domain.LocalConfig{}, "", fmt.Errorf("loading config: %w", err)
	}
	changeName, err := resolveChangeName(changeFlag)
	if err != nil {
		return nil, domain.LocalConfig{}, "", err
	}

	homeDir, _ := os.UserHomeDir()
	store := gatestate.NewWithHome(homeDir, cfg.ProjectHash, changeName)
	return store, cfg, changeName, nil
}

func runGateStatus(cmd *cobra.Command, args []string) error {
	store, cfg, _, err := gateStore()
	if err != nil {
		return err
	}
	gf, _, err := store.LoadOrDefault(cfg.Workflow)
	if err != nil {
		return fmt.Errorf("loading gate state: %w", err)
	}
	if gateJSON {
		return printJSON(cmd, gf)
	}
	return printGateHuman(cmd, gf)
}

func runGateSet(cmd *cobra.Command, args []string) error {
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

	store, cfg, _, err := gateStore()
	if err != nil {
		return err
	}

	// Lazily initialize meta.json on first write (spec Open Q2).
	if _, err := store.Load(); errors.Is(err, domain.ErrGateStateNotFound) {
		if _, err := store.Init(cfg.Workflow); err != nil {
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
	store, _, _, err := gateStore()
	if err != nil {
		return err
	}
	gf, err := store.ReEnter(gateReentryFrom, gateReentryWhy)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "re-entered at %q; superseded: %v\n",
		gateReentryFrom, gf.History[len(gf.History)-1].SupersededGates)
	return nil
}

func printGateHuman(cmd *cobra.Command, gf domain.GatesFile) error {
	fmt.Fprintf(cmd.OutOrStdout(), "Change:    %s\n", gf.Change)
	fmt.Fprintf(cmd.OutOrStdout(), "Workflow:  %s\n", gf.Workflow)
	fmt.Fprintf(cmd.OutOrStdout(), "Pipeline:  %s\n", gf.PipelineState)
	fmt.Fprintf(cmd.OutOrStdout(), "\nGates:\n")
	for _, key := range gatestate.GateKeys() {
		g := gf.Gates[key]
		marker := " "
		if g.Status.IsPassing() {
			marker = "✓"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %-9s %-10s (attempts %d)\n", marker, key, g.Status, g.Attempts)
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

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"sdlaic/internal/domain"
	"sdlaic/internal/state"
	"sdlaic/internal/storage"
	"sdlaic/internal/workspace"
)

var statusJSON bool

// statusCmd represents the `sdlaic status` command.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current workflow state",
	Long: `Analyzes artifact files to determine the current phase of a change.
Outputs either human-readable text or JSON with full artifact status.`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
	// Don't re-register --change: it's already a persistent flag on rootCmd
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output as JSON")
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Find workspace
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	root, err := workspace.FindWorkspace(cwd)
	if err != nil {
		return fmt.Errorf("no SDLAIC workspace found (run 'sdlaic init' first): %w", err)
	}

	workspaceRoot = root

	// Load config
	cfg, err := loadLocalConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Resolve change name
	changeName, err := resolveChangeName(changeFlag)
	if err != nil {
		return err
	}

	// Resolve change path
	homeDir, _ := os.UserHomeDir()
	changePath, err := storage.ResolvePath(cfg.Storage, root, homeDir, changeName)
	if err != nil {
		return fmt.Errorf("resolving change path: %w", err)
	}

	// Analyze the change
	phase, err := state.AnalyzePhase(changePath)
	if err != nil {
		return fmt.Errorf("analyzing change %q: %w", changeName, err)
	}

	artifacts, err := state.AnalyzeArtifacts(changePath)
	if err != nil {
		return fmt.Errorf("analyzing artifacts: %w", err)
	}

	// Build status response
	status := domain.ChangeStatus{
		ActiveChange: changeName,
		StorageMode:  cfg.Storage,
		Workflow:     cfg.Workflow,
		CurrentPhase: phase,
		ChangePath:   changePath,
		Artifacts:    artifacts,
	}

	if statusJSON {
		return printJSON(cmd, status)
	}

	return printStatusHuman(cmd, status)
}

func printStatusHuman(cmd *cobra.Command, status domain.ChangeStatus) error {
	fmt.Fprintf(cmd.OutOrStdout(), "Change:     %s\n", status.ActiveChange)
	fmt.Fprintf(cmd.OutOrStdout(), "Phase:      %s\n", status.CurrentPhase)
	fmt.Fprintf(cmd.OutOrStdout(), "Storage:    %s\n", status.StorageMode)
	fmt.Fprintf(cmd.OutOrStdout(), "Workflow:   %s\n", status.Workflow)
	fmt.Fprintf(cmd.OutOrStdout(), "Path:       %s\n", status.ChangePath)
	fmt.Fprintf(cmd.OutOrStdout(), "\nArtifacts:\n")

	for _, at := range domain.OrderedArtifactTypes() {
		name := string(at)
		artifact, ok := status.Artifacts[name]
		if !ok {
			continue
		}
		marker := " "
		if artifact.Populated {
			marker = "✓"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s\n", marker, at.FileName())
	}

	return nil
}

// resetStatusFlags resets status command flags to defaults.
func resetStatusFlags() {
	changeFlag = ""
	statusJSON = false
}

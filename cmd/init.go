package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/bienwithcode/SDLAIC/internal/storage"
	"github.com/bienwithcode/SDLAIC/internal/workspace"
)

var (
	initWorkflow   string
	initChangesDir string
)

// initCmd represents the `sdlaic init` command.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Register this project with SDLAIC",
	Long: `Registers the current project in ~/.sdlaic/config.json and prepares the
directory that holds change artifacts.

By default artifacts live in <project>/.sdlaic/changes. Pass --changes-dir to
keep them somewhere else — an external path leaves the project untouched.

Re-running init on a registered project updates its entry.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&initWorkflow, "workflow", "strict", "Workflow level: strict, light, free")
	initCmd.Flags().StringVar(&initChangesDir, "changes-dir", "", "Directory holding change artifacts (default <project>/.sdlaic/changes)")
}

func runInit(cmd *cobra.Command, args []string) error {
	workflowLevel, err := domain.ParseWorkflowLevel(initWorkflow)
	if err != nil {
		return fmt.Errorf("invalid --workflow flag: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	hash, err := workspace.ProjectHash(cwd)
	if err != nil {
		return fmt.Errorf("computing project hash: %w", err)
	}

	// An unset --changes-dir keeps artifacts inside the project; anything else is
	// normalised to an absolute path now, so it cannot change meaning later.
	changesDir := storage.DefaultChangesDir(cwd)
	if initChangesDir != "" {
		changesDir, err = storage.NormalizeChangesDir(initChangesDir, cwd, resolveHome())
		if err != nil {
			return fmt.Errorf("invalid --changes-dir flag: %w", err)
		}
	}

	if err := registerProjectEntry(cwd, changesDir, workflowLevel); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Registered SDLAIC project %s\n", cwd)
	fmt.Fprintf(cmd.OutOrStdout(), "  Changes:  %s\n", changesDir)
	fmt.Fprintf(cmd.OutOrStdout(), "  Workflow: %s\n", workflowLevel)
	fmt.Fprintf(cmd.OutOrStdout(), "  Hash:     %s\n", hash)

	return nil
}

// resetInitFlags resets the init command flags to their default values.
// Required for testing since Cobra flag variables persist between executions.
func resetInitFlags() {
	initWorkflow = "strict"
	initChangesDir = ""
	homeFlag = ""
}

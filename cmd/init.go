package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/config"
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
	isDefaultLayout := true
	if initChangesDir != "" {
		changesDir, err = storage.NormalizeChangesDir(initChangesDir, cwd, resolveHome())
		if err != nil {
			return fmt.Errorf("invalid --changes-dir flag: %w", err)
		}
		isDefaultLayout = changesDir == storage.DefaultChangesDir(cwd)
	}

	cfgPath := globalConfigPath()
	if err := ensureChangesDirUnclaimed(cfgPath, hash, changesDir); err != nil {
		return err
	}

	if err := os.MkdirAll(changesDir, 0755); err != nil {
		return fmt.Errorf("creating changes directory: %w", err)
	}

	if err := config.UpdateProject(cfgPath, hash, func(e *domain.ProjectEntry) {
		e.Path = cwd
		e.ChangesDir = changesDir
		e.Workflow = workflowLevel
	}); err != nil {
		return fmt.Errorf("registering project: %w", err)
	}

	// TEMPORARY: commands that have not been migrated yet still read .sdlaicrc.
	// It is only written for the in-project default, so an external changes dir
	// keeps its zero-footprint promise. Removed in T17.
	if isDefaultLayout {
		local := domain.NewLocalConfig(domain.StorageModeLocal, workflowLevel, hash)
		if err := config.SaveLocal(local, filepath.Join(cwd, ".sdlaicrc")); err != nil {
			return fmt.Errorf("writing legacy local config: %w", err)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Registered SDLAIC project %s\n", cwd)
	fmt.Fprintf(cmd.OutOrStdout(), "  Changes:  %s\n", changesDir)
	fmt.Fprintf(cmd.OutOrStdout(), "  Workflow: %s\n", workflowLevel)
	fmt.Fprintf(cmd.OutOrStdout(), "  Hash:     %s\n", hash)

	return nil
}

// ensureChangesDirUnclaimed rejects a directory already registered to a
// different project. One changes directory belongs to exactly one project:
// sharing one would make `list` mix projects, let `archive` overwrite another
// project's tarball, and leave the same change carrying a different gate state
// depending on which project you stand in.
func ensureChangesDirUnclaimed(cfgPath string, hash string, changesDir string) error {
	cfg, err := config.LoadOrCreateGlobal(cfgPath)
	if err != nil {
		return fmt.Errorf("loading global config: %w", err)
	}
	for otherHash, entry := range cfg.Projects {
		if otherHash == hash || entry.ChangesDir != changesDir {
			continue
		}
		return fmt.Errorf("changes directory %s is already used by project %s", changesDir, entry.Path)
	}
	return nil
}

// resetInitFlags resets the init command flags to their default values.
// Required for testing since Cobra flag variables persist between executions.
func resetInitFlags() {
	initWorkflow = "strict"
	initChangesDir = ""
	homeFlag = ""
}

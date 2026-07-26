package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/config"
	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/bienwithcode/SDLAIC/internal/storage"
)

// configCmd represents the `sdlaic config` command.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage SDLAIC configuration",
	Long:  `View and modify this project's entry in ~/.sdlaic/config.json.`,
}

// configSetCmd represents `sdlaic config set`.
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration key to a new value.

Supported keys:
  changes-dir — directory holding change artifacts (stored as an absolute path)
  workflow    — strict, light, free`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

// configListCmd represents `sdlaic config list`.
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List current configuration",
	Long:  `Display the configuration for the current project.`,
	RunE:  runConfigList,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key, value := args[0], args[1]

	project, err := resolveProject()
	if err != nil {
		return err
	}

	switch key {
	case "changes-dir":
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}
		changesDir, err := storage.NormalizeChangesDir(value, cwd, resolveHome())
		if err != nil {
			return fmt.Errorf("invalid changes-dir value: %w", err)
		}
		if err := ensureChangesDirUnclaimed(globalConfigPath(), project.Hash, changesDir); err != nil {
			return err
		}
		if err := os.MkdirAll(changesDir, 0755); err != nil {
			return fmt.Errorf("creating changes directory: %w", err)
		}
		if err := config.UpdateProject(globalConfigPath(), project.Hash, func(e *domain.ProjectEntry) {
			e.Path = project.Root
			e.ChangesDir = changesDir
		}); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		value = changesDir

	case "workflow":
		level, err := domain.ParseWorkflowLevel(value)
		if err != nil {
			return fmt.Errorf("invalid workflow value %q: %w", value, err)
		}
		if err := config.UpdateProject(globalConfigPath(), project.Hash, func(e *domain.ProjectEntry) {
			e.Path = project.Root
			e.Workflow = level
		}); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

	default:
		return fmt.Errorf("unknown configuration key %q; valid keys: changes-dir, workflow", key)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Set %s = %s\n", key, value)
	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "path          = %s", project.Root)
	if _, err := os.Stat(project.Root); err != nil {
		fmt.Fprint(out, "  (stale: directory no longer exists)")
	}
	fmt.Fprintln(out)

	changesDir := project.ChangesDir
	if changesDir == "" {
		changesDir = "(not set — run 'sdlaic init --changes-dir <path>')"
	}
	fmt.Fprintf(out, "changes_dir   = %s\n", changesDir)
	fmt.Fprintf(out, "workflow      = %s\n", project.Workflow)
	if project.ActiveChange != "" {
		fmt.Fprintf(out, "active_change = %s\n", project.ActiveChange)
	}
	fmt.Fprintf(out, "project_hash  = %s\n", project.Hash)

	return nil
}

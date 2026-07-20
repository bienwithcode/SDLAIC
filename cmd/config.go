package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/bienwithcode/SDLAIC/internal/storage"
	"github.com/bienwithcode/SDLAIC/internal/workspace"
)

// configCmd represents the `sdlaic config` command.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage SDLAIC configuration",
	Long:  `View and modify workspace configuration settings.`,
}

// configSetCmd represents `sdlaic config set`.
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration key to a new value.

Supported keys:
  storage  — local, ignored, global
  workflow — strict, light, free`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

// configListCmd represents `sdlaic config list`.
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List current configuration",
	Long:  `Display all configuration key-value pairs for the current workspace.`,
	RunE:  runConfigList,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

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
	cfg, err := loadLocalConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	switch key {
	case "storage":
		mode, err := domain.ParseStorageMode(value)
		if err != nil {
			return fmt.Errorf("invalid storage value %q: %w", value, err)
		}
		oldMode := cfg.Storage
		cfg.Storage = mode

		// Handle gitignore side effects
		if oldMode != domain.StorageModeIgnored && mode == domain.StorageModeIgnored {
			if err := storage.AppendToGitignore(root, ".sdlaic/changes/"); err != nil {
				return fmt.Errorf("updating .gitignore: %w", err)
			}
		}

	case "workflow":
		level, err := domain.ParseWorkflowLevel(value)
		if err != nil {
			return fmt.Errorf("invalid workflow value %q: %w", value, err)
		}
		cfg.Workflow = level

	default:
		return fmt.Errorf("unknown configuration key %q; valid keys: storage, workflow", key)
	}

	if err := saveLocalConfig(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Set %s = %s\n", key, value)
	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	root, err := workspace.FindWorkspace(cwd)
	if err != nil {
		return fmt.Errorf("no SDLAIC workspace found (run 'sdlaic init' first): %w", err)
	}

	workspaceRoot = root
	cfg, err := loadLocalConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "storage  = %s\n", cfg.Storage)
	fmt.Fprintf(cmd.OutOrStdout(), "workflow = %s\n", cfg.Workflow)
	if cfg.ActiveChange != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "active_change = %s\n", cfg.ActiveChange)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "project_hash  = %s\n", cfg.ProjectHash)

	return nil
}

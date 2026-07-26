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
	initStorage  string
	initWorkflow string
)

// initCmd represents the `sdlaic init` command.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new SDLAIC workspace",
	Long: `Creates a .sdlaicrc config file and the directory structure
for managing change artifacts in the current project.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&initStorage, "storage", "local", "Storage mode: local, ignored, global")
	initCmd.Flags().StringVar(&initWorkflow, "workflow", "strict", "Workflow level: strict, light, free")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Reset flags to defaults for this execution
	storageMode, err := domain.ParseStorageMode(initStorage)
	if err != nil {
		return fmt.Errorf("invalid --storage flag: %w", err)
	}

	workflowLevel, err := domain.ParseWorkflowLevel(initWorkflow)
	if err != nil {
		return fmt.Errorf("invalid --workflow flag: %w", err)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	// Compute project hash
	hash, err := workspace.ProjectHash(cwd)
	if err != nil {
		return fmt.Errorf("computing project hash: %w", err)
	}

	// Initialize workspace
	homeDir := resolveHome()

	cfg, err := workspace.InitWorkspaceWithHome(cwd, homeDir, storageMode, workflowLevel, hash)
	if err != nil {
		return fmt.Errorf("initializing workspace: %w", err)
	}

	// Handle gitignore for ignored mode
	if storageMode == domain.StorageModeIgnored {
		changesRelPath := ".sdlaic/changes/"
		if err := storage.AppendToGitignore(cwd, changesRelPath); err != nil {
			return fmt.Errorf("updating .gitignore: %w", err)
		}
	}

	// Also register in global config
	registerProject(globalConfigPath(), hash, cwd, storageMode)

	fmt.Fprintf(cmd.OutOrStdout(), "Initialized SDLAIC workspace in %s\n", cwd)
	fmt.Fprintf(cmd.OutOrStdout(), "  Storage:  %s\n", cfg.Storage)
	fmt.Fprintf(cmd.OutOrStdout(), "  Workflow: %s\n", cfg.Workflow)
	fmt.Fprintf(cmd.OutOrStdout(), "  Hash:     %s\n", cfg.ProjectHash)

	return nil
}

// registerProject adds the project to the global config file.
func registerProject(globalCfgPath string, hash string, path string, storageMode domain.StorageMode) {
	cfg, err := loadOrCreateGlobalConfig(globalCfgPath)
	if err != nil {
		return // Non-fatal: global config is optional
	}

	cfg.Projects[hash] = domain.ProjectEntry{
		Path:    path,
		Storage: storageMode,
	}

	_ = saveGlobalConfig(globalCfgPath, cfg)
}

func loadOrCreateGlobalConfig(path string) (domain.GlobalConfig, error) {
	cfg, err := loadGlobalConfig(path)
	if err != nil {
		return domain.NewGlobalConfig(), nil
	}
	return cfg, nil
}

// resetInitFlags resets the init command flags to their default values.
// Required for testing since Cobra flag variables persist between executions.
func resetInitFlags() {
	initStorage = "local"
	initWorkflow = "strict"
	homeFlag = ""
}

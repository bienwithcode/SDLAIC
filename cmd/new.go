package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/bienwithcode/SDLAIC/internal/storage"
	"github.com/bienwithcode/SDLAIC/internal/templates"
	"github.com/bienwithcode/SDLAIC/internal/workspace"
)

// newCmd represents the `sdlaic new` command group.
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create new resources",
}

// newChangeCmd represents `sdlaic new change`.
var newChangeCmd = &cobra.Command{
	Use:   "change <name>",
	Short: "Create a new change artifact directory",
	Long: `Creates a new change directory with a context.md template
and sets it as the active change.`,
	Args: cobra.ExactArgs(1),
	RunE: runNewChange,
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.AddCommand(newChangeCmd)
}

func runNewChange(cmd *cobra.Command, args []string) error {
	changeName := args[0]

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

	// Load config to determine storage mode
	cfg, err := loadLocalConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Resolve change path
	homeDir, _ := os.UserHomeDir()
	changePath, err := storage.ResolvePath(cfg.Storage, root, homeDir, changeName)
	if err != nil {
		return fmt.Errorf("resolving change path: %w", err)
	}

	// Check if change already exists
	if info, err := os.Stat(changePath); err == nil && info.IsDir() {
		return fmt.Errorf("change %q already exists: %w", changeName, domain.ErrChangeAlreadyExists)
	}

	// Create change directory
	if err := os.MkdirAll(changePath, 0755); err != nil {
		return fmt.Errorf("creating change directory: %w", err)
	}

	// Write context.md template
	templateContent, err := templates.GetTemplate(domain.ArtifactContext)
	if err != nil {
		return fmt.Errorf("loading context template: %w", err)
	}

	contextPath := filepath.Join(changePath, domain.ArtifactContext.FileName())
	if err := os.WriteFile(contextPath, []byte(templateContent), 0644); err != nil {
		return fmt.Errorf("writing context template: %w", err)
	}

	// Set as active change
	cfg.ActiveChange = changeName
	if err := saveLocalConfig(cfg); err != nil {
		return fmt.Errorf("setting active change: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created change %q at %s\n", changeName, changePath)
	fmt.Fprintf(cmd.OutOrStdout(), "Set as active change.\n")

	return nil
}

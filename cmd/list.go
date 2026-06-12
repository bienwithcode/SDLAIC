package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"sdlaic/internal/storage"
	"sdlaic/internal/workspace"
)

var (
	listAll bool
	listJSON bool
)

// listCmd represents the `sdlaic list` command.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all changes",
	Long:  `Lists all changes in the current workspace. Use --all to include archived changes.`,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVar(&listAll, "all", false, "Include archived changes")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
}

func runList(cmd *cobra.Command, args []string) error {
	// Find workspace
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	root, err := workspace.FindWorkspace(cwd)
	if err != nil {
		return fmt.Errorf("no SDLAIC workspace found: %w", err)
	}

	workspaceRoot = root

	cfg, err := loadLocalConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	homeDir, _ := os.UserHomeDir()
	basePath, err := storage.ChangesBasePath(cfg.Storage, root, homeDir)
	if err != nil {
		return fmt.Errorf("resolving changes path: %w", err)
	}

	// List active changes
	changes, err := storage.ListChanges(basePath)
	if err != nil {
		return fmt.Errorf("listing changes: %w", err)
	}

	// Filter out .archive from regular listing
	var active []string
	for _, c := range changes {
		if c != ".archive" {
			active = append(active, c)
		}
	}

	result := map[string]interface{}{
		"changes": active,
	}

	// Include archived if --all
	if listAll {
		archiveDir := filepath.Join(basePath, ".archive")
		archived, err := listArchived(archiveDir)
		if err == nil {
			result["archived"] = archived
		}
	}

	if listJSON {
		return printJSON(cmd, result)
	}

	// Human output
	fmt.Fprintf(cmd.OutOrStdout(), "Changes:\n")
	for _, c := range active {
		marker := " "
		if c == cfg.ActiveChange {
			marker = "*"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %s %s\n", marker, c)
	}

	if listAll {
		archiveDir := filepath.Join(basePath, ".archive")
		archived, err := listArchived(archiveDir)
		if err == nil && len(archived) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "\nArchived:\n")
			for _, a := range archived {
				fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", a)
			}
		}
	}

	return nil
}

// listArchived returns names of archived changes (without .tar.gz extension).
func listArchived(archiveDir string) ([]string, error) {
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		name := e.Name()
		// Strip .tar.gz extension
		if len(name) > 7 && name[len(name)-7:] == ".tar.gz" {
			names = append(names, name[:len(name)-7])
		}
	}
	return names, nil
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/storage"
)

var (
	listAll  bool
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
	project, err := resolveProject()
	if err != nil {
		return err
	}

	basePath, err := project.changesDir()
	if err != nil {
		return err
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

	// Human output — name the directory, since it is per-project configuration
	// and not necessarily inside the project.
	fmt.Fprintf(cmd.OutOrStdout(), "Changes in %s:\n", basePath)
	for _, c := range active {
		marker := " "
		if c == project.ActiveChange {
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

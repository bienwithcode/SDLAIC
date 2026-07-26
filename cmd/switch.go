package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/bienwithcode/SDLAIC/internal/storage"
)

// switchCmd represents the `sdlaic switch` command.
var switchCmd = &cobra.Command{
	Use:   "switch [<change-name>]",
	Short: "Switch the active change",
	Long: `Sets the active change context. If no change name is provided,
lists all available changes.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSwitch,
}

func init() {
	rootCmd.AddCommand(switchCmd)
}

func runSwitch(cmd *cobra.Command, args []string) error {
	// Find workspace
	project, err := resolveProject()
	if err != nil {
		return err
	}

	basePath, err := project.changesDir()
	if err != nil {
		return err
	}

	// If no argument, list available changes
	if len(args) == 0 {
		return listChanges(cmd, basePath)
	}

	changeName := args[0]

	// Verify change exists
	changePath, err := project.changePath(changeName)
	if err != nil {
		return fmt.Errorf("resolving change path: %w", err)
	}

	if info, err := os.Stat(changePath); err != nil || !info.IsDir() {
		return fmt.Errorf("change %q not found: %w", changeName, domain.ErrChangeNotFound)
	}

	// Set as active
	if err := project.setActiveChange(changeName); err != nil {
		return fmt.Errorf("saving active change: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Switched to change %q\n", changeName)
	return nil
}

func listChanges(cmd *cobra.Command, basePath string) error {
	changes, err := storage.ListChanges(basePath)
	if err != nil {
		return fmt.Errorf("listing changes: %w", err)
	}

	// Filter out .archive directory
	var filtered []string
	for _, c := range changes {
		if c != ".archive" {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) == 0 {
		return fmt.Errorf("no changes available; create one with 'sdlaic new change <name>'")
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Available changes:\n")
	for _, name := range filtered {
		fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", name)
	}

	return nil
}

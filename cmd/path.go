package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// pathCmd represents the `sdlaic path` command group.
//
// Agents and scripts must not assume artifacts live in .sdlaic/changes — the
// location is per-project configuration. This is how they ask.
var pathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print resolved SDLAIC paths for this project",
	Long: `Prints absolute paths for the current project, one per line and
undecorated, so they can be used directly in scripts and skills:

  sdlaic path changes            # the directory holding all changes
  sdlaic path change             # the active change's directory
  sdlaic path change -c SDL-1    # a specific change's directory`,
}

var pathChangesCmd = &cobra.Command{
	Use:   "changes",
	Short: "Print the absolute changes directory",
	RunE:  runPathChanges,
}

var pathChangeCmd = &cobra.Command{
	Use:   "change",
	Short: "Print the absolute directory of one change",
	RunE:  runPathChange,
}

func init() {
	rootCmd.AddCommand(pathCmd)
	pathCmd.AddCommand(pathChangesCmd)
	pathCmd.AddCommand(pathChangeCmd)
}

func runPathChanges(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	changesDir, err := project.changesDir()
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), changesDir)
	return nil
}

func runPathChange(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	changeName, err := project.resolveChange(changeFlag)
	if err != nil {
		return err
	}

	changePath, err := project.changePath(changeName)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), changePath)
	return nil
}

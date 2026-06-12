package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionShort bool

// versionCmd represents the `sdlaic version` command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the SDLAIC version",
	Long: `Prints the SDLAIC CLI version. Use --short to output only the
semver version string (useful for scripting).`,
	RunE: runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVar(&versionShort, "short", false, "Print only the version number")
}

func runVersion(cmd *cobra.Command, args []string) error {
	if versionShort {
		fmt.Fprintln(cmd.OutOrStdout(), cliVersion)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "sdlaic %s\n", cliVersion)
	}
	return nil
}

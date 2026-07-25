package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/templates"
)

// instructionsCmd represents the `sdlaic instructions` command.
var instructionsCmd = &cobra.Command{
	Use:   "instructions <type>",
	Short: "Output template instructions for an artifact type",
	Long: `Reads the embedded template for the specified artifact type and
prints it to stdout as raw markdown.

Available types: context, proposal, spec, design, tasks`,
	Args: cobra.ExactArgs(1),
	RunE: runInstructions,
}

func init() {
	rootCmd.AddCommand(instructionsCmd)
}

func runInstructions(cmd *cobra.Command, args []string) error {
	artifactType := args[0]

	content, err := templates.GetTemplateByName(artifactType)
	if err != nil {
		return fmt.Errorf("invalid artifact type %q: %w", artifactType, err)
	}

	fmt.Fprint(cmd.OutOrStdout(), content)
	return nil
}

package cmd

import (
	"github.com/spf13/cobra"
)

// completionCmd represents the `sdlaic completion` command.
var completionCmd = &cobra.Command{
	Use:   "completion [shell]",
	Short: "Generate shell completion script",
	Long: `Generate a shell completion script for SDLAIC.

Supported shells: bash, zsh, fish, powershell

Example:
  sdlaic completion bash > /etc/bash_completion.d/sdlaic
  sdlaic completion zsh > "${fpath[1]}/_sdlaic"
  sdlaic completion fish > ~/.config/fish/completions/sdlaic.fish`,
	Args:               cobra.ExactArgs(1),
	ValidArgs:          []string{"bash", "zsh", "fish", "powershell"},
	DisableFlagParsing: true,
	RunE:               runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func runCompletion(cmd *cobra.Command, args []string) error {
	shell := args[0]
	switch shell {
	case "bash":
		return rootCmd.GenBashCompletion(cmd.OutOrStdout())
	case "zsh":
		return rootCmd.GenZshCompletion(cmd.OutOrStdout())
	case "fish":
		return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
	case "powershell":
		return rootCmd.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
	default:
		return cmd.Help()
	}
}

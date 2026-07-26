package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/bienwithcode/SDLAIC/internal/storage"
	"github.com/bienwithcode/SDLAIC/internal/workspace"
)

var (
	openNoSpawn     bool
	openPrint       bool
	openMarketplace string
	openChangesDir  string
	openWorkflow    string
)

// openCmd represents the `sdlaic open` command group.
var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open and configure AI plugin integrations",
	Long:  `Installs target AI plugins and spawns/launches their respective shells.`,
}

// openClaudeCmd represents `sdlaic open claude`.
var openClaudeCmd = &cobra.Command{
	Use:   "claude",
	Short: "Install SDLAIC plugin in Claude Code and spawn Claude shell",
	RunE:  runOpenClaude,
}

func init() {
	rootCmd.AddCommand(openCmd)
	openCmd.AddCommand(openClaudeCmd)

	openClaudeCmd.Flags().BoolVar(&openNoSpawn, "no-spawn", false, "Install only, do not spawn Claude shell")
	openClaudeCmd.Flags().BoolVar(&openPrint, "print", false, "Print commands to be run, do not execute them")
	openClaudeCmd.Flags().StringVar(&openMarketplace, "marketplace", "bienwithcode/SDLAIC", "Override default marketplace repository owner/repo")
	openClaudeCmd.Flags().StringVar(&openChangesDir, "changes-dir", "", "Directory holding change artifacts, for first-time setup")
	openClaudeCmd.Flags().StringVar(&openWorkflow, "workflow", "", "Override workflow level for auto-init (strict, light, free)")

	// Codex stub
	openCmd.AddCommand(&cobra.Command{
		Use:   "codex",
		Short: "Install SDLAIC plugin in Codex",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "coming in a later release")
		},
	})
}

func runOpenClaude(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	if err := ensureProjectConfigured(cmd, cwd, os.Stdin, workspace.IsTerminal()); err != nil {
		return err
	}

	// Prepare commands
	parts := strings.Split(openMarketplace, "/")
	mName := "bienwithcode"
	if len(parts) > 0 {
		mName = parts[0]
	}

	marketplaceAddCmd := fmt.Sprintf("claude plugin marketplace add %s", openMarketplace)
	pluginInstallCmd := fmt.Sprintf("claude plugin install sdlaic@%s", mName)

	if openPrint {
		fmt.Fprintln(cmd.OutOrStdout(), marketplaceAddCmd)
		fmt.Fprintln(cmd.OutOrStdout(), pluginInstallCmd)
		if !openNoSpawn {
			fmt.Fprintln(cmd.OutOrStdout(), "claude")
		}
		return nil
	}

	// Verify claude is in PATH
	_, err = exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude CLI is not installed or not in PATH. Please install it first")
	}

	// Execute commands
	fmt.Fprintf(cmd.OutOrStdout(), "Adding marketplace: %s...\n", openMarketplace)
	if err := runShellCmd(partsOf(marketplaceAddCmd)...); err != nil {
		return fmt.Errorf("failed to add marketplace: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Installing plugin sdlaic@%s...\n", mName)
	if err := runShellCmd(partsOf(pluginInstallCmd)...); err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	if !openNoSpawn {
		fmt.Fprintln(cmd.OutOrStdout(), "Spawning Claude Code session...")
		c := exec.Command("claude")
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("error running Claude session: %w", err)
		}
	}

	return nil
}

// ensureProjectConfigured makes sure the current directory is a registered
// project with a changes directory and a workflow, asking for anything missing.
//
// Values already stored are never re-asked, and flags win over prompts. In a
// non-interactive session the documented defaults are used rather than blocking
// on a prompt nobody can answer.
func ensureProjectConfigured(cmd *cobra.Command, cwd string, in io.Reader, interactive bool) error {
	out := cmd.OutOrStdout()

	project, err := resolveProject()
	switch {
	case err == nil && project.ChangesDir != "" && project.Workflow != "":
		return nil
	case err != nil && !errors.Is(err, domain.ErrWorkspaceNotFound):
		return err
	}

	// A partially configured project — the shape a v1 config upgrade leaves —
	// must be completed in place. Registering against cwd would add a second
	// entry for a subdirectory, and longest-prefix lookup would then prefer it
	// over the real project.
	root := cwd
	if project.Root != "" {
		root = project.Root
	}

	changesDir := project.ChangesDir
	workflowLevel := project.Workflow

	if openChangesDir != "" {
		changesDir, err = storage.NormalizeChangesDir(openChangesDir, root, resolveHome())
		if err != nil {
			return fmt.Errorf("invalid --changes-dir flag: %w", err)
		}
	}
	if openWorkflow != "" {
		workflowLevel, err = domain.ParseWorkflowLevel(openWorkflow)
		if err != nil {
			return fmt.Errorf("invalid --workflow flag: %w", err)
		}
	}

	needsChangesDir := changesDir == ""
	needsWorkflow := workflowLevel == ""
	if (needsChangesDir || needsWorkflow) && interactive {
		fmt.Fprintln(out, "This project is not set up yet. Let's fix that.")
	}

	// One prompter for both questions: a per-question scanner would read ahead
	// and eat the next answer.
	prompter := workspace.NewPrompter(in, out)

	if needsChangesDir {
		changesDir = storage.DefaultChangesDir(root)
		if interactive {
			choice, err := prompter.Choice(
				"Where should change artifacts live?", []string{"default", "custom"}, "default")
			if err != nil {
				return fmt.Errorf("reading changes directory choice: %w", err)
			}
			if choice == "custom" {
				entered, err := prompter.Line("Path", storage.DefaultChangesDir(root))
				if err != nil {
					return fmt.Errorf("reading changes directory: %w", err)
				}
				changesDir, err = storage.NormalizeChangesDir(entered, root, resolveHome())
				if err != nil {
					return err
				}
			}
		}
	}

	if needsWorkflow {
		workflowLevel = domain.WorkflowStrict
		if interactive {
			choice, err := prompter.Choice(
				"Select Workflow Level", []string{"strict", "light", "free"}, "strict")
			if err != nil {
				return fmt.Errorf("reading workflow level choice: %w", err)
			}
			workflowLevel, _ = domain.ParseWorkflowLevel(choice)
		}
	}

	if err := registerProjectEntry(root, changesDir, workflowLevel); err != nil {
		return err
	}

	fmt.Fprintf(out, "Configured SDLAIC project %s\n", root)
	fmt.Fprintf(out, "  Changes:  %s\n", changesDir)
	fmt.Fprintf(out, "  Workflow: %s\n", workflowLevel)
	return nil
}

func runShellCmd(args ...string) error {
	c := exec.Command(args[0], args[1:]...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func partsOf(cmd string) []string {
	return strings.Fields(cmd)
}

func resetOpenFlags() {
	openNoSpawn = false
	openPrint = false
	openMarketplace = "bienwithcode/SDLAIC"
	openChangesDir = ""
	openWorkflow = ""
}

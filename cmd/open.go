package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	openStorage     string
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
	openClaudeCmd.Flags().StringVar(&openStorage, "storage", "", "Override storage mode for auto-init (local, ignored, global)")
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

	// Check if workspace exists
	hasWorkspace := true
	_, err = workspace.FindWorkspace(cwd)
	if err != nil {
		hasWorkspace = false
	}

	if !hasWorkspace {
		// Needs auto-init
		var storageMode domain.StorageMode
		var workflowLevel domain.WorkflowLevel

		isInteractive := workspace.IsTerminal()

		if isInteractive && openStorage == "" && openWorkflow == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "Workspace is not initialized. Let's set it up!")

			// Prompt for Storage Mode
			sChoice, err := workspace.AskChoice(os.Stdin, cmd.OutOrStdout(), "Select Storage Mode", []string{"local", "ignored", "global"}, "local")
			if err != nil {
				return fmt.Errorf("reading storage mode choice: %w", err)
			}
			storageMode, _ = domain.ParseStorageMode(sChoice)

			// Prompt for Workflow Level
			wChoice, err := workspace.AskChoice(os.Stdin, cmd.OutOrStdout(), "Select Workflow Level", []string{"strict", "light", "free"}, "strict")
			if err != nil {
				return fmt.Errorf("reading workflow level choice: %w", err)
			}
			workflowLevel, _ = domain.ParseWorkflowLevel(wChoice)
		} else {
			// Non-interactive or flags provided
			sVal := "local"
			if openStorage != "" {
				sVal = openStorage
			}
			storageMode, err = domain.ParseStorageMode(sVal)
			if err != nil {
				return fmt.Errorf("invalid storage override: %w", err)
			}

			wVal := "strict"
			if openWorkflow != "" {
				wVal = openWorkflow
			}
			workflowLevel, err = domain.ParseWorkflowLevel(wVal)
			if err != nil {
				return fmt.Errorf("invalid workflow override: %w", err)
			}
		}

		// Run initialization logic
		hash, err := workspace.ProjectHash(cwd)
		if err != nil {
			return fmt.Errorf("computing project hash: %w", err)
		}

		homeDir, _ := os.UserHomeDir()
		_, err = workspace.InitWorkspaceWithHome(cwd, homeDir, storageMode, workflowLevel, hash)
		if err != nil {
			return fmt.Errorf("auto-initializing workspace: %w", err)
		}

		// Append to gitignore if ignored storage
		if storageMode == domain.StorageModeIgnored {
			changesRelPath := ".sdlaic/changes/"
			_ = storage.AppendToGitignore(cwd, changesRelPath)
		}

		// Register in global config
		globalCfgPath := filepath.Join(homeDir, ".sdlaic", "config.json")
		registerProject(globalCfgPath, hash, cwd, storageMode)

		fmt.Fprintf(cmd.OutOrStdout(), "Auto-initialized SDLAIC workspace in %s\n", cwd)
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

// registerProject adds the project to the global config file.
// TEMPORARY: moved here from init.go, which now registers via
// config.UpdateProject. Removed when open is migrated in T16.
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
	openStorage = ""
	openWorkflow = ""
}

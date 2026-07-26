// Package cmd contains all Cobra command definitions for the SDLAIC CLI.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/config"
	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/bienwithcode/SDLAIC/internal/workspace"
)

// workspaceRoot is the discovered workspace root directory.
// Set by PersistentPreRun when a command is executed.
var workspaceRoot string

// rootCmd is the base command for the SDLAIC CLI.
var rootCmd = &cobra.Command{
	Use:   "sdlaic",
	Short: "SDLAIC — Spec-Driven LLM-Assisted Implementation Cycle",
	Long: `SDLAIC CLI manages change artifacts for spec-driven development.
It provides workspace initialization, change lifecycle management,
artifact templating, validation, and status reporting.`,
	SilenceUsage: true,
	Version:      cliVersion,
}

// cliVersion holds the version injected at build time.
var cliVersion = "dev"

// SetVersion allows main.go to inject the build-time version.
func SetVersion(v string) {
	cliVersion = v
	rootCmd.Version = v
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&changeFlag, "change", "c", "", "Target change name (uses active change if omitted)")
}

// changeFlag is shared across commands that accept --change/-c.
var changeFlag string

// resolveChangeName returns the change name from the flag or the active change
// in the workspace config.
func resolveChangeName(flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}

	if workspaceRoot != "" {
		cfgPath := workspaceRoot + "/.sdlaicrc"
		cfg, err := config.LoadLocal(cfgPath)
		if err != nil {
			return "", fmt.Errorf("loading config: %w", err)
		}
		if cfg.ActiveChange != "" {
			return cfg.ActiveChange, nil
		}
	}

	return "", domain.ErrNoActiveChange
}

// discoverWorkspace sets workspaceRoot by walking up from cwd.
func discoverWorkspace() {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	root, err := workspace.FindWorkspace(cwd)
	if err != nil {
		workspaceRoot = ""
		return
	}
	workspaceRoot = root
}

// loadLocalConfig loads the .sdlaicrc from the workspace root.
func loadLocalConfig() (domain.LocalConfig, error) {
	if workspaceRoot == "" {
		return domain.LocalConfig{}, domain.ErrWorkspaceNotFound
	}
	return config.LoadLocal(workspaceRoot + "/.sdlaicrc")
}

// saveLocalConfig saves the .sdlaicrc to the workspace root.
func saveLocalConfig(cfg domain.LocalConfig) error {
	if workspaceRoot == "" {
		return domain.ErrWorkspaceNotFound
	}
	return config.SaveLocal(cfg, workspaceRoot+"/.sdlaicrc")
}

// loadGlobalConfig loads a global config from the given path.
func loadGlobalConfig(path string) (domain.GlobalConfig, error) {
	return config.LoadGlobal(path)
}

// saveGlobalConfig saves a global config to the given path.
func saveGlobalConfig(path string, cfg domain.GlobalConfig) error {
	return config.SaveGlobal(cfg, path)
}

// printJSON outputs a value as formatted JSON to stdout.
func printJSON(cmd *cobra.Command, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}

// printHuman outputs a human-readable string to stdout.
func printHuman(cmd *cobra.Command, msg string) {
	fmt.Fprintln(cmd.OutOrStdout(), msg)
}

// fatal prints an error to stderr and exits with code 1.
func fatal(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
	os.Exit(1)
}

// ExecuteCommand is a test helper that executes a Cobra command with the given
// arguments and captures its output. It resets the command state between calls.
func ExecuteCommand(cmd *cobra.Command, args ...string) (string, error) {
	// Reset flags for re-execution
	cmd.SetArgs(args)

	// Capture output
	oldOut := cmd.OutOrStdout()
	oldErr := cmd.ErrOrStderr()
	r, w, _ := os.Pipe()
	cmd.SetOut(w)
	cmd.SetErr(w)

	err := cmd.Execute()

	w.Close()
	cmd.SetOut(oldOut)
	cmd.SetErr(oldErr)

	// Read captured output
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	return string(buf[:n]), err
}

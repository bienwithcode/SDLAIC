// Package cmd contains all Cobra command definitions for the SDLAIC CLI.
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

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
	rootCmd.PersistentFlags().StringVar(&homeFlag, "home", "", "Override the home directory used for ~/.sdlaic (also SDLAIC_HOME)")
}

// changeFlag is shared across commands that accept --change/-c.
var changeFlag string

// homeFlag overrides the home directory for every command that reads or writes
// ~/.sdlaic. Every command now needs the global config, so this has to be
// injectable or the test suite writes into the developer's real home.
var homeFlag string

// resolveHome returns the home directory backing ~/.sdlaic, in precedence order:
// the --home flag, then SDLAIC_HOME, then the OS home directory.
func resolveHome() string {
	if homeFlag != "" {
		return homeFlag
	}
	if env := os.Getenv("SDLAIC_HOME"); env != "" {
		return env
	}
	home, err := os.UserHomeDir()
	if err != nil {
		// Matches the long-standing fallback in internal/workspace.
		return "/tmp"
	}
	return home
}

// globalConfigPath returns the path of the global config for the resolved home.
func globalConfigPath() string {
	return filepath.Join(resolveHome(), ".sdlaic", "config.json")
}

// resolveChangeName returns the change name from the flag or the active change
// in the workspace config.
func resolveChangeName(flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}

	// The active change lives in the global config now, so commands that have
	// not been migrated yet still have to ask for it through resolveProject —
	// otherwise a change set by `new` would be invisible to `gate` or `validate`.
	project, err := resolveProject()
	if err != nil {
		return "", domain.ErrNoActiveChange
	}
	return project.resolveChange("")
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
//
// The sink is a bytes.Buffer rather than an os.Pipe. A pipe deadlocks here:
// nothing drains the read end until Execute returns, so a command whose output
// exceeds the pipe buffer blocks forever mid-write. `completion bash` emits ~28KB,
// which fits Linux and macOS buffers but not the ~4KB Windows one — the failure
// was invisible until CI ran on Windows. Reading a fixed 4096-byte slice
// afterwards also truncated every longer output, silently, on every platform.
func ExecuteCommand(cmd *cobra.Command, args ...string) (string, error) {
	// Reset flags for re-execution
	cmd.SetArgs(args)

	// Capture output
	oldOut := cmd.OutOrStdout()
	oldErr := cmd.ErrOrStderr()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()

	cmd.SetOut(oldOut)
	cmd.SetErr(oldErr)

	return out.String(), err
}

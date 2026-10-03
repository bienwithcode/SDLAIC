package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// archiveCmd represents the `sdlaic archive` command.
var archiveCmd = &cobra.Command{
	Use:   "archive <change-name>",
	Short: "Archive a completed change",
	Long: `Moves the change directory into the .archive/ directory under a
date-prefixed name (YYYY-MM-DD-<change-name>), keeping the files plain and
readable — no compression. A change name that already starts with a date
prefix keeps its existing name. If the target already exists, the archive
is refused and nothing is overwritten. If the archived change was active,
clears the active change.`,
	Args: cobra.ExactArgs(1),
	RunE: runArchive,
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}

// archiveDatePrefixPattern matches the `YYYY-MM-DD-` prefix that archiving
// prepends to a change name. A change whose name already starts with one is
// archived under its existing name so the prefix is never stacked (same
// convention as OpenSpec #1309).
var archiveDatePrefixPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-`)

func runArchive(cmd *cobra.Command, args []string) error {
	changeName := args[0]

	project, err := resolveProject()
	if err != nil {
		return err
	}

	changePath, err := project.changePath(changeName)
	if err != nil {
		return fmt.Errorf("resolving change path: %w", err)
	}

	// Verify change exists
	if info, err := os.Stat(changePath); err != nil || !info.IsDir() {
		return fmt.Errorf("change %q not found: %w", changeName, domain.ErrChangeNotFound)
	}

	// Create .archive directory
	basePath, err := project.changesDir()
	if err != nil {
		return err
	}
	archiveDir := filepath.Join(basePath, ".archive")

	archiveName := changeName
	if !archiveDatePrefixPattern.MatchString(changeName) {
		archiveName = time.Now().Format("2006-01-02") + "-" + changeName
	}
	archivePath := filepath.Join(archiveDir, archiveName)

	// Refuse to overwrite an existing archive. The old tar.gz behavior
	// silently replaced the previous archive AND deleted the original
	// change directory — destroying both copies on a name collision.
	if _, err := os.Stat(archivePath); err == nil {
		return fmt.Errorf("archive %q already exists at %s — nothing was overwritten; remove the old archive or rename the change first", archiveName, archivePath)
	}

	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("creating archive directory: %w", err)
	}

	// Plain move, no compression: both paths live under the same changes
	// directory, so os.Rename is atomic and never crosses a filesystem.
	// Archived artifacts stay directly readable by agents and by
	// `sdlaic list --all` without extracting anything.
	if err := os.Rename(changePath, archivePath); err != nil {
		return fmt.Errorf("moving change to archive: %w", err)
	}

	// Clear active change if it was this one
	if project.ActiveChange == changeName {
		if err := project.setActiveChange(""); err != nil {
			return fmt.Errorf("clearing active change: %w", err)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Archived %q as %q in %s\n", changeName, archiveName, archiveDir)
	return nil
}

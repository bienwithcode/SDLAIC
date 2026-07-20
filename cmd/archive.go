package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/bienwithcode/SDLAIC/internal/storage"
	"github.com/bienwithcode/SDLAIC/internal/workspace"
)

// archiveCmd represents the `sdlaic archive` command.
var archiveCmd = &cobra.Command{
	Use:   "archive <change-name>",
	Short: "Archive a completed change",
	Long: `Compresses a change directory into a tar.gz archive, moves it to
the .archive/ directory, and removes the original. If the archived
change was active, clears the active change.`,
	Args: cobra.ExactArgs(1),
	RunE: runArchive,
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}

func runArchive(cmd *cobra.Command, args []string) error {
	changeName := args[0]

	// Find workspace
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	root, err := workspace.FindWorkspace(cwd)
	if err != nil {
		return fmt.Errorf("no SDLAIC workspace found: %w", err)
	}

	workspaceRoot = root

	cfg, err := loadLocalConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Resolve change path
	homeDir, _ := os.UserHomeDir()
	changePath, err := storage.ResolvePath(cfg.Storage, root, homeDir, changeName)
	if err != nil {
		return fmt.Errorf("resolving change path: %w", err)
	}

	// Verify change exists
	if info, err := os.Stat(changePath); err != nil || !info.IsDir() {
		return fmt.Errorf("change %q not found: %w", changeName, domain.ErrChangeNotFound)
	}

	// Create .archive directory
	basePath, err := storage.ChangesBasePath(cfg.Storage, root, homeDir)
	if err != nil {
		return fmt.Errorf("resolving base path: %w", err)
	}
	archiveDir := filepath.Join(basePath, ".archive")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("creating archive directory: %w", err)
	}

	// Create tar.gz archive
	archivePath := filepath.Join(archiveDir, changeName+".tar.gz")
	if err := createTarGz(changePath, archivePath); err != nil {
		return fmt.Errorf("creating archive: %w", err)
	}

	// Remove original directory
	if err := os.RemoveAll(changePath); err != nil {
		return fmt.Errorf("removing original directory: %w", err)
	}

	// Clear active change if it was this one
	if cfg.ActiveChange == changeName {
		cfg.ActiveChange = ""
		if err := saveLocalConfig(cfg); err != nil {
			return fmt.Errorf("clearing active change: %w", err)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Archived %q to %s\n", changeName, archivePath)
	return nil
}

// createTarGz creates a tar.gz archive of the source directory.
func createTarGz(srcDir string, destPath string) error {
	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	gzw := gzip.NewWriter(outFile)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	return filepath.Walk(srcDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, filePath)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		return err
	})
}

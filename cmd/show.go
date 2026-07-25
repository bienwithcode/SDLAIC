package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/bienwithcode/SDLAIC/internal/state"
	"github.com/bienwithcode/SDLAIC/internal/storage"
	"github.com/bienwithcode/SDLAIC/internal/workspace"
)

// showCmd represents the `sdlaic show` command.
var showCmd = &cobra.Command{
	Use:   "show <change-name>",
	Short: "Display a change's artifact summary",
	Long:  `Prints artifact names, sizes, last modified dates, and a brief content preview.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)
}

func runShow(cmd *cobra.Command, args []string) error {
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

	homeDir, _ := os.UserHomeDir()
	changePath, err := storage.ResolvePath(cfg.Storage, root, homeDir, changeName)
	if err != nil {
		return fmt.Errorf("resolving change path: %w", err)
	}

	// Verify change exists
	if info, err := os.Stat(changePath); err != nil || !info.IsDir() {
		return fmt.Errorf("change %q not found: %w", changeName, domain.ErrChangeNotFound)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Change: %s\n", changeName)
	fmt.Fprintf(cmd.OutOrStdout(), "Path:   %s\n\n", changePath)
	fmt.Fprintf(cmd.OutOrStdout(), "Artifacts:\n")

	for _, at := range domain.OrderedArtifactTypes() {
		if at == domain.ArtifactSpec {
			specsDir := filepath.Join(changePath, "specs")
			var totalSize int64
			var latestMod time.Time
			var count int
			_ = filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() && state.IsCapabilitySpec(specsDir, path) {
					totalSize += info.Size()
					if info.ModTime().After(latestMod) {
						latestMod = info.ModTime()
					}
					count++
				}
				return nil
			})

			if count == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  %-30s (not found)\n", at.FileName())
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "  %-30s %6d bytes  %s (%d specs)\n",
					at.FileName(),
					totalSize,
					latestMod.Format("2006-01-02 15:04"),
					count,
				)
			}
			continue
		}

		filePath := filepath.Join(changePath, at.FileName())
		info, err := os.Stat(filePath)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "  %-30s (not found)\n", at.FileName())
			continue
		}

		fmt.Fprintf(cmd.OutOrStdout(), "  %-30s %6d bytes  %s\n",
			at.FileName(),
			info.Size(),
			info.ModTime().Format("2006-01-02 15:04"),
		)
	}

	return nil
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bienwithcode/SDLAIC/internal/domain"
	"github.com/bienwithcode/SDLAIC/internal/gatestate"
	"github.com/bienwithcode/SDLAIC/internal/state"
	"github.com/bienwithcode/SDLAIC/internal/storage"
	"github.com/bienwithcode/SDLAIC/internal/workspace"
)

var validateStrict bool

// validateCmd represents the `sdlaic validate` command.
var validateCmd = &cobra.Command{
	Use:   "validate [change-name]",
	Short: "Validate artifact files for format compliance",
	Long: `Checks artifact files for:
  - Valid markdown structure
  - Checkbox syntax in tasks.md (- [ ] / - [x])
  - No template placeholders ({{...}})
  - (strict mode) All artifacts up to current phase exist and are populated`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "Fail on non-critical warnings")
}

// resetValidateFlags resets validate command flags to defaults.
func resetValidateFlags() {
	validateStrict = false
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Find workspace
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	root, err := workspace.FindWorkspace(cwd)
	if err != nil {
		return fmt.Errorf("no SDLAIC workspace found (run 'sdlaic init' first): %w", err)
	}

	workspaceRoot = root

	// Load config
	cfg, err := loadLocalConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Resolve change name from positional arg, then flag, then active change
	var changeName string
	if len(args) > 0 {
		changeName = args[0]
	} else {
		var err error
		changeName, err = resolveChangeName(changeFlag)
		if err != nil {
			return err
		}
	}

	// Resolve change path
	homeDir, _ := os.UserHomeDir()
	changePath, err := storage.ResolvePath(cfg.Storage, root, homeDir, changeName)
	if err != nil {
		return fmt.Errorf("resolving change path: %w", err)
	}

	var violations []string

	// Check each artifact
	for _, at := range domain.OrderedArtifactTypes() {
		var contentsToValidate []struct {
			name    string
			content string
		}

		if at == domain.ArtifactSpec {
			// The spec artifact is directory-based: specs/<capability>/spec.md.
			specsDir := filepath.Join(changePath, "specs")
			_ = filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() && state.IsCapabilitySpec(specsDir, path) {
					data, err := os.ReadFile(path)
					if err == nil {
						relPath, _ := filepath.Rel(changePath, path)
						contentsToValidate = append(contentsToValidate, struct {
							name    string
							content string
						}{relPath, string(data)})
					}
				}
				return nil
			})
		} else {
			filePath := filepath.Join(changePath, at.FileName())
			data, err := os.ReadFile(filePath)
			if err == nil {
				contentsToValidate = append(contentsToValidate, struct {
					name    string
					content string
				}{at.FileName(), string(data)})
			}
		}

		for _, item := range contentsToValidate {
			// Check for placeholders
			if hasPlaceholder(item.content) {
				violations = append(violations, fmt.Sprintf("%s: contains template placeholders", item.name))
			}

			// Check tasks.md has checkbox syntax
			if at == domain.ArtifactTasks && item.content != "" {
				if !hasCheckbox(item.content) {
					violations = append(violations, "tasks.md: missing checkbox syntax (- [ ] or - [x])")
				}
			}
		}
	}

	// Strict mode: check artifacts up to current phase are populated
	if validateStrict {
		// Load gate state to respect explicit skips
		store := gatestate.NewWithHome(homeDir, cfg.ProjectHash, changeName)
		gf, _ := store.Load()

		// Read all artifacts and check phases
		for _, at := range domain.OrderedArtifactTypes() {
			var gateKey string
			switch at {
			case domain.ArtifactProposal:
				gateKey = "proposal"
			case domain.ArtifactSpec:
				gateKey = "spec"
			case domain.ArtifactDesign:
				gateKey = "design"
			case domain.ArtifactTasks:
				gateKey = "tasks"
			}

			if gateKey != "" {
				if g, ok := gf.Gates[gateKey]; ok && g.Status == domain.GateStatusSkipped {
					continue // explicitly skipped, no artifact required
				}
			}

			populated := false

			label := at.FileName()
			if at == domain.ArtifactSpec {
				// The spec artifact is directory-based: specs/<capability>/spec.md.
				label = "specs/<capability>/spec.md"
				specsDir := filepath.Join(changePath, "specs")
				_ = filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
					if err == nil && !info.IsDir() && state.IsCapabilitySpec(specsDir, path) {
						data, err := os.ReadFile(path)
						if err == nil && strings.TrimSpace(string(data)) != "" {
							populated = true
						}
					}
					return nil
				})
			} else {
				filePath := filepath.Join(changePath, at.FileName())
				data, err := os.ReadFile(filePath)
				if err == nil && strings.TrimSpace(string(data)) != "" {
					populated = true
				}
			}

			if !populated {
				violations = append(violations, fmt.Sprintf("%s: missing or empty (strict mode)", label))
			}
		}
	}

	if len(violations) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "Validation failed for %q:\n", changeName)
		for _, v := range violations {
			fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", v)
		}
		return domain.ErrValidationFailed
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Validation passed for %q\n", changeName)
	return nil
}

// hasPlaceholder checks if content contains {{...}} template placeholders or untouched template comments.
func hasPlaceholder(content string) bool {
	// Matches curly braces: {{...}}
	if regexp.MustCompile(`\{\{.*?\}\}`).MatchString(content) {
		return true
	}
	// Matches template comment instructions starting with typical instruction verbs
	instructionRe := regexp.MustCompile(`<!--\s*(Describe|Provide|List|Define|One-paragraph|State|Impact|Security|Backward|New|Requirement|What|Address|Task|Any additional|component|workflow|describe|information)\b`)
	return instructionRe.MatchString(content)
}

// hasCheckbox checks if content contains markdown checkbox syntax.
func hasCheckbox(content string) bool {
	return strings.Contains(content, "- [ ]") || strings.Contains(content, "- [x]")
}

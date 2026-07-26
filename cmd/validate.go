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
	project, err := resolveProject()
	if err != nil {
		return err
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
	changePath, err := project.changePath(changeName)
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
		store := gatestate.NewWithHome(resolveHome(), project.Hash, changeName)
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
				// Only an EXPLICIT skip (gate set --status skipped, SkippedAt set) exempts
				// the artifact. An auto-skip from a light/free default (SkippedAt nil) must
				// still require the artifact, so tightening a change into strict does not
				// silently trust unreviewed work (matches Gate.IsPassingFor).
				if g, ok := gf.Gates[gateKey]; ok && g.Status == domain.GateStatusSkipped && g.SkippedAt != nil {
					continue
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

// placeholderAngle matches a template angle-bracket token, e.g. <change-name>,
// <exact command>, <KEY>. It is letter-led so it ignores HTML-comment openers
// (<!--), comparison operators (<5, < b), and the arrow (->, <-).
var placeholderAngle = regexp.MustCompile(`<[A-Za-z][A-Za-z0-9 _./-]*>`)

// placeholderBracket matches a [...] span, capturing the inner text.
var placeholderBracket = regexp.MustCompile(`\[([^\[\]]*)\]`)

// checkboxInner matches the inner text of a markdown checkbox glyph ([ ] / [x]).
var checkboxInner = regexp.MustCompile(`(?i)^\s*x?\s*$`)

// knownTDDTag matches the structural task tags used in tasks.md (see skills/plan
// and skills/apply). These are real content, not fill-in placeholders, so they
// are excluded from bracket-placeholder detection.
var knownTDDTag = regexp.MustCompile(`^(TEST-RED:(unit|feature|e2e)|IMPL|VERIFY|WIRING|REFACTOR|NO-TEST|COMMIT)$`)

// placeholderComment matches an instructional HTML comment by leading verb. The
// verbs cover every template comment; "Populated"/"From" cover the grill log rows.
var placeholderComment = regexp.MustCompile(`<!--\s*(Describe|Provide|List|Define|One-paragraph|State|Impact|Security|Backward|New|Requirement|What|Address|Task|Any additional|component|workflow|describe|information|Populated|From)\b`)

// hasPlaceholder reports whether content still contains unfilled template
// placeholders: mustache {{...}}, angle tokens like <change-name>, bracket
// fill-ins like [One paragraph: ...], or instructional HTML comments. Markdown
// checkbox glyphs ([ ]/[x]), the known TDD task tags, and markdown link targets
// ([text](url)) are excluded so legitimate tasks.md content is not flagged.
func hasPlaceholder(content string) bool {
	// Mustache: {{...}}
	if regexp.MustCompile(`\{\{.*?\}\}`).MatchString(content) {
		return true
	}
	// Angle-bracket token: <change-name>, <exact command>, ...
	if placeholderAngle.MatchString(content) {
		return true
	}
	// Bracket fill-in: flag any inner text that is not a checkbox, a known TDD
	// tag, or a markdown link target.
	for _, m := range placeholderBracket.FindAllStringSubmatchIndex(content, -1) {
		inner := content[m[2]:m[3]] // first capture group (text inside the brackets)
		if checkboxInner.MatchString(inner) {
			continue
		}
		if knownTDDTag.MatchString(inner) {
			continue
		}
		if m[1] < len(content) && content[m[1]] == '(' {
			continue // markdown link: [text](url)
		}
		return true
	}
	// Instructional HTML comment: <!-- Describe ... -->
	return placeholderComment.MatchString(content)
}

// hasCheckbox checks if content contains markdown checkbox syntax.
func hasCheckbox(content string) bool {
	return strings.Contains(content, "- [ ]") || strings.Contains(content, "- [x]")
}

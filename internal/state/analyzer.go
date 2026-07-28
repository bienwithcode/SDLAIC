// Package state analyzes the state of SDLAIC change artifacts to determine
// the current phase and artifact status.
package state

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/bienwithcode/SDLAIC/internal/domain"
)

// AnalyzePhase determines the current phase of a change by examining which
// artifact files exist and are populated. Phases progress in order:
// EMPTY → CONTEXT → PROPOSED → SPECIFIED → DESIGNED → PLANNED → IMPLEMENTED
//
// This is artifact-only detection. Gate verdicts are tracked separately via
// `sdlaic gate status`; the `enforcer` skill couples artifact presence with
// gate status.
func AnalyzePhase(changeDir string) (domain.Phase, error) {
	// Verify the change directory exists
	info, err := os.Stat(changeDir)
	if err != nil {
		return "", fmt.Errorf("accessing change directory %s: %w", changeDir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", changeDir)
	}

	artifacts, err := AnalyzeArtifacts(changeDir)
	if err != nil {
		return "", fmt.Errorf("analyzing artifacts in %s: %w", changeDir, err)
	}

	// Check phases in reverse order to find the highest achieved phase.
	// Phases must be contiguous — you can't skip ahead.

	// IMPLEMENTED: all artifacts populated + all tasks checked
	allPopulated := true
	for _, at := range domain.OrderedArtifactTypes() {
		status := artifacts[string(at)]
		if !status.Populated {
			allPopulated = false
			break
		}
	}

	if allPopulated {
		tasksContent, err := readFileContent(filepath.Join(changeDir, domain.ArtifactTasks.FileName()))
		if err == nil && AllTasksChecked(tasksContent) {
			return domain.PhaseImplemented, nil
		}
	}

	// Check each phase in reverse (PLANNED → CONTEXT)
	orderedTypes := domain.OrderedArtifactTypes()

	// PLANNED: tasks.md is populated (and all before it)
	if isPhaseComplete(artifacts, orderedTypes, domain.ArtifactTasks) {
		return domain.PhasePlanned, nil
	}

	// DESIGNED: design.md is populated
	if isPhaseComplete(artifacts, orderedTypes, domain.ArtifactDesign) {
		return domain.PhaseDesigned, nil
	}

	// SPECIFIED: specs/<capability>/spec.md is populated
	if isPhaseComplete(artifacts, orderedTypes, domain.ArtifactSpec) {
		return domain.PhaseSpecified, nil
	}

	// PROPOSED: proposal.md is populated
	if isPhaseComplete(artifacts, orderedTypes, domain.ArtifactProposal) {
		return domain.PhaseProposed, nil
	}

	// CONTEXT: context.md is populated
	if isPhaseComplete(artifacts, orderedTypes, domain.ArtifactContext) {
		return domain.PhaseContext, nil
	}

	return domain.PhaseEmpty, nil
}

// isPhaseComplete checks that the given artifact and all preceding artifacts are populated.
func isPhaseComplete(artifacts map[string]domain.ArtifactStatus, orderedTypes []domain.ArtifactType, target domain.ArtifactType) bool {
	for _, at := range orderedTypes {
		status := artifacts[string(at)]
		if !status.Populated {
			return false // Gap found — phase not reached
		}
		if at == target {
			return true // Target and all before it are populated
		}
	}
	return false
}

// AnalyzeArtifacts checks each artifact type and returns a map of artifact name
// to its status (exists, populated, valid).
func AnalyzeArtifacts(changeDir string) (map[string]domain.ArtifactStatus, error) {
	result := make(map[string]domain.ArtifactStatus)

	for _, at := range domain.OrderedArtifactTypes() {
		status := domain.ArtifactStatus{}

		if at == domain.ArtifactSpec {
			// The spec artifact is directory-based: specs/<capability>/spec.md.
			// Emit one entry per capability plus an aggregate "spec" entry. The
			// aggregate's Populated is the AND of all capabilities (so a
			// half-written spec tier does not read as populated); it keeps
			// isPhaseComplete and the status JSON contract working, while the
			// spec:<capability> entries carry the per-capability detail.
			caps, _ := ListCapabilities(changeDir)
			if len(caps) == 0 {
				result[string(at)] = domain.ArtifactStatus{}
				continue
			}
			allPopulated := true
			var anyExists bool
			for _, c := range caps {
				cs := domain.ArtifactStatus{}
				if data, err := os.ReadFile(filepath.Join(changeDir, "specs", c, "spec.md")); err == nil {
					cs.Exists = true
					anyExists = true
					cs.Populated = IsPopulated(string(data))
					cs.Valid = cs.Populated
					if !cs.Populated {
						allPopulated = false
					}
				} else {
					allPopulated = false
				}
				result["spec:"+c] = cs
			}
			result[string(at)] = domain.ArtifactStatus{Exists: anyExists, Populated: allPopulated, Valid: allPopulated}
			continue
		}

		filePath := filepath.Join(changeDir, at.FileName())
		data, err := os.ReadFile(filePath)
		if err != nil {
			// File doesn't exist
			result[string(at)] = status
			continue
		}

		status.Exists = true
		content := string(data)
		status.Populated = IsPopulated(content)
		status.Valid = status.Populated // Basic validity = populated

		result[string(at)] = status
	}

	return result, nil
}

// IsPopulated checks whether file content has real, meaningful content
// (not just whitespace, markdown headers, or HTML comments).
func IsPopulated(content string) bool {
	// Strip HTML comments
	stripped := stripHTMLComments(content)

	// Strip markdown header lines (lines starting with #)
	headerRe := regexp.MustCompile(`(?m)^#+\s.*$`)
	stripped = headerRe.ReplaceAllString(stripped, "")

	// Check if remaining content has non-whitespace characters
	return len(strings.TrimSpace(stripped)) > 0
}

// AllTasksChecked returns true if the tasks.md content has at least one checkbox
// and all checkboxes are checked ([x]).
func AllTasksChecked(content string) bool {
	unchecked := regexp.MustCompile(`- \[ \]`)
	checked := regexp.MustCompile(`- \[x\]`)

	hasChecked := checked.MatchString(content)
	hasUnchecked := unchecked.MatchString(content)

	return hasChecked && !hasUnchecked
}

// stripHTMLComments removes HTML comment blocks from content.
func stripHTMLComments(content string) string {
	re := regexp.MustCompile(`(?s)<!--.*?-->`)
	return re.ReplaceAllString(content, "")
}

// readFileContent reads and returns the content of a file as a string.
func readFileContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// IsCapabilitySpec reports whether path is a spec leaf nested under exactly one
// capability directory beneath specsDir — i.e. specs/<capability>/spec.md. It
// rejects specs/spec.md (no capability dir) and specs/a/b/spec.md (nested too
// deep), matching the documented capability-spec contract.
func IsCapabilitySpec(specsDir, path string) bool {
	rel, err := filepath.Rel(specsDir, path)
	if err != nil {
		return false
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	return len(parts) == 2 && parts[0] != "" && strings.EqualFold(parts[1], "spec.md")
}

// ListCapabilities returns the sorted names of the immediate capability
// directories under <changePath>/specs — i.e. one level deep, specs/<capability>/.
// It returns an empty slice (no error) when specs/ does not exist or is not a
// directory. Files placed directly under specs/ (e.g. a malformed specs/spec.md)
// are not capability directories and are ignored, matching IsCapabilitySpec.
//
// Use this to enumerate capabilities anywhere the spec artifact must be treated
// per-capability (gate keys, validation, status fan-out) rather than collapsed
// into one blob.
func ListCapabilities(changePath string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(changePath, "specs"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var caps []string
	for _, e := range entries {
		if e.IsDir() {
			caps = append(caps, e.Name())
		}
	}
	sort.Strings(caps)
	return caps, nil
}

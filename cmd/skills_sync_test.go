package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sum12(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])[:12]
}

// TestSkillReferencesNoDrift guards reference files bundled under more than
// one skills/<name>/references/ tree: any such file must be byte-identical
// across every copy, so accidental divergence becomes a CI failure instead of
// silent behavioural drift. (The shared docs — references/code-research.md,
// agents/*.md — intentionally live once at the plugin root and are covered by
// TestSharedRootReferences below.)
func TestSkillReferencesNoDrift(t *testing.T) {
	const skillsRoot = "../skills"

	// key: path relative to the skill's references/ dir (e.g. "grills/scope-grill.md")
	type copyInfo struct {
		location string // skill-relative path of the first-seen copy
		content  []byte
		sum      string
	}
	seen := map[string]copyInfo{}

	err := filepath.WalkDir(skillsRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		rel, err := filepath.Rel(skillsRoot, path)
		if err != nil {
			return err
		}
		parts := strings.SplitN(rel, string(filepath.Separator), 3)
		// Only files under skills/<name>/references/... participate.
		if len(parts) < 3 || parts[1] != "references" {
			return nil
		}
		key := parts[2]

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if first, ok := seen[key]; ok {
			if !bytes.Equal(first.content, content) {
				t.Errorf("reference %q drifted between skills:\n  %s (sha256 %s)\n  %s (sha256 %s)\nkeep one canonical copy and sync the others",
					key, first.location, first.sum, path, sum12(content))
			}
			return nil
		}
		seen[key] = copyInfo{location: path, content: content, sum: sum12(content)}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", skillsRoot, err)
	}

	if len(seen) == 0 {
		t.Fatal("no reference files found under skills/ — test is misplaced")
	}
}

// TestSharedRootReferences guards the deliberate split of shared docs back to
// the plugin root: references/code-research.md and the two agent personas in
// agents/ exist exactly once at the root, no per-skill copies or personas/
// trees remain, and every path declared in the plugin manifests resolves.
func TestSharedRootReferences(t *testing.T) {
	const root = ".."

	// 1. The shared code-research doc lives once at the root and is not
	//    duplicated into any skill.
	sharedDocs := []string{"references/code-research.md", "references/sdlaic-standards.md"}
	for _, doc := range sharedDocs {
		if _, err := os.Stat(filepath.Join(root, doc)); err != nil {
			t.Errorf("shared doc %s missing at plugin root: %v", doc, err)
		}
	}

	personaNames := []string{"compliance-reviewer.md", "code-quality-reviewer.md"}
	for _, name := range personaNames {
		if _, err := os.Stat(filepath.Join(root, "agents", name)); err != nil {
			t.Errorf("agent persona agents/%s missing at plugin root: %v", name, err)
		}
	}

	// 2. No stale copies: per-skill code-research.md copies and personas/
	//    trees must not reappear under skills/.
	err := filepath.WalkDir(filepath.Join(root, "skills"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "personas" {
				t.Errorf("stale personas/ tree remains at %s — personas live in root agents/", path)
			}
			return nil
		}
		if d.Name() == "code-research.md" {
			t.Errorf("stale per-skill copy remains at %s — shared doc lives at root references/", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking skills/: %v", err)
	}

	// 3. Every path declared in .claude-plugin/plugin.json resolves.
	pluginRaw, err := os.ReadFile(filepath.Join(root, ".claude-plugin", "plugin.json"))
	if err != nil {
		t.Fatalf("reading .claude-plugin/plugin.json: %v", err)
	}
	var plugin struct {
		Agents []string `json:"agents"`
	}
	if err := json.Unmarshal(pluginRaw, &plugin); err != nil {
		t.Fatalf("parsing .claude-plugin/plugin.json: %v", err)
	}
	if len(plugin.Agents) != len(personaNames) {
		t.Errorf("plugin.json declares %d agents, want %d", len(plugin.Agents), len(personaNames))
	}
	for _, rel := range plugin.Agents {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Errorf("plugin.json agents entry %q does not resolve: %v", rel, err)
		}
	}

	// 4. Every agent name declared in gemini-extension.json maps to a file
	//    in agents/.
	geminiRaw, err := os.ReadFile(filepath.Join(root, "gemini-extension.json"))
	if err != nil {
		t.Fatalf("reading gemini-extension.json: %v", err)
	}
	var gemini struct {
		Agents []string `json:"agents"`
	}
	if err := json.Unmarshal(geminiRaw, &gemini); err != nil {
		t.Fatalf("parsing gemini-extension.json: %v", err)
	}
	if len(gemini.Agents) != len(personaNames) {
		t.Errorf("gemini-extension.json declares %d agents, want %d", len(gemini.Agents), len(personaNames))
	}
	for _, name := range gemini.Agents {
		if _, err := os.Stat(filepath.Join(root, "agents", name+".md")); err != nil {
			t.Errorf("gemini-extension.json agent %q does not resolve to agents/%s.md: %v", name, name, err)
		}
	}
}

package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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

// TestSkillReferencesNoDrift guards the deliberate duplication that
// self-containment requires: any reference file bundled under more than one
// skills/<name>/references/ tree must be byte-identical across every copy.
//
// Skills are self-contained per the Agent Skills standard (each skill ships
// its own references/), so a shared document such as code-research.md exists
// once per skill that needs it. This test turns accidental divergence between
// those copies into a CI failure instead of silent behavioural drift.
func TestSkillReferencesNoDrift(t *testing.T) {
	const skillsRoot = "../skills"

	// key: path relative to the skill's references/ dir (e.g. "code-research.md")
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

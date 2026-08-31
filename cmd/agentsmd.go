// Package cmd implements the sdlaic CLI commands.
//
// This file owns the AGENTS.md bootstrap: an idempotent, marker-delimited
// workflow block injected at the project root so coding agents (Pi, Claude
// Code, Codex, Gemini CLI) anchor on SDLAIC before loading any skill.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Marker comments delimit the managed block. Everything between them is owned
// by sdlaic and replaced on re-init; everything outside is user content and
// preserved byte-for-byte.
const (
	agentsMdBeginMarker = "<!-- sdlaic:begin -->"
	agentsMdEndMarker   = "<!-- sdlaic:end -->"
	agentsMdPath        = "AGENTS.md"
)

// agentsMdBlock is the managed workflow block written into the project's
// AGENTS.md. It is agent-agnostic on purpose: every supported runtime reads
// AGENTS.md, and the Pi-specific entry point (/skill:enforcer) is called out
// inline.
const agentsMdBlock = agentsMdBeginMarker + `
## SDLAIC workflow

This project runs [SDLAIC](https://github.com/bienwithcode/SDLAIC) — a phase-gated development process. Every agent working here follows it:

1. Before any other action, establish the current phase: run ` + "`sdlaic status`" + ` and ` + "`sdlaic gate status --json`" + `. On Pi, load the ` + "`enforcer`" + ` skill (` + "`/skill:enforcer`" + `) — it routes every turn.
2. No phase may be skipped. A phase is unblocked only when its artifact exists **and** its gate is approved or skipped.
3. Implementation code is written only during the ` + "`apply`" + ` phase, from a verified ` + "`tasks.md`" + `.
4. Change artifacts live in the changes directory reported by ` + "`sdlaic status`" + ` — never in ad-hoc files.
` + agentsMdEndMarker

// ensureAgentsMdBlock creates, appends to, or refreshes the managed SDLAIC
// block in <root>/AGENTS.md. Re-running it is a no-op diff: an existing block
// is replaced in place (so upgrades to the block content propagate), and a
// file without markers gets the block appended with a separating blank line.
func ensureAgentsMdBlock(root string) error {
	path := filepath.Join(root, agentsMdPath)

	existing, err := os.ReadFile(path)
	switch {
	case err == nil:
		// fall through
	case os.IsNotExist(err):
		existing = nil
	default:
		return fmt.Errorf("reading %s: %w", path, err)
	}

	updated := spliceAgentsBlock(string(existing), agentsMdBlock)
	if updated == string(existing) {
		return nil
	}

	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

// spliceAgentsBlock merges block into existing AGENTS.md content, normalising
// the blank-line separation on both sides of the managed block.
func spliceAgentsBlock(existing, block string) string {
	begin := strings.Index(existing, agentsMdBeginMarker)
	end := strings.Index(existing, agentsMdEndMarker)

	// Existing managed block: replace it in place, keeping surrounding content.
	if begin >= 0 && end > begin {
		prefix := strings.TrimRight(existing[:begin], "\n")
		rest := strings.Trim(existing[end+len(agentsMdEndMarker):], "\r\n")

		merged := prefix
		if merged != "" {
			merged += "\n\n"
		}
		merged += block
		if rest != "" {
			merged += "\n\n" + rest
		}
		return merged + "\n"
	}

	if strings.TrimSpace(existing) == "" {
		return block + "\n"
	}

	return strings.TrimRight(existing, "\n") + "\n\n" + block + "\n"
}

// changesDirInsideProject reports whether changesDir lives under root, i.e.
// whether the project contract "artifacts stay inside the project" holds. An
// external changes directory keeps the project untouched, and the AGENTS.md
// bootstrap follows the same rule.
func changesDirInsideProject(changesDir, root string) bool {
	rel, err := filepath.Rel(root, changesDir)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

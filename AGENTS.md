# AGENTS.md — SDLAIC

## Project Overview

SDLAIC = **SDLC + AI**. A CLI tool (Go) + AI skill framework (Markdown) that enforces phase-gated development for AI coding agents.

Two distinct parts with different formats and conventions:

### Part 1: `sdlaic` CLI (Go)

Artifact management — workspace init, change lifecycle, templating, validation, status reporting. Built with Cobra.

Key packages:
- `cmd/` — Cobra commands. Each command file (`init.go`, `new.go`, `status.go`, etc.) has a matching `_test.go`.
- `internal/domain/` — shared types: `StorageMode`, `WorkflowLevel`, `Phase`, `ArtifactType`, sentinel errors.
- `internal/state/` — determines phase from artifact file presence.
- `internal/storage/` — resolves change path by storage mode, git integration.
- `internal/config/` — reads/writes `.sdlaicrc` (local) and `~/.sdlaic/config.json` (global).
- `internal/workspace/` — walks up from cwd to find `.sdlaicrc`.
- `internal/templates/` — artifact template rendering.

State machine: the CLI does **not** own one. It tracks artifact files. The `enforcer` skill (Part 2) infers phase from which artifacts exist.

### Part 2: Skills & Agents (Markdown)

8 skill definitions in `skills/<name>/SKILL.md` — loaded by Claude Code, Codex, Gemini CLI via plugin manifests.

- `enforcer` — state machine router, runs every agent turn
- `new` → `grillme` → `brainstorm` → `plan` → `apply` → `review` → `verify` — phase-gated workflow

2 agent personas in `agents/` — `code-quality-reviewer.md`, `compliance-reviewer.md`.

Reference docs in `references/` — consumed by skills, not the CLI.

## Essential Commands

```bash
# Build
go build -ldflags "-X main.version=$(git describe --tags 2>/dev/null || echo dev)" -o sdlaic .

# Run CLI
./sdlaic init
./sdlaic new change "my-feature"
./sdlaic status

# Static analysis (catches bugs compiler misses: wrong printf formats, unreachable code, bad struct tags...)
go vet ./...

# Test (all)
go test ./...

# Test a single package
go test ./cmd/...
go test ./internal/domain/...

# Test with verbose
go test -v ./...
```

No external services needed for tests. Tests use temp directories via `t.TempDir()`.

## Code Style Guidelines

### Go (CLI)

- Module name: `sdlaic` (not `github.com/...`)
- Dependencies: `spf13/cobra`, `stretchr/testify` only. No heavy frameworks.
- Error wrapping: `fmt.Errorf("context: %w", err)`
- Sentinel errors in `internal/domain/` (`errors.New(...)`)
- Package-level doc comment on every package (`// Package X does Y.`)
- Table-driven tests using `assert`/`require` from testify
- Test helpers in `cmd/testutil_test.go` (`initWorkspaceForTest`, `ExecuteCommand`)
- Commands use `RunE` (return error, not `Run` + `os.Exit`)
- `cmd/` shares `workspaceRoot` var set by `discoverWorkspace()` via `PersistentPreRun`
- Flag resets in `_test.go` (`resetInitFlags()`, `resetStatusFlags()`, etc.) — Cobra flags persist between test runs

### Markdown (Skills & Agents) Style

- Frontmatter: `name` + `description` fields
- Sections: `## Core Principle`, `## Process`, `## Output Artifacts`, `## Verification`, `## Common Mistakes`, `## Handoff`
- Tone: imperative, no hedging. "You may NOT skip phases" not "please don't skip"
- Cross-references use relative paths: `references/code-research.md`

### Skill & Agent Authoring Rules

- **Global Applicability**: Rules, instructions, and workflows in skills must be written for global applicability. Do NOT tie rules to specific environments or platform-specific execution mechanisms.
- **Tool & API Abstraction**: Do NOT mandate or reference specific tool names (e.g., `grep_search`, `call_mcp_tool`, or particular MCP server names), specific API endpoints, or credentials. Always describe the required capability conceptually (e.g., "use a search tool", "perform codebase research", "invoke the API") so that the skills remain fully compatible and portable across all agent runtimes (Claude Code, Gemini CLI, etc.).

## Testing Instructions

### Go Tests

- Unit tests only — no integration tests, no network calls
- Each `cmd/*.go` has a `cmd/*_test.go` with the same name
- Test setup: `initWorkspaceForTest(t)` creates a temp dir with `.sdlaicrc` and returns its path
- Use `ExecuteCommand(rootCmd, args...)` to test CLI commands — it captures output
- Use `require` for setup steps (fail fast), `assert` for assertions (collect all failures)
- Reset Cobra flags in each test: call `resetXxxFlags()` at the top

### Skills

Skills are Markdown — no automated tests. Validation is:
1. `sdlaic validate <change> --strict` — checks artifact format (no placeholders, checkbox syntax in tasks.md)
2. Manual review against `references/sdlaic-standards.md`

## Security Considerations

- **No secrets in artifacts.** `context.md` may contain Jira ticket content — never store API tokens or credentials in change directories.
- **`.sdlaicrc` is project-local.** Contains `active_change` and `project_hash` only. No sensitive data.
- **`--storage global` mode** stores artifacts in `~/.sdlaic/` — ensure proper file permissions on multi-user machines.
- **No network calls from the CLI.** Jira interaction happens in the AI agent (skill `new`), not in `sdlaic` binary.
- **Template rendering** is string-based, no eval/exec. Artifacts are plain Markdown.

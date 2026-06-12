# SDLAIC

**SDLC + AI** — a play on *Software Development Life Cycle*, with AI in the name and in the workflow. A CLI tool and AI skill framework that enforces a phase-gated development process for AI coding agents. Agents cannot skip phases, write unverified code, or proceed without research.

SDLAIC consists of two parts:

1. **`sdlaic` CLI** (Go) — manages change artifacts: initialization, templating, validation, status tracking
2. **7 Skill definitions** (Markdown) — loaded by AI agents (Claude Code, Codex, Gemini CLI) to enforce workflow discipline

The state machine is driven by **artifact presence** — each phase requires a specific file to exist and be populated. No artifact, no progression.

## The Problem It Solves

AI coding agents tend to jump straight to writing code, skip design, skip testing, and drift from requirements. SDLAIC forces every change through a fixed pipeline where each phase produces a verifiable artifact before the next phase can begin.

## Workflow

```
new → grillme → brainstorm → plan → apply → review
```

| Skill | Phase | What It Does |
|-------|-------|-------------|
| **enforcer** | Every turn | Routes the agent to the correct phase based on artifact presence. No skipping. |
| **new** | Init | Fetches ticket context, runs codebase research, initializes a change |
| **grillme** | Challenge | Devil's advocate — asks tough questions one at a time, produces `rationale.md` |
| **brainstorm** | Design | Searches for existing patterns, writes `proposal.md` → optional `specs/` → `design.md` |
| **plan** | Planning | Breaks proposal + design into ordered tasks with verification steps → `tasks.md` |
| **apply** | Execution | Implements one task at a time, verifies each, commits per task |
| **review** | Audit | Compares code against proposal + design + spec → `review.md` |

## Artifacts

Each phase produces a file in `.sdlaic/changes/<change-name>/`:

```
.sdlaic/changes/<change-name>/
├── context.md           # Ticket description, research summary, actors & use cases
├── rationale.md         # Validated assumptions, decisions, scope boundaries
├── proposal.md          # Business scope: why, what changes, impact
├── specs/               # Optional — formal WHEN/THEN requirement specs
│   └── <capability>/
│       └── spec.md
├── design.md            # Technical design: architecture, decisions, risks
├── tasks.md             # Ordered tasks with verification steps (checkbox syntax)
└── review.md            # Audit findings against proposal + design + spec
```

## Installation

### Build from source

```bash
git clone <repo-url>
cd SDLAIC/sdlaic
go build -ldflags "-X main.version=v0.1.0" -o sdlaic .
```

### Initialize a workspace

```bash
cd your-project
sdlaic init                          # defaults: local storage, strict workflow
sdlaic init --storage ignored        # artifacts in .sdlaic/ (gitignored)
sdlaic init --storage global         # artifacts in ~/.sdlaic/stores/<hash>/
sdlaic init --workflow light         # some phases can be skipped
```

## CLI Commands

```bash
sdlaic init                           # Initialize workspace
sdlaic new change "<name>"            # Create a new change
sdlaic status                         # Show current phase and artifact status
sdlaic status --json                  # Machine-readable status
sdlaic instructions <artifact> -c <name>  # Get template for an artifact
sdlaic validate <name>                # Validate artifact format
sdlaic validate <name> --strict       # Strict: all artifacts must exist
sdlaic list                           # List active changes
sdlaic show <name>                    # Show change details
sdlaic switch <name>                  # Set active change
sdlaic archive <name>                 # Archive a completed change
```

## AI Agent Integration

### Claude Code / Codex

Install the plugin manifest:

- **Claude Code**: `.claude-plugin/plugin.json`
- **Codex**: `.codex-plugin/plugin.json`

### Gemini CLI

Register the extension via `gemini-extension.json`.

### What agents get

- **8 skills** in `skills/` — one per phase, each with a `SKILL.md` defining behavior
- **2 agent personas** in `agents/` — code quality reviewer, compliance reviewer
- **Reference docs** in `references/` — code research contract, standards reference

The `enforcer` skill runs at the start of every agent turn, checks which artifacts exist via `sdlaic status`, and routes to the correct phase. Agents cannot bypass this.

## Core Rules

1. **No unverified code** — implementation only in `apply`, only with a verified `tasks.md`
2. **Research first** — every phase starts with a code research query
3. **One at a time** — one question, one task, one commit
4. **Evidence over opinion** — verification commands are proof, not "looks right"
5. **No scope creep** — implement what's planned, park everything else

## Project Structure

```
sdlaic/
├── main.go                         # Entry point (version injected at build time)
├── cmd/                            # Cobra CLI commands
│   ├── root.go                     # Root command, workspace discovery
│   ├── init.go                     # sdlaic init
│   ├── new.go                      # sdlaic new change
│   ├── status.go                   # sdlaic status
│   ├── validate.go                 # sdlaic validate
│   ├── instructions.go             # sdlaic instructions
│   ├── show.go                     # sdlaic show
│   ├── list.go                     # sdlaic list
│   ├── switch.go                   # sdlaic switch
│   ├── archive.go                  # sdlaic archive
│   ├── config.go                   # sdlaic config
│   ├── completion.go               # Shell completion
│   └── version.go                  # sdlaic version
├── internal/
│   ├── config/                     # Config file I/O (.sdlaicrc, config.json)
│   ├── domain/                     # Types, constants, phases, errors
│   ├── state/                      # Phase analysis from artifact presence
│   ├── storage/                    # Path resolution, git integration
│   ├── templates/                  # Artifact template rendering
│   └── workspace/                  # Workspace discovery and initialization
├── skills/                         # 8 AI agent skill definitions
│   ├── enforcer/SKILL.md
│   ├── new/SKILL.md
│   ├── grillme/SKILL.md
│   ├── brainstorm/SKILL.md
│   ├── plan/SKILL.md
│   ├── apply/SKILL.md
│   ├── review/SKILL.md
│   └── verify/SKILL.md
├── agents/                         # AI agent persona definitions
│   ├── code-quality-reviewer.md
│   └── compliance-reviewer.md
├── references/                     # Reference docs for skills
│   ├── code-research.md
│   └── sdlaic-standards.md
├── .claude-plugin/plugin.json      # Claude Code manifest
├── .codex-plugin/plugin.json       # Codex manifest
└── gemini-extension.json           # Gemini CLI manifest
```

## Storage Modes

| Mode | Location | Tracked by git |
|------|----------|----------------|
| `local` | `<project>/.sdlaic/changes/` | Yes |
| `ignored` | `<project>/.sdlaic/changes/` | No (auto-added to .gitignore) |
| `global` | `~/.sdlaic/stores/<hash>/changes/` | No |

## Workflow Levels

| Level | Behavior |
|-------|----------|
| `strict` | All phases must be completed in order — default |
| `light` | Some phases can be skipped |
| `free` | No phase enforcement |

## Prerequisites

- **Go 1.24+** — for building the CLI
- **Jira CLI** (`jira`) — for the `new` skill to fetch ticket context (optional if providing context manually)
- **A code research tool** — the skills require evidence-grounded research but don't mandate a specific tool. Use whatever your workspace provides (MCP server, editor search, `grep`). See `references/code-research.md`.

## License

[MIT](LICENSE)

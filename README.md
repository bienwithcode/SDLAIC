# SDLAIC

**SDLC + AI** — a play on *Software Development Life Cycle*, with AI in the name and in the workflow. A CLI tool and AI skill framework that enforces a phase-gated development process for AI coding agents. AI coding agents tend to jump straight to writing code, skip design, skip testing, and drift from requirements. SDLAIC forces every change through a fixed pipeline where each phase produces a verifiable artifact before the next phase can begin.

SDLAIC consists of two parts:
1. **`sdlaic` CLI** (Go) — manages change artifacts: initialization, templating, validation, status tracking, and AI plugin installation.
2. **AI Skill definitions** (Markdown) — loaded by AI agents (Claude Code, Codex, Gemini CLI) to enforce workflow discipline.

---

## Installation

### macOS / Linux
Install `sdlaic` using the single-line installation script:
```bash
curl -fsSL https://raw.githubusercontent.com/bienwithcode/SDLAIC/main/install.sh | sh
```

### Windows
Install `sdlaic` natively using the PowerShell installer:
```powershell
irm https://raw.githubusercontent.com/bienwithcode/SDLAIC/main/install.ps1 | iex
```

### Go Developers (Cross-platform)
If you have the Go compiler installed:
```bash
go install github.com/bienwithcode/SDLAIC@latest
```

---

## AI Agent Integration

### Claude Code (Recommended)
You can automatically configure and launch Claude Code with the SDLAIC plugin using a single command:
```bash
sdlaic open claude
```
*(This command automatically initializes the workspace if needed, registers the `bienwithcode` marketplace, installs the `sdlaic` plugin, and starts a Claude Code shell session).*

#### Manual / Direct Claude Installation
If you do not wish to use the Go CLI, you can register and install the plugin directly within a Claude Code session:
```bash
/plugin marketplace add bienwithcode/SDLAIC
/plugin install sdlaic@bienwithcode
```

### Codex
Support for Codex is coming in a later release.

---

## Workflow

The development lifecycle follows a fixed state machine driven by **artifact presence** — each phase requires a specific file to exist and be populated before progression is allowed:

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

---

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

---

## CLI Commands

```bash
sdlaic init                           # Initialize workspace (if not using auto-init)
sdlaic open claude                    # Install plugin and spawn Claude Code
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

---

## Storage Modes

| Mode | Location | Tracked by git |
|------|----------|----------------|
| `local` | `<project>/.sdlaic/changes/` | Yes |
| `ignored` | `<project>/.sdlaic/changes/` | No (auto-added to .gitignore) |
| `global` | `~/.sdlaic/stores/<hash>/changes/` | No |

---

## Workflow Levels

| Level | Behavior |
|-------|----------|
| `strict` | All phases must be completed in order — default |
| `light` | Some phases can be skipped |
| `free` | No phase enforcement |

---

## License

[MIT](LICENSE)

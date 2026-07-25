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

The development lifecycle is a series of **phase-gated micro-loops**. Each phase produces exactly one artifact, and progression is allowed only when that artifact exists **and** its gate has passed. This prevents an agent from designing against unverified requirements or writing code from an unapproved plan.

Each phase runs the same loop — **grill → draft → review → gate**:

```
new ─► context.md
        │  ┌─────────────── one micro-loop per phase ───────────────┐
        ▼  │ grill (challenge) → draft (write) → review (audit) → gate │
  proposal │ ── proposal.md ───────────────────────────► gate:proposal ✓
  spec     │ ── specs/<capability>/spec.md ─────────────► gate:spec ✓
  design   │ ── design.md ─────────────────────────────► gate:design ✓
  tasks    │ ── tasks.md ──────────────────────────────► gate:tasks ✓
        ▼  └───────────────────────────────────────────────────────┘
  apply ─► code ─► review code ─► review.md
```

Gate verdicts are stored **outside your repo** (see [Gate State](#gate-state)), so approvals never clutter your project.

| Skill | Phase | What It Does |
|-------|-------|-------------|
| **enforcer** | Every turn | Routes to the correct phase from artifact presence **and** gate status. No skipping, no advancing past a failed/pending gate. |
| **new** | Init | Fetches ticket context, runs codebase research, initializes a change → `context.md` |
| **grillme** | Grill (before every draft) | Parameterized Socratic challenge — loads the phase's grill checklist, asks tough questions one at a time; resolutions go in the artifact's *Challenge & Resolution Log* |
| **proposal** | Scope (1A) | Business scope with an explicit IN/OUT-OF-SCOPE contract → `proposal.md` |
| **spec** | Behavior (1B) | Formal GIVEN/WHEN/THEN requirements → `specs/<capability>/spec.md` |
| **design** | Architecture (2) | Input-boundary validation, subsystem boundaries, DRY reuse → `design.md` |
| **plan** | Planning (3) | Ordered TDD tasks grouped by subsystem milestone → `tasks.md` |
| **review** | Review (after every draft) | Parameterized independent audit — loads the phase's audit checklist, issues APPROVE / REQUEST_CHANGES / REJECT, records the verdict via `sdlaic gate set` |
| **apply** | Execution (4) | Implements one task at a time, verifies each, commits per task |
| **review** (code) | Final audit | Two-pass compliance + quality audit of the diff → `review.md` |

> `grillme` and `review` are **optional** and toggled by [workflow level](#workflow-levels): `strict` runs both; `light`/`free` skip them (draft-only fast path). The draft skill is always mandatory.
>
> The old monolithic `brainstorm` skill (which emitted proposal + spec + design in one pass) is **deprecated** — it caused generative overreach. Use `proposal` → `spec` → `design` instead.

---

## Artifacts

Each phase writes one artifact under `.sdlaic/changes/<change-name>/`:

```
.sdlaic/changes/<change-name>/
├── context.md           # Ticket description, research summary, actors & use cases
├── proposal.md          # Scope contract: why, IN/OUT-OF-SCOPE, impact
├── specs/               # Behavioral requirements (if user-facing)
│   └── <capability>/
│       └── spec.md      # Formal GIVEN/WHEN/THEN scenarios
├── design.md            # Technical design: architecture, boundaries, decisions
└── tasks.md             # Ordered TDD tasks by subsystem milestone (checkbox syntax)
```

Socratic-challenge output is recorded in a `## Challenge & Resolution Log` section **inside** each artifact (there is no separate `rationale.md`).

---

## Gate State

Gate verdicts are **never** written into your project repo. They live in a global state store, keyed by project + change:

```
~/.sdlaic/state/<project_hash>/<change-name>/
├── meta.json       # machine-readable gate state — source of truth for "approved?"
├── review.md       # human-readable mirror of the latest verdict + findings
└── history.jsonl   # append-only re-entry / follow-up events
```

A phase is **unblocked** only when its artifact exists **and** its gate is `approved` (or `skipped` in `light`/`free`). When a ticket changes mid-flight, `sdlaic gate reentry` re-enters the earliest affected artifact and supersedes everything downstream.

---

## CLI Commands

```bash
sdlaic init                           # Initialize workspace (if not using auto-init)
sdlaic open claude                    # Install plugin and spawn Claude Code
sdlaic new change "<name>"            # Create a new change
sdlaic status                         # Show current phase and artifact status
sdlaic status --json                  # Machine-readable status
sdlaic instructions <artifact> -c <name>  # Get template (proposal | spec | design | tasks)
sdlaic validate <name>                # Validate artifact format
sdlaic validate <name> --strict       # Strict: all artifacts must exist

# Gate state (verdicts stored in ~/.sdlaic/state/, never in your repo)
sdlaic gate status -c <name>          # Show gate state for each phase
sdlaic gate status -c <name> --json   # Machine-readable gate state
sdlaic gate set -c <name> --phase <proposal|spec|design|tasks> \
    --status <approved|failed|skipped> [--verdict <APPROVE|REQUEST_CHANGES|REJECT>] [--attempt]
sdlaic gate reentry -c <name> --from <phase> --reason "<why>"   # Mid-flight change

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

The workflow level controls whether the grill and review gates run around each draft:

| Level | Behavior |
|-------|----------|
| `strict` | Grill **and** review run every phase; a gate must be `approved` to advance — default |
| `light` | Draft-only fast path — grill/review skipped, gates auto-`skipped` |
| `free` | No gate enforcement — gates auto-`skipped` |

In every level the **draft skill is mandatory** and artifacts are still produced in order; only the grill/review gates are toggled.

---

## License

[MIT](LICENSE)

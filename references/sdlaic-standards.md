# SDLAIC Standards

## Directory Structure

All change artifacts live under:
```
.sdlaic/changes/<change-name>/
```

## Plugin Phase Flow

The plugin manages its own conceptual phases on top of SDLAIC. SDLAIC CLI does NOT have a state machine — it tracks artifact completion via `sdlaic status`. The plugin phases are enforced by the `enforcer` skill, not by the CLI.

```
NO_ACTIVE_CHANGE
  │ sdlaic new change "<name>"
  ▼
INITIALIZED (plugin state)
  │ Write rationale.md manually
  ▼
CHALLENGED (plugin state)
  │ sdlaic instructions proposal --change <name>, then write proposal.md → specs/ (optional) → design.md
  ▼
PROPOSED (plugin state)
  │ sdlaic instructions tasks --change <name>, then write tasks.md
  ▼
PLANNED (plugin state)
  │ Execute tasks, mark complete in tasks.md
  ▼
IMPLEMENTED (plugin state)
  │ Write review.md manually
  ▼
COMPLETE (plugin state)
  │ sdlaic validate <change-name> --strict, then archive change
  ▼
ARCHIVED
```

## Artifact Sequence

Each plugin phase requires a specific artifact. You may NOT advance without it.

| Phase | Required Artifact | Purpose |
|-------|------------------|---------|
| INITIALIZED | `context.md` (optional) | Jira ticket + research summary |
| CHALLENGED | `rationale.md` | Validated assumptions and decisions |
| PROPOSED | `proposal.md` | Business scope (written first) |
| PROPOSED | `specs/<capability>/spec.md` (optional) | Formal requirements — written before `design.md` when user-facing |
| PROPOSED | `design.md` | Technical architecture (written last, satisfies proposal + spec) |
| PLANNED | `tasks.md` | Ordered tasks with verification steps |
| IMPLEMENTED | Committed code | Source code changes |
| COMPLETE | `review.md` | Audit findings against proposal + design + spec |

### Full Artifact Structure

```
.sdlaic/changes/<change-name>/
├── context.md                          # Optional: Jira + research context
├── rationale.md                        # Challenge results
├── proposal.md                         # Business scope (why, what, impact) — written 1st
├── specs/                              # Optional — written 2nd, before design.md
│   └── <capability-name>/
│       └── spec.md                     # Formal requirements (WHEN/THEN scenarios)
├── design.md                           # Technical design (how, decisions, risks) — written 3rd
├── tasks.md                            # Ordered task list
└── review.md                           # Audit findings
```

## Commands Reference

```bash
# Check artifact completion status
sdlaic status --change <name>

# Initialize a new change
sdlaic new change "<kebab-case-name>"

# Get template instructions for writing an artifact
sdlaic instructions <artifact> --change <name>
# Supported artifacts: proposal, specs, design, tasks

# Validate artifacts
sdlaic validate <change-name> --strict

# Archive a completed change
sdlaic archive <change-name>

# List active changes
sdlaic list

# Show change details
sdlaic show <change-name>
```

## Rules

1. **Never advance without the artifact.** Each phase transition requires the corresponding artifact to exist and be populated.
2. **One active change at a time.** Complete or abandon the current change before starting a new one.
3. **Artifacts are the source of truth.** If code and artifact disagree, the artifact wins (during review).
4. **Validate before critical transitions.** Run `sdlaic validate <change-id> --strict` before archiving.
5. **Phase states are plugin-managed.** The enforcer skill tracks which phase you're in — the CLI only tracks artifact files.

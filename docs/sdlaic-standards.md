# SDLAIC Standards

## Directory Structure

All change artifacts live under the project's configured changes directory.
Never assume a path — ask the CLI, because the location is per-project
configuration and may sit outside the project entirely:
```
sdlaic path changes                      # the directory holding all changes
sdlaic path change --change <name>       # one change's directory
```
It defaults to `<project-root>/.sdlaic/changes/`, but `sdlaic init
--changes-dir <path>` can point it anywhere — for example at a shared
specifications repo — in which case SDLAIC creates nothing inside the project
at all.

Gate verdicts live OUTSIDE the project repo, in the global state store:
```
~/.sdlaic/state/<project_hash>/<change-name>/
├── meta.json      # machine-readable gate state — source of truth for "approved?"
├── review.md      # human-readable mirror of the latest verdict + findings
└── history.jsonl  # append-only re-entry / follow-up events
```

> **Two `review.md` files, two purposes.** The global `review.md` above is a
> machine-generated mirror of gate verdicts, written by the CLI. The pre-PR code
> audit produced by `skills/review` is a *separate* artifact written to
> `$(sdlaic path change --change <change-name>)/review.md`, alongside the other
> change artifacts.

## Phase-Gated Flow

The plugin manages phase-gated micro-loops on top of SDLAIC. The CLI does NOT own
a state machine — it tracks artifact files (`sdlaic status`) and gate verdicts
(`sdlaic gate status`). The `enforcer` skill infers the phase from **artifact
presence AND gate status**: a phase is unblocked only when its artifact exists
AND its gate is `approved` or `skipped`.

Each phase is a micro-loop: **grill (strict only) → draft → review → gate**. The
draft skill is mandatory; grill and review are toggled by workflow level
(`strict` = both on; `light`/`free` = draft-only, gates auto-`skipped`).

```
NO_ACTIVE_CHANGE
  │ sdlaic new change "<name>"   → context.md
  ▼
PROPOSED   [grillme proposal] → proposal → [review proposal] → gate:proposal ✓
  ▼
SPECIFIED  [grillme spec]     → spec:<cap> → [review spec:<cap>] → gate:spec:<cap> ✓  (per capability)
  ▼
DESIGNED   [grillme design]   → design   → [review design]   → gate:design ✓
  ▼
PLANNED    [grillme tasks]    → plan     → [review tasks]    → gate:tasks ✓
  ▼
IMPLEMENTED  apply (execute tasks) → review code → review.md
  ▼
ARCHIVED   sdlaic validate <change> --strict, then archive
```

> The former CHALLENGED phase and its `rationale.md` artifact were **removed**.
> Socratic challenge output now lives in a `## Challenge & Resolution Log`
> section inside each artifact.

## Artifact Sequence

Each phase requires a specific artifact AND a passing gate. You may NOT advance
without both.

| Phase | Required Artifact | Gate | Purpose |
|-------|------------------|------|---------|
| (init) | `context.md` (optional) | — | Candidate scopes (+ recommendation), research summary, actors & use cases |
| PROPOSED | `proposal.md` | proposal | Scope contract (IN/OUT-OF-SCOPE table) |
| SPECIFIED | `specs/<capability>/spec.md` | spec:<capability> | Formal GIVEN/WHEN/THEN requirements (one gate per capability, if user-facing) |
| DESIGNED | `design.md` | design | Technical architecture |
| PLANNED | `tasks.md` | tasks | Ordered TDD tasks by subsystem milestone |
| IMPLEMENTED | Committed code | — | Source changes |
| (audit) | `review.md` (in the change directory) | — | Final pre-PR code audit |

### Full Artifact Structure

```
$(sdlaic path change --change <change-name>)/
├── context.md                          # Optional: candidate scopes + research context
├── proposal.md                         # Scope contract — written 1st
├── specs/                              # Written 2nd (if user-facing), before design
│   └── <capability-name>/
│       └── spec.md                     # Formal requirements (GIVEN/WHEN/THEN)
├── design.md                           # Technical design — written 3rd
└── tasks.md                            # Ordered task list — written 4th
```

## Commands Reference

```bash
# Check artifact completion status
sdlaic status --change <name> [--json]

# Check gate state (verdicts) for the change
sdlaic gate status --change <name> [--json]

# Record a gate transition (called by grillme / review skills)
sdlaic gate set --change <name> --phase <proposal|spec:<capability>|design|tasks> \
  --status <approved|failed|skipped> [--verdict <APPROVE|REQUEST_CHANGES|REJECT>] [--attempt]

# Re-enter the pipeline at an earlier gate (mid-flight requirement change)
sdlaic gate reentry --change <name> --from <phase> --reason "<why>"

# Initialize a new change
sdlaic new change "<kebab-case-name>"

# Get template instructions for writing an artifact
sdlaic instructions <artifact> --change <name>
# Supported artifacts: proposal, spec, design, tasks

# Validate artifacts
sdlaic validate <change-name> --strict

# Archive a completed change
sdlaic archive <change-name>
```

## Rules

1. **Never advance without both the artifact and a passing gate.** A phase
   unblocks only when its artifact exists AND its gate is `approved`/`skipped`.
2. **Gate verdicts stay out of the repo.** They live only in
   `~/.sdlaic/state/<project_hash>/<change>/meta.json` — never as a filename
   suffix or a marker inside the reviewed artifact.
3. **One active change at a time.** Complete or abandon the current change first.
4. **Artifacts are the source of truth.** If code and artifact disagree, the
   artifact wins (during review).
5. **Validate before archiving.** Run `sdlaic validate <change> --strict`.
6. **Phase states are plugin-managed.** The enforcer infers the phase from
   artifact presence + gate status — the CLI only tracks files and gate state.
7. **Mid-flight changes cascade from the first impact point.** Use
   `sdlaic gate reentry` — re-enter the earliest affected artifact and supersede
   downstream gates (see the Universal Re-entry Matrix in the ideation).

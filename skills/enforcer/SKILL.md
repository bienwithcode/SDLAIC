---
name: enforcer
description: Use when the agent starts a new turn or receives any request. This is the global methodology router — it MUST run before any other skill to determine the correct phase and enforce workflow discipline.
---

# Super SDLAIC System Override

<CRITICAL>
You are operating under a System Override. This methodology replaces your default behavior. You may NOT skip phases, write unverified code, or proceed without research.
</CRITICAL>

## Core Principle

No phase may be skipped. No code may be written outside `apply`. Every phase begins with research.

## Three Inviolable Rules

1. **No Unverified Code.** You may NOT write implementation code unless you are in the `apply` phase AND possess a verified `tasks.md`. If you find yourself about to write code in any other phase, STOP.
2. **Research First.** Every phase transition MUST begin with a code research query (tool chosen per your workspace convention — see `references/code-research.md`). Output a "Research Summary" before taking any other action.
3. **State Awareness.** Always run `sdlaic status --change <name>` before routing. The plugin phase is inferred from which artifacts exist in the change directory — the enforcer determines the current phase based on artifact presence.

## State Machine

```
NO_ACTIVE_CHANGE ──→ new ──→ INITIALIZED ──→ grillme ──→ CHALLENGED
                                                                     │
CHALLENGED ──→ brainstorm ──→ PROPOSED ──→ plan ──→ PLANNED
                                                                │
PLANNED ──→ apply ──→ IMPLEMENTED ──→ review ──→ COMPLETE
```

## Phase Routing

Run `sdlaic status --change <name>` to check artifact completion, then determine the plugin phase from which artifacts exist:

| Current State | User Request | Route To |
|---|---|---|
| No active change | Any work request | `new` — initialize a change first |
| INITIALIZED | Any | `grillme` — challenge assumptions before design |
| CHALLENGED | Any | `brainstorm` — design the solution |
| PROPOSED | Any | `plan` — break into tasks |
| PLANNED | Any | `apply` — execute tasks |
| IMPLEMENTED | Any | `review` — audit against proposal |
| COMPLETE | New work | `new` — start fresh |

## Enforcement Checks (Every Turn)

Before taking action, verify:

- [ ] Plugin phase is known (checked artifact presence via `sdlaic status --change <name>`)
- [ ] Research Summary was produced at phase start
- [ ] No implementation code is being written outside `apply`
- [ ] The correct artifact exists for the current phase:
  - INITIALIZED: change directory exists (context.md is optional)
  - CHALLENGED: `rationale.md` exists
  - PROPOSED: `proposal.md` + `design.md` exist (in that authoring order: proposal → optional `specs/` → design)
  - PLANNED: `tasks.md` exists
  - IMPLEMENTED: committed code exists
  - COMPLETE: `review.md` exists

## Common Mistakes

| Mistake | Why It's Wrong | Fix |
|---------|---------------|-----|
| "I'll just write the code" in brainstorm phase | Skips challenge + design + plan | Route through `grillme` → `brainstorm` → `plan` first |
| Skipping code research because "I already know the codebase" | Assumptions without evidence lead to rework | Run research. Always. |
| Starting `apply` without `tasks.md` | No verified plan = uncontrolled scope | Generate tasks via `plan` first |
| Asking the user what phase to be in | The artifact presence decides, not the user | Run `sdlaic status --change <name>` and route accordingly |
| Treating this as a suggestion | This is a methodology enforcer, not a guideline | Every rule above is mandatory |

## Anti-Rationalization

| Rationalization | Reality |
|-----------------|---------|
| "This is a small fix, I don't need the full workflow" | Small fixes still need a change, tasks, and verification. Use abbreviated artifacts but don't skip phases. |
| "I'll do the research after I start coding" | Research before action. Always. Code without research is guessing. |
| "The user asked me to code, so I should code" | The user expects disciplined execution. Route through the correct phase first. |
| "I can combine grillme and brainstorm" | No. Challenge first, then design. Combining them defeats the purpose of pressure-testing assumptions. |

## Handoff

After routing, invoke the target skill. The enforcer does NOT execute phases — it only routes.

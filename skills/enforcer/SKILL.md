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
3. **State Awareness.** Before routing, read BOTH sources of truth: `sdlaic status --change <name>` (which artifacts exist) AND `sdlaic gate status --change <name> --json` (whether each phase's gate has passed). A phase is **unblocked only when its artifact exists AND its gate is `approved` or `skipped`.**

## Pipeline (Gated Micro-Loops)

Each phase is a micro-loop: **grill (strict only) → draft → review → gate**. The
draft skill is mandatory; grill/review are toggled by workflow level (`strict` =
both on; `light`/`free` = draft-only, gates auto-`skipped`).

```
NO_ACTIVE_CHANGE ──→ new ──→ context.md
   │
   ▼   [grill proposal] → proposal → [review proposal] → gate:proposal ✓
   ▼   [grill spec]     → spec     → [review spec]     → gate:spec ✓
   ▼   [grill design]   → design   → [review design]   → gate:design ✓
   ▼   [grill tasks]    → plan     → [review tasks]    → gate:tasks ✓
   ▼   apply ──→ code ──→ review code ──→ review.md
```

## Phase Routing

Read artifact presence (`sdlaic status`) AND gate status (`sdlaic gate status --json`), then route on the pair `(highest artifact present, its gate status)`:

| Artifact state | Gate of current phase | Route To |
|---|---|---|
| No active change | — | `new` — initialize a change first |
| context only (or gate not passed for proposal) | proposal not `approved`/`skipped` | `grillme proposal` → `proposal` → `review proposal` |
| proposal approved | spec not passed | `grillme spec` → `spec` → `review spec` |
| spec approved (or skipped as non-user-facing) | design not passed | `grillme design` → `design` → `review design` |
| design approved | tasks not passed | `grillme tasks` → `plan` → `review tasks` |
| tasks approved | — | `apply` — execute tasks |
| code implemented | — | `review code` — terminal pre-PR audit |
| review complete | new work | `new` — start fresh |

**Between phases (context isolation):** once a gate is `approved`, instruct the
user/agent to run `/clear` before starting the next phase. The next phase reads
the approved artifacts from disk with clean context. Within a phase, run `grill`
and `review` in clean-context subagents.

## Enforcement Checks (Every Turn)

Before taking action, verify:

- [ ] Both artifact presence (`sdlaic status`) and gate status (`sdlaic gate status --json`) were checked
- [ ] Research Summary was produced at phase start
- [ ] No implementation code is being written outside `apply`
- [ ] The phase is not advanced past a gate that is not `approved`/`skipped`
- [ ] The correct artifact exists for the current phase:
  - proposal phase: `proposal.md` exists (IN/OUT-OF-SCOPE table + Challenge & Resolution Log)
  - spec phase: `specs/<capability>/spec.md` exists
  - design phase: `design.md` exists
  - tasks phase: `tasks.md` exists
  - apply: committed code exists
  - code review: `review.md` recorded

## Common Mistakes

| Mistake | Why It's Wrong | Fix |
|---------|---------------|-----|
| "I'll just write the code" during a design phase | Skips the gated proposal → spec → design → tasks loops | Route through the phase micro-loops first |
| Advancing to the next phase while the gate is `failed`/`pending` | Unverified artifact poisons everything downstream | A phase unblocks only when its artifact exists AND its gate is `approved`/`skipped` |
| Using the deprecated `brainstorm` skill | It bundled proposal+spec+design (generative overreach) | Use the 1:1 skills: `proposal` → `spec` → `design` |
| Skipping code research because "I already know the codebase" | Assumptions without evidence lead to rework | Run research. Always. |
| Starting `apply` without an `approved` tasks gate | No verified plan = uncontrolled scope | Generate `tasks.md` via `plan` and pass the tasks gate first |
| Asking the user what phase to be in | Artifact presence + gate status decide, not the user | Run `sdlaic status` and `sdlaic gate status --json`, then route |
| Treating this as a suggestion | This is a methodology enforcer, not a guideline | Every rule above is mandatory |

## Anti-Rationalization

| Rationalization | Reality |
|-----------------|---------|
| "This is a small fix, I don't need the full workflow" | Small fixes still need a change, tasks, and verification. Use abbreviated artifacts but don't skip phases. |
| "I'll do the research after I start coding" | Research before action. Always. Code without research is guessing. |
| "The user asked me to code, so I should code" | The user expects disciplined execution. Route through the correct phase first. |
| "I can combine the proposal and design phases" | No. Each artifact has its own gate. Combining them recreates the generative overreach that `brainstorm` was deprecated for. |
| "The gate is just paperwork, I'll advance anyway" | The gate IS the control. Advancing past a non-passing gate is the exact failure this system prevents. |

## Handoff

After routing, invoke the target skill. The enforcer does NOT execute phases — it only routes.

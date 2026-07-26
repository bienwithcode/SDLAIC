---
name: grillme
description: Parameterized Socratic challenge engine. Runs BEFORE drafting an artifact (proposal | spec | design | tasks), loading the matching references/grills/<phase>-grill.md and challenging one question at a time. Records agreed resolutions in the target artifact's Challenge & Resolution Log. Toggled by workflow level (strict = on).
---

# Socratic Challenge Engine (Grill Me)

## Core Principle

Bad assumptions are the #1 source of rework. Grill the idea before code breaks
the system. This is a **parameterized engine**: it runs once per phase, *before*
that phase's draft skill, loading the phase-specific challenge sequence from a
modular grill checklist. It challenges — it does NOT design or draft.

## Parameters

Invoked as `grillme <phase>` where `<phase>` ∈ `proposal | spec | design | tasks`:

| Phase | Grill checklist | Runs before | Target artifact |
|-------|-----------------|-------------|-----------------|
| `proposal` | `references/grills/scope-grill.md` | `skills/proposal` | `proposal.md` |
| `spec` | `references/grills/spec-grill.md` | `skills/spec` | `specs/<capability>/spec.md` |
| `design` | `references/grills/design-grill.md` | `skills/design` | `design.md` |
| `tasks` | `references/grills/plan-grill.md` | `skills/plan` | `tasks.md` |

## Workflow toggle

- `strict` — grill is **mandatory** before the draft skill runs.
- `light` / `free` — grill is **skipped** (draft-only fast path). The enforcer
  marks the gate `skipped`; this skill does not run.

## Context isolation

Run the grill in a **clean-context subagent** with a fresh conversation, passing
only: the phase's grill checklist, the upstream approved artifact(s), and the
change's `context.md`. The subagent self-terminates on completion, returning the
Challenge & Resolution Log rows to the main session. This keeps main-session
context clean across large PRs.

## Process

### Step 1: Load the phase grill

Read `references/grills/<phase>-grill.md`. It defines the auditor stance, the
ordered challenge sequence, the stop condition, and the red flags for this phase.
Read the upstream artifact(s) the phase builds on (e.g. `spec` grills read the
approved `proposal.md`).

### Step 2: Challenge one question at a time

Work the grill checklist's challenge sequence in order. For each question:

1. **Research first.** Run one targeted codebase or web query to ground the
   question in evidence (a `path:line` precedent or `web:<URL>`), or record "no
   precedent". Tool selection follows your workspace convention (see
   `references/code-research.md`).
2. **Ask a single, specific question.** It must name a concrete scenario, be
   evidence-anchored, force a short decision, and target a real risk. Generic
   "any concerns?" questions are filler — drop them.
3. **Present 3 grounded options + a chat option.** Mark the research-grounded
   choice as Recommended.
4. **Process the answer.** For free-form answers, paraphrase back and confirm
   (max 1 retry); if still ambiguous, record `UNRESOLVED`.

Never combine questions. Never proceed without resolution or an explicit
`UNRESOLVED` mark. Stop when the grill checklist's **stop condition** is met.

### Step 3: Record resolutions in the artifact's Challenge & Resolution Log

Do NOT write a separate `rationale.md` — that artifact is removed. Instead, the
draft skill records the agreed resolutions in the target artifact's
`## Challenge & Resolution Log` section:

```markdown
## Challenge & Resolution Log
| Challenge | Resolution |
|-----------|------------|
| [question raised during grill] | [agreed answer, or "UNRESOLVED — <reason>"] |
```

Return these rows to the main session so the draft skill can embed them. Any
`UNRESOLVED` row must be surfaced to the user before the artifact is approved.

## Output Artifacts

- Challenge & Resolution Log rows embedded in the target artifact (no standalone file).

## Verification

- [ ] The correct `references/grills/<phase>-grill.md` was loaded for the phase.
- [ ] Each question was asked one at a time, research-grounded, and specific (no filler).
- [ ] The grill checklist's stop condition was reached (or the user wrapped up).
- [ ] Every free-form answer was paraphrase-confirmed or marked `UNRESOLVED`.
- [ ] Resolutions are captured as Challenge & Resolution Log rows for the artifact.
- [ ] No design/draft content was produced here (grill challenges, it does not draft).

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Loading the wrong or no grill checklist | Load `references/grills/<phase>-grill.md` for the exact phase. |
| Asking soft/open questions | Ask hard, specific, evidence-anchored questions with failure context. |
| Asking all questions at once | One at a time. Always. |
| Writing a rationale.md | Removed. Resolutions go in the artifact's Challenge & Resolution Log. |
| Designing or drafting during the grill | Grill challenges only. Park ideas for the draft skill. |
| Running in the main context on a large PR | Run in a clean-context subagent; return only the log rows. |
| Grilling in light/free mode | Skipped in light/free — the gate is marked `skipped`. |

## Handoff

Return the Challenge & Resolution Log rows, then route to the phase's draft skill
(`proposal` | `spec` | `design` | `plan`). After drafting, the artifact goes to
`skills/review` with the matching `references/reviews/<phase>-audit.md`.

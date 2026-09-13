---
name: grillme
description: Parameterized Socratic challenge engine. Runs BEFORE drafting an artifact (proposal | spec | design | tasks), loading the matching references/grills/<phase>-grill.md and challenging one question at a time. Hands agreed resolutions back for the draft skill to apply into the artifact's content sections. Toggled by workflow level (strict = on).
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

> **Per-capability (`spec` phase):** when the change has multiple `specs/<capability>/`
> directories, the grill runs once per capability, and the downstream review
> records one `spec:<capability>` gate each.

## Workflow toggle

- `strict` — grill is **mandatory** before the draft skill runs.
- `light` / `free` — grill is **skipped** (draft-only fast path). The enforcer
  marks the gate `skipped`; this skill does not run.

## Context isolation

Run the grill in a **clean-context subagent** with a fresh conversation, passing
only: the phase's grill checklist, the upstream approved artifact(s), and the
change's `context.md`. The subagent self-terminates on completion, returning the
resolutions to the main session. This keeps main-session
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

### Step 3: Hand resolutions back for the draft skill to apply

Do NOT write a separate log or `rationale.md` — that artifact was removed, and
artifacts carry **contract content only**. Return the agreed resolutions to the
main session; the phase's draft skill applies each one directly into the
artifact's own content sections:

| Resolution kind | Applied as |
|-----------------|------------|
| Scope decision ("X is out") | `proposal.md` Scope table row (OUT OF SCOPE) or Non-Goal |
| Behavior / edge case | `specs/<capability>/spec.md` requirement or scenario |
| Technical decision | `design.md` `## Decisions` row (or Risks / Open Questions) |
| Sizing / sequencing constraint | `tasks.md` milestone note |

Return the resolutions to the main session so the draft skill can apply them.
Any `UNRESOLVED` (or `DEFERRED`) item must be surfaced to the user before the
artifact is approved.

## Output Artifacts

- Resolutions returned to the main session, to be applied by the phase's draft skill into the target artifact's content sections (no standalone log file, no log section).

## Verification

- [ ] The correct `references/grills/<phase>-grill.md` was loaded for the phase.
- [ ] Each question was asked one at a time, research-grounded, and specific (no filler).
- [ ] The grill checklist's stop condition was reached (or the user wrapped up).
- [ ] Every free-form answer was paraphrase-confirmed or marked `UNRESOLVED`.
- [ ] Every resolution is actionable in the target artifact's content (scope row, requirement/scenario, decision row, or milestone note).
- [ ] No design/draft content was produced here (grill challenges, it does not draft).

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Loading the wrong or no grill checklist | Load `references/grills/<phase>-grill.md` for the exact phase. |
| Asking soft/open questions | Ask hard, specific, evidence-anchored questions with failure context. |
| Asking all questions at once | One at a time. Always. |
| Writing a rationale.md or a log section | Removed. Resolutions are applied into the artifact's content sections. |
| Designing or drafting during the grill | Grill challenges only. Park ideas for the draft skill. |
| Running in the main context on a large PR | Run in a clean-context subagent; return only the resolutions. |
| Grilling in light/free mode | Skipped in light/free — the gate is marked `skipped`. |

## Handoff

Return the resolutions, then route to the phase's draft skill
(`proposal` | `spec` | `design` | `plan`). After drafting, the artifact goes to
`skills/review`, which loads its own matching audit checklist
(`references/reviews/<phase>-audit.md` bundled with the review skill).

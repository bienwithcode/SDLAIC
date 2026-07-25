---
name: proposal
description: Use to author proposal.md for a change — the Scope Proposal (Phase 1A). Runs after an optional scope grill and produces a single artifact defining business rationale and an explicit IN-SCOPE vs OUT-OF-SCOPE contract, then hands off to the review gate.
---

# Scope Proposal Phase (Proposal)

## Core Principle

The proposal is the **scope contract** — why the change is needed and, above all, an explicit boundary between what is IN scope and what is OUT of scope. It locks down "what" before any spec, design, or code exists, so nothing downstream can drift beyond the agreed boundary. This skill produces exactly one artifact: `proposal.md`. It does not design, specify behavior, or plan.

## Pre-conditions

- An active SDLAIC change exists (`context.md` may exist with ticket context — optional).
- If the workflow level is `strict`, the scope grill has run first (see Handoff / Gate).

## Process

### Step 1: Absorb the input

Read `context.md` (if present) and the originating ticket. Identify the problem, its business rationale, and the affected area. Run a codebase-research query to ground the affected-area claims in real paths — do not assume.

### Step 2: Draft the IN/OUT-OF-SCOPE boundary first

Before writing prose, enumerate the boundary. Every item a reader might assume is included must be explicitly placed IN or OUT. Ambiguity here is the primary cause of generative overreach.

### Step 3: Write `proposal.md`

Request the template guidance, then write to `.sdlaic/changes/<name>/proposal.md`:

```markdown
# Proposal: <change-name>

**Change ID**: `<change-name>`
**Ticket**: <KEY> (<summary>)

## Why
[One paragraph: the problem this change solves and its business rationale.]

## Scope

| IN SCOPE | OUT OF SCOPE |
|----------|--------------|
| [thing this change will do] | [related thing it will deliberately NOT do] |
| ... | ... |

## What Changes
- **[Bug fix / Enhancement / New feature]: [description]**
  - [Specific change]

## Success Criteria
- [ ] [Testable criterion tied to the ticket]

## Impact
- **Affected area (indicative)**: [key modules/paths]
- **Breaking changes**: [None intended / list]

## Challenge & Resolution Log
<!-- Populated from the scope grill. Each row: the challenge raised and how it was resolved. -->
| Challenge | Resolution |
|-----------|------------|
| [question raised during grill] | [agreed answer] |
```

The **Scope table** and the **Challenge & Resolution Log** are mandatory. If no grill ran (light/free), state "No grill (workflow: <level>)" in the log.

### Step 4: Hand off to the review gate

Do not advance to the spec phase yourself. Emit a summary and hand off (see Handoff).

## Output Artifacts

- `.sdlaic/changes/<name>/proposal.md` — scope contract with a mandatory IN/OUT-OF-SCOPE table and Challenge & Resolution Log.

## Verification

- [ ] `proposal.md` exists and is populated (no template placeholders).
- [ ] The Scope table has ≥1 IN-SCOPE and ≥1 OUT-OF-SCOPE row.
- [ ] Success Criteria has ≥1 testable criterion tied to the ticket.
- [ ] No technical design or task breakdown leaked into the proposal (that is design/plan).
- [ ] Challenge & Resolution Log reflects the grill (or records that none ran).

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Mixing "how" into the proposal | Proposal is scope (why/what). Keep architecture in `design.md`. |
| Vague or absent OUT-OF-SCOPE column | Every plausible-but-excluded item must be named OUT. This is the anti-overreach control. |
| Writing spec scenarios here | Behavioral GIVEN/WHEN/THEN belong in `skills/spec`. |
| Advancing to spec without the gate | The proposal gate must be `approved`/`skipped` first. |

## Handoff / Gate

1. **Grill first (strict):** the scope grill (`references/grills/scope-grill.md`) runs before drafting; record resolutions in the Challenge & Resolution Log.
2. **Review after:** hand the drafted `proposal.md` to `skills/review` with the `proposal` audit (`references/reviews/proposal-audit.md`).
3. The reviewer records the verdict via `sdlaic gate set --phase proposal --status <approved|failed> [--verdict ...]` — never inside the repo.
4. On `approved`, the enforcer advances to Phase 1B (`skills/spec`). On `failed`, re-draft here and increment the attempt.

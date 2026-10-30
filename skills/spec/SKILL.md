---
name: spec
description: Use to author specs/<capability>/spec.md for a change — the Behavioral Requirements phase (Phase 1B). Runs after an approved proposal and an optional spec grill, producing requirements stated as testable guarantees with compact example tables, then hands off to the review gate.
---

# Behavioral Requirements Phase (Spec)

## Core Principle

The spec is the **verifiable WHAT** — testable behavioral requirements derived
from an approved proposal. Its atomic unit is the **guarantee**: a single
behavioral promise the system must keep (a decision rule, an invariant, or an
error contract), stated so a tester can check it true/false. The spec pins down
**every guarantee exactly once** — edge cases and boundary values become rows
in an example table or their own guarantee, never a pile of near-identical
scenarios. A typical change carries **~10–15 guarantees total**; more than that
means the change is probably several changes. This skill produces one
directory-based artifact: `specs/<capability>/spec.md`. It does not design or
plan.

## Pre-conditions

- `proposal.md` exists and its gate is `approved` (or `skipped` in light/free).
- If workflow is `strict`, the spec grill has run first.

## Process

### Step 1: Derive requirements from the approved scope

Read `proposal.md`. For each IN-SCOPE deliverable, identify the observable
behavior a tester could verify. Ignore OUT-OF-SCOPE items entirely.

### Step 2: Pin the hard cases

Before writing anything, list the failure surface: null / malformed / boundary
inputs, race conditions, partial failures, and error handling. Then pin each
case **once**, choosing the cheapest form that holds:

- Same rule seen from many angles (e.g. every input status mapping to an
  outcome) → **one guarantee + one example table** (each case is a row).
- A genuinely distinct behavior (an invariant, a different error contract) →
  **its own guarantee**.
- Not worth specifying → **explicitly declared out of behavioral scope**.

Never multiply scenarios for one rule. If a requirement is accruing many
scenarios that differ only in input values, collapse them into a table.

### Step 3: Write `specs/<capability>/spec.md`

Request the template guidance (`instructions spec`), then write to
`$(sdlaic path change --change <name>)/specs/<capability-name>/spec.md`:

```markdown
# <capability-name> (delta)

## ADDED Requirements

### Requirement: [Verb statement of what the system MUST do]
[One paragraph, max 5 sentences.]

**Guarantees:**
- **G1** — [SHALL statement: one behavioral promise a tester can check as true/false]
- **G2** — [SHALL statement: the failure behavior of the rule, if distinct]

#### Example: [representative case for the guarantees above]
| [input] | [expected] |
|---------|------------|
| [valid value] | [outcome] |
| [malformed / boundary value] | [defined error behavior — never a crash] |

## CHANGED Requirements

### Requirement: [existing requirement being modified]
[What changed and why.]
```

Use a `#### Scenario:` block (WHEN/THEN) **only** when a flow cannot be
expressed as a table (e.g. a multi-step user journey). Every requirement needs
at least one Example table or Scenario — that is the floor, not a quota:
additional scenarios are justified by genuinely distinct behavior, never by
input variety.

Split multiple capabilities into separate `specs/<capability>/spec.md` files.

**Writing rules — size is part of the contract:**

- **One requirement = one paragraph (max 5 sentences).** If it needs more, the
  extra material is a guarantee, an example table, or a scenario — never more
  prose.
- **One guarantee, one promise.** A guarantee states ONE checkable promise.
  If it needs "and also", it is two guarantees.
- **Enumerate, don't qualify.** Precedence rules, branching, and case lists
  belong in ordered tables or enumerations, not in English universal claims
  (`only`, `never`, `always`, `every`). Tables are auditable; qualifiers drift.
  Input-class rules ("status X grants, status Y denies") belong in an example
  table — one row per input class — not one scenario per input.
- **Re-drafts shrink or convert — they do not just add.** Answer a review
  finding by converting the offending prose into a guarantee/table (or deleting
  it), never by appending a defensive sentence to the same paragraph. Grill
  resolutions are applied into the content (requirements/guarantees), never
  into a separate log.
- **Report the size delta on every re-draft** (e.g. `spec.md: 284 → 271
  lines`, guarantees: 14 → 13) so unbounded growth is visible to the operator.

### Step 4: Hand off to the review gate

Do not advance to design yourself.

## Output Artifacts

- `$(sdlaic path change --change <name>)/specs/<capability>/spec.md` — requirements
  stated as guarantees with compact example tables, one directory per capability.

## Verification

- [ ] At least one `specs/<capability>/spec.md` exists and is populated.
- [ ] Every ADDED requirement lists its **Guarantees** — each one a single promise, checkable true/false.
- [ ] Total guarantee count across the change is ≈10–15; a larger count is justified (or the change split). On a first draft, report the count in your reply to the operator (never as a meta line inside `spec.md`); on a re-draft, report it alongside the size delta.
- [ ] Every requirement has ≥1 Example table or Scenario (the floor) — and input-class variety lives as table rows, not as repeated scenarios.
- [ ] Example tables describe observable behavior, not implementation.
- [ ] Each requirement is one paragraph (max 5 sentences); enumerations and precedence rules are tables, not prose.
- [ ] On re-draft, the size delta was reported, and prose did not grow to answer findings.
- [ ] Null/malformed/boundary inputs and error handling appear as table rows, guarantees, or an explicit exclusion — each pinned exactly once.
- [ ] Requirements trace back to IN-SCOPE items in `proposal.md`; nothing OUT-OF-SCOPE appears.

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| One scenario per input value of the same rule | Collapse into one guarantee + one example table (a row per input class). |
| Guarantees that bundle several promises ("and also") | Split — one promise per guarantee. |
| Undefined failure behavior for malformed input | Add a table row or a guarantee stating the defined error behavior. |
| Describing HOW (functions, tables) | Specs describe observable behavior; architecture is `design.md`. |
| Specifying OUT-OF-SCOPE behavior | Honor the proposal boundary strictly. |
| One giant spec for many capabilities | One `specs/<capability>/spec.md` per capability. |
| Multi-paragraph requirements, or prose appended to defend a rule | Convert to guarantees/tables; one paragraph (max 5 sentences) per requirement. |
| Advancing to design without the gate | Every `spec:<capability>` gate must be `approved`/`skipped` first. |

## Handoff / Gate

1. **Grill first (strict):** the spec grill (bundled with `skills/grillme`) runs before drafting.
2. **Review after:** hand the drafted spec to `skills/review` for the `spec` audit (checklist bundled with the review skill).
3. The reviewer records the verdict **per capability** via `sdlaic gate set --phase spec:<capability> --status <approved|failed> [--verdict ...]` — one `spec:<capability>` gate per `specs/<capability>/` directory.
4. On `approved`, the enforcer advances to Phase 2 (`skills/design`). On `failed`, re-draft here.

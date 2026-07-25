---
name: spec
description: Use to author specs/<capability>/spec.md for a change — the Behavioral Requirements phase (Phase 1B). Runs after an approved proposal and an optional spec grill, producing formal GIVEN/WHEN/THEN scenarios, then hands off to the review gate.
---

# Behavioral Requirements Phase (Spec)

## Core Principle

The spec is the **verifiable WHAT** — formal, testable behavioral requirements
derived from an approved proposal. It locks down observable behavior (including
error and edge behavior) before any architecture is designed, so the design
cannot quietly change the contract. This skill produces one directory-based
artifact: `specs/<capability>/spec.md`. It does not design or plan.

## Pre-conditions

- `proposal.md` exists and its gate is `approved` (or `skipped` in light/free).
- If workflow is `strict`, the spec grill has run first.

## Process

### Step 1: Derive requirements from the approved scope

Read `proposal.md`. For each IN-SCOPE deliverable, identify the observable
behavior a tester could verify. Ignore OUT-OF-SCOPE items entirely.

### Step 2: Enumerate the hard cases

Before writing happy paths, list the failure surface: null / malformed / boundary
inputs, race conditions, partial failures, and error handling. Each must become a
scenario or be explicitly declared out of behavioral scope.

### Step 3: Write `specs/<capability>/spec.md`

Request the template guidance (`instructions spec`), then write to
`.sdlaic/changes/<name>/specs/<capability-name>/spec.md`:

```markdown
# <capability-name> (delta)

## ADDED Requirements

### Requirement: [Verb statement of what the system MUST do]
[One paragraph describing the requirement in full.]

#### Scenario: [happy path — descriptive name]
- **GIVEN** [precondition]
- **WHEN** [action]
- **THEN** [observable outcome]

#### Scenario: [malformed / boundary input]
- **GIVEN** [invalid or edge precondition]
- **WHEN** [action]
- **THEN** [defined error behavior — not a crash]

## CHANGED Requirements

### Requirement: [existing requirement being modified]
[What changed and why.]

## Challenge & Resolution Log
<!-- From the spec grill. State "No grill (workflow: <level>)" if none ran. -->
| Challenge | Resolution |
|-----------|------------|
| [edge case raised] | [scenario added / declared out of scope] |
```

Split multiple capabilities into separate `specs/<capability>/spec.md` files.

### Step 4: Hand off to the review gate

Do not advance to design yourself.

## Output Artifacts

- `.sdlaic/changes/<name>/specs/<capability>/spec.md` — formal GIVEN/WHEN/THEN
  requirements, one directory per capability.

## Verification

- [ ] At least one `specs/<capability>/spec.md` exists and is populated.
- [ ] Every ADDED requirement has ≥1 happy-path scenario AND ≥1 error/edge scenario.
- [ ] Scenarios use GIVEN/WHEN/THEN and describe observable behavior, not implementation.
- [ ] Null/malformed/boundary inputs and error handling are covered or explicitly excluded.
- [ ] Requirements trace back to IN-SCOPE items in `proposal.md`; nothing OUT-OF-SCOPE appears.
- [ ] Challenge & Resolution Log reflects the grill (or records none ran).

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Only happy-path scenarios | Every requirement needs at least one failure/edge scenario. |
| Describing HOW (functions, tables) | Specs describe observable behavior; architecture is `design.md`. |
| Specifying OUT-OF-SCOPE behavior | Honor the proposal boundary strictly. |
| One giant spec for many capabilities | One `specs/<capability>/spec.md` per capability. |
| Advancing to design without the gate | The spec gate must be `approved`/`skipped` first. |

## Handoff / Gate

1. **Grill first (strict):** the spec grill (`references/grills/spec-grill.md`) runs before drafting.
2. **Review after:** hand the drafted spec to `skills/review` with the `spec` audit (`references/reviews/spec-audit.md`).
3. The reviewer records the verdict via `sdlaic gate set --phase spec --status <approved|failed> [--verdict ...]`.
4. On `approved`, the enforcer advances to Phase 2 (`skills/design`). On `failed`, re-draft here.

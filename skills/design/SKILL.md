---
name: design
description: Use to author design.md for a change — the Technical Architecture phase (Phase 2). Runs after an approved spec and an optional design grill, producing the architecture that satisfies the proposal and spec, then hands off to the review gate.
---

# Technical Architecture Phase (Design)

## Core Principle

The design is the **HOW** — the technical architecture that satisfies the
approved proposal (scope) and spec (behavior). It is written last among the
requirement artifacts so it cannot redefine the contract. Its focus areas are
input-boundary validation, subsystem boundaries, and DRY reuse of existing
utilities. This skill produces one artifact: `design.md`. It does not plan tasks
or write code.

## Pre-conditions

- `proposal.md` and (if applicable) `specs/<capability>/spec.md` exist and their
  gates are `approved` (or `skipped`).
- If workflow is `strict`, the design grill has run first.

## Process

### Step 1: Search for existing patterns first

Before designing anything new, run codebase-research queries for similar
implementations, shared utilities, and established architecture patterns. Record
what exists and what can be reused. "No existing patterns found for <terms>" is a
valid, explicit outcome — not a skip.

### Step 2: Design against the three focus areas

- **Input Boundary Validation** — where and how raw/untrusted input is sanitized
  and parsed at the system boundary; reject-early strategy.
- **Subsystem Boundaries** — which modules own what; how boundaries are enforced;
  no cross-boundary reach-through.
- **DRY & Utility Reuse** — reuse existing utilities from Step 1; justify any new
  code that duplicates an existing capability.

### Step 3: Write `design.md`

Request the template guidance (`instructions design`), then write to
`$(sdlaic path change --change <name>)/design.md`:

```markdown
# Design: <change-name>

## Context
[System area touched, current state, how it works today.]

## Goals / Non-Goals
- Goal: [...]
- Non-Goal: [explicitly out of scope]

## Pattern Research
**Queries used**: [...]
**Findings**: [pattern — path — reuse/reference/adapt]
**Decision**: [Reuse X / Build new because Z / No patterns found for <terms>]

## Input Boundary Validation
[Where untrusted input enters; sanitization/parsing strategy; failure behavior.]

## Subsystem Boundaries
[Modules involved, ownership, how boundaries are enforced.]

## Decisions
| Area | Decision | Rationale |
|------|----------|-----------|
| [area] | [decision] | [why; cite reused utility path if any] |

## Risks / Trade-offs
- [risk]: [mitigation]

## Challenge & Resolution Log
<!-- From the design grill. State "No grill (workflow: <level>)" if none ran. -->
| Challenge | Resolution |
|-----------|------------|
| [architectural concern] | [decision] |

## Open Questions
- [unresolved, or "None blocking"]
```

### Step 4: Hand off to the review gate

Do not advance to planning yourself.

## Output Artifacts

- `$(sdlaic path change --change <name>)/design.md` — architecture satisfying proposal + spec,
  with explicit input-boundary, subsystem-boundary, and reuse sections.

## Verification

- [ ] `design.md` exists and is populated; Pattern Research is filled (queries + findings, or explicit "none found").
- [ ] Input Boundary Validation section defines where/how untrusted input is handled.
- [ ] Subsystem Boundaries section names module ownership and enforcement.
- [ ] Decisions cite reused utilities where applicable; new duplication is justified.
- [ ] Design satisfies every ADDED requirement in the spec; introduces no OUT-OF-SCOPE capability.
- [ ] Challenge & Resolution Log reflects the grill (or records none ran).

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Designing without pattern research | Always search first. Reuse > reinvent. Record proof. |
| Redefining scope/behavior in the design | Design satisfies proposal + spec; it does not change them. |
| Ignoring input boundaries | Untrusted input handling is a required section, not optional. |
| Duplicating existing utilities | Reuse from Step 1 or justify the new code explicitly. |
| Breaking out task lists here | Task decomposition is `skills/plan`. |

## Handoff / Gate

1. **Grill first (strict):** the design grill (`references/grills/design-grill.md`) runs before drafting.
2. **Review after:** hand `design.md` to `skills/review` with the `design` audit (`references/reviews/design-audit.md`).
3. The reviewer records the verdict via `sdlaic gate set --phase design --status <approved|failed> [--verdict ...]`.
4. On `approved`, the enforcer advances to Phase 3 (`skills/plan`). On `failed`, re-draft here.

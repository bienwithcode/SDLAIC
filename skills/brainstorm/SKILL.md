---
name: brainstorm
description: DEPRECATED — do not use. The monolithic design phase has been split into three phase-gated 1:1 skills. Use `proposal` (scope), then `spec` (behavior), then `design` (architecture), each behind its own grill/review gate.
---

# Brainstorm (DEPRECATED)

> **This skill is deprecated and must not be invoked.**

## Why it was removed

`brainstorm` generated three artifacts — `proposal.md`, `specs/`, and `design.md`
— in a single pass. That bundling caused **generative overreach**: technical
design decisions were made on top of unverified business requirements, and no
runtime could enforce phase-by-phase boundaries. See
`docs/.ideation/phase_gated_artifact_review_loops.md`.

## Use these instead

The design phase is now four independent, phase-gated micro-loops. Each has a
dedicated 1:1 draft skill and its own grill-before / review-after gate:

| Instead of brainstorm's… | Use skill | Artifact | Gate checklists |
|--------------------------|-----------|----------|-----------------|
| Proposal step | `proposal` | `proposal.md` | `grills/scope-grill.md`, `reviews/proposal-audit.md` |
| Delta spec step | `spec` | `specs/<capability>/spec.md` | `grills/spec-grill.md`, `reviews/spec-audit.md` |
| Design step | `design` | `design.md` | `grills/design-grill.md`, `reviews/design-audit.md` |

Task decomposition remains in `plan` (gated by `grills/plan-grill.md` /
`reviews/plan-audit.md`).

## Routing

The `enforcer` skill routes each phase through `grill → draft → review → gate` and
never routes to `brainstorm`. If you arrive here, stop and route to `proposal`
(or the correct current phase per `sdlaic status` + `sdlaic gate status`).

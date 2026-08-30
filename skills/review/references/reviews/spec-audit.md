# Spec Audit (Spec Gate)

> Loaded by `skills/review` after `skills/spec`. Auditor role: QA Lead. Runs in a
> clean-context subagent with only the `specs/<capability>/spec.md` file(s) + the
> approved `proposal.md`. Verdict recorded **per capability** via `sdlaic gate set --phase spec:<capability>`.

## Claim Verification Rule

Every finding MUST cite primary evidence: `path:line` in the spec, or a quoted
IN/OUT-OF-SCOPE row from `proposal.md`. Unevidenced findings are dropped.

## Audit checklist

### Scenario coverage (blocking)
- [ ] Every ADDED requirement has ≥1 happy-path scenario.
- [ ] Every ADDED requirement has ≥1 error/edge scenario (null/malformed/boundary).
- [ ] Scenarios use GIVEN/WHEN/THEN and describe observable behavior only.

### Boundary & failure states
- [ ] Null / malformed / oversized / wrong-type inputs have defined behavior.
- [ ] Min/max/zero/first/last boundaries are addressed.
- [ ] Race/ordering and partial-failure behavior is defined where applicable.
- [ ] Error contract (how errors surface) is specified.

### Acceptance criteria
- [ ] Each THEN is testable by an external party from the scenario alone.
- [ ] No "it works" outcomes that cannot be verified.

### Traceability & boundary
- [ ] Every requirement traces to an IN-SCOPE item in `proposal.md`.
- [ ] No behavior specified for OUT-OF-SCOPE items.

### Hygiene
- [ ] One `specs/<capability>/spec.md` per capability; no placeholders.
- [ ] Challenge & Resolution Log present (or records no grill ran).

## Verdict

- **APPROVE** — coverage complete, boundaries defined, every requirement testable.
- **REQUEST_CHANGES** — missing edge scenarios or untestable THENs; list each with evidence.
- **REJECT** — spec contradicts the proposal, or specifies out-of-scope behavior.

Map to gate status: APPROVE → `approved`; REQUEST_CHANGES | REJECT → `failed`.

## Severity guide

| Severity | Example |
|----------|---------|
| CRITICAL | Spec contradicts the approved proposal / specifies OUT-OF-SCOPE behavior. |
| HIGH | A requirement has no error/edge scenario; undefined behavior for malformed input. |
| MEDIUM | THEN not verifiable; boundary value unaddressed. |
| LOW / INFO | Wording, scenario naming, formatting. |

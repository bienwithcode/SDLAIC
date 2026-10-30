# Spec Audit (Spec Gate)

> Loaded by `skills/review` after `skills/spec`. Auditor role: QA Lead. Runs in a
> clean-context subagent with only the `specs/<capability>/spec.md` file(s) + the
> approved `proposal.md`. Verdict recorded **per capability** via `sdlaic gate set --phase spec:<capability>`.

## Claim Verification Rule

Every finding MUST cite primary evidence: `path:line` in the spec, or a quoted
IN/OUT-OF-SCOPE row from `proposal.md`. Unevidenced findings are dropped.

## Audit checklist

### Guarantee coverage (blocking)
- [ ] Every ADDED requirement lists its **Guarantees** — each a single behavioral promise, checkable true/false.
- [ ] Every requirement has ≥1 Example table or Scenario (the floor — not a quota).
- [ ] Input-class variety lives as rows in one example table — not as one scenario per input value.
- [ ] Example tables / scenarios describe observable behavior only.

### Boundary & failure states
- [ ] Null / malformed / oversized / wrong-type inputs have defined behavior (as a table row, guarantee, or explicit exclusion).
- [ ] Min/max/zero/first/last boundaries are addressed (rows, where applicable).
- [ ] Race/ordering and partial-failure behavior is defined where applicable (usually its own guarantee).
- [ ] Error contract (how errors surface) is specified.

### Acceptance criteria
- [ ] Each guarantee / expected value is testable by an external party from the requirement alone.
- [ ] No "it works" outcomes that cannot be verified.

### Traceability & boundary
- [ ] Every requirement traces to an IN-SCOPE item in `proposal.md`.
- [ ] No behavior specified for OUT-OF-SCOPE items.

### Size & form (blocking)
- [ ] Every requirement is one paragraph (max 5 sentences); longer requirements
  must be split into guarantees or converted to a table/enumeration.
- [ ] Precedence rules, branching, and case lists are tables or ordered
  enumerations — not prose with universal quantifiers (`only`, `never`,
  `always`, `every`).
- [ ] Total guarantee count across the change is ≈10–15; a larger count is either
  justified (high-stakes change) or the change should be split.

### Hygiene
- [ ] One `specs/<capability>/spec.md` per capability; no placeholders.

## Verdict

- **APPROVE** — guarantees complete and checkable, boundaries defined, size justified.
- **REQUEST_CHANGES** — untestable guarantees, unpinned failure behavior, or scenario fan-out; list each with evidence.
- **REJECT** — spec contradicts the proposal, or specifies out-of-scope behavior.

Map to gate status: APPROVE → `approved`; REQUEST_CHANGES | REJECT → `failed`.

## Severity guide

| Severity | Example |
|----------|---------|
| CRITICAL | Spec contradicts the approved proposal / specifies OUT-OF-SCOPE behavior. |
| HIGH | A requirement has no guarantees; a guarantee is not checkable; undefined behavior for malformed input; one rule exploded into many near-identical scenarios. |
| MEDIUM | THEN/expected value not verifiable; a boundary value unaddressed; oversized or enumeration-shaped requirement prose; guarantee count >15 without justification. |
| LOW / INFO | Wording, example naming, formatting. |

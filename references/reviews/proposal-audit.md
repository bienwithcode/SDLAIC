# Proposal Audit (Scope Gate)

> Loaded by `skills/review` after `skills/proposal`. Auditor role: Business
> Analyst / PM. Runs in a clean-context subagent with only `proposal.md` + the
> originating ticket. Issues a verdict recorded via
> `sdlaic gate set --phase proposal`.

## Claim Verification Rule

Every finding MUST cite primary evidence: a `path:line` reference, a ticket
field/ID, or a direct quote from `proposal.md`. Findings without evidence are
dropped.

## Audit checklist

### Scope completeness (blocking)
- [ ] IN-SCOPE table has ≥1 concrete deliverable.
- [ ] OUT-OF-SCOPE table names each plausible-but-excluded item (no empty column).
- [ ] Scope matches the ticket boundary — no unjustified growth, no missing ticket requirement.

### Rationale
- [ ] "Why" describes the problem, not the solution.
- [ ] Rationale is grounded (ticket evidence or codebase reality), not assumed.

### Success criteria
- [ ] ≥1 criterion, each testable (observable, ideally measurable).
- [ ] Criteria trace back to the stated problem.

### Impact
- [ ] Affected area is indicative and plausible against the codebase.
- [ ] Breaking-change claim is either "none" with justification or an explicit list.

### Hygiene
- [ ] No technical design / task breakdown leaked in (belongs to design/plan).
- [ ] Challenge & Resolution Log present (or records that no grill ran).
- [ ] No template placeholders remain.

## Verdict

- **APPROVE** — all blocking items pass; scope contract is tight and evidenced.
- **REQUEST_CHANGES** — fixable gaps (missing OUT rows, untestable criteria); list each with evidence.
- **REJECT** — scope contradicts the ticket, or rationale is unsupported.

Map to gate status: APPROVE → `approved`; REQUEST_CHANGES | REJECT → `failed`.

## Severity guide

| Severity | Example |
|----------|---------|
| CRITICAL | Scope contradicts the ticket / would authorize out-of-bounds work. |
| HIGH | OUT-OF-SCOPE column empty; no testable success criterion. |
| MEDIUM | Rationale restates solution; impact area vague. |
| LOW / INFO | Wording, formatting, minor traceability gaps. |

# Design Audit (Architecture Gate)

> Loaded by `skills/review` after `skills/design`. Auditor role: System Architect.
> Runs in a clean-context subagent with `design.md` + approved `proposal.md` and
> `specs/`. Verdict recorded via `sdlaic gate set --phase design`.

## Claim Verification Rule

Every finding MUST cite primary evidence: `path:line` in `design.md` or in the
codebase (for reuse/boundary claims), or a quoted spec requirement. Unevidenced
findings are dropped.

## Audit checklist

### Standards conformance (blocking)
- [ ] **Input Boundary Validation** section defines where untrusted input enters and how it is sanitized/parsed.
- [ ] **Subsystem Boundaries** section names ownership and enforcement; no cross-boundary reach-through.
- [ ] **DRY / reuse**: existing utilities are reused where they exist (Pattern Research cites paths); new duplication is justified.

### Spec satisfaction
- [ ] Every ADDED requirement in the spec maps to a design mechanism.
- [ ] No capability beyond the proposal's IN-SCOPE items is introduced.

### API / contract safety
- [ ] Public API/interface changes are backward compatible, or the break is explicitly declared and justified.

### Pattern research
- [ ] Queries listed; findings recorded (or "No existing patterns found for <terms>" stated explicitly).

### Hygiene
- [ ] No task breakdown / code leaked in (belongs to plan/apply).
- [ ] Challenge & Resolution Log present (or records no grill ran).
- [ ] No placeholders remain.

## Verdict

- **APPROVE** — standards met, spec satisfied, reuse evidenced, contracts safe.
- **REQUEST_CHANGES** — fixable gaps (missing boundary section, unjustified new code); list with evidence.
- **REJECT** — design contradicts spec/proposal, or introduces unjustified breaking changes / out-of-scope capability.

Map to gate status: APPROVE → `approved`; REQUEST_CHANGES | REJECT → `failed`.

## Severity guide

| Severity | Example |
|----------|---------|
| CRITICAL | Design fails to satisfy a spec requirement; silent breaking API change; out-of-scope capability. |
| HIGH | No input-validation boundary; module reaches across a subsystem boundary; duplicates an existing utility. |
| MEDIUM | Pattern research thin; weak boundary enforcement. |
| LOW / INFO | Wording, formatting, minor traceability. |

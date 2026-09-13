# Plan Audit (Plan Gate)

> Loaded by `skills/review` after `skills/plan`. Auditor role: Tech Lead / Scrum
> Master. Runs in a clean-context subagent with `tasks.md` + approved `proposal.md`,
> `specs/`, and `design.md`. Verdict recorded via `sdlaic gate set --phase tasks`.

## Claim Verification Rule

Every finding MUST cite primary evidence: a `tasks.md` line reference, or a quoted
success criterion / spec requirement / design subsystem. Unevidenced findings are
dropped.

## Audit checklist

### Granularity (blocking)
- [ ] No task touches more than ~5 files.
- [ ] No "implement the feature"-sized tasks; each is a single verifiable unit.

### TDD ordering (blocking)
- [ ] Every behavioral change has a `[TEST-RED:<level>]` immediately before exactly one `[IMPL]` (strict 1-1).
- [ ] Every `[TEST-RED]` carries a `:level` suffix and a concrete FAIL command; every `[IMPL]` carries a GREEN command.
- [ ] `[WIRING]`/`[VERIFY]`/`[NO-TEST]` tags used per their boundary rules (no `[NO-TEST]` where an empirical check exists).

### Subsystem milestones (blocking)
- [ ] Tasks are grouped under Subsystem Milestones matching `design.md` boundaries.
- [ ] Every milestone ends with a Milestone Integration Verification task exercising the subsystem end-to-end.

### Coverage
- [ ] Every `proposal.md` success criterion maps to ≥1 `[TEST-RED]`/GREEN pair.
- [ ] Every `specs/` requirement maps to ≥1 `[TEST-RED]` task.
- [ ] No task implements anything beyond proposal/spec/design scope.

### Hygiene
- [ ] Completion section with full test-suite command present.
- [ ] No placeholders remain.

## Verdict

- **APPROVE** — granularity, TDD ordering, milestone integration, and coverage all pass.
- **REQUEST_CHANGES** — fixable gaps (oversized task, missing integration task, coverage hole); list with evidence.
- **REJECT** — plan implements out-of-scope work, or `[IMPL]` tasks lack failing tests systemically.

Map to gate status: APPROVE → `approved`; REQUEST_CHANGES | REJECT → `failed`.

## Severity guide

| Severity | Example |
|----------|---------|
| CRITICAL | Plan includes out-of-scope work; systemic missing `[TEST-RED]` before `[IMPL]`. |
| HIGH | Task >5 files; missing Milestone Integration Verification; uncovered success criterion. |
| MEDIUM | Wrong test level; weak milestone grouping. |
| LOW / INFO | Naming, formatting, provisional commit-message wording. |

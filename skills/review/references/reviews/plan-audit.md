# Plan Audit (Plan Gate)

> Loaded by `skills/review` after `skills/plan`. Auditor role: Tech Lead / Scrum
> Master. Runs in a clean-context subagent with `tasks.md` + approved `proposal.md`,
> `specs/`, and `design.md`. Verdict recorded via `sdlaic gate set --phase tasks`.

## Claim Verification Rule

Every finding MUST cite primary evidence: a `tasks.md` line reference, or a quoted
success criterion / guarantee / design subsystem. Unevidenced findings are dropped.

## Audit checklist

### Granularity (blocking)
- [ ] No task touches more than ~5 files.
- [ ] No "implement the feature"-sized tasks; each is a single verifiable unit.

### TDD structure (blocking)
- [ ] Every `[TEST]` task carries a `:level` suffix (`:unit`/`:integration`/`:e2e`; `:feature` accepted as an alias) and its own `RED:` (→ FAIL, right reason stated) and `GREEN:` (→ PASS) commands.
- [ ] Every `[TEST:*]` task has a frozen `Spec:` — actual table rows with hardcoded expected values or exact test code. A placeholder ("write tests for X") fails the gate.
- [ ] `[WIRING]`/`[VERIFY]`/`[NO-TEST]` tags used per their boundary rules (no `[NO-TEST]` where an empirical check exists).

### Test shape (blocking)
- [ ] One test entry per guarantee — no fan-out of one rule into many entries (input classes belong in one table-driven entry).
- [ ] Business rules are proven once, at the lowest level that proves them; entry points carry wiring checks only — no re-testing of rules through every layer.
- [ ] No test inspects source code; no expected values derived from production constants (hardcoded literals only).

### Subsystem milestones (blocking)
- [ ] Tasks are grouped under Subsystem Milestones matching `design.md` boundaries.
- [ ] Every milestone ends with a Milestone Integration Verification task exercising the subsystem end-to-end.

### Coverage
- [ ] Every guarantee from `specs/` maps to exactly one `[TEST:<level>]` entry (or a justified `[WIRING]`/`[VERIFY]`).
- [ ] Every `proposal.md` success criterion maps to exactly one entry — a `[TEST:<level>]` task, or a `[WIRING]`/`[VERIFY]` where the bucket rules say so.
- [ ] No task implements anything beyond proposal/spec/design scope.

### Hygiene
- [ ] Completion section with full test-suite command + guarantee checklist + mutation spot-check present.
- [ ] No placeholders remain.

## Verdict

- **APPROVE** — granularity, TDD structure, test shape, milestone integration, and coverage all pass.
- **REQUEST_CHANGES** — fixable gaps (oversized task, placeholder Spec, fan-out, missing integration task, coverage hole); list with evidence.
- **REJECT** — plan implements out-of-scope work, or behavioral tasks systemically lack test entries.

Map to gate status: APPROVE → `approved`; REQUEST_CHANGES | REJECT → `failed`.

## Severity guide

| Severity | Example |
|----------|---------|
| CRITICAL | Plan includes out-of-scope work; systemic behavioral tasks with no test entry. |
| HIGH | Task >5 files; placeholder `Spec:`; scenario fan-out (many entries for one rule); same rule re-tested at multiple layers; a planned source-inspecting test; uncovered guarantee or success criterion; missing Milestone Integration Verification. |
| MEDIUM | Wrong test level; expected values derived from production constants; weak milestone grouping. |
| LOW / INFO | Naming, formatting, provisional commit-message wording. |

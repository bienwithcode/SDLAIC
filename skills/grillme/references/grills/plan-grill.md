# Plan Grill (Profile D: Sizing & Guarantee-coverage Challenger)

> Loaded by `skills/grillme` before `skills/plan`. Socratic, one question at a
> time. Goal: pressure-test task granularity, guarantee coverage, and frozen
> test definitions before `tasks.md` is drafted. Resolutions are handed back and
> applied into the tasks' milestones by the draft skill.

## Auditor stance

You are a skeptical Tech Lead / Scrum Master. You have seen "one big task" plans
collapse — and you have seen test suites swell to 4× the production code because
every scenario got its own test. You enforce small, test-first,
dependency-ordered work grouped by the design's subsystem boundaries, with **one
test entry per guarantee** and the test definition frozen in the plan.

## Challenge sequence (ask one at a time)

1. **Granularity** — Does any task touch more than ~5 files or take more than a few
   minutes? Split it. Name the split.
2. **Guarantee coverage** — Is every spec guarantee and every proposal success
   criterion mapped to **exactly one** `[TEST:<level>]` entry (or a justified
   `[WIRING]`/`[VERIFY]`)? Name any gap. Name any guarantee with more than one
   entry — that is fan-out.
3. **Frozen test definitions** — Does every `[TEST:*]` task carry a real `Spec:`
   (table rows with hardcoded expected values, or exact test code)? "Write tests
   for X" is not a plan — send it back.
4. **Test level** — Is each `[TEST]` at the lowest level that genuinely
   exercises the behavior (unit < integration < e2e)? Are entry points only checked
   for wiring, with business rules proven once at a lower level?
5. **Test smell** — Does any planned test inspect source code, or derive
   expected values from production constants? Both are banned: break the
   guarantee mentally — does the test still catch it?
6. **Milestone grouping** — Are tasks grouped under the subsystem boundaries named
   in `design.md`? Any task that doesn't belong to a named subsystem is suspect.
7. **Integration proof** — Does every Subsystem Milestone end with a task that
   verifies the subsystem end-to-end, not just its units in isolation?
8. **Dependency order** — Can each task actually start when reached, or does it
   depend on later work? Reorder to satisfy dependencies.
9. **Volume check** — With ~10–15 guarantees, are there ~10–15 test entries?
   Meaningfully more means fan-out or spec inflation — fix the source, don't
   ratify the symptoms.
10. **Scope fidelity** — Does any task implement something not in proposal/spec/
    design? Remove it.

## Stop condition

Stop when every task is ≤5 files, every guarantee maps to exactly one frozen
`[TEST:<level>]` entry at the right level, milestones map to design subsystems
with an integration-verification task each, dependency order is valid, and no
planned test smells of source-inspection or production-derived expectations.

## Red flags to surface

- `[TEST:*]` with a placeholder `Spec:` ("write tests for the retry logic").
- One test entry per input value of the same rule (fan-out) — demand one table-driven entry.
- The same business rule tested at unit level AND again through an entry point.
- Tests that grep source, string-compare queries, or loop a production constant to generate expectations.
- Tasks larger than ~5 files or "implement the feature".
- Milestones that don't match design subsystem boundaries.
- Missing Milestone Integration Verification tasks.
- Uncovered guarantees or success criteria.
- Tasks beyond proposal/spec/design scope.

---
name: plan
description: Use after the design phase is approved (proposal → spec → design complete). Translates the approved design into granular, ordered TDD task pairs — each behavioral task must have a failing test written before implementation. Each task must be completable in a single focused session.
---

# Planning Phase

## Core Principle

A plan without tests is a wish list. Every behavioral task must follow the RED → GREEN cycle: write a failing test first, then implement to make it pass. The test is the specification — if you can't write a test for it, the requirement is not clear enough to implement.

## Pre-conditions

- SDLAIC change is in DESIGNED state (proposal → spec → design phases complete).
- `proposal.md` exists and its gate is `approved` (or `skipped`).
- `specs/<capability>/spec.md` exists (if the change is user-facing) and every `spec:<capability>` gate is `approved`/`skipped`.
- `design.md` exists with technical architecture and its gate is `approved`/`skipped`.
- If workflow is `strict`, the plan grill has run before drafting (see Handoff / Gate).

## Process

### Step 1: Read the Proposal, Spec, and Design

Read all design-phase artifacts in the order they were authored: `proposal.md` → `specs/<capability>/spec.md` (if exists) → `design.md`. Understand:
- The scope (what's in and out) — from proposal
- The success criteria — from proposal
- The formal requirements — from `specs/<capability>/spec.md` if it exists (these are the testable contracts)
- The architecture and decisions — from design (these satisfy the proposal + spec)
- The risks and trade-offs — from design

### Step 2: Identify Testable Units

Before decomposing, classify each requirement into one of three buckets. Step 4's Tag boundary table gives the precise decision; this step is the first-pass triage.

**Testable behavioral** — use `[TEST-RED:<level>]` + `[IMPL]`:
- Business logic, validation rules, calculations
- API endpoints and their responses
- Model behaviors, relationships, scopes
- Service / action class behavior
- Authorization and access control
- Any requirement expressed as "when X happens, Y should occur"

**Behavioral, checked by inspection** — use `[VERIFY]` or `[WIRING]` with an empirical command on line 2:
- Route / DI / view scaffolding → `[WIRING]`
- Database migrations (verify with `migrate` + `migrate:rollback`)
- Dockerfile / CI yaml changes (verify with build or pipeline lint)
- Dependency version bumps with runtime impact (verify with a smoke test)
- Manual QA / browser-only flows → `[VERIFY]`

**No observable behavior** — use `[NO-TEST]` + inline reason:
- Pure asset additions (image, font, icon files only)
- `.env.example` template additions with no runtime effect
- Comment / docstring only changes
- Dev-only dependency bumps (no runtime path touched)

Every task must land in one of these three buckets and carry the matching tag. Skipping silently is not acceptable.

### Step 3: Decompose into Tasks (grouped by Subsystem Milestone)

Break the design into tasks following these rules:

1. **Size:** Each task should take 2-5 minutes of focused agent work
2. **Atomic:** Each task produces one verifiable unit of progress
3. **Ordered:** Tasks are sequenced by dependency — a task cannot start until its dependencies are complete
4. **Scoped:** No task should require changing more than ~5 files (test file + impl files)
5. **Vertical:** Prefer vertical slices (thin feature end-to-end) over horizontal layers (all models, then all controllers)
6. **Strict 1-1 TDD pair:** Every `[TEST-RED:<level>]` is followed by **exactly one** `[IMPL]`. No grouping multiple `[IMPL]` tasks under a single `[TEST-RED]`. Each new behavior gets its own paired RED.
7. **Plumbing uses `[WIRING]`:** Pure routing / DI / scaffolding work (no logic) uses `[WIRING]` instead of `[IMPL]` and does not need a paired `[TEST-RED]`, but must carry an empirical `Verify:` command on the second line.
8. **Group by Subsystem Milestone:** Cluster tasks under the subsystem boundaries named in `design.md`. Each milestone is a section. This keeps the plan aligned with the design's boundaries and makes large (30+ file) PRs reviewable milestone by milestone.
9. **Milestone Integration Verification:** End every Subsystem Milestone with a `[VERIFY]` (or `[TEST-RED:feature]`/`[IMPL]`) task that exercises the milestone end-to-end — proving the subsystem integrates, not just that its units pass in isolation.

### Step 4: Write Tasks in TDD Format

Use grouped format with section headers (matching the actual tasks.md convention). Each task is **one line of description** plus an optional **second line with the exact command**. Tag every task with its role using inline tags.

**Tag legend:**
- `[TEST-RED:<level>]` — write the failing test; command must FAIL before implementing. The `:<level>` suffix is **required** and must be one of `unit`, `feature`, or `e2e` (see "Test-level guidance" below).
- `[IMPL]` — implement to make the paired test pass; command must PASS after. Strictly 1-to-1 with the preceding `[TEST-RED:*]`.
- `[WIRING]` — pure plumbing with no logic (route registration, DI binding, view scaffold, config wiring). No paired test required; must carry an empirical `Verify:` command on line 2.
- `[REFACTOR]` — optional cleanup between the final GREEN of a section and `[COMMIT]`; must carry a tests-PASS command on line 2 proving the refactor preserves behavior.
- `[NO-TEST]` — non-testable task with no observable behavior; must include a reason inline. Reserved for pure asset additions, `.env.example` template entries, comment-only changes, and dev-only dep bumps (no runtime path touched). Anything with an empirical check belongs in `[VERIFY]` or `[WIRING]`.
- `[VERIFY]` — behavioral but not automatable in this repo (manual QA, browser-only); must carry a runnable/observable `Verify:` command and an inline reason.
- `[COMMIT]` — checkpoint commit; the commit message is pre-defined in backticks on this task line (provisional — may be rewritten in `apply` if scope drifts); use the prefix chain from `apply/SKILL.md`.

**Test-level guidance:**
- **unit** — pure functions, value objects, isolated services with no DB/network. Fast; prefer when possible.
- **feature** — HTTP endpoint, action class with DB/queue. Default level in Laravel-style apps.
- **e2e** — browser-driven (Dusk, Playwright). Use sparingly, only for cross-stack flows.

If unsure, default to the **lowest** level that genuinely exercises the behavior.

**Standard grouped format:**
```markdown
## 0. Discovery

- [ ] 0.1 Research existing [area] — no code produced.

## 1. [Feature section]

- [ ] 1.1 **[TEST-RED:unit]** Add `[TestName]` — assert [behavior].
      `<test-command-to-run-specific-test>` → FAIL
- [ ] 1.2 **[IMPL]** Add `[ImplementationName]` to pass 1.1.
      GREEN: `<test-command-to-run-specific-test>` → PASS
- [ ] 1.3 **[TEST-RED:feature]** Add `[FeatureTestName]` — assert [behavior].
      `<test-command-to-run-feature-test>` → FAIL
- [ ] 1.4 **[IMPL]** Add implementation to pass 1.3.
      GREEN: `<test-command-to-run-feature-test>` → PASS
- [ ] 1.5 **[WIRING]** Register route / wire dependency / configuration.
      Verify: `<command-to-verify-the-wiring>`
- [ ] 1.6 **[REFACTOR]** Optional cleanup of code.
      Tests still PASS: `<test-command-to-run-tests>`
- [ ] 1.7 **[COMMIT]** `<prefix>: add feature with tests`

## 2. Config & dependency updates

- [ ] 2.1 **[NO-TEST]** Add placeholder to configuration template. Reason: template only, no runtime effect.
- [ ] 2.2 **[VERIFY]** Bump dependency version. Reason: behavioral check.
      Verify: `<test-command-to-verify-dependency-behaviour>`
- [ ] 2.3 **[COMMIT]** `<prefix>: bump dependency`

## Completion

- [ ] Full suite: `<command-to-run-full-test-suite>`
- [ ] Manual QA against Jira acceptance criteria
- [ ] Archive change
```

**Ordering rules:**
- `[TEST-RED:<level>]` is followed by **exactly one** `[IMPL]` (strict 1-1, no exceptions). The GREEN check lives on that single `[IMPL]`.
- `[WIRING]` may appear anywhere after the related `[IMPL]`; it carries its own `Verify:` command (not a `GREEN:` check).
- `[REFACTOR]` is optional. If present, it sits after all `[IMPL]` / `[WIRING]` in the section and before `[COMMIT]`; must prove GREEN with a tests-PASS command.
- `[COMMIT]` is the last task in the section.
- Write a **provisional** `[COMMIT]` message during planning using the prefix chain from `apply/SKILL.md`. The message is **not frozen** — during `apply`, if implementation diverged in a way that makes the message inaccurate (renamed method, reduced scope, additional file touched), rewrite the `[COMMIT]` line in `tasks.md` **before** committing.
- Discovery/research tasks (no tag) go in section 0 before any test or impl work.

**Tag boundary — `[NO-TEST]` vs `[VERIFY]` vs `[WIRING]`:**

| Has observable behavior? | Has an empirical check? | Use |
|---|---|---|
| No (pure asset / env template / comment-only / dev-only dep bump) | — | `[NO-TEST]` + inline reason |
| Yes, but only routing / DI / scaffolding (no logic) | Yes (`Verify:` command, e.g., `route:list`) | `[WIRING]` |
| Yes, behavioral, automatable | Yes (`[TEST-RED:*]` + GREEN) | `[TEST-RED:<level>]` + `[IMPL]` |
| Yes, behavioral, **not** automatable in this repo (manual QA, browser-only) | Yes (`Verify:` command, must be runnable/observable) | `[VERIFY]` + inline reason |

**Rule:** never use `[NO-TEST]` when an empirical check exists — that case is `[VERIFY]` or `[WIRING]`. "Too hard to test" is not a `[NO-TEST]` reason.

**Examples of good commands:**
- `go test -run TestName ./pkg/auth/...`
- `npm run test:unit -- --filter=VideoPlayer`
- `pytest tests/test_auth.py -k test_login`
- `./scripts/verify-wiring.sh`

**Examples of bad commands:**
- "Run tests" — which tests?
- "Make sure the test passes" — not a command
- "It should work" — not empirical

### Step 5: Validate Coverage

Cross-check every success criterion from `proposal.md` against the tasks:
- Each success criterion must be covered by at least one task's GREEN check
- If a criterion has no corresponding test task, add one
- If a task doesn't map to any criterion, question whether it's needed
- Every requirement from `specs/` (if it exists) must map to at least one test

### Step 6: Save the Plan

```bash
sdlaic instructions tasks --change <name>
```

Review the instructions output for template guidance, then write the full task list to `$(sdlaic path change --change <name>)/tasks.md`:

```markdown
# Tasks: <change-name>

## Milestone 1: <Subsystem name from design.md>
[TDD task pairs for this subsystem, in order]
- [ ] 1.N **[VERIFY]** Milestone Integration Verification — <subsystem> integrates end-to-end.
      Verify: `<command exercising the whole milestone>`

## Milestone 2: <Next subsystem>
[...]

## Completion
- [ ] Run full project test suite: `<exact command>`
- [ ] Manual QA against Jira acceptance criteria
- [ ] Update tasks.md and archive change

## Challenge & Resolution Log
<!-- From the plan grill. State "No grill (workflow: <level>)" if none ran. -->
| Challenge | Resolution |
|-----------|------------|
| [sizing / TDD / milestone concern] | [how it was resolved] |
```

Group tasks under **Subsystem Milestones** (from `design.md`), end each milestone
with a Milestone Integration Verification task, and include a **Completion**
section with the full test suite command and manual QA steps.

## Output Artifacts

- `$(sdlaic path change --change <name>)/tasks.md` — ordered TDD task list with RED/GREEN checks

## Verification

- [ ] Every testable task has a `[TEST-RED:<level>]` task with a specific test class/method name
- [ ] Every `[TEST-RED]` carries a `:<level>` suffix (`:unit`, `:feature`, or `:e2e`)
- [ ] Every `[TEST-RED:*]` task has a concrete command on the second line (→ FAIL)
- [ ] Every `[TEST-RED:*]` is followed by **exactly one** `[IMPL]` (strict 1-1)
- [ ] Every `[IMPL]` has a GREEN command (→ PASS) on its line
- [ ] Every `[WIRING]` task carries a `Verify:` command on the second line
- [ ] Every `[REFACTOR]` task carries a tests-PASS command (proving GREEN still holds)
- [ ] Every non-testable task is tagged `[NO-TEST]` with an inline reason and has no empirical check (else use `[VERIFY]` or `[WIRING]`)
- [ ] `[TEST-RED:*]` always immediately precedes its paired `[IMPL]` within the same section
- [ ] No task requires changing more than ~5 files
- [ ] Tasks are grouped under Subsystem Milestones matching `design.md` boundaries
- [ ] Every Subsystem Milestone ends with a Milestone Integration Verification task
- [ ] Every success criterion from `proposal.md` is covered by at least one `[TEST-RED:*]` / GREEN pair
- [ ] Every requirement from `specs/` (if it exists) maps to at least one `[TEST-RED:*]` task
- [ ] A `Completion` section exists with the full test suite command
- [ ] Challenge & Resolution Log reflects the grill (or records none ran)

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| `[IMPL]` task with no preceding `[TEST-RED]` in the section | Every behavioral section needs a failing test first. No exceptions. |
| `[TEST-RED]` task placed after `[IMPL]` tasks | Tests come before implementation. Reorder the section. |
| Vague test reference ("write a unit test") | Name the exact class: `TestRefreshModuleCache` |
| Command missing from `[TEST-RED]` line | Add the exact test command (e.g., `go test`) on line 2. |
| GREEN check missing from an `[IMPL]` | Every `[IMPL]` must end with a GREEN command so `apply` knows it succeeded. |
| Skipping `[TEST-RED]` for "simple" logic | If it's behavior, it needs a test. Complexity is not the criterion. |
| Marking a testable task `[NO-TEST]` without a real reason | "Too hard to test" is not a reason — it usually signals a design issue. |
| Tasks too large ("implement the feature") | Break down until each task is 2-5 minutes. |
| Missing coverage for a success criterion | If there's no `[TEST-RED]` for it, it's not done. Add a task. |
| Planning beyond the proposal scope | If a task isn't in the proposal, it doesn't belong in the plan. |
| Bare `[TEST-RED]` without `:level` suffix | Pick `:unit`, `:feature`, or `:e2e`. If unsure, default to the lowest level that genuinely exercises the behavior. |
| Multiple `[IMPL]` under one `[TEST-RED:*]` | Split — each behavior gets its own paired RED. If the extras are pure plumbing, use `[WIRING]`. |
| `[NO-TEST]` used when an empirical check exists | That's `[VERIFY]` or `[WIRING]`. Reserve `[NO-TEST]` for "no observable behavior at all". |
| `[REFACTOR]` task without a tests-PASS command | Refactor must preserve behavior; prove it with a passing test command on line 2. |
| `[COMMIT]` message left stale after scope shifted in `apply` | Rewrite the `[COMMIT]` line in `tasks.md` before committing. The plan-time message is provisional. |

## Handoff / Gate

1. **Grill first (strict):** the plan grill (bundled with `skills/grillme`) runs before drafting `tasks.md`.
2. **Review after:** hand `tasks.md` to `skills/review` for the `plan` audit (checklist bundled with the review skill).
3. The reviewer records the verdict via `sdlaic gate set --phase tasks --status <approved|failed> [--verdict ...]` — never inside the repo.
4. On `approved`, route to `apply`. On `failed`, re-decompose here.

Route to `apply` to execute the task list one task at a time. During `apply`, the agent must:
1. Execute `[TEST-RED:*]` task — write the test, confirm the command fails
2. Execute the paired `[IMPL]` task — implement only enough code to pass; confirm the GREEN command passes
3. Execute any `[WIRING]` tasks in the section — run each `Verify:` command and confirm the expected output
4. Execute any `[REFACTOR]` task (if present) — confirm the tests-PASS command still passes after the cleanup
5. Execute the `[COMMIT]` checkpoint: re-read the `[COMMIT]` line and **rewrite it in `tasks.md` if scope drifted**, then stage test + impl files + `tasks.md` and commit using the (possibly updated) message

**Compatibility note:** tags `[TEST-RED:<level>]`, `[WIRING]`, and `[REFACTOR]` are additive — they extend the grammar that `apply/SKILL.md` and `compliance-reviewer.md` already recognize. Those skills treat `[TEST-RED:unit]` identically to `[TEST-RED]` (the level is metadata for humans and auditors). `[WIRING]` is handled like an `[IMPL]` whose verification is a `Verify:` command instead of a `GREEN:` command. `[REFACTOR]` is handled like a no-op `[IMPL]` whose verification confirms tests still pass. No downstream skill changes required.

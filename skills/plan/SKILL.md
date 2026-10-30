---
name: plan
description: Use after the design phase is approved (proposal → spec → design complete). Translates the approved design into granular, ordered tasks — one [TEST:<level>] task per guarantee, each embedding its own RED → IMPL → GREEN cycle with the test definition frozen at plan time. Each task must be completable in a single focused session.
---

# Planning Phase

## Core Principle

A plan without tests is a wish list — but a plan with a test per scenario is a
warehouse. The plan covers **every guarantee with exactly one test entry, at the
lowest level that proves it**. Test volume is decided **once, here** — the test
definition is frozen inside the task (no placeholders); `apply` executes it, and
later review rounds may only add a row to an existing test table, never a new
suite.

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
- The **guarantees** — from `specs/<capability>/spec.md` if it exists (these are the testable contracts; each becomes exactly one test entry)
- The architecture and decisions — from design (these satisfy the proposal + spec)
- The risks and trade-offs — from design

### Step 2: Identify Testable Units

Before decomposing, classify each guarantee / success criterion into one of three buckets. Step 4's Tag boundary table gives the precise decision; this step is the first-pass triage.

**Testable behavioral** — use `[TEST:<level>]`:
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
6. **One guarantee = one `[TEST:<level>]` task:** Each guarantee gets exactly one test entry, at the lowest level that proves it. If a guarantee seems to need several separate tests, it is probably several guarantees — split it in the spec, or use one table-driven test. Never fan out a rule into one test per input value.
7. **Plumbing uses `[WIRING]`:** Pure routing / DI / scaffolding work (no logic) uses `[WIRING]`; it must carry an empirical `Verify:` command on the second line.
8. **Group by Subsystem Milestone:** Cluster tasks under the subsystem boundaries named in `design.md`. Each milestone is a section. This keeps the plan aligned with the design's boundaries and makes large (30+ file) PRs reviewable milestone by milestone.
9. **Milestone Integration Verification:** End every Subsystem Milestone with a `[VERIFY]` task that exercises the milestone end-to-end — proving the subsystem integrates, not just that its units pass in isolation.

### Step 4: Write Tasks in TDD Format

Use grouped format with section headers (matching the actual tasks.md convention). Each task is **one line of description** plus **annotated lines under it**. Tag every task with its role using inline tags.

**Tag legend:**
- `[TEST:<level>]` — one guarantee, one test entry. The `:<level>` suffix is **required** and must be one of `unit`, `integration`, or `e2e` (see "Test-level ladder" below). The task embeds its own RED → IMPL → GREEN cycle on annotated lines:
  - `Spec:` — the frozen test definition (see "No placeholders" below).
  - `RED:` — the exact command that must **FAIL** before implementing, failing *for the right reason* (behavior missing — not a typo or import error).
  - `IMPL:` — one line: the minimal code to add.
  - `GREEN:` — the exact command that must **PASS** after implementing.
- `[WIRING]` — pure plumbing with no logic (route registration, DI binding, view scaffold, config wiring). No test entry; must carry an empirical `Verify:` command on line 2. Entry-point wiring is proven here — never by re-testing business logic through the entry point.
- `[REFACTOR]` — optional cleanup between the final GREEN of a section and `[COMMIT]`; must carry a tests-PASS command on line 2 proving the refactor preserves behavior.
- `[NO-TEST]` — non-testable task with no observable behavior; must include a reason inline. Reserved for pure asset additions, `.env.example` template entries, comment-only changes, and dev-only dep bumps (no runtime path touched). Anything with an empirical check belongs in `[VERIFY]` or `[WIRING]`.
- `[VERIFY]` — behavioral but not automatable in this repo (manual QA, browser-only); must carry a runnable/observable `Verify:` command and an inline reason.
- `[COMMIT]` — checkpoint commit; the commit message is pre-defined in backticks on this task line (provisional — may be rewritten in `apply` if scope drifts); use the prefix chain from `apply/SKILL.md`.

**No placeholders — the test is designed here or not at all.** Every `[TEST:*]` task's `Spec:` line contains the actual test definition:
- For decision logic: the `[input → expected]` table rows with **hardcoded expected values** (never derived from production constants), or
- The exact test code in a fenced block, or
- For non-table flows: the precise Arrange–Act–Assert in one sentence each.

"Write tests for the retry logic" is invalid — the plan that cannot name the inputs and expected outputs is not done.

**Test-shape rules (from the guarantee-first doctrine):**
- Decision logic → one table-driven test with labeled cases.
- Entry points (HTTP handlers, console commands) → wiring checks only: the right function is called, the right response shape returns. Business rules are NOT re-tested here.
- SQL-bound guarantees (`WHERE` clauses) → a few cases against a real DB.
- User journeys → at most 1–2 e2e entries per flow.
- **Banned:** tests that inspect source code (grepping method names, string-comparing queries). If a behavior cannot be observed from outside, run a subprocess or record the gap in the PR description.
- One shared fixture builder + one doubles file per area; tests clean up the DB rows they create.
- If a review round later finds a gap, the fix is **a new row in an existing table**, not a new suite.

**Test-level ladder (unit < integration < e2e):**
- **unit** — pure functions, value objects, services with doubles and no real I/O. Fast; prefer when possible.
- **integration** — one real boundary crossed in-process: DB/queue/filesystem, or an HTTP endpoint driven without a browser. This replaces the older `feature` level.
- **e2e** — multi-process or browser-driven (Playwright, Dusk, a real CLI subprocess). Use sparingly, only for cross-stack flows.

If unsure, default to the **lowest** level that genuinely exercises the behavior. `:feature` is still accepted by the validator as an alias for `:integration` so existing plans keep working, but new plans use `:integration`.

**Standard grouped format:**
```markdown
## 0. Discovery

- [ ] 0.1 Research existing [area] — no code produced.

## 1. [Feature section]

- [ ] 1.1 **[TEST:unit]** G1: cache counter never decreases — `TestCacheCounter`
      Spec: rows: (miss on empty → 0), (hit after write → 1), (second hit → 1, not 2)
      RED: `go test -run TestCacheCounter ./pkg/cache/` → FAIL (function missing)
      IMPL: add counter guards in `cache.Store`
      GREEN: `go test -run TestCacheCounter ./pkg/cache/` → PASS
- [ ] 1.2 **[TEST:integration]** G2: unknown webhook returns 500 — `WebhookNotFoundTest`
      Spec: POST /webhook with unknown id → HTTP 500; body error code `WEBHOOK_NOT_FOUND`
      RED: `php artisan test --filter WebhookNotFoundTest` → FAIL
      IMPL: resolve-or-500 branch in `WebhookController`
      GREEN: `php artisan test --filter WebhookNotFoundTest` → PASS
- [ ] 1.3 **[WIRING]** Register route / wire dependency / configuration.
      Verify: `<command-to-verify-the-wiring>`
- [ ] 1.4 **[REFACTOR]** Optional cleanup of code.
      Tests still PASS: `<test-command-to-run-tests>`
- [ ] 1.5 **[COMMIT]** `<prefix>: add feature with tests`

## 2. Config & dependency updates

- [ ] 2.1 **[NO-TEST]** Add placeholder to configuration template. Reason: template only, no runtime effect.
- [ ] 2.2 **[VERIFY]** Bump dependency version. Reason: behavioral check.
      Verify: `<test-command-to-verify-dependency-behaviour>`
- [ ] 2.3 **[COMMIT]** `<prefix>: bump dependency`

## Completion

- [ ] Full suite: `<command-to-run-full-test-suite>`
- [ ] Guarantee checklist: every guarantee maps to exactly one passing test entry
- [ ] Mutation spot-check: break 2–3 core guarantees, confirm a test fails, restore
- [ ] Archive change
```

**Ordering rules:**
- A `[TEST:<level>]` task is self-contained: its RED is executed and watched failing before its IMPL, its GREEN after. Check the checkbox only when the full cycle is done.
- `[WIRING]` may appear anywhere after the related `[TEST:*]` task; it carries its own `Verify:` command.
- `[REFACTOR]` is optional. If present, it sits after all `[TEST:*]` / `[WIRING]` tasks in the section and before `[COMMIT]`; must prove GREEN with a tests-PASS command.
- `[COMMIT]` is the last task in the section.
- Write a **provisional** `[COMMIT]` message during planning using the prefix chain from `apply/SKILL.md`. The message is **not frozen** — during `apply`, if implementation diverged in a way that makes the message inaccurate (renamed method, reduced scope, additional file touched), rewrite the `[COMMIT]` line in `tasks.md` **before** committing.
- Discovery/research tasks (no tag) go in section 0 before any test or impl work.

**Tag boundary — `[NO-TEST]` vs `[VERIFY]` vs `[WIRING]` vs `[TEST:*]`:**

| Has observable behavior? | Has an empirical check? | Use |
|---|---|---|
| No (pure asset / env template / comment-only / dev-only dep bump) | — | `[NO-TEST]` + inline reason |
| Yes, but only routing / DI / scaffolding (no logic) | Yes (`Verify:` command, e.g., `route:list`) | `[WIRING]` |
| Yes, behavioral, automatable | Yes (RED/IMPL/GREEN cycle in one task) | `[TEST:<level>]` |
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

Cross-check every guarantee against the tasks:
- Every guarantee from `specs/` and every success criterion from `proposal.md` maps to **exactly one** test entry (a `[TEST:*]` task, or a `[WIRING]`/`[VERIFY]` where the bucket rules say so).
- If a guarantee has no test entry, add **one** — do not fan out.
- If a test entry maps to no guarantee, question whether it is needed (it is either redundant, or an unnamed guarantee — name it in the spec first).
- Sanity-check the volume: a change with ~10–15 guarantees should have ~10–15 test entries. Many more means either fan-out (fix the plan) or spec inflation (fix the spec).

### Step 6: Save the Plan

```bash
sdlaic instructions tasks --change <name>
```

Review the instructions output for template guidance, then write the full task list to `$(sdlaic path change --change <name>)/tasks.md`:

```markdown
# Tasks: <change-name>

## Milestone 1: <Subsystem name from design.md>
[One [TEST:<level>] task per guarantee for this subsystem, in order]
- [ ] 1.N **[VERIFY]** Milestone Integration Verification — <subsystem> integrates end-to-end.
      Verify: `<command exercising the whole milestone>`

## Milestone 2: <Next subsystem>
[...]

## Completion
- [ ] Run full project test suite: `<exact command>`
- [ ] Guarantee checklist: every guarantee maps to exactly one passing test entry
- [ ] Mutation spot-check: break 2–3 core guarantees, confirm a test fails, restore
- [ ] Update tasks.md and archive change
```

Group tasks under **Subsystem Milestones** (from `design.md`), end each milestone
with a Milestone Integration Verification task, and include a **Completion**
section with the full test suite command and the guarantee checklist.

## Output Artifacts

- `$(sdlaic path change --change <name>)/tasks.md` — ordered task list, one `[TEST:<level>]` task per guarantee with its frozen test definition

## Verification

- [ ] Every `[TEST:*]` task names the guarantee it proves (G-id or verbatim) and a specific test name
- [ ] Every `[TEST]` carries a `:<level>` suffix (`:unit`, `:integration`, or `:e2e`; `:feature` accepted as an alias)
- [ ] Every `[TEST:*]` task has a `Spec:` line with the actual test definition (table rows with hardcoded expected values, or fenced test code) — no "write tests for X"
- [ ] Every `[TEST:*]` task has a `RED:` command (→ FAIL, right reason stated) and a `GREEN:` command (→ PASS)
- [ ] No test entry duplicates a guarantee already tested at a lower level; entry points carry wiring checks only
- [ ] No test inspects source code; expected values are hardcoded literals, not production constants
- [ ] Every `[WIRING]` task carries a `Verify:` command on the second line
- [ ] Every `[REFACTOR]` task carries a tests-PASS command (proving GREEN still holds)
- [ ] Every non-testable task is tagged `[NO-TEST]` with an inline reason and has no empirical check (else use `[VERIFY]` or `[WIRING]`)
- [ ] No task requires changing more than ~5 files
- [ ] Tasks are grouped under Subsystem Milestones matching `design.md` boundaries
- [ ] Every Subsystem Milestone ends with a Milestone Integration Verification task
- [ ] Every guarantee from `specs/` and every success criterion from `proposal.md` maps to exactly one test entry
- [ ] A `Completion` section exists with the full test suite command and the guarantee checklist

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| One test entry per input value of the same rule | One table-driven `[TEST:*]` entry with a row per input class. |
| `[TEST:*]` task with a placeholder Spec ("write tests for the retry logic") | The plan names inputs and expected outputs, or it is not done. |
| Re-testing a business rule through an entry point | The rule is proven once at unit level; the entry point gets a `[WIRING]`/wiring-check only. |
| Test inspects source (greps method names, string-compares SQL) | Test observable behavior from outside; run a subprocess if needed, or record the gap. |
| Expected values looped from production constants | Hardcode expected values in the test — a mutation that empties the constant must still fail the test. |
| Missing test entry for a guarantee | Add exactly one — the guarantee is the unit of coverage, not the scenario. |
| Vague test reference ("write a unit test") | Name the exact class: `TestRefreshModuleCache` |
| Command missing from `RED:`/`GREEN:` lines | Add the exact test command on the annotated line. |
| Skipping `[TEST]` for "simple" logic | If it's behavior, it needs a test. Complexity is not the criterion. |
| Marking a testable task `[NO-TEST]` without a real reason | "Too hard to test" is not a reason — it usually signals a design issue. |
| Tasks too large ("implement the feature") | Break down until each task is 2-5 minutes. |
| Planning beyond the proposal scope | If a task isn't in the proposal, it doesn't belong in the plan. |
| Bare `[TEST]` without `:level` suffix | Pick `:unit`, `:integration`, or `:e2e`. If unsure, default to the lowest level that genuinely exercises the behavior. |
| `[NO-TEST]` used when an empirical check exists | That's `[VERIFY]` or `[WIRING]`. Reserve `[NO-TEST]` for "no observable behavior at all". |
| `[REFACTOR]` task without a tests-PASS command | Refactor must preserve behavior; prove it with a passing test command on line 2. |
| `[COMMIT]` message left stale after scope shifted in `apply` | Rewrite the `[COMMIT]` line in `tasks.md` before committing. The plan-time message is provisional. |

## Handoff / Gate

1. **Grill first (strict):** the plan grill (bundled with `skills/grillme`) runs before drafting `tasks.md`.
2. **Review after:** hand `tasks.md` to `skills/review` for the `plan` audit (checklist bundled with the review skill).
3. The reviewer records the verdict via `sdlaic gate set --phase tasks --status <approved|failed> [--verdict ...]` — never inside the repo.
4. On `approved`, route to `apply`. On `failed`, re-decompose here.

Route to `apply` to execute the task list one task at a time. During `apply`, the agent must:
1. Execute a `[TEST:<level>]` task as one RED → IMPL → GREEN cycle: write the test from the frozen `Spec:`, run `RED:` and watch it fail for the stated reason, write the minimal `IMPL:`, run `GREEN:` and confirm PASS
2. Execute any `[WIRING]` tasks in the section — run each `Verify:` command and confirm the expected output
3. Execute any `[REFACTOR]` task (if present) — confirm the tests-PASS command still passes after the cleanup
4. Execute the `[COMMIT]` checkpoint: re-read the `[COMMIT]` line and **rewrite it in `tasks.md` if scope drifted**, then stage test + impl files + `tasks.md` and commit using the (possibly updated) message

**Compatibility note (legacy pair grammar):** older plans may carry `[TEST-RED:<level>]` + `[IMPL]` pairs. `apply/SKILL.md` and `compliance-reviewer.md` treat such a pair as one logical `[TEST:<level>]` task split across two checkboxes — the pair's RED/GREEN checks are equivalent to the embedded cycle. New plans must use the single-task `[TEST:<level>]` grammar.

---
name: apply
description: Use when tasks.md exists and is approved. Executes the task list one task at a time — each task is verified before moving to the next. No scope expansion allowed.
---

# Execution Phase (Apply)

## Core Principle

One task at a time. Verify each before moving on. No extras, no scope creep.

## Pre-conditions

- SDLAIC change is in PLANNED state
- `tasks.md` exists with ordered tasks and verification steps
- `proposal.md` exists as the design contract

## Process

### Step 1: Read the Plan

Read `$(sdlaic path change --change <name>)/tasks.md` in full. Identify:
- The first uncompleted task
- Its dependencies (must all be completed)
- Its verification command

### Step 2: Execute One Task

Implement exactly what the task describes — nothing more, nothing less.

**Rules during execution:**
- **One task only.** Do not start the next task until the current one passes verification.
- **Task tag semantics.** A section in `tasks.md` may contain these tagged tasks; apply each one differently:
  - `[TEST:<level>]` — one guarantee, one test entry. Execute as a single RED → IMPL → GREEN cycle **inside this task**: write the test from the task's frozen `Spec:` definition, run the `RED:` command and watch it fail **for the stated reason** (behavior missing — not a typo), write the minimal `IMPL:`, then run the `GREEN:` command and confirm PASS. Never write implementation before the RED was observed. Check the checkbox only when GREEN holds.
  - `[WIRING]` — pure plumbing (route / DI / scaffold). **No test entry required.** Run the `Verify:` command on line 2; it must succeed (e.g., `route:list | grep …` returns the expected line).
  - `[REFACTOR]` — optional cleanup between the final `[TEST:*]` of a section and `[COMMIT]`. Re-run the tests-PASS command on line 2 and confirm it still PASSes (prove behavior is preserved).
  - `[COMMIT]` — checkpoint commit at the end of a section. Use the commit message verbatim from the task line (rewrite it first if scope drifted — see `plan/SKILL.md` for the provisional-message convention).
  - *Legacy:* an older plan may carry `[TEST-RED:<level>]` + `[IMPL]` pairs — treat the pair as one logical `[TEST:<level>]` task: RED on the first checkbox, GREEN on the second.
- **Match existing style.** Follow the code conventions already present in the files you touch.
- **Commit-shaped units.** At the end of a section, after the final `[TEST:*]` / `[WIRING]` / `[REFACTOR]` task in that section has been verified, execute the `[COMMIT]` checkpoint task defined on the last line of the section. The commit message is pre-defined on that `[COMMIT]` task line — use it verbatim. See commit prefix rules below for how to write `[COMMIT]` messages during planning.
- **No silent scope expansion.** If you notice something that "should also be fixed" while working on a task, do NOT fix it. Add it as a comment or note for a separate change.

### Commit Prefix Rules

Use the first applicable rule:

1. **Jira ticket exists** (check `context.md` or change name for ticket ID):
   ```
   JIRA-1013: add TextNormalizer
   ```

2. **No ticket, but SDLAIC change exists** (use the change slug):
   ```
   improve-text-normalization: add TextNormalizer
   ```

3. **No ticket, no SDLAIC change** (conventional commits):
   ```
   feat(text): add TextNormalizer
   fix(parser): remove stray template fence in JSONParser
   ```

For each `[TEST:<level>]` task, the RED and GREEN checks live on the task's own annotated lines — commit test and implementation together after the GREEN check passes:
```
JIRA-1013: add TextNormalizer with tests
```

### Step 3: Verify the Task

The verification commands sit on the **annotated lines** of the task. How to read them depends on the tag:

| Tag | Annotated line | Expected outcome |
|---|---|---|
| `[TEST:<level>]` | `RED:` command | Must **FAIL** for the stated reason (behavior missing) — that's the RED proof. Write impl only after observing it. |
| `[TEST:<level>]` | `GREEN:` command | Must **PASS** — the same test now passes; run the full suite too if cheap. |
| `[WIRING]` | `Verify:` command | Must succeed (exit 0 / expected output) — no RED to compare against |
| `[REFACTOR]` | `Tests still PASS:` (or equivalent re-run) | Must **PASS** — proves the refactor didn't break behavior |

Run the command exactly as written; do not paraphrase.

**If verification passes**, continue with the next task in the section. When the final `[TEST:*]` / `[WIRING]` / `[REFACTOR]` in the section is verified, execute the `[COMMIT]` checkpoint in this exact order:
1. Check off **every** task checkbox in this section that has been verified: `[TEST:*]` tasks (full RED → IMPL → GREEN cycle done) and any `[WIRING]` / `[REFACTOR]` tasks. (Each was verified in turn before reaching this `[COMMIT]`.)
2. Stage all files changed by this section **plus** `tasks.md`:
   ```bash
   git add <impl-files> <test-files> <wiring-files> "$(sdlaic path change --change <name>)/tasks.md"
   ```
3. Commit using the message from the `[COMMIT]` task line verbatim:
   ```bash
   git commit -m "<message from [COMMIT] task>"
   ```
4. Check off the `[COMMIT]` checkbox in `tasks.md`
5. Proceed to the next section

**If verification fails:**
1. Read the error output carefully
2. Determine if the failure is in your implementation or in the verification command itself
3. Fix the implementation (not the verification, unless the verification is provably wrong)
4. Re-run verification
5. If still failing after 2 attempts, STOP and report the issue to the user

### Step 4: Scope Discipline

During implementation, you may discover:
- **New requirements:** Park them. Add a note at the bottom of `tasks.md` under "Parking Lot" for a future change.
- **Issues in existing code:** Do NOT fix them. Note them and move on. This is not a renovation.
- **Uncertainty about a task:** Re-read `proposal.md` for guidance. If still unclear, ask the user.

### Step 5: Advance State

`tasks.md` checkboxes (`[TEST:*]`, any WIRING/REFACTOR, and COMMIT) are all updated as part of the `[COMMIT]` step in Step 3. No separate update is needed after committing.

### Step 6: Completion Check

When all tasks are complete:
1. Re-read the guarantees in `specs/` and the success criteria in `proposal.md`
2. Walk them **line by line** — passing tests alone do not prove requirements met; confirm each guarantee maps to a passing test entry and each criterion holds in the code
3. Run a **mutation spot-check** on 2–3 core guarantees: deliberately break each one, confirm its test fails, restore the code. A guarantee whose break passes silently is not covered.
4. Run the full project test suite
5. Run `sdlaic status --change <name>` to confirm all artifacts are complete

## Output Artifacts

- Modified source files (as specified in tasks)
- Updated `tasks.md` with completed checkboxes
- Git commits (one per `tasks.md` section after its final task is verified, prefixed by ticket/change/conventional)

## Verification

- [ ] Every task in `tasks.md` is checked off
- [ ] Every task's verification command was run and passed (RED was observed failing for the stated reason before each IMPL)
- [ ] No files were modified that aren't listed in any task
- [ ] No scope expansion occurred (compare changes against proposal)
- [ ] Each completed task has a corresponding git commit
- [ ] Every `[WIRING]` task in completed sections has had its `Verify:` command run and observed
- [ ] Every `[REFACTOR]` task in completed sections has had its tests-PASS command re-run and confirmed PASS
- [ ] Every completed section has its `[COMMIT]` checkbox checked, with the commit staging all files changed by `[TEST:*]` / `[WIRING]` / `[REFACTOR]` tasks in that section plus `tasks.md`
- [ ] Guarantee checklist walked line by line; mutation spot-check ran on 2–3 core guarantees

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Implementing multiple tasks at once | One task at a time. Verify. Commit. Then next. |
| Skipping verification because "it looks right" | Run the command. Evidence, not opinions. |
| Writing impl before observing RED fail | The RED failure is the proof the test tests something. Watch it fail first. |
| Fixing unrelated code "while I'm here" | This is scope creep. Don't touch it. |
| Moving to the next task after a failed verification | A failed task is not complete. Fix it or escalate. |
| Adding "just one more thing" | That thing was not in the plan. Park it. |
| Modifying verification to make it pass | Verification is the standard. Fix the implementation. |
| Inventing new tests beyond the plan's frozen `Spec:` | Tests are designed at plan time. If a gap appears, park it and add a table row in a follow-up — do not silently expand the suite. |
| Demanding a test entry for a `[WIRING]` task | Wiring is pure plumbing. The `Verify:` command is the proof — no RED needed. |
| Skipping the `[REFACTOR]` tests-PASS re-run because "I only renamed things" | Refactor must prove behavior is preserved. Re-run the tests-PASS command before the `[COMMIT]`. |

## Anti-Rationalization

| Rationalization | Reality |
|-----------------|---------|
| "I'll verify all tasks at the end" | Failures compound. Verify each task immediately. |
| "This small fix doesn't need a commit" | Every completed section gets a commit at its `[COMMIT]` checkpoint. No exceptions. |
| "The tests pass, so the requirements are met" | Walk the guarantee checklist line by line and spot-check with mutations. Passing tests alone prove nothing about untested guarantees. |
| "I'll clean up the code while I'm at it" | Clean up is a separate change. Stay in scope. |
| "The user will want this extra feature" | The user approved the proposal. Deliver what was planned. |
| "The wiring is obviously correct, no need to run `Verify:`" | Run the command. `route:list \| grep …` takes 2 seconds and is the only evidence. |

## Handoff

When all tasks are complete and verified, route to `review` to audit the implementation against the proposal.

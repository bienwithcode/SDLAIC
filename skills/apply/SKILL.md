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

Read `.sdlaic/changes/<name>/tasks.md` in full. Identify:
- The first uncompleted task
- Its dependencies (must all be completed)
- Its verification command

### Step 2: Execute One Task

Implement exactly what the task describes — nothing more, nothing less.

**Rules during execution:**
- **One task only.** Do not start the next task until the current one passes verification.
- **Task tag semantics.** A section in `tasks.md` may contain these tagged tasks; apply each one differently:
  - `[TEST-RED:<level>]` — write the failing test; the command on line 2 must FAIL before the paired `[IMPL]`.
  - `[IMPL]` — implement to make the paired `[TEST-RED]` pass; the `GREEN:` command must PASS. Strictly 1-to-1 with the preceding `[TEST-RED:*]`.
  - `[WIRING]` — pure plumbing (route / DI / scaffold). **No paired `[TEST-RED]` required.** Run the `Verify:` command on line 2; it must succeed (e.g., `route:list | grep …` returns the expected line).
  - `[REFACTOR]` — optional cleanup between the final `[IMPL]` of a section and `[COMMIT]`. Re-run the tests-PASS command on line 2 and confirm it still PASSes (prove behavior is preserved).
  - `[COMMIT]` — checkpoint commit at the end of a section. Use the commit message verbatim from the task line (rewrite it first if scope drifted — see `plan/SKILL.md` for the provisional-message convention).
- **Match existing style.** Follow the code conventions already present in the files you touch.
- **Commit-shaped units.** At the end of a section, after the final `[IMPL]` / `[WIRING]` / `[REFACTOR]` task in that section has been verified, execute the `[COMMIT]` checkpoint task defined on the last line of the section. The commit message is pre-defined on that `[COMMIT]` task line — use it verbatim. See commit prefix rules below for how to write `[COMMIT]` messages during planning.
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

For TDD pairs — commit test and implementation together after the GREEN check passes, not separately:
```
JIRA-1013: add TextNormalizer with tests
```

### Step 3: Verify the Task

The verification command sits on **line 2** of the task. How to read it depends on the tag:

| Tag | Line-2 command | Expected outcome |
|---|---|---|
| `[TEST-RED:<level>]` | bare command (e.g., `go test -run TestTextNormalizer`) | Must **FAIL** — that's the RED proof |
| `[IMPL]` | prefixed `GREEN:` | Must **PASS** — same test now passes |
| `[WIRING]` | prefixed `Verify:` | Must succeed (exit 0 / expected output) — no paired RED to compare against |
| `[REFACTOR]` | prefixed `Tests still PASS:` (or equivalent re-run) | Must **PASS** — proves the refactor didn't break behavior |

Run the command exactly as written; do not paraphrase.

**If verification passes**, continue with the next task in the section. When the final `[IMPL]` / `[WIRING]` / `[REFACTOR]` in the section is verified, execute the `[COMMIT]` checkpoint in this exact order:
1. Check off **every** task checkbox in this section that has been verified: `[TEST-RED:*]`, `[IMPL]`, and any `[WIRING]` / `[REFACTOR]` tasks. (Each was verified in turn before reaching this `[COMMIT]`.)
2. Stage all files changed by this section **plus** `tasks.md`:
   ```bash
   git add <impl-files> <test-files> <wiring-files> .sdlaic/changes/<name>/tasks.md
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

`tasks.md` checkboxes (TEST-RED, IMPL, any WIRING/REFACTOR, and COMMIT) are all updated as part of the `[COMMIT]` step in Step 3. No separate update is needed after committing.

### Step 6: Completion Check

When all tasks are complete:
1. Re-read `proposal.md` success criteria
2. Confirm each criterion has a passing verification
3. Run `sdlaic status --change <name>` to confirm all artifacts are complete

## Output Artifacts

- Modified source files (as specified in tasks)
- Updated `tasks.md` with completed checkboxes
- Git commits (one per `tasks.md` section after its final task is verified, prefixed by ticket/change/conventional)

## Verification

- [ ] Every task in `tasks.md` is checked off
- [ ] Every task's verification command was run and passed
- [ ] No files were modified that aren't listed in any task
- [ ] No scope expansion occurred (compare changes against proposal)
- [ ] Each completed task has a corresponding git commit
- [ ] Every `[WIRING]` task in completed sections has had its `Verify:` command run and observed
- [ ] Every `[REFACTOR]` task in completed sections has had its tests-PASS command re-run and confirmed PASS
- [ ] Every completed section has its `[COMMIT]` checkbox checked, with the commit staging all files changed by `[TEST-RED:*]` / `[IMPL]` / `[WIRING]` / `[REFACTOR]` tasks in that section plus `tasks.md`

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Implementing multiple tasks at once | One task at a time. Verify. Commit. Then next. |
| Skipping verification because "it looks right" | Run the command. Evidence, not opinions. |
| Fixing unrelated code "while I'm here" | This is scope creep. Don't touch it. |
| Moving to the next task after a failed verification | A failed task is not complete. Fix it or escalate. |
| Adding "just one more thing" | That thing was not in the plan. Park it. |
| Modifying verification to make it pass | Verification is the standard. Fix the implementation. |
| Demanding a paired `[TEST-RED]` for a `[WIRING]` task | Wiring is pure plumbing. The `Verify:` command is the proof — no RED needed. |
| Skipping the `[REFACTOR]` tests-PASS re-run because "I only renamed things" | Refactor must prove behavior is preserved. Re-run the tests-PASS command before the `[COMMIT]`. |

## Anti-Rationalization

| Rationalization | Reality |
|-----------------|---------|
| "I'll verify all tasks at the end" | Failures compound. Verify each task immediately. |
| "This small fix doesn't need a commit" | Every completed section gets a commit at its `[COMMIT]` checkpoint. No exceptions. |
| "I'll clean up the code while I'm at it" | Clean up is a separate change. Stay in scope. |
| "The user will want this extra feature" | The user approved the proposal. Deliver what was planned. |
| "The wiring is obviously correct, no need to run `Verify:`" | Run the command. `route:list \| grep …` takes 2 seconds and is the only evidence. |

## Handoff

When all tasks are complete and verified, route to `review` to audit the implementation against the proposal.

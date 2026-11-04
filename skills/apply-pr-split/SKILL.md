---
name: apply-pr-split
description: Use after the user has confirmed a split plan produced by plan-pr-split. Builds the planned independent PR branches and the GitHub stack (gh stack) hunk by hunk from the frozen source, runs each PR's tests, proves by an equivalence check that the split reproduces the original change, writes final PR descriptions, and publishes as draft PRs only after a final go-ahead.
---

# PR Split Execution (Apply PR Split)

## Core Principle

Execution is mechanical: **the confirmed plan is the contract.** Build exactly the
PRs it describes, prove each one is green on its own, prove together they equal the
original change, then publish.

You may NOT start without a confirmed `split-plan.md`. You may NOT push, open PRs,
comment on the original PR or change tickets until the user gives the publish
go-ahead (Step 7), unless their confirmation of the plan explicitly included
publishing.

If reality contradicts the plan (a layer will not compile, a hunk depends on
something above it), stop, update the plan, tell the user what moved and why —
never quietly improvise a different split.

## Process

### Step 1: Preflight

- Read the confirmed `split-plan.md`.
- Working tree clean; note the current branch so you can return to it.
- Re-fetch. If the PR head or base moved since the plan's recorded SHAs, rebuild
  the snapshot (head + latest base). If the new diff differs from the planned one
  beyond the base drift, stop and re-plan the delta.
- Check tooling: `gh` authenticated and the `gh stack` extension available (`gh
  extension list`). If it is missing, ask whether to install it
  (`gh extension install github/gh-stack`) or fall back to plain chained branches
  (each layer's PR uses the layer below as its base).
- Read the repository's instructions for branch naming, commit message style,
  test/lint commands and any mandatory notices in PR bodies.

### Step 2: Materialise one PR

Applies to each independent PR and each stack layer. Start from the correct base
(base branch for independent PRs; the layer below for stack layers).

- **Whole-file ownership:** check the file out from the snapshot
  (`git checkout <snapshot> -- <path>`); for deleted files, remove them.
- **Hunk ownership in a mixed file:** produce the file's diff from base to
  snapshot, keep only the owned hunks, and apply that partial patch; or edit the
  file by hand to the planned state. Interactive hunk pickers are fine where the
  runtime supports them.
- **Intermediate versions:** where the plan records one, write it exactly; the
  higher layer will replace it.
- **Tests:** bring the owned test files/cases; split test files per case when the
  plan says so.
- Leave approved drops out.

Then verify **this PR alone**: run the focused tests for its scope, then the
project's full checks (formatter, static analysis, full test suite, frontend
tests/build) as the repository defines them. If a check fails because a hunk or
test is in the wrong PR, do **not** move it on your own: stop, propose the move
(which hunk, from which PR to which, and why), and continue only after the user
confirms. Then record the deviation in the plan. Never weaken or delete a test to
make a layer pass.

Commit with a rich message: goal, origin (`requested` / `review-fix` /
`pre-existing`), what was deliberately left to other PRs, and the original PR
reference.

### Step 3: Build the independent PRs

For each independent PR in the plan's merge order: branch from the base branch,
materialise (Step 2), verify, commit. Where two of them share a file, build the
later one so it applies cleanly after the earlier one merges, and record the
expected rebase.

### Step 4: Build the stack

```bash
gh stack init --base <base-or-root-branch> <layer-1-branch>
# materialise layer 1 (Step 2), verify, commit
gh stack add <layer-2-branch>
# materialise layer 2, verify, commit
# … repeat for each layer
gh stack view
```

If a lower layer needs a fix later: check it out (`gh stack down` / `gh stack
checkout <branch>`), commit the fix, then `gh stack rebase --upstack` and re-verify
the layers above.

If the stack must sit on an independent PR that has not merged, root the stack on
that branch as the plan states, and note that `gh stack sync` is needed once it
merges.

### Step 5: Equivalence check

Create a throwaway local branch from the base, merge every independent branch and
the stack top into it, and diff it against the snapshot:

```bash
git diff <snapshot> <integration-branch>
```

The only allowed differences are the approved drops. Any other difference means a
hunk was lost, duplicated or altered — fix the owning PR, re-verify it, and
re-run the check. Run the full suite once on the integration branch. Delete the
integration branch afterwards.

### Step 6: Write final PR descriptions

For each PR, finish the plan's draft using the `write-pr-description` principles:
outcome, acceptance criteria, out-of-scope with reasons ("lands in <PR>"), real
`file:line` citations from this PR's own diff, verification results. Add the stack
position and a link to the original PR. Include any notice the repository requires
(e.g. a machine-generated disclosure).

### Step 7: Publish gate

Show the user, in one summary:

- every branch, its base, commit, and test/check results;
- the equivalence check result;
- deviations from the plan and why;
- what will be pushed and opened.

Ask for the go-ahead. On approval:

```bash
# independent PRs
git push -u origin <branch>
gh pr create --draft --base <base> --head <branch> --title "<title>" --body-file <file>

# stack — pushes every layer and opens one PR per layer
gh stack submit --auto   # --auto: non-interactive, new PRs created as drafts
```

Never run `gh stack submit` without `--auto` and never pass `--open`: in the
interactive editor new PRs default to **ready for review**. `--auto` uses
auto-generated titles, so immediately set each stacked PR's title and description
from Step 6 (`gh pr edit <number> --title "<title>" --body-file <file>`).

Without `gh stack` (plain chained branches), push each layer and open it with
`gh pr create --draft --base <layer-below-branch> --head <layer-branch>`.

### Step 8: Close the loop

- Post one comment on the original PR listing the new PRs, their order and the
  merge plan (after go-ahead). Include any notice the repository requires on
  machine-authored content, as in Step 6.
- Do not close the original PR, delete its branch or change tickets unless the user
  asks; propose these as next steps.
- Update `split-plan.md` with the final PR numbers and any deviations.
- Return to the branch the user started on. Propose (do not run) removing the
  snapshot branch and its temporary worktree from `plan-pr-split` Step 1.

## Output Artifacts

- Local branches for every planned PR (and, after the go-ahead, pushed draft PRs).
- Final PR descriptions (one file per PR, next to `split-plan.md`).
- `split-plan.md` updated with PR numbers, deviations and the equivalence result.

## Verification

- [ ] Started from a confirmed plan; snapshot re-checked against the current PR head and base.
- [ ] Each PR built from its correct base and green on its own (focused + full checks).
- [ ] No test weakened or removed to make a layer pass.
- [ ] Equivalence check: integration diff vs snapshot contains only approved drops.
- [ ] Every deviation from the plan was confirmed by the user before it was made, then recorded.
- [ ] Descriptions cite only lines in their own PR's diff and link the original PR.
- [ ] Nothing pushed, opened or commented before the publish go-ahead.
- [ ] Original PR left open and tickets untouched unless the user asked otherwise.

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Re-deciding the split during execution | The plan is the contract. Stop, update the plan, report the move. |
| Building from the stale PR head | Re-check SHAs; rebuild the snapshot with the latest base before materialising. |
| A layer that only passes once the next layer lands | A hunk or test is in the wrong layer; propose the move, wait for confirmation, then re-verify. |
| Fixing a failing layer by editing the test | Find the missing dependency instead; tests encode the guarantee. |
| Skipping the equivalence check | It is the only proof nothing was lost or changed. Always run it. |
| Pushing as soon as the branches are green | Publishing is outward-facing; wait for the go-ahead. |
| Opening stacked PRs as ready for review by default | Use `gh stack submit --auto` (drafts), never `--open`; the user marks them ready. |
| Closing the original PR automatically | Leave it open and propose closing it once the new PRs are accepted. |

## Handoff

Report the published PR list in merge order, the release plan from the split plan
(merge the stack as one group, or the rollout control), and the proposed
follow-ups (ticket updates, closing the original PR).

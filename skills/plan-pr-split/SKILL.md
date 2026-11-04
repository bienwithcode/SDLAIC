---
name: plan-pr-split
description: Use when a pull request is too large or mixes several scopes and must be broken into reviewable pieces. Read-only analysis that classifies every hunk of the PR (not every file) by scope, using the PR's spec artifacts when present and codebase research when not, then produces a split plan — independent PRs to the base branch plus a dependency-ordered GitHub stack (gh stack) — with per-PR scope, hunk ownership, tests, merge order, release safety and draft PR descriptions. Stops for user confirmation; execution is done by apply-pr-split.
---

# PR Split Planning (Plan PR Split)

## Core Principle

A large PR is hard to review because **reviewers cannot tell which lines serve which
goal**. The fix is not "fewer files per PR" — it is **one coherent, independently
verifiable goal per PR**. This skill produces a plan; it changes nothing.

Three rules drive every decision:

1. **Split by hunk, not by file.** One file may feed several PRs; even one function
   may contain lines that belong to different PRs. File-level splitting is a
   shortcut that produces either broken PRs or mixed scopes.
2. **Every PR must stand alone.** Each PR compiles, passes its own tests and the full
   suite on top of its base. A PR that only works once a later PR lands is a
   mis-ordered split.
3. **Nothing is lost, nothing is added.** The union of all planned PRs must reproduce
   the original change exactly, except for drops the user explicitly approves
   (e.g. unrelated noise).

You may NOT create branches, commit, push, open PRs or edit tickets in this skill.
The single exception is the local-only snapshot of Step 1, which never leaves the
machine and never touches the user's working tree. The output is a plan the user
confirms; `apply-pr-split` executes it.

## Process

### Step 1: Load the PR and freeze the source

Collect, from the hosting service and the local clone:

- PR title, body, base branch, head branch, head SHA, mergeability (is it behind
  the base?), review decision, and **all review comments** (they explain why
  out-of-goal changes were added).
- Every linked or mentioned ticket/issue, including closed ones — read their
  requirements and done-conditions.

Freeze the source: create a **local-only** snapshot branch from the PR head, merge
the latest base into it, and record both SHAs. Do this in a separate temporary
worktree (e.g. `git worktree add <tmp-dir> -b <snapshot> <head-sha>`), so the
user's current branch and uncommitted work are never touched. All analysis runs
against `base..snapshot`, so the plan already accounts for base drift. If the merge
conflicts, stop and report — do not plan against a stale diff.

Never push the snapshot branch. Keep it (and its worktree) for `apply-pr-split`;
record both in the plan.

### Step 2: Measure the diff

Report the shape before judging it:

| Category | Files | +/− |
|---|---|---|
| Production code | | |
| Tests | | |
| Specs / docs | | |
| Generated / lockfiles / fixtures | | |

The production-code share is the real review load; tests and docs follow the code
they belong to.

### Step 3: Find the scope contract

Look for the authoritative statement of what the PR is meant to deliver, in this
order:

1. **Spec artifacts** in the diff or repo: an OpenSpec change folder
   (`openspec/changes/<id>/` — proposal, spec, design, tasks), an SDLAIC change
   directory (`sdlaic path change --change <name>`; never assume the path), or any
   requirements/ADR document the PR adds or references.
2. **The PR description's own structure** — sections such as "What changes",
   "Additional changes carried by this PR", "Out of scope".
3. **The tickets/issues** from Step 1.

**When a spec exists**, every requirement (and every declared "additional change")
is a candidate unit. Map requirements → units first; files come later.

**When no spec exists**, research the codebase before grouping. For each changed
file, establish what capability it serves, who calls it, and what depends on the
changed behaviour. Group into domains such as: data model / lifecycle, access
control / authorisation, API surface, background jobs, business features, UI,
docs. Record the evidence (file:line, caller) behind each grouping — a domain
boundary you cannot ground is a guess.

### Step 4: Classify every hunk

Walk the full diff hunk by hunk. For each hunk record:

| Field | Meaning |
|---|---|
| Unit | The requirement / domain it serves |
| Origin | `requested` (ticket/spec) · `review-fix` (added in review) · `pre-existing defect` (bug that exists on the base, fixed in passing) · `refactor` · `noise` (unrelated wording, formatting, drive-by edits) |
| Introduces | Symbols, routes, enum cases, config keys, factories it adds |
| Consumes | Symbols it depends on that this PR adds elsewhere |

Then:

- List every **mixed file** explicitly, with which hunk (or line range) goes to
  which unit.
- Where two units touch the **same statement**, give the line to the higher layer
  and record the **intermediate version** the lower layer must carry (e.g. the
  lower PR keeps `$data = $merged;`, the higher PR changes it to
  `$data = normalize($merged);`).
- For **tests**, apply the same classification per test case, not per file: a test
  case that uses a symbol introduced by a higher unit (a new enum case, a factory
  state) moves to that unit, even if the rest of its file belongs lower.
- Flag every **noise** hunk and propose dropping it.
- Mark any file you classified **by name only** (not by reading its diff). The
  plan may not be confirmed while such files remain.

### Step 5: Choose the shape

Two kinds of PR:

- **Independent PRs → base branch.** For units that stand alone: pre-existing
  defects, review-driven hardening, small all-tenant fixes. They do not depend on
  the feature and the feature does not depend on them (or only on their final
  state).
- **A stack (gh stack) for the dependent feature.** Order layers by dependency,
  typically: foundation (types, schema, lifecycle, defaults) → enforcement (access
  control, guards, routes) → features (screens, metrics, jobs) → presentation
  (navigation, copy) and final docs. Each layer's diff is reviewed against the
  layer below it.

Shaping rules:

- **Cohesion over size.** Prefer one goal per PR; as a guideline aim for a
  production diff a reviewer can hold in one sitting (roughly ≤ 400 lines), but
  never split one requirement across PRs just to hit a number.
- **Isolate compatibility risk.** A change that alters behaviour for external
  callers (API contract, auth rejection, data migration) gets its own PR so it can
  be held, staged or reverted alone.
- **Keep the spec with the foundation.** If the PR carries spec artifacts, put them
  in the first stack layer so reviewers approve the contract before the code.
- **Offer consolidation, don't force it.** If the plan exceeds the user's
  appetite, list the merges that would reduce the count and what each costs.

### Step 6: Order, conflicts and base drift

- Independent PRs that touch the **same file** need a merge order; the later one
  rebases on the earlier.
- If the stack depends on an independent PR, either land that PR first and root the
  stack on the updated base, or put it at the bottom of the stack. Say which.
- Record the expected conflict points (file + reason).

### Step 7: Release safety

Ask: if only the bottom layer(s) merged and deployed, would users see a partial or
inconsistent feature (e.g. a new type selectable before its guards exist)? If yes,
choose and state one:

- **Merge the stack as one group.** `gh stack merge <top PR>` merges a contiguous
  group from the lowest unmerged PR in a single operation; layers are still
  reviewed separately. It refuses draft PRs, so every layer in the group must be
  marked ready first.
- **Rollout control.** A server-side switch that blocks the unfinished behaviour
  until the last layer lands.

Independent PRs must each be safe to deploy alone.

### Step 8: Tracking

Propose (do not create) the ticket mapping: typically the original ticket as
parent with one sub-item per stack layer, and separate tickets for independent
fixes that have no ticket. Link closed tickets as context instead of reopening
them. Note any project rule that ties tooling (preview environments, CI) to ticket
keys or ticket types.

### Step 9: Draft each PR

For every planned PR record:

- Branch name (follow the repository's naming convention), base, stack position.
- Title (follow the repository's convention; include the ticket key where one
  applies).
- Goal (one sentence) and origin (`requested` / `review-fix` / `pre-existing`).
- Owned files and hunks (with intermediate versions where Step 4 required them).
- Tests that move with it, and the focused + full commands to run.
- Acceptance criteria subset and out-of-scope list ("lands in PR <n>").
- Risks and reviewer focus points.
- A draft description skeleton following the `write-pr-description` principles
  (outcome, AC, out-of-scope with reasons — not a per-file changelog), plus a line
  linking the original PR and the stack position.

### Step 10: Equivalence plan

State how execution will prove nothing was lost: build a throwaway integration
branch that merges every independent branch and the stack top, then diff it
against the frozen snapshot. The only allowed differences are the approved drops.

### Step 11: Write the plan and stop

Write the plan (template below) and present a short summary to the user: the PR
list, the dependency picture, the open decisions, and anything you could not
verify. Ask the user to confirm or adjust. **Stop.** Do not start execution in the
same turn.

## Output Artifacts

- `split-plan.md` — written to the SDLAIC change directory when the PR maps to an
  SDLAIC change (`sdlaic path change --change <name>`), otherwise to a working
  location outside the repository. Never commit it to the PR branch unless the user
  asks.

```markdown
# Split plan: <repo>#<PR> — <title>

Source: head <sha> · base <branch>@<sha> · snapshot <local-branch>@<sha>
Contract: <spec path | PR description | ticket(s) | codebase research>
Diff: prod <files> +<a>/−<d> · tests … · docs …

## Summary
<independent PR count> independent PR(s) + <n>-layer stack = <total> PRs.
Release: <merge stack as one group | rollout control>.

## Dependency picture
<base>
 ├─ I1 <name> (merge first)
 ├─ I2 <name>
 └─ S1 <layer> → S2 <layer> → S3 <layer>

## Independent PRs
### I1 — <title>
- Branch / base:
- Goal / origin:
- Owns: <file> (whole) · <file> (hunks: …, intermediate: …)
- Tests: <files / cases> · verify: <commands>
- AC / out-of-scope:
- Risks / reviewer focus:
- Merge order notes:

## Stack
### S1 — <title>
(same fields)

## Mixed files
| File | Hunk / lines | → PR | Note |

## Dropped (needs approval)
| File | Hunk | Why |

## Ticket mapping (proposed, not created)

## Equivalence check
<integration branch recipe and the allowed differences>

## Open decisions
1. …

## Unverified
- <external facts not checked (CI behaviour, deploy rules, ticket conventions)>
  Every changed file must be classified from its diff; a file classified by name
  only blocks confirmation and does not belong here.
```

## Verification

- [ ] Source frozen: snapshot = PR head + latest base, built in a temporary worktree, SHAs recorded, nothing pushed, user's working tree untouched.
- [ ] Contract identified (spec / description / tickets) or domains grounded by research with file:line evidence.
- [ ] Every hunk of `base..snapshot` is assigned to exactly one PR or to the approved-drop list.
- [ ] Mixed files list hunk ownership and any intermediate versions.
- [ ] Test cases that use higher-layer symbols are placed in that layer.
- [ ] Each PR is planned to compile and pass tests on its own base.
- [ ] Compatibility-risky changes are isolated.
- [ ] Merge order, conflict points and release safety are stated.
- [ ] No file remains classified by name only.
- [ ] Plan written; user asked to confirm; no branch, commit, push or ticket change made.

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Splitting by file | Split by hunk. A file can serve several PRs; record the ownership per hunk. |
| Cherry-picking whole commits | Later commits usually mix scopes. Extract from the final snapshot diff, not from commit history. |
| Planning against the PR head while it is behind the base | Merge the latest base into a local snapshot first; otherwise the plan misses drift and conflicts. |
| A lower PR whose tests use a symbol from a higher PR | Move the test case up, or move the symbol down. Each PR must be green alone. |
| Ignoring the spec | When spec artifacts exist they are the best split boundary — map requirements to units before files. |
| Grouping domains by guess when there is no spec | Research callers and dependents; record evidence per grouping. |
| Bundling a caller-breaking change with safe fixes | Isolate it so it can be held, staged or reverted alone. |
| Silently keeping noise | List unrelated wording/formatting hunks and ask to drop them. |
| Stacking fixes that don't depend on the feature | Pre-existing defects and review hardening go to the base as independent PRs. |
| Ignoring what a partial merge exposes | State release safety: merge the stack as one group or add a rollout control. |
| Executing before confirmation | This skill ends with the plan. Execution belongs to `apply-pr-split` after the user confirms. |

## Handoff

Present the plan summary and the open decisions. When the user confirms (with or
without adjustments), update `split-plan.md` to the confirmed version and invoke
`apply-pr-split` with its path.

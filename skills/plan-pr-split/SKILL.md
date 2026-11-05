---
name: plan-pr-split
description: Use when a pull request is too large or mixes several scopes and must be broken into reviewable pieces. Read-only analysis that classifies every hunk of the PR (not every file) by scope, using the PR's spec artifacts when present and codebase research when not, then produces a split plan — independent PRs to the base branch plus a dependency-ordered GitHub stack (gh stack) — with per-PR scope, hunk ownership, tests, merge order, release safety and draft PR descriptions. Stops for user confirmation; execution is done by apply-pr-split.
---

# PR Split Planning (Plan PR Split)

## Core Principle

A large PR is hard to review because reviewers cannot tell which lines serve which
goal. Split it into the **fewest PRs that each carry one coherent theme** and can be
verified alone. Four rules:

1. **Split by hunk, not by file.** One file, even one function, may feed several PRs.
2. **Every PR stands alone.** It compiles and passes its own and the full tests on its base.
3. **Nothing lost, nothing added.** All PRs together reproduce the original change,
   except drops the user approves.
4. **Fewest PRs.** Each extra PR costs a review, CI run, merge, rebase and often a
   ticket. Finer splits are options, not the plan.

This skill is read-only: you may NOT branch, commit, push, open PRs or edit
tickets. The only exception is the local snapshot of Step 1. `apply-pr-split`
executes the confirmed plan.

## Process

### Step 1: Load the PR and freeze the source

- Collect: title, body, base/head branch, head SHA, behind-base status, review
  decision, **all review comments** (they explain out-of-goal changes), and every
  linked or mentioned ticket/issue (closed ones too) with its acceptance criteria.
- Read the repository's rules now: agent/contributor instructions (`CLAUDE.md`,
  `AGENTS.md`, `CONTRIBUTING.md`) and CI config — branch naming, PR title format,
  ticket types, tooling tied to tickets or titles (preview environments, CI jobs).
  These are readable facts; never park them under "Unverified".
- Freeze: in a temporary worktree (`git worktree add <tmp> -b <snapshot> <head-sha>`),
  merge the latest base into the snapshot and record both SHAs. All analysis runs on
  `base..snapshot`. On conflict, stop and report. Never push the snapshot; keep it
  for `apply-pr-split`.
- Measure: files and +/− for production code, tests, specs/docs, generated files.
  Production code is the real review load.

### Step 2: Establish the scope contract

The contract is the **union** of: spec artifacts (OpenSpec `openspec/changes/<id>/`,
an SDLAIC change dir via `sdlaic path change --change <name>`, ADRs), the PR
description's sections ("What changes", "Additional changes", "Out of scope"),
ticket AC, and review comments. A spec can lag the ticket.

- **Spec exists:** each requirement and each declared additional change is a
  candidate unit. Map requirements → units before files.
- **No spec:** research the codebase first — for each changed file, what capability
  it serves, who calls it, what depends on it. Group into domains (lifecycle, access
  control, API, jobs, features, UI, docs) and record file:line evidence per group.

### Step 3: Classify every hunk

For each hunk record **Unit**, **Origin** (`requested` · `review-fix` ·
`pre-existing defect` · `refactor` · `noise`), **Introduces** / **Consumes**
(symbols, routes, enum cases, factory states), and **Reach** (only the feature's
subjects, or beyond them).

- **Mixed files:** list hunk → PR ownership. If two units share one statement, give
  it to the higher PR and record the lower PR's intermediate version — but prefer
  co-locating the units (Step 4).
- **Tests per case, not per file:** a case using a higher unit's symbol moves up.
- **Reach-beyond changes:** for every hunk that changes existing behaviour, search
  the other callers of the changed function, query, gate or handler **on the base
  branch**, not only in the new code — a method can look new-feature-only in the
  snapshot while older code on the base already calls it. A hunk that **wraps**
  pre-existing code (a `try/catch`, transaction, early return or condition around
  old calls) changes existing behaviour even if it looks additive. Usual suspects:
  broadened `try/catch` (and exception messages surfaced to users), a filter added
  to a shared query/model method, a changed default, middleware on a shared route
  group. Any such change the contract does not declare goes in **Undeclared
  behaviour changes** and **Open decisions**, citing the base callers (file:line).
  "No reach beyond" is a claim: back it with the base-branch search you ran.
- **Before calling anything undeclared or droppable,** check it against every
  contract source and cite the one that covers it, or state that none does.
- **Noise:** propose dropping it. When you drop one, align the kept code with the
  base's convention (e.g. message wording) and record that adjustment.
- Classify every file from its diff. Name-only classification blocks confirmation.

### Step 4: Choose the shape

- **Independent PRs → base:** units the feature doesn't need — pre-existing defects,
  review hardening. Group them by domain or blast radius, not one per finding.
- **Stack (gh stack) for the feature,** ordered by dependency: foundation (types,
  lifecycle, defaults, spec artifacts) → enforcement (access, guards, routes) →
  features (screens, metrics) → presentation and docs. Put a layer below another
  when the upper one needs it to avoid broken behaviour (e.g. a redirect target).

**Budget** (unless the user or reviewer sets one): 4–6 PRs total, ≤ 3 independent,
≤ 4 layers. A unit gets its **own** PR only if it:

1. must be held, staged or reverted alone (caller-breaking or data-changing) —
   such a PR carries **only** the risky change, so reverting it reverts nothing else;
2. has a different release timing;
3. needs a different reviewer or owner; or
4. would push its host PR past one review sitting (~400–500 production lines,
   re-indentation excluded).

Everything else folds in as a **declared change**: tiny review-fixes join the nearest
related PR; a prerequisite fix the feature needs goes into the bottom layer (no
"merge first" PR, no intermediate version). List finer splits under
"Further splits" with what each buys and costs.

### Step 5: Order and release safety

- Independent PRs sharing a file need a merge order; record conflict points.
- If the stack needs an independent PR, land it first or move it to the bottom layer.
- If merging only the lower layers would expose a partial feature, choose:
  **merge the stack as one group** (`gh stack merge <top PR>`; all layers must be
  ready, not draft) or a **rollout control**. Independent PRs must each be safe to
  deploy alone.

### Step 6: Tickets and per-PR drafts

- Propose (do not create) the ticket mapping using ticket types the repo rules
  allow (e.g. if sub-tasks lose preview environments, keep the parent key in each
  layer's title instead). Link closed tickets as context; don't reopen them.
- Per PR, fill the template fields: branch, base, title (repo conventions), goal,
  origin, own-PR reason, owned hunks, folded declared changes, tests and
  focused/full commands, AC and out-of-scope ("lands in <PR>"), risks, approximate
  production lines. Draft descriptions follow `write-pr-description` principles
  and link the original PR and stack position.
- Equivalence: an integration branch merging all independent PRs and the stack top
  must diff against the snapshot only by approved drops and recorded adjustments.

### Step 7: Write the plan and stop

Write `split-plan.md` from `references/split-plan-template.md`. Present the PR list,
dependency picture, undeclared changes, open decisions and anything unverifiable.
Ask the user to confirm or adjust. **Stop.**

## Output Artifacts

- `split-plan.md` — in the SDLAIC change directory when the PR maps to a change
  (`sdlaic path change --change <name>`), otherwise in a working location outside the
  repository. Not committed unless the user asks.

## Verification

- [ ] Snapshot = PR head + latest base, in a temp worktree; SHAs recorded; nothing pushed.
- [ ] Repo rules read and applied; contract = spec ∪ description ∪ ticket AC ∪ review comments.
- [ ] Every hunk assigned to one PR or the drop list; every file classified from its diff.
- [ ] Each PR planned to be green alone (tests moved with the symbols they use).
- [ ] Reach-beyond changes searched on the base branch (callers of changed code, wrappers around old code); undeclared ones listed in the table and Open decisions.
- [ ] A PR justified by reason 1 (hold/revert alone) carries nothing else.
- [ ] Every "undeclared"/"drop" claim cites the contract sources; kept code aligned after drops.
- [ ] Within budget, or each extra PR justified by one of the four reasons.
- [ ] Merge order, release safety, ticket mapping and equivalence check stated.
- [ ] Plan written; user asked to confirm; nothing executed.

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Splitting by file or cherry-picking commits | Split hunks from the final snapshot diff. |
| One PR per finding or requirement | Plan the lean shape within budget; finer splits are options. |
| A lower PR whose tests use a higher PR's symbol | Move the test case up or the symbol down. |
| Calling a change undeclared because the spec omits it | Check ticket AC and review comments too. |
| Missing a change to shared code | Search callers on the **base branch**, not just the snapshot; treat wrappers around old code as behaviour changes. |
| Bundling other fixes into a hold/revert-alone PR | Reverting it would revert them too. Keep that PR to the risky change only. |
| Parking repo rules under "Unverified" | Read them in Step 1 and apply them. |
| Executing before confirmation | Stop after the plan; `apply-pr-split` runs on confirmation. |

## Handoff

When the user confirms (with or without adjustments), update `split-plan.md` to the
confirmed version and invoke `apply-pr-split` with its path.

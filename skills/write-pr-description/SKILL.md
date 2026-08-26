---
name: write-pr-description
description: Use to author the GitHub PR description for review. Checks whether the SDLAIC change directory is inside the PR's git repo (and, if inside, committed) so artifacts the reviewer can't see are never referenced, then emits a tight, review-focused description — ticket/issue link (Jira or GitHub issue, when present), outcome summary, acceptance criteria traced to the ticket/issue, an explicit out-of-scope boundary with reasons, AC→evidence, and brief how-to-verify.
---

# PR Description Phase (Write PR Description)

## Core Principle

The PR description is **what a reviewer reads to scope their review**. A reviewer
decides what belongs to a PR against the *requested outcome* — established
primarily by the ticket/issue (when one exists), then the PR description, then any
clarifying discussion. The description's job is narrow and load-bearing:

1. **Link the ticket/issue** (Jira or GitHub issue) so the reviewer's primary
   context (what was actually asked) resolves — omitted only when the change has
   no ticket or issue at all.
2. **State the requested outcome + acceptance criteria** — a sharp, checkable
   definition of done.
3. **Draw an explicit out-of-scope boundary** (with reasons) — the material a
   reviewer uses to separate findings that belong to *this* PR from follow-up
   work. This is the single highest-leverage section for keeping review focused.

What this artifact is **not**: a per-file changelog. Reviewers read the diff
directly for code-level assessment; narrating every file wastes the description's
real job (scope framing) and bloats the PR.

## Process

### Step 1: Is the SDLAIC change inside the PR's git repo?

```bash
CHANGE_DIR="$(sdlaic path change --change <name>)"   # globally configurable — NEVER assume the path
REPO_ROOT="$(git rev-parse --show-toplevel)"          # the repo this PR lives in
```

The change directory's path is configurable, so it may live **inside** the PR's
repo or **outside** it. Determine which — this decides whether git is even
involved:

- **Outside the repo** (`$CHANGE_DIR` is not under `$REPO_ROOT`) → git does not
  track it, it is never committed, and it cannot be part of this PR. **Do not
  reference it in the description.** Read it locally to derive the content (AC,
  scope, evidence); keep its paths out.
- **Inside the repo** → it is in git's scope, so check whether it is actually
  committed for this PR (`git ls-files "$CHANGE_DIR"` / `git log --oneline --
  "$CHANGE_DIR"`):
  - **Committed** → the artifacts are visible to the reviewer; you may reference
    them (e.g. link `proposal.md` / `spec.md`, trace an AC to a spec scenario).
  - **Not committed** (untracked or `.gitignore`d) → **do not reference them**;
    the reviewer cannot see files that aren't in the PR. Read locally, keep paths
    out.

This is the headline guarantee of the skill: **only reference what the reviewer
can actually see.**

### Step 2: Derive AC and trace them to the ticket/issue

Read the change's `proposal.md` — both its **IN-SCOPE** deliverables (→ AC) and
its **OUT-OF-SCOPE** boundary with reasons (→ the Out-of-scope section) — plus
`specs/<capability>/spec.md` (GIVEN/WHEN/THEN) and `tasks.md` (success
criteria). Derive the acceptance criteria from the **behavior the change
delivers**, not from the file list.

**Each AC must trace to the ticket/issue when one exists.** The ticket/issue is
the source of truth for what was asked; the PR description elaborates it. If an
AC is in the PR but not the ticket/issue, a reviewer may flag over-reach; if a
ticket/issue requirement is missing from the AC, a reviewer may flag
incompleteness. Mirror its done-conditions and cite its key per AC. **When the
change has no ticket or issue**, derive the AC from the artifacts instead and
skip the trace check.

### Step 3: Draft the description

Compose the description using this template (the file write happens in Step 5,
after the gate). Every cited `file:line` must be a real change in this PR — read
the PR diff, cite it, do not invent or overclaim.

```markdown
# [<KEY or #N>] <one-line outcome, not the implementation>
<!-- Drop the [key]/[#N] prefix entirely when the change has no ticket/issue. -->

**Ticket/Issue:** [<KEY or #N> — <summary>](<url>) (Priority: <P>)
<!-- Jira ticket or GitHub issue. Omit this line entirely when the change has neither. -->

**Context:** [proposal](<change-path>/proposal.md) · [spec](<change-path>/spec.md) · [design](<change-path>/design.md)
<!-- Repo-relative links to the SDLAIC artifacts. ONLY when CHANGE_DIR is inside the repo AND committed (Step 1); omit entirely otherwise. -->

## Summary
<1–3 sentences: the problem (the WHY) and the outcome this PR delivers. No
implementation detail, no file list.>

## Acceptance Criteria
<One lead-in sentence: how many AC there are, how they group, and what they were
compressed from — e.g. "Twelve criteria across three independent defects plus the
backfill — roughly three per defect, compressed from the specs' 12 requirements
and 48 scenarios".>

**<Group 1 — the defect/capability stated as the reviewer sees it, not as the fix>**

| # | Acceptance Criterion |
|---|----------------------|
| AC1 | <one observable, binary-checkable assertion> |
| AC2 | <one observable assertion — the negative/absence form counts> |

**<Group 2 — …>**

| # | Acceptance Criterion |
|---|----------------------|
| AC3 | ... |

<!-- One bold header + one table per group. Group by defect / capability / spec
     (one group per `specs/<capability>/spec.md`, plus one for a
     migration/backfill when there is one). Number AC continuously across groups
     (AC1…ACn) — never restart per group — so "How the AC Are Met" and "How to
     verify" stay flat, single-table lookups. A single-capability change with ≤5
     AC keeps one unlabelled table and no group headers. -->

## Out of scope (with reason)
- <adjacent item> — <DEFERRED / PRE-EXISTING / BLOCKED on <ticket> / SEPARATE TICKET <KEY>>: <one-line why>

## How the AC Are Met
| AC | Evidence (file:line) |
|----|----------------------|
| AC1 | <what changed where — every cited file MUST be a real change in the diff> |
| ... | ... |

## How to verify
- **AC1** — <grep / DB column / emitted event / manual step>
- **AC2** — ...
**Not covered (stated, not hidden):**
- <gap> — <why it cannot be asserted now> (blocked on <ticket>)
```

**Section rationale (vs. a traditional PR template):**

| Section | Why it helps the reviewer |
|---|---|
| Ticket/Issue link | Lets the reviewer pull the primary context — what was actually asked. Included when a Jira ticket or GitHub issue exists; omitted otherwise. |
| Context (artifacts) | Optional repo-relative links to proposal/spec/design for depth. Included ONLY when the change dir is committed inside the repo; the PR stays self-contained without them. |
| Summary | Frames the requested outcome for the assessment. |
| AC (ticket/issue-traced) | The positive, verifiable criteria the change is assessed against. |
| Out of scope (with reasons) | The material a reviewer uses to separate findings that belong to this PR from follow-up work. Highest leverage for keeping review focused. |
| How the AC Are Met | Maps each AC to diff evidence — the reviewer's shortcut. |
| How to verify | Replaces a verbose "Test Plan". Brief, per-AC. The "Not covered" lines preempt "missing coverage" findings. |

### Step 4: AC + scope quality gate

Before finalizing, run both gates. These are the controls that keep the review
focused — they are the point of this skill.

**AC quality gate** — AC states what is *delivered and verifiable* (the boundary
lives in Out-of-scope, not here):

- [ ] Each AC is **one observable, binary-checkable assertion** (a column value,
      an emitted event, an HTTP response, an absent side-effect) — never "the
      system should handle X".
- [ ] **No forbidden words**: `correctly`, `properly`, `appropriately`,
      `should handle`, `works as expected`.
- [ ] Each AC has a stated **verification method** (Step 3 "How to verify").
- [ ] AC are **grouped by defect/capability** — one bold header + one table per
      group, preceded by a lead-in sentence naming the count and the grouping —
      and numbered **continuously** across groups (AC1…ACn). A single-capability
      change with ≤5 AC stays one unlabelled table.

**Scope quality gate** — the primary lever for a focused review:

- [ ] Every AC **traces to the ticket/issue requirement** when one exists
      (Step 2); no orphan AC, no dropped requirement. (Skipped when there is no
      ticket/issue.)
- [ ] The **Out-of-scope list names every plausible-but-excluded adjacent item**
      with a reason — this is the material a reviewer buckets findings against. An
      empty or vague Out-of-scope list is the most common cause of review sprawl.
- [ ] Every "Not covered" test gap is **stated with a reason**, not hidden.
- [ ] **No uncommitted artifact paths** are referenced in the description
      (Step 1).

### Step 5: Write and hand off

Write `pr-description.md` to `$CHANGE_DIR`. This file is the canonical draft; its
contents become the PR body when the PR is opened.

## Output Artifacts

- `$(sdlaic path change --change <name>)/pr-description.md` — the review-focused
  PR description.

## Verification

- [ ] `CHANGE_DIR` resolved via `sdlaic path change`; its location (inside vs
      outside the PR repo) and committed status determined.
- [ ] No uncommitted artifact paths are referenced in the description.
- [ ] Context links (proposal/spec/design) appear ONLY when the change dir is
      committed inside the repo; omitted otherwise.
- [ ] Every AC traces to the ticket/issue when one exists; no orphan AC, no
      dropped requirement.
- [ ] No forbidden words in any AC; each AC is one observable, checkable assertion.
- [ ] Out-of-scope list names each plausible-but-excluded item with a reason.
- [ ] Every cited `file:line` in "How the AC Are Met" is a real change in the PR.
- [ ] Test gaps are stated under "Not covered" with reasons — none hidden.
- [ ] No verbose per-file "What Changes" prose; file accountability is one line each.
- [ ] Ticket/Issue link and the title `[key]` prefix are both omitted cleanly when
      the change has no ticket or issue.

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Referencing change artifacts the reviewer can't see | If `CHANGE_DIR` is outside the PR repo, or inside but uncommitted/`.gitignore`d, do not link or mention its files — the reviewer cannot see them. Read them locally to derive the content, but keep their paths out of the description. |
| Making the PR depend on the artifacts | The PR must be self-contained — a reviewer should not have to open proposal/spec/design to understand the change. Context links supplement; they never replace Summary/AC/Out-of-scope. |
| Narrating every file's changes in prose ("What Changes") | Reviewers read the diff directly for code-level assessment — do not re-narrate it. Reserve prose for outcome + AC + scope. |
| AC not traced to the ticket/issue | When a ticket or GitHub issue exists, it is the source of truth for what was asked; the PR description elaborates it. Mirror its done-conditions and cite the key per AC, or the reviewer will flag over-reach / incompleteness. (No trace check when the change has neither.) |
| Putting the scope boundary inside AC | AC states what is *delivered and verifiable*. "What we deliberately did NOT do, and why" belongs in Out-of-scope (with reasons) — that is the material a reviewer buckets findings against. Do not duplicate it as AC. |
| Forcing a Jira link when the change has no ticket | A change may carry a Jira key, a GitHub issue, or neither. Include whichever exists; if neither, omit the link line and the title `[key]` prefix — do not invent one. |
| Citing a `file:line` that is not a real change in the PR | Only cite what is actually in the diff. An overclaim (or a dead reference) is exactly what a reviewer will probe. |
| Empty or vague Out-of-scope list | This is the material a reviewer uses to separate in-PR findings from follow-up work. Name every plausible-but-excluded adjacent item with a reason (DEFERRED / PRE-EXISTING / BLOCKED / SEPARATE TICKET). |
| Hiding untested gaps | State them under "Not covered" with a reason + blocking ticket. An honest gap preempts a "missing coverage" finding; a hidden one invites one. |
| Bloated "How to verify" / "Test Plan" | One verification line per AC, plus the "Not covered" block. Cut hygiene logs and coverage narration — keep what a reviewer needs to confirm each AC. |
| Assuming the change directory path | Always `sdlaic path change --change <name>`. The path is globally configurable; hardcoding it breaks on other machines. |

## Handoff

Open the GitHub PR using `pr-description.md` as the body. The reviewer scopes the
change against the ticket/issue (primary, when present) and this description
(secondary); if a ticket or issue exists, ensure its link is live so the primary
context resolves.

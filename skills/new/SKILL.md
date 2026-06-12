---
name: new
description: Use when starting new work, a user provides a Jira key, or the enforcer routes to the initialization phase. Creates an SDLAIC change with full context from Jira and codebase research.
---

# New Change Initialization

## Core Principle

Every piece of work begins with context. Gather Jira context, research the codebase, and initialize the SDLAIC change before anything else.

## Pre-conditions

- No active SDLAIC change (or user explicitly wants a new one)
- Jira CLI is authenticated (`jira serverinfo` succeeds)

## Input

The user may provide:
- A **Jira key** (e.g., `JIRA-123`) — required for ticket-driven work
- A **change name** (kebab-case, e.g., `add-user-export`) — if not provided, derive from Jira summary or ask
- A **project path** (e.g., `path/to/projects`) — if not provided, ask which project this relates to

## Process

### Step 1: Gather Jira Context (if key provided)

If a Jira key is provided, fetch the full ticket context:

```bash
# Get description, status, priority, assignee, labels
jira issue view <KEY> --plain

# Get all comments (may contain clarifications, decisions, acceptance criteria)
jira issue comment list <KEY> --plain
```

From the output, extract the following metadata. For the description body, do
NOT summarize or restructure — copy verbatim, preserving every section, heading,
and bullet exactly as the ticket author wrote it. If the ticket has sections like
"Open Questions", "Dependencies", "Acceptance Criteria", "Technical Approach",
"Out of Scope", etc., they MUST all appear in `context.md` as-written.

- **Summary:** One-line description of what the ticket asks for
- **Type:** Bug, Story, Task, Epic
- **Priority:** High/Medium/Low
- **Status:** Current workflow status
- **Description (verbatim):** The raw description body — every section, every
  heading, every bullet. Do not invent your own section hierarchy.
- **Linked Issues:** Related/blocking/blocked-by issues with their status
- **Comments:** Key decisions, clarifications, and context from comments
- **Labels/Components:** For scoping

If no Jira key is provided, ask the user for a description of the work.

**Error handling:**
- `401 Unauthorized`: Jira auth issue. Suggest checking `~/.netrc` or `JIRA_API_TOKEN`. Do NOT proceed without context.
- `404 Not Found`: Issue doesn't exist or no permission. Ask user to verify the key.
- `jira: command not found`: Jira CLI not installed. Ask user to install or provide context manually.

### Step 1.5: Ticket Quality Check

A weak ticket poisons every downstream phase. Before proceeding, assess ticket
completeness against this checklist (mark each Y/N):

- [ ] **Description** — exists, > 50 chars, not vague ("improve X", "fix issue" alone fails)
- [ ] **Acceptance Criteria** — explicit list of done-conditions
- [ ] **Scope Boundary** — ticket mentions what is NOT included, or scope is self-evidently narrow
- [ ] **Dependencies / Blockers** — noted explicitly, or genuinely none
- [ ] **Open Questions / Unknowns** — listed explicitly, or ticket is fully resolved

Compute quality:
- **HIGH** = 5/5 pass
- **MEDIUM** = 3–4/5
- **LOW** = ≤ 2/5

**If MEDIUM or LOW**, present options to the user:

> ⚠️ Ticket quality is {LEVEL}. Missing: {list of failed items}.
>
> A. **Update the ticket in Jira first** (recommended for shared work — the
>    ticket is the source of truth and other people will read it later)
> B. **Provide the missing info inline now** (I'll record it in `context.md`
>    with `[INLINE-PROVIDED]` tags so grillme treats it as ground truth)
> C. **Proceed anyway** — `context.md` will be flagged with the quality level;
>    a structured **Gap Manifest** will be written so grillme knows exactly which
>    surfaces are mandatory and what to elicit for each missing item — grillme
>    uses this as ground truth, not as a hint

- User picks A → halt this skill and ask the user to come back when the ticket is updated.
- User picks B → ask each missing item one at a time, record answers inline in `context.md`.
- User picks C → record the quality flag, generate Gap Manifest (see Step 3), and proceed.

The Gap Manifest maps each failed checklist item to the grillme surface(s) that
must cover it. Use this mapping to populate it:

| Failed Item | Grillme Surface | Priority | Elicit |
|-------------|----------------|----------|--------|
| Description vague/missing | Failure modes, Edge cases | 1 | What exactly changes, worst-case failure, known edge cases |
| Acceptance Criteria | ROI / impact | 1 | Done conditions, measurable success metrics, user-visible outcomes |
| Scope Boundary | Scope boundaries | 2 | What is explicitly NOT included, where the change ends |
| Dependencies / Blockers | Failure modes, Rollback | 2 | What happens if a dependency isn't ready or breaks mid-deploy |
| Open Questions | (assign per question) | 1 | For each open question: create a separate row, assign nearest catalog surface based on content; default to Failure modes if unclear |

The result is recorded in `context.md` under `## Ticket Quality Assessment` and
`## Gap Manifest` in Step 3.

If quality is HIGH, this step passes silently — do not bother the user.

### Step 2: Codebase Research

Run a code research query against the relevant project path. Tool selection follows your workspace convention (see `references/code-research.md`).

- **Primary:** one targeted query on the topic from the ticket description, scoped to `projects/<project>`.
- **Fallback if project path is unclear:** a broader query without a path scope to locate the right project first.

Empty result is valid — record "no precedent found for `<terms>`" and proceed. Do not silently skip.

### Step 3: Output Research Summary

Produce a concise summary covering:
- **What exists:** Relevant code, patterns, and architecture found
- **What's missing:** Gaps that the change needs to address
- **Dependencies:** Other systems or modules this change touches
- **Risks:** Potential conflicts or breaking changes identified

### Step 3.5: Extract Actors & Use Cases

A lightweight, evidence-grounded extraction so `grillme` can anchor questions to a specific `(Actor, Use Case)` instead of asking abstractly. This is **not** a formal spec — the SDLAIC `spec.md` (written in PROPOSED phase) still holds the WHEN/THEN clauses.

**Hard provenance bar.** Every actor and every use case must cite either a verbatim ticket quote OR a `file:line` symbol from the Step 2 code research. No source → drop the row.

**Skip rule.** If the change is cosmetic / docs-only / test-only / dependency bump / pure refactor with no behavior change, write `N/A: <one-line reason>` for the whole section in Step 5 and proceed. Do not force actors onto trivial changes.

**Two-pass extraction:**

1. **Pass 1 — Human actors from ticket text.** Re-read the ticket description and key comments. Extract every named role / persona / user type the author mentioned. Each row gets `[ticket: "<verbatim quote>"]` citation.
2. **Pass 2 — System actors from code research.** Re-scan the Step 2 research output for non-human callers: cron jobs, queue workers, webhook receivers, scheduled tasks, policy guards, route middleware, role/permission enums, third-party API callers. Each row gets `[code: <file:line> <symbol>]` citation.
3. **Merge & flag conflicts.** If a ticket-named role does not reconcile with code (e.g. ticket says "Manager" but code has only `Admin` + `Instructor` enum values), flag the row with `⚠ MISMATCH: <description>` rather than silently picking one. The mismatch is a grillme question.

**Caps.** ≤ 6 actors total, ≤ 3 primary use cases per actor. Borderline → merge or drop. The aim is anchoring quality, not exhaustiveness.

**Use case schema (rigid):** `Actor → Trigger → Action → Object → Outcome`.
- *Trigger* and *Action* should point to a concrete route / command / job / event in code. If the change *introduces* the path (no precedent), tag the use case `[NEW SURFACE]` instead of inventing a citation.
- Each use case must have **Happy path** + at least one **Alternate path** OR **Edge case**. Happy-path-only is filler.

**State transitions** go in a separate table keyed by *entity*, not by use case: `Entity | From | Event | To | Side effects`. One row per transition.

**Side-effect-only actions** (notifications, webhook fan-out, async dispatches that do not change persisted state) go in a separate list — do not nest them inside state-transition rows.

**Cross-link to Gap Manifest.** For each row in `## Gap Manifest` (if populated), tag the affected actor(s) and use case(s) in the cross-link table. This is the anchor that makes grillme's gap-driven questions concrete.

#### Quality rubric — apply before committing each row

- **GOOD actor:** named in ticket OR code, single citation, one-line role description that distinguishes them from other actors.
- **BAD actor:** "User" / "Admin" with no citation; an actor invented from "what users might want"; a system actor with no code reference.
- **GOOD use case:** specific trigger → concrete action → named object → observable outcome, citation per element where possible, ≥1 alternate or edge enumerated.
- **BAD use case:** "User uses the system" (vague); paths that don't map to any code path AND aren't tagged `[NEW SURFACE]`; happy-path-only.

#### Worked example (hypothetical ticket: "Team Managers should only see members in their own teams")

```
Actors:
| Team Manager | Human  | [ticket: "Team Managers can currently see members of other teams"] | Manages members within assigned teams |
| System Admin | Human  | [code: pkg/auth/roles.go:12 RoleAdmin] | Sees all members across all teams |
| Sync Job     | System | [code: pkg/jobs/sync.go:45] | Syncs team membership from external directory nightly |

Use Cases — Team Manager:
- Happy path: view team roster → filter by assigned team → see only members in that team [code: pkg/handlers/roster.go:34]
- Edge case: Team Manager assigned to no teams → empty roster list [NEW SURFACE]
- Edge case: team member belongs to multiple teams, one assigned + one not → must still appear [ticket quote]

State Transitions:
| RosterCache | stale | nightly sync job runs | fresh | none |

Gap Manifest Cross-link:
| "Scope: what about multi-team members?" | Team Manager | edge case: multi-team membership |
```

This step's output is consumed downstream by `grillme` (anchors questions) and `brainstorm` (informs design trade-offs). It is **not** copied into `spec.md`.

### Step 4: Initialize SDLAIC Change

```bash
sdlaic new change "<change-name>"
```

Derive the change name:
- From Jira summary if available (convert to kebab-case)
- From user description if no Jira key
- Confirm with user **after** reviewing the research summary — research findings may reveal a scope adjustment that warrants a different name

### Step 5: Save Jira Metadata (if key provided)

If a Jira key was provided, save the gathered context to `context.md`:

```markdown
# Change Context

## Jira
- Key: <KEY>
- Summary: <summary>
- Type: <type>
- Priority: <priority>
- Status: <status>
- URL: https://<your-domain>.atlassian.net/browse/<KEY>

## Linked Issues
<list of linked issues with relationship + status, or "None">

## Dependencies / Blockers
<extracted from the ticket's Dependencies section, or "None noted in ticket">

## Open Questions ⚠️
<extracted verbatim from the ticket's Open Questions section, or "None noted in ticket">
<!-- grillme MUST address each item here -->

## Ticket Description (verbatim)
<paste the raw description from `jira issue view` output verbatim — preserve
 every section (Description, Acceptance Criteria, Technical Approach, Out of
 Scope, etc.) exactly as written. Do NOT reformat, merge, drop, or invent
 sections. If the ticket author wrote "## Foo", the output here must have "## Foo".>

## Key Comments
<extracted decisions and clarifications from comments, or "No comments">

## Ticket Quality Assessment
- Level: HIGH | MEDIUM | LOW
- Checklist: [Description: Y/N, AC: Y/N, Scope: Y/N, Deps: Y/N, OQ: Y/N]
- Missing items handled by: (none / inline-provided / proceed-with-flag)

## Gap Manifest
<!-- Generated by new. Empty when Quality is HIGH or all gaps were inline-provided. -->
<!-- grillme MUST treat every row as a mandatory IN surface — cannot mark N/A without explicit justification. -->
| Gap | Grillme Surface | Priority | Elicit |
|-----|----------------|----------|--------|
| <failed item> | <surface name> | <1–3> | <what grillme must draw out from the user> |

## Actors & Use Cases
<!-- Generated by new Step 3.5. grillme anchors questions to (Actor, Use Case). -->
<!-- If the change is cosmetic/docs/test-only/dep-bump/pure-refactor: replace the tables below with a single line "N/A: <reason>" and skip the rest. -->

### Actors
| Actor | Type | Citation | Notes |
|-------|------|----------|-------|
| <name> | Human / System | [ticket: "<verbatim quote>"] or [code: <file:line> <symbol>] | <one-line role> |

### Use Cases
#### <Actor name>
- **Happy path:** <Trigger → Action → Object → Outcome> [citation]
- **Alternate paths:**
  - <variation> [citation]
- **Edge cases:**
  - <case> [citation or NEW SURFACE]

### State Transitions
<!-- One row per entity transition. Empty if the change is read-only. -->
| Entity | From | Event | To | Side effects |
|--------|------|-------|----|--------------|

### Side-effect-only Actions
<!-- Notifications, webhook fan-out, async dispatches with no persisted state change. -->
- <Actor> → <action> → <effect> [citation]

### Gap Manifest Cross-link
<!-- Empty when ## Gap Manifest is empty. Otherwise one row per gap. -->
| Gap | Affected Actor(s) | Affected Use Case(s) |
|-----|-------------------|---------------------|
```

If no Jira key was provided and the user gave a description, save that instead under `## User Description`.

**`context.md` is always written**, even when no Jira key is provided. This file must exist for grillme to check the Gap Manifest.

### Step 6: Commit Initialized Artifacts

Stage and commit the SDLAIC change directory inside the **target project's repository**:

```bash
git -C projects/<project> add .sdlaic/changes/<change-name>/
git -C projects/<project> commit -m "<prefix>: new context for <change-name>"
```

Commit prefix fallback chain (use first applicable):
1. **Jira key available** → `JIRA-1013: new context for add-user-export`
2. **Change slug only** → `add-user-export: new context`
3. **Neither** → `chore(sdlaic): new context for add-user-export`

This is a stable checkpoint. Future agents can roll back here if direction changes.

### Step 7: Handoff

```
Change "<change-name>" initialized.
- Ticket Quality: {HIGH|MEDIUM|LOW}
- Open Questions captured: {N} (grillme will address each)
- Dependencies noted: {N}
- Actors extracted: {N} (or N/A — trivial change)
- Use cases extracted: {N total}

Next: Run grillme. Pay special attention to the `## Open Questions ⚠️` section,
`## Gap Manifest`, and `## Actors & Use Cases` in context.md — grillme will treat
every row in the manifest as a mandatory surface, and will anchor each question
to a specific (Actor, Use Case) when that section is populated.
```

## Output Artifacts

- `.sdlaic/changes/<change-name>/` — change directory created by SDLAIC
- `.sdlaic/changes/<change-name>/context.md` — Jira + research context (always written; contains Gap Manifest when quality is MEDIUM/LOW + Option C, and `## Actors & Use Cases` with populated tables or `N/A` for trivial changes)

## Verification

- [ ] SDLAIC change directory exists
- [ ] Research Summary was output before handoff
- [ ] User confirmed the change name
- [ ] If Jira key provided: `context.md` exists with ticket description and key comments
- [ ] `context.md > ## Ticket Description (verbatim)` contains the description as-written, with no section dropped or restructured
- [ ] `context.md > ## Open Questions ⚠️` exists (with content or "None noted in ticket")
- [ ] `context.md > ## Dependencies / Blockers` exists (with content or "None noted in ticket")
- [ ] `context.md > ## Ticket Quality Assessment` exists with Level + Checklist + Missing items handled by
- [ ] `context.md > ## Gap Manifest` exists (populated when quality is MEDIUM/LOW + Option C; empty table rows otherwise)
- [ ] `context.md > ## Actors & Use Cases` exists (populated tables OR explicit `N/A: <reason>` for trivial changes)
- [ ] Every actor row has a citation — `[ticket: "..."]` or `[code: <file:line> <symbol>]` — no uncited rows
- [ ] Every use case has Happy path + ≥1 Alternate or Edge case (no happy-path-only rows)
- [ ] Caps respected: ≤ 6 actors, ≤ 3 primary use cases per actor
- [ ] If `## Gap Manifest` is populated, every gap is cross-linked to ≥1 actor/use case in `## Gap Manifest Cross-link`
- [ ] Initialized artifacts committed to target project repo (`.sdlaic/changes/<change-name>/`)

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Fetching only the description, not comments | Comments often contain acceptance criteria and decisions. Always fetch both. |
| Skipping code research | Research prevents rework. Always run it. |
| Creating the change before confirming the name | Always confirm with the user first. |
| Proceeding without any context | If Jira fetch fails and user can't provide context, STOP. Don't guess. |
| Ignoring the project path | Research without project scope returns irrelevant results. Ask which project. |
| Reformatting the ticket description into an agent-chosen section hierarchy | Paste the raw `jira issue view` description verbatim under `## Ticket Description (verbatim)`. Do not invent your own sections, do not merge, do not drop. Open Questions, Dependencies, AC, etc. must all survive intact. |
| Skipping the Ticket Quality Check because the ticket "looks fine" | Always run the 5-item checklist in Step 1.5. The cost is seconds; the cost of a missed gap is rework downstream (gap discovered during manual testing post-review). |
| Listing "User" / "Admin" as generic actors with no citation | Drop. Either cite the ticket's exact wording (`[ticket: "..."]`) or cite the code symbol (`[code: <file:line> <path/to/file>]`). Generic uncited actors are exactly what the provenance bar exists to block. |
| Inventing actors from "what users might want" | Provenance bar. No source → no actor. If the change introduces a new actor type, tag `[NEW SURFACE]` on the use case rather than fabricating a citation. |
| Forcing the Actors & Use Cases section onto a trivial change (typo, dep bump, pure refactor) | Use `N/A: <one-line reason>` for the whole section. The extraction has a cost; skip it when the change has no behavior surface. |
| Nesting side-effect emails/webhooks inside the State Transitions table | State Transition rows are only for persisted state changes (DB / cache / file). Notifications, webhook dispatches, async fan-out with no persisted change go in the separate "Side-effect-only Actions" list. |
| Producing happy-path-only use cases | Every use case needs ≥1 Alternate or Edge case (or a one-line note why none exist). Happy-path-only rows give grillme nothing to anchor on. |
| Treating the Actors section as the SDLAIC `spec.md` | This is lightweight pre-design extraction for grillme anchoring. The formal WHEN/THEN spec is `spec.md`, written later in PROPOSED phase. Do not duplicate or pre-write spec clauses here. |

## Handoff

Route to `grillme` to challenge assumptions before design begins.

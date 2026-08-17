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

### Step 1: Input Gathering & Resolution

#### Step 1.1: Source Input Resolution (Jira / Text / File Path / URL)

First, classify the user's input:
- **Jira Key**: If a Jira key is provided (e.g., `JIRA-123`), fetch the full ticket context:
  ```bash
  # Description, status, priority, assignee, labels — and the comment COUNT in the header
  jira issue view <KEY> --plain

  # ALL comments. Bare `--plain` renders only the LATEST comment.
  # Pass an explicit high count; there is no `jira issue comment list` subcommand.
  jira issue view <KEY> --comments 50 --plain
  ```

  **Comment completeness is mandatory — verify it, do not assume it.** The header of the first
  command reports the true count (e.g. `💭 4 comments`). Count the comments actually returned by the
  second command and compare.

  - Counts match → proceed.
  - Counts differ → **HALT and say so loudly.** Re-run with a higher `--comments` value. Silent
    truncation is indistinguishable from a thin ticket, and the missing comments are usually where
    the investigation, the `file:line` anchors, and the scope decision live.

  From the output, extract issue metadata (key, summary, type, status, priority, URL, **reporter**, **assignee**) plus the full description and all comments as raw source material for Step 4. Reporter and assignee are needed by Step 4.0. The description and comments are SOURCES to decompose into candidate scopes — they are NOT pasted wholesale into context.md. Do NOT proceed without context.
  
- **File Path**: If the user provides a local file path or reference (e.g. `/path/to/file.md` or similar absolute/relative path format):
  1. Use a file reading tool to check if the file exists and is readable.
  2. If the file does NOT exist, is empty, or is unreadable, HALT. Inform the user and ask them to check the path or provide a raw description.
  3. If the file exists and is readable, read its contents. Use the resolved file content as the "description of the work" for all downstream steps.

- **External URL**: If the user provides an external URL (e.g. a link to a GitHub issue, online design doc, or ticket page):
  1. Use a web reading or browser tool to fetch the content of the URL.
  2. If the fetch fails, or the page is unreadable, or contains no meaningful content, HALT. Inform the user and ask them to verify the URL or provide a raw description.
  3. If successful, use the resolved web content as the "description of the work" for all downstream steps.
  
- **Raw Text Description**: If the user provides a raw text description directly, use it as the "description of the work".

If no input is provided, ask the user for a description of the work.

#### Step 1.2: Description Quality Assessment

Assess the completeness of the resolved description (from Jira issue description, resolved file contents, resolved URL contents, or raw text description) against this checklist (mark each Y/N):

- [ ] **Description** — exists, is NOT a file path or URL string, contains actual descriptive content of the work (> 50 characters of descriptive text), and is not vague ("improve X", "fix issue" alone fails)
- [ ] **Acceptance Criteria** — explicit list of done-conditions
- [ ] **Scope Boundary** — mentions what is NOT included, or scope is self-evidently narrow
- [ ] **Dependencies / Blockers** — noted explicitly, or genuinely none
- [ ] **Open Questions / Unknowns** — listed explicitly, or fully resolved
- [ ] **Input Resolved** — if a file path or URL was provided, it was successfully resolved/read and its content is loaded as the description
- [ ] **Evidence Freshness** — every quantitative claim (row counts, affected-user figures, date ranges, "N% of orders") is reproducible against a **live** source, and the source + `as-of` date are stated. Fails if figures come from a stale or sanitised snapshot, or if no source/date is given at all.

**Why Evidence Freshness is separate.** The other six items test whether a ticket is *well-written*.
This one tests whether its evidence is *current and verifiable* — an orthogonal axis. A ticket can be
immaculately written and still rest on numbers nobody can reproduce. When it fails, record in the
Gap Manifest: the snapshot date, and **which columns/tables are unusable and why** (e.g. sanitiser
NULLs the column the sizing query needs). A scope that cannot be sized from live data gets
`Readiness: needs-sizing` in Step 4.1 — this check is where that value comes from.

Compute quality:
- **HIGH** = 7/7 pass
- **MEDIUM** = 5–6/7
- **LOW** = ≤ 4/7

Report the level and the failing axis together — `HIGH`, or e.g. `MEDIUM (evidence-stale)`. "Well-written" and "evidence unreproducible" are both true statements about the same ticket; never let a high score hide an unreproducible figure.

**If MEDIUM or LOW**, present options to the user:

> ⚠️ Description/ticket quality is {LEVEL}. Missing: {list of failed items}.
>
> A. **Update the source first** (recommended — update the ticket in Jira, or update the local description file first so it remains the source of truth)
> B. **Provide the missing info inline now** (I'll record it in `context.md` with `[INLINE-PROVIDED]` tags so grillme treats it as ground truth)
> C. **Proceed anyway** — `context.md` will be flagged with the quality level; a structured **Gap Manifest** will be written so grillme knows exactly which surfaces are mandatory and what to elicit for each missing item

- User picks A → halt this skill and ask the user to come back when the source is updated.
- User picks B → ask each missing item one at a time, record answers inline in `context.md`.
- User picks C → record the quality flag, generate Gap Manifest, and proceed.

The Gap Manifest maps each failed checklist item to the grillme surface(s) that must cover it. Use this mapping to populate it:

| Failed Item | Grillme Surface | Priority | Elicit |
|-------------|----------------|----------|--------|
| Description vague/missing | Failure modes, Edge cases | 1 | What exactly changes, worst-case failure, known edge cases |
| Acceptance Criteria | ROI / impact | 1 | Done conditions, measurable success metrics, user-visible outcomes |
| Scope Boundary | Scope boundaries | 2 | What is explicitly NOT included, where the change ends |
| Dependencies / Blockers | Failure modes, Rollback | 2 | What happens if a dependency isn't ready or breaks mid-deploy |
| Open Questions | (assign per question) | 1 | For each open question: create a separate row, assign nearest catalog surface based on content; default to Failure modes if unclear |
| Evidence Freshness | ROI / impact, Failure modes | 1 | Which figures are unreproducible, the snapshot `as-of` date, which columns/tables are destroyed or unavailable, and what would have to be re-run against live data before the affected scope can be sized |

The result is recorded in `context.md` under `## Ticket Quality Assessment` and `## Gap Manifest` in Step 6.

If quality is HIGH, this step passes silently — do not bother the user.

### Step 2: Codebase & Web Research

Run research queries to gather context about the change. Tool selection follows your workspace convention (see `references/code-research.md`).

1. **Local Codebase Research**:
   - Run a targeted codebase query on the topic from the resolved description, scoped to `projects/<project>` (or a broader query if the project path is unclear).
   - Look for existing precedents, related configurations, or circular dependencies.
   
2. **Prior-art check against existing SDLAIC changes** — mandatory, do not skip:
   ```bash
   sdlaic list          # include archived/completed changes, not just active ones
   ```
   For each existing change whose name or `proposal.md` touches the same surface as this ticket,
   read its `proposal.md` (and `design.md` if present). An existing change may already specify —
   or have already shipped — the remedy you are about to propose.

   Record every hit. In Step 4.1 tag the affected candidate row `⚠ OVERLAPS <change-name>` and cite
   the artifact (`<change>/proposal.md`). Recommending a scope that a sibling change already
   designed risks duplicating or silently contradicting gated design work.

   Empty result is valid — record "no overlapping SDLAIC change found" and proceed.

3. **External / Web Research**:
   - If the resolved description mentions any third-party APIs, external libraries, protocols, or services (e.g., "Antigravity provider", external SDKs, or integrations) that are new or have no established local codebase precedents:
     * Perform a web search to find official documentation, API specifications, or known limitations for these technologies.
     * Do not guess or assume how external services behave.
     * Record key web research findings (including URL citations) in the Research Summary.

Empty result is valid — record "no codebase precedent or web documentation found for `<terms>`" and proceed. Do not silently skip.

### Step 3: Research Output & Actors Extraction

#### Step 3.1: Output Research Summary

Produce a concise summary covering:
- **What exists (Codebase):** Relevant code, patterns, and architecture found locally.
- **Existing SDLAIC changes (prior art):** Overlapping changes found by `sdlaic list`, with the artifact that overlaps (`<change>/proposal.md`) and what it already specifies — or "no overlapping SDLAIC change found".
- **Comment ordering detected:** newest-first / oldest-first, and the comment count verified against the ticket header.
- **External/Web Findings:** Key findings from web search on external dependencies, services, or new concepts mentioned in the description (with URL citations). If none, state "No external research needed".
- **What's missing:** Gaps that the change needs to address.
- **Dependencies:** Other systems or modules this change touches.
- **Risks:** Potential conflicts or breaking changes identified.

#### Step 3.2: Extract Actors & Use Cases

A lightweight, evidence-grounded extraction so `grillme` can anchor questions to a specific `(Actor, Use Case)` instead of asking abstractly. This is **not** a formal spec — the SDLAIC `spec.md` (written in PROPOSED phase) still holds the WHEN/THEN clauses.

**Hard provenance bar.** Every actor and every use case must cite either a verbatim ticket/resolved description quote OR a `file:line` symbol from the Step 2 code research. No source → drop the row.

**Skip rule.** If the change is cosmetic / docs-only / test-only / dependency bump / pure refactor with no behavior change, write `N/A: <one-line reason>` for the whole section in Step 6 and proceed. Do not force actors onto trivial changes.

**Two-pass extraction:**

1. **Pass 1 — Human actors from description text.** Re-read the ticket/resolved description and key comments. Extract every named role / persona / user type mentioned. Each row gets `[ticket: "<verbatim quote>"]` or `[description: "<verbatim quote>"]` citation.
2. **Pass 2 — System actors from code research.** Re-scan the Step 2 research output for non-human callers: cron jobs, queue workers, webhook receivers, scheduled tasks, policy guards, route middleware, role/permission enums, third-party API callers. Each row gets `[code: <file:line> <symbol>]` citation.
3. **Merge & flag conflicts.** If a named role does not reconcile with code (e.g. description says "Manager" but code has only `Admin` + `Instructor` enum values), flag the row with `⚠ MISMATCH: <description>` rather than silently picking one. The mismatch is a grillme question.

**Caps.** ≤ 6 actors total, ≤ 3 primary use cases per actor. Borderline → merge or drop. The aim is anchoring quality, not exhaustiveness.

**Use case schema (rigid):** `Actor → Trigger → Action → Object → Outcome`.
- *Trigger* and *Action* should point to a concrete route / command / job / event in code. If the change *introduces* the path (no precedent), tag the use case `[NEW SURFACE]` instead of inventing a citation.
- Each use case must have **Happy path** + at least one **Alternate path** OR **Edge case**. Happy-path-only is filler.

**State transitions** go in a separate table keyed by *entity*, not by use case: `Entity | From | Event | To | Side effects`. One row per transition.

**Side-effect-only actions** (notifications, webhook fan-out, async dispatches that do not change persisted state) go in a separate list — do not nest them inside state-transition rows.

**Cross-link to Gap Manifest.** For each row in `## Gap Manifest` (if populated), tag the affected actor(s) and use case(s) in the cross-link table. This is the anchor that makes grillme's gap-driven questions concrete.

#### Quality rubric — apply before committing each row

- **GOOD actor:** named in ticket/description OR code, single citation, one-line role description that distinguishes them from other actors.
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

This step's output is consumed downstream by `grillme` (anchors questions) and the `proposal` / `spec` / `design` skills (inform scope, behavior, and design trade-offs). It is **not** copied into `spec.md`.

### Step 4: Scope Extraction (one change = N capabilities)

A ticket often bundles multiple issues, hypotheses, and follow-ons. Do NOT carry the whole ticket into `context.md`. Decompose it into **candidate scopes**. `new` does NOT make the final IN/OUT decision — that is decided and **gated in `proposal.md`** (after `grillme` challenges the candidates). Here you only record the candidate set; nothing is silently dropped.

**A candidate scope maps to a capability, not to a PR.** One SDLAIC change is already a container for N capabilities:

```
<change>/
├── proposal.md                  # 1 — the scope contract
├── specs/<capability>/spec.md   # N — one per capability, each with its own spec:<capability> gate
├── design.md                    # 1
└── tasks.md                     # 1
```

So a ticket carrying three related defects produces **one** change with three capability specs — not three changes. Splitting into separate changes is the exception, governed by the criteria in Step 4.2, and is decided on **delivery unit**, never on "these are different issues."

**Inputs:** the resolved description + comments (Step 1.1), the research findings (Step 2 / Step 3.1), and the extracted actors (Step 3.2).

#### Step 4.0: Prior Agreement Detection

**Run this BEFORE extraction.** Steps 4.1–4.2 assume a ticket needs *narrowing*. When the ticket already carries a plan someone signed off on, narrowing is the wrong operation — re-opening a settled decision is a defect, not diligence.

1. **Establish comment ordering.** Read the timestamps and state which order the CLI returned (this CLI commonly returns **newest-first**). Never assume chronological order — reading the decision trail backwards inverts which comment supersedes which. State the detected ordering explicitly in the Research Summary.

2. **Scan every comment for approval/decision signals.** Common English markers:
   `approve` · `agreed` · `proceed` · `go ahead` · `ship it` · `no questions` · `LGTM` · `confirmed` —
   plus status transitions (→ In Progress) and assignment events.

   **Match intent, not this word list.** Comments are frequently written in the team's working
   language rather than English, and an approval there carries exactly the same weight. Read every
   comment for the *act* of settling a decision — assent, sign-off, "go ahead and do it", closing a
   question rather than opening one — in whatever language it is written in. A missed approval is
   the single most expensive failure in this step: it silently converts a settled ticket back into
   an open one.

3. **Resolve *what* was approved.** An approval comment rarely restates the plan. Walk **back** to the nearest preceding comment that proposes one, and treat that proposal as the referent. Approval without a resolvable referent is not an agreement — record it as an open question instead.

4. **Emit an Agreed Scope Set.** Every scope in the referent proposal gets
   `✅ AGREED [comment: <author>, <date>] "<verbatim approval quote>"` in Step 4.1.

**Authority.** Any commenter's approval counts — but you must **record the author, date, and verbatim quote** on every `AGREED` row so the user can judge authority themselves. Never assert that an approval is authoritative. If the approver is neither the ticket's reporter nor its assignee, append `⚠ approver is neither reporter nor assignee` to the row. You have both names from Step 1.1.

**Transitivity — do not over-extend `AGREED`.** Only scopes explicitly present in the approved referent proposal are `AGREED`. Adjacent or follow-on work mentioned in passing in the same comment (backfills, cleanups, "and we should also…") gets `consider` plus the note `adjacent to agreed scope — not explicitly approved`, and is raised as a question in Step 4.2. Silently inflating an agreement is the same class of error as ignoring one.

**`AGREED` outranks every heuristic.** You may **not** argue against an agreed scope on grounds of impact, effort, or priority. The single permitted move is to surface a **new blocker discovered in Step 2 research** that the approver could not have known about — stated explicitly as such, with its evidence, and still leaving the scope marked `AGREED`.

If no approval signal is found, record `No prior agreement detected` and continue to Step 4.1 normally.

#### Step 4.1: Extract candidate scopes

Re-read the description, every comment, and the investigation output as a single search space. Extract each distinct issue, hypothesis, requested behavior, or proposed fix as a separate candidate scope.

For each candidate scope, determine:
- **Kind** — `defect` / `recovery` (backfill, repair, audit of already-damaged data) / `safeguard` (prevents recurrence) / `refuted` / `separate-ticket`. Extend the list if a ticket genuinely needs a sixth kind; do not force a bad fit. Only `defect` (and sometimes `safeguard`) rows are candidates for THIS change — the rest are context.
- **Scope** — a short noun phrase naming the issue/behavior (e.g., "Roster leaks cross-team members", "Add export-to-CSV button").
- **Root cause** — the underlying mechanism, if known. Cite `file:line` from research when one exists; otherwise quote the ticket/comment that asserts it.
- **Impact** — who/what is affected and how severely.
- **Source** — where it came from: `[ticket body]`, `[comment: <author> <date>]`, `[research: file:line]`, or `[investigation: <note>]`.
- **Readiness** — `ready-now` / `blocked-on-<X>` / `needs-sizing` / `needs-decision`. This is an **observable property of the scope**, not an opinion about it: `needs-sizing` means the data required to size it is unavailable (see the Evidence Freshness check in Step 1.2); `blocked-on-<X>` names the specific blocker.
- **Disposition** — `✅ AGREED [comment: <author>, <date>] "<quote>"` when Step 4.0 identified it; `REFUTED` when a later comment or research disproves it; otherwise `consider`.

**No single-winner recommendation.** Do not mark one scope as the recommended one. Ranking collapses independent axes into a single opinion and silently outranks stakeholder agreement. `Readiness` carries the steer instead — it is checkable, and a reader can disagree with it on evidence.

**Refuted hypotheses.** A common case: the ticket body or an early comment proposes a cause or fix that a later comment or the codebase research disproves. Keep these as rows tagged `REFUTED` with the disproof citation. They are high-value signal — they record a dead-end the team already ruled out, so a future agent does not re-walk it. Every `REFUTED` row additionally carries:

- **Boundary** — *what this refutation does NOT close.*
- **Era/Scope** — the period, subsystem, or data generation the refutation applies to (e.g. "2015–2021 import-created rows, not current checkout code").

An over-broad refutation is worse than no refutation: it retires a live defect by association. If you cannot state the boundary, the refutation is not established — downgrade the row to `consider`.

**Prior art.** If Step 2's `sdlaic list` check found an existing change touching this scope, tag the row `⚠ OVERLAPS <change-name>` and cite the artifact.

**Provenance bar.** Every candidate scope must cite its source. No source → drop it. Do not invent scopes from "what might be wrong."

**Caps.** Aim for the real set; merge near-duplicates. **If extraction yields > 10 candidates, grouping by `Kind` is mandatory** — render `defect` first, then the remaining kinds as clearly secondary groups. This reduces *display noise, not row count*: every issue and hypothesis still gets a row. A 30-row ticket that reads as 4 actionable defects plus 26 rows of recorded context is manageable; the same 30 rows presented at equal weight is not.

#### Step 4.2: Present candidates and confirm the capability set

Present the candidate scopes as a table showing `# | Kind | Scope | Root cause | Impact | Source | Readiness | Disposition`, **grouped by `Kind` with `defect` first**, and within each group sorted `ready-now` before blocked. Use whichever prompt or confirmation capability your runtime provides — do not assume a specific tool name.

**Confirmation is multi-select.** A change routinely carries several capabilities; a single-select question structurally cannot express that and will silently discard the rest.

Frame the question explicitly:

> Here are the candidate scopes I extracted from the ticket, comments, and investigation.
> ✅ AGREED = already signed off in the ticket (author + date + quote shown) — these are pre-selected.
> `Readiness` says how ready each one is; it is not a recommendation.
> **Select all scopes for THIS change** — each becomes its own `specs/<capability>/spec.md`.
> The formal IN/OUT boundary is decided later in `proposal.md`, after `grillme` challenges these candidates.
>
> ✅ 1. [defect] <scope> — <root cause> — <impact> — <readiness> — AGREED [<author>, <date>]
>    2. [defect] <scope> — <root cause> — <impact> — <readiness> — consider
>    3. [recovery] <scope> — … (context — select only if it belongs in this change)

Rules:
- **Pre-select every `✅ AGREED` scope.** Deselecting one is the user's call, not yours; if they do, record it — a rejected prior agreement is a decision `proposal` must account for.
- Selected scopes become capabilities in **one** change. Record their order — `ready-now` before blocked — so `tasks.md` sequences a cheap live-defect fix ahead of a scope waiting on a production query.
- If the user adds a scope NOT in the candidate list, add it with source `[user]`, `Readiness: needs-decision`, disposition `consider`.
- Do NOT record a final IN/OUT decision or "not-selected with reasons" here — that is `proposal`'s gated job. Record only the candidate set + dispositions + selection.

**When to propose splitting into separate changes.** Only when a criterion below is met — and say which one. Otherwise keep the scopes together as capabilities:

| Split into separate changes | Keep as capabilities in one change |
|---|---|
| Different deploy / rollback unit | Ships together |
| One scope blocked indefinitely (`blocked-on-<X>`, `needs-sizing` with no path to size it) | All `ready-now` |
| Different system, or a different reviewer must own it | Same code area, same reviewer |
| Too large for a single PR review | Fits one PR |

"These are different defects" is **not** a split criterion. Splitting a stakeholder-approved scope set across changes also fragments one agreed contract across several `proposal.md` files — do not do it without a criterion above.

#### Step 4.3: Prune dependent extractions

If any Actor or Use Case extracted in Step 3.2 belongs exclusively to a candidate that was **not selected** in Step 4.2 (a `REFUTED` row, or a `separate-ticket` / unselected scope), drop or relocate that row before writing `context.md` in Step 6. Actors and use cases should track the selected capability set — but keep it light, since the final IN/OUT line is settled in `proposal`.

### Step 5: Initialize SDLAIC Change

```bash
sdlaic new change "<change-name>"
```

**Run this exactly once**, regardless of how many scopes Step 4.2 selected. N selected scopes become N `specs/<capability>/` directories inside this one change (written later by the `spec` skill, one gate each) — they do **not** become N changes. Create additional changes only when Step 4.2's split criteria fired, and say which criterion.

Derive the change name:
- From Jira summary if available (convert to kebab-case)
- From user description if no Jira key
- Confirm with user **after** reviewing the research summary AND the Step 4.2 selection — the name must cover the selected capability set, not the whole ticket. When several capabilities share a cause, name the change for the shared cause rather than concatenating them.

### Step 6: Write context.md

Write the gathered context (ticket or non-ticket source) to `context.md` using the template below. Substitute the candidate scopes (with Kind, Readiness, and dispositions) from Step 4 and the actors from Step 3.2:

```markdown
# Change Context

## Ticket
- Source: <Jira key OR "Text input" OR "File: <path>" OR "URL: <url>">
- Key: <KEY>
- Title: <summary>
- Type: <type>  ·  Status: <status>  ·  Priority: <priority>
- URL: https://<domain>.atlassian.net/browse/<KEY>
- Reporter: <name>  ·  Assignee: <name>
- Comments: <N returned> / <N in header> — <verified | ⚠ MISMATCH>
- Comment ordering detected: <newest-first | oldest-first>
- Linked issues: <list with relationship + status, or "None">

## Prior Agreement
<!-- Generated by new Step 4.0. "No prior agreement detected" when the scan found no approval signal. -->
- Approval: <verbatim quote> — [comment: <author>, <date>] <⚠ approver is neither reporter nor assignee, if applicable>
- Referent proposal: [comment: <author>, <date>] — <what was proposed, one line>
- Agreed Scope Set: <scope ids/names explicitly covered by the referent>
- Adjacent, NOT agreed: <follow-on work mentioned in passing — carried as `consider`>

## Candidate Scopes
<!-- Decomposed from ticket + comments + research (new Step 4). Grouped by Kind, `defect` first; grouping is MANDATORY when > 10 rows. Readiness is an observable property, not a recommendation — there is deliberately no single-winner marker. AGREED outranks every heuristic. The formal IN/OUT boundary is DECIDED and GATED in proposal.md (after grillme) — this is the candidate set, not the decision. -->
| # | Kind | Scope | Root cause | Impact | Source | Readiness | Disposition |
|---|------|-------|------------|--------|--------|-----------|-------------|
| 1 | defect / recovery / safeguard / refuted / separate-ticket | <noun phrase> | <mechanism + file:line OR ticket quote> | <who/what, severity> | [ticket body] / [comment: author date] / [research: file:line] / [investigation: note] | ready-now / blocked-on-<X> / needs-sizing / needs-decision | ✅ AGREED [comment: author, date] "<quote>" / REFUTED (cite disproof) / consider  <⚠ OVERLAPS <change-name>, if applicable> |

### Refuted — boundaries
<!-- One row per REFUTED candidate. A refutation with no stated boundary is not established. -->
| # | Refuted claim | Disproof | Boundary — what this does NOT close | Era / Scope |
|---|---------------|----------|-------------------------------------|-------------|

### Selected for this change
<!-- From Step 4.2 multi-select. Each row becomes specs/<capability>/spec.md with its own spec:<capability> gate — in THIS change, not a sibling change. Order: ready-now before blocked. -->
| Order | Capability | Scope | Readiness | Blocked by |
|-------|------------|-------|-----------|------------|

## Open Questions ⚠️
<extracted from the source's Open Questions section, or "None noted in source">
<!-- grillme MUST address each item here -->

## Dependencies / Blockers
<extracted from the ticket, or "None noted in source">

## Ticket Quality Assessment
- Level: HIGH | MEDIUM | LOW  <+ failing axis, e.g. "MEDIUM (evidence-stale)">
- Checklist: [Description: Y/N, AC: Y/N, Scope: Y/N, Deps: Y/N, OQ: Y/N, Resolved: Y/N, Evidence Freshness: Y/N]
- Evidence provenance: <source + as-of date, e.g. "sanitised dump, as-of 2026-07-23"> — <columns/tables unusable and why, or "all figures reproducible against live data">
- Missing items handled by: (none / inline-provided / proceed-with-flag)

## Gap Manifest
<!-- Generated by new. Empty when Quality is HIGH or all gaps were inline-provided. -->
<!-- grillme MUST treat every row as a mandatory IN surface — cannot mark N/A without explicit justification. -->
| Gap | Grillme Surface | Priority | Elicit |
|-----|----------------|----------|--------|
| <failed item> | <surface name> | <1–3> | <what grillme must draw out from the user> |

## Actors & Use Cases
<!-- Generated by new Step 3.2. grillme anchors questions to (Actor, Use Case). -->
<!-- If the change is cosmetic/docs/test-only/dep-bump/pure-refactor: replace the tables below with a single line "N/A: <reason>" and skip the rest. -->

### Actors
| Actor | Type | Citation | Notes |
|-------|------|----------|-------|
| <name> | Human / System | [ticket: "<verbatim quote>"] or [description: "<verbatim quote>"] or [code: <file:line> <symbol>] | <one-line role> |

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

Target branch: <branch>
```

For non-Jira input (text / file / URL), set `Source:` in `## Ticket` accordingly and route the resolved description through scope extraction in Step 4 — do not paste it verbatim.

**`context.md` is always written**, even when no Jira key is provided. This file must exist for grillme to check the Gap Manifest.

### Step 7: Commit Initialized Artifacts

Stage and commit the SDLAIC change directory inside the **target project's repository**:

```bash
git -C projects/<project> add "$(sdlaic path change --change <change-name>)"
git -C projects/<project> commit -m "<prefix>: new context for <change-name>"
```

Commit prefix fallback chain (use first applicable):
1. **Jira key available** → `JIRA-1013: new context for add-user-export`
2. **Change slug only** → `add-user-export: new context`
3. **Neither** → `chore(sdlaic): new context for add-user-export`

This is a stable checkpoint. Future agents can roll back here if direction changes.

### Step 8: Handoff

```
Change "<change-name>" initialized.
- Ticket/Description Quality: {HIGH|MEDIUM|LOW}{ + failing axis, e.g. "(evidence-stale)"}
- Comments: {N}/{N} verified · ordering {newest-first|oldest-first}
- Prior agreement: {none detected | AGREED set of N, per <author> <date>}
- Candidate scopes: {N total} across {N} kinds · selected for this change: {N}
- Prior-art overlaps: {N} (or none)
- Open Questions captured: {N} (grillme will address each)
- Dependencies noted: {N}
- Actors extracted: {N} (or N/A — trivial change)
- Use cases extracted: {N total}

Next: Run grillme. Pay special attention to `## Prior Agreement`, `## Candidate Scopes`
(Kind grouping, Readiness values, `REFUTED` rows + their boundaries, any `⚠ OVERLAPS`),
`### Selected for this change`, `## Open Questions ⚠️`, `## Gap Manifest`, and
`## Actors & Use Cases` in context.md — grillme will pressure-test the candidate set,
treat every row in the manifest as a mandatory surface, and anchor each question to a
specific (Actor, Use Case). An `✅ AGREED` scope is not up for re-litigation: grillme may
probe HOW it is done, and may surface a genuinely new blocker, but not WHETHER to do it.
After grillme, `proposal` writes the gated IN/OUT boundary in `proposal.md`.
```

## Output Artifacts

- `$(sdlaic path change --change <change-name>)/` — change directory created by SDLAIC
- `$(sdlaic path change --change <change-name>)/context.md` — curated scope analysis from the ticket + research (always written; contains `## Prior Agreement`, `## Candidate Scopes` with Kind + Readiness + dispositions and `### Refuted — boundaries`, `### Selected for this change` — the gated IN/OUT decision lives in `proposal.md`, not here — plus Gap Manifest when quality is MEDIUM/LOW + Option C, and `## Actors & Use Cases` with populated tables or `N/A` for trivial changes)

## Verification

- [ ] SDLAIC change directory exists
- [ ] Research Summary was output before handoff
- [ ] User confirmed the change name
- [ ] Comment count returned matches the ticket header count (or the mismatch was raised and resolved before proceeding); ordering was detected and stated
- [ ] `context.md > ## Prior Agreement` exists — either an Agreed Scope Set with author + date + verbatim quote, or `No prior agreement detected`
- [ ] Every `✅ AGREED` row carries `[comment: <author>, <date>]` + a verbatim quote; rows whose approver is neither reporter nor assignee are flagged
- [ ] No scope was argued against on impact/effort/priority grounds after being marked `AGREED` (a newly discovered blocker may be surfaced, but the row stays `AGREED`)
- [ ] `context.md > ## Candidate Scopes` is a COMPLETE table — every issue/hypothesis in the ticket decomposed, each row sourced, with Kind, Readiness, and Disposition columns
- [ ] No single-winner marker anywhere — the steer is carried by `Readiness`, and every row has a value
- [ ] Rows are grouped by `Kind` with `defect` first (grouping is MANDATORY when > 10 rows)
- [ ] Every `REFUTED` row appears in `### Refuted — boundaries` with a stated Boundary and Era/Scope (no boundary → the row should have been downgraded to `consider`)
- [ ] `sdlaic list` prior-art check was run; any overlap is tagged `⚠ OVERLAPS <change-name>` on the candidate row and cited in the Research Summary
- [ ] Scope confirmation was **multi-select**, and `### Selected for this change` lists the selected capabilities ordered `ready-now` before blocked
- [ ] `sdlaic new change` was run **once** — unless a Step 4.2 split criterion fired and was named explicitly
- [ ] `context.md` records NO final IN/OUT decision — it states the boundary is decided+gated in `proposal.md` (after grillme). Candidate set + selection only.
- [ ] `context.md > ## Open Questions ⚠️` exists (with content or "None noted in source")
- [ ] `context.md > ## Dependencies / Blockers` exists (with content or "None noted in source")
- [ ] `context.md > ## Ticket Quality Assessment` exists with Level (+ failing axis) + 7-item Checklist + Evidence provenance + Missing items handled by
- [ ] Evidence Freshness was assessed: every quantitative claim either traced to a live source, or the snapshot date + unusable columns recorded and the affected scope marked `needs-sizing`
- [ ] `context.md > ## Gap Manifest` exists (populated when quality is MEDIUM/LOW + Option C; empty table rows otherwise)
- [ ] `context.md > ## Actors & Use Cases` exists (populated tables OR explicit `N/A: <reason>` for trivial changes)
- [ ] Every actor row has a citation — `[ticket: "..."]` or `[description: "..."]` or `[code: <file:line> <symbol>]` — no uncited rows
- [ ] Every use case has Happy path + ≥1 Alternate or Edge case (no happy-path-only rows)
- [ ] Caps respected: ≤ 6 actors, ≤ 3 primary use cases per actor
- [ ] If `## Gap Manifest` is populated, every gap is cross-linked to ≥1 actor/use case in `## Gap Manifest Cross-link`
- [ ] Initialized artifacts committed to the repo holding the changes directory (`sdlaic path change --change <change-name>`)

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Fetching only the description, not comments | Comments often contain the real scope, refuted hypotheses, and decisions. Always fetch both — comments are a primary source for Step 4 scope extraction. |
| Using `jira issue comment list` | No such subcommand (`jira issue comment` has only `add`). Use `jira issue view <KEY> --comments <N> --plain`. |
| Accepting whatever comments came back | Bare `--plain` renders only the LATEST comment. Compare the count returned against the header count and HALT on mismatch — the missing comments are usually where the decision lives. |
| Assuming comments are chronological | This CLI commonly returns **newest-first**. Detect ordering from timestamps and state it. Reading the trail backwards inverts which comment supersedes which. |
| Re-litigating a scope the ticket already approved | Run Step 4.0 first. `AGREED` outranks every heuristic — you may surface a newly discovered blocker, never argue the scope down on impact/effort/priority. |
| Treating an approval as self-evident | Always record `[comment: <author>, <date>]` + the verbatim quote, and flag when the approver is neither reporter nor assignee. The user judges authority; you supply the evidence. |
| Extending `AGREED` to follow-on work mentioned in passing | Only scopes in the approved referent proposal are `AGREED`. Adjacent work gets `consider` + `adjacent to agreed scope — not explicitly approved`, raised as a question. Inflating an agreement is as wrong as ignoring one. |
| Skipping code research | Research prevents rework. Always run it. |
| Skipping the `sdlaic list` prior-art check | An existing change may already specify — or have shipped — the remedy you are proposing. Tag `⚠ OVERLAPS <change-name>` and cite its artifact. |
| Creating the change before confirming the name | Always confirm with the user first. |
| Proceeding without any context | If Jira fetch fails and user can't provide context, STOP. Don't guess. |
| Ignoring the project path | Research without project scope returns irrelevant results. Ask which project. |
| Dumping the whole ticket body + comments verbatim into `context.md` | Decompose into candidate scopes (Step 4.1) with Kind + Readiness + disposition. `context.md` holds the candidate set + selection only — the gated IN/OUT decision is `proposal`'s job, not here. |
| Making the final IN/OUT scope decision in `new` | Don't. `new` recommends; `proposal` decides (and is grilled + gated). Recording a final In/Out here creates a second, un-gated source of truth that drifts from `proposal.md`. |
| Silently dropping a candidate from the `## Candidate Scopes` table | Every issue/hypothesis in the ticket must appear as a row (tag `REFUTED` if disproved). The table is the complete record; the IN/OUT *split* happens later in `proposal`. |
| Cutting rows to make a long table manageable | Group by `Kind` instead (mandatory > 10 rows). The fix for 25 candidates is 25 rows with 4 defects surfaced first — not 8 rows. Reduce display noise, never row count. |
| Losing the pinned code anchor when distilling a candidate | Every candidate row keeps root cause → `file:line` where known. Distillation removes prose, not evidence — grillme, proposal, spec, and review all need the anchor. |
| Picking one "best" scope and marking it recommended | There is no single-winner marker. Give every row a `Readiness` value instead — an observable property a reader can check and disagree with, rather than an opinion that silently outranks stakeholder agreement. |
| Asking the user to pick ONE scope | Confirmation is multi-select. A change carries N capabilities; single-select structurally discards the rest. |
| Creating one change per defect | N scopes → N `specs/<capability>/` in ONE change. Split only on a Step 4.2 delivery-unit criterion, and name which one. "Different defects" is not a criterion. |
| Recording a `REFUTED` row with no boundary | State what the refutation does NOT close, plus its era/scope. An over-broad refutation retires live defects by association. No boundary → downgrade to `consider`. |
| Skipping the Ticket Quality Check because the ticket "looks fine" | Always run the 7-item checklist in Step 1.2. The cost is seconds; the cost of a missed gap is rework downstream (gap discovered during manual testing post-review). |
| Scoring a ticket HIGH while its numbers are unreproducible | "Well-written" and "evidence-stale" are different axes. Run Evidence Freshness separately and report the failing axis alongside the level. |
| Listing "User" / "Admin" as generic actors with no citation | Drop. Either cite the ticket's exact wording (`[ticket: "..."]`) or description (`[description: "..."]`) or cite the code symbol (`[code: <file:line> <path/to/file>]`). Generic uncited actors are exactly what the provenance bar exists to block. |
| Inventing actors from "what users might want" | Provenance bar. No source → no actor. If the change introduces a new actor type, tag `[NEW SURFACE]` on the use case rather than fabricating a citation. |
| Forcing the Actors & Use Cases section onto a trivial change (typo, dep bump, pure refactor) | Use `N/A: <one-line reason>` for the whole section. The extraction has a cost; skip it when the change has no behavior surface. |
| Nesting side-effect emails/webhooks inside the State Transitions table | State Transition rows are only for persisted state changes (DB / cache / file). Notifications, webhook dispatches, async fan-out with no persisted change go in the separate "Side-effect-only Actions" list. |
| Producing happy-path-only use cases | Every use case needs ≥1 Alternate or Edge case (or a one-line note why none exist). Happy-path-only rows give grillme nothing to anchor on. |
| Treating the Actors section as the SDLAIC `spec.md` | This is lightweight pre-design extraction for grillme anchoring. The formal WHEN/THEN spec is `spec.md`, written later in PROPOSED phase. Do not duplicate or pre-write spec clauses here. |

## Handoff

Route to `grillme` to challenge assumptions before design begins.

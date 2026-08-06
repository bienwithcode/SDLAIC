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
  # Get description, status, priority, assignee, labels
  jira issue view <KEY> --plain

  # Get all comments (may contain clarifications, decisions, acceptance criteria)
  jira issue comment list <KEY> --plain
  ```
  From the output, extract issue metadata (key, summary, type, status, priority, URL) plus the full description and all comments as raw source material for scope extraction in Step 4. The description and comments are SOURCES to decompose into candidate scopes — they are NOT pasted wholesale into context.md. Do NOT proceed without context.
  
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

Compute quality:
- **HIGH** = 6/6 pass
- **MEDIUM** = 4–5/6
- **LOW** = ≤ 3/6

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

The result is recorded in `context.md` under `## Ticket Quality Assessment` and `## Gap Manifest` in Step 6.

If quality is HIGH, this step passes silently — do not bother the user.

### Step 2: Codebase & Web Research

Run research queries to gather context about the change. Tool selection follows your workspace convention (see `references/code-research.md`).

1. **Local Codebase Research**:
   - Run a targeted codebase query on the topic from the resolved description, scoped to `projects/<project>` (or a broader query if the project path is unclear).
   - Look for existing precedents, related configurations, or circular dependencies.
   
2. **External / Web Research**:
   - If the resolved description mentions any third-party APIs, external libraries, protocols, or services (e.g., "Antigravity provider", external SDKs, or integrations) that are new or have no established local codebase precedents:
     * Perform a web search to find official documentation, API specifications, or known limitations for these technologies.
     * Do not guess or assume how external services behave.
     * Record key web research findings (including URL citations) in the Research Summary.

Empty result is valid — record "no codebase precedent or web documentation found for `<terms>`" and proceed. Do not silently skip.

### Step 3: Research Output & Actors Extraction

#### Step 3.1: Output Research Summary

Produce a concise summary covering:
- **What exists (Codebase):** Relevant code, patterns, and architecture found locally.
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

### Step 4: Scope Extraction & Recommendation (one change = one PR)

One SDLAIC change = one PR. A ticket often bundles multiple issues, hypotheses, and follow-ons. Do NOT carry the whole ticket into `context.md`. Decompose it into **candidate scopes** and **recommend** which to do first. `new` does NOT make the final IN/OUT decision — that is decided and **gated in `proposal.md`** (after `grillme` challenges the candidates). Here you only record the candidate set + a recommendation; nothing is silently dropped.

**Inputs:** the resolved description + comments (Step 1.1), the research findings (Step 2 / Step 3.1), and the extracted actors (Step 3.2).

#### Step 4.1: Extract candidate scopes

Re-read the description, every comment, and the investigation output as a single search space. Extract each distinct issue, hypothesis, requested behavior, or proposed fix as a separate candidate scope. A candidate scope is anything that could plausibly be its own PR.

For each candidate scope, determine:
- **Scope** — a short noun phrase naming the issue/behavior (e.g., "Roster leaks cross-team members", "Add export-to-CSV button").
- **Root cause** — the underlying mechanism, if known. Cite `file:line` from research when one exists; otherwise quote the ticket/comment that asserts it.
- **Impact** — who/what is affected and how severely.
- **Source** — where it came from: `[ticket body]`, `[comment: <author> <date>]`, `[research: file:line]`, or `[investigation: <note>]`.
- **Disposition** — `consider` by default; `REFUTED` if a later comment or research disproves it (cite the disproof). Step 4.2 upgrades the strongest one(s) to `🌟 RECOMMENDED`.

**Refuted hypotheses.** A common case: the ticket body or an early comment proposes a cause or fix that a later comment or the codebase research disproves. Keep these as candidate scopes but tag them `REFUTED` with the disproof citation. They are high-value signal — they record a dead-end the team already ruled out, so a future agent does not re-walk it.

**Provenance bar.** Every candidate scope must cite its source. No source → drop it. Do not invent scopes from "what might be wrong."

**Caps.** Aim for the real set; merge near-duplicates. If extraction yields > 10 candidates, group related ones and note the grouping.

#### Step 4.2: Present candidates and recommend

Present the candidate scopes as a numbered list or table showing the fields (scope | root cause | impact | source | disposition). Use whichever prompt or confirmation capability your runtime provides — do not assume a specific tool name.

**Recommend a primary scope.** Do NOT present the list neutrally — mark the scope (or small set) you recommend for THIS change with `🌟 RECOMMENDED` and a one-line rationale grounded in **impact × readiness × risk**: prefer the scope that is high-value, ready to fix now, and low-risk (e.g., "🌟 RECOMMENDED: bundleID leak — live defect, ~1-line fix, highest value-per-hour, reversible"). The recommendation is a suggestion, not a decision.

Frame the question explicitly:

> Here are the candidate scopes I extracted from the ticket, comments, and investigation. One SDLAIC change = one PR. 🌟 = my recommendation. Confirm the focus for THIS change (it informs the change name). The formal IN/OUT boundary is decided later in `proposal.md`, after `grillme` challenges these candidates — so this is a recommendation, not the final scope decision.
>
> 🌟 1. <scope> — <root cause> — <impact> — <source>  (recommended: <one-line reason>)
> 2. <scope> — <root cause> — <impact> — <source>  (<disposition>)

Rules:
- Confirm a focus with the user to derive the change name — typically the recommended scope(s). A typical change has 1–3 in-scope scopes.
- If the user wants everything, flag that large multi-scope tickets usually want splitting into multiple SDLAIC changes.
- If the user adds a scope NOT in the candidate list, add it with source `[user]` and disposition `consider`.
- Always include a recommendation with grounded rationale, even if the user is likely to override — a neutral list with no steer gives them nothing to react to.
- Do NOT record a final IN/OUT decision or "not-selected with reasons" here — that is `proposal`'s gated job. Record only the candidate set + dispositions + recommendation in `context.md`.

#### Step 4.3: Prune dependent extractions

If any Actor or Use Case extracted in Step 3.2 belongs exclusively to a candidate the recommendation steers away from (a `REFUTED` or clearly lower-priority scope), drop or relocate that row before writing `context.md` in Step 6. Actors and use cases should track the recommended focus — but keep it light, since the final IN/OUT line is settled in `proposal`.

### Step 5: Initialize SDLAIC Change

```bash
sdlaic new change "<change-name>"
```

Derive the change name:
- From Jira summary if available (convert to kebab-case)
- From user description if no Jira key
- Confirm with user **after** reviewing the research summary AND the Step 4 recommendation — the name must reflect the recommended focus, not the whole ticket

### Step 6: Write context.md

Write the gathered context (ticket or non-ticket source) to `context.md` using the template below. Substitute the candidate scopes (with dispositions + 🌟 recommendation) from Step 4 and the actors from Step 3.2:

```markdown
# Change Context

## Ticket
- Source: <Jira key OR "Text input" OR "File: <path>" OR "URL: <url>">
- Key: <KEY>
- Title: <summary>
- Type: <type>  ·  Status: <status>  ·  Priority: <priority>
- URL: https://<domain>.atlassian.net/browse/<KEY>
- Linked issues: <list with relationship + status, or "None">

## Candidate Scopes
<!-- Decomposed from ticket + comments + research (new Step 4). 🌟 = recommended scope(s); REFUTED = hypothesis the investigation disproved (cite disproof). The formal IN/OUT boundary is DECIDED and GATED in proposal.md (after grillme) — this is the candidate set + recommendation, not the decision. -->
| # | Scope | Root cause | Impact | Source | Disposition |
|---|-------|------------|--------|--------|-------------|
| 1 | <noun phrase> | <mechanism + file:line OR ticket quote> | <who/what, severity> | [ticket body] / [comment: author date] / [research: file:line] / [investigation: note] | 🌟 RECOMMENDED / REFUTED (cite disproof) / consider |

## Open Questions ⚠️
<extracted from the source's Open Questions section, or "None noted in source">
<!-- grillme MUST address each item here -->

## Dependencies / Blockers
<extracted from the ticket, or "None noted in source">

## Ticket Quality Assessment
- Level: HIGH | MEDIUM | LOW
- Checklist: [Description: Y/N, AC: Y/N, Scope: Y/N, Deps: Y/N, OQ: Y/N, Resolved: Y/N]
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
- Ticket/Description Quality: {HIGH|MEDIUM|LOW}
- Open Questions captured: {N} (grillme will address each)
- Dependencies noted: {N}
- Actors extracted: {N} (or N/A — trivial change)
- Use cases extracted: {N total}

Next: Run grillme. Pay special attention to `## Candidate Scopes` (the `🌟 RECOMMENDED`
marker + any `REFUTED` tags), `## Open Questions ⚠️`, `## Gap Manifest`, and
`## Actors & Use Cases` in context.md — grillme will pressure-test the candidate set
and the recommendation's rationale, treat every row in the manifest as a mandatory
surface, and anchor each question to a specific (Actor, Use Case). After grillme,
`proposal` writes the gated IN/OUT boundary in `proposal.md`.
```

## Output Artifacts

- `$(sdlaic path change --change <change-name>)/` — change directory created by SDLAIC
- `$(sdlaic path change --change <change-name>)/context.md` — curated scope analysis from the ticket + research (always written; contains `## Candidate Scopes` with a `🌟 RECOMMENDED` marker + `REFUTED` tags — the gated IN/OUT decision lives in `proposal.md`, not here — plus Gap Manifest when quality is MEDIUM/LOW + Option C, and `## Actors & Use Cases` with populated tables or `N/A` for trivial changes)

## Verification

- [ ] SDLAIC change directory exists
- [ ] Research Summary was output before handoff
- [ ] User confirmed the change name
- [ ] A `🌟 RECOMMENDED` scope was presented with a grounded one-line rationale (impact × readiness × risk)
- [ ] `context.md > ## Candidate Scopes` is a COMPLETE table — every issue/hypothesis in the ticket decomposed, each row sourced, with a Disposition column; ≥1 row marked `🌟 RECOMMENDED` (with grounded rationale) and any `REFUTED` row citing its disproof
- [ ] `context.md` records NO final IN/OUT decision — it states the boundary is decided+gated in `proposal.md` (after grillme). Candidate set + recommendation only.
- [ ] `context.md > ## Open Questions ⚠️` exists (with content or "None noted in source")
- [ ] `context.md > ## Dependencies / Blockers` exists (with content or "None noted in source")
- [ ] `context.md > ## Ticket Quality Assessment` exists with Level + Checklist + Missing items handled by
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
| Skipping code research | Research prevents rework. Always run it. |
| Creating the change before confirming the name | Always confirm with the user first. |
| Proceeding without any context | If Jira fetch fails and user can't provide context, STOP. Don't guess. |
| Ignoring the project path | Research without project scope returns irrelevant results. Ask which project. |
| Dumping the whole ticket body + comments verbatim into `context.md` | Decompose into candidate scopes (Step 4.1) with a 🌟 recommendation. `context.md` holds the candidate set + recommendation only — the gated IN/OUT decision is `proposal`'s job, not here. |
| Making the final IN/OUT scope decision in `new` | Don't. `new` recommends; `proposal` decides (and is grilled + gated). Recording a final In/Out here creates a second, un-gated source of truth that drifts from `proposal.md`. |
| Silently dropping a candidate from the `## Candidate Scopes` table | Every issue/hypothesis in the ticket must appear as a row (tag `REFUTED` if disproved). The table is the complete record; the IN/OUT *split* happens later in `proposal`. |
| Losing the pinned code anchor when distilling a candidate | Every candidate row keeps root cause → `file:line` where known. Distillation removes prose, not evidence — grillme, proposal, spec, and review all need the anchor. |
| Presenting candidate scopes neutrally with no recommendation | Always mark a `🌟 RECOMMENDED` scope with a grounded rationale (impact × readiness × risk). A neutral list gives the user nothing to react to — the recommendation is a suggestion they can override. |
| Skipping the Ticket Quality Check because the ticket "looks fine" | Always run the 6-item checklist in Step 1.2. The cost is seconds; the cost of a missed gap is rework downstream (gap discovered during manual testing post-review). |
| Listing "User" / "Admin" as generic actors with no citation | Drop. Either cite the ticket's exact wording (`[ticket: "..."]`) or description (`[description: "..."]`) or cite the code symbol (`[code: <file:line> <path/to/file>]`). Generic uncited actors are exactly what the provenance bar exists to block. |
| Inventing actors from "what users might want" | Provenance bar. No source → no actor. If the change introduces a new actor type, tag `[NEW SURFACE]` on the use case rather than fabricating a citation. |
| Forcing the Actors & Use Cases section onto a trivial change (typo, dep bump, pure refactor) | Use `N/A: <one-line reason>` for the whole section. The extraction has a cost; skip it when the change has no behavior surface. |
| Nesting side-effect emails/webhooks inside the State Transitions table | State Transition rows are only for persisted state changes (DB / cache / file). Notifications, webhook dispatches, async fan-out with no persisted change go in the separate "Side-effect-only Actions" list. |
| Producing happy-path-only use cases | Every use case needs ≥1 Alternate or Edge case (or a one-line note why none exist). Happy-path-only rows give grillme nothing to anchor on. |
| Treating the Actors section as the SDLAIC `spec.md` | This is lightweight pre-design extraction for grillme anchoring. The formal WHEN/THEN spec is `spec.md`, written later in PROPOSED phase. Do not duplicate or pre-write spec clauses here. |

## Handoff

Route to `grillme` to challenge assumptions before design begins.

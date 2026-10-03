# Change Context

## Ticket
- Source: <Jira key OR "Text input" OR "File: <path>" OR "URL: <url>">
- Key: <KEY>
- Title: <summary>
- Type: <type>  ·  Status: <status>  ·  Priority: <priority>
- URL: <url>
- Reporter: <name>  ·  Assignee: <name>
- Comments: <N returned> / <N in header> — <verified | ⚠ MISMATCH>
- Comment ordering detected: <newest-first | oldest-first>
- Linked issues: <list with relationship + status, or "None">

## Prior Agreement
<!-- "No prior agreement detected" when the scan found no approval signal. -->
- Approval: <verbatim quote> — [comment: <author>, <date>] <⚠ approver is neither reporter nor assignee, if applicable>
- Referent proposal: [comment: <author>, <date>] — <what was proposed, one line>
- Agreed Scope Set: <scope ids/names explicitly covered by the referent>
- Adjacent, NOT agreed: <follow-on work mentioned in passing — carried as `consider`>

## Candidate Scopes
<!-- Decomposed from ticket + comments + research (new Step 4). Grouped by Kind, `defect` first; grouping is MANDATORY when > 10 rows. Readiness is an observable property, not a recommendation — there is deliberately no single-winner marker. AGREED outranks every heuristic. The formal IN/OUT boundary is DECIDED and GATED in proposal.md (after grillme) — this is the candidate set, not the decision. -->
| # | Kind | Scope | Root cause | Impact | Source | Readiness | Disposition |
|---|------|-------|------------|--------|--------|-----------|-------------|
| 1 | <defect / recovery / safeguard / refuted / separate-ticket> | <noun phrase> | <mechanism + file:line OR ticket quote> | <who/what, severity> | [ticket body] / [comment: <author> <date>] / [research: <file:line>] / [investigation: <note>] | <ready-now / blocked-on-X / needs-sizing / needs-decision> | ✅ AGREED [comment: <author>, <date>] "<quote>" / REFUTED (cite disproof) / consider <⚠ OVERLAPS <change-name>, if applicable> |

### Refuted — boundaries
<!-- One row per REFUTED candidate. A refutation with no stated boundary is not established — downgrade the row to `consider`. -->
| # | Refuted claim | Disproof | Boundary — what this does NOT close | Era / Scope |
|---|---------------|----------|-------------------------------------|-------------|

### Selected for this change
<!-- From the multi-select confirmation. Each row becomes specs/<capability>/spec.md with its own spec:<capability> gate — in THIS change, not a sibling change. Order: ready-now before blocked. -->
| Order | Capability | Scope | Readiness | Blocked by |
|-------|------------|-------|-----------|------------|

## Open Questions ⚠️
<open questions, or "None noted in source">

## Dependencies / Blockers
<dependencies, or "None noted in source">

## Ticket Quality Assessment
- Level: <HIGH | MEDIUM | LOW> <+ failing axis, e.g. "MEDIUM (evidence-stale)">
- Checklist: [Description: Y/N, AC: Y/N, Scope: Y/N, Deps: Y/N, OQ: Y/N, Resolved: Y/N, Evidence Freshness: Y/N]
- Evidence provenance: <source + as-of date> — <columns/tables unusable and why, or "all figures reproducible against live data">
- Missing items handled by: (none / inline-provided / proceed-with-flag)

## Gap Manifest
<!-- Empty when Quality is HIGH or all gaps inline-provided. -->
| Gap | Grillme Surface | Priority | Elicit |
|-----|----------------|----------|--------|
| <failed item> | <surface name> | <1-3> | <what grillme must draw out from the user> |

## Actors & Use Cases
<!-- N/A: <reason> for cosmetic/docs/test-only/dep-bump/pure-refactor changes. -->

### Actors
| Actor | Type | Citation | Notes |
|-------|------|----------|-------|
| <name> | Human / System | [ticket: "<verbatim quote>"] / [description: "<verbatim quote>"] / [code: <file:line> <symbol>] | <one-line role> |

### Use Cases
#### <Actor>
- **Happy path:** <Trigger → Action → Object → Outcome> [citation]
- **Alternate paths:** <variation> [citation]
- **Edge cases:** <case> [citation or NEW SURFACE]

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

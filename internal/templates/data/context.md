# Context

## Ticket
- Source: <Jira key OR "Text input" OR "File: <path>" OR "URL: <url>">
- Key: <KEY>
- Title: <summary>
- Type: <type>  ·  Status: <status>  ·  Priority: <priority>
- URL: <url>
- Linked issues: <list, or "None">

## Candidate Scopes
<!-- Decomposed from ticket + comments + research (new Step 4). 🌟 = recommended; REFUTED = disproved hypothesis. The gated IN/OUT boundary is decided in proposal.md (after grillme) — this is the candidate set + recommendation, not the decision. -->
| # | Scope | Root cause | Impact | Source | Disposition |
|---|-------|------------|--------|--------|-------------|
| 1 | <scope> | <cause + file:line or quote> | <impact> | [source] | 🌟 RECOMMENDED / REFUTED (cite disproof) / consider |

## Open Questions ⚠️
<open questions, or "None noted in source">

## Dependencies / Blockers
<dependencies, or "None noted in source">

## Ticket Quality Assessment
- Level: HIGH | MEDIUM | LOW
- Checklist: [Description: Y/N, AC: Y/N, Scope: Y/N, Deps: Y/N, OQ: Y/N, Resolved: Y/N]
- Missing items handled by: (none / inline-provided / proceed-with-flag)

## Gap Manifest
<!-- Empty when Quality is HIGH or all gaps inline-provided. -->
| Gap | Grillme Surface | Priority | Elicit |
|-----|----------------|----------|--------|
| <gap> | <surface> | <1-3> | <elicit> |

## Actors & Use Cases
<!-- N/A: <reason> for cosmetic/docs/test-only/dep-bump/pure-refactor changes. -->

### Actors
| Actor | Type | Citation | Notes |
|-------|------|----------|-------|
| <name> | Human / System | [citation] | <role> |

### Use Cases
#### <Actor>
- **Happy path:** <Trigger → Action → Object → Outcome> [citation]
- **Alternate/Edge:** <case> [citation]

### State Transitions
| Entity | From | Event | To | Side effects |
|--------|------|-------|----|--------------|

### Side-effect-only Actions
- <Actor> → <action> → <effect> [citation]

### Gap Manifest Cross-link
| Gap | Affected Actor(s) | Affected Use Case(s) |
|-----|-------------------|---------------------|

Target branch: <branch>

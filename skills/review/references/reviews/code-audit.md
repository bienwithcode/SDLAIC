# Code Audit (Final Pre-PR Gate)

> Loaded by `skills/review` for `review code` — the terminal audit after `apply`.
> Auditor role: Multi-Pass Code Reviewer (built for 30+ file PRs). Unlike the four
> artifact gates, this audit records its outcome in `review.md`; it does not gate a
> downstream phase.

## Two-pass structure

Run two independent passes, each in a **clean-context subagent** spawned in a
single message so they run in parallel. Do NOT perform review analysis in the
main context — the subagents do all assessment.

### Pass A — Compliance (persona: `references/personas/compliance-reviewer.md`)

Extract the markdown body of `references/personas/compliance-reviewer.md` (everything after the
YAML frontmatter) as the subagent prompt. Append the change contracts and diff:
`proposal.md`, `specs/<capability>/spec.md` (if any), `design.md`, `tasks.md`, the
git diff (`--stat` + full or `--name-only` per the diff-size rule in
`skills/review`), and the git log. The pass audits whether the implementation
matches the contract.

### Pass B — Code Quality (persona: `references/personas/code-quality-reviewer.md`)

Extract the markdown body of `references/personas/code-quality-reviewer.md` as the subagent
prompt. Append the changed files, full diff, and `tasks.md`. The pass runs a
CURe-style quality assessment (DRY, module boundaries, pattern drift,
architecture, security, performance, test coverage).

## Claim Verification Rule

Every finding from either pass MUST cite primary evidence — `path:line` in the
diff/codebase or a quoted contract line. Unevidenced findings are dropped.

## Verdict rules

Each pass returns an independent verdict (APPROVE / REQUEST_CHANGES / REJECT):

- **APPROVE** — both passes APPROVE.
- **REQUEST_CHANGES** — at least one pass requests changes; none REJECT.
- **REJECT** — either pass REJECTs (contract violation or critical quality defect).

Record the merged result and both pass outputs verbatim in
`$(sdlaic path change --change <name>)/review.md` (see `skills/review` for the full template).

## Severity guide

| Severity | Example |
|----------|---------|
| CRITICAL | Implementation contradicts an approved contract; security defect; data loss. |
| HIGH | Missing test coverage for a success criterion; scope violation; broken module boundary. |
| MEDIUM | DRY violation; pattern drift; weak error handling. |
| LOW / INFO | Naming, formatting, documentation. |

## Scope discipline

- Every changed file should map to a task in `tasks.md`; flag files changed
  outside the planned scope.
- The implementation must not add capability beyond the approved proposal.

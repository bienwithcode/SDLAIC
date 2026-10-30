# Code Audit (Final Pre-PR Gate)

> Loaded by `skills/review` for `review code` — the terminal audit after `apply`.
> Auditor role: Multi-Pass Code Reviewer (built for 30+ file PRs). Unlike the four
> artifact gates, this audit records its outcome in `review.md`; it does not gate a
> downstream phase.

## Two-pass structure

Run two independent passes, each in a **clean-context subagent** spawned in a
single message so they run in parallel. Do NOT perform review analysis in the
main context — the subagents do all assessment.

### Pass A — Compliance (agent: `agents/compliance-reviewer.md`)

Extract the markdown body of `agents/compliance-reviewer.md` (everything after the
YAML frontmatter) as the subagent prompt. Append the change contracts and diff:
`proposal.md`, `specs/<capability>/spec.md` (if any), `design.md`, `tasks.md`, the
git diff (`--stat` + full or `--name-only` per the diff-size rule in
`skills/review`), and the git log. The pass audits whether the implementation
matches the contract.

### Pass B — Code Quality (agent: `agents/code-quality-reviewer.md`)

Extract the markdown body of `agents/code-quality-reviewer.md` as the subagent
prompt. Append the changed files, full diff, and `tasks.md`. The pass runs a
CURe-style quality assessment (DRY, module boundaries, pattern drift,
architecture, security, performance, test quality).

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

## Test quality (both passes)

Test value is proven by what fails when code breaks — not by test count. Judge
the suite guarantee-first:

- **Guarantee coverage:** every spec guarantee has exactly one test entry at the
  lowest level that proves it; entry points assert wiring only.
- **No source-inspecting tests:** tests observe behavior from outside (HTTP
  codes, emitted events, DB rows) — never grep method names or string-compare
  queries/SQL. Such tests break on every refactor and block nothing.
- **Hardcoded expectations:** expected values are literals in the test — a test
  that loops a production constant passes even when the constant is emptied
  (mutation trap).
- **Duplication:** the same rule re-tested at multiple layers is duplication,
  not extra safety.
- **Fixtures:** one shared fixture builder + one doubles file per area; tests
  clean up the DB rows they create.
- **Redundant growth:** a review round answered with a NEW suite instead of a
  row in an existing table is a finding against the PR, not extra credit.

"Missing tests" and "redundant tests" are both findings — do not only hunt for
the former.

## Severity guide

| Severity | Example |
|----------|---------|
| CRITICAL | Implementation contradicts an approved contract; security defect; data loss. |
| HIGH | Missing test entry for a guarantee; source-inspecting test; test deriving expectations from production constants; scope violation; broken module boundary. |
| MEDIUM | Same rule re-tested at multiple layers; per-file duplicate fixtures/doubles; tests that don't clean up their data; DRY violation; pattern drift; weak error handling. |
| LOW / INFO | Naming, formatting, documentation. |

## Scope discipline

- Every changed file should map to a task in `tasks.md`; flag files changed
  outside the planned scope.
- The implementation must not add capability beyond the approved proposal.

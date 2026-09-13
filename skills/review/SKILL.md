---
name: review
description: Parameterized auditor & gatekeeper engine. Runs AFTER an artifact is drafted (proposal | spec | design | tasks | code), loading the matching references/reviews/<phase>-audit.md in a clean subagent and issuing an APPROVE / REQUEST_CHANGES / REJECT verdict. Verdicts are persisted via `sdlaic gate set` — never inside the repo. For `code`, runs the two-pass compliance + quality review.
---

# Auditor & Gatekeeper Engine (Review)

## Core Principle

Review is an **independent gate**. It runs AFTER a draft skill, in a clean-context
subagent, loading only the target artifact + its phase-specific audit checklist —
preventing hallucination from accumulated chat history. It issues one of
**APPROVE / REQUEST_CHANGES / REJECT** and enforces the **Claim Verification
Rule**: every finding cites primary evidence (`path:line` or ticket reference).

## Parameters & Gate Recording

Invoked as `review <phase>` where `<phase>` ∈ `proposal | spec | design | tasks | code`:

| Phase | Audit checklist | Auditor role |
|-------|-----------------|--------------|
| `proposal` | `references/reviews/proposal-audit.md` | Business Analyst / PM |
| `spec` | `references/reviews/spec-audit.md` | QA Lead |
| `design` | `references/reviews/design-audit.md` | System Architect |
| `tasks` | `references/reviews/plan-audit.md` | Tech Lead / Scrum Master |
| `code` | `references/reviews/code-audit.md` | Multi-Pass Code Reviewer (two-pass, below) |

**Recording the verdict (all phases):** the verdict is persisted to the global
gate-state store — NEVER as a marker inside the reviewed artifact or the repo:

```bash
sdlaic gate set --phase <proposal|spec:<capability>|design|tasks> \
  --status <approved|failed> --verdict <APPROVE|REQUEST_CHANGES|REJECT>
```

`APPROVE → approved` (unlocks the next phase); `REQUEST_CHANGES | REJECT → failed`
(the draft skill re-drafts and increments the attempt).

For `spec`, the gate is **per-capability**: record `--phase spec:<capability>` once
per `specs/<capability>/` directory (e.g. `spec:auth`, `spec:billing`).

**Workflow toggle:** in `light`/`free`, review is skipped and the enforcer marks
the gate `skipped`.

## Artifact-gate flow (proposal | spec:<capability> | design | tasks)

1. Spawn a clean-context subagent (`subagent_type="general-purpose"`) with only:
   the target artifact, its upstream approved artifact(s), and the phase's audit
   checklist (`references/reviews/<phase>-audit.md`).
2. The subagent audits against the checklist, applying the Claim Verification
   Rule, and returns a verdict + findings (each with `path:line`/ticket evidence).
3. Persist the verdict via `sdlaic gate set` (above). Do NOT write it into the repo.
4. On `failed`, hand findings back to the draft skill for rework. Findings are
   feedback for the re-draft only — never written into the artifact; the
   `sdlaic gate set` verdict and git history already carry review outcomes.

## Code-gate flow (`review code`)

The `code` phase is the terminal pre-PR audit after `apply`. It runs the
**two-pass** review below (compliance + quality) and its outcome is recorded in
`review.md` (Step 3). It does not gate a downstream phase, so no `sdlaic gate set`
call is required for `code`.

## Two-pass code review

Review is a two-pass gate. Pass 1 checks whether implementation matches the contract (proposal + design). Pass 2 checks whether the code is architecturally sound and would survive a CURe PR review. Both passes produce independent verdicts. Each pass runs in a separate subagent with clean context to prevent hallucination from accumulated chat history.

## Pre-conditions

- SDLAIC change is in IMPLEMENTED state
- All tasks in `tasks.md` are marked complete
- Code has been committed

## Process

### Step 1: Gather Context

Read all artifacts in the order they were authored:
1. `proposal.md` — business scope, what changes, impact
2. `specs/<capability>/spec.md` (if exists) — formal requirements with WHEN/THEN scenarios
3. `design.md` — technical architecture, decisions, goals/non-goals (built to satisfy proposal + spec)
4. `tasks.md` — what was planned and its completion status

Determine the base branch first using this 3-step rule:
1. Explicit `Target branch:` in `context.md`
2. Branch name convention: `hotfix/*` | `release/*` → `main`, everything else → `develop`
3. Fallback: check if `develop` exists (`git show-ref --verify --quiet refs/heads/develop`), otherwise use `main`

Then collect diff data:

```bash
git diff $BASE...HEAD --stat
git diff $BASE...HEAD
git log --oneline --decorate $BASE..HEAD
git diff $BASE...HEAD --name-only
```

Store all outputs — they will be passed to subagents in the next step.

### Step 2: Spawn Parallel Reviewers

Read the two agent definition files from the **plugin root** `agents/` directory (not the skill directory):
- `agents/compliance-reviewer.md`
- `agents/code-quality-reviewer.md`

Each file has YAML frontmatter followed by a markdown body. Extract **only the markdown body** (everything after the closing `---` of the frontmatter) — that is the subagent's prompt. The YAML frontmatter is metadata for other platforms and must NOT be included in the prompt.

**Diff size check before spawning:**
```bash
git diff $BASE...HEAD | wc -l
```
- Under 1 000 lines → pass full diff to both agents
- 1 000–3 000 lines → pass `--stat` + per-file diffs; note the split in Steps taken
- Over 3 000 lines → pass `--stat` + `--name-only` only; instruct agents to use `read_file` on specific changed files rather than relying on an embedded diff

Spawn 2 subagents in a **single message** using the Agent tool with `subagent_type="general-purpose"`. Both run in parallel with clean contexts. Do NOT do any review analysis yourself — the subagents handle all assessment work.

#### Agent A: Compliance Reviewer

Use the Agent tool with `subagent_type="general-purpose"`.

Construct the prompt by combining:

1. **Agent instructions** — the markdown body extracted from `agents/compliance-reviewer.md` (everything after the YAML frontmatter `---` block)
2. **Context data** — append these sections after the agent instructions:

```
## Change Directory
{change_directory_path}

## Contracts
### proposal.md
{proposal.md content}

### specs/<capability>/spec.md (if exists)
{spec.md content or "No specs/ directory found."}

### design.md
{design.md content}

### tasks.md
{tasks.md content}

## Git Diff Stats
{git diff --stat output}

## Full Git Diff
{git diff output}

## Git Log
{git log --oneline output}
```

#### Agent B: Code Quality Reviewer

Use the Agent tool with `subagent_type="general-purpose"`.

Construct the prompt by combining:

1. **Agent instructions** — the markdown body extracted from `agents/code-quality-reviewer.md` (everything after the YAML frontmatter `---` block)
2. **Context data** — append these sections after the agent instructions:

```
## Change Directory
{change_directory_path}

## Changed Files
{git diff --name-only output}

## Full Git Diff
{git diff output}

## Tasks Reference
{tasks.md content}
```

**Fallback:** If the Agent tool fails or subagents cannot run, execute both review passes sequentially in the current context. Read the persona files and follow their instructions directly. Note in the review output that parallel mode was unavailable.

### Step 3: Merge and Write Review

Combine both subagent outputs into the final `review.md`:

1. **Steps taken:** Write a merged list of orchestrator steps ("Read contracts", "Ran git diff", "Spawned reviewers") plus key steps from both agents' own "Steps taken" sections. Keep max 5 words per step. Do not duplicate.
2. **Summary:** Write one sentence based on combined findings. Use strong wording about ticket linkage or business value only when evidence supports it. Otherwise keep wording scoped or uncertain.
3. **Compliance Assessment:** Copy Agent A's output as-is — sections, verdicts, tables, findings. Do not paraphrase or re-evaluate.
4. **Technical Assessment:** Copy Agent B's output as-is — sections, verdicts, tables, findings. Do not paraphrase or re-evaluate.
5. **Final verdict:** Determined by the two independent verdicts following the Verdict Rules below. The orchestrator does not add its own verdict.

Write the review to `$(sdlaic path change --change <name>)/review.md` with this structure:

```markdown
### Steps taken
- [1 line per major action, max 5 words each]

**Summary**: [One sentence: what was implemented and review outcome. Use strong wording about ticket linkage or business value only when evidence supports it. Otherwise keep wording scoped or uncertain.]

## Compliance Assessment
**Verdict**: [APPROVE / REQUEST CHANGES / REJECT]

### Strengths
- [What was done well, specifically]

### In Scope Issues
- [CRITICAL/HIGH/MEDIUM/LOW] <issue description> — path:line

### Out of Scope Issues
- [Adjacent debt or follow-on work worth noting, or "- None."]

### Proposal Compliance

| Proposal Item | Status | Notes |
|---------------|--------|-------|
| [What changes item] | DONE/PARTIAL/MISSING/EXTRA | [Details] |

### Requirements Compliance (if specs/ exists)

| Requirement | Scenario | Covered by Test |
|-------------|----------|-----------------|
| [from spec.md] | [WHEN/THEN] | Yes/No |

### Design Compliance

| Design Decision | Status | Notes |
|----------------|--------|-------|
| [Decision from design.md] | FOLLOWED/DEVIATED | [Details if deviated] |

## Technical Assessment
**Verdict**: [APPROVE / REQUEST CHANGES / REJECT]

### Strengths
- [What was done well technically]

### In Scope Issues
- [CRITICAL/HIGH/MEDIUM/LOW] <issue description> — path:line

### Out of Scope Issues
- [Technical debt or follow-on work, or "- None."]

### Scope Check

| File | Expected (in tasks) | Actually Changed |
|------|---------------------|-----------------|
| ... | Yes/No | Yes/No |

### Scope Violations
[Any changes outside the planned scope, or "- None."]

### Code Quality Checks

| # | Check | Status | Evidence |
|---|-------|--------|----------|
| 1 | DRY | PASS/FAIL | file:line or explanation |
| 2 | Module boundaries | PASS/FAIL | ... |
| 3 | Similar patterns | PASS/FAIL | ... |
| 4 | Duplication | PASS/FAIL | ... |
| 5 | Architecture conformance | PASS/FAIL | ... |
| 6 | Pattern/style matching | PASS/FAIL | ... |
| 7 | Documentation | PASS/FAIL | ... |
| 8 | Security | PASS/FAIL | ... |
| 9 | Performance | PASS/FAIL | ... |
| 10 | Contract maintenance | PASS/FAIL | ... |
| 11 | Test coverage | PASS/FAIL | ... |

### Reusability
- [Observations about reusability patterns, or "- None notable."]
```

### Step 4: Save and Advance

Write the review to `$(sdlaic path change --change <name>)/review.md`.

Note: `review` is a plugin-specific artifact, not a standard SDLAIC artifact. Write the file directly — no CLI command needed.

## Verdict Rules

- **APPROVE** — No CRITICAL or HIGH issues. All proposal items DONE. No scope violations.
- **REQUEST CHANGES** — HIGH or multiple MEDIUM issues. Fixable without architectural rethink.
- **REJECT** — CRITICAL issues or fundamental architectural drift. Requires design re-evaluation.

The two verdicts are **independent**. Compliance can APPROVE while Technical REQUESTS CHANGES (or vice versa).

**Resolution:** If either verdict is not APPROVE, the review is not complete. List required fixes. After fixes are applied, re-run `review`.

## Output Artifacts

- `$(sdlaic path change --change <name>)/review.md` — structured dual-assessment review with independent verdicts

## Verification

- [ ] Every component in proposal.md, specs/ (if exists), and design.md was checked against actual code
- [ ] All changed files were compared against the task list
- [ ] Requirements from specs/ (if exists) were checked against code and tests
- [ ] Design decisions from design.md were verified in implementation
- [ ] Code research search query was executed (or unavailability noted)
- [ ] Code research deep-dive query was executed (or unavailability noted)
- [ ] All 11 code quality checks evaluated with evidence
- [ ] Every finding has a file:line citation
- [ ] Steps taken section present with max 5 words per step
- [ ] Both verdicts (Compliance + Technical) are present and independent
- [ ] No CRITICAL or HIGH severity findings remain unresolved in either assessment
- [ ] Agent definition files were read from agents/ directory
- [ ] Both subagents spawned in parallel (2 Agent calls in single message)
- [ ] Each subagent had clean context (no prior chat history)

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Reviewing style instead of compliance | The proposal is the standard for compliance. CURe checks are the standard for quality. Don't mix them. |
| Giving APPROVE with unresolved HIGH findings | HIGH = must fix. No exceptions in either assessment. |
| Not checking for scope creep | Compare changed files against tasks.md. Extra changes = scope violation. |
| Writing vague findings ("could be improved") | Specific: what, where (file:line), why it matters, what to do. |
| Skipping code research without noting it | If the tool is unavailable, explicitly state it in the review. Don't silently skip. |
| Running research queries in parallel | Run one research call at a time. Wait for the result before issuing another. |
| Speculating about unread code | Never flag an issue about code you haven't read. Investigate first. |
| Giving both assessments the same verdict by default | The verdicts are independent. Evaluate each on its own merits. |
| Running subagents sequentially | Use 2 Agent tool calls in a single message to run both in parallel. |
| Orchestrator doing review work | The orchestrator only gathers context, spawns agents, and merges output. It does NOT analyze code. |
| Not reading agent files | Always read agent definition files from the plugin root agents/ directory. The instructions live there, not in SKILL.md. |
| Including YAML frontmatter in subagent prompt | Extract only the markdown body after the closing ---. The YAML is platform metadata, not prompt content. |
| Passing oversized diff to subagents | Check diff line count first. Over 3 000 lines → pass --stat + --name-only only, let agents read files directly. |
| No fallback when Agent tool fails | If parallel execution fails, fall back to sequential: read agent files and follow instructions directly. |

## Handoff

If **both APPROVE**: The change is COMPLETE. Suggest updating Jira and archiving.
If **any non-APPROVE**: List required fixes. After fixes are applied, re-run `review`.

---
name: brainstorm
description: Use after grillme has challenged assumptions and rationale.md exists. The design phase — searches for existing patterns, conducts Socratic dialogue, and produces the formal proposal, capability delta spec, and design document (in that order).
---

# Design Phase (Brainstorm)

## Core Principle

Design before code. The proposal is the contract (why/what), the spec is the verifiable requirement (formal WHAT), and the design is the architecture that satisfies them (HOW). Write them in that order — locking down "what" before "how" prevents the design from drifting away from the requirement. Review will check implementation against all three.

## Pre-conditions

- SDLAIC change is in CHALLENGED state
- `rationale.md` exists with validated decisions
- `context.md` exists with original ticket context (optional — may not exist if context was gathered from other sources)

## Process

### Step 0: Review Open Items from Rationale

Read `.sdlaic/changes/<name>/rationale.md` section "Open Items".

If any open items exist:
1. Present each one to the user
2. Resolve or explicitly defer with a documented reason
3. Record each resolution in `design.md` section "Resolved Open Items" (added in Step 6)

Do NOT proceed to pattern search until all blocking open items are resolved or explicitly deferred with a reason.

### Step 1: Search for Existing Patterns

Before designing anything new, find what already exists. Run code research queries scoped to `projects/<project>` for the relevant patterns from `rationale.md`. Tool selection follows your workspace convention (see `references/code-research.md`).

Look for:
- **Existing implementations** of similar features
- **Shared utilities** that could be reused
- **Architecture patterns** already established in the codebase
- **Database migrations** or schema patterns that are precedent

If a query returns no results, try broader queries (e.g., the module name, the data entity, the HTTP verb). If still empty, document that no existing patterns were found — this is a valid outcome, not a skip.

Write the **Pattern Summary** directly into `design.md` (Step 6) under section "Pattern Research":
- What patterns exist and where (file paths)
- What can be reused vs what needs to be built new
- Any conventions the design should follow
- If nothing found: "No existing patterns found for [query terms]. Building new."

### Step 2: Socratic Design Dialogue

For each section, formulate a draft design proposal first, then present it to the user for feedback. Do NOT present a blank slate and ask "what do you want?" — present a concrete proposal and ask "does this work?".

Ask ONE section at a time. Get explicit approval before moving to the next.

**Section 1: Architecture**
- How does this fit into the existing system?
- What components/modules are involved?
- What's the interaction flow?

**Section 2: Data Flow**
- What data is read, transformed, and written?
- Where does it come from and where does it go?
- What are the data models/schemas?

**Section 3: Interface Contract**
- What APIs, routes, or interfaces are exposed or consumed?
- What are the inputs and outputs?
- What error states exist?

**Section 4: Edge Cases & Error Handling**
- What happens when dependencies are unavailable?
- How are invalid inputs handled?
- What's the graceful degradation strategy?

**Section 5: Risks & Unknowns**
- What parts of the design are uncertain?
- What might need to change during implementation?
- What are the performance implications?

For each section:
1. Present the draft design proposal (specific, not open-ended)
2. Ask for feedback
3. Incorporate feedback
4. Get explicit approval before moving to the next section

### Step 3: Write the Proposal

The proposal describes **why and what** — the business rationale and scope of change.

```bash
sdlaic instructions proposal --change <name>
```

Review the instructions output for template guidance, then write to `.sdlaic/changes/<name>/proposal.md`:

```markdown
# Proposal: <change-name>

**Change ID**: `<change-name>`
**Jira**: <KEY> (<summary>)

## Why
[One paragraph: why this change is needed — the problem it solves]

## What Changes
- **[Bug fix / Enhancement / New feature]: [description]**
  - [Specific change 1]
  - [Specific change 2]

## Success Criteria
- [ ] [Testable criterion 1 — e.g. "API returns 200 with expected payload when X"]
- [ ] [Testable criterion 2 — e.g. "User sees Y within Z seconds"]

## Impact
- **Affected specs**: [New/updated capability delta — reference spec file]
- **Affected code (indicative)**: [Key areas that will change]
- **Breaking changes**: [None intended / list them]

## Approval
Do not start implementation until this proposal is reviewed and approved.
```

### Step 4: Decide on Capability Delta Spec

Before designing the technical solution, decide explicitly whether this change needs a formal requirements spec. The decision happens here so the spec (if needed) is written **before** the design — locking down "what" before "how".

A delta spec **is required** when:
- The change introduces or modifies a **user-facing capability** (UI flow, API contract consumers depend on, observable behavior change)
- External parties (QA, product, stakeholders) need testable acceptance criteria to verify against

A delta spec **is NOT required** when:
- The change is purely internal (refactor, performance tuning, dependency bump, code reorganization)
- No user-observable behavior changes

Record the decision (Yes/No + reason). The same record will be copied into `design.md` section "Delta Spec Decision" in Step 6 for the audit trail.

- If Yes → proceed to Step 5, then Step 6.
- If No → skip Step 5, jump directly to Step 6.

### Step 5: Write Capability Delta Spec (if applicable)

If Step 4 decided a spec is needed, write the formal requirements spec **now, before the design document**. The design in Step 6 will be built to satisfy this spec.

Write to `.sdlaic/changes/<name>/specs/<capability-name>/spec.md`:

```markdown
# <capability-name> (delta)

## ADDED Requirements

### Requirement: [Verb statement of what the system MUST do]

[One paragraph describing the requirement in full]

#### Scenario: [Descriptive name]
- **WHEN** [precondition]
- **THEN** [expected outcome]
- **AND** [additional constraint]

## CHANGED Requirements

### Requirement: [Existing requirement being modified]
[What changed and why]
```

Skip this step only if Step 4 explicitly decided no spec is needed. Do not silently skip.

### Step 6: Write the Design Document

The design describes **how** — the technical architecture and decisions that satisfy the proposal (Step 3) and, if it exists, the delta spec (Step 5).

Write to `.sdlaic/changes/<name>/design.md`:

```markdown
# Design: <change-name>

## Context
[Technical context: what system area this touches, current state, how it works today]

## Goals
- [Goal 1]
- [Goal 2]

## Non-Goals
- [Explicitly out of scope items]

## Pattern Research
**Queries used**: [list of code research queries]
**Relevant findings**:
- [Pattern name] — [file path] — [reuse / reference / adapt]
**Decision**: [Reuse X from Y / Build new because Z / No existing patterns found for [terms]]

## Resolved Open Items
| Open Item (from rationale.md) | Resolution | Deferred? |
|-------------------------------|------------|-----------|
| [item] | [how resolved] | Yes / No |

## Delta Spec Decision
<!-- Record the outcome of the Step 4 gate. If a spec was created in Step 5, link it here. -->
- **User-facing change**: [Yes / No]
- **Decision** (from Step 4): [Created spec at `.sdlaic/changes/<name>/specs/<capability>/spec.md` / Skipped — reason: <reason>]

## Decisions
| Area | Decision | Rationale |
|------|----------|-----------|
| [e.g. PDF handling] | [e.g. Align with existing download route] | [Why this approach] |

## Risks / Trade-offs
<!-- Import all risks from rationale.md "Risks Identified" as starting point. Add new technical risks found during design. -->
- **[Risk carried from rationale]**: [Mitigation]
- **[New technical risk from design]**: [Mitigation]

## Open Questions
- [Anything unresolved — or "None blocking proposal"]
```

### Step 7: Get Approval

Present a summary of key decisions from each artifact — not the full document. Format as (in the order they were written):

```
**Proposal summary**: [2-3 sentences on why, what, and success criteria]
**Delta spec**: [Created at <path> — N requirements / Skipped — reason]
**Design summary**: [key architectural decision + top risk]

Please confirm you've reviewed each artifact before I commit.
```

Do NOT proceed to `plan` until the user explicitly approves all artifacts.

### Step 8: Commit Design Artifacts

Only after user explicitly approves all artifacts in Step 7.

Stage and commit all design artifacts in the target project's repository in the order they were written (proposal → specs → design):

```bash
git -C projects/<project> add .sdlaic/changes/<name>/proposal.md
# If capability delta spec was created in Step 5:
git -C projects/<project> add .sdlaic/changes/<name>/specs/
git -C projects/<project> add .sdlaic/changes/<name>/design.md
git -C projects/<project> commit -m "<prefix>: brainstorm design for <change-name>"
```

This checkpoint records the approved design contract before planning begins.

## Output Artifacts

Written in this order:

1. `.sdlaic/changes/<name>/proposal.md` — business contract (why, what, success criteria, impact)
2. `.sdlaic/changes/<name>/specs/<capability>/spec.md` — formal requirements with WHEN/THEN scenarios (optional, only if Step 4 decided Yes)
3. `.sdlaic/changes/<name>/design.md` — technical design (pattern research, resolved open items, delta spec decision record, decisions, risks)

## Verification

- [ ] `proposal.md` has "Success Criteria" section with ≥1 testable criterion
- [ ] Step 4 Delta Spec Decision was made explicitly (Yes or No with reason) — not silently skipped
- [ ] If decision was Yes: `specs/<capability>/spec.md` exists and was written **before** `design.md`
- [ ] `design.md` section "Delta Spec Decision" records the Step 4 outcome (and links the spec path if created)
- [ ] `design.md` section "Pattern Research" is filled — queries listed, findings noted (or "No existing patterns found" explicitly stated)
- [ ] `design.md` section "Resolved Open Items" has one row per open item from `rationale.md` (or is empty because rationale had none)
- [ ] `design.md` section "Risks / Trade-offs" imports risks from `rationale.md` plus any new design risks
- [ ] All 5 design sections were discussed — `design.md` "Decisions" table has ≥1 row per section area
- [ ] User explicitly approved artifacts via Step 7 summary
- [ ] Design artifacts committed to target project repo in order: proposal → specs (if any) → design

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Designing without checking existing patterns | Always run a code research query first. Reuse > reinvent. |
| Skipping research because "I know this codebase" | Search anyway. Record what you found. Reviewer needs proof. |
| Not resolving Open Items from rationale before designing | Step 0 is mandatory. Unresolved open items poison the design. |
| Presenting the entire design at once | One section at a time. Get approval per section. |
| Presenting blank sections ("What architecture do you want?") | Draft a concrete proposal first, then ask for feedback. |
| Mixing why/what with how in proposal | Proposal = business (why, what). Design = technical (how). Keep them separate. |
| Skipping the delta spec for user-facing features | If users interact with it, it needs a spec with WHEN/THEN scenarios. |
| Omitting Success Criteria from proposal | Success criteria must be testable — they are the acceptance gate for `review`. |
| Vague success criteria ("make it work") | Success criteria must be testable: "API returns 200 with expected payload" |
| Skipping the approval step | No approval = no proceed. The proposal is a contract. |
| Ignoring risks because "it should be fine" | Every design has risks. Import from rationale, then add design-specific ones. |

## Handoff

Route to `plan` to break the approved proposal, spec (if any), and design into implementable tasks.

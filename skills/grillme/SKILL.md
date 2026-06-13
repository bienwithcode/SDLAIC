---
name: grillme
description: Use when a new change has been initialized and context gathered. The Devil's Advocate phase — challenges assumptions, exposes blind spots, and forces specificity before any design work begins.
---

# Socratic Challenge (Grill Me)

## Core Principle

Bad assumptions are the #1 source of rework. This phase exists to break the idea before code breaks the system.

## Pre-conditions

- An SDLAIC change exists and is in INITIALIZED state
- `context.md` exists with ticket description or user-provided context
- Research Summary was produced during `new` — if missing, re-run `new` or ask the user to provide a written summary before proceeding

## Process

### Step 1: Read & Validate Context

#### Step 1.1: Read Context

Read the following in order:
1. `.sdlaic/changes/<name>/context.md` — Jira description, comments, resolved file description, and research summary
2. Research Summary from the previous `new` phase

Identify:
- **Explicit requirements** — what the ticket/resolved description literally asks for
- **Implicit assumptions** — things not stated but assumed true
- **Vague language** — words like "improve", "faster", "better", "should"
- **Missing information** — what's not mentioned but probably matters

**Gap Manifest check (run after reading):**
If `context.md` contains a `## Gap Manifest` table with populated rows, extract it
into a **Gap-Driven Surface List**. Each row becomes a mandatory IN surface in
Step 2 — it cannot be marked N/A without explicit written justification. The
"Elicit" column defines the minimum scope of questions grillme must cover for
that surface; treat it as a floor, not a ceiling.

**Actors & Use Cases check (run after reading):**
If `context.md` contains a `## Actors & Use Cases` section with populated tables,
build a **Scenario Anchor List** = every `(Actor × Use Case)` pair (happy /
alternate / edge each counts). Each question in Step 3 must name one anchor
from this list — or explicitly justify itself as actor-agnostic. If the section
is `N/A: <reason>` (trivial change) or absent, anchoring is not required and
G1 falls back to "names a concrete situation".

#### Step 1.2: Challenge Context Sanity (Context Logic & Alignment Gate)

Before identifying risk surfaces, actively challenge and validate the integrity of the context:
1. **Sanity Check**: Ensure that the description in `context.md` is actual descriptive text of the change. It must NOT be a raw local file path, a URL, or an empty placeholder.
2. **Alignment Check**: Confirm that the active change name (e.g. `add-antigravity-providers`) logically aligns with the description of the work in the context.
3. **Logic & Codebase Contradictions**: Check if the requirements in the description contradict known patterns in the codebase or contain logical impossibilities (e.g., calling an API that doesn't exist, modifying a dependency that is read-only, making changes to a file outside the workspace, or introducing circular dependencies).
4. **Halt Condition**: If any check fails (e.g., description is a file path, or name is completely mismatched, or there is a glaring logical contradiction):
   - **HALT** the phase. Do NOT proceed to Step 2.
   - Present the mismatch/contradiction/flaw clearly to the user.
   - Instruct the user to re-run the `new` phase with corrected inputs or correct the description in `context.md` before continuing.

### Step 2: Identify Risk Surfaces

Scan the context and list **every risk surface that actually applies to this ticket**. Do not cap the count — the number of questions equals the number of applicable surfaces. Some surfaces may legitimately be N/A (mark them so, with a one-line reason).

**Gap-driven surfaces (from Step 1.1 Gap Manifest check) are always IN.** Merge
them into the inventory first, using the Priority from the manifest. The default
catalog surfaces are then scanned and added after. If a gap-driven surface
overlaps with a catalog surface, keep the gap-driven row's Priority and expand
its Elicit scope with anything the catalog scan surfaces.

Default surface catalog (scan each, keep or drop):

1. **Failure modes** — What breaks? What's the worst case?
2. **Edge cases** — Empty states, large datasets, concurrent access, offline.
3. **Scope boundaries** — Where does this end? What's explicitly NOT included?
4. **Security / data** — Who can access? What data exposed? Permissions needed?
5. **ROI / impact** — Measurable outcome? How do we know it worked?
6. **Rollback** — How do we undo this safely? What's the blast radius?
7. **Observability** — How will we notice it broke in production?
8. **Migration / backfill** — Data shape changes? Existing rows / cache invalidation?
9. **Concurrency** — Races, locking, idempotency under retry.
10. **UX surface** — User-visible behavior, error states, accessibility.

Output a **Surface Inventory** before moving to Step 3 with three columns: Status (IN / N/A), Reason (one line), and Priority (1 = highest risk, ascending). Rank IN surfaces by *risk-to-this-ticket*, not catalog order — Security on an auth change ranks above ROI; Migration on a schema change ranks above UX. Step 3 iterates surfaces in Priority order so the most consequential questions land first, when user attention is freshest.

### Step 3: Ask One Question at a Time (Research-Grounded Loop)

For each question, execute substeps **3a → 3d**. Do NOT combine questions. Do NOT proceed without resolution.

A surface may take **1–3 questions** depending on how many distinct angles it has (e.g. Security can need separate questions for access control, rate limiting, and data exposure). Step 3d decides whether to stay on the current surface for a follow-up or move to the next.

**Iterate surfaces in Priority order from the Surface Inventory** — highest-risk first. Don't ask follow-ups on a low-priority surface while higher-priority surfaces still have main questions pending.

#### Question Quality Rubric — apply before every question

A question is **GOOD** only if it satisfies at least 4 of these 5:

- **G1. Specific scenario.** When `context.md > ## Actors & Use Cases` is populated and not `N/A`, names a concrete `(Actor, Use Case)` from the Scenario Anchor List (or explicitly justifies the question as actor-agnostic — cross-cutting infra / framework / build pipeline). When the section is `N/A` or absent, falls back to naming a concrete situation. Generic "any concerns?" / "anything else?" always fails.
- **G2. Evidence-anchored.** References the research finding from 3a (file:line precedent, web:<URL> for external references, or explicit "no precedent"). **Exception for gap-driven surfaces:** when the gap is intentionally absent from context.md (ticket never stated it), G2 is satisfied by stating `"gap-driven: <Elicit text from manifest> — first-principles question"` instead of a file:line or web citation. This is not a loophole — a gap-driven question must still pass G1, G3, G4, G5.
- **G3. Forces a short decision.** Answer is binary, a short list, or a named choice — not an essay.
- **G4. Materially shapes design.** Answer adds a new row to `rationale.md > Decisions` or `Open Items`. If the answer changes nothing in the design, the question is filler.
- **G5. Targets a real risk.** Surface a real failure / cost / blast-radius — not "for completeness".

A question is **BAD** if any of these hold (don't ask it; reframe or drop):

- **B1.** Open-ended "anything else on X?"
- **B2.** Confirms what `rationale.md` or research already states.
- **B3.** Has no consequence regardless of the answer.
- **B4.** Could be looked up directly without grilling the user.
- **B5.** Restates a prior question with cosmetic word changes.

If you cannot draft a question that passes the rubric for the current surface, **move to the next surface** rather than asking filler.

#### 3a. Pre-question research (code or web)

Before asking, run **one** targeted research query for the question topic. For local codebase analysis, query the codebase scoped to the project path. For third-party libraries, protocols, or external integrations, run a websearch. Tool selection follows your workspace convention (see `references/code-research.md`).

Use the result to ground the Recommended option below. If research returns nothing or the tool is unavailable, mark the question as "no precedent" and proceed — the Recommended slot becomes a reasoning-based default, not an evidence-based one.

#### 3b. Present the question with 3 options + chat

Format:

```
**Question [N]: [Topic]**
**Category**: [Failure mode / Edge case / Scope / Security / ROI]
**Anchor**: (Actor: <name>, Use Case: <happy|alternate|edge — short name>)  — or `actor-agnostic: <one-line reason>` for cross-cutting infra

[The actual question — specific, not vague, named to the anchor when applicable]

Context: [Why this matters — what breaks if we get this wrong, framed in the anchor's scenario]

Research finding: [1-2 lines — existing pattern at file:line, external doc at web:<URL>, or "no precedent found for <terms>"]

How do you want to answer?

A. [Option grounded in research finding] **(Recommended)**
B. [Alternative — different trade-off]
C. [Alternative — narrower or wider scope]
D. Chat — answer in your own words
```

Rules:
- All 4 options must be present every time.
- Recommended (A) must reference the research finding (file:line, web:<URL>, or pattern name). If no finding, mark "(Recommended — reasoning-based, no precedent)".
- B and C must be *meaningfully different* from A, not cosmetic variants.
- **Anchor line is mandatory** when `context.md > ## Actors & Use Cases` is populated and not `N/A`. Either name an `(Actor, Use Case)` from the Scenario Anchor List, or mark `actor-agnostic: <reason>` (allowed for cross-cutting infra / framework / build pipeline questions — not as a loophole to skip anchoring). An abstract question on a ticket with named actors fails G1 and must be reframed before asking.

#### 3c. Process the answer

- **User picks A/B/C** → record the option text verbatim in the Challenge Transcript. Move to 3d.
- **User picks D (Chat)** → apply the **paraphrase gate**:
  1. Paraphrase the answer back: _"Got it — you mean: [concise 1-sentence paraphrase]. Correct?"_
  2. User confirms → record paraphrase in transcript, move to 3d.
  3. User corrects → re-paraphrase. **Max 1 retry.**
  4. Still ambiguous after retry → record `UNRESOLVED` in transcript + add to Open Items. Move to 3d.

#### 3d. Continuation check

After 3c, decide whether to **stay on the current surface** (ask a follow-up) or **move to the next surface**.

**Diminishing-returns check (always run first):**
Look at the last 2 questions in the transcript. Did they each produce a new entry in `rationale.md > Decisions` or `Open Items`? If NO — both recent questions yielded nothing materially new — the loop is filler. **Force MOVE** to the next surface, regardless of caps. Do not STAY.

**STAY (follow-up on current surface)** if ALL of these hold:
- You can name a *specific, concrete* uncovered angle on this surface (e.g. on Security: covered access control, not yet covered rate limit).
- That angle is supported by either:
  - The research finding from 3a revealing a sub-angle, OR
  - The user's answer in 3c surfacing a concrete sub-angle worth grilling.
- You can draft a follow-up question that **passes the Question Quality Rubric** (≥4 of G1–G5, none of B1–B5).
- This surface has had **fewer than 3 questions** so far (per-surface cap).
- Total questions so far is **below 30** (hard cap, see below).

**MOVE to next surface** if ANY of these hold:
- You cannot articulate a specific uncovered angle for this surface.
- You can't draft a question that passes the Quality Rubric.
- The surface already has 3 questions on it.
- The diminishing-returns check fired.
- User said "enough on this" / equivalent.

**Loop control & caps (2-tier):**

- **While surfaces remain in the inventory:** iterate in Priority order. STAY or MOVE per the rules above. Loop back to 3a.

- **Soft cap at Q15 — Quality Meta-Check (does NOT exit).** Before asking the 15th question (and again before the 25th), pause and present:
  > _"Meta-check: {N} questions reached. Most tickets resolve well before this. Current state: {surfaces fully covered / surfaces remaining / avg Qs per surface}. Continue grilling, or wrap up?_
  > _A. Continue — concrete uncovered angles remain_
  > _B. Wrap up — push remaining surfaces to Open Items_
  > _C. Show transcript so far before I decide"_
  - **A** → loop continues; the next meta-check is at Q25.
  - **B** → exit loop, push remaining IN surfaces to Open Items, go to Step 4.
  - **C** → output a transcript summary table, then re-ask the meta-check.

- **Inventory exhausted (before any cap):** ask the user explicitly:
  > _"I've covered every risk surface I identified ({list}). Any other angle to grill, or are we done?"_
  - User names a new angle → add it to the inventory with a Priority, loop back to 3a.
  - User says "done" / equivalent → exit loop, go to Step 4.

- **Hard cap at Q30 — force exit.** If Q30 is reached, the agent **stops the loop unconditionally**, records `HARD_CAP_FIRED` in the rationale metadata, dumps remaining IN surfaces into Open Items, and proceeds to Step 4. The hard cap exists purely to stop runaway loops (prompt bugs, agent confusion) — natural termination via Quality Rubric, diminishing-returns, or user wrap-up should happen well before this.

### Step 4: Synthesize Answers

Once all questions are answered, synthesize into `rationale.md`:

```markdown
# Rationale: <change-name>

## Session Metadata
- **Date**: YYYY-MM-DD
- **Surfaces in inventory**: N (IN: M, N/A: K)
- **Questions asked**: N | **Answered**: N | **Unresolved (Open Items)**: N
- **Soft cap meta-check at Q15**: Not reached / Continue / Wrap up / (not applicable)
- **Soft cap meta-check at Q25**: Not reached / Continue / Wrap up / (not applicable)
- **HARD_CAP_FIRED (Q30)**: Yes / No
- **Reviewed by**: [user handle / email]

## Gap Manifest Coverage
<!-- Only present if context.md had a populated ## Gap Manifest. Omit this section entirely if the manifest was empty. -->
| Gap | Mapped Surface | Status | Notes |
|-----|---------------|--------|-------|
| <gap from manifest> | <surface> | IN (covered) / N/A (reason) | <what was elicited, or why skipped> |

## Problem Statement
[One paragraph: what problem are we solving and for whom]

## Key Constraints
- [Constraint 1]
- [Constraint 2]
- ...

## Surface Inventory
| Surface | Status | Reason | Priority |
|---------|--------|--------|----------|
| Failure modes | IN / N/A | [why included or skipped] | 1–N or — |
| Edge cases | IN / N/A | ... | ... |
| Scope boundaries | IN / N/A | ... | ... |
| Security / data | IN / N/A | ... | ... |
| ROI / impact | IN / N/A | ... | ... |
| Rollback | IN / N/A | ... | ... |
| Observability | IN / N/A | ... | ... |
| Migration / backfill | IN / N/A | ... | ... |
| Concurrency | IN / N/A | ... | ... |
| UX surface | IN / N/A | ... | ... |
| [User-added angle] | IN | [reason] | ... |

Priority: 1 = highest risk for this ticket (asked first). N/A surfaces leave Priority blank.

## Challenge Transcript
| # | Surface | Anchor | Question | Research finding | Answer | Status |
|---|---------|--------|----------|--------------------|--------|--------|
| 1 | [surface] | (Actor: ..., UC: ...) or `actor-agnostic: <reason>` or `—` (if context.md has no Actors section) | [exact question asked] | [file:line or "no precedent"] | [verbatim or paraphrase] | RESOLVED / UNRESOLVED |
| 2 | ... | ... | ... | ... | ... | ... |

## Decisions Made
| Question | Answer | Implication |
|----------|--------|-------------|
| ... | ... | ... |

## Scope Boundaries
### In Scope
- ...

### Out of Scope
- ...

## Open Items
- [Anything still unresolved — these MUST be addressed before or during brainstorm]

## Risks Identified
- **[Risk 1]**: [Mitigation]
- **[Risk 2]**: [Mitigation]
```

### Step 5: Save Rationale

Write the rationale content to `.sdlaic/changes/<name>/rationale.md`.

Note: `rationale` is a plugin-specific artifact, not a standard SDLAIC artifact. Write the file directly — no CLI command needed.

The plugin phase (CHALLENGED) is inferred from the presence of `rationale.md` — there is no CLI state to advance. See `references/sdlaic-standards.md`.

### Step 6: Commit Rationale

Stage and commit `rationale.md` in the target project's repository:

```bash
git -C projects/<project> add .sdlaic/changes/<name>/rationale.md
git -C projects/<project> commit -m "<prefix>: grillme rationale for <change-name>"
```

This checkpoint records the challenged and validated assumptions before design begins.

## Output Artifacts

- `.sdlaic/changes/<name>/rationale.md` — synthesized challenge results including full transcript

## Verification

- [ ] Context Sanity and Logic checks passed (Step 1.2 completed, verifying that the description is valid, aligns with the change name, and does not contain codebase/logic contradictions)
- [ ] If `context.md > ## Gap Manifest` had populated rows, every row appears as IN in the Surface Inventory; N/A requires explicit written justification
- [ ] Gap-driven surfaces used "gap-driven: <elicit>" as G2 anchor when no file:line or web evidence existed
- [ ] If Gap Manifest was populated: `rationale.md` contains `## Gap Manifest Coverage` with one row per manifest entry
- [ ] `rationale.md` section "Surface Inventory" exists with Status, Reason, and Priority columns; every catalog surface is marked IN (with reason + priority) or N/A (with reason)
- [ ] IN surfaces were iterated in Priority order (highest risk first)
- [ ] `rationale.md` section "Challenge Transcript" has at least one row per IN surface (a surface may have up to 3 rows for follow-ups; plus any user-added angles)
- [ ] No surface in the transcript exceeded 3 questions
- [ ] Every question passes the Question Quality Rubric (≥4 of G1–G5, none of B1–B5)
- [ ] When `context.md > ## Actors & Use Cases` is populated and not `N/A`: every transcript row has an `(Actor, Use Case)` anchor OR an explicit `actor-agnostic: <reason>` tag; no bare/empty Anchor cells
- [ ] Every question recorded its research finding (or "no precedent" / "research tool unavailable") in the transcript
- [ ] Every question after the second produced a new entry in Decisions or Open Items (diminishing-returns check honored)
- [ ] No chat answer was recorded without a paraphrase confirmation or an UNRESOLVED tag
- [ ] After the inventory was exhausted, the continuation prompt was asked at least once
- [ ] If the soft cap meta-check fired (Q15 / Q25), the user's choice (A/B/C) is recorded in the rationale metadata
- [ ] If the hard cap (Q30) fired, `HARD_CAP_FIRED` is recorded and remaining surfaces moved to Open Items
- [ ] Every UNRESOLVED row in the transcript has a corresponding entry in "Open Items"
- [ ] `rationale.md` section "Risks Identified" has ≥1 entry (no risk-free change exists)
- [ ] `rationale.md` section "Open Items" — each entry has a resolution plan or is marked for brainstorm (not bare "TBD")
- [ ] `rationale.md` exists in the change directory (plugin phase CHALLENGED is inferred from this)
- [ ] `rationale.md` committed to target project repo

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Blindly trusting the input context description without verification | Always run Step 1.2 sanity checks. If the description is a file path, empty, or mismatched with the change name, halt the phase. |
| Asking soft questions ("Is there anything else?") | Ask hard, specific questions with failure context |
| Asking all questions at once | One at a time. Always. |
| Accepting vague answers ("it should be fine") | Push for specificity: "What does 'fine' mean? How do we measure it?" |
| Accepting "I don't know" as a final answer | Record as UNRESOLVED Open Item — do not skip or leave blank |
| Skipping the research query before asking | Each question needs 1 query. The Recommended option must reference the finding. |
| Recommended option not grounded in evidence | Recommended must cite file:line or web:<URL> from research. No finding → mark "(reasoning-based, no precedent)". |
| Letting Option 4 chat run forever | Paraphrase gate: max 1 retry. After that, record UNRESOLVED and move on. |
| Stopping before the Surface Inventory is drained | Every IN surface gets at least one question. The count is determined by the inventory, not a fixed number. |
| Marking surfaces N/A without reason | Every N/A must have a one-line justification in the Surface Inventory — silent skipping is not allowed. |
| Treating the 15-question cap as a target | The cap is a runaway-safety, not a goal. Most tickets should finish well below it. |
| Asking a vague follow-up ("anything else on security?") | A follow-up must name a specific uncovered angle (e.g. "rate limit") with evidence from research or the prior answer. Generic follow-ups are filler (B1). |
| Maxing every surface to 3 questions because allowed | The 3-per-surface limit is a ceiling, not a target. Most surfaces resolve in 1 question. Only ask follow-ups when you can name the new angle. |
| Asking questions that don't pass the Quality Rubric | If you can't satisfy ≥4 of G1–G5 without hitting B1–B5, the question is filler — drop it and MOVE to the next surface. |
| Iterating surfaces in catalog order regardless of risk | Surfaces must be iterated in Priority order from the inventory. Low-risk first wastes user attention. |
| Asking 2+ questions in a row that yield no new Decision / Open Item | Diminishing-returns trigger. Force MOVE even if other STAY conditions hold. |
| Treating the soft cap (Q15) as a hard stop | Q15 is a meta-check, not an exit. User chooses Continue / Wrap up / Show. Only the hard cap at Q30 forces exit. |
| Aiming to reach Q30 | The hard cap is runaway protection, not a target. Natural termination should happen via Quality Rubric, diminishing-returns, or user wrap-up. |
| Skipping this phase because "requirements are clear" | Requirements are never as clear as they seem. Challenge anyway. |
| Ignoring the Gap Manifest from context.md | Every manifest row is a mandatory IN surface. Skipping it means the ticket's known gaps go unchallenged — exactly the failure new was designed to prevent. |
| Treating gap-driven G2 as a free pass | "gap-driven" satisfies G2 only. The question still needs G1, G3, G4, G5 — vague gap-driven questions are still filler. |
| Asking actor-agnostic questions when context.md lists named actors | Anchor to `(Actor, Use Case)` from the Scenario Anchor List. Abstract questions discard the extraction `new` already paid for and silently fail G1. `actor-agnostic:` is reserved for genuine cross-cutting concerns (infra, build, framework), not a loophole. |
| Using `actor-agnostic:` as a fallback when the anchor is just hard to pick | If multiple anchors apply, pick the highest-risk one and note the rest in the question Context. Defaulting to `actor-agnostic:` because choosing felt arbitrary is filler. |
| Producing rationale without all answers | Every question must have a concrete answer or an explicit UNRESOLVED status. |
| Adding new requirements during grillme | Grillme challenges — it does NOT design. Park new ideas for brainstorm. |
| Handing off without writing `rationale.md` | The plugin phase CHALLENGED is determined by the presence of `rationale.md`. No file, no handoff. |

## Handoff

Route to `brainstorm` to design the solution based on the challenged and validated rationale.

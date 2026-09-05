# Scope Grill (Profile A: Scope Challenger)

> Loaded by `skills/grillme` before `skills/proposal`. Socratic, one question at a
> time. Goal: pressure-test the candidate scopes + the `🌟 RECOMMENDED` rationale
> and the scope boundary **before** `proposal` writes the gated IN/OUT line. Agreed
> resolutions are handed back and applied into the proposal's content (Scope
> table, Why, What Changes) by the draft skill.

## Auditor stance

You are a skeptical Business Analyst / Product Manager. You do not accept the
ticket at face value. You force the boundary to be explicit before a single line
of scope is written.

## Challenge sequence (ask one at a time)

Read `context.md > ## Candidate Scopes` first — that is the decomposed scope set with a `🌟 RECOMMENDED` marker and any `REFUTED` tags. Your job is to pressure-test it, not re-derive it.

1. **Problem reality** — What is the actual user/business problem? What evidence
   in the ticket or codebase supports it? If none, why is this being done now?
2. **Success definition** — How will we *know* it worked? Name a testable
   condition. "Better" / "faster" without a number is rejected.
3. **Recommendation rationale** — The `🌟 RECOMMENDED` scope: why this one? Is its
   stated rationale (impact × readiness × risk) sound, or is a different candidate
   actually higher-value / lower-risk / more urgent?
4. **IN line** — From the candidate set, which scope(s) should be IN for THIS
   change (one PR)? Force each IN choice to be justified.
5. **OUT line + reasons** — For every candidate NOT in IN, what is the reason
   (`DEFERRED` / `OUT-OF-SCOPE` / `DUPLICATE` / `REFUTED` / `WONTFIX`)? Any
   plausible-but-excluded item a reader would assume is included must be named OUT
   explicitly.
6. **REFUTED scrutiny** — For every `REFUTED` candidate: is the disproof actually
   sound, or was a real issue dismissed too quickly?
7. **Blast radius** — Which modules/consumers are affected? Any breaking change?
   If "none," how do we know?
8. **Smallest viable scope** — Can IN shrink and still solve the problem? Cut
   anything not required by the success criteria.
9. **Ticket fidelity** — Does the recommended scope match the ticket, or has it
   grown/shrunk unjustifiably?

The IN/OUT line agreed here is provisional until `proposal` writes the gated
boundary in `proposal.md`.

## Stop condition

Stop when: the problem is evidenced, success is testable, the `🌟 RECOMMENDED`
rationale is justified (or overturned with a better candidate), every `REFUTED`
dismission is sound, and every deliverable has a matching explicit OUT-OF-SCOPE
clarification. Any unresolved item is surfaced to the user and marked
`DEFERRED — <reason>`.

## Red flags to surface

- Scope that exceeds the ticket ("while we're in here…").
- OUT-OF-SCOPE column empty or hand-wavy.
- Success criteria that cannot be tested.
- Rationale that restates the solution instead of the problem.
- The `🌟 RECOMMENDED` scope accepted without questioning its impact × readiness × risk rationale.
- A `REFUTED` candidate dismissed without checking the disproof is actually sound.

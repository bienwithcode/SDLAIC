# Split plan: <repo>#<PR> — <title>

Source: head <sha> · base <branch>@<sha> · snapshot <local-branch>@<sha> (worktree: <path>)
Contract: <spec path> + <PR description sections> + <ticket(s) AC> + <review comments>
Diff: prod <files> +<a>/−<d> · tests … · docs … · generated …

## Summary
<n> independent PR(s) + <n>-layer stack = <total> PRs.
Release: <merge stack as one group | rollout control>.

## Dependency picture
```
<base>
 ├─ I1 <name>
 ├─ I2 <name> (hold/revert alone)
 └─ S1 <layer> → S2 <layer> → S3 <layer>
```

## Independent PRs
### I1 — <title>
- Branch / base:
- Goal / origin:
- Own PR because: <one of the four reasons, or "independent defect group">
- Owns: <file> (whole) · <file> (hunks …; intermediate: …)
- Declared changes folded in:
- Tests: <files / cases> · verify: <focused cmd>, then <full cmd>
- AC / out-of-scope (→ lands in <PR>):
- Risks / reviewer focus:
- Approx. production lines:

## Stack
### S1 — <title>
(same fields; base = layer below)

## Mixed files
| File | Hunk / lines | → PR | Note |
|---|---|---|---|

## Undeclared behaviour changes (reach beyond the feature)
| File:line | Change | Affected callers / subjects | → PR | Decision needed |
|---|---|---|---|---|

## Dropped (needs approval)
| File | Hunk | Why | Retained-code adjustment |
|---|---|---|---|

## Further splits (optional, not planned)
| Split | What it buys | Cost |
|---|---|---|

## Ticket mapping (proposed, not created)
<per PR: ticket key/type, respecting repo rules>

## Equivalence check
<integration branch recipe; the only allowed differences = approved drops + retained-code adjustments>

## Open decisions
1. …

## Unverified
<only facts not checkable from the repo, PR or tickets (e.g. live production data)>

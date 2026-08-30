# Code Research

This plugin assumes you have a code research capability available, but does not mandate any specific tool. Tool selection is deferred to your workspace convention — typically your `AGENTS.md` or `CLAUDE.md`.

## Contract

Every "research" step in this plugin must produce:

- **1 targeted query** scoped to the relevant project path (or an external query if researching libraries, APIs, or protocols)
- **Output:** a 1-2 line summary, each finding cited with:
  - `file:line` for internal codebase precedent.
  - `web:<URL>` for external documentation, library references, or standards.
- **Empty result is valid:** record "no precedent found for `<terms>`" (for local code) or "no documentation found for `<topic>`" (for web searches) — do not silently skip.

The agent decides which underlying tool fulfills the contract.

## Recommended implementations (in order)

### Local Codebase Research
1. **An MCP code research tool** if your environment provides one (e.g. ChunkHound, similar semantic indexes). These give the best signal-to-noise for "find the existing pattern" queries.
2. **Workspace search MCPs** that index the repo and return ranked snippets.
3. **Fallback:** `Grep` + `Read` against the project path. Slower and noisier but always available.

### External Research (Web Search)
1. **An MCP web research tool** if available (e.g., ChunkHound `websearch` which performs deep synthesis over fetched pages).
2. **Standard web search tools** (e.g., `search_web` or search engines).
3. **Direct documentation fetch** via browser/HTTP tools when exact documentation URLs are known.

## Workspace convention wins

If your workspace has a research convention (e.g. *"MUST use tool X before grep"*), follow it. The plugin's job is to demand evidence, not to dictate the tool.

## Output Requirements

1. Output a concise "Research Summary" before proceeding with any action that depends on it
2. The summary must cover:
   - What was found (relevant code, patterns, architecture) with `file:line`
   - What was NOT found (gaps the change needs to fill)
   - Which files are most relevant to the current task
3. Do NOT dump raw tool output — synthesize it into actionable intelligence

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Running research without scoping to a project path | Always scope to the relevant project to avoid noisy results |
| Skipping research because "I know this codebase" | Knowledge without evidence is assumption. Run the query. |
| Dumping raw tool output without summarizing | Synthesize first. The summary IS the deliverable. |
| Mandating a specific tool in a skill | Skills describe the contract. Tool choice belongs to the workspace. |

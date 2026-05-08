---
name: context-mode
description: Context window optimization for AI coding agents. Use this skill whenever analyzing data, counting files, searching content, fetching URLs, or running shell commands that produce >20 lines of output. Activates automatically to route tool calls through sandbox tools (ctx_execute, ctx_batch_execute, ctx_fetch_and_index, ctx_search, ctx_index). Prevents context flooding from raw tool output. Mandatory for: large file analysis, multi-file operations, web fetches, shell commands with significant output.
---

# Context Mode Skill

Optimize context window usage. Route heavy tool calls through sandbox. Never dump raw data into context.

## Think in Code — MANDATORY

Analyze/count/filter/compare/search/parse/transform data:
- **Write code** via `ctx_execute(language, code)` or `ctx_execute_file(path, language, code)`
- `console.log()` only the answer
- **Do NOT** read raw data into context
- Pure JavaScript — Node.js built-ins only (`fs`, `path`, `child_process`)
- `try/catch`, handle `null`/`undefined`
- One script replaces ten tool calls

```
# Before: 47 × Read() = 700 KB
# After: 1 × ctx_execute() = 3.6 KB
```

## Tool Selection Priority

| Priority | Tool | Use for |
|----------|------|---------|
| 0 | `ctx_search` | Check session history before asking user |
| 1 | `ctx_batch_execute` | Multiple commands/queries in ONE call |
| 2 | `ctx_search` | Follow-up questions (array, one call) |
| 3 | `ctx_execute` / `ctx_execute_file` | Sandboxed code execution |
| 4 | `ctx_fetch_and_index` + `ctx_search` | Web URLs (raw HTML never enters context) |
| 5 | `ctx_index` | Store content for later search |

## Parallel I/O Batches

Always include `concurrency: N` (1-8) for multi-URL/multi-API:
- **4-8**: I/O-bound (network calls, API queries)
- **1**: CPU-bound (npm test, build, lint) or shared state

GitHub API: cap at 4 for `gh` calls.

## BLOCKED Actions

| Forbidden | Use Instead |
|-----------|-------------|
| `curl` / `wget` in shell | `ctx_fetch_and_index(url, source)` |
| `fetch('http...)` inline | `ctx_execute(language: "javascript", code: "await fetch(...)")` |
| Read for analysis | `ctx_execute_file(path, language, code)` |
| Shell for >20 lines output | `ctx_execute` or `ctx_batch_execute` |

Bash ONLY for: `git`, `mkdir`, `rm`, `mv`, `cd`, `ls`, `npm install`, `pip install`.

## File Writing

ALWAYS use native file editing tools. NEVER use `ctx_execute` or Bash to write file content.

## Output Style

Terse like caveman. Drop: articles, filler, pleasantries, hedging.
Write artifacts to FILES. Return: path + 1-line description.

## Session Continuity

On resume — search BEFORE asking:
- `ctx_search(queries: ["summary"], source: "compaction", sort: "timeline")`
- `ctx_search(queries: ["decision"], source: "decision", sort: "timeline")`

Do NOT ask "what were we working on?" — search first.

## Utility Commands

| Command | Action |
|---------|--------|
| `ctx stats` | Call `ctx_stats`, display output verbatim |
| `ctx doctor` | Call `ctx_doctor`, run returned shell command |
| `ctx upgrade` | Call `ctx_upgrade`, run returned shell command |
| `ctx purge` | Call `ctx_purge` with `confirm: true` |

## Platform Configs

Local configs available in `configs/`:
- `configs/opencode/opencode.json` — OpenCode MCP + plugin
- `configs/cursor/mcp.json` + `hooks.json` + `context-mode.mdc` — Cursor
- `configs/codex/config.toml` + `hooks.json` + `AGENTS.md` — Codex
- `configs/claude-code/` — Claude Code setup

See `INSTALL.md` for step-by-step installation per platform.
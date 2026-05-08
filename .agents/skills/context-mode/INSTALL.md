# context-mode — Installation Guide

Context Mode installed locally. Configs in this directory.

## Prerequisite

```bash
npm install -g context-mode
# or: npx -y context-mode
```

Verify: `npx -y context-mode --version`

---

## OpenCode

1. Copy config:
   ```bash
   # Project level
   copy configs\opencode\opencode.json opencode.json

   # OR global
   copy configs\opencode\opencode.json %USERPROFILE%\.config\opencode\opencode.json
   ```

2. Copy routing rules:
   ```bash
   copy configs\opencode\AGENTS.md AGENTS.md
   ```

3. Restart OpenCode.

---

## Cursor

1. Copy MCP config:
   ```bash
   copy configs\cursor\mcp.json .cursor\mcp.json
   ```

2. Copy hooks:
   ```bash
   copy configs\cursor\hooks.json .cursor\hooks.json
   ```

3. Copy rules:
   ```bash
   copy configs\cursor\context-mode.mdc .cursor\rules\context-mode.mdc
   ```

4. Restart Cursor.

---

## Claude Code

1. Install via plugin marketplace:
   ```
   /plugin marketplace add mksglu/context-mode
   /plugin install context-mode@context-mode
   ```

2. Copy CLAUDE.md to global:
   ```bash
   copy configs\claude-code\CLAUDE.md %USERPROFILE%\.claude\CLAUDE.md
   ```

3. Restart Claude Code.

---

## Codex CLI

1. Add MCP to config:
   ```bash
   type configs\codex\config.toml >> %USERPROFILE%\.codex\config.toml
   ```

2. Create hooks:
   ```bash
   copy configs\codex\hooks.json %USERPROFILE%\.codex\hooks.json
   ```

3. Copy routing rules:
   ```bash
   copy configs\codex\AGENTS.md %USERPROFILE%\.codex\AGENTS.md
   ```

4. Restart Codex CLI.

---

## Verify

Run `ctx stats` in each platform. Should show:
- Connected MCP tools
- Context savings enabled
- Session tracking active

## Tools Available

| Tool | Description |
|------|-------------|
| `ctx_batch_execute` | Run multiple commands/queries in ONE call |
| `ctx_execute` | Run code in sandbox, only stdout enters context |
| `ctx_execute_file` | Process files in sandbox |
| `ctx_fetch_and_index` | Fetch URLs, raw HTML never enters context |
| `ctx_search` | Search indexed content with BM25 |
| `ctx_index` | Store content for later search |
| `ctx_stats` | Show context savings |
| `ctx_doctor` | Diagnose installation |
| `ctx_upgrade` | Upgrade context-mode |
| `ctx_purge` | Delete all indexed content |
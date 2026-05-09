# Quickstart: Claude Code + Kayron MCP Integration

**Target Audience**: Developers wanting to use Kayron AI trading tools in Claude Code IDE  
**Time to First Trade**: ~5 minutes  
**Prerequisites**: Claude Code IDE (latest), Kayron AI MCP server running, MT5 terminal + account

---

## Step 1: Verify Kayron MCP Server is Running

Start Kayron AI MCP daemon on your machine (or remote LAN):

```bash
# On the machine running MT5 terminal
./cmd/mcp-mt5-server &

# Check server is responding
curl -X POST http://localhost:50051 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","params":{},"id":1}'

# Should respond with list of available tools
```

Expected output: JSON list of 10+ trading tools (place-order, get-quote, close-position, etc.)

---

## Step 2: Configure Claude Code to Connect

Open Claude Code and add Kayron MCP config to `settings.json`:

```json
{
  "mcp.kayron": {
    "enabled": true,
    "host": "localhost",
    "port": 50051,
    "apiKey": "your-api-key-here",
    "cacheTtlMinutes": 60,
    "logLevel": "info"
  }
}
```

**Get your API key**:
- Set environment variable: `export KAYRON_API_KEY="your-key"`
- Or paste into settings (not recommended for production, use env var instead)

**Save settings.json** → Claude Code auto-reloads configuration

---

## Step 3: Verify Connection

Open Command Palette (`Cmd+Shift+P` / `Ctrl+Shift+P`) and run:

```
Kayron: Status
```

Expected output in IDE output panel:
```
✓ Connected to MCP server (localhost:50051)
✓ Loaded 12 trading tools
✓ Schema cache ready (~/.claude/cache/kayron-tools.json)
```

If connection fails, check:
- MCP server running: `ps aux | grep mcp-mt5-server`
- Port correct in settings.json
- API key valid (check logs: `tail -f ~/.claude/logs/kayron-mcp.log`)

---

## Step 4: List Available Tools

Run:

```
Kayron: Tools List
```

Expected output: Table of all available tools with descriptions and input/output schemas.

Example:
```
| Tool Name       | Description                      | Inputs              |
|-----------------|----------------------------------|---------------------|
| place-order     | Place new market or limit order  | symbol, volume, ... |
| get-quote       | Get current bid/ask for symbol   | symbol              |
| close-position  | Close open position by ticket    | ticket              |
| account-info    | Get account balance + margin     | (none)              |
...
```

---

## Step 5: Execute Your First Trade

### Option A: Quick Order (Command Palette)

Run:

```
Kayron: Place Order
```

IDE prompts:
```
? Enter symbol (e.g., EURUSD): EURUSD
? Enter volume (e.g., 0.1): 0.1
? Order type (BUY/SELL): BUY
? Order price (market/limit): market
```

Response in output panel:
```json
{
  "ticket": 12345,
  "symbol": "EURUSD",
  "type": "BUY",
  "volume": 0.1,
  "entryPrice": 1.0850,
  "status": "filled",
  "timestamp": "2026-05-08T10:30:45Z"
}
```

✓ **First order placed!**

### Option B: Using Hotkey

If configured in settings.json:

```json
"mcp.kayron.hotkeys": {
  "place-order": "cmd+k m"
}
```

Press `Cmd+K`, then `M` → same dialog appears

---

## Step 6: View Live Positions

Open side panel (View → Kayron Positions, or icon in status bar) to see:

- **Symbol** | **Entry Price** | **Volume** | **P&L** | **Actions**
- EURUSD | 1.0850 | 0.1 lot | +$50 | [Close] [Modify]

Panel auto-refreshes every 5 seconds. Click **[Close]** to close position.

---

## Step 7: Create a Reusable Skill

Create file `~/.claude/skills/my-first-skill/SKILL.md`:

```markdown
---
name: close-eurusd
description: Close all EURUSD positions
---

# Close All EURUSD Positions

First, get list of open positions:

/kayron positions-list

Then, for each position with symbol "EURUSD", close it:

/kayron close-position {"ticket": 12345}

Log result to audit trail.
```

**Load skill**: 
- Run `Kayron: Reload Skills` in command palette
- Skill appears as `Kayron: My First Skill` command

**Execute skill**:
- Run `Kayron: Close EURUSD` → skill executes, logs result

✓ **Automated trading workflow created!**

---

## Step 8: Check Logs

View all executed trades in audit log:

```bash
cat ~/.claude/logs/kayron-mcp.log | tail -20
```

Each line is JSON with: timestamp, tool, inputs, outputs, errors, retry count

Example:
```json
{"timestamp":"2026-05-08T10:30:45Z","tool":"place-order","inputs":{"symbol":"EURUSD","volume":0.1,"type":"BUY"},"output":{"ticket":12345,"entryPrice":1.0850},"error":null,"durationMs":245,"retryCount":0}
```

---

## Troubleshooting

### MCP Server Unreachable
```
❌ Error: Failed to connect to MCP server (localhost:50051)
```
→ Check server running: `ps aux | grep mcp-mt5-server`  
→ Check port correct in settings.json  
→ Check firewall allows localhost:50051

### API Key Invalid
```
❌ Error: Authentication failed (invalid API key)
```
→ Check `apiKey` in settings.json  
→ Or set env var: `export KAYRON_API_KEY="..."`

### Order Execution Failed
```
❌ Error: INSUFFICIENT_MARGIN (need $500, have $200)
```
→ Account needs more margin  
→ Add funds to MT5 account  
→ Or place smaller order

### Cache Stale
```
⚠️ Warning: Schema cache is 2 hours old
```
→ Run `Kayron: Refresh Schema` to refresh  
→ Or wait 1 hour for auto-refresh

---

## Next Steps

- **Learn Skill Format**: Read `~/.claude/skills/kayron-demo/SKILL.md` (bundled example)
- **Advanced Hotkeys**: Configure chord hotkeys in `settings.json` (see documentation)
- **Real-Time Updates**: Position panel polls every 5s; manual refresh available
- **Audit Trail**: All trades logged to `~/.claude/logs/kayron-mcp.log` (JSON format)
- **Error Recovery**: IDE auto-reconnects on network failure (up to 5 retries)

---

## Quick Reference

| Task | Command |
|------|---------|
| Check status | `Kayron: Status` |
| List tools | `Kayron: Tools List` |
| Place order | `Kayron: Place Order` (or hotkey `Cmd+K M`) |
| View positions | Side panel: Kayron Positions |
| Close position | Click [Close] in position panel |
| Create skill | Create `~/.claude/skills/[name]/SKILL.md` |
| Load skill | `Kayron: Reload Skills` |
| View logs | `tail -f ~/.claude/logs/kayron-mcp.log` |
| Refresh schema | `Kayron: Refresh Schema` |

---

**Congratulations!** You're now using Kayron AI MCP within Claude Code IDE. Happy trading! 🚀

# Contract: MCP Client Interface for Claude Code

**Phase**: 1 (Design)  
**Format**: TypeScript Interface + JSON Schema  
**Purpose**: Define boundary between Claude Code IDE and Kayron AI MCP server

---

## MCP Client Wrapper Interface

```typescript
interface MCPClient {
  // Connection lifecycle
  connect(config: MCPConfig): Promise<void>
  disconnect(): Promise<void>
  isConnected(): boolean
  
  // Tool discovery & metadata
  listTools(): Promise<ToolDefinition[]>
  getTool(toolName: string): Promise<ToolDefinition | null>
  
  // Tool invocation
  invokeTool(
    toolName: string,
    params: Record<string, unknown>,
    options?: ToolInvocationOptions
  ): Promise<ToolExecutionResult>
  
  // Schema caching
  cacheTools(tools: ToolDefinition[]): Promise<void>
  getCachedTools(): Promise<ToolDefinition[] | null>
  clearCache(): Promise<void>
  
  // Pending operations queue
  queueOperation(op: PendingOperation): Promise<void>
  getPendingOperations(): Promise<PendingOperation[]>
  replayOperations(): Promise<ReplayResult[]>
  
  // Events
  on(event: 'connected' | 'disconnected' | 'error' | 'tool-response', 
     handler: (data: unknown) => void): void
  off(event: string, handler: (data: unknown) => void): void
}

interface MCPConfig {
  host: string
  port: number
  apiKey: string
  cacheTtlMinutes?: number
  maxRetries?: number
  backoffMs?: number
}

interface ToolDefinition {
  name: string
  description: string
  inputSchema: JSONSchema
  outputSchema: JSONSchema
  version: string
  category?: string // e.g., "trading", "account", "market"
}

interface ToolInvocationOptions {
  timeout?: number // milliseconds
  idempotencyKey?: string // UUID, generated if not provided
  retryCount?: number // 0-5
  queueIfOffline?: boolean // persist if connection fails
}

interface ToolExecutionResult {
  status: 'success' | 'error' | 'timeout'
  output?: unknown // validated against tool's outputSchema
  error?: {
    code: string // e.g., "INSUFFICIENT_MARGIN", "INVALID_SYMBOL"
    message: string
    details?: unknown
  }
  durationMs: number
  idempotencyKey: string
}

interface PendingOperation {
  id: string // UUID
  toolName: string
  params: Record<string, unknown>
  createdAt: ISO8601Timestamp
  retryCount: number
  idempotencyKey: string
}

interface ReplayResult {
  operationId: string
  status: 'success' | 'failed'
  result?: ToolExecutionResult
  error?: string
}
```

---

## JSON-RPC 2.0 Protocol (IDE ↔ MCP Server)

### Request Format

```json
{
  "jsonrpc": "2.0",
  "method": "tools/invoke",
  "params": {
    "tool": "place-order",
    "input": {
      "symbol": "EURUSD",
      "volume": 0.1,
      "type": "BUY",
      "price": "market"
    },
    "idempotencyKey": "550e8400-e29b-41d4-a716-446655440000"
  },
  "id": "req-12345"
}
```

### Success Response

```json
{
  "jsonrpc": "2.0",
  "result": {
    "ticket": 12345,
    "symbol": "EURUSD",
    "type": "BUY",
    "volume": 0.1,
    "entryPrice": 1.0850,
    "status": "filled",
    "timestamp": "2026-05-08T10:30:45Z"
  },
  "id": "req-12345"
}
```

### Error Response

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32000,
    "message": "INSUFFICIENT_MARGIN",
    "data": {
      "marginRequired": 500,
      "marginAvailable": 200,
      "retryable": false
    }
  },
  "id": "req-12345"
}
```

### Standard Error Codes

| Error Code | Name | Retryable | Meaning |
|-----------|------|-----------|---------|
| -32000 | INSUFFICIENT_MARGIN | false | Account margin insufficient |
| -32001 | INVALID_SYMBOL | false | Symbol not found in market watch |
| -32002 | MARKET_CLOSED | false | Market closed, cannot place order |
| -32003 | INVALID_VOLUME | false | Volume violates position limits |
| -32010 | NETWORK_TIMEOUT | true | Network timeout, safe to retry |
| -32011 | SERVER_UNAVAILABLE | true | MCP server temporarily unavailable |
| -32100 | INVALID_REQUEST | false | Malformed request |
| -32101 | AUTHENTICATION_FAILED | false | API key invalid or expired |

---

## Tool Discovery Contract

### Request

```json
{
  "jsonrpc": "2.0",
  "method": "tools/list",
  "params": {},
  "id": "req-list"
}
```

### Response

```json
{
  "jsonrpc": "2.0",
  "result": {
    "tools": [
      {
        "name": "place-order",
        "description": "Place new market or limit order",
        "inputSchema": {
          "type": "object",
          "properties": {
            "symbol": { "type": "string", "description": "e.g., EURUSD" },
            "volume": { "type": "number", "description": "Lot size, e.g., 0.1" },
            "type": { "enum": ["BUY", "SELL"] },
            "price": { "oneOf": [
              { "const": "market" },
              { "type": "number", "description": "Limit price" }
            ]}
          },
          "required": ["symbol", "volume", "type", "price"]
        },
        "outputSchema": {
          "type": "object",
          "properties": {
            "ticket": { "type": "integer" },
            "symbol": { "type": "string" },
            "entryPrice": { "type": "string", "pattern": "^\\d+\\.\\d+$" },
            "status": { "enum": ["filled", "pending"] }
          },
          "required": ["ticket", "symbol", "entryPrice", "status"]
        },
        "version": "1.0.0"
      },
      {
        "name": "get-quote",
        "description": "Get current bid/ask prices",
        "inputSchema": {
          "type": "object",
          "properties": {
            "symbol": { "type": "string" }
          },
          "required": ["symbol"]
        },
        "outputSchema": {
          "type": "object",
          "properties": {
            "symbol": { "type": "string" },
            "bid": { "type": "string", "pattern": "^\\d+\\.\\d+$" },
            "ask": { "type": "string", "pattern": "^\\d+\\.\\d+$" },
            "timestamp": { "type": "string", "format": "date-time" }
          },
          "required": ["symbol", "bid", "ask", "timestamp"]
        },
        "version": "1.0.0"
      }
    ]
  },
  "id": "req-list"
}
```

---

## Idempotency Contract

Every tool invocation that modifies state (place-order, close-position, modify-order, cancel-order) **MUST** include `idempotencyKey`.

### Duplicate Detection & Response

If client retries with same `idempotencyKey`:

```json
{
  "jsonrpc": "2.0",
  "result": {
    "ticket": 12345,
    "isDuplicate": true,
    "originalRequestId": "req-12345",
    "originalResponse": { /* original result */ }
  },
  "id": "req-retry"
}
```

**Guarantee**: Exactly-once execution. Client sees same result on retry (no double-order).

---

## Streaming Subscriptions (Optional, for v2)

Reserved contract for future real-time position updates.

```json
{
  "jsonrpc": "2.0",
  "method": "subscribe",
  "params": {
    "event": "positions:update",
    "filter": { "symbol": "EURUSD" }
  },
  "id": "sub-1"
}
```

Response: Stream of position updates as server-sent events (SSE) or WebSocket messages.

**Not implemented in v1.** IDE uses polling instead.

---

## Logging & Audit Contract

All executions logged to `~/.claude/logs/kayron-mcp.log` (JSONL format):

```json
{
  "timestamp": "2026-05-08T10:30:45.123Z",
  "requestId": "req-12345",
  "tool": "place-order",
  "input": { "symbol": "EURUSD", "volume": 0.1, "type": "BUY", "price": "market" },
  "output": { "ticket": 12345, "entryPrice": 1.0850, "status": "filled" },
  "error": null,
  "durationMs": 245,
  "idempotencyKey": "550e8400-e29b-41d4-a716-446655440000",
  "retryCount": 0
}
```

**Format**: One JSON object per line, appended sequentially.  
**Retention**: Indefinite (user manages rotation).  
**Privacy**: No sensitive data (API key, credentials) logged — only tool invocations + results.

---

## Next Steps

Contract complete. Implementation proceeds in Phase 2 with TDD approach.

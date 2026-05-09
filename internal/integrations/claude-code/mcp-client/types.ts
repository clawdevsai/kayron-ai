/**
 * MCP Client Type Definitions
 * Core interfaces for MCP communication and tool execution
 */

export interface MCPConfig {
  host: string;
  port: number;
  apiKey: string;
  cacheTtlMinutes?: number;
  maxRetries?: number;
  backoffMs?: number;
  logLevel?: 'debug' | 'info' | 'warn' | 'error';
  reconnectMaxRetries?: number;
}

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: JSONSchema;
  outputSchema: JSONSchema;
  version: string;
  category?: string;
}

export interface JSONSchema {
  type: string;
  properties?: Record<string, unknown>;
  required?: string[];
  description?: string;
  [key: string]: unknown;
}

export interface ToolInvocationOptions {
  timeout?: number;
  idempotencyKey?: string;
  retryCount?: number;
  queueIfOffline?: boolean;
}

export interface ToolExecutionResult {
  status: 'success' | 'error' | 'timeout';
  output?: unknown;
  error?: ToolError;
  durationMs: number;
  idempotencyKey: string;
}

export interface ToolError {
  code: string;
  message: string;
  details?: unknown;
  retryable?: boolean;
}

export interface PendingOperation {
  id: string;
  toolName: string;
  params: Record<string, unknown>;
  createdAt: string;
  retryCount: number;
  idempotencyKey: string;
}

export interface ReplayResult {
  operationId: string;
  status: 'success' | 'failed';
  result?: ToolExecutionResult;
  error?: string;
}

export interface Position {
  ticket: number;
  symbol: string;
  type: 'buy' | 'sell';
  volume: string;
  entryPrice: string;
  currentPrice: string;
  pnl: string;
  pnlPercent: string;
  openTime: string;
  lastUpdateTime: string;
}

export interface ExecutionLog {
  id: string;
  timestamp: string;
  toolName: string;
  inputParams: Record<string, unknown>;
  output?: Record<string, unknown> | string;
  error?: Record<string, unknown> | string;
  executionDurationMs: number;
  retryCount: number;
  idempotencyKey: string;
  userId?: string;
}

export interface SkillDefinition {
  id: string;
  name: string;
  description: string;
  skillPath: string;
  content: string;
  toolDependencies: string[];
  createdAt: string;
  modifiedAt: string;
  enabled: boolean;
}

export type MCPConnectionStatus = 'connected' | 'disconnected' | 'reconnecting' | 'error';

export interface MCPConnection {
  id: string;
  host: string;
  port: number;
  status: MCPConnectionStatus;
  lastConnectTime?: string;
  lastDisconnectTime?: string;
  reconnectAttempts: number;
}

export interface MCPClientInterface {
  // Connection lifecycle
  connect(config: MCPConfig): Promise<void>;
  disconnect(): Promise<void>;
  isConnected(): boolean;
  getStatus(): MCPConnectionStatus;

  // Tool discovery
  listTools(): Promise<ToolDefinition[]>;
  getTool(toolName: string): Promise<ToolDefinition | null>;

  // Tool invocation
  invokeTool(
    toolName: string,
    params: Record<string, unknown>,
    options?: ToolInvocationOptions
  ): Promise<ToolExecutionResult>;

  // Caching
  cacheTools(tools: ToolDefinition[]): Promise<void>;
  getCachedTools(): Promise<ToolDefinition[] | null>;
  clearCache(): Promise<void>;

  // Queue management
  queueOperation(op: PendingOperation): Promise<void>;
  getPendingOperations(): Promise<PendingOperation[]>;
  replayOperations(): Promise<ReplayResult[]>;

  // Events
  on(
    event: 'connected' | 'disconnected' | 'error' | 'tool-response',
    handler: (data: unknown) => void
  ): void;
  off(event: string, handler: (data: unknown) => void): void;
}

// JSON-RPC 2.0 Protocol Types
export interface JSONRPCRequest {
  jsonrpc: '2.0';
  method: string;
  params?: unknown;
  id: string | number;
}

export interface JSONRPCResponse<T = unknown> {
  jsonrpc: '2.0';
  result?: T;
  error?: JSONRPCError;
  id: string | number;
}

export interface JSONRPCError {
  code: number;
  message: string;
  data?: unknown;
}

// Error Codes
export enum ErrorCode {
  INSUFFICIENT_MARGIN = 'INSUFFICIENT_MARGIN',
  INVALID_SYMBOL = 'INVALID_SYMBOL',
  MARKET_CLOSED = 'MARKET_CLOSED',
  INVALID_VOLUME = 'INVALID_VOLUME',
  NETWORK_TIMEOUT = 'NETWORK_TIMEOUT',
  SERVER_UNAVAILABLE = 'SERVER_UNAVAILABLE',
  INVALID_REQUEST = 'INVALID_REQUEST',
  AUTHENTICATION_FAILED = 'AUTHENTICATION_FAILED',
  UNKNOWN_ERROR = 'UNKNOWN_ERROR',
}

// Events
export type MCPEvent =
  | { type: 'connected' }
  | { type: 'disconnected'; reason?: string }
  | { type: 'error'; error: ToolError }
  | { type: 'tool-response'; result: ToolExecutionResult }
  | { type: 'reconnecting'; attempt: number; maxAttempts: number }
  | { type: 'queue-update'; pendingCount: number };

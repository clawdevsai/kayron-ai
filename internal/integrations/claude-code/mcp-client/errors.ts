/**
 * Error Handling & Error Code Translation
 * Maps MCP errors to user-friendly messages
 */

import { ErrorCode, ToolError } from './types';

export const ERROR_CODE_MAP: Record<string, { retryable: boolean; message: string }> = {
  [ErrorCode.INSUFFICIENT_MARGIN]: {
    retryable: false,
    message: 'Insufficient margin. Add funds to your account or reduce order size.',
  },
  [ErrorCode.INVALID_SYMBOL]: {
    retryable: false,
    message: 'Invalid symbol. Verify the instrument is available in market watch.',
  },
  [ErrorCode.MARKET_CLOSED]: {
    retryable: false,
    message: 'Market is closed. Wait for market open to place orders.',
  },
  [ErrorCode.INVALID_VOLUME]: {
    retryable: false,
    message: 'Invalid volume. Check position limits and minimum/maximum lot sizes.',
  },
  [ErrorCode.NETWORK_TIMEOUT]: {
    retryable: true,
    message: 'Network timeout. Will retry automatically with exponential backoff.',
  },
  [ErrorCode.SERVER_UNAVAILABLE]: {
    retryable: true,
    message: 'MCP server temporarily unavailable. Will retry automatically.',
  },
  [ErrorCode.INVALID_REQUEST]: {
    retryable: false,
    message: 'Invalid request format. Check input parameters against tool schema.',
  },
  [ErrorCode.AUTHENTICATION_FAILED]: {
    retryable: false,
    message: 'Authentication failed. Verify API key in settings.json or KAYRON_API_KEY env var.',
  },
  [ErrorCode.UNKNOWN_ERROR]: {
    retryable: false,
    message: 'Unknown error. Check logs for details.',
  },
};

/**
 * Classify error as retryable or permanent
 */
export function isRetryableError(error: ToolError): boolean {
  const info = ERROR_CODE_MAP[error.code];
  return info?.retryable || false;
}

/**
 * Get user-friendly error message
 */
export function getUserFriendlyMessage(error: ToolError): string {
  const info = ERROR_CODE_MAP[error.code];
  return info?.message || error.message;
}

/**
 * Translate MCP error response to ToolError
 */
export function translateMCPError(response: {
  error?: {
    code: number;
    message: string;
    data?: unknown;
  };
}): ToolError | null {
  if (!response.error) return null;

  const code = mapErrorCode(response.error.code);
  const info = ERROR_CODE_MAP[code];

  return {
    code,
    message: response.error.message,
    details: response.error.data,
    retryable: info?.retryable || false,
  };
}

/**
 * Map JSON-RPC error codes to ErrorCode enum
 */
function mapErrorCode(jsonRpcCode: number): string {
  const codeMap: Record<number, string> = {
    [-32000]: ErrorCode.INSUFFICIENT_MARGIN,
    [-32001]: ErrorCode.INVALID_SYMBOL,
    [-32002]: ErrorCode.MARKET_CLOSED,
    [-32003]: ErrorCode.INVALID_VOLUME,
    [-32010]: ErrorCode.NETWORK_TIMEOUT,
    [-32011]: ErrorCode.SERVER_UNAVAILABLE,
    [-32100]: ErrorCode.INVALID_REQUEST,
    [-32101]: ErrorCode.AUTHENTICATION_FAILED,
  };

  return codeMap[jsonRpcCode] || ErrorCode.UNKNOWN_ERROR;
}

/**
 * Custom error class for MCP operations
 */
export class MCPError extends Error {
  constructor(
    public code: string,
    public retryable: boolean,
    public details?: unknown
  ) {
    super(getUserFriendlyMessage({ code, message: '', details }));
    this.name = 'MCPError';
  }
}

/**
 * Validation errors
 */
export class ValidationError extends Error {
  constructor(
    public field: string,
    public reason: string
  ) {
    super(`Validation failed for ${field}: ${reason}`);
    this.name = 'ValidationError';
  }
}

/**
 * Schema errors
 */
export class SchemaError extends Error {
  constructor(
    public tool: string,
    public reason: string
  ) {
    super(`Schema error for tool ${tool}: ${reason}`);
    this.name = 'SchemaError';
  }
}

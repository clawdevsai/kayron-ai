/**
 * Tool Invoker: Wrapper around MCP tool execution
 * Adds validation, idempotency, retries, audit logging
 */

import { v4 as uuidv4 } from 'uuid';
import { ToolDefinition, ExecutionLog } from './types';
import { SchemaValidator, ValidationResult } from './schema-validator';
import { Logger } from './logger';
import { isRetryableError } from './errors';

export interface InvokeResult {
  success: boolean;
  data?: any;
  error?: { code: string; message: string; details?: any };
  retryCount?: number;
}

export interface InvokeOptions {
  maxRetries?: number;
}

export class ToolInvoker {
  private validator: SchemaValidator;
  private logger: Logger;
  private invokedKeys: Set<string> = new Set();
  private executionLogs: ExecutionLog[] = [];

  constructor(validator: SchemaValidator, logger: Logger) {
    this.validator = validator;
    this.logger = logger;
  }

  /**
   * Invoke tool with full pipeline: validate, generate key, check duplicate, invoke, log
   */
  async invoke(tool: ToolDefinition, params: Record<string, unknown>): Promise<InvokeResult> {
    return this.invokeWithRetry(tool, params, { maxRetries: 5 });
  }

  /**
   * Invoke tool with retry logic
   */
  async invokeWithRetry(
    tool: ToolDefinition,
    params: Record<string, unknown>,
    options?: InvokeOptions
  ): Promise<InvokeResult> {
    const maxRetries = options?.maxRetries ?? 5;

    // Step 1: Validate parameters
    const validation = this.validateParams(tool, params);
    if (!validation.valid) {
      this.logger.error(`Tool validation failed: ${tool.name}`, {
        errors: validation.errors,
      });
      return {
        success: false,
        error: {
          code: 'INVALID_REQUEST',
          message: `Validation failed: ${validation.errors[0]}`,
        },
        retryCount: 0,
      };
    }

    // Step 2: Generate idempotency key
    const idempotencyKey = this.generateIdempotencyKey(tool.name, params);

    // Step 3: Check for duplicate
    if (this.isDuplicate(idempotencyKey)) {
      this.logger.warn(`Duplicate invocation detected: ${tool.name}`, {
        idempotencyKey,
      });
      return {
        success: false,
        error: {
          code: 'DUPLICATE_REQUEST',
          message: 'Duplicate invocation detected',
        },
        retryCount: 0,
      };
    }

    // Step 4: Invoke with retries
    const startTime = Date.now();
    let lastError: any = null;
    let retryCount = 0;

    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        const result = await this.executeToolCall(tool, params);

        if (result.success) {
          // Step 5: Record invocation
          this.recordInvocation(idempotencyKey, result.data);

          // Step 6: Log execution
          const duration = Date.now() - startTime;
          this.logExecution({
            toolName: tool.name,
            inputParams: params,
            output: result.data,
            error: null,
            executionDurationMs: duration,
            retryCount: attempt,
            idempotencyKey,
          });

          return {
            success: true,
            data: result.data,
            retryCount: attempt,
          };
        } else {
          lastError = result.error;

          // Check if retryable
          if (!this.isRetryable(lastError) || attempt === maxRetries) {
            // Permanent error or max retries reached
            const duration = Date.now() - startTime;
            this.logExecution({
              toolName: tool.name,
              inputParams: params,
              output: null,
              error: lastError.code,
              executionDurationMs: duration,
              retryCount: attempt,
              idempotencyKey,
            });

            return {
              success: false,
              error: lastError,
              retryCount: attempt,
            };
          }

          retryCount = attempt + 1;

          // Calculate backoff
          const backoff = this.calculateBackoff(attempt);
          this.logger.warn(`Tool invocation failed, retrying: ${tool.name}`, {
            attempt: attempt + 1,
            backoffMs: backoff,
            error: lastError.code,
          });

          // Wait before retry
          await this.delay(backoff);
        }
      } catch (err) {
        lastError = {
          code: 'INTERNAL_ERROR',
          message: (err as Error).message,
        };

        if (attempt === maxRetries) {
          const duration = Date.now() - startTime;
          this.logExecution({
            toolName: tool.name,
            inputParams: params,
            output: null,
            error: lastError.code,
            executionDurationMs: duration,
            retryCount: attempt,
            idempotencyKey,
          });

          return {
            success: false,
            error: lastError,
            retryCount: attempt,
          };
        }

        retryCount = attempt + 1;
        const backoff = this.calculateBackoff(attempt);
        await this.delay(backoff);
      }
    }

    return {
      success: false,
      error: lastError,
      retryCount: maxRetries,
    };
  }

  /**
   * Validate parameters against tool schema
   */
  validateParams(tool: ToolDefinition, params: Record<string, unknown>): ValidationResult {
    return this.validator.validate(tool.inputSchema, params);
  }

  /**
   * Generate idempotency key from tool name + params hash
   */
  generateIdempotencyKey(toolName: string, params: Record<string, unknown>): string {
    const hash = JSON.stringify({ toolName, params });
    return `${toolName}-${Buffer.from(hash).toString('base64').substring(0, 12)}`;
  }

  /**
   * Check if invocation already recorded
   */
  isDuplicate(idempotencyKey: string): boolean {
    return this.invokedKeys.has(idempotencyKey);
  }

  /**
   * Record successful invocation
   */
  recordInvocation(idempotencyKey: string, output: any): void {
    this.invokedKeys.add(idempotencyKey);
  }

  /**
   * Classify error as retryable
   */
  isRetryable(error: { code: string; message: string }): boolean {
    return isRetryableError(error as any);
  }

  /**
   * Get user-friendly error message
   */
  getUserMessage(error: { code: string; message: string }): string {
    const messages: Record<string, string> = {
      INSUFFICIENT_MARGIN: 'Insufficient margin. Add funds or reduce order size.',
      INVALID_SYMBOL: 'Invalid symbol. Check instrument availability.',
      MARKET_CLOSED: 'Market is closed. Trade during market hours.',
      INVALID_VOLUME: 'Invalid volume. Check position limits.',
      NETWORK_TIMEOUT: 'Network timeout. Will retry automatically.',
      SERVER_UNAVAILABLE: 'Server unavailable. Will retry automatically.',
      INVALID_REQUEST: 'Invalid request. Check input parameters.',
      AUTHENTICATION_FAILED: 'Authentication failed. Check API key.',
    };

    return messages[error.code] || error.message;
  }

  /**
   * Log execution for audit trail
   */
  logExecution(log: ExecutionLog): void {
    this.executionLogs.push(log);
    this.logger.logExecution(log);
  }

  /**
   * Get last execution log
   */
  getLastLog(): ExecutionLog | undefined {
    return this.executionLogs[this.executionLogs.length - 1];
  }

  /**
   * Parse tool response
   */
  parseResponse(response: any, tool: ToolDefinition): InvokeResult {
    if (response.error) {
      return {
        success: false,
        error: {
          code: response.error.code || 'UNKNOWN_ERROR',
          message: response.error.message || 'Unknown error',
          details: response.error.data,
        },
      };
    }

    return {
      success: true,
      data: response.result,
    };
  }

  /**
   * Parse error response
   */
  parseError(response: any): { code: string; message: string; details?: any } {
    return {
      code: response.error?.code || 'UNKNOWN_ERROR',
      message: response.error?.message || 'Unknown error',
      details: response.error?.data,
    };
  }

  /**
   * Execute actual tool call (would call MCPClient.invokeTool)
   * This is a placeholder - in real implementation, would use MCPClient
   */
  private async executeToolCall(
    tool: ToolDefinition,
    params: Record<string, unknown>
  ): Promise<InvokeResult> {
    // This would be replaced with actual MCPClient.invokeTool call
    // For now, returns placeholder
    return {
      success: true,
      data: {
        ticket: Math.floor(Math.random() * 1000000),
        status: 'FILLED',
      },
    };
  }

  /**
   * Calculate exponential backoff
   */
  private calculateBackoff(attempt: number): number {
    const baseMs = 1000;
    const maxMs = 32000;
    const backoff = baseMs * Math.pow(2, attempt);
    return Math.min(backoff, maxMs);
  }

  /**
   * Delay execution
   */
  private delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

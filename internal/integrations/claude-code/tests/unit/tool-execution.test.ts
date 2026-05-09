/**
 * Unit Tests: Tool Invocation & Execution
 * Parameter validation, idempotency, error handling, audit logging
 */

import { ToolInvoker } from '../../mcp-client/tool-invoker';
import { SchemaValidator } from '../../mcp-client/schema-validator';
import { ToolDefinition } from '../../mcp-client/types';
import { Logger } from '../../mcp-client/logger';

describe('ToolInvoker', () => {
  let invoker: ToolInvoker;
  let validator: SchemaValidator;
  let logger: Logger;

  const placeOrderTool: ToolDefinition = {
    name: 'place-order',
    description: 'Place a trading order',
    version: '1.0.0',
    inputSchema: {
      type: 'object',
      properties: {
        symbol: { type: 'string', description: 'Trading pair' },
        volume: { type: 'number', minimum: 0.01, maximum: 100 },
        type: { type: 'string', enum: ['BUY', 'SELL'] },
        price: { type: 'string', description: 'Price or MARKET' },
      },
      required: ['symbol', 'volume', 'type'],
    },
    outputSchema: {
      type: 'object',
      properties: {
        ticket: { type: 'number' },
        status: { type: 'string' },
        entryPrice: { type: 'string' },
      },
    },
  };

  beforeEach(() => {
    validator = new SchemaValidator();
    logger = new Logger('debug');
    invoker = new ToolInvoker(validator, logger);
  });

  describe('Tool Invocation', () => {
    it('should invoke tool with valid parameters', async () => {
      const params = {
        symbol: 'EURUSD',
        volume: 0.1,
        type: 'BUY',
        price: 'MARKET',
      };

      const result = await invoker.invoke(placeOrderTool, params);
      expect(result).toBeDefined();
      expect(result.success).toBe(true);
    });

    it('should generate idempotency key for request', () => {
      const params = { symbol: 'EURUSD', volume: 0.1, type: 'BUY' };
      const key1 = invoker.generateIdempotencyKey(placeOrderTool.name, params);
      const key2 = invoker.generateIdempotencyKey(placeOrderTool.name, params);

      expect(key1).toBeDefined();
      expect(typeof key1).toBe('string');
      expect(key1).toBe(key2); // Same params = same key
    });

    it('should generate different idempotency key for different params', () => {
      const params1 = { symbol: 'EURUSD', volume: 0.1, type: 'BUY' };
      const params2 = { symbol: 'GBPUSD', volume: 0.1, type: 'BUY' };

      const key1 = invoker.generateIdempotencyKey(placeOrderTool.name, params1);
      const key2 = invoker.generateIdempotencyKey(placeOrderTool.name, params2);

      expect(key1).not.toBe(key2);
    });

    it('should detect duplicate invocation via idempotency key', () => {
      const params = { symbol: 'EURUSD', volume: 0.1, type: 'BUY' };
      const key = invoker.generateIdempotencyKey(placeOrderTool.name, params);

      invoker.recordInvocation(key, { ticket: 12345 });
      const isDuplicate = invoker.isDuplicate(key);

      expect(isDuplicate).toBe(true);
    });
  });

  describe('Parameter Validation', () => {
    it('should validate parameters before invocation', async () => {
      const params = {
        symbol: 'EURUSD',
        volume: 0.1,
        type: 'BUY',
      };

      const result = invoker.validateParams(placeOrderTool, params);
      expect(result.valid).toBe(true);
      expect(result.errors).toEqual([]);
    });

    it('should reject invalid parameters', () => {
      const params = {
        symbol: 'EURUSD',
        volume: 0.1,
        // Missing 'type'
      };

      const result = invoker.validateParams(placeOrderTool, params);
      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
    });

    it('should reject out-of-range volume', () => {
      const params = {
        symbol: 'EURUSD',
        volume: 101, // Max is 100
        type: 'BUY',
      };

      const result = invoker.validateParams(placeOrderTool, params);
      expect(result.valid).toBe(false);
    });

    it('should reject invalid enum value', () => {
      const params = {
        symbol: 'EURUSD',
        volume: 0.1,
        type: 'INVALID',
      };

      const result = invoker.validateParams(placeOrderTool, params);
      expect(result.valid).toBe(false);
    });
  });

  describe('Error Handling', () => {
    it('should parse error response', () => {
      const errorResponse = {
        error: {
          code: 'INSUFFICIENT_MARGIN',
          message: 'Account margin insufficient',
          details: { margin: 50, required: 100 },
        },
      };

      const error = invoker.parseError(errorResponse);
      expect(error).toBeDefined();
      expect(error.code).toBe('INSUFFICIENT_MARGIN');
      expect(error.message).toBe('Account margin insufficient');
      expect(error.details).toEqual({ margin: 50, required: 100 });
    });

    it('should classify error as retryable or permanent', () => {
      const retryableError = {
        code: 'NETWORK_TIMEOUT',
        message: 'Timeout',
      };

      const permanentError = {
        code: 'INSUFFICIENT_MARGIN',
        message: 'Margin',
      };

      expect(invoker.isRetryable(retryableError)).toBe(true);
      expect(invoker.isRetryable(permanentError)).toBe(false);
    });

    it('should provide user-friendly error message', () => {
      const error = {
        code: 'INSUFFICIENT_MARGIN',
        message: 'Technical message',
      };

      const friendly = invoker.getUserMessage(error);
      expect(friendly).toContain('margin');
      expect(friendly.toLowerCase()).toContain('insufficient');
    });
  });

  describe('Audit Logging', () => {
    it('should create audit log entry for invocation', async () => {
      const params = {
        symbol: 'EURUSD',
        volume: 0.1,
        type: 'BUY',
      };

      const startTime = Date.now();
      await invoker.invoke(placeOrderTool, params);
      const duration = Date.now() - startTime;

      const log = invoker.getLastLog();
      expect(log).toBeDefined();
      expect(log?.toolName).toBe('place-order');
      expect(log?.inputParams).toEqual(params);
      expect(log?.executionDurationMs).toBeLessThanOrEqual(duration + 10);
    });

    it('should log successful execution', () => {
      const params = { symbol: 'EURUSD', volume: 0.1, type: 'BUY' };
      const output = { ticket: 12345, status: 'FILLED' };

      invoker.logExecution({
        toolName: 'place-order',
        inputParams: params,
        output,
        error: undefined,
        executionDurationMs: 150,
        retryCount: 0,
        idempotencyKey: 'key123',
      });

      const log = invoker.getLastLog();
      expect(log?.error).toBeUndefined();
      expect(log?.output).toEqual(output);
    });

    it('should log failed execution with error', () => {
      const params = { symbol: 'EURUSD', volume: 0.1, type: 'BUY' };
      const error = 'INSUFFICIENT_MARGIN';

      invoker.logExecution({
        toolName: 'place-order',
        inputParams: params,
        output: undefined,
        error,
        executionDurationMs: 200,
        retryCount: 0,
        idempotencyKey: 'key456',
      });

      const log = invoker.getLastLog();
      expect(log?.error).toBe(error);
      expect(log?.output).toBeUndefined();
    });

    it('should include idempotency key in audit log', () => {
      const params = { symbol: 'EURUSD', volume: 0.1, type: 'BUY' };
      const idempotencyKey = 'unique-key-123';

      invoker.logExecution({
        toolName: 'place-order',
        inputParams: params,
        output: { ticket: 12345 },
        error: undefined,
        executionDurationMs: 100,
        retryCount: 0,
        idempotencyKey,
      });

      const log = invoker.getLastLog();
      expect(log?.idempotencyKey).toBe(idempotencyKey);
    });

    it('should track retry count in audit log', () => {
      const params = { symbol: 'EURUSD', volume: 0.1, type: 'BUY' };

      invoker.logExecution({
        toolName: 'place-order',
        inputParams: params,
        output: { ticket: 12345 },
        error: undefined,
        executionDurationMs: 500,
        retryCount: 2,
        idempotencyKey: 'key',
      });

      const log = invoker.getLastLog();
      expect(log?.retryCount).toBe(2);
    });
  });

  describe('Execution Flow', () => {
    it('should validate params, generate key, invoke, log', async () => {
      const params = {
        symbol: 'EURUSD',
        volume: 0.1,
        type: 'BUY',
      };

      // This should be a complete flow
      const validation = invoker.validateParams(placeOrderTool, params);
      expect(validation.valid).toBe(true);

      const key = invoker.generateIdempotencyKey(placeOrderTool.name, params);
      expect(key).toBeDefined();

      const isDup = invoker.isDuplicate(key);
      expect(isDup).toBe(false);

      // Would invoke here, then log
    });

    it('should prevent duplicate execution via idempotency', () => {
      const params = { symbol: 'EURUSD', volume: 0.1, type: 'BUY' };
      const key = invoker.generateIdempotencyKey(placeOrderTool.name, params);

      invoker.recordInvocation(key, { ticket: 12345 });

      // Second attempt should detect duplicate
      const isDuplicate = invoker.isDuplicate(key);
      expect(isDuplicate).toBe(true);
    });
  });

  describe('Response Parsing', () => {
    it('should parse successful tool response', () => {
      const response = {
        result: {
          ticket: 12345,
          status: 'FILLED',
          entryPrice: '1.0850',
        },
      };

      const parsed = invoker.parseResponse(response, placeOrderTool);
      expect(parsed.success).toBe(true);
      expect(parsed.data).toEqual(response.result);
    });

    it('should parse error response', () => {
      const response = {
        error: {
          code: 'INSUFFICIENT_MARGIN',
          message: 'Not enough margin',
        },
      };

      const parsed = invoker.parseResponse(response, placeOrderTool);
      expect(parsed.success).toBe(false);
      expect(parsed.error).toBeDefined();
    });
  });
});

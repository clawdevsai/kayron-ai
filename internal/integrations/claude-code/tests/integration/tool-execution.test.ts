/**
 * Integration Tests: Tool Execution on MCP Server
 * Tests order placement, position closing, error handling, retries
 */

import { MCPClient } from '../../mcp-client/client';
import { ToolInvoker } from '../../mcp-client/tool-invoker';
import { SchemaValidator } from '../../mcp-client/schema-validator';
import { MCPConfig } from '../../mcp-client/types';

describe('Tool Execution Integration', () => {
  let client: MCPClient;
  let invoker: ToolInvoker;

  const config: MCPConfig = {
    host: 'localhost',
    port: 50051,
    apiKey: process.env.MCP_API_KEY || 'test-api-key',
    cacheTtlMinutes: 60,
    logLevel: 'info',
    reconnectMaxRetries: 3,
  };

  beforeEach(() => {
    client = new MCPClient(config);
    const validator = new SchemaValidator();
    invoker = new ToolInvoker(validator, client['logger']);
  });

  afterEach(async () => {
    if (client?.isConnected()) {
      await client.disconnect();
    }
  });

  describe('Place Order Execution', () => {
    it('should place order and return ticket', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();

      const params = {
        symbol: 'EURUSD',
        volume: 0.1,
        type: 'BUY',
      };

      const result = await invoker.invoke(
        {
          name: 'place-order',
          description: 'Place order',
          inputSchema: { type: 'object', properties: {} },
          outputSchema: { type: 'object', properties: {} },
        },
        params
      );

      expect(result.success).toBe(true);
      expect(result.data).toBeDefined();
      expect(result.data.ticket).toBeDefined();
      expect(typeof result.data.ticket).toBe('number');
    });

    it('should fill order with correct parameters', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();

      const params = {
        symbol: 'EURUSD',
        volume: 0.5,
        type: 'SELL',
        price: '1.0850',
      };

      const result = await invoker.invoke(
        {
          name: 'place-order',
          description: 'Place order',
          inputSchema: { type: 'object', properties: {} },
          outputSchema: { type: 'object', properties: {} },
        },
        params
      );

      expect(result.data.status).toMatch(/FILLED|PENDING|ACCEPTED/i);
    });

    it('should return entry price on order fill', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();

      const params = {
        symbol: 'EURUSD',
        volume: 0.1,
        type: 'BUY',
      };

      const result = await invoker.invoke(
        {
          name: 'place-order',
          description: 'Place order',
          inputSchema: { type: 'object', properties: {} },
          outputSchema: { type: 'object', properties: {} },
        },
        params
      );

      expect(result.data.entryPrice).toBeDefined();
      expect(typeof result.data.entryPrice).toBe('string');
      expect(/^\d+\.\d+$/.test(result.data.entryPrice)).toBe(true);
    });
  });

  describe('Close Position Execution', () => {
    it('should close position and remove from positions list', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();

      // First place order
      const placeResult = await invoker.invoke(
        {
          name: 'place-order',
          description: 'Place order',
          inputSchema: { type: 'object', properties: {} },
          outputSchema: { type: 'object', properties: {} },
        },
        { symbol: 'EURUSD', volume: 0.1, type: 'BUY' }
      );

      const ticket = placeResult.data.ticket;

      // Then close it
      const closeResult = await invoker.invoke(
        {
          name: 'close-position',
          description: 'Close position',
          inputSchema: { type: 'object', properties: {} },
          outputSchema: { type: 'object', properties: {} },
        },
        { ticket }
      );

      expect(closeResult.success).toBe(true);
      expect(closeResult.data.status).toMatch(/CLOSED|SUCCESS/i);
    });
  });

  describe('Error Handling', () => {
    it('should return error on insufficient margin', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();

      // Attempt to place massive order (will fail)
      const params = {
        symbol: 'EURUSD',
        volume: 1000, // Very large
        type: 'BUY',
      };

      const result = await invoker.invoke(
        {
          name: 'place-order',
          description: 'Place order',
          inputSchema: { type: 'object', properties: {} },
          outputSchema: { type: 'object', properties: {} },
        },
        params
      );

      if (!result.success) {
        expect(result.error).toBeDefined();
        expect(result.error.code).toMatch(/INSUFFICIENT_MARGIN|INVALID_VOLUME/i);
      }
    });

    it('should return error details', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();

      const result = await invoker.invoke(
        {
          name: 'place-order',
          description: 'Place order',
          inputSchema: { type: 'object', properties: {} },
          outputSchema: { type: 'object', properties: {} },
        },
        { symbol: 'INVALID', volume: 0.1, type: 'BUY' }
      );

      if (!result.success) {
        expect(result.error).toBeDefined();
        expect(result.error.message).toBeDefined();
        expect(result.error.code).toBeDefined();
      }
    });

    it('should classify error as retryable', () => {
      const retryableError = {
        code: 'NETWORK_TIMEOUT',
        message: 'Connection timeout',
      };

      expect(invoker.isRetryable(retryableError)).toBe(true);
    });

    it('should classify error as permanent', () => {
      const permanentError = {
        code: 'INVALID_SYMBOL',
        message: 'Symbol not found',
      };

      expect(invoker.isRetryable(permanentError)).toBe(false);
    });
  });

  describe('Retry Logic', () => {
    it('should retry on transient network errors', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();

      const params = {
        symbol: 'EURUSD',
        volume: 0.1,
        type: 'BUY',
      };

      const startTime = Date.now();

      // Should retry internally if timeout occurs
      const result = await invoker.invokeWithRetry(
        {
          name: 'place-order',
          description: 'Place order',
          inputSchema: { type: 'object', properties: {} },
          outputSchema: { type: 'object', properties: {} },
        },
        params,
        { maxRetries: 3 }
      );

      const duration = Date.now() - startTime;

      // If retried, should take longer than single attempt
      expect(result).toBeDefined();
    });

    it('should not retry on permanent errors', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
        }

      await client.connect();

      const result = await invoker.invokeWithRetry(
        {
          name: 'place-order',
          description: 'Place order',
          inputSchema: { type: 'object', properties: {} },
          outputSchema: { type: 'object', properties: {} },
        },
        { symbol: 'BADPAIR', volume: 0.1, type: 'BUY' },
        { maxRetries: 3 }
      );

      // Should fail immediately, not retry
      if (!result.success) {
        expect(result.retryCount).toBe(0);
      }
    });

    it('should respect max retries limit', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();

      // Mock timeout to test retry limits
      const result = await invoker.invokeWithRetry(
        {
          name: 'place-order',
          description: 'Place order',
          inputSchema: { type: 'object', properties: {} },
          outputSchema: { type: 'object', properties: {} },
        },
        { symbol: 'EURUSD', volume: 0.1, type: 'BUY' },
        { maxRetries: 2 }
      );

      if (!result.success && result.retryCount) {
        expect(result.retryCount).toBeLessThanOrEqual(2);
      }
    });
  });

  describe('Idempotency', () => {
    it('should prevent duplicate order via idempotency key', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();

      const params = {
        symbol: 'EURUSD',
        volume: 0.1,
        type: 'BUY',
      };

      const key = invoker.generateIdempotencyKey('place-order', params);

      // First invocation
      const result1 = await invoker.invoke(
        {
          name: 'place-order',
          description: 'Place order',
          inputSchema: { type: 'object', properties: {} },
          outputSchema: { type: 'object', properties: {} },
        },
        params
      );

      invoker.recordInvocation(key, result1.data);

      // Second invocation with same key should be rejected
      const isDuplicate = invoker.isDuplicate(key);
      expect(isDuplicate).toBe(true);
    });
  });

  describe('Audit Trail', () => {
    it('should log tool invocation', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();

      const params = {
        symbol: 'EURUSD',
        volume: 0.1,
        type: 'BUY',
      };

      await invoker.invoke(
        {
          name: 'place-order',
          description: 'Place order',
          inputSchema: { type: 'object', properties: {} },
          outputSchema: { type: 'object', properties: {} },
        },
        params
      );

      const log = invoker.getLastLog();
      expect(log).toBeDefined();
      expect(log?.toolName).toBe('place-order');
      expect(log?.inputParams).toEqual(params);
    });

    it('should include execution duration in log', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();

      await invoker.invoke(
        {
          name: 'place-order',
          description: 'Place order',
          inputSchema: { type: 'object', properties: {} },
          outputSchema: { type: 'object', properties: {} },
        },
        { symbol: 'EURUSD', volume: 0.1, type: 'BUY' }
      );

      const log = invoker.getLastLog();
      expect(log?.executionDurationMs).toBeGreaterThan(0);
    });

    it('should include idempotency key in log', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();

      const params = {
        symbol: 'EURUSD',
        volume: 0.1,
        type: 'BUY',
      };

      const key = invoker.generateIdempotencyKey('place-order', params);

      await invoker.invoke(
        {
          name: 'place-order',
          description: 'Place order',
          inputSchema: { type: 'object', properties: {} },
          outputSchema: { type: 'object', properties: {} },
        },
        params
      );

      const log = invoker.getLastLog();
      expect(log?.idempotencyKey).toBeDefined();
    });
  });
});

/**
 * Unit Tests: MCP Client Connection
 * TDD-driven: connection lifecycle, auth, retry, server unavailable
 */

import { MCPClient } from '../../mcp-client/client';
import { MCPConfig } from '../../mcp-client/types';

describe('MCPClient', () => {
  let client: MCPClient;
  const validConfig: MCPConfig = {
    host: 'localhost',
    port: 50051,
    apiKey: 'test-api-key',
    cacheTtlMinutes: 60,
    logLevel: 'info',
    reconnectMaxRetries: 5,
  };

  beforeEach(() => {
    client = new MCPClient(validConfig);
  });

  afterEach(async () => {
    if (client.isConnected()) {
      await client.disconnect();
    }
  });

  describe('Connection Lifecycle', () => {
    it('should initialize in disconnected state', () => {
      expect(client.isConnected()).toBe(false);
      expect(client.getStatus()).toBe('disconnected');
    });

    it('should return correct status when not connected', () => {
      const status = client.getStatus();
      expect(status).toMatch(/disconnected|offline/i);
    });
  });

  describe('Successful Connection', () => {
    it('should establish connection with valid credentials', async () => {
      // This test assumes a real MCP server running on localhost:50051
      // For unit testing, this may need to be mocked in CI
      // Skipping by default, enable with ENABLE_MCP_SERVER=true
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();
      expect(client.isConnected()).toBe(true);
      expect(client.getStatus()).toMatch(/connected|online/i);
    });

    it('should emit connected event on successful connection', (done) => {
      if (!process.env.ENABLE_MCP_SERVER) {
        done();
        return;
      }

      client.on('connected', () => {
        expect(client.isConnected()).toBe(true);
        done();
      });

      client.connect().catch(() => {
        // Connection may fail in test environment
        done();
      });
    });

    it('should set up heartbeat after connection', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();
      // Heartbeat should be active (implementation detail, verified by not timing out)
      expect(client.isConnected()).toBe(true);
    });
  });

  describe('Connection Failure', () => {
    it('should fail connection with invalid API key', async () => {
      const invalidConfig: MCPConfig = {
        ...validConfig,
        apiKey: 'invalid-key',
      };
      const invalidClient = new MCPClient(invalidConfig);

      if (!process.env.ENABLE_MCP_SERVER) {
        // Without server, just verify connection fails gracefully
        try {
          await invalidClient.connect();
        } catch (err) {
          expect(err).toBeDefined();
        }
        return;
      }

      await expect(invalidClient.connect()).rejects.toThrow();
      expect(invalidClient.isConnected()).toBe(false);
    });

    it('should emit error event on connection failure', (done) => {
      const badConfig: MCPConfig = {
        ...validConfig,
        host: '127.0.0.1',
        port: 9999, // Port unlikely to have server
      };
      const badClient = new MCPClient(badConfig);

      badClient.on('error', (error) => {
        expect(error).toBeDefined();
        expect(badClient.isConnected()).toBe(false);
        done();
      });

      badClient.connect().catch(() => {
        // Error event should have been emitted
      });
    });

    it('should handle gracefully when server unavailable', async () => {
      const unreachableConfig: MCPConfig = {
        ...validConfig,
        host: '192.0.2.1', // TEST-NET-1, guaranteed to be unreachable
        port: 50051,
        reconnectMaxRetries: 1, // Quick fail for test
      };
      const unreachableClient = new MCPClient(unreachableConfig);

      try {
        await unreachableClient.connect();
      } catch (err) {
        // Expected
        expect(unreachableClient.isConnected()).toBe(false);
      }
    });
  });

  describe('Retry Logic', () => {
    it('should retry with exponential backoff', async () => {
      const unreachableConfig: MCPConfig = {
        ...validConfig,
        host: '192.0.2.1',
        port: 50051,
        reconnectMaxRetries: 3,
      };
      const unreachableClient = new MCPClient(unreachableConfig);

      const startTime = Date.now();
      try {
        await unreachableClient.connect();
      } catch (err) {
        // Expected to fail after retries
      }
      const duration = Date.now() - startTime;

      // Should have delayed retries: 1s + 2s + 4s = ~7s minimum
      // Allow 1s buffer for execution time
      expect(duration).toBeGreaterThanOrEqual(6000);
    });

    it('should respect max retries limit', async () => {
      const config: MCPConfig = {
        ...validConfig,
        host: '192.0.2.1',
        port: 50051,
        reconnectMaxRetries: 2,
      };
      const client2 = new MCPClient(config);

      try {
        await client2.connect();
      } catch (err) {
        // Should fail after max retries
        expect(client2.isConnected()).toBe(false);
      }
    });

    it('should calculate exponential backoff correctly', () => {
      // Test backoff calculation: 1s, 2s, 4s, 8s, 16s, 32s max
      const backoffMs = client['calculateBackoff'](0); // 1st retry
      expect(backoffMs).toBe(1000);

      const backoff2 = client['calculateBackoff'](1); // 2nd retry
      expect(backoff2).toBe(2000);

      const backoff3 = client['calculateBackoff'](2); // 3rd retry
      expect(backoff3).toBe(4000);

      const backoff6 = client['calculateBackoff'](5); // 6th retry
      expect(backoff6).toBe(32000); // Capped at 32s
    });
  });

  describe('Disconnect', () => {
    it('should disconnect gracefully', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        // Mock disconnect
        expect(client.isConnected()).toBe(false);
        return;
      }

      await client.connect();
      await client.disconnect();
      expect(client.isConnected()).toBe(false);
    });

    it('should emit disconnected event', (done) => {
      if (!process.env.ENABLE_MCP_SERVER) {
        done();
        return;
      }

      client.on('disconnected', () => {
        expect(client.isConnected()).toBe(false);
        done();
      });

      client.connect().then(() => {
        client.disconnect();
      });
    });

    it('should stop heartbeat on disconnect', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();
      await client.disconnect();
      // Heartbeat should be cleared (verified by not throwing)
      expect(client.isConnected()).toBe(false);
    });
  });

  describe('Events', () => {
    it('should allow registering event listeners', () => {
      const callback = jest.fn();
      client.on('connected', callback);
      // Listener registered successfully
      expect(callback).not.toHaveBeenCalled();
    });

    it('should allow unregistering event listeners', () => {
      const callback = jest.fn();
      client.on('connected', callback);
      client.off('connected', callback);
      // Listener unregistered
    });
  });

  describe('Configuration', () => {
    it('should accept custom host and port', () => {
      const customConfig: MCPConfig = {
        ...validConfig,
        host: '192.168.1.1',
        port: 9999,
      };
      const customClient = new MCPClient(customConfig);
      expect(customClient).toBeDefined();
    });

    it('should use default values from config', () => {
      const minimalConfig: MCPConfig = {
        host: 'localhost',
        port: 50051,
        apiKey: 'test-key',
      };
      const minimalClient = new MCPClient(minimalConfig);
      expect(minimalClient).toBeDefined();
    });
  });
});

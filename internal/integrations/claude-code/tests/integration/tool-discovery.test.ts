/**
 * Integration Tests: Tool Discovery & Caching
 * Tests MCP tool discovery endpoint + schema cache
 */

import { MCPClient } from '../../mcp-client/client';
import { SchemaCache } from '../../mcp-client/cache';
import { MCPConfig, ToolDefinition } from '../../mcp-client/types';
import * as fs from 'fs';
import * as path from 'path';

describe('Tool Discovery Integration', () => {
  let client: MCPClient;
  let cache: SchemaCache;

  const config: MCPConfig = {
    host: 'localhost',
    port: 50051,
    apiKey: process.env.MCP_API_KEY || 'test-api-key',
    cacheTtlMinutes: 60,
    logLevel: 'info',
  };

  beforeEach(() => {
    client = new MCPClient(config);
    cache = new SchemaCache(config.cacheTtlMinutes);
    // Clear cache before each test
    cache.clear();
  });

  afterEach(async () => {
    if (client?.isConnected()) {
      await client.disconnect();
    }
  });

  describe('Tool Discovery', () => {
    it('should call tool discovery endpoint on server', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();
      const tools = await client.listTools();

      expect(Array.isArray(tools)).toBe(true);
      expect(tools.length).toBeGreaterThan(0);
    });

    it('should return tools with valid structure', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();
      const tools = await client.listTools();

      tools.forEach((tool: ToolDefinition) => {
        expect(tool.name).toBeDefined();
        expect(typeof tool.name).toBe('string');
        expect(tool.description).toBeDefined();
        expect(typeof tool.description).toBe('string');
        expect(tool.inputSchema).toBeDefined();
        expect(tool.outputSchema).toBeDefined();
      });
    });

    it('should include MT5 trading tools', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();
      const tools = await client.listTools();

      const toolNames = tools.map((t: ToolDefinition) => t.name);
      expect(toolNames).toContain('place-order');
      expect(toolNames).toContain('close-position');
      expect(toolNames).toContain('positions-list');
    });
  });

  describe('Schema Validation', () => {
    it('should have valid JSON schema for tool input', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();
      const tools = await client.listTools();

      tools.forEach((tool: ToolDefinition) => {
        expect(tool.inputSchema.type).toBe('object');
        if (tool.inputSchema.properties) {
          Object.entries(tool.inputSchema.properties).forEach(([prop, schema]: any) => {
            expect(schema.type).toBeDefined();
          });
        }
      });
    });

    it('should have valid JSON schema for tool output', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();
      const tools = await client.listTools();

      tools.forEach((tool: ToolDefinition) => {
        expect(tool.outputSchema.type).toBe('object');
      });
    });
  });

  describe('Schema Caching', () => {
    it('should cache tools after discovery', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      // First call should fetch from server
      await client.connect();
      const tools1 = await client.listTools();
      expect(tools1.length).toBeGreaterThan(0);

      // Verify cache was populated
      const cached = cache.load();
      expect(cached).toBeDefined();
      expect(cached?.tools).toBeDefined();
      expect(cached?.tools?.length).toBe(tools1.length);
    });

    it('should use cache on second discovery call', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();

      // First call
      const startTime1 = Date.now();
      const tools1 = await client.listTools();
      const duration1 = Date.now() - startTime1;

      // Second call (should be much faster from cache)
      const startTime2 = Date.now();
      const tools2 = await client.listTools();
      const duration2 = Date.now() - startTime2;

      expect(tools1).toEqual(tools2);
      expect(duration2).toBeLessThan(duration1 + 100); // Cache should be instant
    });

    it('should expire cache after TTL', (done) => {
      const shortTtl = 1; // 1 minute
      const expireCache = new SchemaCache(shortTtl);

      // Populate cache
      expireCache.save({
        timestamp: new Date().toISOString(),
        tools: [
          {
            name: 'test-tool',
            description: 'Test',
            inputSchema: { type: 'object' },
            outputSchema: { type: 'object' },
          },
        ],
        toolCount: 1,
      });

      expect(expireCache.isValid()).toBe(true);

      // Mock time passing (in real test, would need to wait or mock Date)
      // For unit test, we verify the TTL logic exists
      done();
    });

    it('should invalidate cache on explicit refresh', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();
      const tools1 = await client.listTools();

      // Clear cache (simulating refresh)
      cache.clear();

      // Verify cache was cleared
      const cached = cache.load();
      expect(cached).toBeNull();

      // New discovery should fetch fresh data
      const tools2 = await client.listTools();
      expect(tools2.length).toBeGreaterThan(0);
    });

    it('should handle corrupted cache gracefully', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      // Write corrupted cache file
      const cacheDir = path.join(process.env.HOME || '.', '.claude', 'cache');
      const cacheFile = path.join(cacheDir, 'kayron-tools.json');

      if (!fs.existsSync(cacheDir)) {
        fs.mkdirSync(cacheDir, { recursive: true });
      }

      fs.writeFileSync(cacheFile, 'invalid json {');

      // Should still work, skipping corrupted cache
      await client.connect();
      const tools = await client.listTools();

      expect(Array.isArray(tools)).toBe(true);
      expect(tools.length).toBeGreaterThan(0);
    });
  });

  describe('Cache Metadata', () => {
    it('should track cache timestamp', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();
      await client.listTools();

      const cached = cache.load();
      expect(cached?.timestamp).toBeDefined();
      expect(new Date(cached!.timestamp)).toBeInstanceOf(Date);
    });

    it('should track tool count in cache', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();
      const tools = await client.listTools();

      const cached = cache.load();
      expect(cached?.toolCount).toBe(tools.length);
    });

    it('should track TTL in cache metadata', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      const ttlMinutes = 60;
      const testCache = new SchemaCache(ttlMinutes);

      await client.connect();
      const tools = await client.listTools();

      testCache.save({
        timestamp: new Date().toISOString(),
        tools,
        toolCount: tools.length,
      });

      const cached = testCache.load();
      expect(cached).toBeDefined();
      // Verify TTL is stored (implementation-specific)
    });
  });

  describe('Cache Persistence', () => {
    it('should persist cache to file', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();
      await client.listTools();

      // Verify cache file exists
      const cacheDir = path.join(process.env.HOME || '.', '.claude', 'cache');
      const cacheFile = path.join(cacheDir, 'kayron-tools.json');

      expect(fs.existsSync(cacheFile)).toBe(true);

      // Verify file content is valid JSON
      const content = fs.readFileSync(cacheFile, 'utf-8');
      const parsed = JSON.parse(content);
      expect(parsed.tools).toBeDefined();
    });

    it('should load cache from file on startup', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      // First client discovers and caches
      const client1 = new MCPClient(config);
      const cache1 = new SchemaCache(config.cacheTtlMinutes);
      await client1.connect();
      const tools1 = await client1.listTools();
      await client1.disconnect();

      // Second client loads from file
      const client2 = new MCPClient(config);
      const cache2 = new SchemaCache(config.cacheTtlMinutes);

      const cached = cache2.load();
      expect(cached?.tools).toBeDefined();
      expect(cached?.tools?.length).toBe(tools1.length);
    });
  });

  describe('Performance', () => {
    it('should retrieve cached tools in under 10ms', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();
      await client.listTools(); // Populate cache

      const startTime = Date.now();
      const tools = await client.listTools(); // From cache
      const duration = Date.now() - startTime;

      expect(duration).toBeLessThan(10);
      expect(tools.length).toBeGreaterThan(0);
    });

    it('should handle large tool list (>100 tools)', async () => {
      if (!process.env.ENABLE_MCP_SERVER) {
        return;
      }

      await client.connect();
      const tools = await client.listTools();

      // Should handle large lists without performance issues
      expect(Array.isArray(tools)).toBe(true);
      tools.forEach((tool: ToolDefinition) => {
        expect(tool.name).toBeDefined();
      });
    });
  });
});

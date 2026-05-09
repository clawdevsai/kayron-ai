/**
 * IDE Event Handlers for Kayron MCP
 * Handles command palette, skill execution, UI updates
 */

import { MCPClientInterface, MCPEvent } from '../mcp-client/types';

export interface IDEEventHandler {
  onConnected(): void;
  onDisconnected(reason?: string): void;
  onError(error: Error): void;
  onToolResponse(result: unknown): void;
  onQueueUpdate(pendingCount: number): void;
}

export class CommandPaletteHandler {
  /**
   * Handle "Kayron: Status" command
   */
  static async handleStatus(client: MCPClientInterface): Promise<string> {
    const isConnected = client.isConnected();
    const status = isConnected ? 'Connected ✓' : 'Offline ⚠️';

    return `MCP Status: ${status}`;
  }

  /**
   * Handle "Kayron: Tools List" command
   */
  static async handleToolsList(client: MCPClientInterface): Promise<string> {
    try {
      const tools = await client.listTools();
      if (tools.length === 0) {
        return 'No tools available';
      }

      const table = tools
        .map((tool) => `${tool.name} - ${tool.description}`)
        .join('\n');

      return `Available Tools (${tools.length}):\n${table}`;
    } catch (err) {
      return `Error fetching tools: ${(err as Error).message}`;
    }
  }

  /**
   * Handle "Kayron: Place Order" command
   */
  static async handlePlaceOrder(client: MCPClientInterface, params: {
    symbol: string;
    volume: number;
    type: 'BUY' | 'SELL';
    price: string;
  }): Promise<string> {
    try {
      const result = await client.invokeTool('place-order', params);

      if (result.status === 'success') {
        return JSON.stringify(result.output, null, 2);
      } else {
        return `Error: ${result.error?.message}`;
      }
    } catch (err) {
      return `Failed to place order: ${(err as Error).message}`;
    }
  }

  /**
   * Handle "Kayron: Refresh Schema" command
   */
  static async handleRefreshSchema(client: MCPClientInterface): Promise<string> {
    try {
      const tools = await client.listTools();
      await client.cacheTools(tools);
      return `Schema refreshed: ${tools.length} tools cached`;
    } catch (err) {
      return `Failed to refresh schema: ${(err as Error).message}`;
    }
  }
}

export class SkillExecutionHandler {
  /**
   * Execute a skill (calls multiple MCP tools in sequence)
   */
  static async executeSkill(client: MCPClientInterface, skillSteps: Array<{
    toolName: string;
    params: Record<string, unknown>;
  }>): Promise<Array<{
    step: number;
    toolName: string;
    status: 'success' | 'error';
    result?: unknown;
    error?: string;
  }>> {
    const results = [];

    for (let i = 0; i < skillSteps.length; i++) {
      const step = skillSteps[i];

      try {
        const result = await client.invokeTool(step.toolName, step.params);

        results.push({
          step: i + 1,
          toolName: step.toolName,
          status: result.status,
          result: result.output,
          error: result.error?.message,
        });

        // Stop execution on error (fail-fast)
        if (result.status === 'error') {
          break;
        }
      } catch (err) {
        results.push({
          step: i + 1,
          toolName: step.toolName,
          status: 'error',
          error: (err as Error).message,
        });
        break;
      }
    }

    return results;
  }
}

export class EventListener {
  private handlers: Map<string, Function[]> = new Map();

  /**
   * Register event handler
   */
  on(event: string, handler: (data: unknown) => void): void {
    if (!this.handlers.has(event)) {
      this.handlers.set(event, []);
    }
    this.handlers.get(event)!.push(handler);
  }

  /**
   * Unregister event handler
   */
  off(event: string, handler: (data: unknown) => void): void {
    const handlers = this.handlers.get(event);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index !== -1) {
        handlers.splice(index, 1);
      }
    }
  }

  /**
   * Emit event to all listeners
   */
  emit(event: string, data: unknown): void {
    const handlers = this.handlers.get(event);
    if (handlers) {
      handlers.forEach((handler) => {
        try {
          handler(data);
        } catch (err) {
          console.error(`Error in event handler for ${event}:`, err);
        }
      });
    }
  }

  /**
   * Get listener count for event
   */
  listenerCount(event: string): number {
    return this.handlers.get(event)?.length || 0;
  }

  /**
   * Clear all listeners
   */
  clear(): void {
    this.handlers.clear();
  }
}

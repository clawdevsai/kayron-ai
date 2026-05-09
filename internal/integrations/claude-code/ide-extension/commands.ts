/**
 * Command Handlers for Kayron MCP IDE Commands
 * /kayron tools-list, /kayron place-order, /kayron refresh-schema, etc.
 */

import { MCPClient } from '../mcp-client/client';
import { ToolDefinition } from '../mcp-client/types';

export class CommandHandler {
  private client: MCPClient;
  private lastOrderParams: any = null;

  constructor(client: MCPClient) {
    this.client = client;
  }

  /**
   * Handle /kayron tools-list command
   * Lists all available tools in table format
   */
  async handleToolsList(): Promise<string> {
    if (!this.client.isConnected()) {
      return 'Error: Not connected to MCP server. Check status.';
    }

    try {
      const tools = await this.client.listTools();

      if (tools.length === 0) {
        return 'No tools available from MCP server.';
      }

      // Format as table: Tool Name | Description | Input Schema | Output Schema
      let table = '| Tool Name | Description | Input Fields | Output Fields |\n';
      table += '|-----------|-------------|--------------|---------------|\n';

      for (const tool of tools) {
        const inputFields = Object.keys(tool.inputSchema?.properties || {}).join(', ') || 'N/A';
        const outputFields = Object.keys(tool.outputSchema?.properties || {}).join(', ') || 'N/A';
        const desc = tool.description.substring(0, 50) + (tool.description.length > 50 ? '...' : '');

        table += `| ${tool.name} | ${desc} | ${inputFields} | ${outputFields} |\n`;
      }

      return (
        `## Available Tools (${tools.length})\n\n` +
        table +
        `\nClick tool name to view detailed schema.`
      );
    } catch (err) {
      return `Error listing tools: ${(err as Error).message}`;
    }
  }

  /**
   * Handle /kayron refresh-schema command
   * Force re-query of tool schemas from server
   */
  async handleRefreshSchema(): Promise<string> {
    if (!this.client.isConnected()) {
      return 'Error: Not connected to MCP server.';
    }

    try {
      // Clear cache first
      // TODO: Expose cache clear on client

      // Fetch fresh tools
      const tools = await this.client.listTools();

      return `✅ Schema updated: ${tools.length} tools cached`;
    } catch (err) {
      return `Error refreshing schema: ${(err as Error).message}`;
    }
  }

  /**
   * Handle /kayron place-order command
   * Show dialog to collect order parameters
   */
  async handlePlaceOrder(params?: {
    symbol?: string;
    volume?: number;
    type?: string;
    price?: string;
  }): Promise<string> {
    if (!this.client.isConnected()) {
      return 'Error: Not connected to MCP server. Check status.';
    }

    try {
      // Collect parameters (via IDE dialog or passed in)
      const orderParams = {
        symbol: params?.symbol || 'EURUSD',
        volume: params?.volume || 0.1,
        type: params?.type || 'BUY',
        price: params?.price || 'MARKET',
      };

      // TODO: Show IDE quick-input dialog with fields:
      // 1. Symbol (default: EURUSD)
      // 2. Volume (default: 0.1)
      // 3. Type (dropdown: BUY/SELL)
      // 4. Price (default: MARKET, or specific price)

      // For now, return formatted dialog response
      return (
        `## Place Order\n\n` +
        `**Symbol**: ${orderParams.symbol}\n` +
        `**Volume**: ${orderParams.volume}\n` +
        `**Type**: ${orderParams.type}\n` +
        `**Price**: ${orderParams.price}\n\n` +
        `[Execute] [Cancel]`
      );
    } catch (err) {
      return `Error preparing order: ${(err as Error).message}`;
    }
  }

  /**
   * Execute placed order
   */
  async executeOrder(params: { symbol: string; volume: number; type: string; price?: string }): Promise<string> {
    if (!this.client.isConnected()) {
      return 'Error: Not connected to MCP server.';
    }

    try {
      const result = await this.client.invokeTool('place-order', params);

      this.lastOrderParams = params;

      return (
        `✅ Order Executed\n\n` +
        `**Ticket**: ${result.ticket}\n` +
        `**Status**: ${result.status}\n` +
        `**Entry Price**: ${result.entryPrice}\n` +
        `**Symbol**: ${params.symbol}\n` +
        `**Volume**: ${params.volume}\n` +
        `**Type**: ${params.type}`
      );
    } catch (err) {
      const errorMsg = (err as Error).message;

      if (errorMsg.includes('INSUFFICIENT_MARGIN')) {
        return (
          `❌ Order Failed: Insufficient Margin\n\n` +
          `Your account balance is too low for this trade.\n` +
          `[Add Funds] [Adjust Size]`
        );
      }

      if (errorMsg.includes('INVALID_SYMBOL')) {
        return (
          `❌ Order Failed: Invalid Symbol\n\n` +
          `"${params.symbol}" is not available in your market watch.\n` +
          `[View Available Symbols]`
        );
      }

      if (errorMsg.includes('MARKET_CLOSED')) {
        return `⏰ Order Failed: Market Closed\n\nTrade during market hours only.`;
      }

      return `❌ Order Failed: ${errorMsg}`;
    }
  }

  /**
   * Format tool for display in browser
   */
  formatToolForDisplay(tool: ToolDefinition): string {
    const inputProps = Object.entries(tool.inputSchema?.properties || {});
    const outputProps = Object.entries(tool.outputSchema?.properties || {});

    let formatted = `## ${tool.name}\n\n`;
    formatted += `${tool.description}\n\n`;

    formatted += `### Input Parameters\n`;
    if (inputProps.length > 0) {
      formatted += '| Parameter | Type | Required | Description |\n';
      formatted += '|-----------|------|----------|-------------|\n';

      const required = tool.inputSchema?.required || [];
      for (const [name, prop] of inputProps) {
        const isRequired = required.includes(name) ? '✓' : '✗';
        const type = (prop as any).type || 'unknown';
        const desc = (prop as any).description || '';

        formatted += `| ${name} | ${type} | ${isRequired} | ${desc} |\n`;
      }
    } else {
      formatted += 'None\n';
    }

    formatted += `\n### Output\n`;
    if (outputProps.length > 0) {
      formatted += '| Field | Type | Description |\n';
      formatted += '|-------|------|-------------|\n';

      for (const [name, prop] of outputProps) {
        const type = (prop as any).type || 'unknown';
        const desc = (prop as any).description || '';

        formatted += `| ${name} | ${type} | ${desc} |\n`;
      }
    } else {
      formatted += 'None\n';
    }

    return formatted;
  }
}

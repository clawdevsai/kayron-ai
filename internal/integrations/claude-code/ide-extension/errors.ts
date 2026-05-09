/**
 * IDE Error Messages & Notifications
 * Shows user-friendly error messages with actionable guidance
 */

import { MCPClient } from '../mcp-client/client';
import { Logger } from '../mcp-client/logger';

export class ErrorNotifications {
  private client: MCPClient;
  private logger: Logger;
  private notificationQueue: any[] = [];

  constructor(client: MCPClient) {
    this.client = client;
    this.logger = new Logger('info');
    this.setupErrorListeners();
  }

  /**
   * Setup listeners for errors and connection failures
   */
  private setupErrorListeners(): void {
    this.client.on('error', (data: any) => {
      this.handleConnectionError(data);
    });

    this.client.on('disconnected', () => {
      // Disconnection is normal, only show error if unexpected
    });
  }

  /**
   * Handle connection error with user-friendly message
   */
  private handleConnectionError(error: any): void {
    const message = this.getErrorMessage(error);
    this.logger.error('Connection error', { error: error.error || error.message });
    this.showNotification('error', message);
  }

  /**
   * Generate user-friendly error message
   */
  private getErrorMessage(error: any): string {
    const errorStr = error.error || error.message || '';

    if (errorStr.includes('ECONNREFUSED') || errorStr.includes('Connection refused')) {
      return (
        '❌ MCP server not found at localhost:50051\n\n' +
        'Check settings.json:\n' +
        '  • mcp.kayron.host: localhost\n' +
        '  • mcp.kayron.port: 50051\n\n' +
        '[View Setup Guide] [Troubleshoot]'
      );
    }

    if (errorStr.includes('ETIMEDOUT') || errorStr.includes('timed out')) {
      return (
        '⏱️ Connection timeout: MCP server not responding\n\n' +
        'Verify:\n' +
        '  • Server is running (MT5 + Kayron gRPC daemon)\n' +
        '  • Network connectivity to localhost:50051\n' +
        '  • Firewall not blocking port 50051\n\n' +
        '[Retry] [Troubleshoot]'
      );
    }

    if (errorStr.includes('ENOTFOUND') || errorStr.includes('not found')) {
      return (
        '🔍 Cannot resolve hostname\n\n' +
        'Check:\n' +
        '  • Hostname/IP in settings.json is correct\n' +
        '  • DNS resolution working\n' +
        '  • For custom hosts, ensure they are reachable\n\n' +
        '[Edit Settings] [Troubleshoot]'
      );
    }

    if (errorStr.includes('EACCES') || errorStr.includes('permission denied')) {
      return (
        '🔒 Permission denied connecting to MCP server\n\n' +
        'Check:\n' +
        '  • API key in settings.json is correct\n' +
        '  • User has permission to access this host:port\n' +
        '  • Firewall rules allow connection\n\n' +
        '[Check API Key] [Troubleshoot]'
      );
    }

    // Generic error
    return (
      `❌ Failed to connect to MCP server: ${errorStr}\n\n` +
      '[View Logs] [Troubleshoot] [Retry]'
    );
  }

  /**
   * Show notification to user
   */
  private showNotification(level: 'error' | 'warning' | 'info', message: string): void {
    this.notificationQueue.push({
      level,
      message,
      timestamp: new Date().toISOString(),
    });

    // TODO: Show in IDE notification system (VS Code showErrorMessage, etc.)
    console.log(`[${level.toUpperCase()}] ${message}`);

    // Log to file for audit trail
    if (level === 'error') {
      this.logger.error(message);
    } else if (level === 'warning') {
      this.logger.warn(message);
    } else {
      this.logger.info(message);
    }
  }

  /**
   * Show error with action buttons
   */
  showErrorWithActions(
    message: string,
    actions: Array<{ label: string; callback: () => void }>
  ): void {
    this.logger.error(message);
    console.log(`Error: ${message}`);
    console.log('Actions:', actions.map((a) => a.label));

    // TODO: Show in IDE with action buttons
    // VS Code: window.showErrorMessage(message, ...actions.map(a => a.label))
  }

  /**
   * Show error on tool invocation failure
   */
  handleToolError(toolName: string, error: any): void {
    let userMessage = `Failed to execute ${toolName}: ${error.message}`;

    // Translate MCP errors to user-friendly messages
    if (error.code === 'INSUFFICIENT_MARGIN') {
      userMessage =
        `❌ Insufficient margin for ${toolName}\n\n` +
        'Your account balance is too low for this operation.\n' +
        'Add funds or reduce order size.\n\n' +
        '[View Account] [Add Funds]';
    } else if (error.code === 'INVALID_SYMBOL') {
      userMessage =
        `❌ Invalid symbol for ${toolName}\n\n` +
        'Verify the symbol is available in your market watch.\n' +
        'Check symbol spelling and broker availability.\n\n' +
        '[View Available Symbols]';
    } else if (error.code === 'MARKET_CLOSED') {
      userMessage = `⏰ Market closed\n\nOrders can only be placed during market hours.`;
    } else if (error.code === 'INVALID_VOLUME') {
      userMessage =
        `❌ Invalid volume for ${toolName}\n\n` +
        'Check position limits and minimum/maximum lot sizes.\n' +
        'Adjust volume and try again.\n\n' +
        '[View Limits]';
    }

    this.showErrorWithActions(userMessage, [
      { label: 'Retry', callback: () => {} },
      { label: 'View Logs', callback: () => {} },
    ]);
  }

  /**
   * Get notification history
   */
  getNotificationHistory(): any[] {
    return [...this.notificationQueue].slice(-10); // Last 10
  }

  /**
   * Clear notification queue
   */
  clearHistory(): void {
    this.notificationQueue = [];
  }

  /**
   * Dispose resources
   */
  dispose(): void {
    this.clearHistory();
  }
}

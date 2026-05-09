/**
 * IDE Status Bar Badge
 * Shows connection status, click to open details panel
 */

import { MCPClient } from '../mcp-client/client';
import { MCPConnectionStatus } from '../mcp-client/types';

export class StatusBarBadge {
  private client: MCPClient;
  private badge: any; // IDE status bar item
  private updateInterval: NodeJS.Timeout | null = null;
  private pendingCount: number = 0;

  constructor(client: MCPClient) {
    this.client = client;
    this.setupEventListeners();
  }

  /**
   * Initialize status bar (called by IDE extension)
   */
  init(statusBar: any): void {
    this.badge = statusBar.createStatusBarItem('kayron-status', 1);
    this.updateBadge();
    this.startPolling();
  }

  /**
   * Setup event listeners for connection state changes
   */
  private setupEventListeners(): void {
    this.client.on('connected', () => {
      this.updateBadge();
    });

    this.client.on('disconnected', () => {
      this.updateBadge();
    });

    this.client.on('error', (error) => {
      this.updateBadge();
    });

    this.client.on('queue-update', (data: any) => {
      this.pendingCount = data.pendingCount || 0;
      this.updateBadge();
    });
  }

  /**
   * Update badge appearance based on connection state
   */
  private updateBadge(): void {
    const isConnected = this.client.isConnected();
    const status = this.client.getStatus();

    if (isConnected) {
      this.badge.text = '$(check) Connected ✓';
      this.badge.color = '#4ec9b0'; // Green
      this.badge.tooltip = 'Kayron MCP: Connected';
    } else if (this.pendingCount > 0) {
      this.badge.text = `$(warning) ${this.pendingCount} pending`;
      this.badge.color = '#f4d03f'; // Yellow
      this.badge.tooltip = `Kayron MCP: ${this.pendingCount} operations queued`;
    } else {
      this.badge.text = '$(error) Offline ⚠️';
      this.badge.color = '#d4534f'; // Red
      this.badge.tooltip = 'Kayron MCP: Offline - Click to reconnect';
    }

    this.badge.command = 'kayron.showStatus';
  }

  /**
   * Start polling for status updates
   */
  private startPolling(): void {
    this.updateInterval = setInterval(() => {
      this.updateBadge();
    }, 1000); // Update every 1 second for real-time feedback
  }

  /**
   * Stop polling
   */
  stopPolling(): void {
    if (this.updateInterval) {
      clearInterval(this.updateInterval);
      this.updateInterval = null;
    }
  }

  /**
   * Show details panel (opened by badge click)
   */
  showDetails(): void {
    const status = this.client.getStatus();
    const isConnected = this.client.isConnected();

    const details = {
      status: status,
      connected: isConnected,
      host: 'localhost', // TODO: get from config
      port: 50051,
      pendingOperations: this.pendingCount,
      timestamp: new Date().toISOString(),
    };

    // TODO: Show panel in IDE (VS Code WebviewPanel, etc.)
    console.log('Status Details:', details);
  }

  /**
   * Update pending operation count
   */
  setPendingCount(count: number): void {
    this.pendingCount = count;
    this.updateBadge();
  }

  /**
   * Dispose resources
   */
  dispose(): void {
    this.stopPolling();
    if (this.badge) {
      this.badge.dispose();
    }
  }
}

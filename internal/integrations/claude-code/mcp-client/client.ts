/**
 * MCP Client: Socket connection + JSON-RPC communication
 * Handles connection lifecycle, authentication, retry, heartbeat
 */

import { createConnection, Socket } from 'net';
import { MCPConfig, MCPEvent, JSONRPCRequest, JSONRPCResponse, MCPConnectionStatus, ToolDefinition } from './types';
import { Logger } from './logger';
import { EventListener } from '../ide-extension/handlers';
import { SchemaCache } from './cache';

export class MCPClient {
  private socket: Socket | null = null;
  private config: MCPConfig;
  private logger: Logger;
  private eventListener: EventListener;
  private cache: SchemaCache;
  private status: MCPConnectionStatus = 'disconnected';
  private heartbeatTimer: NodeJS.Timeout | null = null;
  private reconnectAttempts: number = 0;
  private requestId: number = 1;
  private requestMap: Map<number, (response: JSONRPCResponse) => void> = new Map();
  private buffer: string = '';

  constructor(config: MCPConfig) {
    this.config = {
      cacheTtlMinutes: 60,
      logLevel: 'info',
      reconnectMaxRetries: 5,
      ...config,
    };
    this.logger = new Logger(this.config.logLevel);
    this.eventListener = new EventListener();
    this.cache = new SchemaCache(this.config.cacheTtlMinutes);
  }

  /**
   * Establish connection to MCP server
   */
  async connect(): Promise<void> {
    if (this.socket) {
      throw new Error('Already connected');
    }

    return new Promise((resolve, reject) => {
      const attemptConnect = () => {
        this.logger.debug('Attempting connection', {
          host: this.config.host,
          port: this.config.port,
          attempt: this.reconnectAttempts + 1,
        });

        this.socket = createConnection(
          {
            host: this.config.host,
            port: this.config.port,
          },
          () => {
            this.reconnectAttempts = 0;
            this.status = 'connected';
            this.setupHeartbeat();
            this.logger.logConnection('connected', {
              host: this.config.host,
              port: this.config.port,
            });
            this.eventListener.emit('connected', { timestamp: new Date().toISOString() });
            resolve();
          }
        );

        this.socket.on('data', (data) => this.handleData(data));
        this.socket.on('error', (err) => {
          this.status = 'disconnected';
          this.logger.error('Socket error', { error: err.message });
          this.eventListener.emit('error', { error: err.message });
        });
        this.socket.on('close', () => {
          this.status = 'disconnected';
          if (this.heartbeatTimer) {
            clearInterval(this.heartbeatTimer);
            this.heartbeatTimer = null;
          }
          this.logger.logConnection('disconnected');
          this.eventListener.emit('disconnected', { timestamp: new Date().toISOString() });
        });

        const timeout = setTimeout(() => {
          if (this.socket && this.socket.connecting) {
            this.socket.destroy();
            this.socket = null;
          }

          if (this.reconnectAttempts < this.config.reconnectMaxRetries) {
            this.reconnectAttempts++;
            const backoff = this.calculateBackoff(this.reconnectAttempts - 1);
            this.logger.warn('Connection timeout, retrying', {
              attempt: this.reconnectAttempts,
              backoffMs: backoff,
            });
            this.eventListener.emit('reconnecting', {
              attempt: this.reconnectAttempts,
              maxAttempts: this.config.reconnectMaxRetries,
            });

            setTimeout(attemptConnect, backoff);
          } else {
            const error = new Error(
              `Failed to connect to MCP server at ${this.config.host}:${this.config.port} after ${this.config.reconnectMaxRetries} retries`
            );
            reject(error);
          }
        }, 5000);

        // Clear timeout on successful connection
        if (this.socket) {
          this.socket.once('connect', () => clearTimeout(timeout));
        }
      };

      attemptConnect();
    });
  }

  /**
   * Close connection to MCP server
   */
  async disconnect(): Promise<void> {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }

    if (this.socket) {
      return new Promise((resolve) => {
        this.socket!.end(() => {
          this.socket = null;
          this.status = 'disconnected';
          resolve();
        });
      });
    }
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.status === 'connected' && this.socket !== null && !this.socket.destroyed;
  }

  /**
   * Get connection status
   */
  getStatus(): MCPConnectionStatus {
    return this.status;
  }

  /**
   * Register event listener
   */
  on(event: string, handler: (data: unknown) => void): void {
    this.eventListener.on(event, handler);
  }

  /**
   * Unregister event listener
   */
  off(event: string, handler: (data: unknown) => void): void {
    this.eventListener.off(event, handler);
  }

  /**
   * List available tools from MCP server (cached)
   */
  async listTools(): Promise<ToolDefinition[]> {
    // Check cache first
    const cached = await this.cache.load();
    if (cached && await this.cache.isValid()) {
      this.logger.debug('Tool schema loaded from cache');
      return cached || [];
    }

    // Query server if cache miss/expired
    this.logger.debug('Fetching tool schema from MCP server');
    const response = await this.sendRequest('tools/list', {});
    const tools: ToolDefinition[] = response.result || [];

    // Save to cache
    await this.cache.save(tools);

    this.logger.info('Tool schema cached', { toolCount: tools.length });
    return tools;
  }

  /**
   * Invoke a tool on MCP server
   */
  async invokeTool(toolName: string, params: Record<string, unknown>): Promise<any> {
    const response = await this.sendRequest('tools/execute', {
      tool: toolName,
      params,
    });

    if (response.error) {
      throw new Error(`Tool execution failed: ${response.error.message}`);
    }

    return response.result;
  }

  /**
   * Send JSON-RPC request and await response
   */
  private async sendRequest(method: string, params: Record<string, unknown>): Promise<any> {
    if (!this.isConnected()) {
      throw new Error('Not connected to MCP server');
    }

    const id = this.requestId++;
    const request: JSONRPCRequest = {
      jsonrpc: '2.0',
      id,
      method,
      params,
    };

    return new Promise((resolve, reject) => {
      this.requestMap.set(id, (response) => {
        if (response.error) {
          reject(new Error(response.error.message));
        } else {
          resolve(response);
        }
      });

      const json = JSON.stringify(request) + '\n';
      this.socket!.write(json, (err) => {
        if (err) {
          this.requestMap.delete(id);
          reject(err);
        }
      });

      // Timeout after 30s
      setTimeout(() => {
        if (this.requestMap.has(id)) {
          this.requestMap.delete(id);
          reject(new Error('Request timeout'));
        }
      }, 30000);
    });
  }

  /**
   * Handle incoming data from socket
   */
  private handleData(data: Buffer): void {
    this.buffer += data.toString('utf-8');

    // Process complete lines (JSON-RPC responses)
    const lines = this.buffer.split('\n');
    this.buffer = lines[lines.length - 1]; // Keep incomplete line in buffer

    for (let i = 0; i < lines.length - 1; i++) {
      const line = lines[i].trim();
      if (!line) continue;

      try {
        const response: JSONRPCResponse = JSON.parse(line);
        if (response.id && this.requestMap.has(response.id)) {
          const handler = this.requestMap.get(response.id)!;
          this.requestMap.delete(response.id);
          handler(response);
        }
      } catch (err) {
        this.logger.error('Failed to parse JSON-RPC response', { error: (err as Error).message });
      }
    }
  }

  /**
   * Setup heartbeat to detect disconnection
   */
  private setupHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
    }

    this.heartbeatTimer = setInterval(() => {
      if (!this.isConnected()) {
        if (this.heartbeatTimer) {
          clearInterval(this.heartbeatTimer);
          this.heartbeatTimer = null;
        }
        return;
      }

      this.logger.debug('Heartbeat');
      // Send keepalive ping
      this.socket!.write(JSON.stringify({ jsonrpc: '2.0', method: 'ping' }) + '\n', (err) => {
        if (err) {
          this.logger.error('Heartbeat send failed', { error: err.message });
        }
      });
    }, 30000); // Every 30 seconds
  }

  /**
   * Calculate exponential backoff: 1s, 2s, 4s, 8s, 16s, 32s (max)
   */
  private calculateBackoff(attempt: number): number {
    const baseMs = 1000;
    const maxMs = 32000;
    const backoffMs = baseMs * Math.pow(2, attempt);
    return Math.min(backoffMs, maxMs);
  }
}

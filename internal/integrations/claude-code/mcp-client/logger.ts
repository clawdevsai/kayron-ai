/**
 * JSON Logger for Kayron MCP Operations
 * Logs all tool invocations to ~/.claude/logs/kayron-mcp.log (JSONL format)
 */

import * as fs from 'fs';
import * as path from 'path';
import { ExecutionLog } from './types';

export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

export interface LogEntry {
  timestamp: string;
  level: LogLevel;
  message: string;
  data?: unknown;
}

export class Logger {
  private logDir: string;
  private logFile: string;
  private logLevel: LogLevel;
  private levelPriority: Record<LogLevel, number> = {
    debug: 0,
    info: 1,
    warn: 2,
    error: 3,
  };

  constructor(level: LogLevel = 'info') {
    this.logDir = path.join(process.env.HOME || '.', '.claude', 'logs');
    this.logFile = path.join(this.logDir, 'kayron-mcp.log');
    this.logLevel = level;
    this.ensureLogDir();
  }

  /**
   * Ensure log directory exists
   */
  private ensureLogDir(): void {
    if (!fs.existsSync(this.logDir)) {
      fs.mkdirSync(this.logDir, { recursive: true });
    }
  }

  /**
   * Write log entry to file (JSONL format)
   */
  private writeLog(entry: LogEntry): void {
    const line = JSON.stringify(entry) + '\n';

    fs.appendFile(this.logFile, line, (err) => {
      if (err) {
        console.error(`Failed to write log: ${err.message}`);
      }
    });
  }

  /**
   * Check if level should be logged
   */
  private shouldLog(level: LogLevel): boolean {
    return this.levelPriority[level] >= this.levelPriority[this.logLevel];
  }

  /**
   * Log execution result (for audit trail)
   */
  logExecution(execution: ExecutionLog): void {
    if (!this.shouldLog('info')) return;

    const entry: LogEntry = {
      timestamp: execution.timestamp,
      level: 'info',
      message: `Tool execution: ${execution.toolName}`,
      data: {
        toolName: execution.toolName,
        inputParams: execution.inputParams,
        output: execution.output,
        error: execution.error,
        durationMs: execution.executionDurationMs,
        retryCount: execution.retryCount,
        idempotencyKey: execution.idempotencyKey,
        success: !execution.error,
      },
    };

    this.writeLog(entry);
  }

  /**
   * Log debug message
   */
  debug(message: string, data?: unknown): void {
    if (!this.shouldLog('debug')) return;

    this.writeLog({
      timestamp: new Date().toISOString(),
      level: 'debug',
      message,
      data,
    });
  }

  /**
   * Log info message
   */
  info(message: string, data?: unknown): void {
    if (!this.shouldLog('info')) return;

    this.writeLog({
      timestamp: new Date().toISOString(),
      level: 'info',
      message,
      data,
    });
  }

  /**
   * Log warning message
   */
  warn(message: string, data?: unknown): void {
    if (!this.shouldLog('warn')) return;

    this.writeLog({
      timestamp: new Date().toISOString(),
      level: 'warn',
      message,
      data,
    });
  }

  /**
   * Log error message
   */
  error(message: string, data?: unknown): void {
    if (!this.shouldLog('error')) return;

    this.writeLog({
      timestamp: new Date().toISOString(),
      level: 'error',
      message,
      data,
    });
  }

  /**
   * Log MCP connection event
   */
  logConnection(event: 'connected' | 'disconnected' | 'reconnecting' | 'error', data?: unknown): void {
    this.info(`MCP ${event}`, data);
  }

  /**
   * Get log file path
   */
  getLogFilePath(): string {
    return this.logFile;
  }

  /**
   * Set log level
   */
  setLevel(level: LogLevel): void {
    this.logLevel = level;
  }
}

/**
 * Global logger instance
 */
export const globalLogger = new Logger('info');

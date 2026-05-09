/**
 * Data Model Implementation
 * Represents entities and state in the Kayron MCP integration
 */

import { Position, ExecutionLog, PendingOperation, SkillDefinition } from './types';
import { v4 as uuidv4 } from 'uuid';

/**
 * Position Model
 * Represents open trading positions
 */
export class PositionModel implements Position {
  ticket: number;
  symbol: string;
  type: 'buy' | 'sell';
  volume: string;
  entryPrice: string;
  currentPrice: string;
  pnl: string;
  pnlPercent: string;
  openTime: string;
  lastUpdateTime: string;

  constructor(data: Position) {
    this.ticket = data.ticket;
    this.symbol = data.symbol;
    this.type = data.type;
    this.volume = data.volume;
    this.entryPrice = data.entryPrice;
    this.currentPrice = data.currentPrice;
    this.pnl = data.pnl;
    this.pnlPercent = data.pnlPercent;
    this.openTime = data.openTime;
    this.lastUpdateTime = data.lastUpdateTime;
  }

  /**
   * Recalculate P&L based on current price
   */
  updatePrice(newPrice: string): void {
    this.currentPrice = newPrice;
    this.lastUpdateTime = new Date().toISOString();
    // P&L calculation: (currentPrice - entryPrice) * volume (buy) or (entryPrice - currentPrice) * volume (sell)
    // Decimal arithmetic handled by caller (no floating point)
  }

  toJSON(): Position {
    return {
      ticket: this.ticket,
      symbol: this.symbol,
      type: this.type,
      volume: this.volume,
      entryPrice: this.entryPrice,
      currentPrice: this.currentPrice,
      pnl: this.pnl,
      pnlPercent: this.pnlPercent,
      openTime: this.openTime,
      lastUpdateTime: this.lastUpdateTime,
    };
  }
}

/**
 * Execution Log Model
 * Immutable audit trail of tool invocations
 */
export class ExecutionLogModel implements ExecutionLog {
  id: string;
  timestamp: string;
  toolName: string;
  inputParams: Record<string, unknown>;
  output?: Record<string, unknown> | string;
  error?: Record<string, unknown> | string;
  executionDurationMs: number;
  retryCount: number;
  idempotencyKey: string;
  userId?: string;

  constructor(data: Omit<ExecutionLog, 'id'> & { id?: string }) {
    this.id = data.id || uuidv4();
    this.timestamp = data.timestamp;
    this.toolName = data.toolName;
    this.inputParams = data.inputParams;
    this.output = data.output;
    this.error = data.error;
    this.executionDurationMs = data.executionDurationMs;
    this.retryCount = data.retryCount;
    this.idempotencyKey = data.idempotencyKey;
    this.userId = data.userId;
  }

  /**
   * Serialize to JSONL format for file logging
   */
  toJSONL(): string {
    return JSON.stringify({
      id: this.id,
      timestamp: this.timestamp,
      toolName: this.toolName,
      inputParams: this.inputParams,
      output: this.output,
      error: this.error,
      executionDurationMs: this.executionDurationMs,
      retryCount: this.retryCount,
      idempotencyKey: this.idempotencyKey,
      userId: this.userId,
    });
  }

  /**
   * Create from JSONL line
   */
  static fromJSONL(line: string): ExecutionLogModel {
    const data = JSON.parse(line);
    return new ExecutionLogModel(data);
  }
}

/**
 * Pending Operation Model
 * Represents queued operations awaiting execution
 */
export class PendingOperationModel implements PendingOperation {
  id: string;
  toolName: string;
  params: Record<string, unknown>;
  createdAt: string;
  retryCount: number;
  idempotencyKey: string;

  constructor(data: Omit<PendingOperation, 'id'> & { id?: string }) {
    this.id = data.id || uuidv4();
    this.toolName = data.toolName;
    this.params = data.params;
    this.createdAt = data.createdAt;
    this.retryCount = data.retryCount || 0;
    this.idempotencyKey = data.idempotencyKey;
  }

  /**
   * Increment retry count
   */
  incrementRetry(): void {
    this.retryCount += 1;
  }

  /**
   * Check if retries exhausted (max 5)
   */
  isRetryExhausted(): boolean {
    return this.retryCount >= 5;
  }

  toJSON(): PendingOperation {
    return {
      id: this.id,
      toolName: this.toolName,
      params: this.params,
      createdAt: this.createdAt,
      retryCount: this.retryCount,
      idempotencyKey: this.idempotencyKey,
    };
  }
}

/**
 * Skill Definition Model
 * Represents reusable trading skills
 */
export class SkillDefinitionModel implements SkillDefinition {
  id: string;
  name: string;
  description: string;
  skillPath: string;
  content: string;
  toolDependencies: string[];
  createdAt: string;
  modifiedAt: string;
  enabled: boolean;

  constructor(data: SkillDefinition) {
    this.id = data.id;
    this.name = data.name;
    this.description = data.description;
    this.skillPath = data.skillPath;
    this.content = data.content;
    this.toolDependencies = data.toolDependencies;
    this.createdAt = data.createdAt;
    this.modifiedAt = data.modifiedAt;
    this.enabled = data.enabled;
  }

  /**
   * Validate skill content (basic markdown + frontmatter check)
   */
  validate(): string[] {
    const errors: string[] = [];

    if (!this.name || !/^[a-z0-9-]+$/.test(this.name)) {
      errors.push('Skill name must be lowercase alphanumeric + hyphens');
    }

    if (!this.skillPath.includes('.claude/skills/') || !this.skillPath.endsWith('SKILL.md')) {
      errors.push('Skill path must be in ~/.claude/skills/[name]/SKILL.md');
    }

    if (!this.content.includes('---')) {
      errors.push('Skill content must include YAML frontmatter (---)');
    }

    if (!Array.isArray(this.toolDependencies) || this.toolDependencies.length === 0) {
      errors.push('Skill must declare at least one tool dependency');
    }

    return errors;
  }

  /**
   * Update modification timestamp
   */
  touch(): void {
    this.modifiedAt = new Date().toISOString();
  }

  toJSON(): SkillDefinition {
    return {
      id: this.id,
      name: this.name,
      description: this.description,
      skillPath: this.skillPath,
      content: this.content,
      toolDependencies: this.toolDependencies,
      createdAt: this.createdAt,
      modifiedAt: this.modifiedAt,
      enabled: this.enabled,
    };
  }
}

/**
 * Factory functions for creating models
 */
export namespace ModelFactory {
  export function createPosition(data: Position): PositionModel {
    return new PositionModel(data);
  }

  export function createExecutionLog(data: Omit<ExecutionLog, 'id'>): ExecutionLogModel {
    return new ExecutionLogModel(data);
  }

  export function createPendingOperation(
    toolName: string,
    params: Record<string, unknown>,
    idempotencyKey: string
  ): PendingOperationModel {
    return new PendingOperationModel({
      toolName,
      params,
      createdAt: new Date().toISOString(),
      retryCount: 0,
      idempotencyKey,
    });
  }

  export function createSkill(data: SkillDefinition): SkillDefinitionModel {
    return new SkillDefinitionModel(data);
  }
}

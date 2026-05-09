/**
 * Pending Operations Queue
 * Persists operations during disconnection, replays on reconnect
 */

import * as fs from 'fs';
import * as path from 'path';
import { PendingOperation } from './types';
import { PendingOperationModel } from './data-model';

export class OperationsQueue {
  private queueDir: string;
  private queueFile: string;
  private operations: Map<string, PendingOperationModel> = new Map();

  constructor() {
    this.queueDir = path.join(process.env.HOME || '.', '.claude', 'cache');
    this.queueFile = path.join(this.queueDir, 'kayron-queue.json');
  }

  /**
   * Ensure queue directory exists
   */
  private ensureQueueDir(): void {
    if (!fs.existsSync(this.queueDir)) {
      fs.mkdirSync(this.queueDir, { recursive: true });
    }
  }

  /**
   * Add operation to queue (in-memory + persist to file)
   */
  async add(operation: PendingOperationModel): Promise<void> {
    this.operations.set(operation.id, operation);
    await this.persistToFile();
  }

  /**
   * Get all pending operations
   */
  getAll(): PendingOperation[] {
    return Array.from(this.operations.values()).map((op) => op.toJSON());
  }

  /**
   * Get operation by ID
   */
  get(id: string): PendingOperationModel | undefined {
    return this.operations.get(id);
  }

  /**
   * Remove operation from queue
   */
  async remove(id: string): Promise<void> {
    this.operations.delete(id);
    await this.persistToFile();
  }

  /**
   * Remove multiple operations
   */
  async removeMany(ids: string[]): Promise<void> {
    ids.forEach((id) => this.operations.delete(id));
    await this.persistToFile();
  }

  /**
   * Get next operation to retry (oldest first, FIFO)
   */
  getNextToRetry(): PendingOperationModel | undefined {
    const ops = Array.from(this.operations.values()).sort(
      (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
    );

    return ops.find((op) => !op.isRetryExhausted());
  }

  /**
   * Persist queue to file
   */
  private async persistToFile(): Promise<void> {
    this.ensureQueueDir();

    const queueData = {
      timestamp: new Date().toISOString(),
      operationCount: this.operations.size,
      operations: Array.from(this.operations.values()).map((op) => op.toJSON()),
    };

    return new Promise((resolve, reject) => {
      fs.writeFile(this.queueFile, JSON.stringify(queueData, null, 2), (err) => {
        if (err) reject(err);
        else resolve();
      });
    });
  }

  /**
   * Load queue from file (on startup)
   */
  async loadFromFile(): Promise<void> {
    if (!fs.existsSync(this.queueFile)) {
      this.operations.clear();
      return;
    }

    return new Promise((resolve) => {
      fs.readFile(this.queueFile, 'utf-8', (err, data) => {
        if (err) {
          this.operations.clear();
          resolve();
          return;
        }

        try {
          const queueData = JSON.parse(data);
          this.operations.clear();

          if (Array.isArray(queueData.operations)) {
            queueData.operations.forEach((op: PendingOperation) => {
              const model = new PendingOperationModel(op);
              this.operations.set(op.id, model);
            });
          }

          resolve();
        } catch {
          this.operations.clear();
          resolve();
        }
      });
    });
  }

  /**
   * Clear queue (both in-memory and file)
   */
  async clear(): Promise<void> {
    this.operations.clear();
    await this.persistToFile();
  }

  /**
   * Get queue size
   */
  size(): number {
    return this.operations.size;
  }

  /**
   * Check if queue is empty
   */
  isEmpty(): boolean {
    return this.operations.size === 0;
  }

  /**
   * Get queue file path
   */
  getQueueFilePath(): string {
    return this.queueFile;
  }
}

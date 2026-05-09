/**
 * Tool Schema Cache Manager
 * Caches MCP tool schemas locally with TTL validation
 */

import * as fs from 'fs';
import * as path from 'path';
import { ToolDefinition } from './types';

export interface CacheMetadata {
  timestamp: number;
  ttlMs: number;
  toolCount: number;
}

export interface CachedSchema {
  metadata: CacheMetadata;
  tools: ToolDefinition[];
}

export class SchemaCache {
  private cacheDir: string;
  private cacheFile: string;
  private ttlMs: number;

  constructor(ttlMinutes: number = 60) {
    this.cacheDir = path.join(process.env.HOME || '.', '.claude', 'cache');
    this.cacheFile = path.join(this.cacheDir, 'kayron-tools.json');
    this.ttlMs = ttlMinutes * 60 * 1000;
  }

  /**
   * Ensure cache directory exists
   */
  private ensureCacheDir(): void {
    if (!fs.existsSync(this.cacheDir)) {
      fs.mkdirSync(this.cacheDir, { recursive: true });
    }
  }

  /**
   * Save tool schemas to cache
   */
  async save(tools: ToolDefinition[]): Promise<void> {
    this.ensureCacheDir();

    const cached: CachedSchema = {
      metadata: {
        timestamp: Date.now(),
        ttlMs: this.ttlMs,
        toolCount: tools.length,
      },
      tools,
    };

    return new Promise((resolve, reject) => {
      fs.writeFile(this.cacheFile, JSON.stringify(cached, null, 2), (err) => {
        if (err) reject(err);
        else resolve();
      });
    });
  }

  /**
   * Load tool schemas from cache if valid
   */
  async load(): Promise<ToolDefinition[] | null> {
    if (!fs.existsSync(this.cacheFile)) {
      return null;
    }

    return new Promise((resolve) => {
      fs.readFile(this.cacheFile, 'utf-8', (err, data) => {
        if (err) {
          resolve(null);
          return;
        }

        try {
          const cached: CachedSchema = JSON.parse(data);

          // Check if cache is still valid
          const age = Date.now() - cached.metadata.timestamp;
          if (age > cached.metadata.ttlMs) {
            resolve(null);
            return;
          }

          resolve(cached.tools);
        } catch {
          resolve(null);
        }
      });
    });
  }

  /**
   * Check if cache is valid (exists and not expired)
   */
  async isValid(): Promise<boolean> {
    const tools = await this.load();
    return tools !== null;
  }

  /**
   * Get cache age in milliseconds
   */
  async getAge(): Promise<number | null> {
    if (!fs.existsSync(this.cacheFile)) {
      return null;
    }

    return new Promise((resolve) => {
      fs.stat(this.cacheFile, (err, stats) => {
        if (err) {
          resolve(null);
        } else {
          resolve(Date.now() - stats.mtimeMs);
        }
      });
    });
  }

  /**
   * Clear cache
   */
  async clear(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (fs.existsSync(this.cacheFile)) {
        fs.unlink(this.cacheFile, (err) => {
          if (err) reject(err);
          else resolve();
        });
      } else {
        resolve();
      }
    });
  }

  /**
   * Find tool by name in cache
   */
  async findTool(toolName: string): Promise<ToolDefinition | null> {
    const tools = await this.load();
    if (!tools) return null;

    return tools.find((t) => t.name === toolName) || null;
  }

  /**
   * Get cache file path
   */
  getCacheFilePath(): string {
    return this.cacheFile;
  }
}

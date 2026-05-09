/**
 * Unit Tests: Persistent Queue
 * Queue persists to ~/.kayron-queue.jsonl, survives reconnect
 */

import { OperationsQueue } from '../../mcp-client/queue';
import { PendingOperationModel } from '../../mcp-client/data-model';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';

describe('OperationsQueue', () => {
  let queue: OperationsQueue;
  const testQueuePath = path.join(os.homedir(), '.kayron-queue-test.jsonl');

  beforeEach(async () => {
    if (fs.existsSync(testQueuePath)) {
      fs.unlinkSync(testQueuePath);
    }
    queue = new OperationsQueue();
  });

  afterEach(() => {
    if (fs.existsSync(testQueuePath)) {
      fs.unlinkSync(testQueuePath);
    }
  });

  describe('Enqueue & Dequeue', () => {
    it('should enqueue operation', () => {
      const op = { toolName: 'place-order', params: { symbol: 'EURUSD', volume: 0.1, type: 'BUY' } };
      const queued = queue.enqueue(op);

      expect(queued.id).toBeDefined();
      expect(queued.toolName).toBe('place-order');
      expect(queued.status).toBe('pending');
    });

    it('should dequeue FIFO', () => {
      queue.enqueue({ toolName: 'place-order', params: { symbol: 'EURUSD', volume: 0.1, type: 'BUY' } });
      queue.enqueue({ toolName: 'close-position', params: { ticket: 123 } });

      const first = queue.dequeue();
      expect(first?.toolName).toBe('place-order');

      const second = queue.dequeue();
      expect(second?.toolName).toBe('close-position');

      expect(queue.dequeue()).toBeUndefined();
    });

    it('should generate unique IDs', () => {
      const op1 = queue.enqueue({ toolName: 'place-order', params: { symbol: 'EURUSD', volume: 0.1, type: 'BUY' } });
      const op2 = queue.enqueue({ toolName: 'place-order', params: { symbol: 'EURUSD', volume: 0.1, type: 'BUY' } });

      expect(op1.id).not.toBe(op2.id);
    });
  });

  describe('Persistence', () => {
    it('should persist enqueued operation to disk', () => {
      queue.enqueue({ toolName: 'place-order', params: { symbol: 'EURUSD', volume: 0.1, type: 'BUY' } });

      expect(fs.existsSync(testQueuePath)).toBe(true);
      const content = fs.readFileSync(testQueuePath, 'utf-8');
      expect(content).toContain('place-order');
    });

    it('should load operations from disk', () => {
      const queue1 = new PersistentQueue(testQueuePath);
      queue1.enqueue({ toolName: 'place-order', params: { symbol: 'EURUSD', volume: 0.1, type: 'BUY' } });
      queue1.enqueue({ toolName: 'close-position', params: { ticket: 456 } });

      const queue2 = new PersistentQueue(testQueuePath);
      expect(queue2.length()).toBe(2);

      const first = queue2.dequeue();
      expect(first?.toolName).toBe('place-order');
    });

    it('should persist on dequeue', () => {
      queue.enqueue({ toolName: 'place-order', params: { symbol: 'EURUSD', volume: 0.1, type: 'BUY' } });
      queue.enqueue({ toolName: 'close-position', params: { ticket: 789 } });

      queue.dequeue();

      const queue2 = new PersistentQueue(testQueuePath);
      expect(queue2.length()).toBe(1);
      expect(queue2.dequeue()?.toolName).toBe('close-position');
    });

    it('should handle empty queue file', () => {
      fs.writeFileSync(testQueuePath, '');
      const newQueue = new PersistentQueue(testQueuePath);
      expect(newQueue.length()).toBe(0);
    });

    it('should skip corrupted lines in queue file', () => {
      const op1 = { id: 'op1', toolName: 'place-order', params: {}, timestamp: Date.now(), status: 'pending' as const };
      fs.writeFileSync(testQueuePath, JSON.stringify(op1) + '\n' + 'INVALID JSON\n');

      const newQueue = new PersistentQueue(testQueuePath);
      expect(newQueue.length()).toBe(1);
      expect(newQueue.dequeue()?.id).toBe('op1');
    });
  });

  describe('Status Tracking', () => {
    it('should mark operation as executed', () => {
      const op = queue.enqueue({ toolName: 'place-order', params: { symbol: 'EURUSD', volume: 0.1, type: 'BUY' } });
      queue.markExecuted(op.id);

      const all = queue.getAll();
      const executed = all.find(o => o.id === op.id);
      expect(executed?.status).toBe('executed');
    });

    it('should persist executed status', () => {
      const op = queue.enqueue({ toolName: 'place-order', params: { symbol: 'EURUSD', volume: 0.1, type: 'BUY' } });
      queue.markExecuted(op.id);

      const queue2 = new PersistentQueue(testQueuePath);
      const all = queue2.getAll();
      const executed = all.find(o => o.id === op.id);
      expect(executed?.status).toBe('executed');
    });
  });

  describe('Query & Remove', () => {
    it('should get all operations', () => {
      queue.enqueue({ toolName: 'place-order', params: { symbol: 'EURUSD', volume: 0.1, type: 'BUY' } });
      queue.enqueue({ toolName: 'close-position', params: { ticket: 111 } });

      const all = queue.getAll();
      expect(all.length).toBe(2);
      expect(all[0].toolName).toBe('place-order');
      expect(all[1].toolName).toBe('close-position');
    });

    it('should remove operation by ID', () => {
      const op1 = queue.enqueue({ toolName: 'place-order', params: { symbol: 'EURUSD', volume: 0.1, type: 'BUY' } });
      const op2 = queue.enqueue({ toolName: 'close-position', params: { ticket: 222 } });

      queue.remove(op1.id);

      const all = queue.getAll();
      expect(all.length).toBe(1);
      expect(all[0].id).toBe(op2.id);
    });

    it('should clear all operations', () => {
      queue.enqueue({ toolName: 'place-order', params: { symbol: 'EURUSD', volume: 0.1, type: 'BUY' } });
      queue.enqueue({ toolName: 'close-position', params: { ticket: 333 } });

      queue.clear();

      expect(queue.length()).toBe(0);
    });

    it('should persist removal', () => {
      const op = queue.enqueue({ toolName: 'place-order', params: { symbol: 'EURUSD', volume: 0.1, type: 'BUY' } });
      queue.remove(op.id);

      const queue2 = new PersistentQueue(testQueuePath);
      expect(queue2.length()).toBe(0);
    });
  });

  describe('Timestamps', () => {
    it('should record enqueue timestamp', () => {
      const beforeEnqueue = Date.now();
      const op = queue.enqueue({ toolName: 'place-order', params: { symbol: 'EURUSD', volume: 0.1, type: 'BUY' } });
      const afterEnqueue = Date.now();

      expect(op.timestamp).toBeGreaterThanOrEqual(beforeEnqueue);
      expect(op.timestamp).toBeLessThanOrEqual(afterEnqueue);
    });
  });
});

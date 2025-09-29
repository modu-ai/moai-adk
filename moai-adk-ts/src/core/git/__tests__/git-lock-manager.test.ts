/**
import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';
 * @file GitLockManager test suite - RED phase TDD
 * @author MoAI Team
 * @tags @TEST:GIT-LOCK-001 @REQ:CORE-SYSTEM-013
 */

import fs from 'fs-extra';
import { GitLockedException } from '../../../types/git';
import { GitLockManager } from '../git-lock-manager';

// Mock fs-extra
vi.mock('fs-extra');
const mockFs = fs as vi.Mocked<typeof fs>;

describe('GitLockManager', () => {
  let lockManager: GitLockManager;
  let tempDir: string;

  beforeEach(() => {
    tempDir = '/test/project';
    lockManager = new GitLockManager(tempDir);
    vi.clearAllMocks();

    // Reset all mocks to return resolved promises by default
    (mockFs.pathExists as vi.Mock).mockResolvedValue(false);
    (mockFs.ensureDir as vi.Mock).mockResolvedValue(undefined);
    (mockFs.writeJson as vi.Mock).mockResolvedValue(undefined);
    (mockFs.readJson as vi.Mock).mockResolvedValue({});
    (mockFs.remove as vi.Mock).mockResolvedValue(undefined);
  });

  describe('constructor', () => {
    it('should initialize with default values', () => {
      const manager = new GitLockManager();
      expect(manager).toBeDefined();
    });

    it('should initialize with custom project directory', () => {
      const manager = new GitLockManager('/custom/path');
      expect(manager).toBeDefined();
    });

    it('should throw error for invalid project directory', () => {
      expect(() => {
        new GitLockManager(123 as any);
      }).toThrow('Invalid project directory');
    });
  });

  describe('isLocked', () => {
    it('should return false when lock file does not exist', async () => {
      (mockFs.pathExists as vi.Mock).mockResolvedValue(false);

      const result = await lockManager.isLocked();

      expect(result).toBe(false);
    });

    it('should return true when valid lock file exists', async () => {
      const lockInfo = {
        pid: process.pid, // Use current process to ensure it's running
        timestamp: Date.now(),
        operation: 'commit',
        user: 'test-user',
      };

      (mockFs.pathExists as vi.Mock).mockResolvedValue(true);
      (mockFs.readJson as vi.Mock).mockResolvedValue(lockInfo);

      const result = await lockManager.isLocked();

      expect(result).toBe(true);
    });

    it('should clean up stale lock and return false', async () => {
      const staleLockInfo = {
        pid: 99999, // Non-existent process
        timestamp: Date.now() - 60000, // 1 minute old
        operation: 'commit',
        user: 'test-user',
      };

      (mockFs.pathExists as vi.Mock).mockResolvedValue(true);
      (mockFs.readJson as vi.Mock).mockResolvedValue(staleLockInfo);

      const result = await lockManager.isLocked();

      expect(result).toBe(false);
      expect(mockFs.remove).toHaveBeenCalled();
    });

    it('should handle corrupt lock file gracefully', async () => {
      (mockFs.pathExists as vi.Mock).mockResolvedValue(true);
      (mockFs.readJson as vi.Mock).mockRejectedValue(new Error('Invalid JSON'));

      const result = await lockManager.isLocked();

      expect(result).toBe(false);
    });
  });

  describe('acquireLock', () => {
    it('should acquire lock when not locked', async () => {
      (mockFs.pathExists as vi.Mock).mockResolvedValue(false);

      const lockContext = await lockManager.acquireLock(false); // Don't wait

      expect(lockContext).toBeDefined();
      expect(lockContext.lockInfo).toBeDefined();
      expect(lockContext.acquired).toBeInstanceOf(Date);
      expect(typeof lockContext.release).toBe('function');
      expect(mockFs.ensureDir).toHaveBeenCalled();
      expect(mockFs.writeJson).toHaveBeenCalled();
    });

    it('should throw error when already locked', async () => {
      const existingLock = {
        pid: process.pid,
        timestamp: Date.now(),
        operation: 'commit',
        user: 'other-user',
      };

      (mockFs.pathExists as vi.Mock).mockResolvedValue(true);
      (mockFs.readJson as vi.Mock).mockResolvedValue(existingLock);

      await expect(lockManager.acquireLock(false)).rejects.toThrow(
        GitLockedException
      );
    });
  });

  describe('releaseLock', () => {
    it('should release lock when exists', async () => {
      (mockFs.pathExists as vi.Mock).mockResolvedValue(true);

      await lockManager.releaseLock();

      expect(mockFs.remove).toHaveBeenCalled();
    });

    it('should not throw when lock does not exist', async () => {
      (mockFs.pathExists as vi.Mock).mockResolvedValue(false);

      await expect(lockManager.releaseLock()).resolves.toBeUndefined();
    });
  });

  describe('withLock', () => {
    it('should execute operation with lock', async () => {
      (mockFs.pathExists as vi.Mock).mockResolvedValue(false);

      const operation = vi.fn().mockResolvedValue('result');
      const releaseLockSpy = vi.spyOn(lockManager, 'releaseLock');

      const result = await lockManager.withLock(operation, 'test-operation');

      expect(result).toBe('result');
      expect(operation).toHaveBeenCalled();
      expect(mockFs.writeJson).toHaveBeenCalled();
      expect(releaseLockSpy).toHaveBeenCalled();
    });

    it('should release lock even if operation fails', async () => {
      (mockFs.pathExists as vi.Mock).mockResolvedValue(false);

      const operation = vi
        .fn()
        .mockRejectedValue(new Error('Operation failed'));
      const releaseLockSpy = vi.spyOn(lockManager, 'releaseLock');

      await expect(
        lockManager.withLock(operation, 'test-operation')
      ).rejects.toThrow('Operation failed');

      expect(releaseLockSpy).toHaveBeenCalled(); // Lock should still be released
    });
  });

  describe('getLockStatus', () => {
    it('should return complete lock status', async () => {
      const lockInfo = {
        pid: process.pid,
        timestamp: Date.now() - 5000, // 5 seconds ago
        operation: 'commit',
        user: 'test-user',
      };

      (mockFs.pathExists as vi.Mock).mockResolvedValue(true);
      (mockFs.readJson as vi.Mock).mockResolvedValue(lockInfo);

      const status = await lockManager.getLockStatus();

      expect(status).toMatchObject({
        isLocked: true,
        lockFileExists: true,
        lockInfo,
        processRunning: true,
        lockAgeSeconds: expect.any(Number),
      });
    });

    it('should return status when no lock exists', async () => {
      (mockFs.pathExists as vi.Mock).mockResolvedValue(false);

      const status = await lockManager.getLockStatus();

      expect(status).toMatchObject({
        isLocked: false,
        lockFileExists: false,
        lockInfo: null,
      });
    });
  });

  describe('cleanupStaleLocks', () => {
    it('should remove stale lock files', async () => {
      const staleLock = {
        pid: 99999,
        timestamp: Date.now() - 300000, // 5 minutes ago
        operation: 'commit',
        user: 'test-user',
      };

      (mockFs.pathExists as vi.Mock).mockResolvedValue(true);
      (mockFs.readJson as vi.Mock).mockResolvedValue(staleLock);

      await lockManager.cleanupStaleLocks();

      expect(mockFs.remove).toHaveBeenCalled();
    });

    it('should preserve active locks', async () => {
      const activeLock = {
        pid: process.pid,
        timestamp: Date.now(),
        operation: 'commit',
        user: 'test-user',
      };

      (mockFs.pathExists as vi.Mock).mockResolvedValue(true);
      (mockFs.readJson as vi.Mock).mockResolvedValue(activeLock);

      await lockManager.cleanupStaleLocks();

      expect(mockFs.remove).not.toHaveBeenCalled();
    });
  });
});

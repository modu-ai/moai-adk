/**
 * @file GitLockManager test suite
 * @author MoAI Team
 * @tags @TEST:GIT-LOCK-001 @SPEC:CORE-SYSTEM-013
 */

import { describe, test, expect, beforeEach, vi, afterEach } from 'vitest';
import { GitLockedException } from '../../../types/git';
import { GitLockManager } from '../git-lock-manager';

// Mock fs-extra module - 팩토리 함수 내부에서 모든 것 정의
vi.mock('fs-extra', () => {
  const mockPathExists = vi.fn();
  const mockEnsureDir = vi.fn();
  const mockWriteJson = vi.fn();
  const mockReadJson = vi.fn();
  const mockRemove = vi.fn();

  return {
    default: {
      pathExists: mockPathExists,
      ensureDir: mockEnsureDir,
      writeJson: mockWriteJson,
      readJson: mockReadJson,
      remove: mockRemove,
    },
    pathExists: mockPathExists,
    ensureDir: mockEnsureDir,
    writeJson: mockWriteJson,
    readJson: mockReadJson,
    remove: mockRemove,
  };
});

// Import mocked functions after vi.mock
import * as fs from 'fs-extra';
const mockPathExists = vi.mocked(fs.pathExists);
const mockEnsureDir = vi.mocked(fs.ensureDir);
const mockWriteJson = vi.mocked(fs.writeJson);
const mockReadJson = vi.mocked(fs.readJson);
const mockRemove = vi.mocked(fs.remove);

describe('GitLockManager', () => {
  let lockManager: GitLockManager;
  let tempDir: string;

  beforeEach(() => {
    tempDir = '/test/project';
    lockManager = new GitLockManager(tempDir);
    vi.clearAllMocks();

    // Reset all mocks to return resolved promises by default
    mockPathExists.mockResolvedValue(false);
    mockEnsureDir.mockResolvedValue(undefined);
    mockWriteJson.mockResolvedValue(undefined);
    mockReadJson.mockResolvedValue({});
    mockRemove.mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('constructor', () => {
    test('should initialize with default values', () => {
      const manager = new GitLockManager();
      expect(manager).toBeDefined();
    });

    test('should initialize with custom project directory', () => {
      const manager = new GitLockManager('/custom/path');
      expect(manager).toBeDefined();
    });

    test('should throw error for invalid project directory', () => {
      expect(() => {
        new GitLockManager(123 as any);
      }).toThrow('Invalid project directory');
    });
  });

  describe('isLocked', () => {
    test('should return false when lock file does not exist', async () => {
      mockPathExists.mockResolvedValue(false);

      const result = await lockManager.isLocked();

      expect(result).toBe(false);
    });

    test('should return true when valid lock file exists', async () => {
      const lockInfo = {
        pid: process.pid, // Use current process to ensure it's running
        timestamp: Date.now(),
        operation: 'commit',
        user: 'test-user',
        hostname: 'test-host',
        workingDir: tempDir,
      };

      mockPathExists.mockResolvedValue(true);
      mockReadJson.mockResolvedValue(lockInfo);

      const result = await lockManager.isLocked();

      expect(result).toBe(true);
    });

    test('should clean up stale lock and return false', async () => {
      const staleLockInfo = {
        pid: 99999, // Non-existent process
        timestamp: Date.now() - 60000, // 1 minute old
        operation: 'commit',
        user: 'test-user',
        hostname: 'test-host',
        workingDir: tempDir,
      };

      mockPathExists.mockResolvedValue(true);
      mockReadJson.mockResolvedValue(staleLockInfo);

      const result = await lockManager.isLocked();

      expect(result).toBe(false);
      expect(mockRemove).toHaveBeenCalled();
    });

    test('should handle corrupt lock file gracefully', async () => {
      mockPathExists.mockResolvedValue(true);
      mockReadJson.mockRejectedValue(new Error('Invalid JSON'));

      const result = await lockManager.isLocked();

      expect(result).toBe(false);
    });

    test('should clean up expired lock', async () => {
      const expiredLock = {
        pid: process.pid,
        timestamp: Date.now() - 400000, // > 5 minutes old
        operation: 'commit',
        user: 'test-user',
        hostname: 'test-host',
        workingDir: tempDir,
      };

      mockPathExists.mockResolvedValue(true);
      mockReadJson.mockResolvedValue(expiredLock);

      const result = await lockManager.isLocked();

      expect(result).toBe(false);
      expect(mockRemove).toHaveBeenCalled();
    });
  });

  describe('acquireLock', () => {
    test('should acquire lock when not locked', async () => {
      mockPathExists.mockResolvedValue(false);

      const lockContext = await lockManager.acquireLock(false); // Don't wait

      expect(lockContext).toBeDefined();
      expect(lockContext.lockInfo).toBeDefined();
      expect(lockContext.acquired).toBeInstanceOf(Date);
      expect(typeof lockContext.release).toBe('function');
      expect(mockEnsureDir).toHaveBeenCalled();
      expect(mockWriteJson).toHaveBeenCalled();
    });

    test('should throw error when already locked and not waiting', async () => {
      const existingLock = {
        pid: process.pid,
        timestamp: Date.now(),
        operation: 'commit',
        user: 'other-user',
        hostname: 'other-host',
        workingDir: tempDir,
      };

      mockPathExists.mockResolvedValue(true);
      mockReadJson.mockResolvedValue(existingLock);

      await expect(lockManager.acquireLock(false)).rejects.toThrow(
        GitLockedException
      );
    });

    test('should wait and acquire lock when wait is true', async () => {
      // First call: locked, second call: unlocked
      let callCount = 0;
      mockPathExists.mockImplementation(async () => {
        callCount++;
        return callCount === 1;
      });

      const lockContext = await lockManager.acquireLock(true, 5);

      expect(lockContext).toBeDefined();
      expect(mockPathExists.mock.calls.length).toBeGreaterThanOrEqual(2);
    });

    test('should timeout when waiting too long', async () => {
      mockPathExists.mockResolvedValue(true);
      mockReadJson.mockResolvedValue({
        pid: process.pid,
        timestamp: Date.now(),
        operation: 'commit',
        user: 'other-user',
        hostname: 'other-host',
        workingDir: tempDir,
      });

      await expect(lockManager.acquireLock(true, 0.05)).rejects.toThrow(
        'Lock acquisition timeout'
      );
    }, 10000);
  });

  describe('releaseLock', () => {
    test('should release lock when exists', async () => {
      mockPathExists.mockResolvedValue(true);

      await lockManager.releaseLock();

      expect(mockRemove).toHaveBeenCalled();
    });

    test('should not throw when lock does not exist', async () => {
      mockPathExists.mockResolvedValue(false);

      await expect(lockManager.releaseLock()).resolves.toBeUndefined();
    });

    test('should handle removal errors gracefully', async () => {
      mockPathExists.mockResolvedValue(true);
      mockRemove.mockRejectedValue(new Error('Permission denied'));

      // Should not throw
      await expect(lockManager.releaseLock()).resolves.toBeUndefined();
    });
  });

  describe('withLock', () => {
    test('should execute operation with lock', async () => {
      mockPathExists.mockResolvedValue(false);

      const operation = vi.fn().mockResolvedValue('result');
      const releaseLockSpy = vi.spyOn(lockManager, 'releaseLock');

      const result = await lockManager.withLock(operation, 'test-operation');

      expect(result).toBe('result');
      expect(operation).toHaveBeenCalled();
      expect(mockWriteJson).toHaveBeenCalled();
      expect(releaseLockSpy).toHaveBeenCalled();
    });

    test('should release lock even if operation fails', async () => {
      mockPathExists.mockResolvedValue(false);

      const operation = vi
        .fn()
        .mockRejectedValue(new Error('Operation failed'));
      const releaseLockSpy = vi.spyOn(lockManager, 'releaseLock');

      await expect(
        lockManager.withLock(operation, 'test-operation')
      ).rejects.toThrow('Operation failed');

      expect(releaseLockSpy).toHaveBeenCalled(); // Lock should still be released
    });

    test('should use default operation name', async () => {
      mockPathExists.mockResolvedValue(false);

      const operation = vi.fn().mockResolvedValue('result');

      await lockManager.withLock(operation);

      expect(operation).toHaveBeenCalled();
    });
  });

  describe('getLockStatus', () => {
    test('should return complete lock status', async () => {
      const lockInfo = {
        pid: process.pid,
        timestamp: Date.now() - 5000, // 5 seconds ago
        operation: 'commit',
        user: 'test-user',
        hostname: 'test-host',
        workingDir: tempDir,
      };

      mockPathExists.mockResolvedValue(true);
      mockReadJson.mockResolvedValue(lockInfo);

      const status = await lockManager.getLockStatus();

      expect(status.isLocked).toBe(true);
      expect(status.lockFileExists).toBe(true);
      expect(status.lockInfo).toEqual(lockInfo);
      expect(status.processRunning).toBe(true);
      expect(status.lockAgeSeconds).toBeGreaterThanOrEqual(5);
    });

    test('should return status when no lock exists', async () => {
      mockPathExists.mockResolvedValue(false);

      const status = await lockManager.getLockStatus();

      expect(status).toMatchObject({
        isLocked: false,
        lockFileExists: false,
        lockInfo: null,
      });
    });

    test('should handle corrupt lock file in status check', async () => {
      mockPathExists.mockResolvedValue(true);
      mockReadJson.mockRejectedValue(new Error('Invalid JSON'));

      const status = await lockManager.getLockStatus();

      expect(status.isLocked).toBe(false);
      expect(status.lockFileExists).toBe(true);
      expect(status.lockInfo).toBeNull();
    });
  });

  describe('cleanupStaleLocks', () => {
    test('should remove stale lock from dead process', async () => {
      const staleLock = {
        pid: 99999,
        timestamp: Date.now() - 60000, // 1 minute ago
        operation: 'commit',
        user: 'test-user',
        hostname: 'test-host',
        workingDir: tempDir,
      };

      mockPathExists.mockResolvedValue(true);
      mockReadJson.mockResolvedValue(staleLock);

      await lockManager.cleanupStaleLocks();

      expect(mockRemove).toHaveBeenCalled();
    });

    test('should remove expired lock', async () => {
      const expiredLock = {
        pid: process.pid, // Current process but lock is too old
        timestamp: Date.now() - 400000, // 6.67 minutes ago (> 5 minute max)
        operation: 'commit',
        user: 'test-user',
        hostname: 'test-host',
        workingDir: tempDir,
      };

      mockPathExists.mockResolvedValue(true);
      mockReadJson.mockResolvedValue(expiredLock);

      await lockManager.cleanupStaleLocks();

      expect(mockRemove).toHaveBeenCalled();
    });

    test('should preserve active locks', async () => {
      const activeLock = {
        pid: process.pid,
        timestamp: Date.now(),
        operation: 'commit',
        user: 'test-user',
        hostname: 'test-host',
        workingDir: tempDir,
      };

      mockPathExists.mockResolvedValue(true);
      mockReadJson.mockResolvedValue(activeLock);

      await lockManager.cleanupStaleLocks();

      expect(mockRemove).not.toHaveBeenCalled();
    });

    test('should handle missing lock file', async () => {
      mockPathExists.mockResolvedValue(false);

      await expect(lockManager.cleanupStaleLocks()).resolves.toBeUndefined();
      expect(mockRemove).not.toHaveBeenCalled();
    });

    test('should handle corrupt lock during cleanup', async () => {
      mockPathExists.mockResolvedValue(true);
      mockReadJson.mockRejectedValue(new Error('Invalid JSON'));

      await expect(lockManager.cleanupStaleLocks()).resolves.toBeUndefined();
      // Should still attempt to clean up corrupt lock
      expect(mockRemove).toHaveBeenCalled();
    });
  });
});

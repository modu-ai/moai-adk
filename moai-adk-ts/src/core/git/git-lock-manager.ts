/**
 * @file Git Lock Manager for preventing concurrent Git operations
 * @author MoAI Team
 * @tags @FEATURE:GIT-LOCK-001 @REQ:CORE-SYSTEM-013
 */

import * as os from 'node:os';
import * as path from 'node:path';
import * as fs from 'fs-extra';
import {
  type GitLockContext,
  GitLockedException,
  type GitLockInfo,
  type GitLockStatus,
} from '../../types/git';

/**
 * GitLockManager class for preventing concurrent Git operations
 * Ports Python git_lock_manager.py with TypeScript improvements
 * @tags @TASK:GIT-LOCK-MANAGER-001
 */
export class GitLockManager {
  private readonly projectDir: string;
  private readonly lockDir: string;
  private readonly lockFile: string;
  private readonly maxLockAge: number = 300000; // 5 minutes
  private readonly pollInterval: number = 100; // 100ms

  constructor(projectDir?: string) {
    if (projectDir && typeof projectDir !== 'string') {
      throw new Error('Invalid project directory');
    }

    this.projectDir = projectDir || process.cwd();
    this.lockDir = path.join(this.projectDir, '.moai', 'locks');
    this.lockFile = path.join(this.lockDir, 'git.lock');
  }

  /**
   * Check if Git operations are currently locked
   * @returns True if locked, false otherwise
   * @tags @API:IS-LOCKED-001
   */
  public async isLocked(): Promise<boolean> {
    try {
      if (!(await fs.pathExists(this.lockFile))) {
        return false;
      }

      const lockInfo = await this.readLockInfo();
      if (!lockInfo) {
        await this.cleanupCorruptLock();
        return false;
      }

      // Check if process is still running
      if (!this.isProcessRunning(lockInfo.pid)) {
        console.log(`Cleaning up stale lock from process ${lockInfo.pid}`);
        await this.releaseLock();
        return false;
      }

      // Check if lock is too old
      const lockAge = Date.now() - lockInfo.timestamp;
      if (lockAge > this.maxLockAge) {
        console.log(`Cleaning up expired lock (age: ${lockAge}ms)`);
        await this.releaseLock();
        return false;
      }

      return true;
    } catch (error) {
      console.error('Error checking lock status:', error);
      // Clean up corrupt lock file
      await this.cleanupCorruptLock();
      return false;
    }
  }

  /**
   * Acquire lock with context manager pattern
   * @param wait Whether to wait for lock
   * @param timeout Timeout in seconds
   * @returns Lock context
   * @tags @API:ACQUIRE-LOCK-001
   */
  public async acquireLock(
    wait: boolean = true,
    timeout: number = 30
  ): Promise<GitLockContext> {
    const startTime = Date.now();
    const timeoutMs = timeout * 1000;

    while (true) {
      if (!(await this.isLocked())) {
        // Create lock
        const lockInfo = await this.createLock('unknown');
        return {
          lockInfo,
          acquired: new Date(),
          release: () => this.releaseLock(),
        };
      }

      if (!wait) {
        const existingLock = await this.readLockInfo();
        throw new GitLockedException(
          'Git operations are locked by another process',
          existingLock || undefined
        );
      }

      // Check timeout
      if (Date.now() - startTime > timeoutMs) {
        throw new GitLockedException(
          'Lock acquisition timeout',
          undefined,
          timeout
        );
      }

      // Wait before retrying
      await this.sleep(this.pollInterval);
    }
  }

  /**
   * Release the Git lock
   * @tags @API:RELEASE-LOCK-001
   */
  public async releaseLock(): Promise<void> {
    try {
      if (await fs.pathExists(this.lockFile)) {
        await fs.remove(this.lockFile);
        console.log('Git lock released');
      }
    } catch (error) {
      console.error('Error releasing lock:', error);
      // Don't throw - we want to be resilient
    }
  }

  /**
   * Execute operation with lock (context manager pattern)
   * @param operation Operation to execute
   * @param operationName Name of the operation for logging
   * @param wait Whether to wait for lock
   * @param timeout Timeout in seconds
   * @returns Operation result
   * @tags @API:WITH-LOCK-001
   */
  public async withLock<T>(
    operation: () => Promise<T>,
    operationName: string = 'unknown',
    wait: boolean = true,
    timeout: number = 30
  ): Promise<T> {
    const lockContext = await this.acquireLock(wait, timeout);

    try {
      console.log(`Executing Git operation: ${operationName}`);
      return await operation();
    } finally {
      await lockContext.release();
    }
  }

  /**
   * Get comprehensive lock status
   * @returns Lock status information
   * @tags @API:GET-LOCK-STATUS-001
   */
  public async getLockStatus(): Promise<GitLockStatus> {
    const lockFileExists = await fs.pathExists(this.lockFile);

    if (!lockFileExists) {
      return {
        isLocked: false,
        lockFileExists: false,
        lockInfo: null,
      };
    }

    const lockInfo = await this.readLockInfo();
    if (!lockInfo) {
      return {
        isLocked: false,
        lockFileExists: true,
        lockInfo: null,
      };
    }

    const processRunning = this.isProcessRunning(lockInfo.pid);
    const lockAgeSeconds = (Date.now() - lockInfo.timestamp) / 1000;

    return {
      isLocked: processRunning && lockAgeSeconds < this.maxLockAge / 1000,
      lockFileExists: true,
      lockInfo,
      processRunning,
      lockAgeSeconds,
    };
  }

  /**
   * Clean up stale locks
   * @tags @API:CLEANUP-STALE-LOCKS-001
   */
  public async cleanupStaleLocks(): Promise<void> {
    if (!(await fs.pathExists(this.lockFile))) {
      return;
    }

    const lockInfo = await this.readLockInfo();
    if (!lockInfo) {
      await this.cleanupCorruptLock();
      return;
    }

    // Check if process is still running
    if (!this.isProcessRunning(lockInfo.pid)) {
      console.log(`Cleaning up stale lock from dead process ${lockInfo.pid}`);
      await this.releaseLock();
      return;
    }

    // Check if lock is too old
    const lockAge = Date.now() - lockInfo.timestamp;
    if (lockAge > this.maxLockAge) {
      console.log(`Cleaning up expired lock (age: ${lockAge}ms)`);
      await this.releaseLock();
      return;
    }

    console.log('Lock is active and valid');
  }

  // Private helper methods

  /**
   * Create lock file with current process information
   */
  private async createLock(
    operation: string = 'unknown'
  ): Promise<GitLockInfo> {
    await fs.ensureDir(this.lockDir);

    const lockInfo: GitLockInfo = {
      pid: process.pid,
      timestamp: Date.now(),
      operation,
      user: os.userInfo().username,
      hostname: os.hostname(),
      workingDir: this.projectDir,
    };

    await fs.writeJson(this.lockFile, lockInfo, { spaces: 2 });
    console.log(
      `Git lock acquired by process ${lockInfo.pid} for operation: ${operation}`
    );

    return lockInfo;
  }

  /**
   * Read lock information from file
   */
  private async readLockInfo(): Promise<GitLockInfo | null> {
    try {
      if (!(await fs.pathExists(this.lockFile))) {
        return null;
      }

      const lockInfo = (await fs.readJson(this.lockFile)) as GitLockInfo;

      // Validate lock info structure
      if (!lockInfo.pid || !lockInfo.timestamp || !lockInfo.operation) {
        throw new Error('Invalid lock file structure');
      }

      return lockInfo;
    } catch (error) {
      console.error('Error reading lock file:', error);
      return null;
    }
  }

  /**
   * Check if a process is still running
   */
  private isProcessRunning(pid: number): boolean {
    try {
      // On Unix systems, sending signal 0 checks if process exists
      process.kill(pid, 0);
      return true;
    } catch (_error) {
      // Process not found or permission denied
      return false;
    }
  }

  /**
   * Clean up corrupt lock file
   */
  private async cleanupCorruptLock(): Promise<void> {
    try {
      if (await fs.pathExists(this.lockFile)) {
        await fs.remove(this.lockFile);
        console.log('Cleaned up corrupt lock file');
      }
    } catch (error) {
      console.error('Error cleaning up corrupt lock:', error);
    }
  }

  /**
   * Sleep for specified milliseconds
   */
  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

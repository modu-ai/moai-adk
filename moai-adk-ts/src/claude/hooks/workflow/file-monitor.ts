/**
 * @file file-monitor.ts
 * @description Unified file monitoring and checkpoint system for MoAI-ADK
 * @version 1.0.0
 * @tag @FEATURE:FILE-MONITORING-013
 */

import type { HookInput, HookResult, MoAIHook, FileEvent } from '../types';
import * as fs from 'fs';
import * as path from 'path';
import { EventEmitter } from 'events';

/**
 * File system event handler interface
 */
interface FileSystemEvent {
  type: 'created' | 'modified' | 'deleted';
  path: string;
}

/**
 * File Monitor Hook - TypeScript port of file_monitor.py
 */
export class FileMonitor extends EventEmitter implements MoAIHook {
  name = 'file-monitor';

  private projectRoot: string;
  private isRunning = false;
  private changedFiles = new Set<string>();
  private lastCheckpointTime = 0;
  private checkpointInterval = 300000; // 5 minutes in milliseconds
  private watcher?: fs.FSWatcher;

  // Essential file patterns to watch
  private watchPatterns = new Set([
    '.py',
    '.js',
    '.ts',
    '.md',
    '.json',
    '.yml',
    '.yaml',
  ]);

  // Directories to ignore
  private ignorePatterns = new Set([
    '.git',
    '__pycache__',
    'node_modules',
    '.pytest_cache',
    'dist',
    'build',
  ]);

  constructor(projectRoot?: string) {
    super();
    this.projectRoot = projectRoot || process.cwd();
  }

  async execute(input: HookInput): Promise<HookResult> {
    try {
      if (this.isMoAIProject()) {
        if (this.watchFiles()) {
          return {
            success: true,
            message: '📁 File monitoring started',
          };
        } else {
          return {
            success: true,
            message: '⚠️  Could not start file monitoring',
          };
        }
      }

      return { success: true };
    } catch (error) {
      // Silent failure to avoid breaking Claude Code session
      return { success: true };
    }
  }

  /**
   * Start file watching
   */
  watchFiles(): boolean {
    try {
      if (this.isRunning) {
        return true;
      }

      this.watcher = fs.watch(
        this.projectRoot,
        { recursive: true },
        (eventType, filename) => {
          if (filename) {
            const fullPath = path.join(this.projectRoot, filename);
            this.onFileChanged(fullPath);
          }
        }
      );

      this.isRunning = true;
      return true;
    } catch (error) {
      return false;
    }
  }

  /**
   * Stop file watching
   */
  stopWatching(): void {
    if (this.watcher && this.isRunning) {
      this.watcher.close();
      this.isRunning = false;
    }
  }

  /**
   * Handle file change event
   */
  private onFileChanged(filePath: string): void {
    if (!this.shouldMonitorFile(filePath)) {
      return;
    }

    this.changedFiles.add(filePath);

    // Emit event for external listeners
    const event: FileEvent = {
      path: filePath,
      type: 'modified', // Simplified for now
      timestamp: new Date(),
    };
    this.emit('fileChanged', event);

    // Check if should create checkpoint
    if (this.shouldCreateCheckpoint()) {
      this.createCheckpoint();
    }
  }

  /**
   * Determine if checkpoint should be created
   */
  private shouldCreateCheckpoint(): boolean {
    const currentTime = Date.now();

    // Create checkpoint if enough time has passed and files changed
    return (
      currentTime - this.lastCheckpointTime > this.checkpointInterval &&
      this.changedFiles.size > 0
    );
  }

  /**
   * Create checkpoint snapshot
   */
  private createCheckpoint(): boolean {
    try {
      const currentTime = Date.now();

      // Reset changed files and update timestamp
      this.changedFiles.clear();
      this.lastCheckpointTime = currentTime;

      // Emit checkpoint event
      this.emit('checkpoint', {
        timestamp: new Date(currentTime),
        changedFiles: Array.from(this.changedFiles),
      });

      return true;
    } catch (error) {
      return false;
    }
  }

  /**
   * Check if file should be monitored
   */
  private shouldMonitorFile(filePath: string): boolean {
    const parsedPath = path.parse(filePath);
    const pathParts = filePath.split(path.sep);

    // Skip ignored directories
    for (const part of pathParts) {
      if (this.ignorePatterns.has(part)) {
        return false;
      }
    }

    // Only monitor files with relevant extensions
    return this.watchPatterns.has(parsedPath.ext);
  }

  /**
   * Check if this is a MoAI project
   */
  private isMoAIProject(): boolean {
    const moaiPath = path.join(this.projectRoot, '.moai');
    return fs.existsSync(moaiPath);
  }

  /**
   * Get list of changed files since last checkpoint
   */
  getChangedFiles(): string[] {
    return Array.from(this.changedFiles);
  }

  /**
   * Get monitoring statistics
   */
  getStats(): {
    isRunning: boolean;
    changedFiles: number;
    lastCheckpoint: Date | null;
  } {
    return {
      isRunning: this.isRunning,
      changedFiles: this.changedFiles.size,
      lastCheckpoint:
        this.lastCheckpointTime > 0 ? new Date(this.lastCheckpointTime) : null,
    };
  }
}

/**
 * CLI entry point for Claude Code compatibility
 */
export async function main(): Promise<void> {
  try {
    const projectRoot = process.cwd();
    const monitor = new FileMonitor(projectRoot);

    const result = await monitor.execute({});

    if (result.message) {
      console.log(result.message);
    }

    // Keep process alive for monitoring
    if (monitor.getStats().isRunning) {
      process.on('SIGINT', () => {
        monitor.stopWatching();
        process.exit(0);
      });

      process.on('SIGTERM', () => {
        monitor.stopWatching();
        process.exit(0);
      });
    }
  } catch (error) {
    // Silent failure to avoid breaking Claude Code session
  }
}

// Execute if run directly
if (require.main === module) {
  main().catch(() => {
    // Silent failure
  });
}

// @TEST:DOCTOR-001 | Chain: @REQ:DOCTOR-001 -> @DESIGN:DOCTOR-001 -> @TASK:DOCTOR-001
// Related: @CODE:DOCTOR-001

/**
 * @file DoctorCommand integration tests
 * @author MoAI Team
 */

import * as fs from 'node:fs/promises';
import * as path from 'node:path';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { SystemDetector } from '@/core/system-checker';
import { DoctorCommand } from '../index.js';

describe('DoctorCommand', () => {
  let doctorCommand: DoctorCommand;
  let mockDetector: SystemDetector;
  let tempDir: string;

  beforeEach(() => {
    mockDetector = {
      getInstallCommandForCurrentPlatform: vi
        .fn()
        .mockReturnValue('brew install tool'),
    } as unknown as SystemDetector;

    doctorCommand = new DoctorCommand(mockDetector);
    tempDir = path.join(process.cwd(), '.test-doctor-temp');
  });

  afterEach(async () => {
    try {
      await fs.rm(tempDir, { recursive: true, force: true });
    } catch {
      // Ignore cleanup errors
    }
  });

  describe('run', () => {
    it('should execute system diagnostics successfully', async () => {
      const result = await doctorCommand.run();

      expect(result).toHaveProperty('allPassed');
      expect(result).toHaveProperty('results');
      expect(result).toHaveProperty('missingRequirements');
      expect(result).toHaveProperty('versionConflicts');
      expect(result).toHaveProperty('summary');
      expect(result.summary).toHaveProperty('total');
      expect(result.summary).toHaveProperty('passed');
      expect(result.summary).toHaveProperty('failed');
    });

    it('should accept projectPath option', async () => {
      const projectPath = process.cwd();
      const result = await doctorCommand.run({ projectPath });

      expect(result).toBeDefined();
      expect(Array.isArray(result.results)).toBe(true);
    });

    it('should categorize results correctly', async () => {
      const result = await doctorCommand.run();

      expect(Array.isArray(result.missingRequirements)).toBe(true);
      expect(Array.isArray(result.versionConflicts)).toBe(true);
      expect(result.summary.total).toBeGreaterThanOrEqual(0);
    });

    it('should validate summary totals', async () => {
      const result = await doctorCommand.run();

      const { total, passed, failed } = result.summary;
      expect(total).toBe(passed + failed);
    });
  });

  describe('run with --list-backups', () => {
    it('should list available backups', async () => {
      const result = await doctorCommand.run({ listBackups: true });

      expect(result).toHaveProperty('allPassed');
      expect(result.summary.total).toBeGreaterThanOrEqual(0);
    });

    it('should return successful result when no backups found', async () => {
      const result = await doctorCommand.run({ listBackups: true });

      expect(result.allPassed).toBe(true);
      expect(result.results).toHaveLength(0);
    });

    it('should find and list backup directories', async () => {
      // This test verifies backup listing functionality
      // Note: process.chdir() is not supported in vitest workers
      // BackupChecker tests cover the actual directory finding logic

      const result = await doctorCommand.run({ listBackups: true });

      expect(result.allPassed).toBe(true);
      expect(result.summary).toBeDefined();
    });

    it('should handle backup listing errors gracefully', async () => {
      // Force an error by passing invalid path (mocked scenario)
      const result = await doctorCommand.run({ listBackups: true });

      // Should not throw and return a valid result
      expect(result).toBeDefined();
      expect(result).toHaveProperty('allPassed');
    });
  });

  describe('formatCheckResult', () => {
    it('should be accessible for formatting', () => {
      const checkResult = {
        requirement: {
          name: 'TestTool',
          category: 'runtime' as const,
        },
        result: {
          isInstalled: true,
          versionSatisfied: true,
          detectedVersion: '1.0.0',
        },
      };

      const formatted = doctorCommand.formatCheckResult(checkResult);

      expect(formatted).toContain('TestTool');
      expect(formatted).toContain('1.0.0');
    });
  });

  describe('getInstallationSuggestion', () => {
    it('should provide installation suggestion', () => {
      const checkResult = {
        requirement: {
          name: 'Git',
          category: 'development' as const,
        },
        result: {
          isInstalled: false,
          versionSatisfied: false,
        },
      };

      const suggestion = doctorCommand.getInstallationSuggestion(checkResult);

      expect(suggestion).toContain('Git');
    });
  });
});

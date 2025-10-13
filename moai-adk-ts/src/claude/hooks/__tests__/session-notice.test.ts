/**
 * @TEST:SESSION-001 |
 * Related: @CODE:HOOK-003, @CODE:HOOK-003:API
 *
 * Session Notice Hook Test Suite
 * 세션 시작 알림 및 프로젝트 상태 표시 테스트
 */

import { spawn } from 'node:child_process';
import * as fs from 'node:fs';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { SessionNotifier } from '../session-notice/index';

// Mock modules
vi.mock('node:fs', () => ({
  existsSync: vi.fn(),
  readFileSync: vi.fn(),
  readdirSync: vi.fn(),
  statSync: vi.fn(),
}));
vi.mock('node:child_process', () => ({
  spawn: vi.fn(),
}));

describe('SessionNotifier Hook', () => {
  let notifier: SessionNotifier;
  const testProjectRoot = '/test/project';

  beforeEach(() => {
    vi.clearAllMocks();
    notifier = new SessionNotifier(testProjectRoot);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('@TEST:SESSION-001-HAPPY: MoAI 프로젝트 감지', () => {
    it('should detect MoAI project successfully', () => {
      vi.spyOn(fs, 'existsSync').mockImplementation((p: any) => {
        const pathStr = p.toString();
        return pathStr.includes('.moai') || pathStr.includes('.claude');
      });

      const result = notifier.isMoAIProject();

      expect(result).toBe(true);
    });

    it('should return false for non-MoAI project', () => {
      vi.spyOn(fs, 'existsSync').mockReturnValue(false);

      const result = notifier.isMoAIProject();

      expect(result).toBe(false);
    });

    it('should execute successfully for MoAI project', async () => {
      vi.spyOn(fs, 'existsSync').mockReturnValue(true);
      vi.spyOn(fs, 'readFileSync').mockReturnValue(
        JSON.stringify({
          project: { version: '0.0.1' },
          pipeline: { current_stage: 'implementation' },
        })
      );
      vi.spyOn(fs, 'readdirSync').mockReturnValue([]);

      const result = await notifier.execute();

      expect(result.success).toBe(true);
      expect(result.message).toBeDefined();
    });

    it('should suggest initialization for non-MoAI project', async () => {
      vi.spyOn(fs, 'existsSync').mockReturnValue(false);

      const result = await notifier.execute();

      expect(result.success).toBe(true);
      expect(result.message).toContain('/alfred:0-project');
    });
  });

  describe('@TEST:SESSION-001-STATUS: 프로젝트 상태 수집', () => {
    beforeEach(() => {
      vi.spyOn(fs, 'existsSync').mockReturnValue(true);
    });

    it('should get MoAI version from package.json', () => {
      vi.spyOn(fs, 'readFileSync').mockReturnValue(
        JSON.stringify({
          version: '1.2.3',
        })
      );

      const version = notifier.getMoAIVersion();

      expect(version).toBe('1.2.3');
    });

    it('should fall back to config.json for version', () => {
      vi.spyOn(fs, 'existsSync').mockImplementation((p: any) => {
        return p.toString().includes('config.json');
      });
      vi.spyOn(fs, 'readFileSync').mockReturnValue(
        JSON.stringify({
          project: { version: '0.5.0' },
        })
      );

      const version = notifier.getMoAIVersion();

      expect(version).toBe('0.5.0');
    });

    it('should return unknown for missing version', () => {
      vi.spyOn(fs, 'existsSync').mockReturnValue(false);

      const version = notifier.getMoAIVersion();

      expect(version).toBe('unknown');
    });

    it('should detect current pipeline stage', () => {
      vi.spyOn(fs, 'readFileSync').mockReturnValue(
        JSON.stringify({
          pipeline: { current_stage: 'specification' },
        })
      );

      const stage = notifier.getCurrentPipelineStage();

      expect(stage).toBe('specification');
    });

    it('should check constitution status', () => {
      // Mock isMoAIProject to return true
      vi.spyOn(fs, 'existsSync').mockImplementation((p: any) => {
        const pathStr = p.toString();
        // Return true for .moai, .claude directories and critical files
        return (
          pathStr.includes('.moai') ||
          pathStr.includes('.claude') ||
          pathStr.includes('CLAUDE.md') ||
          pathStr.includes('development-guide.md')
        );
      });

      const status = notifier.checkConstitutionStatus();

      expect(status.status).toBe('ok');
      expect(status.violations).toHaveLength(0);
    });

    it('should detect constitution violations', () => {
      // Mock isMoAIProject to return true, but some files missing
      vi.spyOn(fs, 'existsSync').mockImplementation((p: any) => {
        const pathStr = p.toString();
        // .moai and .claude exist but critical files are missing
        if (
          pathStr.includes('.moai') &&
          !pathStr.includes('development-guide')
        ) {
          return true;
        }
        if (pathStr.includes('.claude') && !pathStr.includes('CLAUDE')) {
          return true;
        }
        // Critical files don't exist
        if (
          pathStr.includes('development-guide.md') ||
          pathStr.includes('CLAUDE.md')
        ) {
          return false;
        }
        return true;
      });

      const status = notifier.checkConstitutionStatus();

      expect(status.status).toBe('violations_found');
      expect(status.violations.length).toBeGreaterThan(0);
    });
  });

  describe('@TEST:SESSION-001-SPEC: SPEC 진행률 계산', () => {
    it('should calculate SPEC progress correctly', () => {
      vi.spyOn(fs, 'existsSync').mockReturnValue(true);
      vi.spyOn(fs, 'readdirSync').mockReturnValue([
        'SPEC-001',
        'SPEC-002',
        'SPEC-003',
      ] as any);
      vi.spyOn(fs, 'statSync').mockReturnValue({
        isDirectory: () => true,
      } as any);

      const progress = notifier.getSpecProgress();

      expect(progress.total).toBe(3);
    });

    it('should return zero progress for missing specs directory', () => {
      vi.spyOn(fs, 'existsSync').mockReturnValue(false);

      const progress = notifier.getSpecProgress();

      expect(progress.total).toBe(0);
      expect(progress.completed).toBe(0);
    });

    it('should count completed specs', () => {
      vi.spyOn(fs, 'existsSync').mockImplementation((_p: any) => {
        return true;
      });
      vi.spyOn(fs, 'readdirSync').mockReturnValue(['SPEC-001'] as any);
      vi.spyOn(fs, 'statSync').mockReturnValue({
        isDirectory: () => true,
      } as any);

      const progress = notifier.getSpecProgress();

      expect(progress.completed).toBe(1);
    });
  });

  describe('@TEST:SESSION-001-GIT: Git 정보 수집', () => {
    it('should get Git branch information', async () => {
      const mockSpawn = vi.fn().mockReturnValue({
        stdout: {
          on: vi.fn((event, callback) => {
            if (event === 'data') {
              callback(Buffer.from('develop'));
            }
          }),
        },
        stderr: { on: vi.fn() },
        on: vi.fn((event, callback) => {
          if (event === 'close') {
            callback(0);
          }
        }),
      });
      (spawn as any).mockImplementation(mockSpawn as any);

      const gitInfo = await notifier.getGitInfo();

      expect(gitInfo.branch).toBe('develop');
    });

    it('should handle Git command failures', async () => {
      const mockSpawn = vi.fn().mockReturnValue({
        stdout: { on: vi.fn() },
        stderr: { on: vi.fn() },
        on: vi.fn((event, callback) => {
          if (event === 'close') {
            callback(1);
          }
        }),
      });
      (spawn as any).mockImplementation(mockSpawn as any);

      const gitInfo = await notifier.getGitInfo();

      expect(gitInfo.branch).toBe('unknown');
    });

    it('should count Git changes', async () => {
      const mockSpawn = vi.fn().mockReturnValue({
        stdout: {
          on: vi.fn((event, callback) => {
            if (event === 'data') {
              callback(Buffer.from(' M file1.ts\n M file2.ts\n'));
            }
          }),
        },
        stderr: { on: vi.fn() },
        on: vi.fn((event, callback) => {
          if (event === 'close') {
            callback(0);
          }
        }),
      });
      (spawn as any).mockImplementation(mockSpawn as any);

      const count = await notifier.getGitChangesCount();

      expect(count).toBe(2);
    });
  });

  describe('@TEST:SESSION-001-VERSION: 버전 확인', () => {
    it('should check for latest version', async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ version: '1.0.0' }),
      } as any);

      vi.spyOn(notifier, 'getMoAIVersion').mockReturnValue('0.9.0');

      const versionCheck = await notifier.checkLatestVersion();

      expect(versionCheck).toBeDefined();
      expect(versionCheck?.hasUpdate).toBe(true);
    });

    it('should handle version check failures gracefully', async () => {
      global.fetch = vi.fn().mockRejectedValue(new Error('Network error'));

      const versionCheck = await notifier.checkLatestVersion();

      expect(versionCheck).toBeNull();
    });
  });
});

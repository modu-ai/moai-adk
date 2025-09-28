/**
 * @file session-notice.test.ts
 * @description Tests for SessionNotifier hook
 */

import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';
import { SessionNotifier } from '../../../../claude/hooks/session/session-notice';
// import { HookInput } from '../../../../claude/hooks/types';
import * as fs from 'fs';
// import * as path from 'path';

// Mock filesystem
vi.mock('fs');

const mockFs = fs as vi.Mocked<typeof fs>;

describe('SessionNotifier', () => {
  let sessionNotifier: SessionNotifier;
  const mockProjectRoot = '/test/project';

  beforeEach(() => {
    sessionNotifier = new SessionNotifier(mockProjectRoot);

    // Reset all mocks
    vi.clearAllMocks();

    // Mock console.log to avoid output during tests
    vi.spyOn(console, 'log').mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('execute', () => {
    it('should show initialization message for non-MoAI project', async () => {
      mockFs.existsSync.mockReturnValue(false);

      const result = await sessionNotifier.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toBe(
        '💡 Run `/moai:0-project` to initialize MoAI-ADK'
      );
    });

    it('should show project status for MoAI project', async () => {
      // Mock MoAI project detection
      mockFs.existsSync.mockImplementation((filePath: fs.PathLike) => {
        const pathStr = filePath.toString();
        return (
          pathStr.includes('.moai') || pathStr.includes('.claude/commands/moai')
        );
      });

      mockFs.readFileSync.mockImplementation(() => '{}');
      mockFs.readdirSync.mockImplementation(() => []);

      const result = await sessionNotifier.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain('🗿 MoAI-ADK 프로젝트');
      expect(result.data).toBeDefined();
    });

    it('should handle file system errors gracefully', async () => {
      mockFs.existsSync.mockImplementation(() => {
        throw new Error('File system error');
      });

      const result = await sessionNotifier.execute({});

      expect(result.success).toBe(true);
    });
  });

  describe('getProjectStatus', () => {
    beforeEach(() => {
      // Mock MoAI project
      mockFs.existsSync.mockImplementation((filePath: fs.PathLike) => {
        const pathStr = filePath.toString();
        return (
          pathStr.includes('.moai') || pathStr.includes('.claude/commands/moai')
        );
      });
    });

    it('should return comprehensive project status', async () => {
      mockFs.readFileSync.mockImplementation(() =>
        JSON.stringify({
          project: { version: '1.0.0' },
          pipeline: { current_stage: 'implementation' },
        })
      );

      mockFs.readdirSync.mockImplementation(
        () => ['SPEC-001', 'SPEC-002'] as any
      );
      mockFs.statSync.mockImplementation(
        () => ({ isDirectory: () => true }) as any
      );

      const status = await sessionNotifier.getProjectStatus();

      expect(status.projectName).toBe('project');
      expect(status.moaiVersion).toBe('1.0.0');
      expect(status.initialized).toBe(true);
      expect(status.pipelineStage).toBe('implementation');
      expect(status.specProgress.total).toBe(2);
    });

    it('should detect constitution violations', async () => {
      mockFs.existsSync.mockImplementation((filePath: fs.PathLike) => {
        const pathStr = filePath.toString();
        // Mock missing development-guide.md
        if (pathStr.includes('development-guide.md')) {
          return false;
        }
        return (
          pathStr.includes('.moai') || pathStr.includes('.claude/commands/moai')
        );
      });

      const status = await sessionNotifier.getProjectStatus();

      expect(status.constitutionStatus.status).toBe('violations_found');
      expect(status.constitutionStatus.violations).toContain(
        'Missing critical file: .moai/memory/development-guide.md'
      );
    });

    it('should handle missing config files gracefully', async () => {
      mockFs.readFileSync.mockImplementation(() => {
        throw new Error('Config file not found');
      });

      const status = await sessionNotifier.getProjectStatus();

      expect(status.moaiVersion).toBe('unknown');
      expect(status.pipelineStage).toBe('specification'); // fallback
    });
  });

  describe('SPEC progress tracking', () => {
    beforeEach(() => {
      mockFs.existsSync.mockImplementation((filePath: fs.PathLike) => {
        const pathStr = filePath.toString();
        return (
          pathStr.includes('.moai') ||
          pathStr.includes('.claude/commands/moai') ||
          pathStr.includes('.moai/specs')
        );
      });
    });

    it('should count completed SPECs correctly', async () => {
      mockFs.readdirSync.mockImplementation((dir: string) => {
        if (dir.toString().includes('.moai/specs')) {
          return ['SPEC-001', 'SPEC-002', 'SPEC-003'] as any;
        }
        return [] as any;
      });

      mockFs.statSync.mockImplementation(
        () => ({ isDirectory: () => true }) as any
      );

      mockFs.existsSync.mockImplementation((filePath: fs.PathLike) => {
        const pathStr = filePath.toString();
        if (
          pathStr.includes('SPEC-001') &&
          (pathStr.includes('spec.md') || pathStr.includes('plan.md'))
        ) {
          return true;
        }
        if (pathStr.includes('SPEC-002') && pathStr.includes('spec.md')) {
          return true; // Only spec.md, missing plan.md
        }
        return (
          pathStr.includes('.moai') ||
          pathStr.includes('.claude/commands/moai') ||
          pathStr.includes('.moai/specs')
        );
      });

      const status = await sessionNotifier.getProjectStatus();

      expect(status.specProgress.total).toBe(3);
      expect(status.specProgress.completed).toBe(1); // Only SPEC-001 has both files
    });

    it('should handle missing specs directory', async () => {
      mockFs.existsSync.mockImplementation((filePath: fs.PathLike) => {
        const pathStr = filePath.toString();
        if (pathStr.includes('.moai/specs')) {
          return false;
        }
        return (
          pathStr.includes('.moai') || pathStr.includes('.claude/commands/moai')
        );
      });

      const status = await sessionNotifier.getProjectStatus();

      expect(status.specProgress.total).toBe(0);
      expect(status.specProgress.completed).toBe(0);
    });
  });

  describe('pipeline stage detection', () => {
    beforeEach(() => {
      mockFs.existsSync.mockImplementation((filePath: fs.PathLike) => {
        const pathStr = filePath.toString();
        return (
          pathStr.includes('.moai') || pathStr.includes('.claude/commands/moai')
        );
      });
    });

    it('should detect stage from config file', async () => {
      mockFs.readFileSync.mockImplementation(() =>
        JSON.stringify({
          pipeline: { current_stage: 'implementation' },
        })
      );

      const status = await sessionNotifier.getProjectStatus();

      expect(status.pipelineStage).toBe('implementation');
    });

    it('should fallback to heuristic detection', async () => {
      mockFs.readFileSync.mockImplementation(() => {
        throw new Error('Config not found');
      });

      mockFs.existsSync.mockImplementation((filePath: fs.PathLike) => {
        const pathStr = filePath.toString();
        if (pathStr.includes('.moai/specs')) {
          return true;
        }
        return (
          pathStr.includes('.moai') || pathStr.includes('.claude/commands/moai')
        );
      });

      mockFs.readdirSync.mockImplementation(() => ['SPEC-001'] as any);
      mockFs.statSync.mockImplementation(
        () => ({ isDirectory: () => true }) as any
      );

      const status = await sessionNotifier.getProjectStatus();

      expect(status.pipelineStage).toBe('implementation');
    });

    it('should detect specification stage when no specs exist', async () => {
      mockFs.readFileSync.mockImplementation(() => {
        throw new Error('Config not found');
      });

      mockFs.existsSync.mockImplementation((filePath: fs.PathLike) => {
        const pathStr = filePath.toString();
        if (pathStr.includes('.moai/specs')) {
          return false;
        }
        return (
          pathStr.includes('.moai') || pathStr.includes('.claude/commands/moai')
        );
      });

      const status = await sessionNotifier.getProjectStatus();

      expect(status.pipelineStage).toBe('specification');
    });

    it('should detect initialization stage for non-MoAI project', async () => {
      mockFs.existsSync.mockReturnValue(false);

      const status = await sessionNotifier.getProjectStatus();

      expect(status.pipelineStage).toBe('initialization');
    });
  });

  describe('Git integration', () => {
    it('should handle Git errors gracefully', async () => {
      mockFs.existsSync.mockImplementation((filePath: fs.PathLike) => {
        const pathStr = filePath.toString();
        return (
          pathStr.includes('.moai') || pathStr.includes('.claude/commands/moai')
        );
      });

      // Git commands will fail in test environment, but should be handled gracefully
      const result = await sessionNotifier.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain('🗿 MoAI-ADK 프로젝트');
    });
  });

  describe('output generation', () => {
    beforeEach(() => {
      mockFs.existsSync.mockImplementation((filePath: fs.PathLike) => {
        const pathStr = filePath.toString();
        return (
          pathStr.includes('.moai') || pathStr.includes('.claude/commands/moai')
        );
      });
    });

    it('should include violations in output when detected', async () => {
      mockFs.existsSync.mockImplementation((filePath: fs.PathLike) => {
        const pathStr = filePath.toString();
        if (pathStr.includes('CLAUDE.md')) {
          return false; // Missing critical file
        }
        return (
          pathStr.includes('.moai') || pathStr.includes('.claude/commands/moai')
        );
      });

      const result = await sessionNotifier.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain(
        '⚠️  Development guide violations detected'
      );
      expect(result.message).toContain('Missing critical file: CLAUDE.md');
    });

    it('should include SPEC progress information', async () => {
      mockFs.readdirSync.mockImplementation(
        () => ['SPEC-001', 'SPEC-002'] as any
      );
      mockFs.statSync.mockImplementation(
        () => ({ isDirectory: () => true }) as any
      );

      const result = await sessionNotifier.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain('📝 SPEC 진행률');
      expect(result.message).toContain('(미완료 2개)');
    });

    it('should include system status confirmation', async () => {
      const result = await sessionNotifier.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain('✅ 통합 체크포인트 시스템 사용 가능');
    });
  });
});

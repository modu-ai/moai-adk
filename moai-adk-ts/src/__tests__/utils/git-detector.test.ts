// @TEST:INIT-004 | SPEC: SPEC-INIT-004.md
// Related: @CODE:INIT-004

/**
 * @file Git detection utilities tests
 * @author MoAI Team
 * @tags @TEST:INIT-004:GIT-DETECTOR
 */

import type { SimpleGit } from 'simple-git';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

// Mock node:fs module (ESM-compatible)
vi.mock('node:fs', () => ({
  existsSync: vi.fn(),
}));

// Mock simple-git module
const mockSimpleGit = vi.fn();
vi.mock('simple-git', () => ({
  simpleGit: mockSimpleGit,
}));

// Import mocked modules after mock setup
import { existsSync } from 'node:fs';

describe('@TEST:INIT-004 - GitDetector', () => {
  let mockGit: Partial<SimpleGit>;
  const testDir = '/test/project';

  beforeEach(() => {
    // Create mock Git instance
    mockGit = {
      checkIsRepo: vi.fn(),
      raw: vi.fn(),
      branch: vi.fn(),
      getRemotes: vi.fn(),
      init: vi.fn(),
    };

    // Mock simpleGit to return our mock instance
    mockSimpleGit.mockReturnValue(mockGit);

    // Mock fs.existsSync (ESM-compatible)
    (existsSync as any).mockReturnValue(true);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('detectGitStatus', () => {
    it('should detect existing Git repository with commits', async () => {
      // Arrange
      (mockGit.checkIsRepo as any).mockResolvedValue(true);
      (mockGit.raw as any).mockResolvedValue('42\n');
      (mockGit.branch as any).mockResolvedValue({
        current: 'main',
        all: [],
        branches: {},
        detached: false,
      });
      (mockGit.getRemotes as any).mockResolvedValue([
        {
          name: 'origin',
          refs: {
            fetch: 'https://github.com/user/repo.git',
            push: 'https://github.com/user/repo.git',
          },
        },
      ]);

      const { detectGitStatus } = await import('@/utils/git-detector');

      // Act
      const result = await detectGitStatus(testDir);

      // Assert
      expect(result.exists).toBe(true);
      expect(result.commits).toBe(42);
      expect(result.currentBranch).toBe('main');
      expect(result.remotes).toHaveLength(1);
      expect(result.githubUrl).toBe('https://github.com/user/repo.git');
    });

    it('should handle non-existent Git repository', async () => {
      // Arrange
      (existsSync as any).mockReturnValue(false);

      const { detectGitStatus } = await import('@/utils/git-detector');

      // Act
      const result = await detectGitStatus(testDir);

      // Assert
      expect(result.exists).toBe(false);
      expect(result.commits).toBe(0);
      expect(result.currentBranch).toBe('');
      expect(result.remotes).toEqual([]);
      expect(result.githubUrl).toBeUndefined();
    });

    it('should handle invalid Git repository (checkIsRepo returns false)', async () => {
      // Arrange
      (mockGit.checkIsRepo as any).mockResolvedValue(false);

      const { detectGitStatus } = await import('@/utils/git-detector');

      // Act
      const result = await detectGitStatus(testDir);

      // Assert
      expect(result.exists).toBe(false);
      expect(result.commits).toBe(0);
      expect(result.currentBranch).toBe('');
      expect(result.remotes).toEqual([]);
      expect(result.githubUrl).toBeUndefined();
    });

    it('should handle Git repository without commits', async () => {
      // Arrange
      (mockGit.checkIsRepo as any).mockResolvedValue(true);
      (mockGit.raw as any).mockRejectedValue(
        new Error('fatal: your current branch does not have any commits yet')
      );
      (mockGit.branch as any).mockResolvedValue({
        current: 'main',
        all: [],
        branches: {},
        detached: false,
      });
      (mockGit.getRemotes as any).mockResolvedValue([]);

      const { detectGitStatus } = await import('@/utils/git-detector');

      // Act
      const result = await detectGitStatus(testDir);

      // Assert
      expect(result.exists).toBe(true);
      expect(result.commits).toBe(0);
      expect(result.currentBranch).toBe('main');
      expect(result.remotes).toEqual([]);
    });

    it('should handle multiple remotes', async () => {
      // Arrange
      (mockGit.checkIsRepo as any).mockResolvedValue(true);
      (mockGit.raw as any).mockResolvedValue('10\n');
      (mockGit.branch as any).mockResolvedValue({
        current: 'develop',
        all: [],
        branches: {},
        detached: false,
      });
      (mockGit.getRemotes as any).mockResolvedValue([
        {
          name: 'origin',
          refs: {
            fetch: 'https://github.com/user/repo.git',
            push: 'https://github.com/user/repo.git',
          },
        },
        {
          name: 'upstream',
          refs: {
            fetch: 'https://github.com/org/repo.git',
            push: 'https://github.com/org/repo.git',
          },
        },
      ]);

      const { detectGitStatus } = await import('@/utils/git-detector');

      // Act
      const result = await detectGitStatus(testDir);

      // Assert
      expect(result.remotes).toHaveLength(2);
      expect(result.remotes[0].name).toBe('origin');
      expect(result.remotes[1].name).toBe('upstream');
    });

    it('should handle different fetch and push URLs', async () => {
      // Arrange
      (mockGit.checkIsRepo as any).mockResolvedValue(true);
      (mockGit.raw as any).mockResolvedValue('5\n');
      (mockGit.branch as any).mockResolvedValue({
        current: 'main',
        all: [],
        branches: {},
        detached: false,
      });
      (mockGit.getRemotes as any).mockResolvedValue([
        {
          name: 'origin',
          refs: {
            fetch: 'https://github.com/user/repo.git',
            push: 'git@github.com:user/repo.git',
          },
        },
      ]);

      const { detectGitStatus } = await import('@/utils/git-detector');

      // Act
      const result = await detectGitStatus(testDir);

      // Assert
      expect(result.remotes).toHaveLength(2);
      expect(result.remotes[0].type).toBe('fetch');
      expect(result.remotes[0].url).toBe('https://github.com/user/repo.git');
      expect(result.remotes[1].type).toBe('push');
      expect(result.remotes[1].url).toBe('git@github.com:user/repo.git');
    });

    it('should handle Git error during status detection', async () => {
      // Arrange
      (mockGit.checkIsRepo as any).mockRejectedValue(
        new Error('Git operation failed')
      );

      const { detectGitStatus } = await import('@/utils/git-detector');

      // Act
      const result = await detectGitStatus(testDir);

      // Assert
      expect(result.exists).toBe(false);
      expect(result.commits).toBe(0);
      expect(result.currentBranch).toBe('');
      expect(result.remotes).toEqual([]);
    });
  });

  describe('detectGitHubRemote', () => {
    it('should detect HTTPS GitHub URL', async () => {
      const { detectGitHubRemote } = await import('@/utils/git-detector');

      const remotes = [
        {
          name: 'origin',
          url: 'https://github.com/user/repo.git',
          type: 'fetch' as const,
        },
      ];

      const result = detectGitHubRemote(remotes);

      expect(result).toBe('https://github.com/user/repo.git');
    });

    it('should detect SSH GitHub URL', async () => {
      const { detectGitHubRemote } = await import('@/utils/git-detector');

      const remotes = [
        {
          name: 'origin',
          url: 'git@github.com:user/repo.git',
          type: 'fetch' as const,
        },
      ];

      const result = detectGitHubRemote(remotes);

      expect(result).toBe('git@github.com:user/repo.git');
    });

    it('should prioritize origin remote', async () => {
      const { detectGitHubRemote } = await import('@/utils/git-detector');

      const remotes = [
        {
          name: 'upstream',
          url: 'https://github.com/org/repo.git',
          type: 'fetch' as const,
        },
        {
          name: 'origin',
          url: 'https://github.com/user/repo.git',
          type: 'fetch' as const,
        },
      ];

      const result = detectGitHubRemote(remotes);

      expect(result).toBe('https://github.com/user/repo.git');
    });

    it('should return null for non-GitHub remotes', async () => {
      const { detectGitHubRemote } = await import('@/utils/git-detector');

      const remotes = [
        {
          name: 'origin',
          url: 'https://gitlab.com/user/repo.git',
          type: 'fetch' as const,
        },
      ];

      const result = detectGitHubRemote(remotes);

      expect(result).toBeNull();
    });

    it('should return null for empty remotes', async () => {
      const { detectGitHubRemote } = await import('@/utils/git-detector');

      const result = detectGitHubRemote([]);

      expect(result).toBeNull();
    });

    it('should handle remotes without github.com', async () => {
      const { detectGitHubRemote } = await import('@/utils/git-detector');

      const remotes = [
        {
          name: 'origin',
          url: 'https://bitbucket.org/user/repo.git',
          type: 'fetch' as const,
        },
        {
          name: 'github',
          url: 'https://github.com/user/fork.git',
          type: 'fetch' as const,
        },
      ];

      const result = detectGitHubRemote(remotes);

      expect(result).toBe('https://github.com/user/fork.git');
    });
  });

  describe('autoInitGit', () => {
    it('should initialize Git repository', async () => {
      // Arrange
      (mockGit.init as any).mockResolvedValue(undefined);

      const { autoInitGit } = await import('@/utils/git-detector');

      // Act
      await autoInitGit(testDir);

      // Assert
      expect(mockGit.init).toHaveBeenCalled();
    });

    it('should throw error if Git init fails', async () => {
      // Arrange
      const error = new Error('Permission denied');
      (mockGit.init as any).mockRejectedValue(error);

      const { autoInitGit } = await import('@/utils/git-detector');

      // Act & Assert
      await expect(autoInitGit(testDir)).rejects.toThrow(
        'Failed to initialize Git repository: Permission denied'
      );
    });
  });

  describe('validateGitHubUrl', () => {
    it('should validate HTTPS GitHub URL', async () => {
      const { validateGitHubUrl } = await import('@/utils/git-detector');

      expect(validateGitHubUrl('https://github.com/user/repo')).toBe(true);
      expect(validateGitHubUrl('https://github.com/user/repo.git')).toBe(true);
    });

    it('should validate SSH GitHub URL', async () => {
      const { validateGitHubUrl } = await import('@/utils/git-detector');

      expect(validateGitHubUrl('git@github.com:user/repo.git')).toBe(true);
      expect(validateGitHubUrl('git@github.com:user/repo')).toBe(true);
    });

    it('should reject invalid URLs', async () => {
      const { validateGitHubUrl } = await import('@/utils/git-detector');

      expect(validateGitHubUrl('https://gitlab.com/user/repo')).toBe(false);
      expect(validateGitHubUrl('https://bitbucket.org/user/repo')).toBe(false);
      expect(validateGitHubUrl('invalid-url')).toBe(false);
      expect(validateGitHubUrl('')).toBe(false);
    });

    it('should handle GitHub enterprise URLs', async () => {
      const { validateGitHubUrl } = await import('@/utils/git-detector');

      // Should reject non-standard GitHub URLs
      expect(validateGitHubUrl('https://github.company.com/user/repo')).toBe(
        false
      );
    });
  });

  describe('Edge Cases', () => {
    it('should handle very long commit count', async () => {
      // Arrange
      (mockGit.checkIsRepo as any).mockResolvedValue(true);
      (mockGit.raw as any).mockResolvedValue('999999\n');
      (mockGit.branch as any).mockResolvedValue({
        current: 'main',
        all: [],
        branches: {},
        detached: false,
      });
      (mockGit.getRemotes as any).mockResolvedValue([]);

      const { detectGitStatus } = await import('@/utils/git-detector');

      // Act
      const result = await detectGitStatus(testDir);

      // Assert
      expect(result.commits).toBe(999999);
    });

    it('should handle detached HEAD state', async () => {
      // Arrange
      (mockGit.checkIsRepo as any).mockResolvedValue(true);
      (mockGit.raw as any).mockResolvedValue('5\n');
      (mockGit.branch as any).mockResolvedValue({
        current: 'HEAD',
        all: [],
        branches: {},
        detached: true,
      });
      (mockGit.getRemotes as any).mockResolvedValue([]);

      const { detectGitStatus } = await import('@/utils/git-detector');

      // Act
      const result = await detectGitStatus(testDir);

      // Assert
      expect(result.currentBranch).toBe('HEAD');
    });

    it('should handle special characters in branch names', async () => {
      // Arrange
      (mockGit.checkIsRepo as any).mockResolvedValue(true);
      (mockGit.raw as any).mockResolvedValue('3\n');
      (mockGit.branch as any).mockResolvedValue({
        current: 'feature/SPEC-INIT-004',
        all: [],
        branches: {},
        detached: false,
      });
      (mockGit.getRemotes as any).mockResolvedValue([]);

      const { detectGitStatus } = await import('@/utils/git-detector');

      // Act
      const result = await detectGitStatus(testDir);

      // Assert
      expect(result.currentBranch).toBe('feature/SPEC-INIT-004');
    });
  });

  describe('Performance', () => {
    it('should complete Git detection in under 100ms', async () => {
      // Arrange
      (mockGit.checkIsRepo as any).mockResolvedValue(true);
      (mockGit.raw as any).mockResolvedValue('10\n');
      (mockGit.branch as any).mockResolvedValue({
        current: 'main',
        all: [],
        branches: {},
        detached: false,
      });
      (mockGit.getRemotes as any).mockResolvedValue([]);

      const { detectGitStatus } = await import('@/utils/git-detector');

      // Act
      const start = Date.now();
      await detectGitStatus(testDir);
      const duration = Date.now() - start;

      // Assert
      expect(duration).toBeLessThan(100);
    });
  });
});

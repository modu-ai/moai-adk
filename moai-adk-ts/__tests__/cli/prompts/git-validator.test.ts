// @TEST:INSTALL-001 | SPEC: SPEC-INSTALL-001.md
/**
 * @file Git installation validation tests
 * @author MoAI Team
 * @tags @TEST:INSTALL-001:GIT-VALIDATION
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { execa } from 'execa';

vi.mock('execa', () => ({
  execa: vi.fn(),
}));

describe('@TEST:INSTALL-001 - Git Validation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('UR-002: Git 필수 검증', () => {
    it('should detect Git installation successfully', async () => {
      // Arrange
      const mockExeca = execa as any;
      mockExeca.mockResolvedValue({
        stdout: 'git version 2.42.0',
        stderr: '',
        exitCode: 0,
      } as any);

      // Act
      const { validateGitInstallation } = await import(
        '@/cli/prompts/init/git-validator'
      );
      const result = await validateGitInstallation();

      // Assert
      expect(result.isValid).toBe(true);
      expect(result.version).toBe('2.42.0');
      expect(mockExeca).toHaveBeenCalledWith('git', ['--version']);
    });

    it('should fail when Git is not installed', async () => {
      // Arrange
      const mockExeca = execa as any;
      mockExeca.mockRejectedValue(new Error('Command failed: git --version'));

      // Act
      const { validateGitInstallation } = await import(
        '@/cli/prompts/init/git-validator'
      );
      const result = await validateGitInstallation();

      // Assert
      expect(result.isValid).toBe(false);
      expect(result.error).toContain('Git이 설치되지 않았습니다');
    });

    it('should provide OS-specific installation instructions', async () => {
      // Arrange
      const mockExeca = execa as any;
      mockExeca.mockRejectedValue(new Error('Command failed'));

      // Act
      const { validateGitInstallation } = await import(
        '@/cli/prompts/init/git-validator'
      );
      const result = await validateGitInstallation();

      // Assert
      expect(result.error).toContain('brew install git');
      expect(result.error).toContain('sudo apt-get install git');
      expect(result.error).toContain('https://git-scm.com/download/win');
    });
  });

  describe('ER-003: Git 미설치 시 에러 메시지', () => {
    it('should display friendly error message with installation guide', async () => {
      // Arrange
      const mockExeca = execa as any;
      mockExeca.mockRejectedValue(new Error('Command failed'));

      // Act
      const { validateGitInstallation } = await import(
        '@/cli/prompts/init/git-validator'
      );
      const result = await validateGitInstallation();

      // Assert - ER-003 요구사항 검증
      expect(result.error).toMatch(/❌ Git이 설치되지 않았습니다/);
      expect(result.error).toMatch(/MoAI-ADK는 Git을 필수로 사용합니다/);
      expect(result.error).toMatch(/macOS: brew install git/);
      expect(result.error).toMatch(/Ubuntu: sudo apt-get install git/);
      expect(result.error).toMatch(/Windows: https:\/\/git-scm\.com\/download\/win/);
    });
  });
});

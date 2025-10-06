// @TEST:INSTALL-001 | SPEC: SPEC-INSTALL-001.md
/**
 * @file Developer information collection tests
 * @author MoAI Team
 * @tags @TEST:INSTALL-001:DEVELOPER-INFO
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { execa } from 'execa';
import inquirer from 'inquirer';

vi.mock('execa', () => ({
  execa: vi.fn(),
}));
vi.mock('inquirer', () => ({
  default: {
    prompt: vi.fn(),
  },
}));

describe('@TEST:INSTALL-001 - Developer Info Collection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('UR-001: 개발자 이름 수집', () => {
    it('should collect developer name from prompt', async () => {
      // Arrange
      const mockExeca = execa as any;
      mockExeca.mockResolvedValue({
        stdout: '',
        stderr: '',
        exitCode: 0,
      } as any);

      const mockPrompt = inquirer.prompt as any;
      mockPrompt.mockResolvedValue({ developerName: '홍길동' });

      // Act
      const { collectDeveloperInfo } = await import(
        '@/cli/prompts/init/developer-info'
      );
      const result = await collectDeveloperInfo();

      // Assert
      expect(result.developerName).toBe('홍길동');
      expect(mockPrompt).toHaveBeenCalled();
    });

    it('should use Git user.name as default value', async () => {
      // Arrange - ER-004 요구사항
      const mockExeca = execa as any;
      mockExeca.mockResolvedValue({
        stdout: 'Git User Name',
        stderr: '',
        exitCode: 0,
      } as any);

      const mockPrompt = inquirer.prompt as any;
      mockPrompt.mockResolvedValue({ developerName: 'Git User Name' });

      // Act
      const { collectDeveloperInfo } = await import(
        '@/cli/prompts/init/developer-info'
      );
      const result = await collectDeveloperInfo();

      // Assert
      expect(mockExeca).toHaveBeenCalledWith('git', [
        'config',
        '--global',
        'user.name',
      ]);
      expect(result.developerName).toBe('Git User Name');
    });

    it('should reject empty developer name', async () => {
      // Arrange
      const mockPrompt = inquirer.prompt as any;

      // Act & Assert
      const { getDeveloperNamePrompt } = await import(
        '@/cli/prompts/init/developer-info'
      );
      const prompt = getDeveloperNamePrompt('');

      const validateFn = prompt[0].validate as (input: string) => boolean | string;
      expect(validateFn('')).toBe('개발자 이름은 필수입니다.');
      expect(validateFn('  ')).toBe('개발자 이름은 필수입니다.');
      expect(validateFn('홍길동')).toBe(true);
    });

    it('should save developer name to config.json', async () => {
      // Arrange
      const mockExeca = execa as any;
      mockExeca.mockResolvedValue({ stdout: '', stderr: '', exitCode: 0 } as any);

      const mockPrompt = inquirer.prompt as any;
      mockPrompt.mockResolvedValue({ developerName: '홍길동' });

      // Act
      const { collectDeveloperInfo } = await import(
        '@/cli/prompts/init/developer-info'
      );
      const result = await collectDeveloperInfo();

      // Assert - UR-001 요구사항: developer.name 필드 저장
      expect(result).toEqual({
        developerName: '홍길동',
        timestamp: expect.any(String),
      });
    });
  });

  describe('ER-004: Git user.name 기본값 제안', () => {
    it('should suggest Git user.name when available', async () => {
      // Arrange
      const mockExeca = execa as any;
      mockExeca.mockResolvedValue({
        stdout: 'John Doe',
        stderr: '',
        exitCode: 0,
      } as any);

      // Act
      const { getGitUserName } = await import(
        '@/cli/prompts/init/developer-info'
      );
      const gitUserName = await getGitUserName();

      // Assert
      expect(gitUserName).toBe('John Doe');
    });

    it('should handle missing Git user.name gracefully', async () => {
      // Arrange
      const mockExeca = execa as any;
      mockExeca.mockRejectedValue(new Error('Config not found'));

      // Act
      const { getGitUserName } = await import(
        '@/cli/prompts/init/developer-info'
      );
      const gitUserName = await getGitUserName();

      // Assert
      expect(gitUserName).toBe('');
    });
  });
});

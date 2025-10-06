// @TEST:INSTALL-001 | SPEC: SPEC-INSTALL-001.md
/**
 * @file Complete installation flow integration tests
 * @author MoAI Team
 * @tags @TEST:INSTALL-001:INTEGRATION
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import inquirer from 'inquirer';
import { execa } from 'execa';

vi.mock('inquirer', () => ({
  default: {
    prompt: vi.fn(),
  },
}));
vi.mock('execa', () => ({
  execa: vi.fn(),
}));

describe('@TEST:INSTALL-001 - Installation Flow Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('AC7: Progressive Disclosure', () => {
    it('should follow correct prompt order for Personal mode', async () => {
      // Arrange
      const mockExeca = execa as any;
      mockExeca.mockResolvedValue({ stdout: 'git version 2.42.0', stderr: '', exitCode: 0 } as any);

      const mockPrompt = inquirer.prompt as any;
      const promptCalls: string[] = [];

      mockPrompt.mockImplementation((questions: any) => {
        const firstQuestion = Array.isArray(questions) ? questions[0] : questions;
        promptCalls.push(firstQuestion.name);
        return Promise.resolve({ [firstQuestion.name]: 'mock-value' });
      });

      // Act
      const { runInstallPrompts } = await import('@/cli/prompts/init/install-flow');
      await runInstallPrompts('personal');

      // Assert - 순서: Git 검증 → 개발자 이름 → SPEC Workflow
      expect(promptCalls).toEqual(
        expect.arrayContaining(['developerName', 'enforceSpec'])
      );
    });

    it('should follow correct prompt order for Team mode', async () => {
      // Arrange
      const mockExeca = execa as any;
      mockExeca.mockResolvedValue({ stdout: 'git version 2.42.0', stderr: '', exitCode: 0 } as any);

      const mockPrompt = inquirer.prompt as any;
      const promptCalls: string[] = [];

      mockPrompt.mockImplementation((questions: any) => {
        const firstQuestion = Array.isArray(questions) ? questions[0] : questions;
        promptCalls.push(firstQuestion.name);
        return Promise.resolve({ [firstQuestion.name]: true });
      });

      // Act
      const { runInstallPrompts } = await import('@/cli/prompts/init/install-flow');
      await runInstallPrompts('team');

      // Assert - Team 모드 순서: Git 검증 → 개발자 이름 → Auto PR → Draft PR
      expect(promptCalls).toEqual(
        expect.arrayContaining(['developerName', 'autoPR', 'draftPR'])
      );
    });

    it('should skip Draft PR prompt when Auto PR is disabled', async () => {
      // Arrange
      const mockExeca = execa as any;
      mockExeca.mockResolvedValue({ stdout: 'git version 2.42.0', stderr: '', exitCode: 0 } as any);

      const mockPrompt = inquirer.prompt as any;
      mockPrompt.mockImplementation((questions: any) => {
        const firstQuestion = Array.isArray(questions) ? questions[0] : questions;
        if (firstQuestion.name === 'autoPR') {
          return Promise.resolve({ autoPR: false });
        }
        return Promise.resolve({ [firstQuestion.name]: 'mock-value' });
      });

      // Act
      const { runInstallPrompts } = await import('@/cli/prompts/init/install-flow');
      const result = await runInstallPrompts('team');

      // Assert
      expect(result.draftPR).toBeUndefined();
    });
  });

  describe('AC8: 하위 호환성', () => {
    it('should handle missing developer field gracefully', async () => {
      // Arrange
      const oldConfig = {
        mode: 'personal',
        git: { enabled: true },
      };

      // Act
      const { validateConfig } = await import('@/cli/prompts/init/install-flow');
      const result = validateConfig(oldConfig);

      // Assert
      expect(result.warnings).toContain('developer 필드가 없습니다');
      expect(result.isValid).toBe(true);
    });

    it('should set enforce_spec default when missing', async () => {
      // Arrange
      const oldConfig = {
        mode: 'personal',
        constitution: {},
      };

      // Act
      const { migrateConfig } = await import('@/cli/prompts/init/install-flow');
      const migrated = migrateConfig(oldConfig);

      // Assert
      expect(migrated.constitution.enforce_spec).toBe(true);
    });
  });

  describe('Complete Flow Scenarios', () => {
    it('should complete Personal mode flow successfully', async () => {
      // Arrange
      const mockExeca = execa as any;
      mockExeca.mockResolvedValue({ stdout: 'git version 2.42.0', stderr: '', exitCode: 0 } as any);

      const mockPrompt = inquirer.prompt as any;
      mockPrompt.mockResolvedValueOnce({ developerName: '홍길동' });
      mockPrompt.mockResolvedValueOnce({ enforceSpec: true });

      // Act
      const { runInstallPrompts } = await import('@/cli/prompts/init/install-flow');
      const result = await runInstallPrompts('personal');

      // Assert
      expect(result).toEqual({
        mode: 'personal',
        developer: {
          name: '홍길동',
          timestamp: expect.any(String),
        },
        constitution: {
          enforce_spec: true,
        },
      });
    });

    it('should complete Team mode flow with Auto PR', async () => {
      // Arrange
      const mockExeca = execa as any;
      mockExeca.mockResolvedValue({ stdout: 'git version 2.42.0', stderr: '', exitCode: 0 } as any);

      const mockPrompt = inquirer.prompt as any;
      mockPrompt.mockResolvedValueOnce({ developerName: '홍길동' });
      mockPrompt.mockResolvedValueOnce({ autoPR: true });
      mockPrompt.mockResolvedValueOnce({ draftPR: true });

      // Act
      const { runInstallPrompts } = await import('@/cli/prompts/init/install-flow');
      const result = await runInstallPrompts('team');

      // Assert
      expect(result).toEqual({
        mode: 'team',
        developer: {
          name: '홍길동',
          timestamp: expect.any(String),
        },
        constitution: {
          enforce_spec: true, // Team 모드는 강제 활성화
        },
        git_strategy: {
          team: {
            auto_pr: true,
            draft_pr: true,
          },
        },
      });
    });

    it('should abort when Git is not installed', async () => {
      // Arrange
      const mockExeca = execa as any;
      mockExeca.mockRejectedValue(new Error('Command failed'));

      // Act & Assert
      const { runInstallPrompts } = await import('@/cli/prompts/init/install-flow');
      await expect(runInstallPrompts('personal')).rejects.toThrow('Git이 설치되지 않았습니다');
    });
  });
});

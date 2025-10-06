// @TEST:INSTALL-001 | SPEC: SPEC-INSTALL-001.md
/**
 * @file Auto PR and Draft PR configuration tests
 * @author MoAI Team
 * @tags @TEST:INSTALL-001:PR-CONFIG
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import inquirer from 'inquirer';

vi.mock('inquirer', () => ({
  default: {
    prompt: vi.fn(),
  },
}));

describe('@TEST:INSTALL-001 - PR Configuration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('ER-001: Auto PR 선택 (Team 모드)', () => {
    it('should prompt Auto PR for Team mode', async () => {
      // Arrange
      const mockPrompt = inquirer.prompt as any;
      mockPrompt.mockResolvedValue({ autoPR: true });

      // Act
      const { promptAutoPR } = await import('@/cli/prompts/init/pr-config');
      const result = await promptAutoPR('team');

      // Assert
      expect(mockPrompt).toHaveBeenCalled();
      expect(result.autoPR).toBe(true);
    });

    it('should default Auto PR to true', async () => {
      // Arrange
      const { getAutoPRPrompt } = await import('@/cli/prompts/init/pr-config');
      const prompt = getAutoPRPrompt();

      // Assert
      expect(prompt[0].default).toBe(true);
      expect(prompt[0].message).toContain('자동으로 PR을 생성할까요?');
    });

    it('should allow disabling Auto PR', async () => {
      // Arrange
      const mockPrompt = inquirer.prompt as any;
      mockPrompt.mockResolvedValue({ autoPR: false });

      // Act
      const { promptAutoPR } = await import('@/cli/prompts/init/pr-config');
      const result = await promptAutoPR('team');

      // Assert
      expect(result.autoPR).toBe(false);
    });

    it('should not prompt Auto PR for Personal mode', async () => {
      // Arrange
      const mockPrompt = inquirer.prompt as any;

      // Act
      const { promptAutoPR } = await import('@/cli/prompts/init/pr-config');
      const result = await promptAutoPR('personal');

      // Assert
      expect(mockPrompt).not.toHaveBeenCalled();
      expect(result.autoPR).toBeUndefined();
    });
  });

  describe('ER-002: Draft PR 선택 (Auto PR 활성화 시)', () => {
    it('should prompt Draft PR when Auto PR is enabled', async () => {
      // Arrange
      const mockPrompt = inquirer.prompt as any;
      mockPrompt.mockResolvedValue({ draftPR: true });

      // Act
      const { promptDraftPR } = await import('@/cli/prompts/init/pr-config');
      const result = await promptDraftPR(true);

      // Assert
      expect(mockPrompt).toHaveBeenCalled();
      expect(result.draftPR).toBe(true);
    });

    it('should default Draft PR to true', async () => {
      // Arrange
      const { getDraftPRPrompt } = await import('@/cli/prompts/init/pr-config');
      const prompt = getDraftPRPrompt();

      // Assert
      expect(prompt[0].default).toBe(true);
      expect(prompt[0].message).toContain('PR을 Draft 상태로 생성할까요?');
    });

    it('should not prompt Draft PR when Auto PR is disabled', async () => {
      // Arrange
      const mockPrompt = inquirer.prompt as any;

      // Act
      const { promptDraftPR } = await import('@/cli/prompts/init/pr-config');
      const result = await promptDraftPR(false);

      // Assert
      expect(mockPrompt).not.toHaveBeenCalled();
      expect(result.draftPR).toBeUndefined();
    });
  });

  describe('AC4: Auto PR 설정 저장', () => {
    it('should save autoPR to git_strategy.team.auto_pr', async () => {
      // Arrange
      const mockPrompt = inquirer.prompt as any;
      mockPrompt.mockResolvedValue({ autoPR: true });

      // Act
      const { buildPRConfig } = await import('@/cli/prompts/init/pr-config');
      const config = await buildPRConfig('team', true, true);

      // Assert
      expect(config).toEqual({
        git_strategy: {
          team: {
            auto_pr: true,
            draft_pr: true,
          },
        },
      });
    });

    it('should handle Auto PR disabled', async () => {
      // Act
      const { buildPRConfig } = await import('@/cli/prompts/init/pr-config');
      const config = await buildPRConfig('team', false, undefined);

      // Assert
      expect(config.git_strategy.team.auto_pr).toBe(false);
      expect(config.git_strategy.team.draft_pr).toBeUndefined();
    });
  });
});

// @TEST:INSTALL-001 | SPEC: SPEC-INSTALL-001.md
/**
 * @file SPEC Workflow selection tests
 * @author MoAI Team
 * @tags @TEST:INSTALL-001:SPEC-WORKFLOW
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import inquirer from 'inquirer';

vi.mock('inquirer', () => ({
  default: {
    prompt: vi.fn(),
  },
}));

describe('@TEST:INSTALL-001 - SPEC Workflow Selection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('ER-005: Personal 모드 SPEC Workflow 선택', () => {
    it('should prompt SPEC workflow for Personal mode', async () => {
      // Arrange
      const mockPrompt = inquirer.prompt as any;
      mockPrompt.mockResolvedValue({ enforceSpec: true });

      // Act
      const { promptSpecWorkflowPersonal } = await import(
        '@/cli/prompts/init/spec-workflow'
      );
      const result = await promptSpecWorkflowPersonal();

      // Assert
      expect(mockPrompt).toHaveBeenCalled();
      expect(result.enforceSpec).toBe(true);
    });

    it('should default to true for SPEC workflow', async () => {
      // Arrange
      const { getSpecWorkflowPersonalPrompt } = await import(
        '@/cli/prompts/init/spec-workflow'
      );
      const prompt = getSpecWorkflowPersonalPrompt();

      // Assert
      expect(prompt[0].default).toBe(true);
      expect(prompt[0].message).toContain('SPEC-First Workflow를 사용할까요?');
      expect(prompt[0].message).toContain('권장');
    });

    it('should allow disabling SPEC workflow', async () => {
      // Arrange
      const mockPrompt = inquirer.prompt as any;
      mockPrompt.mockResolvedValue({ enforceSpec: false });

      // Act
      const { promptSpecWorkflowPersonal } = await import(
        '@/cli/prompts/init/spec-workflow'
      );
      const result = await promptSpecWorkflowPersonal();

      // Assert
      expect(result.enforceSpec).toBe(false);
    });
  });

  describe('SR-002: Team 모드 SPEC 강제 활성화', () => {
    it('should automatically enable SPEC for Team mode', async () => {
      // Act
      const { getSpecConfigForTeam } = await import(
        '@/cli/prompts/init/spec-workflow'
      );
      const result = getSpecConfigForTeam();

      // Assert - Team 모드는 SPEC 필수
      expect(result.enforceSpec).toBe(true);
    });

    it('should not prompt for Team mode SPEC workflow', async () => {
      // Arrange
      const mockPrompt = inquirer.prompt as any;

      // Act
      const { promptSpecWorkflow } = await import(
        '@/cli/prompts/init/spec-workflow'
      );
      await promptSpecWorkflow('team');

      // Assert - Team 모드는 프롬프트 표시 안 함
      expect(mockPrompt).not.toHaveBeenCalled();
    });
  });

  describe('UR-003: SPEC Workflow 설정 저장', () => {
    it('should save enforceSpec to constitution.enforce_spec', async () => {
      // Arrange
      const mockPrompt = inquirer.prompt as any;
      mockPrompt.mockResolvedValue({ enforceSpec: true });

      // Act
      const { buildSpecConfig } = await import(
        '@/cli/prompts/init/spec-workflow'
      );
      const config = await buildSpecConfig('personal', true);

      // Assert
      expect(config).toEqual({
        constitution: {
          enforce_spec: true,
        },
      });
    });
  });
});

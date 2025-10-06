// @CODE:INSTALL-001 | SPEC: SPEC-INSTALL-001.md | TEST: __tests__/cli/prompts/pr-config.test.ts
/**
 * @file Auto PR and Draft PR configuration module
 * @author MoAI Team
 * @tags @CODE:INSTALL-001:PR-CONFIG
 */

import inquirer from 'inquirer';
import type { Answers } from 'inquirer';

/**
 * PR configuration
 */
export interface PRConfig {
  git_strategy: {
    team: {
      auto_pr?: boolean | undefined;
      draft_pr?: boolean | undefined;
    };
  };
}

/**
 * Auto PR result
 */
export interface AutoPRResult {
  autoPR?: boolean | undefined;
}

/**
 * Draft PR result
 */
export interface DraftPRResult {
  draftPR?: boolean | undefined;
}

/**
 * Get Auto PR prompt
 * @returns Inquirer prompt configuration
 */
export function getAutoPRPrompt(): Answers {
  return [
    {
      type: 'confirm',
      name: 'autoPR',
      message: '자동으로 PR을 생성할까요? (Auto PR)',
      default: true,
    },
  ];
}

/**
 * Get Draft PR prompt
 * @returns Inquirer prompt configuration
 */
export function getDraftPRPrompt(): Answers {
  return [
    {
      type: 'confirm',
      name: 'draftPR',
      message: 'PR을 Draft 상태로 생성할까요? (검토 후 Ready 전환)',
      default: true,
    },
  ];
}

/**
 * Prompt Auto PR for Team mode
 * @param mode Project mode
 * @returns Auto PR result (undefined for Personal mode)
 */
export async function promptAutoPR(mode: 'personal' | 'team'): Promise<AutoPRResult> {
  if (mode !== 'team') {
    return { autoPR: undefined };
  }

  const prompts = getAutoPRPrompt();
  const answers = await inquirer.prompt(prompts);
  return { autoPR: answers.autoPR };
}

/**
 * Prompt Draft PR (only when Auto PR is enabled)
 * @param autoPREnabled Whether Auto PR is enabled
 * @returns Draft PR result (undefined when Auto PR is disabled)
 */
export async function promptDraftPR(autoPREnabled: boolean): Promise<DraftPRResult> {
  if (!autoPREnabled) {
    return { draftPR: undefined };
  }

  const prompts = getDraftPRPrompt();
  const answers = await inquirer.prompt(prompts);
  return { draftPR: answers.draftPR };
}

/**
 * Build PR configuration
 * @param _mode Project mode (unused, reserved for future use)
 * @param autoPR Auto PR enabled
 * @param draftPR Draft PR enabled
 * @returns Complete PR configuration
 */
export async function buildPRConfig(
  _mode: 'personal' | 'team',
  autoPR: boolean,
  draftPR?: boolean | undefined
): Promise<PRConfig> {
  return {
    git_strategy: {
      team: {
        auto_pr: autoPR,
        draft_pr: draftPR,
      },
    },
  };
}

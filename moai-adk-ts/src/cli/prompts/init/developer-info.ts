// @CODE:INSTALL-001 | SPEC: SPEC-INSTALL-001.md | TEST: __tests__/cli/prompts/developer-info.test.ts
/**
 * @file Developer information collection module
 * @author MoAI Team
 * @tags @CODE:INSTALL-001:DEVELOPER-INFO
 */

import { execa } from 'execa';
import type { Answers } from 'inquirer';
import inquirer from 'inquirer';

/**
 * Developer information
 */
export interface DeveloperInfo {
  developerName: string;
  timestamp: string;
}

/**
 * Get Git user.name from global config
 * @returns Git user name or empty string
 */
export async function getGitUserName(): Promise<string> {
  try {
    const result = await execa('git', ['config', '--global', 'user.name']);
    return result.stdout.trim();
  } catch {
    return '';
  }
}

/**
 * Get developer name prompt
 * @param defaultName Default developer name from Git
 * @returns Inquirer prompt configuration
 */
export function getDeveloperNamePrompt(defaultName: string): Answers {
  return [
    {
      type: 'input',
      name: 'developerName',
      message: '개발자 이름을 입력해주세요:',
      default: defaultName,
      validate: (input: string) => {
        const trimmed = input.trim();
        if (trimmed === '') {
          return '개발자 이름은 필수입니다.';
        }
        return true;
      },
    },
  ];
}

/**
 * Collect developer information
 * @returns Developer info with name and timestamp
 */
export async function collectDeveloperInfo(): Promise<DeveloperInfo> {
  const gitUserName = await getGitUserName();
  const prompts = getDeveloperNamePrompt(gitUserName);
  const answers = await inquirer.prompt(prompts);

  return {
    developerName: answers.developerName,
    timestamp: new Date().toISOString(),
  };
}

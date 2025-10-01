/**
 * @file Prompt definitions and configurations
 * @author MoAI Team
 * @tags @CODE:INIT-DEFINITIONS-001 | Chain: @SPEC:INTERACTIVE-INIT-019 -> @CODE:INTERACTIVE-INIT-019
 * Related: @DOC:INTERACTIVE-INIT-019
 */

import chalk from 'chalk';
import type { QuestionCollection } from 'inquirer';
import { t } from '@/utils/i18n';
import { validateGitHubUrl, validateProjectName } from './validators';

/**
 * Locale selection prompt
 */
export function getLocalePrompt(): QuestionCollection {
  return [
    {
      type: 'list',
      name: 'locale',
      message: chalk.cyan('Choose CLI language / CLI 언어를 선택하세요:'),
      choices: [
        {
          name: chalk.white('🇰🇷 한국어') + chalk.gray(' - Korean'),
          value: 'ko',
          short: '한국어',
        },
        {
          name: chalk.white('🇺🇸 English') + chalk.gray(' - English'),
          value: 'en',
          short: 'English',
        },
      ],
      default: 'ko',
    },
  ];
}

/**
 * Project name prompt
 * @param defaultName Default project name
 */
export function getProjectNamePrompt(defaultName: string): QuestionCollection {
  return [
    {
      type: 'input',
      name: 'projectName',
      message: chalk.cyan(t('init.prompts.projectNameLabel')),
      default: defaultName,
      validate: validateProjectName,
      transformer: (input: string) => chalk.green(input),
    },
  ];
}

/**
 * Mode selection prompt (Personal/Team)
 */
export function getModePrompt(): QuestionCollection {
  return [
    {
      type: 'list',
      name: 'mode',
      message: chalk.cyan(t('init.prompts.selectMode')),
      choices: [
        {
          name:
            chalk.white(t('init.prompts.modePersonal')) +
            chalk.gray(` - ${t('init.prompts.modePersonalDesc')}`),
          value: 'personal',
          short: 'Personal',
        },
        {
          name:
            chalk.white(t('init.prompts.modeTeam')) +
            chalk.gray(`     - ${t('init.prompts.modeTeamDesc')}`),
          value: 'team',
          short: 'Team',
        },
      ],
      default: 'personal',
    },
  ];
}

/**
 * Git configuration prompt
 */
export function getGitConfigPrompt(): QuestionCollection {
  return [
    {
      type: 'confirm',
      name: 'gitEnabled',
      message: chalk.cyan(t('init.prompts.initGit')),
      default: true,
    },
  ];
}

/**
 * GitHub enabled prompt
 */
export function getGitHubEnabledPrompt(): QuestionCollection {
  return [
    {
      type: 'confirm',
      name: 'githubEnabled',
      message: chalk.cyan(t('init.prompts.useGithub')),
      default: true,
    },
  ];
}

/**
 * GitHub URL prompt
 */
export function getGitHubUrlPrompt(): QuestionCollection {
  return [
    {
      type: 'input',
      name: 'githubUrl',
      message: chalk.cyan(t('init.prompts.githubUrl')),
      default: 'https://github.com/username/project-name',
      validate: validateGitHubUrl,
      transformer: (input: string) => chalk.green(input),
    },
  ];
}

/**
 * SPEC workflow prompt
 */
export function getSpecWorkflowPrompt(): QuestionCollection {
  return [
    {
      type: 'list',
      name: 'specWorkflow',
      message: chalk.cyan(`${t('init.prompts.specWorkflow')}:`),
      choices: [
        {
          name:
            chalk.white(t('init.prompts.workflowBranch')) +
            chalk.gray(` - ${t('init.prompts.workflowBranchDesc')}`),
          value: 'branch',
          short: 'Branch',
        },
        {
          name:
            chalk.white(t('init.prompts.workflowCommit')) +
            chalk.gray(`  - ${t('init.prompts.workflowCommitDesc')}`),
          value: 'commit',
          short: 'Commit',
        },
      ],
      default: 'branch',
    },
  ];
}

/**
 * Auto-push configuration prompt
 */
export function getAutoPushPrompt(): QuestionCollection {
  return [
    {
      type: 'confirm',
      name: 'autoPush',
      message: chalk.cyan(t('init.prompts.autoPush')),
      default: true,
    },
  ];
}

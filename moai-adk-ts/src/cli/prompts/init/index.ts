/**
 * @file Interactive prompts orchestrator for moai init command
 * @author MoAI Team
 * @tags @CODE:INTERACTIVE-INIT-019 | Chain: @SPEC:INTERACTIVE-INIT-019 -> @CODE:INTERACTIVE-INIT-019 -> @TEST:INTERACTIVE-INIT-019
 * Related: @DOC:INTERACTIVE-INIT-019
 */

import inquirer from 'inquirer';
import { setLocale, t } from '@/utils/i18n';
import {
  getAutoPushPrompt,
  getGitConfigPrompt,
  getGitHubEnabledPrompt,
  getGitHubUrlPrompt,
  getLocalePrompt,
  getModePrompt,
  getProjectNamePrompt,
  getSpecWorkflowPrompt,
} from './definitions';
import type { InitAnswers, PartialInitAnswers } from './types';
import { displayStep, displaySummary, displayTip } from './ui-helpers';

/**
 * Prompt locale selection
 */
export async function promptLocale(): Promise<PartialInitAnswers> {
  displayStep(1, 4, 'Language Selection / 언어 선택');
  const answers = await inquirer.prompt(getLocalePrompt());
  setLocale(answers['locale']);
  displayTip(
    answers['locale'] === 'ko'
      ? '한국어가 선택되었습니다. 이후 모든 메시지가 한국어로 표시됩니다.'
      : 'English selected. All subsequent messages will be in English.'
  );
  return answers;
}

/**
 * Prompt basic project info
 */
export async function promptBasicInfo(
  defaultName?: string,
  isCurrentDirMode = false
): Promise<PartialInitAnswers> {
  displayStep(2, 4, t('init.prompts.projectInfo'));

  const effectiveDefaultName = isCurrentDirMode
    ? process.cwd().split('/').pop() || 'moai-project'
    : defaultName || 'moai-project';

  const answers = await inquirer.prompt(
    getProjectNamePrompt(effectiveDefaultName)
  );

  displayTip(
    isCurrentDirMode
      ? t('init.prompts.projectNameTipCurrent')
      : t('init.prompts.projectNameTipNew')
  );

  return answers;
}

/**
 * Prompt mode selection
 */
export async function promptMode(): Promise<PartialInitAnswers> {
  displayStep(3, 4, t('init.prompts.devMode'));
  const answers = await inquirer.prompt(getModePrompt());
  displayTip(
    answers['mode'] === 'personal'
      ? t('init.prompts.tipPersonal')
      : t('init.prompts.tipTeam')
  );
  return answers;
}

/**
 * Prompt Git config
 */
export async function promptGitConfig(): Promise<PartialInitAnswers> {
  displayStep(4, 4, t('init.prompts.versionControl'));
  const answers = await inquirer.prompt(getGitConfigPrompt());
  displayTip(
    answers['gitEnabled']
      ? t('init.prompts.tipGitEnabled')
      : t('init.prompts.tipGitDisabled')
  );
  return answers;
}

/**
 * Prompt GitHub config (team mode only)
 */
export async function promptGitHubConfig(
  mode: 'personal' | 'team'
): Promise<PartialInitAnswers> {
  if (mode !== 'team') return {};

  displayStep(4, 7, t('init.prompts.github'));
  const githubAnswers = await inquirer.prompt(getGitHubEnabledPrompt());

  if (!githubAnswers['githubEnabled']) {
    displayTip(t('init.prompts.tipGithubDisabled'));
    return githubAnswers;
  }

  displayStep(5, 7, t('init.prompts.githubRepo'));
  const urlAnswers = await inquirer.prompt(getGitHubUrlPrompt());
  displayTip(t('init.prompts.tipGithubUrl'));

  return { ...githubAnswers, ...urlAnswers };
}

/**
 * Prompt SPEC workflow (team mode only)
 */
export async function promptSpecWorkflow(
  mode: 'personal' | 'team',
  githubEnabled?: boolean
): Promise<PartialInitAnswers> {
  if (mode !== 'team' || !githubEnabled) {
    return { specWorkflow: 'commit' };
  }

  displayStep(6, 7, t('init.prompts.specWorkflow'));
  const answers = await inquirer.prompt(getSpecWorkflowPrompt());
  displayTip(
    answers['specWorkflow'] === 'branch'
      ? t('init.prompts.tipBranch')
      : t('init.prompts.tipCommit')
  );
  return answers;
}

/**
 * Prompt auto-push (GitHub enabled only)
 */
export async function promptAutoPush(
  githubEnabled?: boolean
): Promise<PartialInitAnswers> {
  if (!githubEnabled) return { autoPush: false };

  displayStep(7, 7, t('init.prompts.remoteSyn'));
  const answers = await inquirer.prompt(getAutoPushPrompt());
  displayTip(
    answers['autoPush']
      ? t('init.prompts.tipAutoPushEnabled')
      : t('init.prompts.tipAutoPushDisabled')
  );
  return answers;
}

/**
 * Run all prompts in sequence
 */
export async function runInteractivePrompts(
  defaultName?: string,
  isCurrentDirMode = false
): Promise<InitAnswers> {
  const locale = await promptLocale();
  const basic = await promptBasicInfo(defaultName, isCurrentDirMode);
  const mode = await promptMode();
  const git = await promptGitConfig();
  const github = await promptGitHubConfig(mode.mode as 'personal' | 'team');
  const workflow = await promptSpecWorkflow(
    mode.mode as 'personal' | 'team',
    github.githubEnabled
  );
  const push = await promptAutoPush(github.githubEnabled);

  const answers: InitAnswers = {
    locale: locale['locale'] ?? 'ko',
    projectName: basic['projectName'] as string,
    mode: mode['mode'] as 'personal' | 'team',
    gitEnabled: git['gitEnabled'] as boolean,
    githubEnabled: github['githubEnabled'] ?? undefined,
    githubUrl: github['githubUrl'] ?? undefined,
    specWorkflow: workflow['specWorkflow'] ?? undefined,
    autoPush: push['autoPush'] ?? undefined,
  };

  displaySummary(answers);
  return answers;
}

// Backwards compatibility
export const promptProjectSetup = runInteractivePrompts;

// Re-exports
export type { InitAnswers, PartialInitAnswers } from './types';
export {
  displayStep,
  displaySummary,
  displayTip,
  displayWelcomeBanner,
} from './ui-helpers';

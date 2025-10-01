/**
 * @file Interactive prompts for moai init command
 * @author MoAI Team
 * @tags @CODE:INTERACTIVE-INIT-019 | Chain: @SPEC:INTERACTIVE-INIT-019 -> @SPEC:INTERACTIVE-INIT-019 -> @CODE:INTERACTIVE-INIT-019 -> @TEST:INTERACTIVE-INIT-019
 * Related: @DOC:INTERACTIVE-INIT-019
 */

import inquirer from 'inquirer';
import chalk from 'chalk';
import { InputValidator } from '@/utils/input-validator';
import { type Locale, setLocale, t } from '../../utils/i18n.js';

/**
 * User answers from interactive prompts
 */
export interface InitAnswers {
  locale?: Locale;
  projectName: string;
  mode: 'personal' | 'team';
  gitEnabled: boolean;
  githubEnabled?: boolean;
  githubUrl?: string;
  specWorkflow?: 'commit' | 'branch';
  autoPush?: boolean;
}

/**
 * Display welcome banner for moai init
 */
export function displayWelcomeBanner(): void {
  console.log(
    chalk.gray("  Let's set up your project with a few questions...")
  );
  console.log(
    chalk.gray('  You can change these settings later in .moai/config.json\n')
  );
}

/**
 * Display step indicator
 * Note: Total is calculated dynamically based on mode
 */
function displayStep(current: number, total: number, question: string): void {
  const progress = `[${current}/${total}]`;
  console.log('\n');
  console.log(chalk.blue.bold(`❓ Question ${progress}`));
  console.log(chalk.white(`→ ${question}`));
}

/**
 * Get locale (language) selection
 * This is the FIRST prompt users see
 */
export async function promptLocale(): Promise<Partial<InitAnswers>> {
  displayStep(1, 4, 'Language Selection / 언어 선택');

  const answers = await inquirer.prompt([
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
  ]);

  // Immediately apply locale for subsequent prompts
  setLocale(answers.locale);

  if (answers.locale === 'ko') {
    displayTip(
      '한국어가 선택되었습니다. 이후 모든 메시지가 한국어로 표시됩니다.'
    );
  } else {
    displayTip('English selected. All subsequent messages will be in English.');
  }

  return answers;
}

/**
 * Display helpful tip
 */
function displayTip(tip: string): void {
  console.log(chalk.gray(`  💡 ${tip}`));
}

/**
 * Get basic project information
 */
export async function promptBasicInfo(
  defaultName?: string,
  isCurrentDirMode = false
): Promise<Partial<InitAnswers>> {
  displayStep(2, 4, t('init.prompts.projectInfo'));

  // Determine appropriate default name and tip based on mode
  let effectiveDefaultName: string;
  let tipMessage: string;

  if (isCurrentDirMode) {
    // For "moai init ." - suggest current directory name
    const cwd = process.cwd();
    const currentDirName = cwd.split('/').pop() || 'moai-project';
    effectiveDefaultName = currentDirName;
    tipMessage = t('init.prompts.projectNameTipCurrent');
  } else {
    // For "moai init project-name" - use provided name or default
    effectiveDefaultName = defaultName || 'moai-project';
    tipMessage = t('init.prompts.projectNameTipNew');
  }

  const answers = await inquirer.prompt([
    {
      type: 'input',
      name: 'projectName',
      message: chalk.cyan(t('init.prompts.projectNameLabel')),
      default: effectiveDefaultName,
      validate: (input: string) => {
        const result = InputValidator.validateProjectName(input, {
          maxLength: 50,
          allowSpaces: false,
        });
        return result.isValid ? true : result.errors.join(', ');
      },
      transformer: (input: string) => chalk.green(input),
    },
  ]);

  displayTip(tipMessage);

  return answers;
}

/**
 * Get mode selection (Personal/Team)
 */
export async function promptMode(): Promise<Partial<InitAnswers>> {
  displayStep(3, 4, t('init.prompts.devMode'));

  const answers = await inquirer.prompt([
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
  ]);

  if (answers.mode === 'personal') {
    displayTip(t('init.prompts.tipPersonal'));
  } else {
    displayTip(t('init.prompts.tipTeam'));
  }

  return answers;
}

/**
 * Get Git configuration
 */
export async function promptGitConfig(): Promise<Partial<InitAnswers>> {
  displayStep(4, 4, t('init.prompts.versionControl'));

  const answers = await inquirer.prompt([
    {
      type: 'confirm',
      name: 'gitEnabled',
      message: chalk.cyan(t('init.prompts.initGit')),
      default: true,
    },
  ]);

  if (answers.gitEnabled) {
    displayTip(t('init.prompts.tipGitEnabled'));
  } else {
    displayTip(t('init.prompts.tipGitDisabled'));
  }

  return answers;
}

/**
 * Get GitHub configuration (Team mode only)
 */
export async function promptGitHubConfig(
  mode: 'personal' | 'team'
): Promise<Partial<InitAnswers>> {
  if (mode !== 'team') {
    return {};
  }

  displayStep(4, 7, t('init.prompts.github'));

  const githubAnswers = await inquirer.prompt([
    {
      type: 'confirm',
      name: 'githubEnabled',
      message: chalk.cyan(t('init.prompts.useGithub')),
      default: true,
    },
  ]);

  if (!githubAnswers.githubEnabled) {
    displayTip(t('init.prompts.tipGithubDisabled'));
    return githubAnswers;
  }

  displayStep(5, 7, t('init.prompts.githubRepo'));

  const urlAnswers = await inquirer.prompt([
    {
      type: 'input',
      name: 'githubUrl',
      message: chalk.cyan(t('init.prompts.githubUrl')),
      default: 'https://github.com/username/project-name',
      validate: (input: string) => {
        const githubRegex = /^https:\/\/github\.com\/[\w-]+\/[\w-]+$/;
        if (!githubRegex.test(input)) {
          return 'Please enter a valid GitHub URL (https://github.com/username/repo)';
        }
        return true;
      },
      transformer: (input: string) => chalk.green(input),
    },
  ]);

  displayTip(t('init.prompts.tipGithubUrl'));

  return { ...githubAnswers, ...urlAnswers };
}

/**
 * Get SPEC workflow configuration (Team mode only)
 */
export async function promptSpecWorkflow(
  mode: 'personal' | 'team',
  githubEnabled?: boolean
): Promise<Partial<InitAnswers>> {
  if (mode !== 'team' || !githubEnabled) {
    return { specWorkflow: 'commit' };
  }

  displayStep(6, 7, t('init.prompts.specWorkflow'));

  const answers = await inquirer.prompt([
    {
      type: 'list',
      name: 'specWorkflow',
      message: chalk.cyan(t('init.prompts.specWorkflow') + ':'),
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
  ]);

  if (answers.specWorkflow === 'branch') {
    displayTip(t('init.prompts.tipBranch'));
  } else {
    displayTip(t('init.prompts.tipCommit'));
  }

  return answers;
}

/**
 * Get auto-push configuration (GitHub enabled only)
 */
export async function promptAutoPush(
  githubEnabled?: boolean
): Promise<Partial<InitAnswers>> {
  if (!githubEnabled) {
    return { autoPush: false };
  }

  displayStep(7, 7, t('init.prompts.remoteSyn'));

  const answers = await inquirer.prompt([
    {
      type: 'confirm',
      name: 'autoPush',
      message: chalk.cyan(t('init.prompts.autoPush')),
      default: true,
    },
  ]);

  if (answers.autoPush) {
    displayTip(t('init.prompts.tipAutoPushEnabled'));
  } else {
    displayTip(t('init.prompts.tipAutoPushDisabled'));
  }

  return answers;
}

/**
 * Display configuration summary
 */
export function displaySummary(answers: InitAnswers): void {
  console.log('\n');
  console.log(chalk.green.bold('✅ Configuration Complete!'));
  console.log('\n');
  console.log(chalk.white.bold('📋 Summary:'));
  console.log(chalk.gray('─'.repeat(60)));
  console.log(
    chalk.cyan('  Project Name:  ') + chalk.white(answers.projectName)
  );
  console.log(
    chalk.cyan('  Mode:          ') +
      chalk.white(answers.mode === 'personal' ? '🧑 Personal' : '👥 Team')
  );
  console.log(
    chalk.cyan('  Git:           ') +
      chalk.white(answers.gitEnabled ? '✓ Enabled' : '✗ Disabled')
  );

  if (answers.mode === 'team') {
    console.log(
      chalk.cyan('  GitHub:        ') +
        chalk.white(answers.githubEnabled ? '✓ Enabled' : '✗ Disabled')
    );
    if (answers.githubEnabled) {
      console.log(
        chalk.cyan('  Repository:    ') +
          chalk.white(answers.githubUrl || 'N/A')
      );
      console.log(
        chalk.cyan('  Workflow:      ') +
          chalk.white(
            answers.specWorkflow === 'branch'
              ? '🌿 Branch + Merge'
              : '📝 Commits'
          )
      );
      console.log(
        chalk.cyan('  Auto-push:     ') +
          chalk.white(answers.autoPush ? '✓ Enabled' : '✗ Disabled')
      );
    }
  }

  console.log(chalk.gray('─'.repeat(60)));
  console.log('\n');
}

/**
 * Run all prompts in sequence
 */
export async function runInteractivePrompts(
  defaultName?: string,
  isCurrentDirMode = false
): Promise<InitAnswers> {
  // Banner is already displayed in init.ts, no need to display again

  // Step 0: Select language FIRST (before any other prompts)
  const localeInfo = await promptLocale();

  const basicInfo = await promptBasicInfo(defaultName, isCurrentDirMode);
  const modeInfo = await promptMode();
  const gitInfo = await promptGitConfig();
  const githubInfo = await promptGitHubConfig(
    modeInfo.mode as 'personal' | 'team'
  );
  const workflowInfo = await promptSpecWorkflow(
    modeInfo.mode as 'personal' | 'team',
    githubInfo.githubEnabled
  );
  const pushInfo = await promptAutoPush(githubInfo.githubEnabled);

  const answers: InitAnswers = {
    locale: localeInfo.locale ?? 'ko',
    projectName: basicInfo.projectName as string,
    mode: modeInfo.mode as 'personal' | 'team',
    gitEnabled: gitInfo.gitEnabled as boolean,
    githubEnabled: githubInfo.githubEnabled ?? undefined,
    githubUrl: githubInfo.githubUrl ?? undefined,
    specWorkflow: workflowInfo.specWorkflow ?? undefined,
    autoPush: pushInfo.autoPush ?? undefined,
  };

  displaySummary(answers);

  return answers;
}

/**
 * Alias for backwards compatibility
 */
export { runInteractivePrompts as promptProjectSetup };

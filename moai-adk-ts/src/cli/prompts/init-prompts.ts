/**
 * @file Interactive prompts for moai init command
 * @author MoAI Team
 * @tags @FEATURE:INTERACTIVE-INIT-019 | Chain: @REQ:INTERACTIVE-INIT-019 -> @DESIGN:INTERACTIVE-INIT-019 -> @TASK:INTERACTIVE-INIT-019 -> @TEST:INTERACTIVE-INIT-019
 * Related: @DOCS:INTERACTIVE-INIT-019
 */

import inquirer from 'inquirer';
import chalk from 'chalk';
import { InputValidator } from '@/utils/input-validator';

/**
 * User answers from interactive prompts
 */
export interface InitAnswers {
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
  console.log(chalk.gray('  Let\'s set up your project with a few questions...'));
  console.log(chalk.gray('  You can change these settings later in .moai/config.json\n'));
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
 * Display helpful tip
 */
function displayTip(tip: string): void {
  console.log(chalk.gray(`  💡 ${tip}`));
}

/**
 * Get basic project information
 */
export async function promptBasicInfo(defaultName?: string, isCurrentDirMode = false): Promise<Partial<InitAnswers>> {
  displayStep(1, 3, 'Project Information');

  // Determine appropriate default name and tip based on mode
  let effectiveDefaultName: string;
  let tipMessage: string;

  if (isCurrentDirMode) {
    // For "moai init ." - suggest current directory name
    const cwd = process.cwd();
    const currentDirName = cwd.split('/').pop() || 'moai-project';
    effectiveDefaultName = currentDirName;
    tipMessage = 'This will be used in configuration (current directory will NOT be renamed)';
  } else {
    // For "moai init project-name" - use provided name or default
    effectiveDefaultName = defaultName || 'moai-project';
    tipMessage = 'This will be used as the folder name and project identifier';
  }

  const answers = await inquirer.prompt([
    {
      type: 'input',
      name: 'projectName',
      message: chalk.cyan('Project name:'),
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
  displayStep(2, 3, 'Development Mode');

  const answers = await inquirer.prompt([
    {
      type: 'list',
      name: 'mode',
      message: chalk.cyan('Select mode:'),
      choices: [
        {
          name: chalk.white('🧑 Personal') + chalk.gray(' - Local development with .moai/specs/'),
          value: 'personal',
          short: 'Personal',
        },
        {
          name: chalk.white('👥 Team') + chalk.gray('     - GitHub Issues for SPEC management'),
          value: 'team',
          short: 'Team',
        },
      ],
      default: 'personal',
    },
  ]);

  if (answers.mode === 'personal') {
    displayTip('Personal mode: SPEC files stored locally, simpler workflow');
  } else {
    displayTip('Team mode: GitHub Issues for SPECs, PR-based workflow');
  }

  return answers;
}

/**
 * Get Git configuration
 */
export async function promptGitConfig(): Promise<Partial<InitAnswers>> {
  displayStep(3, 3, 'Version Control');

  const answers = await inquirer.prompt([
    {
      type: 'confirm',
      name: 'gitEnabled',
      message: chalk.cyan('Initialize local Git repository?'),
      default: true,
    },
  ]);

  if (answers.gitEnabled) {
    displayTip('Git will be initialized with initial commit');
  } else {
    displayTip('You can initialize Git later with: git init');
  }

  return answers;
}

/**
 * Get GitHub configuration (Team mode only)
 */
export async function promptGitHubConfig(mode: 'personal' | 'team'): Promise<Partial<InitAnswers>> {
  if (mode !== 'team') {
    return {};
  }

  displayStep(4, 7, 'GitHub Integration');

  const githubAnswers = await inquirer.prompt([
    {
      type: 'confirm',
      name: 'githubEnabled',
      message: chalk.cyan('Use GitHub for remote repository?'),
      default: true,
    },
  ]);

  if (!githubAnswers.githubEnabled) {
    displayTip('GitHub integration disabled - local Git only');
    return githubAnswers;
  }

  displayStep(5, 7, 'GitHub Repository');

  const urlAnswers = await inquirer.prompt([
    {
      type: 'input',
      name: 'githubUrl',
      message: chalk.cyan('GitHub repository URL:'),
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

  displayTip('Example: https://github.com/username/project-name');

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

  displayStep(6, 7, 'SPEC Workflow');

  const answers = await inquirer.prompt([
    {
      type: 'list',
      name: 'specWorkflow',
      message: chalk.cyan('SPEC workflow:'),
      choices: [
        {
          name: chalk.white('🌿 Branch + Merge') + chalk.gray(' - GitHub PR workflow (recommended)'),
          value: 'branch',
          short: 'Branch',
        },
        {
          name: chalk.white('📝 Local commits') + chalk.gray('  - Direct commits to main'),
          value: 'commit',
          short: 'Commit',
        },
      ],
      default: 'branch',
    },
  ]);

  if (answers.specWorkflow === 'branch') {
    displayTip('Branch workflow: feature/* branches + Pull Requests');
  } else {
    displayTip('Commit workflow: Direct commits to main branch');
  }

  return answers;
}

/**
 * Get auto-push configuration (GitHub enabled only)
 */
export async function promptAutoPush(githubEnabled?: boolean): Promise<Partial<InitAnswers>> {
  if (!githubEnabled) {
    return { autoPush: false };
  }

  displayStep(7, 7, 'Remote Synchronization');

  const answers = await inquirer.prompt([
    {
      type: 'confirm',
      name: 'autoPush',
      message: chalk.cyan('Auto-push commits to remote repository?'),
      default: true,
    },
  ]);

  if (answers.autoPush) {
    displayTip('Commits will be automatically pushed to GitHub');
  } else {
    displayTip('You\'ll need to manually push with: git push');
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
  console.log(chalk.cyan('  Project Name:  ') + chalk.white(answers.projectName));
  console.log(chalk.cyan('  Mode:          ') + chalk.white(answers.mode === 'personal' ? '🧑 Personal' : '👥 Team'));
  console.log(chalk.cyan('  Git:           ') + chalk.white(answers.gitEnabled ? '✓ Enabled' : '✗ Disabled'));

  if (answers.mode === 'team') {
    console.log(chalk.cyan('  GitHub:        ') + chalk.white(answers.githubEnabled ? '✓ Enabled' : '✗ Disabled'));
    if (answers.githubEnabled) {
      console.log(chalk.cyan('  Repository:    ') + chalk.white(answers.githubUrl || 'N/A'));
      console.log(chalk.cyan('  Workflow:      ') + chalk.white(answers.specWorkflow === 'branch' ? '🌿 Branch + Merge' : '📝 Commits'));
      console.log(chalk.cyan('  Auto-push:     ') + chalk.white(answers.autoPush ? '✓ Enabled' : '✗ Disabled'));
    }
  }

  console.log(chalk.gray('─'.repeat(60)));
  console.log('\n');
}

/**
 * Run all prompts in sequence
 */
export async function runInteractivePrompts(defaultName?: string, isCurrentDirMode = false): Promise<InitAnswers> {
  // Banner is already displayed in init.ts, no need to display again

  const basicInfo = await promptBasicInfo(defaultName, isCurrentDirMode);
  const modeInfo = await promptMode();
  const gitInfo = await promptGitConfig();
  const githubInfo = await promptGitHubConfig(modeInfo.mode as 'personal' | 'team');
  const workflowInfo = await promptSpecWorkflow(
    modeInfo.mode as 'personal' | 'team',
    githubInfo.githubEnabled
  );
  const pushInfo = await promptAutoPush(githubInfo.githubEnabled);

  const answers: InitAnswers = {
    projectName: basicInfo.projectName as string,
    mode: modeInfo.mode as 'personal' | 'team',
    gitEnabled: gitInfo.gitEnabled as boolean,
    githubEnabled: githubInfo.githubEnabled,
    githubUrl: githubInfo.githubUrl,
    specWorkflow: workflowInfo.specWorkflow,
    autoPush: pushInfo.autoPush,
  };

  displaySummary(answers);

  return answers;
}

/**
 * Alias for backwards compatibility
 */
export { runInteractivePrompts as promptProjectSetup };
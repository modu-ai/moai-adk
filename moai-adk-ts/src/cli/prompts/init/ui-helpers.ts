/**
 * @file UI helper functions for init prompts
 * @author MoAI Team
 * @tags @CODE:UI-HELPERS-001 | Chain: @SPEC:INTERACTIVE-INIT-019 -> @CODE:INTERACTIVE-INIT-019
 * Related: @DOC:INTERACTIVE-INIT-019
 */

import chalk from 'chalk';
import type { InitAnswers } from './types';

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
 * @param current Current step number
 * @param total Total steps
 * @param question Question text
 */
export function displayStep(current: number, total: number, question: string): void {
  const progress = `[${current}/${total}]`;
  console.log('\n');
  console.log(chalk.blue.bold(`❓ Question ${progress}`));
  console.log(chalk.white(`→ ${question}`));
}

/**
 * Display helpful tip
 * @param tip Tip message to display
 */
export function displayTip(tip: string): void {
  console.log(chalk.gray(`  💡 ${tip}`));
}

/**
 * Display configuration summary
 * @param answers Complete initialization answers
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

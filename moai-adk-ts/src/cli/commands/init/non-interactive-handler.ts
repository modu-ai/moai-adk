// @CODE:INIT-001 | SPEC: SPEC-INIT-001.md | TEST: __tests__/cli/init-noninteractive.test.ts
// Related: @CODE:CLI-001, @SPEC:INIT-001

/**
 * @file Non-interactive mode handler for moai init
 * @author MoAI Team
 * @tags @CODE:INIT-001:HANDLER
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import chalk from 'chalk';
import { InstallationOrchestrator } from '@/core/installer/orchestrator';
import type { InstallationConfig } from '@/core/installer/types';
import type { DoctorCommand } from '../doctor';
import type { InitResult, InitOptions } from '@/types/project';
import { printBanner } from '@/utils/banner';
import { InputValidator } from '@/utils/input-validator';
import { validateProjectPath } from '@/utils/path-validator';
import { logger } from '@/utils/winston-logger.js';

/**
 * Progress callback for installation progress display
 * @param message Progress message
 * @param current Current step
 * @param total Total steps
 */
function displayProgress(
  message: string,
  current: number,
  total: number
): void {
  const percentage = Math.round((current / total) * 100);
  const progressBar =
    '█'.repeat(Math.floor(percentage / 5)) +
    '░'.repeat(20 - Math.floor(percentage / 5));
  logger.log(chalk.blue(`[${progressBar}] ${percentage}% - ${message}`));
}

/**
 * Handle non-interactive initialization mode
 * @param doctorCommand Doctor command instance for system verification
 * @param options Init command options
 * @returns Complete initialization result
 * @description
 * Non-interactive mode features:
 * - No prompts (uses default settings)
 * - Auto-triggered when TTY is not available
 * - Can be explicitly requested with --yes flag
 * - Default: { mode: "personal", gitEnabled: true }
 */
export async function handleNonInteractiveInit(
  doctorCommand: DoctorCommand,
  options?: InitOptions
): Promise<InitResult> {
  try {
    // Display the MoAI banner
    printBanner();

    // Display non-interactive mode message
    console.log(
      chalk.cyan.bold(
        '\n🤖 Running in non-interactive mode with default settings\n'
      )
    );

    // Step 1: System verification with clean separator
    console.log(chalk.gray('─'.repeat(60)));
    console.log(chalk.yellow.bold('📋 Step 1: System Verification'));
    console.log(`${chalk.gray('─'.repeat(60))}\n`);

    const doctorResult = await doctorCommand.run();

    if (!doctorResult.allPassed) {
      console.log(chalk.red.bold('\n❌ System verification failed\n'));
      return {
        success: false,
        projectPath: '',
        config: { name: '', type: 'typescript' as any },
        createdFiles: [],
        errors: ['System verification failed'],
      };
    }

    // Step 2: Use default configuration (no prompts)
    console.log(chalk.gray(`\n${'─'.repeat(60)}`));
    console.log(chalk.yellow.bold('⚙️  Step 2: Default Configuration'));
    console.log(`${chalk.gray('─'.repeat(60))}\n`);

    const projectName = options?.name || 'alfred-project';
    const mode = options?.mode || 'personal';
    const gitEnabled = true;

    console.log(chalk.blue('Using default settings:'));
    console.log(chalk.gray(`  Project: ${projectName}`));
    console.log(chalk.gray(`  Mode: ${mode}`));
    console.log(chalk.gray(`  Git: ${gitEnabled ? 'enabled' : 'disabled'}\n`));

    // Determine initialization mode
    const isCurrentDirMode = options?.name === '.';

    // Determine project path
    let projectPathInput: string;
    if (options?.path) {
      projectPathInput = options.path;
    } else if (isCurrentDirMode) {
      projectPathInput = process.cwd();
    } else {
      projectPathInput = path.join(process.cwd(), projectName);
    }

    // Path validation
    const packageValidation = validateProjectPath(projectPathInput);
    if (!packageValidation.isValid) {
      console.log(chalk.red.bold('\n❌ Invalid project location\n'));
      console.log(chalk.red(packageValidation.error));
      return {
        success: false,
        projectPath: '',
        config: { name: '', type: 'typescript' as any },
        createdFiles: [],
        errors: [packageValidation.error || 'Path validation failed'],
      };
    }

    const pathValidation = await InputValidator.validatePath(projectPathInput, {
      mustBeDirectory: false,
      maxDepth: 10,
    });

    if (!pathValidation.isValid) {
      return {
        success: false,
        projectPath: '',
        config: { name: '', type: 'typescript' as any },
        createdFiles: [],
        errors: ['Project path validation failed:', ...pathValidation.errors],
      };
    }

    // Build default MoAI config
    const moaiConfig = {
      mode,
      gitEnabled,
      projectName,
      locale: 'ko' as const,
    };

    // Save MoAI config
    const finalProjectPath = pathValidation.sanitizedValue || projectPathInput;
    const moaiConfigPath = path.join(finalProjectPath, '.moai', 'config.json');

    const moaiDir = path.join(finalProjectPath, '.moai');
    if (!fs.existsSync(moaiDir)) {
      fs.mkdirSync(moaiDir, { recursive: true });
    }

    fs.writeFileSync(
      moaiConfigPath,
      JSON.stringify(moaiConfig, null, 2),
      'utf-8'
    );
    console.log(chalk.green(`✅ Configuration saved\n`));

    // Check if backup is needed
    const needsBackup =
      isCurrentDirMode &&
      (fs.existsSync(path.join(finalProjectPath, '.moai')) ||
        fs.existsSync(path.join(finalProjectPath, '.claude')));

    const config: InstallationConfig = {
      projectPath: finalProjectPath,
      projectName: projectName,
      mode: moaiConfig.mode,
      backupEnabled: needsBackup,
      overwriteExisting: options?.force || false,
      additionalFeatures: options?.features || [],
    };

    // Step 3: Installation
    console.log(chalk.gray('─'.repeat(60)));
    console.log(chalk.yellow.bold('📦 Step 3: Installation'));
    console.log(`${chalk.gray('─'.repeat(60))}\n`);
    const orchestrator = new InstallationOrchestrator(config);
    const installResult =
      await orchestrator.executeInstallation(displayProgress);

    // Convert to InitResult
    const result: InitResult = {
      success: installResult.success,
      projectPath: installResult.projectPath,
      config: { name: config.projectName, type: 'typescript' as any },
      createdFiles: [...installResult.filesCreated],
      errors: [...installResult.errors],
    };

    if (result.success) {
      console.log(chalk.gray(`\n${'─'.repeat(60)}`));
      console.log(
        chalk.green.bold('✅ Initialization Completed Successfully!')
      );
      console.log(chalk.gray('─'.repeat(60)));

      console.log(chalk.cyan('\n📊 Summary:'));
      console.log(
        chalk.gray(`  📁 Location:  ${chalk.white(result.projectPath)}`)
      );
      console.log(
        chalk.gray(
          `  📄 Files:     ${chalk.white(result.createdFiles.length)} created`
        )
      );
      console.log(
        chalk.gray(
          `  ⏱️  Duration:  ${chalk.white(`${installResult.duration}ms`)}`
        )
      );

      if (installResult.nextSteps.length > 0) {
        console.log(chalk.cyan('\n🚀 Next Steps:'));
        installResult.nextSteps.forEach((step, index) => {
          console.log(chalk.blue(`  ${index + 1}. ${step}`));
        });
      }

      console.log(chalk.gray(`\n${'─'.repeat(60)}\n`));
    } else {
      console.log(chalk.red.bold('\n❌ Initialization failed!'));
      if (result.errors && result.errors.length > 0) {
        console.log(chalk.red('\nErrors:'));
        result.errors.forEach(error => {
          console.log(chalk.red(`  • ${error}`));
        });
      }
      console.log();
    }

    return result;
  } catch (error) {
    const errorMessage =
      error instanceof Error ? error.message : 'Unknown error';

    if (errorMessage === 'User cancelled') {
      logger.log(chalk.yellow('\nInitialization cancelled by user.'));
    } else {
      logger.log(chalk.red(`\nError during initialization: ${errorMessage}`));
    }

    return {
      success: false,
      projectPath: '',
      config: { name: '', type: 'typescript' as any },
      createdFiles: [],
      errors: [errorMessage],
    };
  }
}

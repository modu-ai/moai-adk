// @CODE:INIT-001 | SPEC: SPEC-INIT-001.md | TEST: __tests__/cli/init-noninteractive.test.ts
// Related: @CODE:CLI-001, @SPEC:INIT-001

/**
 * @file Interactive mode handler for moai init
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
import { buildMoAIConfig } from '@/cli/config/config-builder';
import { displayWelcomeBanner, promptProjectSetup } from '@/cli/prompts/init';

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
 * Handle interactive initialization mode
 * @param doctorCommand Doctor command instance for system verification
 * @param options Init command options
 * @returns Complete initialization result
 */
export async function handleInteractiveInit(
  doctorCommand: DoctorCommand,
  options?: InitOptions
): Promise<InitResult> {
  try {
    // Display the MoAI banner
    printBanner();

    // Display initialization header with modern design
    const inputProjectName = options?.name || 'alfred-project';
    console.log(
      chalk.cyan.bold(
        `\n🚀 Initializing ${chalk.white(inputProjectName)} project...\n`
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

    // Step 2: Interactive Configuration with modern separator
    console.log(chalk.gray(`\n${'─'.repeat(60)}`));
    console.log(chalk.yellow.bold('⚙️  Step 2: Interactive Configuration'));
    console.log(`${chalk.gray('─'.repeat(60))}\n`);

    // Display welcome banner for interactive setup
    displayWelcomeBanner();

    // Determine initialization mode BEFORE prompting
    const isCurrentDirMode = options?.name === '.';

    // Run interactive prompts with mode context
    const answers = await promptProjectSetup(options?.name, isCurrentDirMode);

    // Build MoAI config from answers
    const moaiConfig = buildMoAIConfig(answers);

    // Determine project path based on initialization mode
    let projectPathInput: string;
    const projectName = answers.projectName;

    if (options?.path) {
      // Explicit path provided
      projectPathInput = options.path;
    } else if (isCurrentDirMode) {
      // moai init . mode - always use current directory
      projectPathInput = process.cwd();
    } else {
      // moai init project-name mode - create new directory
      projectPathInput = path.join(process.cwd(), projectName);
    }

    // Step 2.1: Check if path is inside MoAI-ADK package
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

    // Step 2.2: Validate path format and structure
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

    // Save MoAI config to .moai/config.json
    const finalProjectPath = pathValidation.sanitizedValue || projectPathInput;
    const moaiConfigPath = path.join(finalProjectPath, '.moai', 'config.json');

    // Ensure .moai directory exists
    const moaiDir = path.join(finalProjectPath, '.moai');
    if (!fs.existsSync(moaiDir)) {
      fs.mkdirSync(moaiDir, { recursive: true });
    }

    // Write config.json
    fs.writeFileSync(
      moaiConfigPath,
      JSON.stringify(moaiConfig, null, 2),
      'utf-8'
    );
    console.log(chalk.green(`\n✅ Configuration saved\n`));

    // Check if backup is needed (only for existing projects in current directory mode)
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

    // Step 3: Full installation with orchestrator and modern header
    console.log(chalk.gray('─'.repeat(60)));
    console.log(chalk.yellow.bold('📦 Step 3: Installation'));
    console.log(`${chalk.gray('─'.repeat(60))}\n`);
    const orchestrator = new InstallationOrchestrator(config);
    const installResult =
      await orchestrator.executeInstallation(displayProgress);

    // Convert InstallationResult to InitResult format
    const result: InitResult = {
      success: installResult.success,
      projectPath: installResult.projectPath,
      config: { name: config.projectName, type: 'typescript' as any },
      createdFiles: [...installResult.filesCreated],
      errors: [...installResult.errors],
    };

    if (result.success) {
      // Modern success message with clean design
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

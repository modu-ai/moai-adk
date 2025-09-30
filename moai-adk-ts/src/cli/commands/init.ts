/**
 * @file CLI init command implementation with integrated TypeScript components
 * @author MoAI Team
 * @tags @FEATURE:CLI-INIT-INTEGRATION-001 @REQ:CLI-FOUNDATION-012
 */

import * as path from 'node:path';
import * as fs from 'node:fs';
import chalk from 'chalk';
import { InstallationOrchestrator } from '@/core/installer/orchestrator';
import type { InstallationConfig } from '@/core/installer/types';
import type { SystemDetector } from '@/core/system-checker/detector';
import type { InitResult } from '@/types/project';
import { createHeader, printBanner } from '@/utils/banner';
import { InputValidator } from '@/utils/input-validator';
import { DoctorCommand } from './doctor';
import { logger } from '../../utils/winston-logger.js';
import { promptProjectSetup, displayWelcomeBanner } from '../prompts/init-prompts';
import { buildMoAIConfig } from '../config/config-builder';

/**
 * Progress callback for installation progress display
 * @param message Progress message
 * @param current Current step
 * @param total Total steps
 * @tags @UTIL:PROGRESS-DISPLAY-001
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
  logger.info(chalk.blue(`[${progressBar}] ${percentage}% - ${message}`));
}

/**
 * Initialize command for project setup with TypeScript orchestrator
 * @tags @FEATURE:CLI-INIT-INTEGRATION-001
 */
export class InitCommand {
  private readonly doctorCommand: DoctorCommand;

  constructor(detector: SystemDetector) {
    this.doctorCommand = new DoctorCommand(detector);
  }

  /**
   * Run project initialization with TypeScript orchestrator
   * @param options Init command options
   * @returns Complete initialization result
   * @tags @API:INIT-INTERACTIVE-001
   */
  public async runInteractive(options?: {
    name?: string;
    mode?: 'personal' | 'team';
    path?: string;
    force?: boolean;
    backup?: boolean;
    features?: string[];
  }): Promise<InitResult> {
    try {
      // Display the MoAI banner
      printBanner();

      // Display initialization header
      const inputProjectName = options?.name || 'moai-project';
      logger.info(createHeader(`Initializing ${inputProjectName} project...`));

      // Step 1: System verification
      logger.info(chalk.yellow.bold('Step 1: System Verification'));
      const doctorResult = await this.doctorCommand.run();

      if (!doctorResult.allPassed) {
        return {
          success: false,
          projectPath: '',
          config: { name: '', type: 'typescript' as any },
          createdFiles: [],
          errors: ['System verification failed'],
        };
      }

      // Step 2: Interactive Configuration
      logger.info(chalk.yellow.bold('\nStep 2: Interactive Configuration'));

      // Display welcome banner for interactive setup
      displayWelcomeBanner();

      // Run interactive prompts
      const answers = await promptProjectSetup(options?.name);

      // Build MoAI config from answers
      const moaiConfig = buildMoAIConfig(answers);

      // Determine project path
      let projectPathInput: string;
      const projectName = answers.projectName;

      if (options?.path) {
        projectPathInput = options.path;
      } else if (projectName === '.' || projectName === 'moai-project') {
        projectPathInput = process.cwd();
      } else {
        projectPathInput = path.join(process.cwd(), projectName);
      }

      const pathValidation = await InputValidator.validatePath(
        projectPathInput,
        {
          mustBeDirectory: false,
          maxDepth: 10,
        }
      );

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
      fs.writeFileSync(moaiConfigPath, JSON.stringify(moaiConfig, null, 2), 'utf-8');
      logger.info(chalk.green(`✅ Config saved: ${moaiConfigPath}`));

      const config: InstallationConfig = {
        projectPath: finalProjectPath,
        projectName: projectName,
        mode: moaiConfig.mode,
        backupEnabled: options?.backup !== false,
        overwriteExisting: options?.force || false,
        additionalFeatures: options?.features || [],
      };

      logger.info(chalk.gray(`  Project: ${config.projectName}`));
      logger.info(chalk.gray(`  Mode: ${config.mode}`));
      logger.info(chalk.gray(`  Path: ${config.projectPath}`));

      // Step 3: Full installation with orchestrator
      logger.info(chalk.yellow.bold('\nStep 3: Installation'));
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
        logger.info(
          chalk.green.bold('\n✅ Initialization completed successfully!')
        );
        logger.info(chalk.gray(`📁 Project created at: ${result.projectPath}`));
        logger.info(
          chalk.gray(`📄 Files created: ${result.createdFiles.length}`)
        );
        logger.info(chalk.gray(`⏱️  Duration: ${installResult.duration}ms`));

        if (installResult.nextSteps.length > 0) {
          logger.info(chalk.blue.bold('\n📋 Next Steps:'));
          installResult.nextSteps.forEach((step, index) => {
            logger.info(chalk.blue(`${index + 1}. ${step}`));
          });
        }
      } else {
        logger.info(chalk.red.bold('\n❌ Initialization failed!'));
        if (result.errors && result.errors.length > 0) {
          logger.info(chalk.red('\nErrors:'));
          result.errors.forEach(error => {
            logger.info(chalk.red(`  • ${error}`));
          });
        }
      }

      return result;
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';

      if (errorMessage === 'User cancelled') {
        logger.info(chalk.yellow('\nInitialization cancelled by user.'));
      } else {
        logger.info(
          chalk.red(`\nError during initialization: ${errorMessage}`)
        );
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

  /**
   * Run project initialization (legacy method)
   * @param projectName - Name of the project to initialize
   * @returns Success status
   * @tags @API:INIT-RUN-001
   */
  public async run(projectName?: string): Promise<boolean> {
    const options = projectName
      ? {
          name: projectName,
          mode: 'personal' as const,
          // Remove explicit path to let runInteractive handle path logic
        }
      : undefined;

    const result = await this.runInteractive(options);
    return result.success;
  }
}

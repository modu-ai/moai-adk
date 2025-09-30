// @FEATURE:CLI-001 | Chain: @REQ:CLI-001 -> @DESIGN:CLI-001 -> @TASK:CLI-001 -> @TEST:CLI-001
// Related: @API:INST-001, @UI:PROMPT-001, @DATA:CFG-001

/**
 * @file CLI init command implementation with integrated TypeScript components
 * @author MoAI Team
 */

import * as path from 'node:path';
import * as fs from 'node:fs';
import chalk from 'chalk';
import { InstallationOrchestrator } from '@/core/installer/orchestrator';
import type { InstallationConfig } from '@/core/installer/types';
import type { SystemDetector } from '@/core/system-checker/detector';
import type { InitResult } from '@/types/project';
import { printBanner } from '@/utils/banner';
import { InputValidator } from '@/utils/input-validator';
import { validateProjectPath } from '@/utils/path-validator';
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

      // Display initialization header with modern design
      const inputProjectName = options?.name || 'moai-project';
      console.log(chalk.cyan.bold(`\n🚀 Initializing ${chalk.white(inputProjectName)} project...\n`));

      // Step 1: System verification with clean separator
      console.log(chalk.gray('─'.repeat(60)));
      console.log(chalk.yellow.bold('📋 Step 1: System Verification'));
      console.log(chalk.gray('─'.repeat(60)) + '\n');

      const doctorResult = await this.doctorCommand.run();

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
      console.log(chalk.gray('\n' + '─'.repeat(60)));
      console.log(chalk.yellow.bold('⚙️  Step 2: Interactive Configuration'));
      console.log(chalk.gray('─'.repeat(60)) + '\n');

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
      console.log(chalk.green(`\n✅ Configuration saved\n`));

      const config: InstallationConfig = {
        projectPath: finalProjectPath,
        projectName: projectName,
        mode: moaiConfig.mode,
        backupEnabled: false, // No backup needed for new project initialization
        overwriteExisting: options?.force || false,
        additionalFeatures: options?.features || [],
      };

      // Step 3: Full installation with orchestrator and modern header
      console.log(chalk.gray('─'.repeat(60)));
      console.log(chalk.yellow.bold('📦 Step 3: Installation'));
      console.log(chalk.gray('─'.repeat(60)) + '\n');
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
        console.log(chalk.gray('\n' + '─'.repeat(60)));
        console.log(chalk.green.bold('✅ Initialization Completed Successfully!'));
        console.log(chalk.gray('─'.repeat(60)));

        console.log(chalk.cyan('\n📊 Summary:'));
        console.log(chalk.gray(`  📁 Location:  ${chalk.white(result.projectPath)}`));
        console.log(chalk.gray(`  📄 Files:     ${chalk.white(result.createdFiles.length)} created`));
        console.log(chalk.gray(`  ⏱️  Duration:  ${chalk.white(installResult.duration + 'ms')}`));

        if (installResult.nextSteps.length > 0) {
          console.log(chalk.cyan('\n🚀 Next Steps:'));
          installResult.nextSteps.forEach((step, index) => {
            console.log(chalk.blue(`  ${index + 1}. ${step}`));
          });
        }

        console.log(chalk.gray('\n' + '─'.repeat(60) + '\n'));
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

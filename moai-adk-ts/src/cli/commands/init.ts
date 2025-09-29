/**
 * @file CLI init command implementation with integrated TypeScript components
 * @author MoAI Team
 * @tags @FEATURE:CLI-INIT-INTEGRATION-001 @REQ:CLI-FOUNDATION-012
 */

import * as path from 'node:path';
import chalk from 'chalk';
import { InstallationOrchestrator } from '@/core/installer/orchestrator';
import type { InstallationConfig } from '@/core/installer/types';
import type { SystemDetector } from '@/core/system-checker/detector';
import type { InitResult } from '@/types/project';
import { createHeader, printBanner } from '@/utils/banner';
import { InputValidator } from '@/utils/input-validator';
import { DoctorCommand } from './doctor';

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
  console.log(chalk.blue(`[${progressBar}] ${percentage}% - ${message}`));
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
      console.log(createHeader(`Initializing ${inputProjectName} project...`));

      // Step 1: System verification
      console.log(chalk.yellow.bold('Step 1: System Verification'));
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

      // Step 2: Input validation and configuration setup
      console.log(chalk.yellow.bold('\nStep 2: Configuration'));

      // Validate project name
      const projectNameValidation = InputValidator.validateProjectName(
        options?.name || 'moai-project',
        { maxLength: 50, allowSpaces: false }
      );

      if (!projectNameValidation.isValid) {
        return {
          success: false,
          projectPath: '',
          config: { name: '', type: 'typescript' as any },
          createdFiles: [],
          errors: [
            'Project name validation failed:',
            ...projectNameValidation.errors,
          ],
        };
      }

      // Determine project path
      let projectPathInput: string;
      const projectName =
        projectNameValidation.sanitizedValue || 'moai-project';

      if (options?.path) {
        // Explicit path provided
        projectPathInput = options.path;
      } else if (projectName === '.' || projectName === 'moai-project') {
        // Current directory installation (moai init . or default name)
        projectPathInput = process.cwd();
      } else {
        // Create new directory with project name
        projectPathInput = path.join(process.cwd(), projectName);
      }

      const pathValidation = await InputValidator.validatePath(
        projectPathInput,
        {
          mustBeDirectory: false, // Directory will be created if needed
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

      // Validate mode
      const validModes = ['personal', 'team'];
      const mode = options?.mode || 'personal';
      if (!validModes.includes(mode)) {
        return {
          success: false,
          projectPath: '',
          config: { name: '', type: 'typescript' as any },
          createdFiles: [],
          errors: [
            `Invalid mode: ${mode}. Must be one of: ${validModes.join(', ')}`,
          ],
        };
      }

      const config: InstallationConfig = {
        projectPath: pathValidation.sanitizedValue || projectPathInput,
        projectName: projectNameValidation.sanitizedValue || 'moai-project',
        mode: mode as 'personal' | 'team',
        backupEnabled: options?.backup !== false,
        overwriteExisting: options?.force || false,
        additionalFeatures: options?.features || [],
      };

      console.log(chalk.gray(`  Project: ${config.projectName}`));
      console.log(chalk.gray(`  Mode: ${config.mode}`));
      console.log(chalk.gray(`  Path: ${config.projectPath}`));

      // Step 3: Full installation with orchestrator
      console.log(chalk.yellow.bold('\nStep 3: Installation'));
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
        console.log(
          chalk.green.bold('\n✅ Initialization completed successfully!')
        );
        console.log(chalk.gray(`📁 Project created at: ${result.projectPath}`));
        console.log(
          chalk.gray(`📄 Files created: ${result.createdFiles.length}`)
        );
        console.log(chalk.gray(`⏱️  Duration: ${installResult.duration}ms`));

        if (installResult.nextSteps.length > 0) {
          console.log(chalk.blue.bold('\n📋 Next Steps:'));
          installResult.nextSteps.forEach((step, index) => {
            console.log(chalk.blue(`${index + 1}. ${step}`));
          });
        }
      } else {
        console.log(chalk.red.bold('\n❌ Initialization failed!'));
        if (result.errors && result.errors.length > 0) {
          console.log(chalk.red('\nErrors:'));
          result.errors.forEach(error => {
            console.log(chalk.red(`  • ${error}`));
          });
        }
      }

      return result;
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';

      if (errorMessage === 'User cancelled') {
        console.log(chalk.yellow('\nInitialization cancelled by user.'));
      } else {
        console.log(
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

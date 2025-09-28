/**
 * @file CLI init command implementation with integrated TypeScript components
 * @author MoAI Team
 * @tags @FEATURE:CLI-INIT-INTEGRATION-001 @REQ:CLI-FOUNDATION-012
 */

import chalk from 'chalk';
import type { SystemDetector } from '@/core/system-checker/detector';
import { InstallationOrchestrator } from '@/core/installer/orchestrator';
import type { InstallationConfig } from '@/core/installer/types';
import { DoctorCommand } from './doctor';
import type { InitResult } from '@/types/project';

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
      console.log(
        chalk.blue.bold('🚀 MoAI-ADK TypeScript Project Initialization')
      );
      console.log(chalk.blue('Integrating all ported components...\n'));

      // Step 1: System verification
      console.log(chalk.yellow('Step 1: System Verification'));
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

      // Step 2: Configuration setup
      console.log(chalk.yellow('\nStep 2: Configuration Setup'));
      const config: InstallationConfig = {
        projectPath: options?.path || process.cwd(),
        projectName: options?.name || 'moai-project',
        mode: options?.mode || 'personal',
        backupEnabled: options?.backup !== false,
        overwriteExisting: options?.force || false,
        additionalFeatures: options?.features || [],
      };

      console.log(chalk.gray(`  Project: ${config.projectName}`));
      console.log(chalk.gray(`  Mode: ${config.mode}`));
      console.log(chalk.gray(`  Path: ${config.projectPath}`));

      // Step 3: Full installation with orchestrator
      console.log(chalk.yellow('\nStep 3: Project Installation'));
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
          path: process.cwd(),
        }
      : undefined;

    const result = await this.runInteractive(options);
    return result.success;
  }
}

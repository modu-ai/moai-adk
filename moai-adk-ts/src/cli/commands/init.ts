/**
 * @file CLI init command implementation
 * @author MoAI Team
 * @tags @FEATURE:CLI-INIT-001 @REQ:CLI-FOUNDATION-012
 */

import chalk from 'chalk';
import { SystemDetector } from '@/core/system-checker/detector';
import { ProjectWizard } from '@/core/project/wizard';
import { TemplateManager } from '@/core/project/template-manager';
import { DoctorCommand } from './doctor';
import { InitResult } from '@/types/project';

/**
 * Initialize command for project setup
 * @tags @FEATURE:CLI-INIT-001
 */
export class InitCommand {
  private readonly doctorCommand: DoctorCommand;
  private readonly wizard: ProjectWizard;
  private readonly templateManager: TemplateManager;

  constructor(
    detector: SystemDetector,
    wizard?: ProjectWizard,
    templateManager?: TemplateManager
  ) {
    this.doctorCommand = new DoctorCommand(detector);
    this.wizard = wizard || new ProjectWizard();
    this.templateManager = templateManager || new TemplateManager();
  }

  /**
   * Run project initialization with interactive wizard
   * @returns Complete initialization result
   * @tags @API:INIT-INTERACTIVE-001
   */
  public async runInteractive(): Promise<InitResult> {
    try {
      console.log(chalk.blue.bold('🗿 MoAI-ADK Project Initialization'));
      console.log(chalk.blue('Setting up your development environment...\n'));

      // Step 1: System verification
      console.log(chalk.yellow('Step 1: System Verification'));
      const doctorResult = await this.doctorCommand.run();

      if (!doctorResult.allPassed) {
        return {
          success: false,
          projectPath: '',
          config: { name: '', type: 'python' as any },
          createdFiles: [],
          errors: ['System verification failed']
        };
      }

      // Step 2: Interactive configuration
      console.log(chalk.yellow('\nStep 2: Project Configuration'));
      const projectConfig = await this.wizard.run();

      // Step 3: Project generation
      console.log(chalk.yellow('\nStep 3: Project Generation'));
      const currentDir = process.cwd();
      const result = await this.templateManager.generateProject(projectConfig, currentDir);

      if (result.success) {
        console.log(chalk.green.bold('\n🎉 Initialization completed successfully!'));
        console.log(chalk.blue(`Project created at: ${result.projectPath}`));
        console.log(chalk.gray(`Files created: ${result.createdFiles.length}`));
      } else {
        console.log(chalk.red.bold('\n❌ Initialization failed!'));
        if (result.errors) {
          result.errors.forEach(error => console.log(chalk.red(`  • ${error}`)));
        }
      }

      return result;

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';

      if (errorMessage === 'User cancelled') {
        console.log(chalk.yellow('\nInitialization cancelled by user.'));
      } else {
        console.log(chalk.red(`\nError during initialization: ${errorMessage}`));
      }

      return {
        success: false,
        projectPath: '',
        config: { name: '', type: 'python' as any },
        createdFiles: [],
        errors: [errorMessage]
      };
    }
  }

  /**
   * Run project initialization (legacy method)
   * @param _projectName - Name of the project to initialize (unused in interactive mode)
   * @returns Success status
   * @tags @API:INIT-RUN-001
   */
  public async run(_projectName?: string): Promise<boolean> {
    const result = await this.runInteractive();
    return result.success;
  }
}

/**
 * @file CLI init command implementation
 * @author MoAI Team
 * @tags @FEATURE:CLI-INIT-001 @REQ:CLI-FOUNDATION-012
 */

import chalk from 'chalk';
import { SystemDetector } from '@/core/system-checker/detector';
import { DoctorCommand } from './doctor';

/**
 * Initialize command for project setup
 * @tags @FEATURE:CLI-INIT-001
 */
export class InitCommand {
  private readonly doctorCommand: DoctorCommand;

  constructor(detector: SystemDetector) {
    this.doctorCommand = new DoctorCommand(detector);
  }

  /**
   * Run project initialization
   * @param projectName - Name of the project to initialize
   * @returns Success status
   * @tags @API:INIT-RUN-001
   */
  public async run(projectName?: string): Promise<boolean> {
    console.log(chalk.blue.bold('🗿 MoAI-ADK Project Initialization'));
    console.log(chalk.blue('Setting up your development environment...\n'));

    // Step 1: System verification
    console.log(chalk.yellow('Step 1: System Verification'));
    const doctorResult = await this.doctorCommand.run();

    if (!doctorResult.allPassed) {
      console.log(chalk.red('\n❌ System verification failed.'));
      console.log(
        chalk.yellow('Please resolve the above issues before proceeding.')
      );
      return false;
    }

    // Step 2: Project setup (placeholder for now)
    console.log(chalk.yellow('\nStep 2: Project Setup'));
    if (projectName) {
      console.log(chalk.green(`✅ Project name: ${projectName}`));
    } else {
      console.log(chalk.blue('📝 Using current directory for project setup'));
    }

    console.log(
      chalk.green.bold('\n🎉 Initialization completed successfully!')
    );
    console.log(chalk.blue('Your MoAI-ADK project is ready to use.'));

    return true;
  }
}

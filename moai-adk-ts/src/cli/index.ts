#!/usr/bin/env node
/**
 * @file CLI entry point
 * @author MoAI Team
 * @tags @FEATURE:CLI-ENTRY-001 @REQ:CLI-FOUNDATION-012
 */

import { Command } from 'commander';
import chalk from 'chalk';
import { SystemDetector } from '@/core/system-checker/detector';
import { DoctorCommand } from './commands/doctor';
import { InitCommand } from './commands/init';

/**
 * CLI Application
 * @tags @FEATURE:CLI-APP-001
 */
export class CLIApp {
  private readonly program: Command;
  private readonly detector: SystemDetector;
  private readonly doctorCommand: DoctorCommand;
  private readonly initCommand: InitCommand;

  constructor() {
    this.program = new Command();
    this.detector = new SystemDetector();
    this.doctorCommand = new DoctorCommand(this.detector);
    this.initCommand = new InitCommand(this.detector);

    this.setupCommands();
  }

  /**
   * Setup CLI commands
   * @tags @SETUP:CLI-COMMANDS-001
   */
  private setupCommands(): void {
    this.program
      .name('moai')
      .description('🗿 MoAI-ADK: Modu-AI Agentic Development kit')
      .version('0.0.1', '-v, --version', 'output the current version');

    // Doctor command
    this.program
      .command('doctor')
      .description('Run system diagnostics')
      .action(async () => {
        try {
          await this.doctorCommand.run();
        } catch (error) {
          console.error(chalk.red('Error running diagnostics:'), error);
          process.exit(1);
        }
      });

    // Init command
    this.program
      .command('init [project]')
      .description('Initialize a new MoAI-ADK project')
      .action(async (project: string | undefined) => {
        try {
          const success = await this.initCommand.run(project);
          process.exit(success ? 0 : 1);
        } catch (error) {
          console.error(chalk.red('Error during initialization:'), error);
          process.exit(1);
        }
      });
  }

  /**
   * Run CLI application
   * @param argv - Command line arguments
   * @tags @API:CLI-RUN-001
   */
  public run(argv: string[]): void {
    this.program.parse(argv);
  }
}

// Main execution
if (require.main === module) {
  const app = new CLIApp();
  app.run(process.argv);
}

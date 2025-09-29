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
import { RestoreCommand } from './commands/restore';
import { StatusCommand } from './commands/status';
import { UpdateCommand } from './commands/update';
import { HelpCommand } from './commands/help';
import { createBanner } from '@/utils/banner';
import { getCurrentVersion } from '@/utils/version';

/**
 * CLI Application
 * @tags @FEATURE:CLI-APP-001
 */
export class CLIApp {
  private readonly program: Command;
  private readonly detector: SystemDetector;
  private readonly doctorCommand: DoctorCommand;
  private readonly initCommand: InitCommand;
  private readonly restoreCommand: RestoreCommand;
  private readonly statusCommand: StatusCommand;
  private readonly updateCommand: UpdateCommand;
  private readonly helpCommand: HelpCommand;

  constructor() {
    this.program = new Command();
    this.detector = new SystemDetector();
    this.doctorCommand = new DoctorCommand(this.detector);
    this.initCommand = new InitCommand(this.detector);
    this.restoreCommand = new RestoreCommand();
    this.statusCommand = new StatusCommand();
    this.updateCommand = new UpdateCommand();
    this.helpCommand = new HelpCommand();

    this.setupCommands();
  }

  /**
   * Setup CLI commands
   * @tags @SETUP:CLI-COMMANDS-001
   */
  private setupCommands(): void {
    this.program
      .name('moai')
      .description('')  // Remove duplicate description since it's in the banner
      .version(getCurrentVersion(), '-v, --version', 'output the current version')
      .configureHelp({
        helpWidth: 80,
        sortSubcommands: false,
        formatHelp: () => {
          return createBanner({ showUsage: true });
        }
      })
      .helpOption('-h, --help', 'display help for command');

    // Doctor command
    this.program
      .command('doctor')
      .description('Run system diagnostics')
      .option('-l, --list-backups', 'List available backups')
      .action(async (options: { listBackups?: boolean }) => {
        try {
          await this.doctorCommand.run(options);
        } catch (error) {
          console.error(chalk.red('Error running diagnostics:'), error);
          process.exit(1);
        }
      });

    // Init command
    this.program
      .command('init [project]')
      .description('Initialize a new MoAI-ADK project')
      .option(
        '-t, --template <type>',
        'Template to use (standard, minimal, advanced)',
        'standard'
      )
      .option('-i, --interactive', 'Run interactive setup wizard')
      .option('-b, --backup', 'Create backup before installation')
      .option('-f, --force', 'Force overwrite existing files')
      .option('--personal', 'Initialize in personal mode (default)')
      .option('--team', 'Initialize in team mode')
      .action(async (project: string | undefined) => {
        try {
          const success = await this.initCommand.run(project);
          process.exit(success ? 0 : 1);
        } catch (error) {
          console.error(chalk.red('Error during initialization:'), error);
          process.exit(1);
        }
      });

    // Restore command
    this.program
      .command('restore <backup-path>')
      .description('Restore MoAI-ADK from a backup directory')
      .option('--dry-run', 'Show what would be restored without making changes')
      .option('--force', 'Force overwrite existing files')
      .action(
        async (
          backupPath: string,
          options: { dryRun?: boolean; force?: boolean }
        ) => {
          try {
            const result = await this.restoreCommand.run(backupPath, {
              dryRun: options.dryRun || false,
              force: options.force,
            });
            process.exit(result.success ? 0 : 1);
          } catch (error) {
            console.error(chalk.red('Error during restore:'), error);
            process.exit(1);
          }
        }
      );

    // Status command
    this.program
      .command('status')
      .description('Show MoAI-ADK project status')
      .option('-v, --verbose', 'Show detailed status information')
      .option('-p, --project-path <path>', 'Path to project directory')
      .action(async (options: { verbose?: boolean; projectPath?: string }) => {
        try {
          const result = await this.statusCommand.run({
            verbose: options.verbose || false,
            projectPath: options.projectPath,
          });
          process.exit(result.success ? 0 : 1);
        } catch (error) {
          console.error(chalk.red('Error getting status:'), error);
          process.exit(1);
        }
      });

    // Update command
    this.program
      .command('update')
      .description('Update MoAI-ADK to the latest version')
      .option('-c, --check', 'Check for updates without installing')
      .option('--no-backup', 'Skip backup creation before update')
      .option('-v, --verbose', 'Show detailed update information')
      .option('--package-only', 'Update only the package')
      .option('--resources-only', 'Update only project resources')
      .action(
        async (options: {
          check?: boolean;
          noBackup?: boolean;
          verbose?: boolean;
          packageOnly?: boolean;
          resourcesOnly?: boolean;
        }) => {
          try {
            const result = await this.updateCommand.run({
              check: options.check || false,
              noBackup: options.noBackup,
              verbose: options.verbose,
              packageOnly: options.packageOnly,
              resourcesOnly: options.resourcesOnly,
            });
            process.exit(result.success ? 0 : 1);
          } catch (error) {
            console.error(chalk.red('Error during update:'), error);
            process.exit(1);
          }
        }
      );

    // Help command
    this.program
      .command('help [command]')
      .description('Show help for MoAI-ADK commands')
      .action(async (command: string | undefined) => {
        try {
          const result = await this.helpCommand.run({ command });
          process.exit(result.success ? 0 : 1);
        } catch (error) {
          console.error(chalk.red('Error showing help:'), error);
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

// Main execution - ESM compatible
if (import.meta.url === `file://${process.argv[1]}`) {
  const app = new CLIApp();
  app.run(process.argv);
}

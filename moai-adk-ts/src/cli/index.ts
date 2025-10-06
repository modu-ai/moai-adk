#!/usr/bin/env node

/**
 * @file CLI entry point
 * @author MoAI Team
 * @tags @CODE:CLI-ENTRY-001 @SPEC:CLI-FOUNDATION-012
 */

import { existsSync, readFileSync, realpathSync } from 'node:fs';
import { join } from 'node:path';
import chalk from 'chalk';
import { Command } from 'commander';
import { SystemDetector } from '@/core/system-checker/detector';
import { createBanner } from '@/utils/banner';
import { type Locale, setLocale } from '@/utils/i18n';
import { getCurrentVersion } from '@/utils/version';
import { logger } from '../utils/winston-logger.js';
import { DoctorCommand } from './commands/doctor';
import { HelpCommand } from './commands/help';
import { InitCommand } from './commands/init';
import { RestoreCommand } from './commands/restore';
import { StatusCommand } from './commands/status';

/**
 * CLI Application
 * @tags @CODE:CLI-APP-001
 */
export class CLIApp {
  private readonly program: Command;
  private readonly detector: SystemDetector;
  private readonly doctorCommand: DoctorCommand;
  private readonly initCommand: InitCommand;
  private readonly restoreCommand: RestoreCommand;
  private readonly statusCommand: StatusCommand;
  private readonly helpCommand: HelpCommand;

  constructor() {
    // Load locale from config.json if available
    this.loadLocaleFromConfig();

    this.program = new Command();
    this.detector = new SystemDetector();
    this.doctorCommand = new DoctorCommand(this.detector);
    this.initCommand = new InitCommand(this.detector);
    this.restoreCommand = new RestoreCommand();
    this.statusCommand = new StatusCommand();
    this.helpCommand = new HelpCommand();

    this.setupCommands();
  }

  /**
   * Load locale from config.json if available
   * @tags @CODE:I18N-INIT-001
   */
  private loadLocaleFromConfig(): void {
    try {
      const configPath = join(process.cwd(), '.moai', 'config.json');
      if (existsSync(configPath)) {
        const configContent = readFileSync(configPath, 'utf-8');
        const config = JSON.parse(configContent) as { locale?: Locale };

        if (config.locale) {
          setLocale(config.locale);
        }
      }
    } catch (_error) {
      // Silently ignore errors - will use default locale (ko)
      // This is expected when running outside of a MoAI project directory
    }
  }

  /**
   * Setup CLI commands
   * @tags SETUP:CLI-COMMANDS-001
   */
  private setupCommands(): void {
    this.program
      .name('moai')
      .description('') // Remove duplicate description since it's in the banner
      .version(
        getCurrentVersion(),
        '-v, --version',
        'output the current version'
      )
      .configureHelp({
        helpWidth: 80,
        sortSubcommands: false,
        formatHelp: () => {
          return createBanner({ showUsage: true });
        },
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
          logger.error(chalk.red('Error running diagnostics:'), error);
          process.exit(1);
        }
      });

    // Init command
    this.program
      .command('init [project]')
      .description('Initialize a new MoAI-ADK project')
      .option('-y, --yes', 'Skip prompts and use default settings')
      .option('-b, --backup', 'Create backup before installation')
      .option('-f, --force', 'Force overwrite existing files')
      .option('--personal', 'Initialize in personal mode (default)')
      .option('--team', 'Initialize in team mode')
      .action(
        async (
          project: string | undefined,
          options: {
            yes?: boolean;
            backup?: boolean;
            force?: boolean;
            personal?: boolean;
            team?: boolean;
          }
        ) => {
          try {
            // Import TTY detector
            const { isTTYAvailable } = await import('@/utils/tty-detector');

            // Determine mode: team takes precedence over personal
            const mode = options.team ? 'team' : 'personal';

            // Check if non-interactive mode should be used
            // 1. Explicit --yes flag
            // 2. TTY not available (Claude Code, CI/CD, Docker)
            const useNonInteractive = options.yes || !isTTYAvailable();

            if (useNonInteractive) {
              // Non-interactive mode
              const result = await this.initCommand.runNonInteractive({
                ...(project && { name: project }),
                mode: mode as 'personal' | 'team',
                ...(options.backup !== undefined && { backup: options.backup }),
                ...(options.force !== undefined && { force: options.force }),
              });

              process.exit(result.success ? 0 : 1);
            } else {
              // Interactive mode (existing behavior)
              const result = await this.initCommand.runInteractive({
                ...(project && { name: project }),
                mode: mode as 'personal' | 'team',
                ...(options.backup !== undefined && { backup: options.backup }),
                ...(options.force !== undefined && { force: options.force }),
              });

              process.exit(result.success ? 0 : 1);
            }
          } catch (error) {
            logger.error(chalk.red('Error during initialization:'), error);
            process.exit(1);
          }
        }
      );

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
            logger.error(chalk.red('Error during restore:'), error);
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
          logger.error(chalk.red('Error getting status:'), error);
          process.exit(1);
        }
      });

    // Help command
    this.program
      .command('help [command]')
      .description('Show help for MoAI-ADK commands')
      .action(async (command: string | undefined) => {
        try {
          const result = await this.helpCommand.run({ command });
          process.exit(result.success ? 0 : 1);
        } catch (error) {
          logger.error(chalk.red('Error showing help:'), error);
          process.exit(1);
        }
      });
  }

  /**
   * Run CLI application
   * @param argv - Command line arguments
   * @tags @CODE:CLI-RUN-001:API
   */
  public run(argv: string[]): void {
    this.program.parse(argv);
  }
}

/**
 * Main execution entry point
 * Uses realpathSync to handle symlink execution (e.g., via bun/npm global install)
 * Guards against undefined argv[1] in REPL/eval contexts
 * @see https://nodejs.org/api/fs.html#fsrealpathsyncpath-options
 */
if (
  process.argv[1] &&
  import.meta.url === `file://${realpathSync(process.argv[1])}`
) {
  const app = new CLIApp();
  app.run(process.argv);
}

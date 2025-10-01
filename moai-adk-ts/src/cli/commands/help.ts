// @CODE:CLI-006 |
// Related: @CODE:HELP-001:UI

/**
 * @file CLI help command and documentation display
 * @author MoAI Team
 */

import chalk from 'chalk';
import { logger } from '../../utils/winston-logger.js';

/**
 * Help command options
 * @tags @SPEC:HELP-OPTIONS-001
 */
export interface HelpOptions {
  readonly command?: string | undefined;
}

/**
 * Command help information
 * @tags @SPEC:COMMAND-HELP-001
 */
export interface CommandHelp {
  readonly name: string;
  readonly description: string;
  readonly usage: string;
  readonly options: Array<{
    readonly flag: string;
    readonly description: string;
  }>;
  readonly examples: string[];
}

/**
 * Help result
 * @tags @SPEC:HELP-RESULT-001
 */
export interface HelpResult {
  readonly success: boolean;
  readonly command?: string;
  readonly helpText: string;
}

/**
 * Help command for comprehensive help system
 * @tags @CODE:CLI-HELP-001
 */
export class HelpCommand {
  private readonly commands: Map<string, CommandHelp> = new Map();

  constructor() {
    this.initializeCommands();
  }

  /**
   * Initialize command help information
   * @tags UTIL:INIT-COMMANDS-001
   */
  private initializeCommands(): void {
    this.commands.set('init', {
      name: 'init',
      description: 'Initialize a new MoAI-ADK project',
      usage: 'moai init [project-path]',
      options: [
        {
          flag: '--template, -t',
          description: 'Template to use (standard, minimal, advanced)',
        },
        {
          flag: '--interactive, -i',
          description: 'Run interactive setup wizard',
        },
        {
          flag: '--backup, -b',
          description: 'Create backup before installation',
        },
        { flag: '--force, -f', description: 'Force overwrite existing files' },
        {
          flag: '--personal',
          description: 'Initialize in personal mode (default)',
        },
        { flag: '--team', description: 'Initialize in team mode' },
      ],
      examples: [
        'moai init',
        'moai init my-project',
        'moai init --interactive',
        'moai init --team --backup',
      ],
    });

    this.commands.set('doctor', {
      name: 'doctor',
      description: 'Run system diagnostics and check health',
      usage: 'moai doctor',
      options: [
        { flag: '--list-backups, -l', description: 'List available backups' },
      ],
      examples: ['moai doctor', 'moai doctor --list-backups'],
    });

    this.commands.set('restore', {
      name: 'restore',
      description: 'Restore MoAI-ADK from a backup directory',
      usage: 'moai restore <backup-path>',
      options: [
        {
          flag: '--dry-run',
          description: 'Show what would be restored without making changes',
        },
        { flag: '--force', description: 'Force overwrite existing files' },
      ],
      examples: [
        'moai restore ./backup',
        'moai restore --dry-run ./backup',
        'moai restore --force ./backup',
      ],
    });

    this.commands.set('status', {
      name: 'status',
      description: 'Show MoAI-ADK project status',
      usage: 'moai status',
      options: [
        {
          flag: '--verbose, -v',
          description: 'Show detailed status information',
        },
        {
          flag: '--project-path, -p',
          description: 'Path to project directory',
        },
      ],
      examples: [
        'moai status',
        'moai status --verbose',
        'moai status --project-path /path/to/project',
      ],
    });

    this.commands.set('update', {
      name: 'update',
      description: 'Update MoAI-ADK to the latest version',
      usage: 'moai update',
      options: [
        {
          flag: '--check, -c',
          description: 'Check for updates without installing',
        },
        {
          flag: '--no-backup',
          description: 'Skip backup creation before update',
        },
        {
          flag: '--verbose, -v',
          description: 'Show detailed update information',
        },
        { flag: '--package-only', description: 'Update only the package' },
        {
          flag: '--resources-only',
          description: 'Update only project resources',
        },
      ],
      examples: [
        'moai update',
        'moai update --check',
        'moai update --no-backup',
        'moai update --package-only',
      ],
    });

    this.commands.set('help', {
      name: 'help',
      description: 'Show help for MoAI-ADK commands',
      usage: 'moai help [command]',
      options: [],
      examples: ['moai help', 'moai help init', 'moai help status'],
    });
  }

  /**
   * Get help for a specific command
   * @param commandName - Name of the command
   * @returns Command help information
   * @tags @CODE:GET-COMMAND-HELP-001:API
   */
  public getCommandHelp(commandName: string): CommandHelp | undefined {
    return this.commands.get(commandName);
  }

  /**
   * Get list of all available commands
   * @returns Array of command names
   * @tags @CODE:GET-COMMANDS-001:API
   */
  public getAvailableCommands(): string[] {
    return Array.from(this.commands.keys());
  }

  /**
   * Format general help text
   * @returns Formatted help text
   * @tags UTIL:FORMAT-GENERAL-HELP-001
   */
  public formatGeneralHelp(): string {
    const banner = `
🗿 MoAI-ADK: Modu-AI Agentic Development Kit

Usage: moai <command> [options]

Available commands:`;

    const commandList = Array.from(this.commands.values())
      .map(cmd => `  ${cmd.name.padEnd(12)} ${cmd.description}`)
      .join('\n');

    const footer = `
Run 'moai help <command>' for detailed information about a specific command.
Run 'moai <command> --help' for command-specific help.

Examples:
  moai init                 Initialize a new project
  moai doctor              Check system health
  moai status              Show project status
  moai help init           Get help for init command
`;

    return `${banner}\n${commandList}${footer}`;
  }

  /**
   * Format help for a specific command
   * @param commandHelp - Command help information
   * @returns Formatted help text
   * @tags UTIL:FORMAT-COMMAND-HELP-001
   */
  public formatCommandHelp(commandHelp: CommandHelp): string {
    let help = `
🗿 MoAI-ADK Command: ${commandHelp.name}

Description:
  ${commandHelp.description}

Usage:
  ${commandHelp.usage}
`;

    if (commandHelp.options.length > 0) {
      help += '\nOptions:\n';
      help += commandHelp.options
        .map(opt => `  ${opt.flag.padEnd(20)} ${opt.description}`)
        .join('\n');
    }

    if (commandHelp.examples.length > 0) {
      help += '\n\nExamples:\n';
      help += commandHelp.examples.map(example => `  ${example}`).join('\n');
    }

    return `${help}\n`;
  }

  /**
   * Run help command
   * @param options - Help options
   * @returns Help result
   * @tags @CODE:HELP-RUN-001:API
   */
  public async run(options: HelpOptions): Promise<HelpResult> {
    try {
      if (options.command) {
        // Show help for specific command
        const commandHelp = this.getCommandHelp(options.command);

        if (!commandHelp) {
          const helpText = `❌ Unknown command: ${options.command}\n\nAvailable commands: ${this.getAvailableCommands().join(', ')}\n\nRun 'moai help' to see all commands.`;
          logger.info(chalk.red(helpText));

          return {
            success: false,
            command: options.command,
            helpText,
          };
        }

        const helpText = this.formatCommandHelp(commandHelp);
        logger.info(helpText);

        return {
          success: true,
          command: options.command,
          helpText,
        };
      } else {
        // Show general help
        const helpText = this.formatGeneralHelp();
        logger.info(helpText);

        return {
          success: true,
          helpText,
        };
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      const helpText = `❌ Failed to show help: ${errorMessage}`;
      logger.info(chalk.red(helpText));

      return {
        success: false,
        helpText,
      };
    }
  }
}

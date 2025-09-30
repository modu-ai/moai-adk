// @FEATURE:CLI-002 | Chain: @REQ:CLI-002 -> @DESIGN:CLI-002 -> @TASK:CLI-002 -> @TEST:CLI-002
// Related: @API:SYS-001, @DATA:SYS-INFO-001

/**
 * @file CLI doctor command for system diagnostics
 * @author MoAI Team
 */

import * as fs from 'node:fs/promises';
import * as os from 'node:os';
import * as path from 'node:path';
import chalk from 'chalk';
import { logger } from '../../utils/winston-logger.js';
import {
  type RequirementCheckResult,
  SystemChecker,
  type SystemCheckSummary,
  type SystemDetector,
} from '@/core/system-checker';

/**
 * Doctor command result summary
 * @tags @DESIGN:DOCTOR-RESULT-001
 */
export interface DoctorResult {
  readonly allPassed: boolean;
  readonly results: RequirementCheckResult[];
  readonly missingRequirements: RequirementCheckResult[];
  readonly versionConflicts: RequirementCheckResult[];
  readonly summary: {
    readonly total: number;
    readonly passed: number;
    readonly failed: number;
  };
}

/**
 * Categorized check results type
 * @tags @DESIGN:CATEGORIZED-RESULTS-001
 */
type CategorizedResults = {
  readonly missing: RequirementCheckResult[];
  readonly conflicts: RequirementCheckResult[];
  readonly passed: RequirementCheckResult[];
  readonly allPassed: boolean;
};

/**
 * Doctor command for system diagnostics with enhanced language detection
 * @tags @FEATURE:CLI-DOCTOR-001
 */
export class DoctorCommand {
  private readonly systemChecker = new SystemChecker();
  constructor(private readonly detector: SystemDetector) {}

  /**
   * Run system diagnostics with language detection
   * @param options - Doctor command options
   * @returns Doctor result with all checks
   * @tags @API:DOCTOR-RUN-001
   */
  public async run(
    options: { listBackups?: boolean; projectPath?: string } = {}
  ): Promise<DoctorResult> {
    // Handle --list-backups option
    if (options.listBackups) {
      return this.listBackups();
    }

    this.printHeader();

    // Use enhanced system checker with language detection
    const projectPath = options.projectPath || process.cwd();
    const checkSummary = await this.systemChecker.runSystemCheck(projectPath);

    this.printEnhancedResults(checkSummary);
    this.printEnhancedSummary(checkSummary);

    const results = [
      ...checkSummary.runtime,
      ...checkSummary.development,
      ...checkSummary.optional,
    ];
    const categorizedResults = this.categorizeResults(results);

    return {
      allPassed: checkSummary.passedChecks === checkSummary.totalChecks,
      results,
      missingRequirements: categorizedResults.missing,
      versionConflicts: categorizedResults.conflicts,
      summary: {
        total: checkSummary.totalChecks,
        passed: checkSummary.passedChecks,
        failed: checkSummary.failedChecks,
      },
    };
  }

  /**
   * Print diagnostic header (simplified for clean output)
   * @tags @UTIL:PRINT-HEADER-001
   */
  private printHeader(): void {
    console.log(chalk.cyan('🔍 Checking system requirements...\n'));
  }

  /**
   * Categorize check results
   * @param results - Raw check results
   * @returns Categorized results
   * @tags @UTIL:CATEGORIZE-RESULTS-001
   */
  private categorizeResults(
    results: RequirementCheckResult[]
  ): CategorizedResults {
    const missing = results.filter(r => !r.result.isInstalled);
    const conflicts = results.filter(
      r => r.result.isInstalled && !r.result.versionSatisfied
    );
    const passed = results.filter(
      r => r.result.isInstalled && r.result.versionSatisfied
    );
    const allPassed = missing.length === 0 && conflicts.length === 0;

    return { missing, conflicts, passed, allPassed };
  }

  /**
   * Format individual check result
   * @param checkResult - Requirement check result
   * @returns Formatted string
   * @tags @UTIL:FORMAT-CHECK-001
   */
  public formatCheckResult(checkResult: RequirementCheckResult): string {
    const { requirement, result } = checkResult;

    if (result.isInstalled && result.versionSatisfied) {
      const version = result.detectedVersion || 'unknown';
      return `${chalk.green('✅')} ${chalk.bold(requirement.name)} ${chalk.gray(`(${version})`)}`;
    }

    if (result.isInstalled && !result.versionSatisfied) {
      const version = result.detectedVersion || 'unknown';
      const minVersion = requirement.minVersion || 'N/A';
      return `${chalk.yellow('⚠️ ')} ${chalk.bold(requirement.name)} ${chalk.gray(`(${version})`)} - ${chalk.yellow(`requires >= ${minVersion}`)}`;
    }

    const error = result.error || 'Not found';
    return `${chalk.red('❌')} ${chalk.bold(requirement.name)} - ${chalk.red(error)}`;
  }

  /**
   * Get installation suggestion for failed requirement
   * @param checkResult - Failed requirement check result
   * @returns Installation suggestion string
   * @tags @UTIL:INSTALL-SUGGESTION-001
   */
  public getInstallationSuggestion(
    checkResult: RequirementCheckResult
  ): string {
    const installCommand = this.detector.getInstallCommandForCurrentPlatform(
      checkResult.requirement
    );

    if (!installCommand) {
      return `${chalk.gray('Manual installation required for')} ${chalk.bold(checkResult.requirement.name)}`;
    }

    return `${chalk.blue('Install')} ${chalk.bold(checkResult.requirement.name)} ${chalk.blue('with:')} ${chalk.cyan(installCommand)}`;
  }

  /**
   * Print enhanced check results with language detection
   * @param checkSummary - System check summary
   * @tags @UTIL:PRINT-ENHANCED-RESULTS-001
   */
  private printEnhancedResults(checkSummary: SystemCheckSummary): void {
    // Show detected languages first if any
    if (checkSummary.detectedLanguages.length > 0) {
      console.log(chalk.cyan.bold('  Languages:'), chalk.white(checkSummary.detectedLanguages.join(', ')));
      console.log();
    }

    console.log(chalk.bold('  ⚙️  Runtime:'));
    checkSummary.runtime.forEach(result => {
      console.log(`    ${this.formatCheckResult(result)}`);
      if (!result.result.isInstalled || !result.result.versionSatisfied) {
        console.log(`      ${this.getInstallationSuggestion(result)}`);
      }
    });

    console.log();
    console.log(chalk.bold('  🛠️  Development:'));
    checkSummary.development.forEach(result => {
      console.log(`    ${this.formatCheckResult(result)}`);
      if (!result.result.isInstalled || !result.result.versionSatisfied) {
        console.log(`      ${this.getInstallationSuggestion(result)}`);
      }
    });

    // Show optional requirements if any
    if (checkSummary.optional.length > 0) {
      console.log();
      console.log(chalk.bold('  📦 Optional:'));
      checkSummary.optional.forEach(result => {
        console.log(`    ${this.formatCheckResult(result)}`);
        if (!result.result.isInstalled || !result.result.versionSatisfied) {
          console.log(
            `      ${chalk.gray(this.getInstallationSuggestion(result))}`
          );
        }
      });
    }

    console.log();
  }

  /**
   * Print enhanced summary with language info
   * @param checkSummary - System check summary
   * @tags @UTIL:PRINT-ENHANCED-SUMMARY-001
   */
  private printEnhancedSummary(checkSummary: SystemCheckSummary): void {
    console.log(chalk.gray('─'.repeat(60)));
    console.log(chalk.bold('  📊 Summary:'));
    console.log(chalk.gray(`     Checks: ${chalk.white(checkSummary.totalChecks)} total`));
    console.log(chalk.gray(`     Status: ${chalk.green(checkSummary.passedChecks + ' passed')} ${checkSummary.failedChecks > 0 ? chalk.red(checkSummary.failedChecks + ' failed') : ''}`));
    console.log(chalk.gray('─'.repeat(60)));

    if (checkSummary.passedChecks === checkSummary.totalChecks) {
      console.log(chalk.green.bold('\n✅ All requirements satisfied!\n'));
    } else {
      console.log(chalk.yellow.bold('\n⚠️  Some requirements need attention\n'));
      console.log(chalk.gray('Install missing tools as suggested above.\n'));
    }
  }

  /**
   * List available MoAI-ADK backups
   * @returns Doctor result with backup information
   * @tags @API:LIST-BACKUPS-001
   */
  private async listBackups(): Promise<DoctorResult> {
    logger.info(chalk.blue.bold('📦 MoAI-ADK Backup Directory Listing'));
    logger.info(chalk.blue('Searching for available backups...\n'));

    try {
      const backupPaths = await this.findBackupDirectories();

      if (backupPaths.length === 0) {
        logger.info(chalk.yellow('📁 No backup directories found.'));
        logger.info(
          chalk.gray('  Backup directories are typically created in:')
        );
        logger.info(chalk.gray('  • .moai-backup/ (current directory)'));
        logger.info(chalk.gray('  • ~/.moai/backups/ (global backups)'));
        logger.info('');
        logger.info(
          chalk.blue(
            '💡 Tip: Run "moai init --backup" to create a backup during initialization.'
          )
        );
      } else {
        logger.info(
          chalk.green(
            `📁 Found ${backupPaths.length} backup director${backupPaths.length === 1 ? 'y' : 'ies'}:`
          )
        );
        logger.info('');

        for (const backupPath of backupPaths) {
          await this.printBackupInfo(backupPath);
        }

        logger.info('');
        logger.info(
          chalk.blue(
            '💡 To restore from a backup, use: "moai restore <backup-path>"'
          )
        );
      }

      // Return a successful result for backup listing
      return {
        allPassed: true,
        results: [],
        missingRequirements: [],
        versionConflicts: [],
        summary: {
          total: backupPaths.length,
          passed: backupPaths.length,
          failed: 0,
        },
      };
    } catch (error) {
      logger.error(chalk.red('❌ Error scanning for backups:'), error);

      return {
        allPassed: false,
        results: [],
        missingRequirements: [],
        versionConflicts: [],
        summary: {
          total: 0,
          passed: 0,
          failed: 1,
        },
      };
    }
  }

  /**
   * Find backup directories in common locations
   * @returns Array of backup directory paths
   * @tags @UTIL:FIND-BACKUP-DIRS-001
   */
  private async findBackupDirectories(): Promise<string[]> {
    const backupPaths: string[] = [];
    const searchPaths = [
      path.join(process.cwd(), '.moai-backup'),
      path.join(process.cwd(), '.moai-backup'),
      path.join(os.homedir(), '.moai', 'backups'),
    ];

    for (const searchPath of searchPaths) {
      try {
        const exists = await this.directoryExists(searchPath);
        if (exists) {
          const subdirs = await this.getSubdirectories(searchPath);
          backupPaths.push(
            ...subdirs.map(subdir => path.join(searchPath, subdir))
          );
        }
      } catch {
        // Directory doesn't exist or can't be accessed
      }
    }

    return backupPaths.sort();
  }

  /**
   * Check if directory exists
   * @param dirPath - Directory path to check
   * @returns True if directory exists
   * @tags @UTIL:DIRECTORY-EXISTS-001
   */
  private async directoryExists(dirPath: string): Promise<boolean> {
    try {
      const stat = await fs.stat(dirPath);
      return stat.isDirectory();
    } catch {
      return false;
    }
  }

  /**
   * Get subdirectories in a directory
   * @param dirPath - Directory path
   * @returns Array of subdirectory names
   * @tags @UTIL:GET-SUBDIRECTORIES-001
   */
  private async getSubdirectories(dirPath: string): Promise<string[]> {
    try {
      const entries = await fs.readdir(dirPath, { withFileTypes: true });
      return entries
        .filter(entry => entry.isDirectory())
        .map(entry => entry.name)
        .filter(
          name => name.startsWith('backup-') || /^\d{4}-\d{2}-\d{2}/.test(name)
        );
    } catch {
      return [];
    }
  }

  /**
   * Print information about a backup directory
   * @param backupPath - Path to backup directory
   * @tags @UTIL:PRINT-BACKUP-INFO-001
   */
  private async printBackupInfo(backupPath: string): Promise<void> {
    try {
      const stat = await fs.stat(backupPath);
      const backupName = path.basename(backupPath);
      const backupDate = stat.mtime.toLocaleDateString();
      const backupTime = stat.mtime.toLocaleTimeString();

      logger.info(`  📦 ${chalk.bold(backupName)}`);
      logger.info(`     📍 Path: ${chalk.gray(backupPath)}`);
      logger.info(
        `     📅 Created: ${chalk.cyan(backupDate)} ${chalk.gray(backupTime)}`
      );

      // Check backup contents
      const contents = await this.getBackupContents(backupPath);
      if (contents.length > 0) {
        logger.info(`     📄 Contains: ${chalk.green(contents.join(', '))}`);
      }
      logger.info('');
    } catch (_error) {
      logger.info(`  ❌ ${chalk.red('Error reading backup:')} ${backupPath}`);
      logger.info('');
    }
  }

  /**
   * Get backup directory contents summary
   * @param backupPath - Path to backup directory
   * @returns Array of content descriptions
   * @tags @UTIL:GET-BACKUP-CONTENTS-001
   */
  private async getBackupContents(backupPath: string): Promise<string[]> {
    const contents: string[] = [];

    try {
      const entries = await fs.readdir(backupPath);

      if (entries.includes('.claude')) contents.push('Claude Code config');
      if (entries.includes('.moai')) contents.push('MoAI config');
      if (entries.includes('package.json')) contents.push('Package config');
      if (entries.includes('tsconfig.json')) contents.push('TypeScript config');
      if (entries.some(e => e.endsWith('.py'))) contents.push('Python files');
      if (entries.some(e => e.endsWith('.ts')))
        contents.push('TypeScript files');

      const totalFiles = entries.filter(e => !e.startsWith('.')).length;
      if (totalFiles > 0) {
        contents.push(`${totalFiles} files`);
      }
    } catch {
      // Can't read contents
    }

    return contents;
  }
}

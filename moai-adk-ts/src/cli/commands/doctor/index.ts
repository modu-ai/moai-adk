// @CODE:DOCTOR-001 | Chain: @SPEC:DOCTOR-001 -> @SPEC:DOCTOR-001 -> @CODE:DOCTOR-001
// Related: @CODE:SYS-001:API, @CODE:SYS-INFO-001, @TEST:DOCTOR-001

/**
 * @file CLI doctor command orchestrator for system diagnostics
 * @author MoAI Team
 */

import * as fs from 'node:fs/promises';
import * as path from 'node:path';
import chalk from 'chalk';
import {
  type RequirementCheckResult,
  SystemChecker,
  type SystemCheckSummary,
  type SystemDetector,
} from '@/core/system-checker';
import { logger } from '../../../utils/winston-logger.js';
import { BackupChecker } from './checkers/backup-checker.js';
import { ResultFormatter } from './reporters/result-formatter.js';
import { SummaryReporter } from './reporters/summary-reporter.js';

/**
 * Doctor command result summary
 * @tags @SPEC:DOCTOR-RESULT-001
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
 * @tags @SPEC:CATEGORIZED-RESULTS-001
 */
type CategorizedResults = {
  readonly missing: RequirementCheckResult[];
  readonly conflicts: RequirementCheckResult[];
  readonly passed: RequirementCheckResult[];
  readonly allPassed: boolean;
};

/**
 * Doctor command orchestrator for system diagnostics
 * @tags @CODE:DOCTOR-001
 */
export class DoctorCommand {
  private readonly systemChecker = new SystemChecker();
  private readonly backupChecker = new BackupChecker();
  private readonly formatter = new ResultFormatter();
  private readonly reporter = new SummaryReporter();

  constructor(private readonly detector: SystemDetector) {}

  /**
   * Run system diagnostics with language detection
   * @param options - Doctor command options
   * @returns Doctor result with all checks
   * @tags @CODE:DOCTOR-RUN-001:API
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
    this.reporter.printEnhancedSummary(checkSummary);

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
   * Print diagnostic header
   * @tags UTIL:PRINT-HEADER-001
   */
  private printHeader(): void {
    console.log(chalk.cyan('🔍 Checking system requirements...\n'));
  }

  /**
   * Categorize check results
   * @param results - Raw check results
   * @returns Categorized results
   * @tags UTIL:CATEGORIZE-RESULTS-001
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
   * Format individual check result (delegated to formatter)
   * @param checkResult - Requirement check result
   * @returns Formatted string
   * @tags @CODE:FORMAT-CHECK-001:API
   */
  public formatCheckResult(checkResult: RequirementCheckResult): string {
    return this.formatter.formatCheckResult(checkResult);
  }

  /**
   * Get installation suggestion (delegated to formatter)
   * @param checkResult - Failed requirement check result
   * @returns Installation suggestion string
   * @tags @CODE:INSTALL-SUGGESTION-001:API
   */
  public getInstallationSuggestion(
    checkResult: RequirementCheckResult
  ): string {
    const installCommand = this.detector.getInstallCommandForCurrentPlatform(
      checkResult.requirement
    );
    return this.formatter.getInstallationSuggestion(
      checkResult,
      installCommand ?? null
    );
  }

  /**
   * Print enhanced check results with language detection
   * @param checkSummary - System check summary
   * @tags UTIL:PRINT-ENHANCED-RESULTS-001
   */
  private printEnhancedResults(checkSummary: SystemCheckSummary): void {
    const results = {
      runtime: checkSummary.runtime,
      development: checkSummary.development,
      optional: checkSummary.optional,
      detectedLanguages: checkSummary.detectedLanguages,
    };

    this.reporter.printEnhancedResults(results, this.formatter);

    // Print installation suggestions for failed checks
    const allChecks = [
      ...checkSummary.runtime,
      ...checkSummary.development,
      ...checkSummary.optional,
    ];

    allChecks.forEach(result => {
      if (!result.result.isInstalled || !result.result.versionSatisfied) {
        const suggestion = this.getInstallationSuggestion(result);
        console.log(`      ${suggestion}`);
      }
    });
  }

  /**
   * List available MoAI-ADK backups
   * @returns Doctor result with backup information
   * @tags @CODE:LIST-BACKUPS-001:API
   */
  private async listBackups(): Promise<DoctorResult> {
    logger.log(chalk.blue.bold('📦 MoAI-ADK Backup Directory Listing'));
    logger.log(chalk.blue('Searching for available backups...\n'));

    try {
      const backupPaths = await this.backupChecker.findBackupDirectories();

      if (backupPaths.length === 0) {
        logger.log(chalk.yellow('📁 No backup directories found.'));
        logger.log(
          chalk.gray('  Backup directories are typically created in:')
        );
        logger.log(chalk.gray('  • .moai-backup/ (current directory)'));
        logger.log(chalk.gray('  • ~/.moai/backups/ (global backups)'));
        logger.log('');
        logger.log(
          chalk.blue(
            '💡 Tip: Run "moai init --backup" to create a backup during initialization.'
          )
        );
      } else {
        logger.log(
          chalk.green(
            `📁 Found ${backupPaths.length} backup director${backupPaths.length === 1 ? 'y' : 'ies'}:`
          )
        );
        logger.log('');

        for (const backupPath of backupPaths) {
          await this.printBackupInfo(backupPath);
        }

        logger.log('');
        logger.log(
          chalk.blue(
            '💡 To restore from a backup, use: "moai restore <backup-path>"'
          )
        );
      }

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
   * Print information about a backup directory
   * @param backupPath - Path to backup directory
   * @tags UTIL:PRINT-BACKUP-INFO-001
   */
  private async printBackupInfo(backupPath: string): Promise<void> {
    try {
      const stat = await fs.stat(backupPath);
      const backupName = path.basename(backupPath);
      const backupDate = stat.mtime.toLocaleDateString();
      const backupTime = stat.mtime.toLocaleTimeString();

      logger.log(`  📦 ${chalk.bold(backupName)}`);
      logger.log(`     📍 Path: ${chalk.gray(backupPath)}`);
      logger.log(
        `     📅 Created: ${chalk.cyan(backupDate)} ${chalk.gray(backupTime)}`
      );

      // Check backup contents using BackupChecker
      const contents = await this.backupChecker.getBackupContents(backupPath);
      if (contents.length > 0) {
        logger.log(`     📄 Contains: ${chalk.green(contents.join(', '))}`);
      }
      logger.log('');
    } catch (_error) {
      logger.log(`  ❌ ${chalk.red('Error reading backup:')} ${backupPath}`);
      logger.log('');
    }
  }
}

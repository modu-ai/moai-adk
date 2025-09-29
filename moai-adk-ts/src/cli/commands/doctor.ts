/**
 * @file CLI doctor command implementation
 * @author MoAI Team
 * @tags @FEATURE:CLI-DOCTOR-001 @REQ:CLI-FOUNDATION-012
 */

import chalk from 'chalk';
import * as fs from 'fs/promises';
import * as path from 'path';
import * as os from 'os';
import type {
  SystemDetector,
  RequirementCheckResult,
} from '@/core/system-checker/detector';
import {
  requirementRegistry,
  type SystemRequirement,
} from '@/core/system-checker/requirements';

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
 * Doctor command for system diagnostics
 * @tags @FEATURE:CLI-DOCTOR-001
 */
export class DoctorCommand {
  constructor(private readonly detector: SystemDetector) {}

  /**
   * Run system diagnostics
   * @param options - Doctor command options
   * @returns Doctor result with all checks
   * @tags @API:DOCTOR-RUN-001
   */
  public async run(options: { listBackups?: boolean } = {}): Promise<DoctorResult> {
    // Handle --list-backups option
    if (options.listBackups) {
      return this.listBackups();
    }

    this.printHeader();

    const requirements = this.gatherRequirements();
    const results = await this.executeChecks(requirements);
    const categorizedResults = this.categorizeResults(results);

    this.printResults(results);
    const summary = this.generateSummary(categorizedResults);
    this.printSummary(summary, categorizedResults.allPassed);

    return {
      allPassed: categorizedResults.allPassed,
      results,
      missingRequirements: categorizedResults.missing,
      versionConflicts: categorizedResults.conflicts,
      summary,
    };
  }

  /**
   * Print diagnostic header
   * @tags @UTIL:PRINT-HEADER-001
   */
  private printHeader(): void {
    console.log(chalk.blue.bold('🔍 MoAI-ADK System Diagnostics'));
    console.log(chalk.blue('Checking system requirements...\n'));
  }

  /**
   * Gather all system requirements
   * @returns Array of system requirements
   * @tags @UTIL:GATHER-REQUIREMENTS-001
   */
  private gatherRequirements(): SystemRequirement[] {
    const runtimeRequirements = requirementRegistry.getByCategory('runtime');
    const developmentRequirements =
      requirementRegistry.getByCategory('development');
    return [...runtimeRequirements, ...developmentRequirements];
  }

  /**
   * Execute requirement checks
   * @param requirements - Requirements to check
   * @returns Check results
   * @tags @UTIL:EXECUTE-CHECKS-001
   */
  private async executeChecks(
    requirements: SystemRequirement[]
  ): Promise<RequirementCheckResult[]> {
    try {
      return await this.detector.checkMultipleRequirements(requirements);
    } catch (error) {
      console.error(chalk.red('❌ Failed to execute system checks:'), error);
      throw new Error('System diagnostics failed');
    }
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
   * Generate summary statistics
   * @param categorized - Categorized results
   * @returns Summary object
   * @tags @UTIL:GENERATE-SUMMARY-001
   */
  private generateSummary(
    categorized: CategorizedResults
  ): DoctorResult['summary'] {
    return {
      total:
        categorized.missing.length +
        categorized.conflicts.length +
        categorized.passed.length,
      passed: categorized.passed.length,
      failed: categorized.missing.length + categorized.conflicts.length,
    };
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
   * Print all check results
   * @param results - Array of check results
   * @tags @UTIL:PRINT-RESULTS-001
   */
  private printResults(results: RequirementCheckResult[]): void {
    console.log(chalk.bold('Runtime Requirements:'));
    const runtimeResults = results.filter(
      r => r.requirement.category === 'runtime'
    );
    runtimeResults.forEach(result => {
      console.log(`  ${this.formatCheckResult(result)}`);
      if (!result.result.isInstalled || !result.result.versionSatisfied) {
        console.log(`    ${this.getInstallationSuggestion(result)}`);
      }
    });

    console.log('');
    console.log(chalk.bold('Development Requirements:'));
    const devResults = results.filter(
      r => r.requirement.category === 'development'
    );
    devResults.forEach(result => {
      console.log(`  ${this.formatCheckResult(result)}`);
      if (!result.result.isInstalled || !result.result.versionSatisfied) {
        console.log(`    ${this.getInstallationSuggestion(result)}`);
      }
    });
    console.log('');
  }

  /**
   * Print summary of checks
   * @param summary - Check summary
   * @param allPassed - Whether all checks passed
   * @tags @UTIL:PRINT-SUMMARY-001
   */
  private printSummary(
    summary: { total: number; passed: number; failed: number },
    allPassed: boolean
  ): void {
    console.log(chalk.bold('Summary:'));
    console.log(`  Total checks: ${summary.total}`);
    console.log(`  ${chalk.green('Passed:')} ${summary.passed}`);
    console.log(`  ${chalk.red('Failed:')} ${summary.failed}`);
    console.log('');

    if (allPassed) {
      console.log(chalk.green.bold('✅ All system requirements satisfied!'));
    } else {
      console.log(
        chalk.red.bold('❌ Some system requirements need attention.')
      );
      console.log(
        chalk.yellow(
          'Please install missing tools or upgrade versions as suggested above.'
        )
      );
    }
  }

  /**
   * List available MoAI-ADK backups
   * @returns Doctor result with backup information
   * @tags @API:LIST-BACKUPS-001
   */
  private async listBackups(): Promise<DoctorResult> {
    console.log(chalk.blue.bold('📦 MoAI-ADK Backup Directory Listing'));
    console.log(chalk.blue('Searching for available backups...\n'));

    try {
      const backupPaths = await this.findBackupDirectories();

      if (backupPaths.length === 0) {
        console.log(chalk.yellow('📁 No backup directories found.'));
        console.log(chalk.gray('  Backup directories are typically created in:'));
        console.log(chalk.gray('  • .moai-backup/ (current directory)'));
        console.log(chalk.gray('  • ~/.moai/backups/ (global backups)'));
        console.log('');
        console.log(chalk.blue('💡 Tip: Run "moai init --backup" to create a backup during initialization.'));
      } else {
        console.log(chalk.green(`📁 Found ${backupPaths.length} backup director${backupPaths.length === 1 ? 'y' : 'ies'}:`));
        console.log('');

        for (const backupPath of backupPaths) {
          await this.printBackupInfo(backupPath);
        }

        console.log('');
        console.log(chalk.blue('💡 To restore from a backup, use: "moai restore <backup-path>"'));
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
      console.error(chalk.red('❌ Error scanning for backups:'), error);

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
          backupPaths.push(...subdirs.map(subdir => path.join(searchPath, subdir)));
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
        .filter(name => name.startsWith('backup-') || /^\d{4}-\d{2}-\d{2}/.test(name));
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

      console.log(`  📦 ${chalk.bold(backupName)}`);
      console.log(`     📍 Path: ${chalk.gray(backupPath)}`);
      console.log(`     📅 Created: ${chalk.cyan(backupDate)} ${chalk.gray(backupTime)}`);

      // Check backup contents
      const contents = await this.getBackupContents(backupPath);
      if (contents.length > 0) {
        console.log(`     📄 Contains: ${chalk.green(contents.join(', '))}`);
      }
      console.log('');
    } catch (error) {
      console.log(`  ❌ ${chalk.red('Error reading backup:')} ${backupPath}`);
      console.log('');
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
      if (entries.some(e => e.endsWith('.ts'))) contents.push('TypeScript files');

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

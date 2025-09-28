/**
 * @file CLI doctor command implementation
 * @author MoAI Team
 * @tags @FEATURE:CLI-DOCTOR-001 @REQ:CLI-FOUNDATION-012
 */

import chalk from 'chalk';
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
   * @returns Doctor result with all checks
   * @tags @API:DOCTOR-RUN-001
   */
  public async run(): Promise<DoctorResult> {
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
}

/**
 * @file CLI doctor command implementation
 * @author MoAI Team
 * @tags @FEATURE:CLI-DOCTOR-001 @REQ:CLI-FOUNDATION-012
 */

import chalk from 'chalk';
import {
  SystemDetector,
  RequirementCheckResult,
} from '@/core/system-checker/detector';
import { requirementRegistry } from '@/core/system-checker/requirements';

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
    console.log(chalk.blue.bold('🔍 MoAI-ADK System Diagnostics'));
    console.log(chalk.blue('Checking system requirements...\n'));

    // Get runtime requirements
    const runtimeRequirements = requirementRegistry.getByCategory('runtime');
    const developmentRequirements =
      requirementRegistry.getByCategory('development');
    const allRequirements = [
      ...runtimeRequirements,
      ...developmentRequirements,
    ];

    // Check all requirements
    const results =
      await this.detector.checkMultipleRequirements(allRequirements);

    // Categorize results
    const missingRequirements = results.filter(r => !r.result.isInstalled);
    const versionConflicts = results.filter(
      r => r.result.isInstalled && !r.result.versionSatisfied
    );
    const passedChecks = results.filter(
      r => r.result.isInstalled && r.result.versionSatisfied
    );

    // Print results
    this.printResults(results);

    // Generate summary
    const allPassed =
      missingRequirements.length === 0 && versionConflicts.length === 0;
    const summary = {
      total: results.length,
      passed: passedChecks.length,
      failed: missingRequirements.length + versionConflicts.length,
    };

    // Print summary
    this.printSummary(summary, allPassed);

    return {
      allPassed,
      results,
      missingRequirements,
      versionConflicts,
      summary,
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

    return `${chalk.blue('Install with:')} ${chalk.cyan(installCommand)}`;
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

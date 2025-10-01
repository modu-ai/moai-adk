// @CODE:SUMMARY-REPORTER-001 | Chain: @SPEC:DOCTOR-001 -> @SPEC:DOCTOR-001 -> @CODE:DOCTOR-001
// Related: @TEST:REPORTERS-001

/**
 * @file Summary reporter for doctor command diagnostics
 * @author MoAI Team
 */

import chalk from 'chalk';
import type { SystemCheckSummary } from '@/core/system-checker';
import type { ResultFormatter } from './result-formatter.js';

/**
 * Results to display
 */
interface CheckResults {
  readonly runtime: readonly any[];
  readonly development: readonly any[];
  readonly optional: readonly any[];
  readonly detectedLanguages: readonly string[];
}

/**
 * SummaryReporter displays enhanced check results and summary
 * @tags @CODE:SUMMARY-REPORTER-001
 */
export class SummaryReporter {
  /**
   * Print enhanced check results with language detection
   * @param results - Categorized check results
   * @param formatter - Result formatter instance
   * @tags @CODE:PRINT-RESULTS-001:API
   */
  public printEnhancedResults(
    results: CheckResults,
    formatter: ResultFormatter
  ): void {
    // Show detected languages first if any
    if (results.detectedLanguages.length > 0) {
      console.log(
        chalk.cyan.bold('  Languages:'),
        chalk.white(results.detectedLanguages.join(', '))
      );
      console.log();
    }

    console.log(chalk.bold('  ⚙️  Runtime:'));
    results.runtime.forEach(result => {
      console.log(`    ${formatter.formatCheckResult(result)}`);
    });

    console.log();
    console.log(chalk.bold('  🛠️  Development:'));
    results.development.forEach(result => {
      console.log(`    ${formatter.formatCheckResult(result)}`);
    });

    // Show optional requirements if any
    if (results.optional.length > 0) {
      console.log();
      console.log(chalk.bold('  📦 Optional:'));
      results.optional.forEach(result => {
        console.log(`    ${formatter.formatCheckResult(result)}`);
      });
    }

    console.log();
  }

  /**
   * Print enhanced summary with language info
   * @param checkSummary - System check summary
   * @tags @CODE:PRINT-SUMMARY-001:API
   */
  public printEnhancedSummary(checkSummary: SystemCheckSummary): void {
    console.log(chalk.gray('─'.repeat(60)));
    console.log(chalk.bold('  📊 Summary:'));
    console.log(
      chalk.gray(`     Checks: ${chalk.white(checkSummary.totalChecks)} total`)
    );
    console.log(
      chalk.gray(
        `     Status: ${chalk.green(`${checkSummary.passedChecks} passed`)} ${checkSummary.failedChecks > 0 ? chalk.red(`${checkSummary.failedChecks} failed`) : ''}`
      )
    );
    console.log(chalk.gray('─'.repeat(60)));

    if (checkSummary.passedChecks === checkSummary.totalChecks) {
      console.log(chalk.green.bold('\n✅ All requirements satisfied!\n'));
    } else {
      console.log(chalk.yellow.bold('\n⚠️  Some requirements need attention\n'));
      console.log(chalk.gray('Install missing tools as suggested above.\n'));
    }
  }
}

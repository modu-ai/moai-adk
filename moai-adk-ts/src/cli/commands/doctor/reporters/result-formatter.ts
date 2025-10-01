// @CODE:RESULT-FORMATTER-001 | Chain: @SPEC:DOCTOR-001 -> @SPEC:DOCTOR-001 -> @CODE:DOCTOR-001
// Related: @TEST:REPORTERS-001

/**
 * @file Check result formatter for doctor command
 * @author MoAI Team
 */

import chalk from 'chalk';
import type { RequirementCheckResult } from '@/core/system-checker';

/**
 * ResultFormatter formats individual check results for display
 * @tags @CODE:RESULT-FORMATTER-001
 */
export class ResultFormatter {
  /**
   * Format individual check result with colors and symbols
   * @param checkResult - Requirement check result
   * @returns Formatted string for display
   * @tags @CODE:FORMAT-CHECK-001:API
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
   * @param installCommand - Platform-specific install command or null
   * @returns Installation suggestion string
   * @tags @CODE:INSTALL-SUGGESTION-001:API
   */
  public getInstallationSuggestion(
    checkResult: RequirementCheckResult,
    installCommand: string | null
  ): string {
    if (!installCommand) {
      return `${chalk.gray('Manual installation required for')} ${chalk.bold(checkResult.requirement.name)}`;
    }

    return `${chalk.blue('Install')} ${chalk.bold(checkResult.requirement.name)} ${chalk.blue('with:')} ${chalk.cyan(installCommand)}`;
  }
}

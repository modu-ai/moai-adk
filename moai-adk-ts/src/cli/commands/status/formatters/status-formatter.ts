// @CODE:CLI-FMT-001 | SPEC: status.ts 리팩토링
// Related: @CODE:CLI-003

/**
 * @file Status output formatting functionality
 * @author MoAI Team
 */

import chalk from 'chalk';
import { logger } from '../../../../utils/winston-logger.js';
import type { FileCount } from '../collectors/file-counter.js';
import type { ProjectStatus } from '../collectors/status-collector.js';
import type { VersionInfo } from '../collectors/version-collector.js';

/**
 * Status formatter for console output
 * @tags @CODE:STATUS-FORMATTER-001
 */
export class StatusFormatter {
  /**
   * Display project status information
   * @param status - Project status
   * @tags @CODE:FORMAT-STATUS-001:API
   */
  public displayStatus(status: ProjectStatus): void {
    logger.log(chalk.cyan('📊 MoAI-ADK Project Status'));
    logger.log(`\n📂 Project: ${status.path}`);
    logger.log(`   Type: ${status.projectType}`);

    logger.log('\n🗿 MoAI-ADK Components:');
    logger.log(`   MoAI System: ${status.moaiInitialized ? '✅' : '❌'}`);
    logger.log(
      `   Claude Integration: ${status.claudeInitialized ? '✅' : '❌'}`
    );
    logger.log(`   Memory File: ${status.memoryFile ? '✅' : '❌'}`);
    logger.log(`   Git Repository: ${status.gitRepository ? '✅' : '❌'}`);
  }

  /**
   * Display version information
   * @param versions - Version info
   * @tags @CODE:FORMAT-VERSION-001:API
   */
  public displayVersions(versions: VersionInfo): void {
    logger.log('\n🧭 Versions:');
    logger.log(`   Package: v${versions.package}`);
    logger.log(`   Templates: v${versions.resources}`);

    if (versions.available && versions.available !== versions.resources) {
      logger.log(`   Available template update: v${versions.available}`);
    }

    if (versions.outdated) {
      logger.log(
        chalk.yellow(
          "   ⚠️  Templates are outdated. Run 'moai update' to refresh."
        )
      );
    }
  }

  /**
   * Display file counts
   * @param fileCounts - File count information
   * @tags @CODE:FORMAT-FILES-001:API
   */
  public displayFileCounts(fileCounts: FileCount): void {
    logger.log('\n📁 File Counts:');

    for (const [component, count] of Object.entries(fileCounts)) {
      if (component !== 'total') {
        logger.log(`   ${component}: ${count} files`);
      }
    }
  }

  /**
   * Generate recommendations based on status
   * @param status - Project status
   * @returns Recommendations list
   * @tags @CODE:GENERATE-RECOMMENDATIONS-001:API
   */
  public generateRecommendations(status: ProjectStatus): string[] {
    const recommendations: string[] = [];

    if (!status.moaiInitialized) {
      recommendations.push("Run 'moai init' to initialize MoAI-ADK");
    }
    if (!status.gitRepository) {
      recommendations.push('Initialize Git repository: git init');
    }

    return recommendations;
  }

  /**
   * Display recommendations
   * @param recommendations - Recommendations list
   * @tags @CODE:FORMAT-RECOMMENDATIONS-001:API
   */
  public displayRecommendations(recommendations: string[]): void {
    if (recommendations.length > 0) {
      logger.log('\n💡 Recommendations:');
      for (const rec of recommendations) {
        logger.log(`   - ${rec}`);
      }
    }
  }
}

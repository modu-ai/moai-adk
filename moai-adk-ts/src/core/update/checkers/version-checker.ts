// @CODE:UPD-VER-001 | SPEC: update-orchestrator.ts 리팩토링
// Related: @CODE:UPD-001

/**
 * @file Version checking functionality
 * @author MoAI Team
 */

import chalk from 'chalk';
import { checkLatestVersion, getCurrentVersion } from '../../../utils/version.js';
import { logger } from '../../../utils/winston-logger.js';

/**
 * Version check result (re-export with consistent naming)
 * @tags @SPEC:VERSION-CHECK-RESULT-001
 */
export interface VersionCheckResult {
  readonly currentVersion: string;
  readonly latestVersion: string | null;
  readonly hasUpdate: boolean;
}

/**
 * Version checker for MoAI-ADK updates
 * @tags @CODE:VERSION-CHECKER-001
 */
export class VersionChecker {
  /**
   * Check for available updates
   * @returns Version check result
   * @tags @CODE:CHECK-UPDATES-001:API
   */
  public async checkForUpdates(): Promise<VersionCheckResult> {
    logger.log(chalk.cyan('🔍 MoAI-ADK 업데이트 확인 중...'));

    const currentVersion = getCurrentVersion();
    const versionCheck = await checkLatestVersion();

    logger.log(chalk.blue(`📦 현재 버전: v${currentVersion}`));

    if (!versionCheck.hasUpdate || !versionCheck.latest) {
      logger.log(chalk.green('✅ 최신 버전을 사용 중입니다'));
      return {
        currentVersion,
        latestVersion: versionCheck.latest,
        hasUpdate: false,
      };
    }

    logger.log(chalk.yellow(`⚡ 최신 버전: v${versionCheck.latest}`));
    logger.log(chalk.green('✅ 업데이트 가능'));

    return {
      currentVersion,
      latestVersion: versionCheck.latest,
      hasUpdate: true,
    };
  }
}

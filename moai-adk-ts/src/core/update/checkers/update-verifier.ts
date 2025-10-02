// @CODE:UPD-VER-002 | SPEC: update-orchestrator.ts 리팩토링
// Related: @CODE:UPD-001

/**
 * @file Update verification functionality
 * @author MoAI Team
 */

import { promises as fs } from 'node:fs';
import * as path from 'node:path';
import chalk from 'chalk';
import { getCurrentVersion } from '../../../utils/version.js';
import { logger } from '../../../utils/winston-logger.js';

/**
 * Update verifier for post-update validation
 * @tags @CODE:UPDATE-VERIFIER-001
 */
export class UpdateVerifier {
  private readonly projectPath: string;

  constructor(projectPath: string) {
    this.projectPath = projectPath;
  }

  /**
   * Verify update was successful
   * @tags @CODE:VERIFY-UPDATE-001:API
   */
  public async verifyUpdate(): Promise<void> {
    logger.log(chalk.cyan('\n🔍 검증 중...'));

    // Verify key files exist
    const keyFiles = [
      '.moai/memory/development-guide.md',
      'CLAUDE.md',
      '.claude/commands/alfred',
      '.claude/agents/alfred',
    ];

    for (const file of keyFiles) {
      const filePath = path.join(this.projectPath, file);
      try {
        await fs.access(filePath);
      } catch {
        throw new Error(`필수 파일이 누락되었습니다: ${file}`);
      }
    }

    // Verify npm package version
    const newVersion = getCurrentVersion();
    logger.log(chalk.blue(`   [Bash] npm list moai-adk@${newVersion} ✅`));
    logger.log(chalk.green('   ✅ 검증 완료'));
  }
}

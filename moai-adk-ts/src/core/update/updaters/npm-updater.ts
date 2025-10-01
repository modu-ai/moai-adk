// @CODE:UPD-NPM-001 | SPEC: update-orchestrator.ts 리팩토링
// Related: @CODE:UPD-001

/**
 * @file npm package update functionality
 * @author MoAI Team
 */

import { promises as fs } from 'node:fs';
import * as path from 'node:path';
import chalk from 'chalk';
import { execa } from 'execa';
import { logger } from '../../../utils/winston-logger.js';

/**
 * npm updater for MoAI-ADK package
 * @tags @CODE:NPM-UPDATER-001
 */
export class NpmUpdater {
  private readonly projectPath: string;

  constructor(projectPath: string) {
    this.projectPath = projectPath;
  }

  /**
   * Update npm package to latest version
   * @tags @CODE:UPDATE-PACKAGE-001:API
   */
  public async updatePackage(): Promise<void> {
    logger.log(chalk.cyan('\n📦 패키지 업데이트 중...'));

    const packageJsonPath = path.join(this.projectPath, 'package.json');

    try {
      await fs.access(packageJsonPath);
      // Local installation
      await execa('npm', ['install', 'moai-adk@latest'], {
        cwd: this.projectPath,
      });
    } catch {
      // Global installation
      await execa('npm', ['install', '-g', 'moai-adk@latest']);
    }

    logger.log(chalk.green('   ✅ 패키지 업데이트 완료'));
  }

  /**
   * Get npm root directory
   * @returns npm root path
   * @tags @CODE:GET-NPM-ROOT-001:API
   */
  public async getNpmRoot(): Promise<string> {
    logger.log(chalk.cyan('\n📁 패키지 경로 확인 중...'));

    try {
      // Try local first
      const { stdout } = await execa('npm', ['root'], {
        cwd: this.projectPath,
      });
      const npmRoot = stdout.trim();
      logger.log(chalk.blue(`   npm root → ${npmRoot}`));
      return npmRoot;
    } catch {
      // Try global
      const { stdout } = await execa('npm', ['root', '-g']);
      const npmRoot = stdout.trim();
      logger.log(chalk.blue(`   npm root → ${npmRoot}`));
      return npmRoot;
    }
  }
}

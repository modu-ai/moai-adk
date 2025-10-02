// @CODE:UPDATE-REFACTOR-001 | SPEC: SPEC-UPDATE-REFACTOR-001.md
// Related: @CODE:UPDATE-REFACTOR-001:ALFRED-BRIDGE

/**
 * @file Alfred Update Bridge Utilities
 * @author MoAI Team
 * @description File operation utilities for Alfred Update Bridge
 * @tags @CODE:UPDATE-REFACTOR-001:FILE-UTILS
 */

import { promises as fs } from 'node:fs';
import * as path from 'node:path';
import chalk from 'chalk';
import { logger } from '../../../utils/winston-logger.js';

/**
 * 파일 백업
 * @param filePath - File to backup
 * @tags UTIL:BACKUP-FILE
 */
export async function backupFile(filePath: string): Promise<void> {
  const backup = `${filePath}.backup-${Date.now()}`;
  await fs.copyFile(filePath, backup);
  logger.log(chalk.gray(`   → 백업: ${path.basename(backup)}`));
}

/**
 * 디렉토리 재귀 복사
 * @param source - Source directory
 * @param target - Target directory
 * @returns Number of files copied
 * @tags UTIL:COPY-DIRECTORY
 */
export async function copyDirectory(
  source: string,
  target: string
): Promise<number> {
  await fs.mkdir(target, { recursive: true });
  const entries = await fs.readdir(source, { withFileTypes: true });
  let count = 0;

  for (const entry of entries) {
    const srcPath = path.join(source, entry.name);
    const tgtPath = path.join(target, entry.name);

    if (entry.isDirectory()) {
      count += await copyDirectory(srcPath, tgtPath);
    } else {
      await fs.copyFile(srcPath, tgtPath);
      count++;
    }
  }

  return count;
}

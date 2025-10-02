// @CODE:UPD-TPL-001 | SPEC: update-orchestrator.ts 리팩토링
// Related: @CODE:UPD-001

/**
 * @file Template file copying functionality
 * @author MoAI Team
 */

import { promises as fs } from 'node:fs';
import * as path from 'node:path';
import chalk from 'chalk';
import { logger } from '../../../utils/winston-logger.js';

/**
 * Template copier for update operations
 * @tags @CODE:TEMPLATE-COPIER-001
 */
export class TemplateCopier {
  private readonly projectPath: string;

  constructor(projectPath: string) {
    this.projectPath = projectPath;
  }

  /**
   * Copy template files to project
   * @param templatePath - Template directory path
   * @returns Number of files copied
   * @tags @CODE:COPY-TEMPLATES-001:API
   */
  public async copyTemplates(templatePath: string): Promise<number> {
    // Verify template directory exists
    try {
      await fs.access(templatePath);
    } catch {
      throw new Error(`템플릿 디렉토리를 찾을 수 없습니다: ${templatePath}`);
    }

    logger.log(chalk.green(`   ✅ 템플릿 경로: ${templatePath}`));
    logger.log(chalk.cyan('\n📄 템플릿 파일 복사 중...'));

    let filesCopied = 0;

    const filesToCopy = [
      { src: '.claude/commands/alfred', dest: '.claude/commands/alfred' },
      { src: '.claude/agents/alfred', dest: '.claude/agents/alfred' },
      { src: '.claude/hooks/alfred', dest: '.claude/hooks/alfred' },
      {
        src: '.moai/memory/development-guide.md',
        dest: '.moai/memory/development-guide.md',
      },
      { src: '.moai/project/product.md', dest: '.moai/project/product.md' },
      {
        src: '.moai/project/structure.md',
        dest: '.moai/project/structure.md',
      },
      { src: '.moai/project/tech.md', dest: '.moai/project/tech.md' },
      { src: 'CLAUDE.md', dest: 'CLAUDE.md' },
    ];

    for (const { src, dest } of filesToCopy) {
      const sourcePath = path.join(templatePath, src);
      const targetPath = path.join(this.projectPath, dest);

      try {
        const stat = await fs.stat(sourcePath);

        if (stat.isDirectory()) {
          // Copy directory
          await this.copyDirectory(sourcePath, targetPath);
          const files = await this.countFiles(sourcePath);
          filesCopied += files;
        } else {
          // Copy file
          await fs.mkdir(path.dirname(targetPath), { recursive: true });
          await fs.copyFile(sourcePath, targetPath);
          filesCopied++;
        }
      } catch (error) {
        logger.log(
          chalk.yellow(
            `   ⚠️  건너뛰기: ${src} (${error instanceof Error ? error.message : '알 수 없는 오류'})`
          )
        );
      }
    }

    logger.log(chalk.green(`   ✅ ${filesCopied}개 파일 복사 완료`));
    return filesCopied;
  }

  /**
   * Copy directory recursively
   * @param source - Source directory
   * @param target - Target directory
   * @tags UTIL:COPY-DIRECTORY-002
   */
  private async copyDirectory(source: string, target: string): Promise<void> {
    await fs.mkdir(target, { recursive: true });

    const entries = await fs.readdir(source, { withFileTypes: true });

    for (const entry of entries) {
      const sourcePath = path.join(source, entry.name);
      const targetPath = path.join(target, entry.name);

      if (entry.isDirectory()) {
        await this.copyDirectory(sourcePath, targetPath);
      } else {
        await fs.copyFile(sourcePath, targetPath);
      }
    }
  }

  /**
   * Count files in directory recursively
   * @param dirPath - Directory path
   * @returns File count
   * @tags UTIL:COUNT-FILES-002
   */
  private async countFiles(dirPath: string): Promise<number> {
    let count = 0;
    const entries = await fs.readdir(dirPath, { withFileTypes: true });

    for (const entry of entries) {
      if (entry.isDirectory()) {
        count += await this.countFiles(path.join(dirPath, entry.name));
      } else {
        count++;
      }
    }

    return count;
  }
}

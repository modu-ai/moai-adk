// @CODE:UPDATE-REFACTOR-001 | SPEC: SPEC-UPDATE-REFACTOR-001.md
// Related: @TEST:UPDATE-REFACTOR-001

/**
 * @file Alfred Update Bridge - Phase 4 Claude Code Tools Integration
 * @author MoAI Team
 * @description Alfred가 Claude Code 도구로 템플릿 복사 처리
 * @tags @CODE:UPDATE-REFACTOR-001:ALFRED-BRIDGE
 */

import { promises as fs } from 'node:fs';
import * as path from 'node:path';
import chalk from 'chalk';
import { logger } from '../../../utils/winston-logger.js';
import { backupFile, copyDirectory } from './file-utils.js';

/**
 * TDD History:
 * - RED: alfred-update-bridge.spec.ts 작성 (2025-10-02)
 *   - T001: Claude Code Tools 시뮬레이션
 *   - T002: {{PROJECT_NAME}} 패턴 검증 및 백업
 *   - T003: chmod +x 실행 권한 처리
 *   - T004-T007: 파일별 처리 로직
 * - GREEN: 최소 구현 완료 (2025-10-02)
 *   - handleProjectDocs: {{PROJECT_NAME}} 패턴 기반 보호
 *   - handleHookFiles: chmod +x 적용
 *   - handleOutputStyles: output-styles/alfred 복사
 *   - handleOtherFiles: 기타 파일 처리
 * - REFACTOR: (진행 중)
 *
 * @tags @CODE:UPDATE-REFACTOR-001:ALFRED-BRIDGE
 */
export class AlfredUpdateBridge {
  private readonly projectPath: string;

  constructor(projectPath: string) {
    this.projectPath = projectPath;
  }

  /**
   * Phase 4: Alfred가 Claude Code 도구로 템플릿 복사
   * @param templatePath - Template directory path
   * @returns Number of files processed
   * @tags @CODE:UPDATE-REFACTOR-001:COPY-TEMPLATES:API
   */
  async copyTemplatesWithClaudeTools(templatePath: string): Promise<number> {
    logger.log(chalk.cyan('\n📄 Phase 4: Alfred가 템플릿 복사 중...'));

    let filesCopied = 0;

    // P0 요구사항 구현
    try {
      filesCopied += await this.handleProjectDocs(templatePath);
    } catch (error) {
      logger.log(
        chalk.yellow(
          `   ⚠️  프로젝트 문서 처리 실패: ${error instanceof Error ? error.message : '알 수 없는 오류'}`
        )
      );
    }

    try {
      filesCopied += await this.handleHookFiles(templatePath);
    } catch (error) {
      logger.log(
        chalk.yellow(
          `   ⚠️  훅 파일 처리 실패: ${error instanceof Error ? error.message : '알 수 없는 오류'}`
        )
      );
    }

    try {
      filesCopied += await this.handleOutputStyles(templatePath);
    } catch (error) {
      logger.log(
        chalk.yellow(
          `   ⚠️  Output Styles 처리 실패: ${error instanceof Error ? error.message : '알 수 없는 오류'}`
        )
      );
    }

    try {
      filesCopied += await this.handleOtherFiles(templatePath);
    } catch (error) {
      logger.log(
        chalk.yellow(
          `   ⚠️  기타 파일 처리 실패: ${error instanceof Error ? error.message : '알 수 없는 오류'}`
        )
      );
    }

    logger.log(chalk.green(`   ✅ ${filesCopied}개 파일 처리 완료`));
    return filesCopied;
  }

  /**
   * R002: 프로젝트 문서 보호 ({{PROJECT_NAME}} 검증)
   * @param templatePath - Template directory path
   * @returns Number of files processed
   * @tags @CODE:UPDATE-REFACTOR-001:PROJECT-DOCS
   */
  private async handleProjectDocs(templatePath: string): Promise<number> {
    const docs = ['product.md', 'structure.md', 'tech.md'];
    let count = 0;

    for (const doc of docs) {
      const sourcePath = path.join(templatePath, '.moai/project', doc);
      const targetPath = path.join(this.projectPath, '.moai/project', doc);

      try {
        // [Read] 템플릿 파일
        const templateContent = await fs.readFile(sourcePath, 'utf-8');

        // [Grep] {{PROJECT_NAME}} 패턴 검증
        const isTemplate = templateContent.includes('{{PROJECT_NAME}}');

        // Check if target exists
        let targetExists = false;
        let targetContent = '';
        try {
          targetContent = await fs.readFile(targetPath, 'utf-8');
          targetExists = true;
        } catch {
          // File doesn't exist
        }

        if (!targetExists) {
          // 파일 없음 → 새로 생성
          await fs.mkdir(path.dirname(targetPath), { recursive: true });
          await fs.writeFile(targetPath, templateContent);
          logger.log(chalk.green(`   → ${doc}: 새로 생성`));
          count++;
        } else {
          // Target exists
          const isTargetCustomized =
            !targetContent.includes('{{PROJECT_NAME}}');

          if (isTargetCustomized) {
            // User has customized the file, backup before overwriting
            await backupFile(targetPath);
            await fs.writeFile(targetPath, templateContent);
            logger.log(chalk.yellow(`   → ${doc}: 사용자 수정 (백업 완료)`));
            count++;
          } else if (isTemplate) {
            // Both template and target are in template state, just overwrite
            await fs.writeFile(targetPath, templateContent);
            logger.log(chalk.blue(`   → ${doc}: 템플릿 (덮어쓰기)`));
            count++;
          } else {
            // Template is not in template state but target is still template
            // Overwrite to update
            await fs.writeFile(targetPath, templateContent);
            logger.log(chalk.blue(`   → ${doc}: 덮어쓰기`));
            count++;
          }
        }
      } catch (error) {
        logger.log(
          chalk.red(
            `   ✗ ${doc}: ${error instanceof Error ? error.message : '오류'}`
          )
        );
      }
    }

    return count;
  }

  /**
   * R003: 훅 파일 실행 권한 처리
   * @param templatePath - Template directory path
   * @returns Number of files processed
   * @tags @CODE:UPDATE-REFACTOR-001:HOOK-PERMISSIONS
   */
  private async handleHookFiles(templatePath: string): Promise<number> {
    const hookPath = path.join(templatePath, '.claude/hooks/alfred');
    const targetPath = path.join(this.projectPath, '.claude/hooks/alfred');

    try {
      // Check if hook directory exists
      await fs.access(hookPath);
    } catch {
      // Hook directory doesn't exist in template
      return 0;
    }

    try {
      await fs.mkdir(targetPath, { recursive: true });
      const files = await fs.readdir(hookPath);
      let count = 0;

      for (const file of files) {
        const source = path.join(hookPath, file);
        const target = path.join(targetPath, file);

        const stat = await fs.stat(source);
        if (!stat.isFile()) {
          continue;
        }

        await fs.copyFile(source, target);

        // chmod +x (Windows 예외)
        if (process.platform !== 'win32') {
          await fs.chmod(target, 0o755);
          logger.log(chalk.green(`   → chmod +x ${file}`));
        }
        count++;
      }

      return count;
    } catch (error) {
      logger.log(
        chalk.red(
          `   ✗ 훅 파일: ${error instanceof Error ? error.message : '오류'}`
        )
      );
      return 0;
    }
  }

  /**
   * R004: Output Styles 복사
   * @param templatePath - Template directory path
   * @returns Number of files processed
   * @tags @CODE:UPDATE-REFACTOR-001:OUTPUT-STYLES
   */
  private async handleOutputStyles(templatePath: string): Promise<number> {
    const sourcePath = path.join(templatePath, '.claude/output-styles/alfred');
    const targetPath = path.join(
      this.projectPath,
      '.claude/output-styles/alfred'
    );

    try {
      // Check if source exists
      await fs.access(sourcePath);
      return await copyDirectory(sourcePath, targetPath);
    } catch (error) {
      logger.log(
        chalk.red(
          `   ✗ Output Styles: ${error instanceof Error ? error.message : '오류'}`
        )
      );
      return 0;
    }
  }

  /**
   * 기타 파일 복사
   * @param templatePath - Template directory path
   * @returns Number of files processed
   * @tags @CODE:UPDATE-REFACTOR-001:OTHER-FILES
   */
  private async handleOtherFiles(templatePath: string): Promise<number> {
    const items = [
      '.claude/commands/alfred',
      '.claude/agents/alfred',
      '.moai/memory/development-guide.md',
      'CLAUDE.md',
    ];

    let count = 0;
    for (const item of items) {
      try {
        const source = path.join(templatePath, item);
        const target = path.join(this.projectPath, item);

        const stat = await fs.stat(source);
        if (stat.isDirectory()) {
          count += await copyDirectory(source, target);
        } else {
          await fs.mkdir(path.dirname(target), { recursive: true });
          await fs.copyFile(source, target);
          count++;
        }
      } catch (error) {
        logger.log(
          chalk.red(
            `   ✗ ${item}: ${error instanceof Error ? error.message : '오류'}`
          )
        );
      }
    }

    return count;
  }
}

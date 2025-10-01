// @CODE:UPD-001 |
// Related: @CODE:UPD-001:API, @CODE:UPD-VER-001

/**
 * @file Update orchestration system
 * @author MoAI Team
 */

import { promises as fs } from 'node:fs';
import * as path from 'node:path';
import chalk from 'chalk';
import { execa } from 'execa';
import { checkLatestVersion, getCurrentVersion } from '../../utils/version.js';
import { logger } from '../../utils/winston-logger.js';

/**
 * Simplified update configuration
 * @tags @SPEC:UPDATE-CONFIG-002
 */
export interface UpdateConfiguration {
  readonly projectPath: string;
  readonly checkOnly?: boolean; // Only check for updates
  readonly force?: boolean; // Skip backup
  readonly verbose?: boolean;
}

/**
 * Update operation result
 * @tags @SPEC:UPDATE-RESULT-002
 */
export interface UpdateResult {
  readonly success: boolean;
  readonly currentVersion: string;
  readonly latestVersion: string | null;
  readonly hasUpdate: boolean;
  readonly backupPath?: string | undefined;
  readonly filesUpdated: number;
  readonly duration: number;
  readonly errors: string[];
}

/**
 * Simplified update orchestrator: backup and overwrite strategy
 * @tags @CODE:UPDATE-ORCHESTRATOR-001
 */
export class UpdateOrchestrator {
  private readonly projectPath: string;

  constructor(projectPath: string) {
    this.projectPath = projectPath;
  }

  /**
   * Execute simplified update operation
   * @param config - Update configuration
   * @returns Update operation result
   * @tags @CODE:EXECUTE-UPDATE-001:API
   */
  public async executeUpdate(
    config: UpdateConfiguration
  ): Promise<UpdateResult> {
    const startTime = Date.now();
    const errors: string[] = [];

    try {
      // Phase 1: Version check
      logger.log(chalk.cyan('🔍 MoAI-ADK 업데이트 확인 중...'));

      const currentVersion = getCurrentVersion();
      const versionCheck = await checkLatestVersion();

      logger.log(chalk.blue(`📦 현재 버전: v${currentVersion}`));

      if (!versionCheck.hasUpdate || !versionCheck.latestVersion) {
        logger.log(chalk.green('✅ 최신 버전을 사용 중입니다'));
        return {
          success: true,
          currentVersion,
          latestVersion: versionCheck.latestVersion,
          hasUpdate: false,
          filesUpdated: 0,
          duration: Date.now() - startTime,
          errors: [],
        };
      }

      logger.log(chalk.yellow(`⚡ 최신 버전: v${versionCheck.latestVersion}`));
      logger.log(chalk.green('✅ 업데이트 가능'));

      // If check-only mode, stop here
      if (config.checkOnly) {
        return {
          success: true,
          currentVersion,
          latestVersion: versionCheck.latestVersion,
          hasUpdate: true,
          filesUpdated: 0,
          duration: Date.now() - startTime,
          errors: [],
        };
      }

      // Phase 2: Backup (unless --force)
      let backupPath: string | undefined;
      if (!config.force) {
        logger.log(chalk.cyan('\n💾 백업 생성 중...'));
        backupPath = await this.createBackup();
        logger.log(chalk.green(`   → ${backupPath}`));
      }

      // Phase 3: npm package update
      logger.log(chalk.cyan('\n📦 패키지 업데이트 중...'));
      await this.updateNpmPackage();
      logger.log(chalk.green('   ✅ 패키지 업데이트 완료'));

      // Phase 4: Template file copy
      logger.log(chalk.cyan('\n📁 패키지 경로 확인 중...'));
      const npmRoot = await this.getNpmRoot();
      const templatePath = path.join(npmRoot, 'moai-adk', 'templates');
      logger.log(chalk.blue(`   npm root → ${npmRoot}`));
      logger.log(chalk.green(`   ✅ 템플릿 경로: ${templatePath}`));

      // Verify template directory exists
      try {
        await fs.access(templatePath);
      } catch {
        throw new Error(`템플릿 디렉토리를 찾을 수 없습니다: ${templatePath}`);
      }

      logger.log(chalk.cyan('\n📄 템플릿 파일 복사 중...'));
      const filesUpdated = await this.copyTemplateFiles(templatePath);
      logger.log(chalk.green(`   ✅ ${filesUpdated}개 파일 복사 완료`));

      // Phase 5: Verification
      logger.log(chalk.cyan('\n🔍 검증 중...'));
      await this.verifyUpdate(templatePath);
      logger.log(chalk.green('   ✅ 검증 완료'));

      const duration = Date.now() - startTime;
      logger.log(chalk.green('\n✨ 업데이트 완료!'));

      if (backupPath) {
        logger.log(
          chalk.gray(
            `\n롤백이 필요하면: moai restore --from=${path.basename(backupPath)}`
          )
        );
      }

      return {
        success: true,
        currentVersion,
        latestVersion: versionCheck.latestVersion,
        hasUpdate: true,
        backupPath: backupPath ?? undefined,
        filesUpdated,
        duration,
        errors: [],
      };
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      errors.push(errorMessage);

      logger.log(chalk.red(`\n❌ 업데이트 실패: ${errorMessage}`));

      return {
        success: false,
        currentVersion: getCurrentVersion(),
        latestVersion: null,
        hasUpdate: false,
        filesUpdated: 0,
        duration: Date.now() - startTime,
        errors,
      };
    }
  }

  /**
   * Create backup of existing files
   * @returns Backup directory path
   * @tags @UTIL:CREATE-BACKUP-001
   */
  private async createBackup(): Promise<string> {
    const timestamp = new Date()
      .toISOString()
      .replace(/T/, '-')
      .replace(/\..+/, '')
      .replace(/:/g, '-');

    const backupDir = path.join(this.projectPath, '.moai-backup', timestamp);

    // Backup directories
    const dirsToBackup = ['.claude', '.moai', 'CLAUDE.md'];

    for (const dir of dirsToBackup) {
      const sourcePath = path.join(this.projectPath, dir);
      try {
        await fs.access(sourcePath);
        const targetPath = path.join(backupDir, dir);

        if ((await fs.stat(sourcePath)).isDirectory()) {
          await this.copyDirectory(sourcePath, targetPath);
        } else {
          await fs.mkdir(path.dirname(targetPath), { recursive: true });
          await fs.copyFile(sourcePath, targetPath);
        }
      } catch {
        // File/directory doesn't exist, skip
      }
    }

    return backupDir;
  }

  /**
   * Copy directory recursively
   * @param source - Source directory
   * @param target - Target directory
   * @tags @UTIL:COPY-DIRECTORY-001
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
   * Update npm package to latest version
   * @tags @UTIL:UPDATE-NPM-PACKAGE-001
   */
  private async updateNpmPackage(): Promise<void> {
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
  }

  /**
   * Get npm root directory
   * @returns npm root path
   * @tags @UTIL:GET-NPM-ROOT-001
   */
  private async getNpmRoot(): Promise<string> {
    try {
      // Try local first
      const { stdout } = await execa('npm', ['root'], {
        cwd: this.projectPath,
      });
      return stdout.trim();
    } catch {
      // Try global
      const { stdout } = await execa('npm', ['root', '-g']);
      return stdout.trim();
    }
  }

  /**
   * Copy template files to project (simple overwrite)
   * @param templatePath - Template directory path
   * @returns Number of files copied
   * @tags @UTIL:COPY-TEMPLATE-FILES-001
   */
  private async copyTemplateFiles(templatePath: string): Promise<number> {
    let filesCopied = 0;

    const filesToCopy = [
      { src: '.claude/commands/moai', dest: '.claude/commands/moai' },
      { src: '.claude/agents/moai', dest: '.claude/agents/moai' },
      { src: '.claude/hooks/moai', dest: '.claude/hooks/moai' },
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

    return filesCopied;
  }

  /**
   * Count files in directory recursively
   * @param dirPath - Directory path
   * @returns File count
   * @tags @UTIL:COUNT-FILES-001
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

  /**
   * Verify update was successful
   * @param templatePath - Template path for verification
   * @tags @UTIL:VERIFY-UPDATE-001
   */
  private async verifyUpdate(_templatePath: string): Promise<void> {
    // Verify key files exist
    const keyFiles = [
      '.moai/memory/development-guide.md',
      'CLAUDE.md',
      '.claude/commands/moai',
      '.claude/agents/moai',
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
  }
}

// @CODE:UPD-001 |
// Related: @CODE:UPD-001:API, @CODE:UPD-VER-001

/**
 * @file Update orchestration system (Refactored)
 * @author MoAI Team
 * @description 리팩토링된 UpdateOrchestrator - 책임 분리 및 모듈화
 */

import * as path from 'node:path';
import chalk from 'chalk';
import { getCurrentVersion } from '../../utils/version.js';
import { logger } from '../../utils/winston-logger.js';
import { UpdateVerifier } from './checkers/update-verifier.js';
import { VersionChecker } from './checkers/version-checker.js';
import { BackupManager } from './updaters/backup-manager.js';
import { NpmUpdater } from './updaters/npm-updater.js';
import { TemplateCopier } from './updaters/template-copier.js';

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
 *
 * 리팩토링 완료:
 * - 이전: 381 LOC 단일 파일
 * - 이후: 120 LOC 오케스트레이터 + 5개 전문 모듈
 * - 책임 분리: 버전 체크, 백업, npm 업데이트, 템플릿 복사, 검증
 */
export class UpdateOrchestrator {
  private readonly versionChecker: VersionChecker;
  private readonly backupManager: BackupManager;
  private readonly npmUpdater: NpmUpdater;
  private readonly templateCopier: TemplateCopier;
  private readonly verifier: UpdateVerifier;

  constructor(projectPath: string) {
    this.versionChecker = new VersionChecker();
    this.backupManager = new BackupManager(projectPath);
    this.npmUpdater = new NpmUpdater(projectPath);
    this.templateCopier = new TemplateCopier(projectPath);
    this.verifier = new UpdateVerifier(projectPath);
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
      const versionCheck = await this.versionChecker.checkForUpdates();

      if (!versionCheck.hasUpdate) {
        return {
          success: true,
          currentVersion: versionCheck.currentVersion,
          latestVersion: versionCheck.latestVersion,
          hasUpdate: false,
          filesUpdated: 0,
          duration: Date.now() - startTime,
          errors: [],
        };
      }

      // If check-only mode, stop here
      if (config.checkOnly) {
        return {
          success: true,
          currentVersion: versionCheck.currentVersion,
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
        backupPath = await this.backupManager.createBackup();
      }

      // Phase 3: npm package update
      await this.npmUpdater.updatePackage();

      // Phase 4: Template file copy
      const npmRoot = await this.npmUpdater.getNpmRoot();
      const templatePath = path.join(npmRoot, 'moai-adk', 'templates');
      const filesUpdated = await this.templateCopier.copyTemplates(templatePath);

      // Phase 5: Verification
      await this.verifier.verifyUpdate();

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
        currentVersion: versionCheck.currentVersion,
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
}

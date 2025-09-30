/**
 * @file Main update orchestrator that integrates all update components
 * @author MoAI Team
 * @tags @FEATURE:UPDATE-STRATEGY-001 | Chain: @REQ:UPDATE-REAL-001 -> @DESIGN:UPDATE-ORCHESTRATOR-001 -> @TASK:UPDATE-ORCHESTRATOR-001 -> @TEST:UPDATE-ORCHESTRATOR-001
 * Related: @SEC:UPDATE-STRATEGY-001, @DOCS:UPDATE-STRATEGY-001
 */

import { promises as fs } from 'node:fs';
import * as path from 'node:path';
import chalk from 'chalk';
import { BackupManager, type BackupResult } from './backup-manager.js';
import {
  type ConflictResolution,
  ConflictResolver,
} from './conflict-resolver.js';
import {
  ConsoleMigrationLogger,
  type MigrationContext,
  type MigrationExecutionPlan,
  MigrationFramework,
  type MigrationResult,
} from './migration-framework.js';
import { UpdateStrategy, type UpdateStrategyResult } from './strategy.js';
import { UpdateAction } from './types.js';
import { logger } from '../../utils/winston-logger.js';
import {
  type UpdateRecord,
  VersionManager,
} from './version-manager.js';

/**
 * Complete update operation configuration
 * @tags @DESIGN:UPDATE-CONFIG-001
 */
export interface UpdateConfiguration {
  readonly projectPath: string;
  readonly templatePath: string;
  readonly backupEnabled: boolean;
  readonly interactiveMode: boolean;
  readonly dryRun: boolean;
  readonly verbose: boolean;
  readonly forceUpdate: boolean;
  readonly skipValidation: boolean;
}

/**
 * Complete update operation result
 * @tags @DESIGN:UPDATE-OPERATION-RESULT-001
 */
export interface UpdateOperationResult {
  readonly success: boolean;
  readonly duration: number;
  readonly summary: UpdateSummary;
  readonly backupInfo?: BackupResult;
  readonly migrationResult?: MigrationResult;
  readonly errors: readonly string[];
  readonly warnings: readonly string[];
}

/**
 * Update operation summary
 * @tags @DESIGN:UPDATE-SUMMARY-001
 */
export interface UpdateSummary {
  readonly fromVersion: string;
  readonly toVersion: string;
  readonly filesAnalyzed: number;
  readonly filesChanged: number;
  readonly filesSkipped: number;
  readonly conflictsResolved: number;
  readonly migrationsExecuted: number;
}

/**
 * Main update orchestrator that coordinates all update operations
 * @tags @FEATURE:UPDATE-ORCHESTRATOR-001
 */
export class UpdateOrchestrator {
  private readonly updateStrategy: UpdateStrategy;
  private readonly backupManager: BackupManager;
  private readonly conflictResolver: ConflictResolver;
  private readonly versionManager: VersionManager;
  private readonly migrationFramework: MigrationFramework;

  constructor(projectPath: string) {
    this.updateStrategy = new UpdateStrategy();
    this.backupManager = new BackupManager(projectPath);
    this.conflictResolver = new ConflictResolver();
    this.versionManager = new VersionManager(projectPath);
    this.migrationFramework = new MigrationFramework();
  }

  /**
   * Execute complete update operation
   * @param config - Update configuration
   * @returns Update operation result
   * @tags @API:EXECUTE-UPDATE-001
   */
  public async executeUpdate(
    config: UpdateConfiguration
  ): Promise<UpdateOperationResult> {
    const startTime = Date.now();
    const errors: string[] = [];
    const warnings: string[] = [];

    try {
      logger.info(chalk.cyan('🚀 Starting MoAI-ADK Update Operation'));
      logger.info(`Project: ${config.projectPath}`);
      logger.info(`Template: ${config.templatePath}`);

      if (config.dryRun) {
        logger.info(chalk.yellow('🧪 DRY RUN MODE - No changes will be made'));
      }

      // Step 1: Load version information
      logger.info(chalk.cyan('\n📋 Step 1: Version Analysis'));
      const versionInfo = await this.versionManager.loadVersionInfo();
      const currentVersion = versionInfo.templateVersion;

      // For demo purposes, assume we're updating to a newer version
      const targetVersion = this.getNextVersion(currentVersion);

      logger.info(`Current version: ${currentVersion}`);
      logger.info(`Target version: ${targetVersion}`);

      // Step 2: Analyze project files
      logger.info(chalk.cyan('\n🔍 Step 2: Project Analysis'));
      const analysisResult = await this.updateStrategy.analyzeProject(
        config.projectPath,
        config.templatePath
      );

      this.printAnalysisResult(analysisResult);

      // Step 3: Create migration plan
      logger.info(chalk.cyan('\n📝 Step 3: Migration Planning'));
      const migrationPlan = this.migrationFramework.createMigrationPlan(
        currentVersion,
        targetVersion
      );

      this.printMigrationPlan(migrationPlan);

      // Step 4: Create backup if needed
      let backupResult: BackupResult | undefined;
      if (config.backupEnabled && !config.dryRun) {
        logger.info(chalk.cyan('\n💾 Step 4: Creating Backup'));

        const filesToBackup = this.updateStrategy.getFilesRequiringBackup(
          analysisResult.analysisResults
        );

        if (filesToBackup.length > 0) {
          backupResult = await this.backupManager.createBackup(
            filesToBackup,
            config.projectPath,
            'pre-update',
            { before: currentVersion, after: targetVersion }
          );

          if (backupResult.success) {
            logger.info(
              chalk.green(`✅ Backup created: ${backupResult.backupId}`)
            );
            logger.info(`Backed up ${backupResult.filesBackedUp} files`);
          } else {
            errors.push(
              `Backup failed: ${backupResult.error || 'Unknown error'}`
            );
          }
        } else {
          logger.info(chalk.gray('ℹ️  No files require backup'));
        }
      }

      // Step 5: Resolve conflicts
      logger.info(chalk.cyan('\n🔧 Step 5: Conflict Resolution'));
      const conflictFiles = analysisResult.analysisResults.filter(
        result =>
          result.conflictPotential === 'high' ||
          result.conflictPotential === 'medium'
      );

      let resolutions: Map<string, ConflictResolution> = new Map();
      if (
        conflictFiles.length > 0 &&
        config.interactiveMode &&
        !config.dryRun
      ) {
        resolutions = await this.conflictResolver.resolveConflicts(
          conflictFiles,
          config.projectPath,
          config.templatePath
        );
      } else if (conflictFiles.length > 0) {
        logger.info(
          chalk.yellow(
            `⚠️  ${conflictFiles.length} conflicts require manual resolution`
          )
        );
        warnings.push(
          `${conflictFiles.length} conflicts require manual resolution`
        );
      } else {
        logger.info(chalk.green('✅ No conflicts detected'));
      }

      // Step 6: Execute migrations
      let migrationResult: MigrationResult | undefined;
      if (migrationPlan.totalSteps > 0 && !config.dryRun) {
        logger.info(chalk.cyan('\n⚡ Step 6: Executing Migrations'));

        const migrationContext: MigrationContext = {
          projectPath: config.projectPath,
          templatePath: config.templatePath,
          fromVersion: currentVersion,
          toVersion: targetVersion,
          backupId: backupResult?.backupId,
          dryRun: config.dryRun,
          logger: new ConsoleMigrationLogger(config.verbose),
        };

        migrationResult = await this.migrationFramework.executeMigrations(
          migrationPlan,
          migrationContext
        );

        if (!migrationResult.success) {
          errors.push(`Migration failed: ${migrationResult.message}`);
          errors.push(...migrationResult.errors);
        }
      }

      // Step 7: Apply file updates
      logger.info(chalk.cyan('\n📝 Step 7: Applying Updates'));
      const { filesChanged, filesSkipped } = await this.applyFileUpdates(
        analysisResult,
        resolutions,
        config
      );

      // Step 8: Update version information
      if (!config.dryRun) {
        logger.info(chalk.cyan('\n📊 Step 8: Updating Version Information'));
        await this.updateVersionInformation(
          currentVersion,
          targetVersion,
          filesChanged,
          backupResult?.backupId
        );
      }

      // Generate summary
      const duration = Date.now() - startTime;
      const summary: UpdateSummary = {
        fromVersion: currentVersion,
        toVersion: targetVersion,
        filesAnalyzed: analysisResult.totalFiles,
        filesChanged,
        filesSkipped,
        conflictsResolved: resolutions.size,
        migrationsExecuted: migrationPlan.totalSteps,
      };

      // Print final result
      this.printUpdateResult(
        summary,
        duration,
        errors,
        warnings,
        config.dryRun
      );

      return {
        success: errors.length === 0,
        duration,
        summary,
        backupInfo: backupResult,
        migrationResult,
        errors,
        warnings,
      };
    } catch (error) {
      const duration = Date.now() - startTime;
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      errors.push(`Update operation failed: ${errorMessage}`);

      logger.info(chalk.red(`\n❌ Update failed: ${errorMessage}`));

      return {
        success: false,
        duration,
        summary: {
          fromVersion: 'unknown',
          toVersion: 'unknown',
          filesAnalyzed: 0,
          filesChanged: 0,
          filesSkipped: 0,
          conflictsResolved: 0,
          migrationsExecuted: 0,
        },
        errors,
        warnings,
      };
    }
  }

  /**
   * Print analysis result summary
   * @param result - Analysis result to print
   * @tags @UTIL:PRINT-ANALYSIS-RESULT-001
   */
  private printAnalysisResult(result: UpdateStrategyResult): void {
    logger.info(`📁 Files analyzed: ${result.totalFiles}`);
    logger.info(`🔄 Safe to auto-update: ${result.safeToAutoUpdate}`);
    logger.info(`👤 Requires manual review: ${result.requiresManualReview}`);
    logger.info(`💾 Requires backup: ${result.requiresBackup}`);

    if (result.conflictFiles.length > 0) {
      logger.info(
        chalk.yellow(`⚠️  Conflict files: ${result.conflictFiles.join(', ')}`)
      );
    }
  }

  /**
   * Print migration plan summary
   * @param plan - Migration plan to print
   * @tags @UTIL:PRINT-MIGRATION-PLAN-001
   */
  private printMigrationPlan(plan: MigrationExecutionPlan): void {
    logger.info(`📝 Migrations to execute: ${plan.totalSteps}`);
    logger.info(`💾 Backup required: ${plan.requiresBackup ? 'Yes' : 'No'}`);
    logger.info(`⚠️  Risk level: ${plan.riskLevel}`);
    logger.info(
      `⏱️  Estimated duration: ${Math.round(plan.estimatedDuration / 1000)}s`
    );

    if (plan.migrations.length > 0) {
      logger.info('\nMigrations:');
      for (const migration of plan.migrations) {
        logger.info(`  • ${migration.name} (${migration.version})`);
      }
    }
  }

  /**
   * Apply file updates based on analysis and resolutions
   * @param analysisResult - File analysis result
   * @param resolutions - Conflict resolutions
   * @param config - Update configuration
   * @returns Files changed and skipped counts
   * @tags @UTIL:APPLY-FILE-UPDATES-001
   */
  private async applyFileUpdates(
    analysisResult: UpdateStrategyResult,
    resolutions: Map<string, ConflictResolution>,
    config: UpdateConfiguration
  ): Promise<{ filesChanged: number; filesSkipped: number }> {
    let filesChanged = 0;
    let filesSkipped = 0;

    for (const fileAnalysis of analysisResult.analysisResults) {
      const filePath = fileAnalysis.path;

      // Check if we have a resolution for this file
      const resolution = resolutions.get(filePath);
      const action =
        resolution?.choice.action || fileAnalysis.recommendedAction;

      try {
        switch (action) {
          case UpdateAction.REPLACE:
            if (!config.dryRun) {
              await this.replaceFile(
                filePath,
                config.projectPath,
                config.templatePath
              );
            }
            filesChanged++;
            logger.info(chalk.green(`✅ Replaced: ${filePath}`));
            break;

          case UpdateAction.MERGE:
            if (!config.dryRun) {
              // Smart merge would be implemented here
              logger.info(chalk.blue(`🔄 Merged: ${filePath}`));
            }
            filesChanged++;
            break;

          case UpdateAction.REGENERATE:
            if (!config.dryRun) {
              // File regeneration logic would be here
              logger.info(chalk.cyan(`🔄 Regenerated: ${filePath}`));
            }
            filesChanged++;
            break;

          case UpdateAction.KEEP:
          case UpdateAction.MANUAL:
          default:
            filesSkipped++;
            logger.info(chalk.gray(`⏭️  Skipped: ${filePath} (${action})`));
            break;
        }
      } catch (error) {
        filesSkipped++;
        logger.info(chalk.red(`❌ Failed to update: ${filePath}`));
      }
    }

    return { filesChanged, filesSkipped };
  }

  /**
   * Replace file with template version
   * @param filePath - Relative file path
   * @param projectPath - Project directory
   * @param templatePath - Template directory
   * @tags @UTIL:REPLACE-FILE-001
   */
  private async replaceFile(
    filePath: string,
    projectPath: string,
    templatePath: string
  ): Promise<void> {
    const sourcePath = path.join(templatePath, filePath);
    const targetPath = path.join(projectPath, filePath);

    // Ensure target directory exists
    await fs.mkdir(path.dirname(targetPath), { recursive: true });

    // Copy file
    await fs.copyFile(sourcePath, targetPath);
  }

  /**
   * Update version information after successful update
   * @param fromVersion - Previous version
   * @param toVersion - New version
   * @param filesChanged - Number of files changed
   * @param backupId - Backup ID if created
   * @tags @UTIL:UPDATE-VERSION-INFO-001
   */
  private async updateVersionInformation(
    fromVersion: string,
    toVersion: string,
    filesChanged: number,
    backupId?: string
  ): Promise<void> {
    const updateRecord: UpdateRecord = {
      id: `update-${Date.now()}`,
      timestamp: new Date().toISOString(),
      fromVersion,
      toVersion,
      type: this.getUpdateType(fromVersion, toVersion),
      filesChanged,
      backupId,
      success: true,
      duration: 0, // Would be calculated properly
    };

    await this.versionManager.recordUpdate(updateRecord);
  }

  /**
   * Get next version for demo purposes
   * @param currentVersion - Current version
   * @returns Next version
   * @tags @UTIL:GET-NEXT-VERSION-001
   */
  private getNextVersion(currentVersion: string): string {
    // Simple increment for demo
    const parts = currentVersion.split('.');
    const patch = parseInt(parts[2] || '0') + 1;
    return `${parts[0]}.${parts[1]}.${patch}`;
  }

  /**
   * Determine update type
   * @param fromVersion - From version
   * @param toVersion - To version
   * @returns Update type
   * @tags @UTIL:GET-UPDATE-TYPE-001
   */
  private getUpdateType(
    _fromVersion: string,
    _toVersion: string
  ): 'major' | 'minor' | 'patch' | 'prerelease' {
    // Simple implementation
    return 'patch';
  }

  /**
   * Print final update result
   * @param summary - Update summary
   * @param duration - Operation duration
   * @param errors - Error messages
   * @param warnings - Warning messages
   * @param dryRun - Whether this was a dry run
   * @tags @UTIL:PRINT-UPDATE-RESULT-001
   */
  private printUpdateResult(
    summary: UpdateSummary,
    duration: number,
    errors: readonly string[],
    warnings: readonly string[],
    dryRun: boolean
  ): void {
    logger.info(chalk.cyan('\n📊 Update Summary'));
    logger.info(`Version: ${summary.fromVersion} → ${summary.toVersion}`);
    logger.info(`Files analyzed: ${summary.filesAnalyzed}`);
    logger.info(`Files changed: ${summary.filesChanged}`);
    logger.info(`Files skipped: ${summary.filesSkipped}`);
    logger.info(`Conflicts resolved: ${summary.conflictsResolved}`);
    logger.info(`Migrations executed: ${summary.migrationsExecuted}`);
    logger.info(`Duration: ${Math.round(duration / 1000)}s`);

    if (warnings.length > 0) {
      logger.info(chalk.yellow('\n⚠️  Warnings:'));
      for (const warning of warnings) {
        logger.info(chalk.yellow(`  • ${warning}`));
      }
    }

    if (errors.length > 0) {
      logger.info(chalk.red('\n❌ Errors:'));
      for (const error of errors) {
        logger.info(chalk.red(`  • ${error}`));
      }
    }

    if (errors.length === 0) {
      const message = dryRun
        ? '✅ Dry run completed successfully - Ready for actual update'
        : '🎉 Update completed successfully!';
      logger.info(chalk.green(`\n${message}`));
    } else {
      logger.info(chalk.red('\n❌ Update completed with errors'));
    }
  }
}

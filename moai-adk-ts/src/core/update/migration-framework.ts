/**
 * @file Migration framework for handling version-specific update scripts
 * @author MoAI Team
 * @tags @FEATURE:UPDATE-STRATEGY-001 | Chain: @REQ:UPDATE-REAL-001 -> @DESIGN:MIGRATION-FRAMEWORK-001 -> @TASK:MIGRATION-FRAMEWORK-001 -> @TEST:MIGRATION-FRAMEWORK-001
 * Related: @SEC:UPDATE-STRATEGY-001, @DOCS:UPDATE-STRATEGY-001
 */

import { promises as fs } from 'node:fs';
import * as path from 'node:path';
import semver from 'semver';
import { logger } from '../../utils/winston-logger.js';

/**
 * Migration script interface
 * @tags @DESIGN:MIGRATION-SCRIPT-001
 */
export interface MigrationScript {
  readonly version: string;
  readonly name: string;
  readonly description: string;
  readonly requiredBackup: boolean;
  up(context: MigrationContext): Promise<MigrationResult>;
  down?(context: MigrationContext): Promise<MigrationResult>;
  validate?(context: MigrationContext): Promise<boolean>;
}

/**
 * Migration context with project information
 * @tags @DESIGN:MIGRATION-CONTEXT-001
 */
export interface MigrationContext {
  readonly projectPath: string;
  readonly templatePath: string;
  readonly fromVersion: string;
  readonly toVersion: string;
  readonly backupId?: string;
  readonly dryRun: boolean;
  readonly logger: MigrationLogger;
}

/**
 * Migration execution result
 * @tags @DESIGN:MIGRATION-RESULT-001
 */
export interface MigrationResult {
  readonly success: boolean;
  readonly message: string;
  readonly filesChanged: readonly string[];
  readonly warnings: readonly string[];
  readonly errors: readonly string[];
}

/**
 * Migration logger interface
 * @tags @DESIGN:MIGRATION-LOGGER-001
 */
export interface MigrationLogger {
  info(message: string): void;
  warn(message: string): void;
  error(message: string): void;
  debug(message: string): void;
}

/**
 * Migration execution plan
 * @tags @DESIGN:MIGRATION-PLAN-001
 */
export interface MigrationExecutionPlan {
  readonly migrations: readonly MigrationScript[];
  readonly totalSteps: number;
  readonly requiresBackup: boolean;
  readonly estimatedDuration: number;
  readonly riskLevel: 'low' | 'medium' | 'high';
}

/**
 * Migration framework for handling version-specific updates
 * @tags @FEATURE:MIGRATION-FRAMEWORK-001
 */
export class MigrationFramework {
  private readonly registeredMigrations: Map<string, MigrationScript> =
    new Map();

  /**
   * Register a migration script
   * @param migration - Migration script to register
   * @tags @API:REGISTER-MIGRATION-001
   */
  public registerMigration(migration: MigrationScript): void {
    if (!semver.valid(migration.version)) {
      throw new Error(`Invalid version format: ${migration.version}`);
    }

    this.registeredMigrations.set(migration.version, migration);
  }

  /**
   * Create migration plan for version update
   * @param fromVersion - Current version
   * @param toVersion - Target version
   * @returns Migration execution plan
   * @tags @API:CREATE-MIGRATION-PLAN-001
   */
  public createMigrationPlan(
    fromVersion: string,
    toVersion: string
  ): MigrationExecutionPlan {
    const relevantMigrations = this.getRelevantMigrations(
      fromVersion,
      toVersion
    );

    const requiresBackup = relevantMigrations.some(m => m.requiredBackup);
    const estimatedDuration = relevantMigrations.length * 30000; // 30s per migration
    const riskLevel = this.assessRiskLevel(relevantMigrations);

    return {
      migrations: relevantMigrations,
      totalSteps: relevantMigrations.length,
      requiresBackup,
      estimatedDuration,
      riskLevel,
    };
  }

  /**
   * Execute migration plan
   * @param plan - Migration execution plan
   * @param context - Migration context
   * @returns Overall migration result
   * @tags @API:EXECUTE-MIGRATIONS-001
   */
  public async executeMigrations(
    plan: MigrationExecutionPlan,
    context: MigrationContext
  ): Promise<MigrationResult> {
    const allResults: MigrationResult[] = [];
    const filesChanged: string[] = [];
    const warnings: string[] = [];
    const errors: string[] = [];

    context.logger.verbose(
      `Starting migration from ${context.fromVersion} to ${context.toVersion}`
    );
    context.logger.verbose(`Executing ${plan.totalSteps} migration(s)`);

    for (const [index, migration] of plan.migrations.entries()) {
      context.logger.verbose(
        `[${index + 1}/${plan.totalSteps}] Running: ${migration.name}`
      );

      try {
        // Validate migration if validation method exists
        if (migration.validate) {
          context.logger.debug('Validating migration preconditions...');
          const isValid = await migration.validate(context);

          if (!isValid) {
            const error = `Migration validation failed: ${migration.name}`;
            context.logger.error(error);
            errors.push(error);
            continue;
          }
        }

        // Execute migration
        const result = await migration.up(context);
        allResults.push(result);

        if (result.success) {
          context.logger.verbose(`✅ Migration completed: ${migration.name}`);
          filesChanged.push(...result.filesChanged);
          warnings.push(...result.warnings);
        } else {
          context.logger.error(
            `❌ Migration failed: ${migration.name} - ${result.message}`
          );
          errors.push(`${migration.name}: ${result.message}`);
          errors.push(...result.errors);

          // Stop on first failure for safety
          break;
        }
      } catch (error) {
        const errorMessage = `Migration crashed: ${migration.name} - ${error instanceof Error ? error.message : 'Unknown error'}`;
        context.logger.error(errorMessage);
        errors.push(errorMessage);
        break;
      }
    }

    const overallSuccess = errors.length === 0;
    const message = overallSuccess
      ? `All migrations completed successfully`
      : `Migration failed with ${errors.length} error(s)`;

    context.logger.verbose(message);

    return {
      success: overallSuccess,
      message,
      filesChanged: [...new Set(filesChanged)], // Deduplicate
      warnings: [...new Set(warnings)],
      errors: [...new Set(errors)],
    };
  }

  /**
   * Rollback migrations if needed
   * @param plan - Original migration plan
   * @param context - Migration context
   * @returns Rollback result
   * @tags @API:ROLLBACK-MIGRATIONS-001
   */
  public async rollbackMigrations(
    plan: MigrationExecutionPlan,
    context: MigrationContext
  ): Promise<MigrationResult> {
    const rollbackResults: MigrationResult[] = [];
    const filesChanged: string[] = [];
    const warnings: string[] = [];
    const errors: string[] = [];

    context.logger.verbose('Starting migration rollback...');

    // Execute rollbacks in reverse order
    const reversedMigrations = [...plan.migrations].reverse();

    for (const migration of reversedMigrations) {
      if (!migration.down) {
        warnings.push(`No rollback available for: ${migration.name}`);
        continue;
      }

      try {
        context.logger.verbose(`Rolling back: ${migration.name}`);
        const result = await migration.down(context);
        rollbackResults.push(result);

        if (result.success) {
          context.logger.verbose(`✅ Rollback completed: ${migration.name}`);
          filesChanged.push(...result.filesChanged);
          warnings.push(...result.warnings);
        } else {
          context.logger.error(
            `❌ Rollback failed: ${migration.name} - ${result.message}`
          );
          errors.push(`${migration.name}: ${result.message}`);
          errors.push(...result.errors);
        }
      } catch (error) {
        const errorMessage = `Rollback crashed: ${migration.name} - ${error instanceof Error ? error.message : 'Unknown error'}`;
        context.logger.error(errorMessage);
        errors.push(errorMessage);
      }
    }

    const success = errors.length === 0;
    const message = success
      ? 'Rollback completed successfully'
      : `Rollback completed with ${errors.length} error(s)`;

    return {
      success,
      message,
      filesChanged: [...new Set(filesChanged)],
      warnings: [...new Set(warnings)],
      errors: [...new Set(errors)],
    };
  }

  /**
   * Get list of registered migrations
   * @returns Array of migration scripts
   * @tags @API:LIST-MIGRATIONS-001
   */
  public listMigrations(): readonly MigrationScript[] {
    return Array.from(this.registeredMigrations.values()).sort((a, b) =>
      semver.compare(a.version, b.version)
    );
  }

  /**
   * Get relevant migrations for version range
   * @param fromVersion - Starting version
   * @param toVersion - Target version
   * @returns Applicable migrations
   * @tags @UTIL:GET-RELEVANT-MIGRATIONS-001
   */
  private getRelevantMigrations(
    fromVersion: string,
    toVersion: string
  ): readonly MigrationScript[] {
    const allMigrations = this.listMigrations();

    return allMigrations.filter(migration => {
      const migrationVersion = migration.version;
      return (
        semver.gt(migrationVersion, fromVersion) &&
        semver.lte(migrationVersion, toVersion)
      );
    });
  }

  /**
   * Assess risk level of migration plan
   * @param migrations - List of migrations to assess
   * @returns Risk level
   * @tags @UTIL:ASSESS-RISK-LEVEL-001
   */
  private assessRiskLevel(
    migrations: readonly MigrationScript[]
  ): 'low' | 'medium' | 'high' {
    if (migrations.length === 0) {
      return 'low';
    }

    // High risk if any migration requires backup
    if (migrations.some(m => m.requiredBackup)) {
      return 'high';
    }

    // Medium risk if more than 3 migrations
    if (migrations.length > 3) {
      return 'medium';
    }

    // Check for major version jumps
    const hasMajorVersionJump = migrations.some(m => {
      const prevMigration = migrations[migrations.indexOf(m) - 1];
      if (!prevMigration) return false;

      return semver.major(m.version) > semver.major(prevMigration.version);
    });

    return hasMajorVersionJump ? 'high' : 'low';
  }
}

/**
 * Simple console logger for migrations
 * @tags @UTIL:CONSOLE-LOGGER-001
 */
export class ConsoleMigrationLogger implements MigrationLogger {
  constructor(private readonly verbose: boolean = false) {}

  public info(message: string): void {
    logger.verbose(`ℹ️  ${message}`);
  }

  public warn(message: string): void {
    logger.verbose(`⚠️  ${message}`);
  }

  public error(message: string): void {
    logger.verbose(`❌ ${message}`);
  }

  public debug(message: string): void {
    if (this.verbose) {
      logger.verbose(`🔍 ${message}`);
    }
  }
}

/**
 * Abstract base class for migration scripts
 * @tags @UTIL:BASE-MIGRATION-001
 */
export abstract class BaseMigration implements MigrationScript {
  constructor(
    public readonly version: string,
    public readonly name: string,
    public readonly description: string,
    public readonly requiredBackup: boolean = false
  ) {}

  abstract up(context: MigrationContext): Promise<MigrationResult>;

  public async down(_context: MigrationContext): Promise<MigrationResult> {
    return {
      success: false,
      message: 'Rollback not implemented',
      filesChanged: [],
      warnings: ['This migration does not support rollback'],
      errors: [],
    };
  }

  public async validate(_context: MigrationContext): Promise<boolean> {
    return true; // Default: no validation required
  }

  /**
   * Helper method to read file safely
   * @param filePath - File path to read
   * @returns File content or null if not found
   * @tags @UTIL:READ-FILE-SAFE-001
   */
  protected async readFileSafe(filePath: string): Promise<string | null> {
    try {
      return await fs.readFile(filePath, 'utf-8');
    } catch {
      return null;
    }
  }

  /**
   * Helper method to write file with directory creation
   * @param filePath - File path to write
   * @param content - File content
   * @tags @UTIL:WRITE-FILE-SAFE-001
   */
  protected async writeFileSafe(
    filePath: string,
    content: string
  ): Promise<void> {
    await fs.mkdir(path.dirname(filePath), { recursive: true });
    await fs.writeFile(filePath, content);
  }

  /**
   * Helper method to check if file exists
   * @param filePath - File path to check
   * @returns True if file exists
   * @tags @UTIL:FILE-EXISTS-SAFE-001
   */
  protected async fileExists(filePath: string): Promise<boolean> {
    try {
      await fs.access(filePath);
      return true;
    } catch {
      return false;
    }
  }
}

/**
 * @file Version tracking and management for update operations
 * @author MoAI Team
 * @tags @FEATURE:UPDATE-STRATEGY-001 | Chain: @REQ:UPDATE-REAL-001 -> @DESIGN:VERSION-MANAGER-001 -> @TASK:VERSION-MANAGER-001 -> @TEST:VERSION-MANAGER-001
 * Related: @SEC:UPDATE-STRATEGY-001, @DOCS:UPDATE-STRATEGY-001
 */

import { promises as fs } from 'node:fs';
import * as path from 'node:path';
import semver from 'semver';

/**
 * Version information structure
 * @tags @DESIGN:VERSION-INFO-001
 */
export interface VersionInfo {
  readonly moaiVersion: string;
  readonly templateVersion: string;
  readonly projectVersion: string;
  readonly lastUpdated: string;
  readonly updateHistory: readonly UpdateRecord[];
  readonly compatibility: CompatibilityInfo;
}

/**
 * Update history record
 * @tags @DESIGN:UPDATE-RECORD-001
 */
export interface UpdateRecord {
  readonly id: string;
  readonly timestamp: string;
  readonly fromVersion: string;
  readonly toVersion: string;
  readonly type: 'major' | 'minor' | 'patch' | 'prerelease';
  readonly filesChanged: number;
  readonly backupId?: string;
  readonly success: boolean;
  readonly duration: number;
}

/**
 * Compatibility information
 * @tags @DESIGN:COMPATIBILITY-INFO-001
 */
export interface CompatibilityInfo {
  readonly minMoaiVersion: string;
  readonly maxMoaiVersion?: string;
  readonly supportedFeatures: readonly string[];
  readonly deprecatedFeatures: readonly string[];
  readonly breakingChanges: readonly BreakingChange[];
}

/**
 * Breaking change information
 * @tags @DESIGN:BREAKING-CHANGE-001
 */
export interface BreakingChange {
  readonly version: string;
  readonly description: string;
  readonly migration: string;
  readonly affectedFiles: readonly string[];
}

/**
 * Version comparison result
 * @tags @DESIGN:VERSION-COMPARISON-001
 */
export interface VersionComparison {
  readonly current: string;
  readonly available: string;
  readonly updateType: 'major' | 'minor' | 'patch' | 'prerelease' | 'none';
  readonly hasBreakingChanges: boolean;
  readonly isCompatible: boolean;
  readonly recommendations: readonly string[];
}

/**
 * Version manager for tracking and comparing versions
 * @tags @FEATURE:VERSION-MANAGER-001
 */
export class VersionManager {
  private readonly versionFilePath: string;

  constructor(projectPath: string) {
    this.versionFilePath = path.join(projectPath, '.moai', 'version.json');
  }

  /**
   * Load current version information
   * @returns Version information or default if not found
   * @tags @API:LOAD-VERSION-INFO-001
   */
  public async loadVersionInfo(): Promise<VersionInfo> {
    try {
      const content = await fs.readFile(this.versionFilePath, 'utf-8');
      return JSON.parse(content) as VersionInfo;
    } catch {
      return this.createDefaultVersionInfo();
    }
  }

  /**
   * Save version information to disk
   * @param versionInfo - Version information to save
   * @tags @API:SAVE-VERSION-INFO-001
   */
  public async saveVersionInfo(versionInfo: VersionInfo): Promise<void> {
    // Ensure directory exists
    await fs.mkdir(path.dirname(this.versionFilePath), { recursive: true });

    await fs.writeFile(
      this.versionFilePath,
      JSON.stringify(versionInfo, null, 2)
    );
  }

  /**
   * Compare current version with available version
   * @param currentVersion - Current version string
   * @param availableVersion - Available version string
   * @param compatibilityInfo - Compatibility information
   * @returns Version comparison result
   * @tags @API:COMPARE-VERSIONS-001
   */
  public compareVersions(
    currentVersion: string,
    availableVersion: string,
    compatibilityInfo?: CompatibilityInfo
  ): VersionComparison {
    if (!semver.valid(currentVersion) || !semver.valid(availableVersion)) {
      throw new Error('Invalid version format');
    }

    const comparison = semver.compare(availableVersion, currentVersion);

    if (comparison === 0) {
      return {
        current: currentVersion,
        available: availableVersion,
        updateType: 'none',
        hasBreakingChanges: false,
        isCompatible: true,
        recommendations: ['Project is up to date'],
      };
    }

    if (comparison < 0) {
      return {
        current: currentVersion,
        available: availableVersion,
        updateType: 'none',
        hasBreakingChanges: false,
        isCompatible: true,
        recommendations: ['Current version is newer than available version'],
      };
    }

    // Determine update type
    const updateType = this.getUpdateType(currentVersion, availableVersion);

    // Check for breaking changes
    const hasBreakingChanges = this.hasBreakingChangesBetween(
      currentVersion,
      availableVersion,
      compatibilityInfo?.breakingChanges || []
    );

    // Check compatibility
    const isCompatible = this.checkCompatibility(
      availableVersion,
      compatibilityInfo
    );

    // Generate recommendations
    const recommendations = this.generateRecommendations(
      updateType,
      hasBreakingChanges,
      isCompatible
    );

    return {
      current: currentVersion,
      available: availableVersion,
      updateType,
      hasBreakingChanges,
      isCompatible,
      recommendations,
    };
  }

  /**
   * Record update operation in history
   * @param updateRecord - Update record to add
   * @tags @API:RECORD-UPDATE-001
   */
  public async recordUpdate(updateRecord: UpdateRecord): Promise<void> {
    const versionInfo = await this.loadVersionInfo();

    const updatedInfo: VersionInfo = {
      ...versionInfo,
      templateVersion: updateRecord.toVersion,
      lastUpdated: updateRecord.timestamp,
      updateHistory: [...versionInfo.updateHistory, updateRecord],
    };

    await this.saveVersionInfo(updatedInfo);
  }

  /**
   * Get update history with filtering
   * @param limit - Maximum number of records to return
   * @param successOnly - Only return successful updates
   * @returns Filtered update history
   * @tags @API:GET-UPDATE-HISTORY-001
   */
  public async getUpdateHistory(
    limit: number = 10,
    successOnly: boolean = false
  ): Promise<readonly UpdateRecord[]> {
    const versionInfo = await this.loadVersionInfo();
    let history = versionInfo.updateHistory;

    if (successOnly) {
      history = history.filter(record => record.success);
    }

    // Sort by timestamp (newest first) and limit
    return history
      .sort(
        (a, b) =>
          new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
      )
      .slice(0, limit);
  }

  /**
   * Check if rollback is possible to specific version
   * @param targetVersion - Version to rollback to
   * @returns True if rollback is possible
   * @tags @API:CAN-ROLLBACK-001
   */
  public async canRollbackTo(targetVersion: string): Promise<boolean> {
    const history = await this.getUpdateHistory(50, true);

    // Check if we have a successful update record for target version
    return history.some(
      record => record.fromVersion === targetVersion && record.backupId
    );
  }

  /**
   * Generate migration plan for version update
   * @param fromVersion - Current version
   * @param toVersion - Target version
   * @param compatibilityInfo - Compatibility information
   * @returns Migration plan steps
   * @tags @API:GENERATE-MIGRATION-PLAN-001
   */
  public generateMigrationPlan(
    fromVersion: string,
    toVersion: string,
    compatibilityInfo?: CompatibilityInfo
  ): readonly string[] {
    const plan: string[] = [];

    if (!semver.valid(fromVersion) || !semver.valid(toVersion)) {
      plan.push('❌ Invalid version format detected');
      return plan;
    }

    const comparison = this.compareVersions(
      fromVersion,
      toVersion,
      compatibilityInfo
    );

    if (comparison.updateType === 'none') {
      plan.push('✅ No migration required - versions are the same');
      return plan;
    }

    // Add pre-migration steps
    plan.push('📋 Pre-migration checklist:');
    plan.push('  1. Create backup of current state');
    plan.push('  2. Review changed files');

    // Add breaking change migrations
    if (comparison.hasBreakingChanges) {
      plan.push('⚠️  Breaking changes detected:');

      const breakingChanges = compatibilityInfo?.breakingChanges || [];
      const relevantChanges = breakingChanges.filter(
        change =>
          semver.gt(change.version, fromVersion) &&
          semver.lte(change.version, toVersion)
      );

      for (const change of relevantChanges) {
        plan.push(`  • ${change.description}`);
        plan.push(`    Migration: ${change.migration}`);
      }
    }

    // Add update type specific steps
    switch (comparison.updateType) {
      case 'major':
        plan.push('🔄 Major update steps:');
        plan.push('  1. Review API changes');
        plan.push('  2. Update configuration files');
        plan.push('  3. Run compatibility tests');
        break;

      case 'minor':
        plan.push('🔄 Minor update steps:');
        plan.push('  1. Review new features');
        plan.push('  2. Update documentation');
        break;

      case 'patch':
        plan.push('🔄 Patch update steps:');
        plan.push('  1. Apply bug fixes');
        plan.push('  2. Verify functionality');
        break;

      case 'prerelease':
        plan.push('🧪 Prerelease update steps:');
        plan.push('  1. Test experimental features');
        plan.push('  2. Report feedback');
        break;
    }

    // Add post-migration steps
    plan.push('📋 Post-migration checklist:');
    plan.push('  1. Run tests');
    plan.push('  2. Verify functionality');
    plan.push('  3. Update version tracking');

    return plan;
  }

  /**
   * Create default version information
   * @returns Default version info structure
   * @tags @UTIL:CREATE-DEFAULT-VERSION-001
   */
  private createDefaultVersionInfo(): VersionInfo {
    return {
      moaiVersion: '0.0.1',
      templateVersion: '0.0.1',
      projectVersion: '0.1.0',
      lastUpdated: new Date().toISOString(),
      updateHistory: [],
      compatibility: {
        minMoaiVersion: '0.0.1',
        supportedFeatures: [],
        deprecatedFeatures: [],
        breakingChanges: [],
      },
    };
  }

  /**
   * Determine update type between versions
   * @param fromVersion - Current version
   * @param toVersion - Target version
   * @returns Update type
   * @tags @UTIL:GET-UPDATE-TYPE-001
   */
  private getUpdateType(
    fromVersion: string,
    toVersion: string
  ): 'major' | 'minor' | 'patch' | 'prerelease' {
    if (semver.prerelease(toVersion)) {
      return 'prerelease';
    }

    if (semver.major(toVersion) > semver.major(fromVersion)) {
      return 'major';
    }

    if (semver.minor(toVersion) > semver.minor(fromVersion)) {
      return 'minor';
    }

    return 'patch';
  }

  /**
   * Check if there are breaking changes between versions
   * @param fromVersion - Current version
   * @param toVersion - Target version
   * @param breakingChanges - List of breaking changes
   * @returns True if breaking changes exist
   * @tags @UTIL:HAS-BREAKING-CHANGES-001
   */
  private hasBreakingChangesBetween(
    fromVersion: string,
    toVersion: string,
    breakingChanges: readonly BreakingChange[]
  ): boolean {
    return breakingChanges.some(
      change =>
        semver.gt(change.version, fromVersion) &&
        semver.lte(change.version, toVersion)
    );
  }

  /**
   * Check version compatibility
   * @param version - Version to check
   * @param compatibilityInfo - Compatibility information
   * @returns True if compatible
   * @tags @UTIL:CHECK-COMPATIBILITY-001
   */
  private checkCompatibility(
    version: string,
    compatibilityInfo?: CompatibilityInfo
  ): boolean {
    if (!compatibilityInfo) {
      return true;
    }

    const { minMoaiVersion, maxMoaiVersion } = compatibilityInfo;

    if (semver.lt(version, minMoaiVersion)) {
      return false;
    }

    if (maxMoaiVersion && semver.gt(version, maxMoaiVersion)) {
      return false;
    }

    return true;
  }

  /**
   * Generate update recommendations
   * @param updateType - Type of update
   * @param hasBreakingChanges - Whether breaking changes exist
   * @param isCompatible - Whether version is compatible
   * @returns List of recommendations
   * @tags @UTIL:GENERATE-RECOMMENDATIONS-001
   */
  private generateRecommendations(
    updateType: 'major' | 'minor' | 'patch' | 'prerelease' | 'none',
    hasBreakingChanges: boolean,
    isCompatible: boolean
  ): readonly string[] {
    const recommendations: string[] = [];

    if (!isCompatible) {
      recommendations.push('❌ Version incompatibility detected');
      recommendations.push('💡 Check minimum version requirements');
      return recommendations;
    }

    if (hasBreakingChanges) {
      recommendations.push('⚠️  Breaking changes require manual review');
      recommendations.push('📋 Review migration guide before updating');
      recommendations.push('💾 Create backup before proceeding');
    }

    switch (updateType) {
      case 'major':
        recommendations.push('🔄 Major update available');
        recommendations.push('📖 Review changelog for new features');
        recommendations.push('🧪 Test thoroughly after update');
        break;

      case 'minor':
        recommendations.push('✨ New features available');
        recommendations.push('🔄 Safe to update with minimal risk');
        break;

      case 'patch':
        recommendations.push('🐛 Bug fixes available');
        recommendations.push('✅ Recommended to update');
        break;

      case 'prerelease':
        recommendations.push('🧪 Experimental version available');
        recommendations.push('⚠️  Use with caution in production');
        break;
    }

    if (recommendations.length === 0) {
      recommendations.push('✅ Update looks safe to proceed');
    }

    return recommendations;
  }
}

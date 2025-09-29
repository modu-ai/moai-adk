/**
 * @file Update strategy module exports
 * @author MoAI Team
 * @tags @FEATURE:UPDATE-STRATEGY-001 | Chain: @REQ:UPDATE-REAL-001 -> @DESIGN:UPDATE-STRATEGY-001 -> @TASK:UPDATE-INDEX-001 -> @TEST:UPDATE-INDEX-001
 * Related: @SEC:UPDATE-STRATEGY-001, @DOCS:UPDATE-STRATEGY-001
 */

export type {
  BackupMetadata,
  BackupResult,
  RollbackResult,
} from './backup-manager.js';
// Backup and rollback management
export { BackupManager } from './backup-manager.js';
export { ChangeAnalyzer } from './change-analyzer.js';
export type {
  ConflictChoice,
  ConflictResolution,
} from './conflict-resolver.js';
// Conflict resolution system
export { ConflictResolver } from './conflict-resolver.js';
export { FileClassifier } from './file-classifier.js';
export type {
  MigrationContext,
  MigrationExecutionPlan,
  MigrationResult,
  MigrationScript,
} from './migration-framework.js';
// Migration framework
export {
  BaseMigration,
  ConsoleMigrationLogger,
  MigrationFramework,
} from './migration-framework.js';
// Core type definitions and interfaces
export type { UpdateStrategyResult } from './strategy.js';
// Core strategy components
export { UpdateStrategy } from './strategy.js';
export {
  FileChangeAnalysis,
  FilePatterns,
  FileType,
  UpdateAction,
} from './types.js';
export type {
  UpdateConfiguration,
  UpdateOperationResult,
  UpdateSummary,
} from './update-orchestrator.js';
// Main orchestrator
export { UpdateOrchestrator } from './update-orchestrator.js';
export type {
  UpdateRecord,
  VersionComparison,
  VersionInfo,
} from './version-manager.js';
// Version tracking and management
export { VersionManager } from './version-manager.js';

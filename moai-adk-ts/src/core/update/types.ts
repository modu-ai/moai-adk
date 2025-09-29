/**
 * @file Update strategy types and interfaces
 * @author MoAI Team
 * @tags @FEATURE:UPDATE-STRATEGY-001 | Chain: @REQ:UPDATE-REAL-001 -> @DESIGN:UPDATE-STRATEGY-001 -> @TASK:UPDATE-TYPES-001 -> @TEST:UPDATE-TYPES-001
 * Related: @SEC:UPDATE-STRATEGY-001, @DOCS:UPDATE-STRATEGY-001
 */

/**
 * File classification types for update strategy
 * @tags @DESIGN:FILE-CLASSIFICATION-001
 */
export enum FileType {
  /** Template files from MoAI-ADK that should be updated */
  TEMPLATE = 'TEMPLATE',

  /** User-created files that should never be overwritten */
  USER = 'USER',

  /** Files that need intelligent merging (mix of template and user content) */
  HYBRID = 'HYBRID',

  /** Auto-generated files that can be safely recreated */
  GENERATED = 'GENERATED',

  /** Metadata files that need special handling */
  METADATA = 'METADATA',
}

/**
 * File change analysis result
 * @tags @DESIGN:CHANGE-ANALYSIS-001
 */
export interface FileChangeAnalysis {
  readonly type: FileType;
  readonly path: string;
  readonly hasLocalChanges: boolean;
  readonly hasTemplateUpdates: boolean;
  readonly conflictPotential: 'none' | 'low' | 'medium' | 'high';
  readonly recommendedAction: UpdateAction;
  readonly backupRequired: boolean;
}

/**
 * Recommended update actions
 * @tags @DESIGN:UPDATE-ACTIONS-001
 */
export enum UpdateAction {
  /** Replace file completely with template */
  REPLACE = 'REPLACE',

  /** Keep existing file unchanged */
  KEEP = 'KEEP',

  /** Attempt intelligent merge */
  MERGE = 'MERGE',

  /** Require manual conflict resolution */
  MANUAL = 'MANUAL',

  /** Regenerate file from scratch */
  REGENERATE = 'REGENERATE',
}

/**
 * Project file patterns for classification
 * @tags @DESIGN:FILE-PATTERNS-001
 */
export interface FilePatterns {
  readonly template: readonly string[];
  readonly user: readonly string[];
  readonly hybrid: readonly string[];
  readonly generated: readonly string[];
  readonly metadata: readonly string[];
}

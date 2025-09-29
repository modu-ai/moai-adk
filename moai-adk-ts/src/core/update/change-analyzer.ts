/**
 * @file Change analysis logic for update strategy
 * @author MoAI Team
 * @tags @FEATURE:UPDATE-STRATEGY-001 | Chain: @REQ:UPDATE-REAL-001 -> @DESIGN:UPDATE-STRATEGY-001 -> @TASK:CHANGE-ANALYZER-001 -> @TEST:CHANGE-ANALYZER-001
 * Related: @SEC:UPDATE-STRATEGY-001, @DOCS:UPDATE-STRATEGY-001
 */

import { promises as fs } from 'node:fs';
import * as path from 'node:path';
import { type FileChangeAnalysis, FileType, UpdateAction } from './types.js';

/**
 * Analyzes changes between current files and template updates
 * @tags @FEATURE:CHANGE-ANALYZER-001
 */
export class ChangeAnalyzer {
  /**
   * Analyze changes for a specific file
   * @param filePath - Relative file path from project root
   * @param fileType - Classified file type
   * @param projectPath - Project root directory
   * @param templatePath - Template source directory
   * @returns File change analysis result
   * @tags @API:ANALYZE-FILE-CHANGES-001
   */
  public async analyzeChanges(
    filePath: string,
    fileType: FileType,
    projectPath: string,
    templatePath: string
  ): Promise<FileChangeAnalysis> {
    const currentFilePath = path.join(projectPath, filePath);
    const templateFilePath = path.join(templatePath, filePath);

    const hasLocalFile = await this.fileExists(currentFilePath);
    const hasTemplateFile = await this.fileExists(templateFilePath);

    // Determine if files have changed
    const hasLocalChanges = await this.hasLocalModifications(
      currentFilePath,
      templateFilePath,
      fileType
    );

    const hasTemplateUpdates =
      hasTemplateFile && (!hasLocalFile || hasLocalChanges);

    // Assess conflict potential and recommend action
    const conflictPotential = this.assessConflictPotential(
      fileType,
      hasLocalChanges,
      hasTemplateUpdates
    );

    const recommendedAction = this.getRecommendedAction(
      fileType,
      hasLocalChanges,
      hasTemplateUpdates
    );

    const backupRequired = this.isBackupRequired(
      fileType,
      hasLocalChanges,
      recommendedAction
    );

    return {
      type: fileType,
      path: filePath,
      hasLocalChanges,
      hasTemplateUpdates,
      conflictPotential,
      recommendedAction,
      backupRequired,
    };
  }

  /**
   * Check if file exists
   * @param filePath - File path to check
   * @returns True if file exists
   * @tags @UTIL:FILE-EXISTS-001
   */
  private async fileExists(filePath: string): Promise<boolean> {
    try {
      await fs.access(filePath);
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Check if local file has modifications compared to template
   * @param currentPath - Current file path
   * @param templatePath - Template file path
   * @param fileType - File type for analysis strategy
   * @returns True if local modifications detected
   * @tags @UTIL:LOCAL-MODIFICATIONS-001
   */
  private async hasLocalModifications(
    currentPath: string,
    templatePath: string,
    fileType: FileType
  ): Promise<boolean> {
    if (!(await this.fileExists(currentPath))) {
      return false;
    }

    if (!(await this.fileExists(templatePath))) {
      return true; // Local file exists but no template = user created
    }

    // For USER files, always consider them modified
    if (fileType === FileType.USER) {
      return true;
    }

    try {
      const currentContent = await fs.readFile(currentPath, 'utf-8');
      const templateContent = await fs.readFile(templatePath, 'utf-8');

      // Simple content comparison for now
      // TODO: Implement more sophisticated diff analysis
      return currentContent !== templateContent;
    } catch {
      return true; // Error reading = assume modified
    }
  }

  /**
   * Assess potential for conflicts during update
   * @param fileType - File type
   * @param hasLocalChanges - Whether file has local changes
   * @param hasTemplateUpdates - Whether template has updates
   * @returns Conflict potential level
   * @tags @UTIL:CONFLICT-ASSESSMENT-001
   */
  private assessConflictPotential(
    fileType: FileType,
    hasLocalChanges: boolean,
    hasTemplateUpdates: boolean
  ): 'none' | 'low' | 'medium' | 'high' {
    if (!hasLocalChanges || !hasTemplateUpdates) {
      return 'none';
    }

    switch (fileType) {
      case FileType.USER:
        return 'none'; // Never update user files

      case FileType.TEMPLATE:
        return 'low'; // Replace template files

      case FileType.GENERATED:
        return 'none'; // Regenerate these files

      case FileType.HYBRID:
        return 'high'; // Requires careful merging

      case FileType.METADATA:
        return 'medium'; // Special handling needed

      default:
        return 'medium';
    }
  }

  /**
   * Get recommended update action based on analysis
   * @param fileType - File type
   * @param hasLocalChanges - Whether file has local changes
   * @param hasTemplateUpdates - Whether template has updates
   * @returns Recommended update action
   * @tags @UTIL:RECOMMEND-ACTION-001
   */
  private getRecommendedAction(
    fileType: FileType,
    hasLocalChanges: boolean,
    hasTemplateUpdates: boolean
  ): UpdateAction {
    if (!hasTemplateUpdates) {
      return UpdateAction.KEEP;
    }

    switch (fileType) {
      case FileType.USER:
        return UpdateAction.KEEP;

      case FileType.TEMPLATE:
        return hasLocalChanges ? UpdateAction.MANUAL : UpdateAction.REPLACE;

      case FileType.GENERATED:
        return UpdateAction.REGENERATE;

      case FileType.HYBRID:
        return hasLocalChanges ? UpdateAction.MERGE : UpdateAction.REPLACE;

      case FileType.METADATA:
        return hasLocalChanges ? UpdateAction.MANUAL : UpdateAction.REPLACE;

      default:
        return UpdateAction.MANUAL;
    }
  }

  /**
   * Determine if backup is required before update
   * @param fileType - File type
   * @param hasLocalChanges - Whether file has local changes
   * @param action - Recommended action
   * @returns True if backup is required
   * @tags @UTIL:BACKUP-REQUIRED-001
   */
  private isBackupRequired(
    fileType: FileType,
    hasLocalChanges: boolean,
    action: UpdateAction
  ): boolean {
    // Always backup user files that might be affected
    if (fileType === FileType.USER && hasLocalChanges) {
      return true;
    }

    // Backup if we're going to modify or merge
    return (
      action === UpdateAction.REPLACE ||
      action === UpdateAction.MERGE ||
      action === UpdateAction.MANUAL
    );
  }
}

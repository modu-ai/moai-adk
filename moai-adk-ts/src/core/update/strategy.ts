/**
 * @file Main update strategy orchestration
 * @author MoAI Team
 * @tags @FEATURE:UPDATE-STRATEGY-001 | Chain: @REQ:UPDATE-REAL-001 -> @DESIGN:UPDATE-STRATEGY-001 -> @TASK:UPDATE-STRATEGY-001 -> @TEST:UPDATE-STRATEGY-001
 * Related: @SEC:UPDATE-STRATEGY-001, @DOCS:UPDATE-STRATEGY-001
 */

import { promises as fs } from 'node:fs';
import * as path from 'node:path';
import { ChangeAnalyzer } from './change-analyzer.js';
import { FileClassifier } from './file-classifier.js';
import { type FileChangeAnalysis, UpdateAction } from './types.js';

/**
 * Update strategy results for entire project
 * @tags @DESIGN:UPDATE-RESULTS-001
 */
export interface UpdateStrategyResult {
  readonly analysisResults: readonly FileChangeAnalysis[];
  readonly totalFiles: number;
  readonly requiresBackup: number;
  readonly requiresManualReview: number;
  readonly safeToAutoUpdate: number;
  readonly conflictFiles: readonly string[];
}

/**
 * Main update strategy orchestrator
 * @tags @FEATURE:UPDATE-STRATEGY-001
 */
export class UpdateStrategy {
  private readonly fileClassifier: FileClassifier;
  private readonly changeAnalyzer: ChangeAnalyzer;

  constructor() {
    this.fileClassifier = new FileClassifier();
    this.changeAnalyzer = new ChangeAnalyzer();
  }

  /**
   * Analyze all files in project for update strategy
   * @param projectPath - Project root directory
   * @param templatePath - Template source directory
   * @returns Complete update strategy analysis
   * @tags @API:ANALYZE-PROJECT-001
   */
  public async analyzeProject(
    projectPath: string,
    templatePath: string
  ): Promise<UpdateStrategyResult> {
    const allFiles = await this.getAllProjectFiles(projectPath, templatePath);
    const analysisResults: FileChangeAnalysis[] = [];

    for (const filePath of allFiles) {
      const fileType = this.fileClassifier.classifyFile(filePath);
      const analysis = await this.changeAnalyzer.analyzeChanges(
        filePath,
        fileType,
        projectPath,
        templatePath
      );
      analysisResults.push(analysis);
    }

    return this.generateSummary(analysisResults);
  }

  /**
   * Get recommended files for backup before update
   * @param analysisResults - Analysis results from analyzeProject
   * @returns List of files that should be backed up
   * @tags @API:GET-BACKUP-FILES-001
   */
  public getFilesRequiringBackup(
    analysisResults: readonly FileChangeAnalysis[]
  ): readonly string[] {
    return analysisResults
      .filter(result => result.backupRequired)
      .map(result => result.path);
  }

  /**
   * Get files requiring manual review/intervention
   * @param analysisResults - Analysis results from analyzeProject
   * @returns List of files needing manual attention
   * @tags @API:GET-MANUAL-FILES-001
   */
  public getFilesRequiringManualReview(
    analysisResults: readonly FileChangeAnalysis[]
  ): readonly string[] {
    return analysisResults
      .filter(
        result =>
          result.recommendedAction === UpdateAction.MANUAL ||
          result.conflictPotential === 'high'
      )
      .map(result => result.path);
  }

  /**
   * Get all relevant files from both project and template
   * @param projectPath - Project directory
   * @param templatePath - Template directory
   * @returns Combined list of unique file paths
   * @tags @UTIL:GET-ALL-FILES-001
   */
  private async getAllProjectFiles(
    projectPath: string,
    templatePath: string
  ): Promise<readonly string[]> {
    const projectFiles = await this.getFilesRecursive(projectPath);
    const templateFiles = await this.getFilesRecursive(templatePath);

    // Combine and deduplicate files
    const allFiles = new Set([...projectFiles, ...templateFiles]);

    // Filter out system and generated files that shouldn't be analyzed
    return Array.from(allFiles).filter(
      file =>
        !file.includes('node_modules/') &&
        !file.includes('.git/') &&
        !file.includes('__pycache__/') &&
        !file.includes('dist/') &&
        !file.includes('build/')
    );
  }

  /**
   * Recursively get all files in directory
   * @param dirPath - Directory to scan
   * @returns List of relative file paths
   * @tags @UTIL:GET-FILES-RECURSIVE-001
   */
  private async getFilesRecursive(dirPath: string): Promise<string[]> {
    const files: string[] = [];

    try {
      const entries = await fs.readdir(dirPath, { withFileTypes: true });

      for (const entry of entries) {
        const fullPath = path.join(dirPath, entry.name);

        if (entry.isDirectory()) {
          const subFiles = await this.getFilesRecursive(fullPath);
          files.push(...subFiles.map(f => path.join(entry.name, f)));
        } else {
          files.push(entry.name);
        }
      }
    } catch {
      // Directory doesn't exist or not accessible
    }

    return files;
  }

  /**
   * Generate summary statistics from analysis results
   * @param analysisResults - Individual file analysis results
   * @returns Summary of update strategy
   * @tags @UTIL:GENERATE-SUMMARY-001
   */
  private generateSummary(
    analysisResults: readonly FileChangeAnalysis[]
  ): UpdateStrategyResult {
    const requiresBackup = analysisResults.filter(r => r.backupRequired).length;

    const requiresManualReview = analysisResults.filter(
      r =>
        r.recommendedAction === UpdateAction.MANUAL ||
        r.conflictPotential === 'high'
    ).length;

    const safeToAutoUpdate = analysisResults.filter(
      r =>
        r.recommendedAction === UpdateAction.REPLACE ||
        r.recommendedAction === UpdateAction.REGENERATE
    ).length;

    const conflictFiles = analysisResults
      .filter(
        r => r.conflictPotential === 'high' || r.conflictPotential === 'medium'
      )
      .map(r => r.path);

    return {
      analysisResults,
      totalFiles: analysisResults.length,
      requiresBackup,
      requiresManualReview,
      safeToAutoUpdate,
      conflictFiles,
    };
  }
}

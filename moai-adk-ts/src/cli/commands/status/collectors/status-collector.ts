// @CODE:CLI-STA-001 | SPEC: status.ts 리팩토링
// Related: @CODE:CLI-003

/**
 * @file Project status collection functionality
 * @author MoAI Team
 */

import * as path from 'node:path';
import * as fs from 'fs-extra';

/**
 * Project status information
 * @tags @SPEC:PROJECT-STATUS-002
 */
export interface ProjectStatus {
  readonly path: string;
  readonly projectType: string;
  readonly moaiInitialized: boolean;
  readonly claudeInitialized: boolean;
  readonly memoryFile: boolean;
  readonly gitRepository: boolean;
}

/**
 * Status collector for project components
 * @tags @CODE:STATUS-COLLECTOR-001
 */
export class StatusCollector {
  /**
   * Check project status and components
   * @param projectPath - Path to project directory
   * @returns Project status information
   * @tags @CODE:COLLECT-STATUS-001:API
   */
  public async collectStatus(projectPath: string): Promise<ProjectStatus> {
    try {
      const resolvedPath = path.resolve(projectPath);

      // Check if MoAI is initialized
      const moaiDir = path.join(resolvedPath, '.moai');
      const moaiInitialized = await fs.pathExists(moaiDir);

      // Check if Claude is initialized
      const claudeDir = path.join(resolvedPath, '.claude');
      const claudeInitialized = await fs.pathExists(claudeDir);

      // Check if memory file exists
      const memoryFile = path.join(resolvedPath, 'CLAUDE.md');
      const memoryFileExists = await fs.pathExists(memoryFile);

      // Check if Git repository exists
      const gitDir = path.join(resolvedPath, '.git');
      const gitRepository = await fs.pathExists(gitDir);

      // Determine project type
      let projectType = 'Regular Directory';
      if (moaiInitialized && claudeInitialized) {
        projectType = 'MoAI Project (Full)';
      } else if (moaiInitialized) {
        projectType = 'MoAI Project (Partial)';
      } else if (claudeInitialized) {
        projectType = 'Claude Project';
      }

      return {
        path: resolvedPath,
        projectType,
        moaiInitialized,
        claudeInitialized,
        memoryFile: memoryFileExists,
        gitRepository,
      };
    } catch (error) {
      throw new Error(
        `Failed to check project status: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
    }
  }
}

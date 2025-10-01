// @CODE:CLI-CNT-001 | SPEC: status.ts 리팩토링
// Related: @CODE:CLI-003

/**
 * @file File counting functionality
 * @author MoAI Team
 */

import * as path from 'node:path';
import * as fs from 'fs-extra';

/**
 * File count information
 * @tags @SPEC:FILE-COUNT-002
 */
export interface FileCount {
  [key: string]: number;
}

/**
 * File counter for project components
 * @tags @CODE:FILE-COUNTER-001
 */
export class FileCounter {
  /**
   * Count project files
   * @param projectPath - Path to project directory
   * @returns File count information
   * @tags @CODE:COUNT-FILES-001:API
   */
  public async countProjectFiles(projectPath: string): Promise<FileCount> {
    try {
      const counts: FileCount = {};
      let total = 0;

      // Count files in .moai directory
      const moaiDir = path.join(projectPath, '.moai');
      const moaiExists = await fs.pathExists(moaiDir);
      if (moaiExists) {
        const moaiCount = await this.countFilesInDirectory(moaiDir);
        counts['.moai'] = moaiCount;
        total += moaiCount;
      } else {
        counts['.moai'] = 0;
      }

      // Count files in .claude directory
      const claudeDir = path.join(projectPath, '.claude');
      const claudeExists = await fs.pathExists(claudeDir);
      if (claudeExists) {
        const claudeCount = await this.countFilesInDirectory(claudeDir);
        counts['.claude'] = claudeCount;
        total += claudeCount;
      } else {
        counts['.claude'] = 0;
      }

      // Count CLAUDE.md if it exists
      const memoryFile = path.join(projectPath, 'CLAUDE.md');
      const memoryExists = await fs.pathExists(memoryFile);
      if (memoryExists) {
        counts['CLAUDE.md'] = 1;
        total += 1;
      } else {
        counts['CLAUDE.md'] = 0;
      }

      counts.total = total;
      return counts;
    } catch (error) {
      throw new Error(
        `Failed to count project files: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
    }
  }

  /**
   * Count files in a directory recursively
   * @param dirPath - Directory path
   * @returns Number of files
   * @tags UTIL:COUNT-FILES-DIRECTORY-002
   */
  private async countFilesInDirectory(dirPath: string): Promise<number> {
    try {
      const items = await fs.readdir(dirPath, { withFileTypes: true });
      let count = 0;

      for (const item of items) {
        if (item.isFile()) {
          count++;
        } else if (item.isDirectory()) {
          const subDirPath = path.join(dirPath, item.name);
          count += await this.countFilesInDirectory(subDirPath);
        }
      }

      return count;
    } catch {
      // Return 0 if directory cannot be read
      return 0;
    }
  }
}

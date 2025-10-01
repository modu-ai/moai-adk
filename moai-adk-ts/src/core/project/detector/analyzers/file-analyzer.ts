// @CODE:PROJECT-FILE-ANALYZER-001 | SPEC: project-detector
// Related: @CODE:PROJ-003

/**
 * @file File system analyzer for project detection
 * @author MoAI Team
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { logger } from '@/utils/winston-logger.js';
import type { FileInfo, LanguageExtensions } from '../../types';

/**
 * 파일 시스템 분석 담당
 */
export class FileAnalyzer {
  private readonly moaiExclusions = [
    '.claude',
    '.moai',
    'CLAUDE.md',
    'node_modules',
    '.git',
    'dist',
    'build',
    'coverage',
  ];

  /**
   * Scan directory recursively for files
   */
  async scanDirectory(dirPath: string): Promise<FileInfo[]> {
    const files: FileInfo[] = [];

    const scanRecursive = (currentPath: string) => {
      try {
        const entries = fs.readdirSync(currentPath, { withFileTypes: true });

        for (const entry of entries) {
          const fullPath = path.join(currentPath, entry.name);

          // Skip MoAI framework files and directories
          if (this.moaiExclusions.includes(entry.name)) {
            logger.info(
              `Skipping MoAI framework file/directory: ${entry.name}`
            );
            continue;
          }

          if (entry.isDirectory() && !entry.name.startsWith('.')) {
            // Skip hidden directories but recurse into others
            scanRecursive(fullPath);
          } else if (entry.isFile()) {
            files.push({
              name: entry.name,
              path: fullPath,
              exists: true,
              isFile: true,
            });
          }
        }
      } catch (error) {
        logger.error(`Error scanning directory ${currentPath}:`, error);
      }
    };

    scanRecursive(dirPath);
    return files;
  }

  /**
   * Count files by language extensions
   */
  countFilesByLanguage(
    files: FileInfo[],
    languageExtensions: LanguageExtensions
  ): Record<string, number> {
    const fileCounts: Record<string, number> = {};

    for (const lang of Object.keys(languageExtensions)) {
      fileCounts[lang] = 0;
    }

    for (const file of files) {
      const ext = path.extname(file.path).toLowerCase();
      for (const [lang, extensions] of Object.entries(languageExtensions)) {
        if (extensions.includes(ext)) {
          fileCounts[lang] = (fileCounts[lang] || 0) + 1;
        }
      }
    }

    return fileCounts;
  }

  /**
   * Find dominant language from file counts
   */
  findDominantLanguage(fileCounts: Record<string, number>): string {
    const detectedLanguage = Object.keys(fileCounts).reduce((a, b) =>
      (fileCounts[a] || 0) > (fileCounts[b] || 0) ? a : b
    );

    if ((fileCounts[detectedLanguage] || 0) > 0) {
      return detectedLanguage;
    }

    return 'unknown';
  }
}

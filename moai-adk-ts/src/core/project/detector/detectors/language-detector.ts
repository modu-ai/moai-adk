// @CODE:PROJECT-LANGUAGE-DETECTOR-001 | SPEC: project-detector
// Related: @CODE:PROJ-003

/**
 * @file Language detector based on file extensions
 * @author MoAI Team
 */

import * as fs from 'node:fs';
import { logger } from '@/utils/winston-logger.js';
import type { LanguageExtensions } from '../../types';
import type { FileAnalyzer } from '../analyzers/file-analyzer';

/**
 * 파일 확장자 기반 언어 감지 담당
 */
export class LanguageDetector {
  constructor(
    private readonly fileAnalyzer: FileAnalyzer,
    private readonly languageExtensions: LanguageExtensions
  ) {}

  /**
   * Detect primary language based on file extensions in project
   */
  async detectLanguageFromFiles(projectPath: string): Promise<string> {
    if (!fs.existsSync(projectPath)) {
      logger.warn(`Project path does not exist: ${projectPath}`);
      return 'unknown';
    }

    try {
      const files = await this.fileAnalyzer.scanDirectory(projectPath);

      const fileCounts = this.fileAnalyzer.countFilesByLanguage(
        files,
        this.languageExtensions
      );

      const detectedLanguage =
        this.fileAnalyzer.findDominantLanguage(fileCounts);

      if (detectedLanguage !== 'unknown') {
        logger.info(
          `Detected language: ${detectedLanguage} (${fileCounts[detectedLanguage]} files)`
        );
      } else {
        logger.info('No specific language detected from file extensions');
      }

      return detectedLanguage;
    } catch (error) {
      logger.error(`Error scanning files: ${error}`);
      return 'unknown';
    }
  }
}

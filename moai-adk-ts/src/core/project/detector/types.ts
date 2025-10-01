// @CODE:PROJECT-DETECT-TYPES-001 | SPEC: project-detector
// Related: @CODE:PROJ-003

/**
 * @file Project detector internal types
 * @author MoAI Team
 */

/**
 * Language detection result
 */
export interface LanguageDetectionResult {
  readonly language: string;
  readonly fileCount: number;
  readonly confidence: number;
}

/**
 * Framework detection result
 */
export interface FrameworkDetectionResult {
  readonly frameworks: string[];
  readonly buildTools: string[];
  readonly hasTypeScript: boolean;
}

/**
 * File scanning options
 */
export interface FileScanOptions {
  readonly excludePatterns: string[];
  readonly maxDepth?: number;
  readonly followSymlinks?: boolean;
}

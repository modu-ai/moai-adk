/**
 * @file Project system type definitions
 * @author MoAI Team
 * @tags @DESIGN:PROJECT-TYPES-001 @REQ:CORE-SYSTEM-013
 */

/**
 * Project detection result interface
 * @tags @DESIGN:PROJECT-INFO-001
 */
export interface ProjectInfo {
  type: string;
  language: string;
  frameworks: string[];
  buildTools: string[];
  filesFound: string[];
  hasScripts?: boolean;
  scripts?: string[];
}

/**
 * Package.json analysis result
 * @tags @DESIGN:PACKAGE-ANALYSIS-001
 */
export interface PackageAnalysis {
  frameworks: string[];
  buildTools: string[];
  hasScripts: boolean;
  scripts: string[];
}

/**
 * Project configuration interface for detection
 * @tags @DESIGN:PROJECT-CONFIG-001
 */
export interface ProjectConfig {
  readonly runtime: {
    readonly name: string;
  };
  readonly techStack: readonly string[];
}

/**
 * File information for language detection
 * @tags @DESIGN:FILE-INFO-001
 */
export interface FileInfo {
  readonly path: string;
  readonly isFile: () => boolean;
}

/**
 * Language extension mapping
 * @tags @DESIGN:LANGUAGE-MAP-001
 */
export interface LanguageExtensions {
  readonly [language: string]: readonly string[];
}

/**
 * Project file indicators
 * @tags @DESIGN:PROJECT-INDICATORS-001
 */
export interface ProjectFileIndicators {
  readonly [fileName: string]: {
    readonly type: string;
    readonly language: string;
  };
}

/**
 * Framework indicators for detection
 * @tags @DESIGN:FRAMEWORK-INDICATORS-001
 */
export interface FrameworkIndicators {
  readonly [framework: string]: readonly string[];
}

/**
 * Build tool indicators for detection
 * @tags @DESIGN:BUILD-TOOL-INDICATORS-001
 */
export interface BuildToolIndicators {
  readonly [buildTool: string]: readonly string[];
}

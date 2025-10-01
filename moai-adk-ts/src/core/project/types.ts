// @CODE:PROJ-TYPES-002 |
// Related: @CODE:PROJ-003

/**
 * @file Project detector type definitions
 * @author MoAI Team
 */

// Local ProjectConfig for detector (different from base config)
export interface ProjectConfig {
  runtime: {
    name: string;
  };
  techStack: string[];
}

/**
 * Project information detected from filesystem
 */
export interface ProjectInfo {
  type: string;
  language: string;
  frameworks: string[];
  buildTools: string[];
  packageManager?: string;
  hasTests: boolean;
  hasDocker: boolean;
  hasCI: boolean;
  filesFound: string[];
  hasScripts: boolean;
  scripts: string[];
}

/**
 * File information for analysis
 */
export interface FileInfo {
  name: string;
  path: string;
  exists: boolean;
  size?: number;
  isFile: boolean;
}

/**
 * Project file indicators mapping
 */
export interface ProjectFileIndicators {
  [filename: string]: {
    type: string;
    language: string;
  };
}

/**
 * Language extensions mapping
 */
export interface LanguageExtensions {
  [language: string]: string[];
}

/**
 * Framework indicators mapping
 */
export interface FrameworkIndicators {
  [framework: string]: string[];
}

/**
 * Build tool indicators mapping
 */
export interface BuildToolIndicators {
  [tool: string]: string[];
}

/**
 * Package.json analysis result
 */
export interface PackageAnalysis {
  dependencies: string[];
  devDependencies: string[];
  scripts: string[];
  hasTypeScript: boolean;
  hasReact: boolean;
  hasVue: boolean;
  hasNext: boolean;
  hasNest: boolean;
  hasExpress: boolean;
  frameworks: string[];
  buildTools: string[];
}

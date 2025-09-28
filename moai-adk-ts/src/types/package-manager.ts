/**
 * @file Package manager type definitions
 * @author MoAI Team
 * @tags @DATA:PACKAGE-MANAGER-001 @REQ:PACKAGE-MANAGER-002
 */

/**
 * Supported package managers
 * @tags @DATA:PACKAGE-MANAGER-TYPES-001
 */
export enum PackageManagerType {
  NPM = 'npm',
  YARN = 'yarn',
  PNPM = 'pnpm'
}

/**
 * Package manager information
 * @tags @DATA:PACKAGE-MANAGER-INFO-001
 */
export interface PackageManagerInfo {
  type: PackageManagerType;
  version: string;
  isAvailable: boolean;
  executablePath?: string;
  isPreferred?: boolean;
}

/**
 * Package manager detection result
 * @tags @DATA:PACKAGE-MANAGER-DETECTION-001
 */
export interface PackageManagerDetectionResult {
  available: PackageManagerInfo[];
  preferred?: PackageManagerInfo | undefined;
  recommended?: PackageManagerInfo | undefined;
}

/**
 * Package manager operation commands
 * @tags @DATA:PACKAGE-MANAGER-COMMANDS-001
 */
export interface PackageManagerCommands {
  install: string;
  installDev: string;
  installGlobal: string;
  run: string;
  build: string;
  test: string;
  init: string;
}

/**
 * Package.json dependencies structure
 * @tags @DATA:PACKAGE-JSON-DEPS-001
 */
export interface PackageJsonDependencies {
  dependencies?: Record<string, string>;
  devDependencies?: Record<string, string>;
  peerDependencies?: Record<string, string>;
  optionalDependencies?: Record<string, string>;
}

/**
 * Package.json structure for project generation
 * @tags @DATA:PACKAGE-JSON-001
 */
export interface PackageJsonConfig {
  name: string;
  version: string;
  description?: string;
  main?: string;
  type?: 'module' | 'commonjs';
  scripts?: Record<string, string>;
  keywords?: string[];
  author?: string;
  license?: string;
  repository?: {
    type: string;
    url: string;
  };
  bugs?: {
    url: string;
  };
  homepage?: string;
  engines?: {
    node?: string;
    npm?: string;
  };
  files?: string[];
  dependencies?: Record<string, string>;
  devDependencies?: Record<string, string>;
  peerDependencies?: Record<string, string>;
  optionalDependencies?: Record<string, string>;
}

/**
 * Package installation options
 * @tags @DATA:PACKAGE-INSTALL-OPTS-001
 */
export interface PackageInstallOptions {
  packageManager: PackageManagerType;
  isDevelopment?: boolean;
  isGlobal?: boolean;
  exact?: boolean;
  save?: boolean;
  silent?: boolean;
  workingDirectory?: string;
}
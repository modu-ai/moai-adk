// @CODE:PKG-TYPES-001 |
// Related: @CODE:PKG-001

/**
 * @file Package manager type definitions
 * @author MoAI Team
 */

/**
 * Supported package managers
 * @tags @CODE:PACKAGE-MANAGER-TYPES-001:DATA
 */
export enum PackageManagerType {
  NPM = 'npm',
  YARN = 'yarn',
  PNPM = 'pnpm',
}

/**
 * Package manager information
 * @tags @CODE:PACKAGE-MANAGER-INFO-001:DATA
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
 * @tags @CODE:PACKAGE-MANAGER-DETECTION-001:DATA
 */
export interface PackageManagerDetectionResult {
  available: PackageManagerInfo[];
  preferred?: PackageManagerInfo | undefined;
  recommended?: PackageManagerInfo | undefined;
}

/**
 * Package manager operation commands
 * @tags @CODE:PACKAGE-MANAGER-COMMANDS-001:DATA
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
 * @tags @CODE:PACKAGE-JSON-DEPS-001:DATA
 */
export interface PackageJsonDependencies {
  dependencies?: Record<string, string>;
  devDependencies?: Record<string, string>;
  peerDependencies?: Record<string, string>;
  optionalDependencies?: Record<string, string>;
}

/**
 * Package.json structure for project generation
 * @tags @CODE:PACKAGE-JSON-001:DATA
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
 * @tags @CODE:PACKAGE-INSTALL-OPTS-001:DATA
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

// @CODE:UTIL-003 | 
// Related: @CODE:VALID-001:API

/**
 * @file Path validation utilities
 * @author MoAI Team
 */

import * as fs from 'node:fs';
import * as path from 'node:path';

/**
 * MoAI-ADK package name constant
 * @internal
 */
const MOAI_PACKAGE_NAME = 'moai-adk';

/**
 * Package.json structure interface
 * @internal
 */
interface PackageJson {
  name?: string;
  [key: string]: unknown;
}

/**
 * Resolve a path to its real location, handling symlinks and non-existent paths
 * @param targetPath - Path to resolve
 * @returns Resolved absolute path
 * @internal
 */
function resolveRealPath(targetPath: string): string {
  try {
    return fs.realpathSync(targetPath);
  } catch {
    // If path doesn't exist or can't be resolved, use absolute path
    return path.resolve(targetPath);
  }
}

/**
 * Read and parse package.json file safely
 * @param packageJsonPath - Path to package.json
 * @returns Parsed package.json or null if invalid
 * @internal
 */
function readPackageJson(packageJsonPath: string): PackageJson | null {
  try {
    const content = fs.readFileSync(packageJsonPath, 'utf-8');
    return JSON.parse(content) as PackageJson;
  } catch {
    return null;
  }
}

/**
 * Check if a package.json belongs to the MoAI-ADK package
 * @param packageJson - Parsed package.json object
 * @returns true if this is the MoAI-ADK package
 * @internal
 */
function isMoAIPackage(packageJson: PackageJson | null): boolean {
  return packageJson?.name === MOAI_PACKAGE_NAME;
}

/**
 * Find the nearest package.json file by traversing up the directory tree
 * @param startPath - Path to start searching from
 * @returns Path to package.json if found, null otherwise
 * @internal
 */
function findPackageRoot(startPath: string): string | null {
  let searchPath = startPath;
  const root = path.parse(searchPath).root;

  while (searchPath !== root) {
    const packageJsonPath = path.join(searchPath, 'package.json');

    if (fs.existsSync(packageJsonPath)) {
      const packageJson = readPackageJson(packageJsonPath);
      if (isMoAIPackage(packageJson)) {
        return searchPath;
      }
    }

    // Move up one directory
    const parentPath = path.dirname(searchPath);
    if (parentPath === searchPath) {
      break; // Safety check: prevent infinite loop
    }
    searchPath = parentPath;
  }

  return null;
}

/**
 * Check if a given path is inside the MoAI-ADK package
 * @param targetPath - Path to check (defaults to current working directory)
 * @returns true if inside MoAI-ADK package, false otherwise
 * @tags @CODE:PATH-VALIDATOR-CHECK-001:API
 *
 * @example
 * ```typescript
 * // Check current directory
 * if (isInsideMoAIPackage()) {
 *   console.error('Cannot run inside MoAI-ADK package');
 * }
 *
 * // Check specific path
 * if (isInsideMoAIPackage('/path/to/project')) {
 *   console.error('Path is inside MoAI-ADK package');
 * }
 * ```
 */
export function isInsideMoAIPackage(targetPath?: string): boolean {
  const checkPath = targetPath || process.cwd();
  const resolvedPath = resolveRealPath(checkPath);
  const packageRoot = findPackageRoot(resolvedPath);

  return packageRoot !== null;
}

/**
 * Validation result interface
 */
export interface ValidationResult {
  isValid: boolean;
  error?: string;
}

/**
 * Error message for package internal initialization attempt
 * @internal
 */
const ERROR_INSIDE_PACKAGE =
  'Cannot initialize project inside MoAI-ADK package directory. ' +
  'Please choose a location outside the package.';

/**
 * Validate if a project path is safe for initialization
 * @param projectPath - Path to validate
 * @returns Validation result with error message if invalid
 * @tags @CODE:PATH-VALIDATOR-VALIDATE-001:API
 *
 * @example
 * ```typescript
 * const result = validateProjectPath('/path/to/new/project');
 * if (!result.isValid) {
 *   console.error(result.error);
 *   process.exit(1);
 * }
 * ```
 */
export function validateProjectPath(projectPath: string): ValidationResult {
  if (isInsideMoAIPackage(projectPath)) {
    return {
      isValid: false,
      error: ERROR_INSIDE_PACKAGE,
    };
  }

  return {
    isValid: true,
  };
}

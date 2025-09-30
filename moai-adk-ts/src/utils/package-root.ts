// @CODE:UTIL-006 | 
// Related: @CODE:PATH-INFO-001

/**
 * @file Package root directory utilities
 * @author MoAI Team
 */

import { existsSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

/**
 * Find the package root directory by searching upward for package.json
 * Works in development (src/), build (dist/), and installed (node_modules/) environments
 *
 * @param currentModuleUrl - import.meta.url from the calling module
 * @returns Absolute path to the package root directory
 * @throws Error if package.json is not found within 10 directory levels
 * @tags @CODE:FIND-PACKAGE-ROOT-001:API
 *
 * @example
 * ```typescript
 * // In any module
 * const packageRoot = findPackageRoot(import.meta.url);
 * // Returns: /path/to/moai-adk (where package.json exists)
 * ```
 */
export function findPackageRoot(currentModuleUrl: string): string {
  // Convert file:// URL to absolute path
  const currentFile = fileURLToPath(currentModuleUrl);
  let dir = dirname(currentFile);

  // Search upward for package.json (max 10 levels to prevent infinite loop)
  for (let i = 0; i < 10; i++) {
    const packageJsonPath = join(dir, 'package.json');

    if (existsSync(packageJsonPath)) {
      return dir;
    }

    // Move to parent directory
    const parentDir = dirname(dir);

    // Check if we've reached the filesystem root
    if (parentDir === dir) {
      break;
    }

    dir = parentDir;
  }

  throw new Error(
    `Could not find package root (package.json not found within 10 levels from ${currentFile})`
  );
}

/**
 * Get the absolute path to the templates directory
 *
 * @param currentModuleUrl - import.meta.url from the calling module
 * @returns Absolute path to the templates directory
 * @throws Error if package root or templates directory is not found
 * @tags @CODE:GET-TEMPLATES-PATH-001:API
 *
 * @example
 * ```typescript
 * // In update.ts
 * const templatesPath = getTemplatesPath(import.meta.url);
 * // Development: /path/to/moai-adk-ts/templates
 * // Global install: /usr/local/lib/node_modules/moai-adk/templates
 * ```
 */
export function getTemplatesPath(currentModuleUrl: string): string {
  const packageRoot = findPackageRoot(currentModuleUrl);
  const templatesPath = join(packageRoot, 'templates');

  if (!existsSync(templatesPath)) {
    throw new Error(
      `Templates directory not found at ${templatesPath}. Ensure 'templates' is included in package.json 'files' field.`
    );
  }

  return templatesPath;
}

/**
 * Get the package root directory for the current module
 * Convenience wrapper around findPackageRoot
 *
 * @param currentModuleUrl - import.meta.url from the calling module
 * @returns Absolute path to the package root directory
 * @tags @CODE:GET-PACKAGE-ROOT-001:API
 */
export function getPackageRoot(currentModuleUrl: string): string {
  return findPackageRoot(currentModuleUrl);
}
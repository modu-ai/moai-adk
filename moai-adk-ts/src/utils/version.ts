/**
 * @file Version utilities with package root resolution
 * @author MoAI Team
 * @tags @FEATURE:VERSION-UTILS-001 @REQ:CLI-FOUNDATION-012
 */

import * as fs from 'node:fs';
import * as path from 'node:path';

/**
 * Package information interface
 * @tags @DESIGN:PACKAGE-INFO-001
 */
export interface PackageInfo {
  readonly name: string;
  readonly version: string;
  readonly description: string;
}

/**
 * Get package information from package.json using package root resolution
 * Works reliably in development (src/), build (dist/), and global install environments
 *
 * @returns Package information
 * @tags @API:GET-PACKAGE-INFO-001
 */
export function getPackageInfo(): PackageInfo {
  try {
    // Use import.meta.url to find package root
    // Note: This is a workaround since we can't use import.meta.url in a regular function
    // We'll search from current file location or use process.cwd() as fallback
    let packageJsonPath: string;

    // Try to find package root from common locations
    const possibleRoots = [
      // Try from current module (if running from dist/ or src/)
      path.resolve(
        new URL('.', import.meta.url).pathname,
        '..',
        '..',
        'package.json'
      ),
      path.resolve(new URL('.', import.meta.url).pathname, '..', 'package.json'),
      // Fallback to cwd
      path.resolve(process.cwd(), 'package.json'),
    ];

    for (const testPath of possibleRoots) {
      if (fs.existsSync(testPath)) {
        packageJsonPath = testPath;
        const packageJson = JSON.parse(
          fs.readFileSync(packageJsonPath, 'utf-8')
        ) as {
          name: string;
          version: string;
          description: string;
        };

        return {
          name: packageJson.name,
          version: packageJson.version,
          description: packageJson.description,
        };
      }
    }

    throw new Error('package.json not found in any expected location');
  } catch (_error) {
    // Fallback for unexpected environments
    return {
      name: 'moai-adk',
      version: '0.0.1',
      description: '🗿 MoAI-ADK: Modu-AI Agentic Development kit',
    };
  }
}

/**
 * Get current version string
 * @returns Version string
 * @tags @API:GET-VERSION-001
 */
export function getCurrentVersion(): string {
  return getPackageInfo().version;
}

/**
 * Version comparison result
 * @tags @DESIGN:VERSION-CHECK-001
 */
export interface VersionCheckResult {
  readonly current: string;
  readonly latest: string | null;
  readonly hasUpdate: boolean;
  readonly error?: string;
}

/**
 * Check for latest version from npm registry
 * Uses a short timeout to avoid blocking session start
 *
 * @param timeout - Timeout in milliseconds (default: 2000ms)
 * @returns Version check result
 * @tags @API:CHECK-LATEST-VERSION-001
 */
export async function checkLatestVersion(
  timeout = 2000
): Promise<VersionCheckResult> {
  const current = getCurrentVersion();
  const packageInfo = getPackageInfo();

  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);

    const response = await fetch(
      `https://registry.npmjs.org/${packageInfo.name}/latest`,
      {
        signal: controller.signal,
        headers: {
          Accept: 'application/json',
        },
      }
    );

    clearTimeout(timeoutId);

    if (!response.ok) {
      return {
        current,
        latest: null,
        hasUpdate: false,
        error: `HTTP ${response.status}`,
      };
    }

    const data = (await response.json()) as { version: string };
    const latest = data.version;

    // Simple version comparison (assumes semantic versioning)
    const hasUpdate = compareVersions(current, latest) < 0;

    return {
      current,
      latest,
      hasUpdate,
    };
  } catch (error) {
    // Silently fail - don't block session start
    return {
      current,
      latest: null,
      hasUpdate: false,
      error: error instanceof Error ? error.message : 'Unknown error',
    };
  }
}

/**
 * Compare two semantic version strings
 * @param v1 - First version
 * @param v2 - Second version
 * @returns -1 if v1 < v2, 0 if v1 === v2, 1 if v1 > v2
 * @tags @UTIL:VERSION-COMPARE-001
 */
function compareVersions(v1: string, v2: string): number {
  const parts1 = v1.split('.').map(Number);
  const parts2 = v2.split('.').map(Number);

  for (let i = 0; i < Math.max(parts1.length, parts2.length); i++) {
    const num1 = parts1[i] || 0;
    const num2 = parts2[i] || 0;

    if (num1 < num2) return -1;
    if (num1 > num2) return 1;
  }

  return 0;
}

/**
 * @file Version utilities
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
 * Get package information from package.json
 * @returns Package information
 * @tags @API:GET-PACKAGE-INFO-001
 */
export function getPackageInfo(): PackageInfo {
  try {
    // Try multiple possible locations for package.json
    const possiblePaths = [
      path.resolve(__dirname, '../../package.json'), // From dist/
      path.resolve(__dirname, '../package.json'), // From src/
      path.resolve(process.cwd(), 'package.json'), // From current working directory
    ];

    for (const packageJsonPath of possiblePaths) {
      if (fs.existsSync(packageJsonPath)) {
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
    // Fallback for development/test environments
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

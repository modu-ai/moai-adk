/**
 * @file System checker module exports and main system checker class
 * @author MoAI Team
 * @tags @CODE:SYSTEM-CHECKER-001 @SPEC:AUTO-VERIFY-012
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { logger } from '../../utils/winston-logger.js';
import { type RequirementCheckResult, SystemDetector } from './detector';
import { requirementRegistry } from './requirements';

export type { DetectionResult, RequirementCheckResult } from './detector';
export { SystemDetector } from './detector';
export type { SystemRequirement } from './requirements';
export { RequirementRegistry, requirementRegistry } from './requirements';

/**
 * System check summary interface
 * @tags @SPEC:CHECK-SUMMARY-001
 */
export interface SystemCheckSummary {
  readonly runtime: RequirementCheckResult[];
  readonly development: RequirementCheckResult[];
  readonly optional: RequirementCheckResult[];
  readonly totalChecks: number;
  readonly passedChecks: number;
  readonly failedChecks: number;
  readonly detectedLanguages: string[];
}

/**
 * Enhanced system checker with language detection
 * @tags @CODE:ENHANCED-SYSTEM-CHECKER-001
 */
export class SystemChecker {
  private readonly detector = new SystemDetector();

  /**
   * Run comprehensive system check with language detection
   * @param projectPath - Project path to analyze
   * @returns System check summary
   * @tags @CODE:COMPREHENSIVE-CHECK-001:API
   */
  public async runSystemCheck(
    projectPath?: string
  ): Promise<SystemCheckSummary> {
    // Detect languages if project path is provided
    const detectedLanguages: string[] = [];
    if (projectPath && fs.existsSync(projectPath)) {
      detectedLanguages.push(...this.detectProjectLanguages(projectPath));

      // Add language-specific requirements
      for (const language of detectedLanguages) {
        requirementRegistry.addLanguageRequirements(language);
      }
    }

    // Get all requirements by category
    const runtimeRequirements = requirementRegistry.getByCategory('runtime');
    const developmentRequirements =
      requirementRegistry.getByCategory('development');
    const optionalRequirements = requirementRegistry.getByCategory('optional');

    // Run checks concurrently
    const [runtimeResults, developmentResults, optionalResults] =
      await Promise.all([
        this.detector.checkMultipleRequirements(runtimeRequirements),
        this.detector.checkMultipleRequirements(developmentRequirements),
        this.detector.checkMultipleRequirements(optionalRequirements),
      ]);

    // Calculate summary
    const allResults = [
      ...runtimeResults,
      ...developmentResults,
      ...optionalResults,
    ];
    const passedChecks = allResults.filter(
      r => r.result.isInstalled && r.result.versionSatisfied
    ).length;
    const failedChecks = allResults.length - passedChecks;

    return {
      runtime: runtimeResults,
      development: developmentResults,
      optional: optionalResults,
      totalChecks: allResults.length,
      passedChecks,
      failedChecks,
      detectedLanguages,
    };
  }

  /**
   * Detect programming languages in project
   * @param projectPath - Project directory path
   * @returns Array of detected languages
   * @tags UTIL:DETECT-LANGUAGES-001
   */
  private detectProjectLanguages(projectPath: string): string[] {
    const languages: Set<string> = new Set();

    try {
      const files = fs.readdirSync(projectPath);

      // Check for language-specific files
      for (const file of files) {
        const ext = path.extname(file).toLowerCase();

        switch (ext) {
          case '.ts':
          case '.tsx':
            languages.add('typescript');
            break;
          case '.js':
          case '.jsx':
          case '.mjs':
            languages.add('javascript');
            break;
          case '.py':
            languages.add('python');
            break;
          case '.java':
            languages.add('java');
            break;
          case '.go':
            languages.add('go');
            break;
          case '.rs':
            languages.add('rust');
            break;
          case '.cpp':
          case '.cc':
          case '.cxx':
            languages.add('cpp');
            break;
          case '.cs':
            languages.add('csharp');
            break;
        }

        // Check for framework/language-specific files
        if (file === 'package.json') {
          languages.add('javascript');
        } else if (file === 'tsconfig.json') {
          languages.add('typescript');
        } else if (file === 'requirements.txt' || file === 'pyproject.toml') {
          languages.add('python');
        } else if (file === 'pom.xml' || file === 'build.gradle') {
          languages.add('java');
        } else if (file === 'go.mod') {
          languages.add('go');
        } else if (file === 'Cargo.toml') {
          languages.add('rust');
        }
      }
    } catch (error) {
      // If we can't read the directory, just return empty array
      logger.warn(
        `Could not analyze project at ${projectPath}`,
        error instanceof Error
          ? { error: error.message }
          : { error: String(error) }
      );
    }

    return Array.from(languages);
  }
}

// @CODE:PROJECT-PACKAGE-ANALYZER-001 | SPEC: project-detector
// Related: @CODE:PROJ-003

/**
 * @file Package.json analyzer
 * @author MoAI Team
 */

import * as fs from 'node:fs';
import { logger } from '@/utils/winston-logger.js';
import type {
  BuildToolIndicators,
  FrameworkIndicators,
  PackageAnalysis,
} from '../../types';

/**
 * package.json 파일 분석 담당
 */
export class PackageAnalyzer {
  constructor(
    private readonly frameworkIndicators: FrameworkIndicators,
    private readonly buildToolIndicators: BuildToolIndicators
  ) {}

  /**
   * Analyze package.json for frameworks and dependencies
   */
  async analyzePackageJson(packageJsonPath: string): Promise<PackageAnalysis> {
    try {
      const packageData = JSON.parse(fs.readFileSync(packageJsonPath, 'utf-8'));

      const frameworks: string[] = [];
      const buildTools: string[] = [];

      // Combine dependencies and devDependencies
      const allDeps = {
        ...(packageData.dependencies || {}),
        ...(packageData.devDependencies || {}),
      };

      // Detect frameworks
      for (const [framework, indicators] of Object.entries(
        this.frameworkIndicators
      )) {
        if (indicators.some(indicator => indicator in allDeps)) {
          frameworks.push(framework);
          logger.info(`Detected framework: ${framework}`);
        }
      }

      // Detect build tools
      for (const [tool, indicators] of Object.entries(
        this.buildToolIndicators
      )) {
        if (indicators.some(indicator => indicator in allDeps)) {
          buildTools.push(tool);
          logger.info(`Detected build tool: ${tool}`);
        }
      }

      const scripts = Object.keys(packageData.scripts || {});

      logger.info(
        `Package.json analysis: frameworks=${frameworks}, buildTools=${buildTools}`
      );

      return {
        dependencies: Object.keys(packageData.dependencies || {}),
        devDependencies: Object.keys(packageData.devDependencies || {}),
        scripts,
        hasTypeScript: 'typescript' in allDeps,
        hasReact: 'react' in allDeps,
        hasVue: 'vue' in allDeps,
        hasNext: 'next' in allDeps,
        hasNest: '@nestjs/core' in allDeps,
        hasExpress: 'express' in allDeps,
        frameworks,
        buildTools,
      };
    } catch (error) {
      logger.error('Error analyzing package.json:', error);
      return {
        dependencies: [],
        devDependencies: [],
        scripts: [],
        hasTypeScript: false,
        hasReact: false,
        hasVue: false,
        hasNext: false,
        hasNest: false,
        hasExpress: false,
        frameworks: [],
        buildTools: [],
      };
    }
  }

  /**
   * Check if package.json is MoAI-ADK's own package.json
   */
  isMoAIPackageJson(packageJsonPath: string): boolean {
    try {
      const packageData = JSON.parse(fs.readFileSync(packageJsonPath, 'utf-8'));
      return (
        packageData.name === 'moai-adk' ||
        packageData.description?.includes('MoAI-ADK')
      );
    } catch {
      return false;
    }
  }
}

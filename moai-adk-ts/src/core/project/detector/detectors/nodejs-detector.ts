// @CODE:PROJECT-NODEJS-DETECTOR-001 | SPEC: project-detector
// Related: @CODE:PROJ-003

/**
 * @file Node.js project detector
 * @author MoAI Team
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { logger } from '@/utils/winston-logger.js';
import type { ProjectInfo } from '../../types';
import type { PackageAnalyzer } from '../analyzers/package-analyzer';

/**
 * Node.js 프로젝트 감지 담당
 */
export class NodeJsDetector {
  constructor(private readonly packageAnalyzer: PackageAnalyzer) {}

  /**
   * Detect if directory is a Node.js project
   */
  async detect(projectPath: string): Promise<ProjectInfo | null> {
    const packageJsonPath = path.join(projectPath, 'package.json');

    if (!fs.existsSync(packageJsonPath)) {
      return null;
    }

    // Skip MoAI-ADK's own package.json
    if (this.packageAnalyzer.isMoAIPackageJson(packageJsonPath)) {
      logger.info('Skipping MoAI-ADK package.json from project analysis');
      return null;
    }

    logger.info(`Found package.json, detected as Node.js project`);

    const packageAnalysis =
      await this.packageAnalyzer.analyzePackageJson(packageJsonPath);

    return {
      type: 'nodejs',
      language: 'javascript',
      frameworks: packageAnalysis.frameworks,
      buildTools: packageAnalysis.buildTools,
      packageManager: this.detectPackageManager(projectPath) || undefined,
      hasTests: packageAnalysis.scripts.some(s =>
        s.toLowerCase().includes('test')
      ),
      hasDocker: fs.existsSync(path.join(projectPath, 'Dockerfile')),
      hasCI: this.detectCI(projectPath),
      filesFound: ['package.json'],
      hasScripts: packageAnalysis.scripts.length > 0,
      scripts: packageAnalysis.scripts,
    };
  }

  /**
   * Detect package manager (npm, yarn, pnpm, bun)
   */
  private detectPackageManager(projectPath: string): string | undefined {
    if (fs.existsSync(path.join(projectPath, 'pnpm-lock.yaml'))) {
      return 'pnpm';
    }
    if (fs.existsSync(path.join(projectPath, 'yarn.lock'))) {
      return 'yarn';
    }
    if (fs.existsSync(path.join(projectPath, 'bun.lockb'))) {
      return 'bun';
    }
    if (fs.existsSync(path.join(projectPath, 'package-lock.json'))) {
      return 'npm';
    }
    return undefined;
  }

  /**
   * Detect CI configuration
   */
  private detectCI(projectPath: string): boolean {
    const ciFiles = [
      '.github/workflows',
      '.gitlab-ci.yml',
      '.circleci/config.yml',
      'azure-pipelines.yml',
      'Jenkinsfile',
    ];

    return ciFiles.some(file => fs.existsSync(path.join(projectPath, file)));
  }
}

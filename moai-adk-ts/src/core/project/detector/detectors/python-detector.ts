// @CODE:PROJECT-PYTHON-DETECTOR-001 | SPEC: project-detector
// Related: @CODE:PROJ-003

/**
 * @file Python project detector
 * @author MoAI Team
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { logger } from '@/utils/winston-logger.js';
import type { ProjectInfo } from '../../types';

/**
 * Python 프로젝트 감지 담당
 */
export class PythonDetector {
  /**
   * Detect if directory is a Python project
   */
  async detect(projectPath: string): Promise<ProjectInfo | null> {
    const pythonFiles = [
      'pyproject.toml',
      'requirements.txt',
      'setup.py',
      'Pipfile',
    ];

    const foundFiles: string[] = [];

    for (const file of pythonFiles) {
      if (fs.existsSync(path.join(projectPath, file))) {
        foundFiles.push(file);
      }
    }

    if (foundFiles.length === 0) {
      return null;
    }

    logger.info(`Found ${foundFiles.join(', ')}, detected as Python project`);

    return {
      type: 'python',
      language: 'python',
      frameworks: this.detectPythonFrameworks(projectPath),
      buildTools: this.detectPythonBuildTools(foundFiles),
      packageManager: this.detectPackageManager(foundFiles) || undefined,
      hasTests: this.detectTests(projectPath),
      hasDocker: fs.existsSync(path.join(projectPath, 'Dockerfile')),
      hasCI: this.detectCI(projectPath),
      filesFound: foundFiles,
      hasScripts: false,
      scripts: [],
    };
  }

  /**
   * Detect Python frameworks (Django, Flask, FastAPI, etc.)
   */
  private detectPythonFrameworks(projectPath: string): string[] {
    const frameworks: string[] = [];

    try {
      // Check requirements.txt
      const reqPath = path.join(projectPath, 'requirements.txt');
      if (fs.existsSync(reqPath)) {
        const content = fs.readFileSync(reqPath, 'utf-8').toLowerCase();
        if (content.includes('django')) frameworks.push('django');
        if (content.includes('flask')) frameworks.push('flask');
        if (content.includes('fastapi')) frameworks.push('fastapi');
        if (content.includes('pytest')) frameworks.push('pytest');
      }

      // Check pyproject.toml
      const pyprojectPath = path.join(projectPath, 'pyproject.toml');
      if (fs.existsSync(pyprojectPath)) {
        const content = fs.readFileSync(pyprojectPath, 'utf-8').toLowerCase();
        if (content.includes('django')) frameworks.push('django');
        if (content.includes('flask')) frameworks.push('flask');
        if (content.includes('fastapi')) frameworks.push('fastapi');
      }
    } catch (error: unknown) {
      logger.warn('Error detecting Python frameworks:', { error });
    }

    return [...new Set(frameworks)]; // Remove duplicates
  }

  /**
   * Detect Python build tools
   */
  private detectPythonBuildTools(foundFiles: string[]): string[] {
    const buildTools: string[] = [];

    if (foundFiles.includes('pyproject.toml')) {
      buildTools.push('poetry');
    }
    if (foundFiles.includes('setup.py')) {
      buildTools.push('setuptools');
    }
    if (foundFiles.includes('Pipfile')) {
      buildTools.push('pipenv');
    }

    return buildTools;
  }

  /**
   * Detect package manager
   */
  private detectPackageManager(foundFiles: string[]): string | undefined {
    if (foundFiles.includes('Pipfile')) return 'pipenv';
    if (foundFiles.includes('pyproject.toml')) return 'poetry';
    if (foundFiles.includes('requirements.txt')) return 'pip';
    return undefined;
  }

  /**
   * Detect test framework
   */
  private detectTests(projectPath: string): boolean {
    const testDirs = ['tests', 'test', '__tests__'];
    return testDirs.some(dir => fs.existsSync(path.join(projectPath, dir)));
  }

  /**
   * Detect CI configuration
   */
  private detectCI(projectPath: string): boolean {
    const ciFiles = [
      '.github/workflows',
      '.gitlab-ci.yml',
      '.circleci/config.yml',
    ];

    return ciFiles.some(file => fs.existsSync(path.join(projectPath, file)));
  }
}

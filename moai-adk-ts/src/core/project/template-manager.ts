// @FEATURE:PROJ-002 | Chain: @REQ:PROJ-002 -> @DESIGN:PROJ-002 -> @TASK:PROJ-002 -> @TEST:PROJ-002
// Related: @API:PROJ-002, @DATA:PROJ-TPL-001

/**
 * @file Project template management
 * @author MoAI Team
 */

import * as fs from 'node:fs/promises';
import * as path from 'node:path';
import { getTemplatesPath } from '../../utils/package-root.js';
import {
  type InitResult,
  type ProjectConfig,
  ProjectType,
  type TemplateData,
} from '@/types/project';
// import { getDefaultTagDatabase } from '../tag-system/tag-database';

/**
 * Template manager for project structure generation
 * Resolves templates from installed package location
 * @tags @FEATURE:TEMPLATE-MANAGER-001
 */
export class TemplateManager {
  private readonly templatesPath: string;

  /**
   * Create a new TemplateManager instance
   * @param templatesPath - Optional explicit templates path. If not provided,
   *                        automatically resolves to the installed package's templates directory
   */
  constructor(templatesPath?: string) {
    if (templatesPath) {
      this.templatesPath = templatesPath;
    } else {
      // Use package root resolution to find templates in installed package
      // Works in development (src/), build (dist/), and global install
      this.templatesPath = getTemplatesPath(import.meta.url);
    }
  }

  /**
   * Get templates directory path
   * @returns Templates directory path
   * @tags @API:TEMPLATE-PATH-001
   */
  public getTemplatesPath(): string {
    return this.templatesPath;
  }

  /**
   * Generate complete project structure from configuration
   * @param config - Project configuration
   * @param targetPath - Target directory path
   * @returns Generation result
   * @tags @API:TEMPLATE-GENERATE-001
   */
  public async generateProject(
    config: ProjectConfig,
    targetPath: string
  ): Promise<InitResult> {
    try {
      const result: InitResult = {
        success: false,
        projectPath: targetPath,
        config,
        createdFiles: [],
        errors: [],
        warnings: [],
      };

      // Validate project name
      if (!/^[a-zA-Z0-9-_]+$/.test(config.name)) {
        result.errors = ['Invalid project name format'];
        return result;
      }

      // Create project directory
      const projectPath = path.join(targetPath, config.name);
      await this.ensureDirectory(projectPath);

      // Generate template data
      const templateData = this.createTemplateData(config);

      // Create base structure
      await this.createBaseStructure(projectPath, result);

      // Create project-specific files
      await this.createProjectFiles(projectPath, config, templateData, result);

      // Create .moai directory structure
      await this.createMoaiStructure(projectPath, templateData, result);

      // Create .claude directory structure (if enabled)
      if (this.hasFeature(config, 'claude-integration')) {
        await this.createClaudeStructure(projectPath, templateData, result);
      }

      result.success = true;
      return result;
    } catch (error) {
      return {
        success: false,
        projectPath: targetPath,
        config,
        createdFiles: [],
        errors: [error instanceof Error ? error.message : 'Unknown error'],
      };
    }
  }

  /**
   * Create template data from project configuration
   * @param config - Project configuration
   * @returns Template data for file generation
   * @tags @API:TEMPLATE-DATA-001
   */
  private createTemplateData(config: ProjectConfig): TemplateData {
    return {
      projectName: config.name,
      projectType: config.type,
      timestamp: new Date().toISOString(),
      author: config.author || 'MoAI Developer',
      description:
        config.description || `A ${config.type} project built with MoAI-ADK`,
      license: config.license || 'MIT',
      packageManager: config.packageManager || 'npm',
      features: Object.fromEntries(
        (config.features || []).map(f => [f.name, f.enabled])
      ),
    };
  }

  /**
   * Create base project structure
   * @param projectPath - Project directory path
   * @param result - Result object to update
   * @tags @API:TEMPLATE-BASE-001
   */
  private async createBaseStructure(
    projectPath: string,
    _result: InitResult
  ): Promise<void> {
    const directories = [
      'src',
      'tests',
      'docs',
      '.moai',
      '.moai/project',
      '.moai/reports',
      // NOTE: [v0.0.3+] .moai/indexes 제거 - CODE-FIRST 방식으로 전환
    ];

    for (const dir of directories) {
      await this.ensureDirectory(path.join(projectPath, dir));
    }
  }

  /**
   * Create project-specific files based on type
   * @param projectPath - Project directory path
   * @param config - Project configuration
   * @param templateData - Template data
   * @param result - Result object to update
   * @tags @API:TEMPLATE-PROJECT-001
   */
  private async createProjectFiles(
    projectPath: string,
    config: ProjectConfig,
    templateData: TemplateData,
    result: InitResult
  ): Promise<void> {
    switch (config.type) {
      case ProjectType.PYTHON:
        await this.createPythonFiles(projectPath, templateData, result);
        break;

      case ProjectType.NODEJS:
      case ProjectType.TYPESCRIPT:
        await this.createNodeJSFiles(projectPath, templateData, result);
        break;

      case ProjectType.FRONTEND:
        await this.createFrontendFiles(projectPath, templateData, result);
        break;

      case ProjectType.MIXED:
        await this.createMixedProjectFiles(projectPath, templateData, result);
        break;
    }
  }

  /**
   * Create Python project files
   * @param projectPath - Project directory path
   * @param templateData - Template data
   * @param result - Result object to update
   * @tags @API:TEMPLATE-PYTHON-001
   */
  private async createPythonFiles(
    projectPath: string,
    templateData: TemplateData,
    result: InitResult
  ): Promise<void> {
    // pyproject.toml
    const pyprojectContent = this.generatePyprojectToml(templateData);
    const pyprojectPath = path.join(projectPath, 'pyproject.toml');
    await fs.writeFile(pyprojectPath, pyprojectContent);
    result.createdFiles.push('pyproject.toml');

    // src/__init__.py
    const srcInitPath = path.join(projectPath, 'src', '__init__.py');
    await fs.writeFile(srcInitPath, '# Package initialization\n');
    result.createdFiles.push('src/__init__.py');

    // Basic Python files if features are enabled
    if (templateData.features['pytest']) {
      const testConfigPath = path.join(projectPath, 'pytest.ini');
      await fs.writeFile(testConfigPath, this.generatePytestConfig());
      result.createdFiles.push('pytest.ini');
    }
  }

  /**
   * Create Node.js/TypeScript project files
   * @param projectPath - Project directory path
   * @param templateData - Template data
   * @param result - Result object to update
   * @tags @API:TEMPLATE-NODEJS-001
   */
  private async createNodeJSFiles(
    projectPath: string,
    templateData: TemplateData,
    result: InitResult
  ): Promise<void> {
    // package.json
    const packageJsonContent = this.generatePackageJson(templateData);
    const packageJsonPath = path.join(projectPath, 'package.json');
    await fs.writeFile(
      packageJsonPath,
      JSON.stringify(packageJsonContent, null, 2)
    );
    result.createdFiles.push('package.json');

    // tsconfig.json (for TypeScript projects)
    if (
      templateData.projectType === ProjectType.TYPESCRIPT ||
      templateData.features['typescript']
    ) {
      const tsconfigContent = this.generateTsConfig();
      const tsconfigPath = path.join(projectPath, 'tsconfig.json');
      await fs.writeFile(
        tsconfigPath,
        JSON.stringify(tsconfigContent, null, 2)
      );
      result.createdFiles.push('tsconfig.json');
    }

    // Jest configuration
    if (templateData.features['jest']) {
      const jestConfigContent = this.generateJestConfig();
      const jestConfigPath = path.join(projectPath, 'jest.config.js');
      await fs.writeFile(jestConfigPath, jestConfigContent);
      result.createdFiles.push('jest.config.js');
    }
  }

  /**
   * Create frontend project files
   * @param projectPath - Project directory path
   * @param templateData - Template data
   * @param result - Result object to update
   * @tags @API:TEMPLATE-FRONTEND-001
   */
  private async createFrontendFiles(
    projectPath: string,
    templateData: TemplateData,
    result: InitResult
  ): Promise<void> {
    // Create package.json for frontend
    await this.createNodeJSFiles(projectPath, templateData, result);

    // Add frontend-specific directories
    await this.ensureDirectory(path.join(projectPath, 'public'));
    await this.ensureDirectory(path.join(projectPath, 'src', 'components'));
  }

  /**
   * Create mixed project files (full-stack)
   * @param projectPath - Project directory path
   * @param templateData - Template data
   * @param result - Result object to update
   * @tags @API:TEMPLATE-MIXED-001
   */
  private async createMixedProjectFiles(
    projectPath: string,
    templateData: TemplateData,
    result: InitResult
  ): Promise<void> {
    // Create backend directory structure
    const backendPath = path.join(projectPath, 'backend');
    await this.ensureDirectory(backendPath);

    if (templateData.features['backend-python']) {
      await this.createPythonFiles(backendPath, templateData, result);
      // Update file paths to include backend/
      const lastFiles = result.createdFiles.splice(-2);
      result.createdFiles.push(...lastFiles.map(f => `backend/${f}`));
    }

    // Create frontend directory structure
    const frontendPath = path.join(projectPath, 'frontend');
    await this.ensureDirectory(frontendPath);

    if (templateData.features['frontend-typescript']) {
      await this.createNodeJSFiles(frontendPath, templateData, result);
      // Update file paths to include frontend/
      const lastFiles = result.createdFiles.splice(-2);
      result.createdFiles.push(...lastFiles.map(f => `frontend/${f}`));
    }
  }

  /**
   * Create .moai directory structure
   * @param projectPath - Project directory path
   * @param templateData - Template data
   * @param result - Result object to update
   * @tags @API:TEMPLATE-MOAI-001
   */
  private async createMoaiStructure(
    projectPath: string,
    templateData: TemplateData,
    result: InitResult
  ): Promise<void> {
    // .moai/config.json
    const moaiConfig = this.generateMoaiConfig(templateData);
    const configPath = path.join(projectPath, '.moai', 'config.json');
    await fs.writeFile(configPath, JSON.stringify(moaiConfig, null, 2));
    result.createdFiles.push('.moai/config.json');

    // .moai/project files
    const projectFiles = ['product.md', 'structure.md', 'tech.md'];
    for (const file of projectFiles) {
      const content = this.generateProjectFile(file, templateData);
      const filePath = path.join(projectPath, '.moai', 'project', file);
      await fs.writeFile(filePath, content);
      result.createdFiles.push(`.moai/project/${file}`);
    }

    // NOTE: [v0.0.3+] .moai/indexes/tags.db 제거 - CODE-FIRST 방식으로 전환
    // TAG 인덱스 파일 불필요 - 코드에서 직접 스캔하는 방식으로 변경됨

    // .moai/reports/sync-report.md
    const syncReportContent = this.generateSyncReport(templateData);
    const syncReportPath = path.join(
      projectPath,
      '.moai',
      'reports',
      'sync-report.md'
    );
    await fs.writeFile(syncReportPath, syncReportContent);
    result.createdFiles.push('.moai/reports/sync-report.md');
  }

  /**
   * Create .claude directory structure
   * @param projectPath - Project directory path
   * @param templateData - Template data
   * @param result - Result object to update
   * @tags @API:TEMPLATE-CLAUDE-001
   */
  private async createClaudeStructure(
    projectPath: string,
    templateData: TemplateData,
    result: InitResult
  ): Promise<void> {
    // Create .claude directories
    const claudeDirs = [
      '.claude',
      '.claude/agents',
      '.claude/agents/moai',
      '.claude/commands',
      '.claude/commands/moai',
      '.claude/hooks',
      '.claude/hooks/moai',
    ];

    for (const dir of claudeDirs) {
      await this.ensureDirectory(path.join(projectPath, dir));
    }

    // Create agent files
    const agents = ['spec-builder.md', 'code-builder.md', 'doc-syncer.md'];
    for (const agent of agents) {
      const content = this.generateAgentFile(agent, templateData);
      const agentPath = path.join(
        projectPath,
        '.claude',
        'agents',
        'moai',
        agent
      );
      await fs.writeFile(agentPath, content);
      result.createdFiles.push(`.claude/agents/moai/${agent}`);
    }

    // Create command files
    const commands = ['8-project.md', '1-spec.md', '2-build.md', '3-sync.md'];
    for (const command of commands) {
      const content = this.generateCommandFile(command, templateData);
      const commandPath = path.join(
        projectPath,
        '.claude',
        'commands',
        'moai',
        command
      );
      await fs.writeFile(commandPath, content);
      result.createdFiles.push(`.claude/commands/moai/${command}`);
    }

    // Create hook files
    const hookContent = this.generatePreCommitHook(templateData);
    const hookPath = path.join(
      projectPath,
      '.claude',
      'hooks',
      'moai',
      'pre-commit.py'
    );
    await fs.writeFile(hookPath, hookContent);
    result.createdFiles.push('.claude/hooks/moai/pre-commit.py');
  }

  /**
   * Check if project has specific feature enabled
   * @param config - Project configuration
   * @param featureName - Feature name to check
   * @returns Whether feature is enabled
   * @tags @API:TEMPLATE-FEATURE-001
   */
  private hasFeature(config: ProjectConfig, featureName: string): boolean {
    return (
      config.features?.some(f => f.name === featureName && f.enabled) || false
    );
  }

  /**
   * Ensure directory exists, create if needed
   * @param dirPath - Directory path
   * @tags @API:TEMPLATE-DIR-001
   */
  private async ensureDirectory(dirPath: string): Promise<void> {
    try {
      await fs.access(dirPath);
    } catch {
      await fs.mkdir(dirPath, { recursive: true });
    }
  }

  // Template generation methods (minimal implementations for GREEN phase)
  private generatePyprojectToml(data: TemplateData): string {
    return `[build-system]
requires = ["setuptools>=45", "wheel"]
build-backend = "setuptools.build_meta"

[project]
name = "${data.projectName}"
description = "${data.description}"
authors = [{name = "${data.author}"}]
license = {text = "${data.license}"}
version = "0.1.0"
`;
  }

  private generatePackageJson(data: TemplateData): any {
    return {
      name: data.projectName,
      version: '0.1.0',
      description: data.description,
      author: data.author,
      license: data.license,
      scripts: {
        build: 'tsc',
        test: 'jest',
        start: 'node dist/index.js',
      },
      dependencies: {},
      devDependencies: {},
    };
  }

  private generateTsConfig(): any {
    return {
      compilerOptions: {
        target: 'ES2020',
        module: 'commonjs',
        outDir: './dist',
        rootDir: './src',
        strict: true,
        esModuleInterop: true,
      },
      include: ['src/**/*'],
      exclude: ['node_modules', 'dist'],
    };
  }

  private generateJestConfig(): string {
    return `module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  roots: ['<rootDir>/src'],
  testMatch: ['**/__tests__/**/*.test.ts']
};`;
  }

  private generatePytestConfig(): string {
    return `[tool:pytest]
testpaths = tests
python_files = test_*.py
python_classes = Test*
python_functions = test_*
`;
  }

  private generateMoaiConfig(data: TemplateData): any {
    return {
      project: {
        name: data.projectName,
        type: data.projectType,
        version: '0.1.0',
        created_at: data.timestamp,
      },
      constitution: {
        enforce_tdd: true,
        test_coverage_target: 85,
      },
    };
  }

  private generateProjectFile(filename: string, data: TemplateData): string {
    return `# ${data.projectName} ${filename.replace('.md', '').toUpperCase()}

Generated on ${data.timestamp}
Project Type: ${data.projectType}
Author: ${data.author}

This file was auto-generated by MoAI-ADK.
`;
  }

  private generateSyncReport(data: TemplateData): string {
    return `# Sync Report - ${data.projectName}

Generated: ${data.timestamp}

## Status
- Project initialized successfully
- All required files created
- Ready for development

## Next Steps
1. Run \`/moai:1-spec\` to create your first specification
2. Begin TDD development with \`/moai:2-build\`
`;
  }

  private generateAgentFile(filename: string, data: TemplateData): string {
    const agentName = filename.replace('.md', '');
    return `# ${agentName} Agent

Agent for ${data.projectName}
Generated: ${data.timestamp}

This agent handles ${agentName} functionality.
`;
  }

  private generateCommandFile(filename: string, data: TemplateData): string {
    const commandName = filename.replace('.md', '');
    return `# ${commandName} Command

Command for ${data.projectName}
Generated: ${data.timestamp}

Handles ${commandName} workflow step.
`;
  }

  private generatePreCommitHook(data: TemplateData): string {
    return `#!/usr/bin/env python3
"""
Pre-commit hook for ${data.projectName}
Generated: ${data.timestamp}
"""

import sys

def main():
    print("Running pre-commit checks...")
    # Add validation logic here
    return 0

if __name__ == "__main__":
    sys.exit(main())
`;
  }
}

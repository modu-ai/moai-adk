// @CODE:REFACTOR-002 |
// Related: @CODE:TEMPLATE-MANAGER-REFACTOR-001

/**
 * @file Project template management - Refactored Version
 * @author MoAI Team
 * @tags @CODE:TEMPLATE-MANAGER-REFACTOR-001
 *
 * Phase 3: Refactored TemplateManager
 * - Uses TemplateValidator for validation
 * - Uses TemplateProcessor for content generation
 * - Focuses on orchestration and file I/O
 * - Reduced from 609 LOC to < 300 LOC
 */

import * as fs from 'node:fs/promises';
import * as path from 'node:path';
import {
  type InitResult,
  type ProjectConfig,
  ProjectType,
  type TemplateData,
} from '@/types/project';
import { getTemplatesPath } from '../../utils/package-root.js';
import { TemplateProcessor } from './template-processor.js';
import { TemplateValidator } from './template-validator.js';

/**
 * Template manager for project structure generation
 * Orchestrates validation and file generation using injected dependencies
 * @tags @CODE:TEMPLATE-MANAGER-REFACTOR-001
 */
export class TemplateManager {
  private readonly templatesPath: string;
  private readonly validator: TemplateValidator;
  private readonly processor: TemplateProcessor;

  /**
   * Create a new TemplateManager instance
   * @param templatesPath - Optional explicit templates path
   * @param validator - Optional validator instance (for testing)
   * @param processor - Optional processor instance (for testing)
   */
  constructor(
    templatesPath?: string,
    validator?: TemplateValidator,
    processor?: TemplateProcessor
  ) {
    this.templatesPath = templatesPath || getTemplatesPath(import.meta.url);
    this.validator = validator || new TemplateValidator();
    this.processor = processor || new TemplateProcessor();
  }

  /**
   * Get templates directory path
   * @returns Templates directory path
   * @tags @CODE:TEMPLATE-PATH-001:API
   */
  public getTemplatesPath(): string {
    return this.templatesPath;
  }

  /**
   * Generate complete project structure from configuration
   * @param config - Project configuration
   * @param targetPath - Target directory path
   * @returns Generation result
   * @tags @CODE:TEMPLATE-GENERATE-001:API
   */
  public async generateProject(
    config: ProjectConfig,
    targetPath: string
  ): Promise<InitResult> {
    const result: InitResult = {
      success: false,
      projectPath: targetPath,
      config,
      createdFiles: [],
      errors: [],
      warnings: [],
    };

    try {
      // Phase 1: Validate configuration
      const validationResult = this.validator.validateConfig(config);
      if (!validationResult.isValid) {
        result.errors = validationResult.errors;
        return result;
      }

      // Phase 2: Validate target path
      if (!this.validator.validatePath(targetPath)) {
        result.errors = ['Invalid or unsafe target path'];
        return result;
      }

      // Phase 3: Create project directory
      const projectPath = path.join(targetPath, config.name);
      await this.ensureDirectory(projectPath);

      // Phase 4: Generate template data
      const templateData = this.processor.createTemplateData(config);

      // Phase 5: Create project structure
      await this.createBaseStructure(projectPath, result);
      await this.createProjectFiles(projectPath, config, templateData, result);
      await this.createMoaiStructure(projectPath, templateData, result);

      // Phase 6: Create Claude structure if enabled
      if (this.hasFeature(config, 'claude-integration')) {
        await this.createClaudeStructure(projectPath, templateData, result);
      }

      result.success = true;
      return result;
    } catch (error) {
      result.errors = [
        error instanceof Error ? error.message : 'Unknown error',
      ];
      return result;
    }
  }

  /**
   * Create base project structure
   * @param projectPath - Project directory path
   * @param result - Result object to update
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
   */
  private async createPythonFiles(
    projectPath: string,
    templateData: TemplateData,
    result: InitResult
  ): Promise<void> {
    // pyproject.toml
    const pyprojectContent = this.processor.generatePyprojectToml(templateData);
    await fs.writeFile(
      path.join(projectPath, 'pyproject.toml'),
      pyprojectContent
    );
    result.createdFiles.push('pyproject.toml');

    // src/__init__.py
    await fs.writeFile(
      path.join(projectPath, 'src', '__init__.py'),
      '# Package initialization\n'
    );
    result.createdFiles.push('src/__init__.py');

    // pytest.ini if feature enabled
    if (templateData.features.pytest) {
      const pytestConfig = this.processor.generatePytestConfig();
      await fs.writeFile(path.join(projectPath, 'pytest.ini'), pytestConfig);
      result.createdFiles.push('pytest.ini');
    }
  }

  /**
   * Create Node.js/TypeScript project files
   */
  private async createNodeJSFiles(
    projectPath: string,
    templateData: TemplateData,
    result: InitResult
  ): Promise<void> {
    // package.json
    const packageJson = this.processor.generatePackageJson(templateData);
    await fs.writeFile(
      path.join(projectPath, 'package.json'),
      JSON.stringify(packageJson, null, 2)
    );
    result.createdFiles.push('package.json');

    // tsconfig.json for TypeScript
    if (
      templateData.projectType === ProjectType.TYPESCRIPT ||
      templateData.features.typescript
    ) {
      const tsconfig = this.processor.generateTsConfig();
      await fs.writeFile(
        path.join(projectPath, 'tsconfig.json'),
        JSON.stringify(tsconfig, null, 2)
      );
      result.createdFiles.push('tsconfig.json');
    }

    // jest.config.js if feature enabled
    if (templateData.features.jest) {
      const jestConfig = this.processor.generateJestConfig();
      await fs.writeFile(path.join(projectPath, 'jest.config.js'), jestConfig);
      result.createdFiles.push('jest.config.js');
    }
  }

  /**
   * Create frontend project files
   */
  private async createFrontendFiles(
    projectPath: string,
    templateData: TemplateData,
    result: InitResult
  ): Promise<void> {
    await this.createNodeJSFiles(projectPath, templateData, result);
    await this.ensureDirectory(path.join(projectPath, 'public'));
    await this.ensureDirectory(path.join(projectPath, 'src', 'components'));
  }

  /**
   * Create mixed (full-stack) project files
   */
  private async createMixedProjectFiles(
    projectPath: string,
    templateData: TemplateData,
    result: InitResult
  ): Promise<void> {
    // Backend
    const backendPath = path.join(projectPath, 'backend');
    await this.ensureDirectory(backendPath);
    if (templateData.features['backend-python']) {
      await this.createPythonFiles(backendPath, templateData, result);
      const lastFiles = result.createdFiles.splice(-2);
      result.createdFiles.push(...lastFiles.map(f => `backend/${f}`));
    }

    // Frontend
    const frontendPath = path.join(projectPath, 'frontend');
    await this.ensureDirectory(frontendPath);
    if (templateData.features['frontend-typescript']) {
      await this.createNodeJSFiles(frontendPath, templateData, result);
      const lastFiles = result.createdFiles.splice(-2);
      result.createdFiles.push(...lastFiles.map(f => `frontend/${f}`));
    }
  }

  /**
   * Create .moai directory structure
   */
  private async createMoaiStructure(
    projectPath: string,
    templateData: TemplateData,
    result: InitResult
  ): Promise<void> {
    // .moai/config.json
    const moaiConfig = this.processor.generateMoaiConfig(templateData);
    await fs.writeFile(
      path.join(projectPath, '.moai', 'config.json'),
      JSON.stringify(moaiConfig, null, 2)
    );
    result.createdFiles.push('.moai/config.json');

    // .moai/project files
    for (const file of ['product.md', 'structure.md', 'tech.md']) {
      const content = this.processor.generateProjectFile(file, templateData);
      await fs.writeFile(
        path.join(projectPath, '.moai', 'project', file),
        content
      );
      result.createdFiles.push(`.moai/project/${file}`);
    }

    // .moai/reports/sync-report.md
    const syncReport = this.processor.generateSyncReport(templateData);
    await fs.writeFile(
      path.join(projectPath, '.moai', 'reports', 'sync-report.md'),
      syncReport
    );
    result.createdFiles.push('.moai/reports/sync-report.md');
  }

  /**
   * Create .claude directory structure
   */
  private async createClaudeStructure(
    projectPath: string,
    templateData: TemplateData,
    result: InitResult
  ): Promise<void> {
    const claudeDirs = [
      '.claude',
      '.claude/agents',
      '.claude/agents/alfred',
      '.claude/commands',
      '.claude/commands/alfred',
      '.claude/hooks',
      '.claude/hooks/alfred',
    ];

    for (const dir of claudeDirs) {
      await this.ensureDirectory(path.join(projectPath, dir));
    }

    // Agent files
    for (const agent of [
      'spec-builder.md',
      'code-builder.md',
      'doc-syncer.md',
    ]) {
      const content = this.processor.generateAgentFile(agent, templateData);
      await fs.writeFile(
        path.join(projectPath, '.claude', 'agents', 'moai', agent),
        content
      );
      result.createdFiles.push(`.claude/agents/alfred/${agent}`);
    }

    // Command files
    for (const cmd of [
      '8-project.md',
      '1-spec.md',
      '2-build.md',
      '3-sync.md',
    ]) {
      const content = this.processor.generateCommandFile(cmd, templateData);
      await fs.writeFile(
        path.join(projectPath, '.claude', 'commands', 'moai', cmd),
        content
      );
      result.createdFiles.push(`.claude/commands/alfred/${cmd}`);
    }

    // Note: Python hooks have been replaced with TypeScript hooks
    // See: src/claude/hooks/* for TS-based hook implementations
  }

  /**
   * Check if project has specific feature enabled
   */
  private hasFeature(config: ProjectConfig, featureName: string): boolean {
    return (
      config.features?.some(f => f.name === featureName && f.enabled) || false
    );
  }

  /**
   * Ensure directory exists, create if needed
   */
  private async ensureDirectory(dirPath: string): Promise<void> {
    try {
      await fs.access(dirPath);
    } catch {
      await fs.mkdir(dirPath, { recursive: true });
    }
  }
}

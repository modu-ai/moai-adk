/**
 * @file Interactive project wizard
 * @author MoAI Team
 * @tags @FEATURE:PROJECT-WIZARD-001 @REQ:CLI-WIZARD-001
 */

import inquirer from 'inquirer';
import { logger } from '../../utils/winston-logger.js';
import {
  type ProjectConfig,
  type ProjectFeature,
  ProjectType,
} from '@/types/project';

/**
 * Interactive wizard for project configuration
 * @tags @FEATURE:PROJECT-WIZARD-001
 */
export class ProjectWizard {
  /**
   * Run interactive project configuration wizard
   * @returns Configured project settings
   * @tags @API:WIZARD-RUN-001
   */
  public async run(): Promise<ProjectConfig> {
    logger.info('🧙‍♂️ Project Configuration Wizard');
    logger.info('Step 1/5: Basic Information\n');

    const basicAnswers = await inquirer.prompt([
      {
        type: 'input',
        name: 'projectName',
        message: 'Project name:',
        validate: (input: string) => {
          if (!input || input.trim().length === 0) {
            return 'Project name is required';
          }
          if (!/^[a-zA-Z0-9-_]+$/.test(input)) {
            return 'Invalid project name format';
          }
          return true;
        },
      },
      {
        type: 'list',
        name: 'projectType',
        message: 'Project type:',
        choices: [
          {
            name: 'Python - Backend/CLI applications',
            value: ProjectType.PYTHON,
          },
          {
            name: 'Node.js - JavaScript/TypeScript applications',
            value: ProjectType.NODEJS,
          },
          {
            name: 'TypeScript - Pure TypeScript projects',
            value: ProjectType.TYPESCRIPT,
          },
          {
            name: 'Frontend - React/Vue/Angular applications',
            value: ProjectType.FRONTEND,
          },
          { name: 'Mixed - Full-stack applications', value: ProjectType.MIXED },
        ],
      },
      {
        type: 'input',
        name: 'description',
        message: 'Project description (optional):',
        default: '',
      },
      {
        type: 'input',
        name: 'author',
        message: 'Author name:',
        default: 'MoAI Developer',
      },
    ]);

    // Validate project name format
    if (!/^[a-zA-Z0-9-_]+$/.test(basicAnswers.projectName)) {
      throw new Error('Invalid project name format');
    }

    logger.info('\nStep 2/5: License and Package Manager\n');

    const configAnswers = await inquirer.prompt([
      {
        type: 'list',
        name: 'license',
        message: 'License:',
        choices: ['MIT', 'Apache-2.0', 'GPL-3.0', 'BSD-3-Clause', 'Unlicense'],
        default: 'MIT',
      },
      {
        type: 'list',
        name: 'packageManager',
        message: 'Package manager:',
        choices: ['npm', 'yarn', 'pnpm'],
        default: 'npm',
      },
    ]);

    logger.info('\nStep 3/5: Project Features\n');

    const features = await this.collectFeatures(basicAnswers.projectType);

    return {
      name: basicAnswers.projectName,
      type: basicAnswers.projectType,
      description: basicAnswers.description,
      author: basicAnswers.author,
      license: configAnswers.license,
      packageManager: configAnswers.packageManager,
      features,
    };
  }

  /**
   * Collect project-specific features based on type
   * @param projectType - The selected project type
   * @returns Array of selected features
   * @tags @API:WIZARD-FEATURES-001
   */
  private async collectFeatures(
    projectType: ProjectType
  ): Promise<ProjectFeature[]> {
    const featureChoices = this.getFeatureChoices(projectType);

    const featureAnswers = await inquirer.prompt([
      {
        type: 'checkbox',
        name: 'selectedFeatures',
        message: 'Select features to include:',
        choices: featureChoices,
      },
    ]);

    return featureAnswers.selectedFeatures.map((featureName: string) => ({
      name: featureName,
      enabled: true,
      config: {},
    }));
  }

  /**
   * Get available feature choices for project type
   * @param projectType - The project type
   * @returns Available feature choices
   * @tags @API:WIZARD-CHOICES-001
   */
  private getFeatureChoices(
    projectType: ProjectType
  ): Array<{ name: string; value: string }> {
    const commonFeatures = [
      { name: 'Testing framework', value: 'testing' },
      { name: 'Linting (code quality)', value: 'linting' },
      { name: 'Documentation generation', value: 'documentation' },
      { name: 'CI/CD configuration', value: 'ci-cd' },
    ];

    switch (projectType) {
      case ProjectType.PYTHON:
        return [
          ...commonFeatures,
          { name: 'Pytest testing', value: 'pytest' },
          { name: 'Black formatter', value: 'black' },
          { name: 'MyPy type checking', value: 'mypy' },
          { name: 'FastAPI web framework', value: 'fastapi' },
        ];

      case ProjectType.NODEJS:
      case ProjectType.TYPESCRIPT:
        return [
          ...commonFeatures,
          { name: 'Jest testing', value: 'jest' },
          { name: 'ESLint linting', value: 'eslint' },
          { name: 'Prettier formatting', value: 'prettier' },
          { name: 'Express.js framework', value: 'express' },
        ];

      case ProjectType.FRONTEND:
        return [
          ...commonFeatures,
          { name: 'React framework', value: 'react' },
          { name: 'Vue.js framework', value: 'vue' },
          { name: 'Tailwind CSS', value: 'tailwind' },
          { name: 'Webpack bundling', value: 'webpack' },
        ];

      case ProjectType.MIXED:
        return [
          ...commonFeatures,
          { name: 'Backend: Python + FastAPI', value: 'backend-python' },
          {
            name: 'Frontend: TypeScript + React',
            value: 'frontend-typescript',
          },
          { name: 'Docker containerization', value: 'docker' },
          { name: 'Claude integration', value: 'claude-integration' },
          { name: 'Database integration', value: 'database' },
          { name: 'API documentation', value: 'api-docs' },
        ];

      default:
        return commonFeatures;
    }
  }

  /**
   * Get description for project type
   * @param projectType - The project type
   * @returns Human-readable description
   * @tags @API:WIZARD-DESCRIPTION-001
   */
  public getProjectTypeDescription(projectType: ProjectType): string {
    const descriptions = {
      [ProjectType.PYTHON]:
        'Python backend applications, CLI tools, and data processing scripts',
      [ProjectType.NODEJS]:
        'Node.js applications with JavaScript or TypeScript',
      [ProjectType.TYPESCRIPT]:
        'Pure TypeScript projects with advanced type safety',
      [ProjectType.FRONTEND]:
        'Frontend web applications using modern frameworks',
      [ProjectType.MIXED]:
        'Full-stack applications with both backend and frontend components',
    };

    return descriptions[projectType] || 'Custom project configuration';
  }
}

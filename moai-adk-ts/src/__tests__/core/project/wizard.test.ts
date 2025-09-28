/**
 * @file Project wizard test suite
 * @author MoAI Team
 * @tags @TEST:PROJECT-WIZARD-001 @REQ:CLI-WIZARD-001
 */

import { describe, test, expect, beforeEach, jest, vi } from 'vitest';
import '@/__tests__/setup';
import { ProjectWizard } from '@/core/project/wizard';
import { ProjectType, ProjectConfig } from '@/types/project';
import inquirer from 'inquirer';

// Mock inquirer
vi.mock('inquirer');
const mockInquirer = inquirer as vi.Mocked<typeof inquirer>;

describe('ProjectWizard', () => {
  let wizard: ProjectWizard;

  beforeEach(() => {
    vi.clearAllMocks();
    wizard = new ProjectWizard();
  });

  describe('Interactive Project Configuration', () => {
    test('should collect complete project configuration through prompts', async () => {
      // Arrange
      const mockAnswers = {
        projectName: 'my-test-project',
        projectType: ProjectType.PYTHON,
        description: 'A test project for Python development',
        author: 'Test Developer',
        license: 'MIT',
        packageManager: 'npm',
        features: ['testing', 'linting', 'documentation'],
      };

      mockInquirer.prompt.mockResolvedValue(mockAnswers);

      // Act
      const result = await wizard.run();

      // Assert
      expect(result.name).toBe('my-test-project');
      expect(result.type).toBe(ProjectType.PYTHON);
      expect(result.description).toBe('A test project for Python development');
      expect(result.author).toBe('Test Developer');
      expect(result.license).toBe('MIT');
      expect(result.packageManager).toBe('npm');
      expect(result.features).toHaveLength(3);
    });

    test('should validate project name format during input', async () => {
      // Arrange
      const invalidNameAnswers = {
        projectName: 'Invalid Name!',
        projectType: ProjectType.NODEJS,
      };

      mockInquirer.prompt.mockResolvedValue(invalidNameAnswers);

      // Act & Assert
      await expect(wizard.run()).rejects.toThrow('Invalid project name format');
    });

    test('should provide different feature options based on project type', async () => {
      // Arrange - Python project
      const pythonAnswers = {
        projectName: 'python-project',
        projectType: ProjectType.PYTHON,
        features: ['pytest', 'black', 'mypy'],
      };

      mockInquirer.prompt.mockResolvedValueOnce(pythonAnswers);

      // Act
      const pythonResult = await wizard.run();

      // Reset for Node.js test
      vi.clearAllMocks();

      // Arrange - Node.js project
      const nodeAnswers = {
        projectName: 'node-project',
        projectType: ProjectType.NODEJS,
        features: ['jest', 'eslint', 'prettier'],
      };

      mockInquirer.prompt.mockResolvedValueOnce(nodeAnswers);

      // Act
      const nodeResult = await wizard.run();

      // Assert
      expect(pythonResult.features?.map(f => f.name)).toEqual([
        'pytest',
        'black',
        'mypy',
      ]);
      expect(nodeResult.features?.map(f => f.name)).toEqual([
        'jest',
        'eslint',
        'prettier',
      ]);
    });

    test('should handle mixed project type with both backend and frontend options', async () => {
      // Arrange
      const mixedAnswers = {
        projectName: 'fullstack-app',
        projectType: ProjectType.MIXED,
        backendLanguage: 'python',
        frontendFramework: 'react',
        features: ['docker', 'ci-cd', 'monitoring'],
      };

      mockInquirer.prompt.mockResolvedValue(mixedAnswers);

      // Act
      const result = await wizard.run();

      // Assert
      expect(result.type).toBe(ProjectType.MIXED);
      expect(result.features?.some(f => f.name === 'docker')).toBe(true);
    });
  });

  describe('Package Manager Detection and Selection', () => {
    test('should auto-detect available package managers', async () => {
      // Arrange
      const answers = {
        projectName: 'package-manager-test',
        projectType: ProjectType.NODEJS,
        autoDetectPackageManager: true,
      };

      mockInquirer.prompt.mockResolvedValue(answers);

      // Mock package manager detection
      vi.spyOn(wizard as any, 'detectPackageManagers').mockResolvedValue([
        'npm',
        'yarn',
        'pnpm',
      ]);

      // Act
      const result = await wizard.run();

      // Assert
      expect(result.packageManager).toBeOneOf(['npm', 'yarn', 'pnpm']);
    });

    test('should allow manual package manager selection', async () => {
      // Arrange
      const answers = {
        projectName: 'manual-pm-test',
        projectType: ProjectType.NODEJS,
        autoDetectPackageManager: false,
        packageManager: 'yarn',
      };

      mockInquirer.prompt.mockResolvedValue(answers);

      // Act
      const result = await wizard.run();

      // Assert
      expect(result.packageManager).toBe('yarn');
    });
  });

  describe('User Experience and Validation', () => {
    test('should provide helpful descriptions for each project type', async () => {
      // This test ensures the wizard provides clear information to users
      const wizard = new ProjectWizard();

      // Verify that the wizard has descriptions for all project types
      const projectTypes = Object.values(ProjectType);

      for (const type of projectTypes) {
        expect(wizard.getProjectTypeDescription(type)).toBeDefined();
        expect(wizard.getProjectTypeDescription(type).length).toBeGreaterThan(
          10
        );
      }
    });

    test('should show progress indicators during multi-step configuration', async () => {
      // Arrange
      const answers = {
        projectName: 'progress-test',
        projectType: ProjectType.MIXED,
      };

      mockInquirer.prompt.mockResolvedValue(answers);

      // Mock console.log to capture progress messages
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation();

      // Act
      await wizard.run();

      // Assert
      expect(consoleSpy).toHaveBeenCalledWith(expect.stringContaining('Step'));
      consoleSpy.mockRestore();
    });

    test('should allow users to go back and modify previous choices', async () => {
      // Arrange
      const firstAttempt = {
        projectName: 'first-attempt',
        projectType: ProjectType.PYTHON,
        goBack: true,
      };

      const secondAttempt = {
        projectName: 'corrected-name',
        projectType: ProjectType.NODEJS,
        goBack: false,
      };

      mockInquirer.prompt
        .mockResolvedValueOnce(firstAttempt)
        .mockResolvedValueOnce(secondAttempt);

      // Act
      const result = await wizard.run();

      // Assert
      expect(result.name).toBe('corrected-name');
      expect(result.type).toBe(ProjectType.NODEJS);
    });
  });
});

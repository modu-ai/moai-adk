/**
 * @file Init command test suite
 * @author MoAI Team
 * @tags @TEST:CLI-INIT-001 @REQ:CLI-INIT-002
 */

import { beforeEach, describe, expect, test, vi, type Mocked } from 'vitest';
import '@/__tests__/setup';
import type { DoctorResult } from '@/cli/commands/doctor';
import { InitCommand } from '@/cli/commands/init';
import { TemplateManager } from '@/core/project/template-manager';
import { ProjectWizard } from '@/core/project/wizard';
import { SystemDetector } from '@/core/system-checker/detector';
import { type ProjectConfig, ProjectType } from '@/types/project';

// Mock modules
vi.mock('@/core/project/wizard');
vi.mock('@/core/project/template-manager');
vi.mock('@/core/system-checker/detector');

describe('InitCommand Advanced Features', () => {
  let initCommand: InitCommand;
  let mockDetector: Mocked<SystemDetector>;
  let mockWizard: Mocked<ProjectWizard>;
  let mockTemplateManager: Mocked<TemplateManager>;

  beforeEach(() => {
    // Reset all mocks
    vi.clearAllMocks();

    // Setup mock instances
    mockDetector = new SystemDetector() as unknown as Mocked<SystemDetector>;
    mockWizard = new ProjectWizard() as unknown as Mocked<ProjectWizard>;
    mockTemplateManager = new TemplateManager() as unknown as Mocked<TemplateManager>;

    initCommand = new InitCommand(
      mockDetector,
      mockWizard,
      mockTemplateManager
    );
  });

  describe('Project Type Detection and Template Selection', () => {
    test('should create Python project structure with correct templates', async () => {
      // Arrange
      const projectConfig: ProjectConfig = {
        name: 'test-python-project',
        type: ProjectType.PYTHON,
        description: 'Test Python project',
        author: 'Test Author',
        packageManager: 'npm',
      };

      mockWizard.run.mockResolvedValue(projectConfig);
      mockTemplateManager.generateProject.mockResolvedValue({
        success: true,
        projectPath: '/test/path',
        config: projectConfig,
        createdFiles: [
          'pyproject.toml',
          'src/__init__.py',
          '.moai/config.json',
          '.claude/agents/moai/spec-builder.md',
        ],
      });

      // Act
      const result = await initCommand.runInteractive();

      // Assert
      expect(result.success).toBe(true);
      expect(result.config.type).toBe(ProjectType.PYTHON);
      expect(result.createdFiles).toContain('pyproject.toml');
      expect(result.createdFiles).toContain('.moai/config.json');
      expect(mockTemplateManager.generateProject).toHaveBeenCalledWith(
        projectConfig,
        expect.any(String)
      );
    });

    test('should create Node.js project with package.json and npm configuration', async () => {
      // Arrange
      const projectConfig: ProjectConfig = {
        name: 'test-node-project',
        type: ProjectType.NODEJS,
        packageManager: 'yarn',
        features: [
          { name: 'typescript', enabled: true },
          { name: 'testing', enabled: true },
        ],
      };

      mockWizard.run.mockResolvedValue(projectConfig);
      mockTemplateManager.generateProject.mockResolvedValue({
        success: true,
        projectPath: '/test/node/path',
        config: projectConfig,
        createdFiles: [
          'package.json',
          'tsconfig.json',
          'vi.config.js',
          '.moai/config.json',
          '.claude/commands/moai/init.md',
        ],
      });

      // Act
      const result = await initCommand.runInteractive();

      // Assert
      expect(result.success).toBe(true);
      expect(result.config.type).toBe(ProjectType.NODEJS);
      expect(result.config.packageManager).toBe('yarn');
      expect(result.createdFiles).toContain('package.json');
      expect(result.createdFiles).toContain('tsconfig.json');
    });

    test('should create mixed project with both Python and Node.js structure', async () => {
      // Arrange
      const projectConfig: ProjectConfig = {
        name: 'test-mixed-project',
        type: ProjectType.MIXED,
        packageManager: 'pnpm',
        features: [
          { name: 'backend-python', enabled: true },
          { name: 'frontend-typescript', enabled: true },
          { name: 'claude-integration', enabled: true },
        ],
      };

      mockWizard.run.mockResolvedValue(projectConfig);
      mockTemplateManager.generateProject.mockResolvedValue({
        success: true,
        projectPath: '/test/mixed/path',
        config: projectConfig,
        createdFiles: [
          'backend/pyproject.toml',
          'frontend/package.json',
          'frontend/tsconfig.json',
          '.moai/config.json',
          '.claude/agents/moai/code-builder.md',
        ],
      });

      // Act
      const result = await initCommand.runInteractive();

      // Assert
      expect(result.success).toBe(true);
      expect(result.config.type).toBe(ProjectType.MIXED);
      expect(result.createdFiles).toContain('backend/pyproject.toml');
      expect(result.createdFiles).toContain('frontend/package.json');
    });

    test('should fail when system requirements are not met', async () => {
      // Arrange
      // Mock DoctorCommand.run to return failed result
      vi.spyOn(initCommand as any, 'doctorCommand', 'get').mockReturnValue({
        run: vi.fn().mockResolvedValue({
          allPassed: false,
          results: [],
          missingRequirements: [],
          versionConflicts: [],
          summary: { total: 1, passed: 0, failed: 1 },
        }),
      });

      // Act
      const result = await initCommand.runInteractive();

      // Assert
      expect(result.success).toBe(false);
      expect(result.errors).toContain('System verification failed');
      expect(mockWizard.run).not.toHaveBeenCalled();
    });
  });

  describe('Interactive Wizard Integration', () => {
    test('should run interactive wizard and collect user preferences', async () => {
      // Arrange
      const expectedConfig: ProjectConfig = {
        name: 'my-awesome-project',
        type: ProjectType.TYPESCRIPT,
        description: 'An awesome TypeScript project',
        author: 'John Doe',
        license: 'MIT',
        packageManager: 'npm',
        features: [
          { name: 'eslint', enabled: true },
          { name: 'prettier', enabled: true },
          { name: 'jest', enabled: true },
        ],
      };

      // Mock successful doctor run
      const mockDoctorResult = {
        allPassed: true,
        results: [],
        missingRequirements: [],
        versionConflicts: [],
        summary: { total: 5, passed: 5, failed: 0 },
      } as DoctorResult;
      (initCommand as any).doctorCommand.run = vi
        .fn()
        .mockResolvedValue(mockDoctorResult) as any;
      mockWizard.run.mockResolvedValue(expectedConfig);
      mockTemplateManager.generateProject.mockResolvedValue({
        success: true,
        projectPath: '/generated/path',
        config: expectedConfig,
        createdFiles: ['package.json', 'tsconfig.json'],
      });

      // Act
      const result = await initCommand.runInteractive();

      // Assert
      expect(mockWizard.run).toHaveBeenCalledWith();
      expect(result.config).toEqual(expectedConfig);
      expect(mockTemplateManager.generateProject).toHaveBeenCalledWith(
        expectedConfig,
        expect.any(String)
      );
    });

    test('should handle wizard cancellation gracefully', async () => {
      // Arrange
      // Mock successful doctor run
      const mockDoctorResult = {
        allPassed: true,
        results: [],
        missingRequirements: [],
        versionConflicts: [],
        summary: { total: 5, passed: 5, failed: 0 },
      } as DoctorResult;
      (initCommand as any).doctorCommand.run = vi
        .fn()
        .mockResolvedValue(mockDoctorResult) as any;
      mockWizard.run.mockRejectedValue(new Error('User cancelled'));

      // Act
      const result = await initCommand.runInteractive();

      // Assert
      expect(result.success).toBe(false);
      expect(result.errors).toContain('User cancelled');
    });
  });

  describe('Directory Structure Generation', () => {
    test('should create .moai directory structure with all required files', async () => {
      // Arrange
      const projectConfig: ProjectConfig = {
        name: 'structure-test',
        type: ProjectType.PYTHON,
      };

      mockWizard.run.mockResolvedValue(projectConfig);
      mockTemplateManager.generateProject.mockResolvedValue({
        success: true,
        projectPath: '/test/structure',
        config: projectConfig,
        createdFiles: [
          '.moai/config.json',
          '.moai/project/product.md',
          '.moai/project/structure.md',
          '.moai/project/tech.md',
          '.moai/reports/sync-report.md',
        ],
      });

      // Act
      const result = await initCommand.runInteractive();

      // Assert
      expect(result.createdFiles).toContain('.moai/config.json');
      expect(result.createdFiles).toContain('.moai/project/product.md');
      // NOTE: [v0.0.3+] .moai/indexes/tags.db 제거 - CODE-FIRST 방식으로 전환
    });

    test('should create .claude directory with agent configurations', async () => {
      // Arrange
      const projectConfig: ProjectConfig = {
        name: 'claude-test',
        type: ProjectType.NODEJS,
        features: [{ name: 'claude-integration', enabled: true }],
      };

      mockWizard.run.mockResolvedValue(projectConfig);
      mockTemplateManager.generateProject.mockResolvedValue({
        success: true,
        projectPath: '/test/claude',
        config: projectConfig,
        createdFiles: [
          '.claude/agents/moai/spec-builder.md',
          '.claude/agents/moai/code-builder.md',
          '.claude/agents/moai/doc-syncer.md',
          '.claude/commands/moai/0-project.md',
          '.claude/commands/moai/1-spec.md',
          '.claude/hooks/moai/pre-commit.py',
        ],
      });

      // Act
      const result = await initCommand.runInteractive();

      // Assert
      expect(result.createdFiles).toContain(
        '.claude/agents/moai/spec-builder.md'
      );
      expect(result.createdFiles).toContain(
        '.claude/commands/moai/0-project.md'
      );
      expect(result.createdFiles).toContain('.claude/hooks/moai/pre-commit.py');
    });
  });

  describe('Error Handling and Edge Cases', () => {
    test('should handle template generation failures gracefully', async () => {
      // Arrange
      const projectConfig: ProjectConfig = {
        name: 'fail-test',
        type: ProjectType.PYTHON,
      };

      mockWizard.run.mockResolvedValue(projectConfig);
      mockTemplateManager.generateProject.mockResolvedValue({
        success: false,
        projectPath: '',
        config: projectConfig,
        createdFiles: [],
        errors: ['Failed to create directory', 'Permission denied'],
      });

      // Act
      const result = await initCommand.runInteractive();

      // Assert
      expect(result.success).toBe(false);
      expect(result.errors).toContain('Failed to create directory');
    });

    test('should validate project name format', async () => {
      // Arrange
      const invalidConfig: ProjectConfig = {
        name: 'invalid name with spaces!',
        type: ProjectType.PYTHON,
      };

      mockWizard.run.mockResolvedValue(invalidConfig);

      // Act
      const result = await initCommand.runInteractive();

      // Assert
      expect(result.success).toBe(false);
      expect(result.errors).toContain('Invalid project name format');
    });
  });
});

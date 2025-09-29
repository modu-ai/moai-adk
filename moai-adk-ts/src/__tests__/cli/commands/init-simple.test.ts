/**
 * @file Simplified Init command test suite
 * @author MoAI Team
 * @tags @TEST:CLI-INIT-SIMPLE-001
 */

import { beforeEach, describe, expect, test, vi } from 'vitest';
import { InitCommand } from '@/cli/commands/init';
import { TemplateManager } from '@/core/project/template-manager';
import { ProjectWizard } from '@/core/project/wizard';
import { ProjectType } from '@/types/project';

describe('InitCommand Basic Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  test('should define InitCommand class', () => {
    expect(InitCommand).toBeDefined();
    expect(typeof InitCommand).toBe('function');
  });

  test('should define ProjectWizard class', () => {
    expect(ProjectWizard).toBeDefined();
    expect(typeof ProjectWizard).toBe('function');
  });

  test('should define TemplateManager class', () => {
    expect(TemplateManager).toBeDefined();
    expect(typeof TemplateManager).toBe('function');
  });

  test('should define project types', () => {
    expect(ProjectType).toBeDefined();
    expect(ProjectType.PYTHON).toBe('python');
    expect(ProjectType.NODEJS).toBe('nodejs');
    expect(ProjectType.TYPESCRIPT).toBe('typescript');
    expect(ProjectType.FRONTEND).toBe('frontend');
    expect(ProjectType.MIXED).toBe('mixed');
  });

  test('should create wizard with proper project type descriptions', () => {
    const wizard = new ProjectWizard();

    // Test each project type has a description
    const pythonDesc = wizard.getProjectTypeDescription(ProjectType.PYTHON);
    const nodeDesc = wizard.getProjectTypeDescription(ProjectType.NODEJS);
    const tsDesc = wizard.getProjectTypeDescription(ProjectType.TYPESCRIPT);
    const frontendDesc = wizard.getProjectTypeDescription(ProjectType.FRONTEND);
    const mixedDesc = wizard.getProjectTypeDescription(ProjectType.MIXED);

    expect(pythonDesc).toContain('Python');
    expect(nodeDesc).toContain('Node.js');
    expect(tsDesc).toContain('TypeScript');
    expect(frontendDesc).toContain('Frontend');
    expect(mixedDesc).toContain('Full-stack');
  });

  test('should create template manager with project generation capability', async () => {
    const templateManager = new TemplateManager();
    const mockConfig = {
      name: 'test-project',
      type: ProjectType.PYTHON,
      description: 'Test project',
      author: 'Test Author',
    };

    // Test that generateProject method exists and has correct signature
    expect(typeof templateManager.generateProject).toBe('function');

    // Test with invalid project name
    const invalidConfig = { ...mockConfig, name: 'invalid project name!' };
    const result = await templateManager.generateProject(invalidConfig, '/tmp');

    expect(result.success).toBe(false);
    expect(result.errors).toContain('Invalid project name format');
  });
});

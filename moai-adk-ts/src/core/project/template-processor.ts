// @CODE:REFACTOR-002-PROCESSOR | -PROCESSOR
// Related: @CODE:TEMPLATE-PROCESSOR-001

/**
 * @file Template processing and content generation
 * @author MoAI Team
 * @tags @CODE:TEMPLATE-PROCESSOR-001
 *
 * Phase 2: Template file content generation
 * - Template data creation
 * - Config file generation (Python, Node.js, TypeScript)
 * - MoAI structure generation
 */

import type { ProjectConfig, TemplateData } from '@/types/project';
import { ProjectType } from '@/types/project';

/**
 * Package.json structure
 */
interface PackageJson {
  name: string;
  version: string;
  description: string;
  author: string;
  license: string;
  scripts: Record<string, string>;
  dependencies: Record<string, string>;
  devDependencies: Record<string, string>;
}

/**
 * TypeScript configuration structure
 */
interface TsConfig {
  compilerOptions: {
    target: string;
    module: string;
    outDir: string;
    rootDir: string;
    strict: boolean;
    esModuleInterop: boolean;
  };
  include: string[];
  exclude: string[];
}

/**
 * MoAI configuration structure
 */
interface MoaiConfig {
  project: {
    name: string;
    type: ProjectType;
    version: string;
    created_at: string;
  };
  constitution: {
    enforce_tdd: boolean;
    test_coverage_target: number;
  };
}

/**
 * Template content generator
 * Extracted from template-manager.ts for better separation of concerns
 * Handles file content generation for various project types
 * @tags @CODE:TEMPLATE-PROCESSOR-001
 */
export class TemplateProcessor {
  /**
   * Create template data from project configuration
   * @param config - Project configuration
   * @returns Template data for file generation
   * @tags @CODE:PROCESSOR-DATA-001:API
   */
  public createTemplateData(config: ProjectConfig): TemplateData {
    return {
      projectName: config.name,
      projectType: config.type || ProjectType.TYPESCRIPT,
      timestamp: new Date().toISOString(),
      author: config.author || 'MoAI Developer',
      description:
        config.description ||
        `A ${(config.type || 'typescript').toLowerCase()} project built with MoAI-ADK`,
      license: config.license || 'MIT',
      packageManager: config.packageManager || 'npm',
      features: Object.fromEntries(
        (config.features || []).map(f => [f.name, f.enabled])
      ),
    };
  }

  /**
   * Generate pyproject.toml content for Python projects
   * @param data - Template data
   * @returns pyproject.toml content
   * @tags @CODE:PROCESSOR-PYTHON-001:API
   */
  public generatePyprojectToml(data: TemplateData): string {
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

  /**
   * Generate package.json content for Node.js projects
   * @param data - Template data
   * @returns package.json object
   * @tags @CODE:PROCESSOR-NODEJS-001:API
   */
  public generatePackageJson(data: TemplateData): PackageJson {
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

  /**
   * Generate tsconfig.json content
   * @returns TypeScript configuration object
   * @tags @CODE:PROCESSOR-TYPESCRIPT-001:API
   */
  public generateTsConfig(): TsConfig {
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

  /**
   * Generate jest.config.js content
   * @returns Jest configuration string
   * @tags @CODE:PROCESSOR-JEST-001:API
   */
  public generateJestConfig(): string {
    return `module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  roots: ['<rootDir>/src'],
  testMatch: ['**/__tests__/**/*.test.ts']
};`;
  }

  /**
   * Generate pytest.ini content
   * @returns Pytest configuration string
   * @tags @CODE:PROCESSOR-PYTEST-001:API
   */
  public generatePytestConfig(): string {
    return `[tool:pytest]
testpaths = tests
python_files = test_*.py
python_classes = Test*
python_functions = test_*
`;
  }

  /**
   * Generate MoAI configuration
   * @param data - Template data
   * @returns MoAI config object
   * @tags @CODE:PROCESSOR-MOAI-001:API
   */
  public generateMoaiConfig(data: TemplateData): MoaiConfig {
    return {
      project: {
        name: data.projectName || 'unknown',
        type: data.projectType || ProjectType.TYPESCRIPT,
        version: '0.1.0',
        created_at: data.timestamp || new Date().toISOString(),
      },
      constitution: {
        enforce_tdd: true,
        test_coverage_target: 85,
      },
    };
  }

  /**
   * Generate project documentation file content
   * @param filename - Documentation filename
   * @param data - Template data
   * @returns Documentation content
   * @tags @CODE:PROCESSOR-DOCS-001:API
   */
  public generateProjectFile(filename: string, data: TemplateData): string {
    return `# ${data.projectName} ${filename.replace('.md', '').toUpperCase()}

Generated on ${data.timestamp}
Project Type: ${data.projectType}
Author: ${data.author}

This file was auto-generated by MoAI-ADK.
`;
  }

  /**
   * Generate sync report content
   * @param data - Template data
   * @returns Sync report markdown
   * @tags @CODE:PROCESSOR-SYNC-001:API
   */
  public generateSyncReport(data: TemplateData): string {
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

  /**
   * Generate agent file content
   * @param filename - Agent filename
   * @param data - Template data
   * @returns Agent configuration markdown
   * @tags @CODE:PROCESSOR-AGENT-001:API
   */
  public generateAgentFile(filename: string, data: TemplateData): string {
    const agentName = filename.replace('.md', '');
    return `# ${agentName} Agent

Agent for ${data.projectName}
Generated: ${data.timestamp}

This agent handles ${agentName} functionality.
`;
  }

  /**
   * Generate command file content
   * @param filename - Command filename
   * @param data - Template data
   * @returns Command documentation markdown
   * @tags @CODE:PROCESSOR-COMMAND-001:API
   */
  public generateCommandFile(filename: string, data: TemplateData): string {
    const commandName = filename.replace('.md', '');
    return `# ${commandName} Command

Command for ${data.projectName}
Generated: ${data.timestamp}

Handles ${commandName} workflow step.
`;
  }

  /**
   * Generate pre-commit hook content
   * @param data - Template data
   * @returns Pre-commit hook Python script
   * @tags @CODE:PROCESSOR-HOOK-001:API
   */
  public generatePreCommitHook(data: TemplateData): string {
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

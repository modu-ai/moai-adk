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
 * Aligned with templates/.moai/config.json
 */
interface MoaiConfig {
  _meta: {
    '@CODE:CONFIG-STRUCTURE-001': string;
    '@SPEC:PROJECT-CONFIG-001': string;
  };
  project: {
    name: string;
    version: string;
    mode: 'personal' | 'team';
    description?: string;
    initialized: boolean;
    created_at: string;
    locale: 'ko' | 'en';
  };
  constitution: {
    enforce_tdd: boolean;
    require_tags: boolean;
    test_coverage_target: number;
    simplicity_threshold: number;
    principles: {
      simplicity: {
        max_projects: number;
        notes: string;
      };
    };
  };
  git_strategy: {
    personal: {
      auto_checkpoint: boolean;
      auto_commit: boolean;
      branch_prefix: string;
      checkpoint_interval: number;
      cleanup_days: number;
      max_checkpoints: number;
    };
    team: {
      auto_pr: boolean;
      develop_branch: string;
      draft_pr: boolean;
      feature_prefix: string;
      main_branch: string;
      use_gitflow: boolean;
    };
  };
  tags: {
    auto_sync: boolean;
    storage_type: 'code_scan';
    categories: string[];
    code_scan_policy: {
      no_intermediate_cache: boolean;
      realtime_validation: boolean;
      scan_tools: string[];
      scan_command: string;
      philosophy: string;
    };
  };
  pipeline: {
    available_commands: string[];
    current_stage: string;
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
      projectMode: config.mode || 'personal',
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
      _meta: {
        '@CODE:CONFIG-STRUCTURE-001': '@DOC:JSON-CONFIG-001',
        '@SPEC:PROJECT-CONFIG-001': '@SPEC:MOAI-CONFIG-001',
      },
      project: {
        name: data.projectName || 'unknown',
        version: '0.1.0',
        mode: 'personal',
        description: data.description,
        initialized: true,
        created_at: data.timestamp || new Date().toISOString(),
        locale: 'ko',
      },
      constitution: {
        enforce_tdd: true,
        require_tags: true,
        test_coverage_target: 85,
        simplicity_threshold: 5,
        principles: {
          simplicity: {
            max_projects: 5,
            notes:
              '기본 권장값. 프로젝트 규모에 따라 .moai/config.json 또는 SPEC/ADR로 근거와 함께 조정하세요.',
          },
        },
      },
      git_strategy: {
        personal: {
          auto_checkpoint: true,
          auto_commit: true,
          branch_prefix: 'feature/',
          checkpoint_interval: 300,
          cleanup_days: 7,
          max_checkpoints: 50,
        },
        team: {
          auto_pr: true,
          develop_branch: 'develop',
          draft_pr: true,
          feature_prefix: 'feature/SPEC-',
          main_branch: 'main',
          use_gitflow: true,
        },
      },
      tags: {
        auto_sync: true,
        storage_type: 'code_scan',
        categories: [
          'REQ',
          'DESIGN',
          'TASK',
          'TEST',
          'FEATURE',
          'API',
          'UI',
          'DATA',
        ],
        code_scan_policy: {
          no_intermediate_cache: true,
          realtime_validation: true,
          scan_tools: ['rg', 'grep'],
          scan_command: "rg '@TAG' -n",
          philosophy: 'TAG의 진실은 코드 자체에만 존재',
        },
      },
      pipeline: {
        available_commands: [
          '/alfred:1-spec',
          '/alfred:2-build',
          '/alfred:3-sync',
          '/alfred:4-debug',
        ],
        current_stage: 'initialized',
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
1. Run \`/alfred:1-spec\` to create your first specification
2. Begin TDD development with \`/alfred:2-build\`
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

  // Note: Pre-commit hooks are now implemented in TypeScript
  // See: src/claude/hooks/* for TS-based hook implementations
  // Legacy Python hook generation has been removed
}

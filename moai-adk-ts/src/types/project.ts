// @CODE:PROJ-TYPES-001 |
// Related: @CODE:PROJ-001

/**
 * @file Project type definitions
 * @author MoAI Team
 */

/**
 * Supported project types for initialization
 * @tags @CODE:PROJECT-TYPES-001:DATA
 */
export enum ProjectType {
  PYTHON = 'python',
  NODEJS = 'nodejs',
  MIXED = 'mixed',
  TYPESCRIPT = 'typescript',
  FRONTEND = 'frontend',
}

/**
 * Project configuration interface
 * @tags @CODE:PROJECT-CONFIG-001:DATA
 */
export interface ProjectConfig {
  name: string;
  type: ProjectType;
  mode?: 'personal' | 'team';
  description?: string;
  author?: string;
  license?: string;
  packageManager?: 'npm' | 'yarn' | 'pnpm';
  features?: ProjectFeature[];
}

/**
 * Optional project features
 * @tags @CODE:PROJECT-FEATURES-001:DATA
 */
export interface ProjectFeature {
  name: string;
  enabled: boolean;
  config?: Record<string, any>;
}

/**
 * Project initialization options
 * @tags @CODE:INIT-004 | SPEC: SPEC-INIT-004.md
 * Related: @CODE:INIT-OPTIONS-001:DATA
 */
export interface InitOptions {
  name?: string;
  mode?: 'personal' | 'team';
  path?: string;
  force?: boolean;
  backup?: boolean;
  features?: string[];
  autoGit?: boolean; // Skip Git prompts and auto-initialize
  locale?: 'ko' | 'en'; // Language selection
}

/**
 * Project initialization result
 * @tags @CODE:INIT-RESULT-001:DATA
 */
export interface InitResult {
  success: boolean;
  projectPath: string;
  config: ProjectConfig;
  createdFiles: string[];
  errors?: string[];
  warnings?: string[];
}

/**
 * Template data for project generation
 * @tags @CODE:TEMPLATE-001:DATA
 */
export interface TemplateData {
  projectName: string;
  projectType: ProjectType;
  projectMode?: string;
  timestamp: string;
  author: string;
  description: string;
  license: string;
  packageManager: string;
  features: Record<string, boolean>;
}

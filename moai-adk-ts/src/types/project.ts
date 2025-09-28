/**
 * @file Project type definitions
 * @author MoAI Team
 * @tags @DATA:PROJECT-TYPES-001 @REQ:CLI-INIT-002
 */

/**
 * Supported project types for initialization
 * @tags @DATA:PROJECT-TYPES-001
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
 * @tags @DATA:PROJECT-CONFIG-001
 */
export interface ProjectConfig {
  name: string;
  type: ProjectType;
  description?: string;
  author?: string;
  license?: string;
  packageManager?: 'npm' | 'yarn' | 'pnpm';
  features?: ProjectFeature[];
}

/**
 * Optional project features
 * @tags @DATA:PROJECT-FEATURES-001
 */
export interface ProjectFeature {
  name: string;
  enabled: boolean;
  config?: Record<string, any>;
}

/**
 * Project initialization result
 * @tags @DATA:INIT-RESULT-001
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
 * @tags @DATA:TEMPLATE-001
 */
export interface TemplateData {
  projectName: string;
  projectType: ProjectType;
  timestamp: string;
  author: string;
  description: string;
  license: string;
  packageManager: string;
  features: Record<string, boolean>;
}

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
  timestamp: string;
  author: string;
  description: string;
  license: string;
  packageManager: string;
  features: Record<string, boolean>;
}

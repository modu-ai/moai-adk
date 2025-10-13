// @CODE:REFACTOR-002-VALIDATOR | -VALIDATOR
// Related: @CODE:TEMPLATE-VALIDATOR-001

/**
 * @file Template validation logic
 * @author MoAI Team
 * @tags @CODE:TEMPLATE-VALIDATOR-001
 *
 * Phase 1: Template configuration validation
 * - Project name validation
 * - Config validation
 * - Path validation
 * - Feature validation
 */

import type { ProjectConfig, ProjectFeature } from '@/types/project';
import { ProjectType } from '@/types/project';

/**
 * Validation result interface
 */
export interface ValidationResult {
  isValid: boolean;
  errors: string[];
}

/**
 * Template configuration validator
 * Extracted from template-manager.ts for better separation of concerns
 * @tags @CODE:TEMPLATE-VALIDATOR-001
 */
export class TemplateValidator {
  private readonly PROJECT_NAME_PATTERN = /^[a-zA-Z0-9-_]+$/;
  private readonly UNSAFE_PATHS = [
    '/etc',
    '/root',
    '/sys',
    '/bin',
    '/sbin',
    '/usr/bin',
    '/usr/sbin',
  ];
  private readonly WINDOWS_UNSAFE_PATHS = ['C:\\Windows', 'C:\\Program Files'];

  private readonly VALID_FEATURES: Record<ProjectType, string[]> = {
    [ProjectType.PYTHON]: [
      'pytest',
      'mypy',
      'black',
      'ruff',
      'claude-integration',
    ],
    [ProjectType.NODEJS]: ['jest', 'eslint', 'prettier', 'claude-integration'],
    [ProjectType.TYPESCRIPT]: [
      'typescript',
      'jest',
      'eslint',
      'prettier',
      'biome',
      'claude-integration',
    ],
    [ProjectType.FRONTEND]: [
      'typescript',
      'react',
      'vue',
      'jest',
      'vitest',
      'claude-integration',
    ],
    [ProjectType.MIXED]: [
      'typescript',
      'jest',
      'pytest',
      'mypy',
      'react',
      'claude-integration',
    ],
  };

  /**
   * Validate project name format
   * @param name - Project name to validate
   * @returns Whether the name is valid
   * @tags @CODE:VALIDATOR-NAME-001:API
   */
  public validateProjectName(name: string): boolean {
    if (!name || name.length === 0) {
      return false;
    }

    return this.PROJECT_NAME_PATTERN.test(name);
  }

  /**
   * Get validation errors for a project name
   * @param name - Project name to validate
   * @returns Array of error messages
   * @tags @CODE:VALIDATOR-ERROR-001:API
   */
  public getValidationErrors(name: string): string[] {
    const errors: string[] = [];

    if (!name || name.length === 0) {
      errors.push('Invalid project name: name cannot be empty');
      return errors;
    }

    if (!this.PROJECT_NAME_PATTERN.test(name)) {
      errors.push(
        'Invalid project name format: only alphanumeric characters, hyphens, and underscores are allowed'
      );
    }

    return errors;
  }

  /**
   * Validate complete project configuration
   * @param config - Project configuration to validate
   * @returns Validation result with errors if any
   * @tags @CODE:VALIDATOR-CONFIG-001:API
   */
  public validateConfig(config: ProjectConfig): ValidationResult {
    const errors: string[] = [];

    // Validate required fields
    if (!config.name) {
      errors.push('Invalid config: name is required');
    } else if (!this.validateProjectName(config.name)) {
      errors.push(...this.getValidationErrors(config.name));
    }

    if (!config.type) {
      errors.push('Invalid config: type is required');
    } else if (!Object.values(ProjectType).includes(config.type)) {
      errors.push(
        `Invalid config: type must be one of ${Object.values(ProjectType).join(', ')}`
      );
    }

    // Validate features if present (pass project type for compatibility check)
    if (config.features && config.features.length > 0 && config.type) {
      const featureResult = this.validateFeatures(config.features, config.type);
      errors.push(...featureResult.errors);
    }

    return {
      isValid: errors.length === 0,
      errors,
    };
  }

  /**
   * Validate path safety
   * @param targetPath - Path to validate
   * @returns Whether the path is safe
   * @tags @CODE:VALIDATOR-PATH-001:API
   */
  public validatePath(targetPath: string): boolean {
    if (!targetPath || targetPath.length === 0) {
      return false;
    }

    return !this.isUnsafePath(targetPath);
  }

  /**
   * Check if path is unsafe (system directory)
   * @param targetPath - Path to check
   * @returns Whether the path is unsafe
   */
  private isUnsafePath(targetPath: string): boolean {
    // Check Unix unsafe paths
    if (this.isUnixUnsafePath(targetPath)) {
      return true;
    }

    // Check Windows unsafe paths
    if (process.platform === 'win32' && this.isWindowsUnsafePath(targetPath)) {
      return true;
    }

    return false;
  }

  /**
   * Check Unix unsafe paths
   */
  private isUnixUnsafePath(targetPath: string): boolean {
    return this.UNSAFE_PATHS.some(unsafePath =>
      targetPath.startsWith(unsafePath)
    );
  }

  /**
   * Check Windows unsafe paths
   */
  private isWindowsUnsafePath(targetPath: string): boolean {
    const upperPath = targetPath.toUpperCase();
    return this.WINDOWS_UNSAFE_PATHS.some(unsafePath =>
      upperPath.startsWith(unsafePath.toUpperCase())
    );
  }

  /**
   * Validate features configuration
   * @param features - Features to validate
   * @param projectType - Optional project type for compatibility check
   * @returns Validation result
   * @tags @CODE:VALIDATOR-FEATURE-001:API
   */
  public validateFeatures(
    features: ProjectFeature[],
    projectType?: ProjectType
  ): ValidationResult {
    const errors: string[] = [];

    for (const feature of features) {
      const featureErrors = this.validateSingleFeature(feature, projectType);
      errors.push(...featureErrors);
    }

    return {
      isValid: errors.length === 0,
      errors,
    };
  }

  /**
   * Validate a single feature
   * @param feature - Feature to validate
   * @param projectType - Optional project type for compatibility
   * @returns Array of error messages
   */
  private validateSingleFeature(
    feature: ProjectFeature,
    projectType?: ProjectType
  ): string[] {
    const errors: string[] = [];

    // Validate feature name
    if (!feature.name || feature.name.length === 0) {
      errors.push('Invalid feature: name cannot be empty');
      return errors;
    }

    // Check compatibility if project type is provided
    if (projectType && projectType !== ProjectType.MIXED) {
      const compatibilityError = this.checkFeatureCompatibility(
        feature.name,
        projectType
      );
      if (compatibilityError) {
        errors.push(compatibilityError);
      }
    }

    return errors;
  }

  /**
   * Check feature compatibility with project type
   */
  private checkFeatureCompatibility(
    featureName: string,
    projectType: ProjectType
  ): string | null {
    const validFeatures = this.VALID_FEATURES[projectType] || [];

    if (!validFeatures.includes(featureName)) {
      return `Invalid feature: '${featureName}' is incompatible with project type '${projectType}'`;
    }

    return null;
  }
}

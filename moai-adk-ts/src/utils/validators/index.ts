// @CODE:VALIDATOR-001 | Chain: @REQ:QUAL-001 -> @DESIGN:QUAL-001 -> @TASK:QUAL-001 -> @TEST:UTIL-006
// Related: @CODE:VALID-002:API, @SECURITY:INPUT-VALIDATION-001

/**
 * @file Input validation facade (refactored from monolithic input-validator.ts)
 * @author MoAI Team
 */

export type { PathValidationOptions } from './path-validator';

// Export validation options types
export type { ProjectNameOptions } from './project-name-validator';
// Export types
export type { ValidationResult } from './types';

// Import individual validators
import { validateBranchName as validateBranchNameImpl } from './branch-validator';
import { validateCommandOptions as validateCommandOptionsImpl } from './command-options-validator';
import type { PathValidationOptions } from './path-validator';
import { validatePath as validatePathImpl } from './path-validator';
import {
  type ProjectNameOptions,
  validateProjectName as validateProjectNameImpl,
} from './project-name-validator';
import { validateTemplateType as validateTemplateTypeImpl } from './template-validator';
import type { ValidationResult } from './types';

/**
 * Input validator class for CLI parameters and user inputs
 * @tags @SECURITY:INPUT-VALIDATION-001
 */
export class InputValidator {
  /**
   * Validate project name
   * @tags @CODE:VALIDATE-PROJECT-NAME-001:API
   */
  static validateProjectName(
    projectName: string,
    options?: ProjectNameOptions
  ): ValidationResult {
    return validateProjectNameImpl(projectName, options);
  }

  /**
   * Validate file/directory path
   * @tags @CODE:VALIDATE-PATH-001:API
   */
  static async validatePath(
    inputPath: string,
    options?: PathValidationOptions
  ): Promise<ValidationResult> {
    return validatePathImpl(inputPath, options);
  }

  /**
   * Validate template type
   * @tags @CODE:VALIDATE-TEMPLATE-TYPE-001:API
   */
  static validateTemplateType(templateType: string): ValidationResult {
    return validateTemplateTypeImpl(templateType);
  }

  /**
   * Validate Git branch name
   * @tags @CODE:VALIDATE-BRANCH-NAME-001:API
   */
  static validateBranchName(branchName: string): ValidationResult {
    return validateBranchNameImpl(branchName);
  }

  /**
   * Validate command options
   * @tags @CODE:VALIDATE-COMMAND-OPTIONS-001:API
   */
  static validateCommandOptions(
    options: Record<string, any>
  ): ValidationResult {
    return validateCommandOptionsImpl(options);
  }
}

/**
 * Helper function for quick project name validation
 * @tags @CODE:QUICK-VALIDATE-PROJECT-NAME-001:API
 */
export function validateProjectName(name: string): ValidationResult {
  return InputValidator.validateProjectName(name);
}

/**
 * Helper function for quick path validation
 * @tags @CODE:QUICK-VALIDATE-PATH-001:API
 */
export async function validatePath(
  inputPath: string
): Promise<ValidationResult> {
  return InputValidator.validatePath(inputPath);
}

/**
 * Helper function for quick branch name validation
 * @tags @CODE:QUICK-VALIDATE-BRANCH-NAME-001:API
 */
export function validateBranchName(branchName: string): ValidationResult {
  return InputValidator.validateBranchName(branchName);
}

// @CODE:VALIDATOR-004:BRANCH | Chain: @SPEC:QUAL-001 -> @SPEC:QUAL-001 -> @CODE:QUAL-001 -> @TEST:UTIL-006
// Related: @CODE:INPUT-VALIDATION-001

/**
 * @file Git branch name validation module
 * @author MoAI Team
 */

import type { ValidationResult } from './types';
import {
  BRANCH_NAME_CONSTANTS,
  INVALID_BRANCH_PATTERNS,
  RESERVED_BRANCH_NAMES,
} from './validation-rules';

/**
 * Validate Git branch name
 * @tags @CODE:VALIDATE-BRANCH-NAME-001:API
 */
export function validateBranchName(branchName: string): ValidationResult {
  const errors: string[] = [];

  if (!branchName || typeof branchName !== 'string') {
    errors.push('Branch name must be a non-empty string');
    return { isValid: false, errors };
  }

  const trimmed = branchName.trim();

  // Length validation
  if (
    trimmed.length < BRANCH_NAME_CONSTANTS.MIN_LENGTH ||
    trimmed.length > BRANCH_NAME_CONSTANTS.MAX_LENGTH
  ) {
    errors.push(
      `Branch name must be between ${BRANCH_NAME_CONSTANTS.MIN_LENGTH} and ${BRANCH_NAME_CONSTANTS.MAX_LENGTH} characters`
    );
  }

  // Git branch naming rules
  for (const pattern of INVALID_BRANCH_PATTERNS) {
    if (pattern.test(trimmed)) {
      errors.push('Branch name contains invalid characters or patterns');
      break;
    }
  }

  // Git reserved names
  if (RESERVED_BRANCH_NAMES.includes(trimmed as any)) {
    errors.push('Branch name is reserved');
  }

  return {
    isValid: errors.length === 0,
    errors: Object.freeze(errors),
    sanitizedValue: trimmed,
  };
}

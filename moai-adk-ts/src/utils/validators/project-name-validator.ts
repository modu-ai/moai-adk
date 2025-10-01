// @CODE:VALIDATOR-002:PROJECT-NAME | Chain: @SPEC:QUAL-001 -> @SPEC:QUAL-001 -> @CODE:QUAL-001 -> @TEST:UTIL-006
// Related: @CODE:INPUT-VALIDATION-001

/**
 * @file Project name validation module
 * @author MoAI Team
 */

import type { ValidationResult } from './types';
import {
  DANGEROUS_PROJECT_PATTERNS,
  PROJECT_NAME_CONSTANTS,
  SPECIAL_CHARS_PATTERN,
} from './validation-rules';

/**
 * Project name validation options
 */
export interface ProjectNameOptions {
  readonly minLength?: number;
  readonly maxLength?: number;
  readonly allowSpaces?: boolean;
  readonly allowSpecialChars?: boolean;
}

/**
 * Validate project name
 * @tags @CODE:VALIDATE-PROJECT-NAME-001:API
 */
export function validateProjectName(
  projectName: string,
  options: ProjectNameOptions = {}
): ValidationResult {
  const {
    minLength = PROJECT_NAME_CONSTANTS.DEFAULT_MIN_LENGTH,
    maxLength = PROJECT_NAME_CONSTANTS.DEFAULT_MAX_LENGTH,
    allowSpaces = false,
    allowSpecialChars = false,
  } = options;

  const errors: string[] = [];

  // Basic checks
  if (!projectName || typeof projectName !== 'string') {
    errors.push('Project name must be a non-empty string');
    return { isValid: false, errors };
  }

  const trimmed = projectName.trim();

  // Length validation
  if (trimmed.length < minLength) {
    errors.push(`Project name must be at least ${minLength} characters long`);
  }

  if (trimmed.length > maxLength) {
    errors.push(`Project name must not exceed ${maxLength} characters`);
  }

  // Character validation
  if (!allowSpaces && /\s/.test(trimmed)) {
    errors.push('Project name cannot contain spaces');
  }

  // Special character validation
  if (!allowSpecialChars && SPECIAL_CHARS_PATTERN.test(trimmed)) {
    errors.push('Project name cannot contain special characters');
  }

  // Dangerous pattern detection
  for (const pattern of DANGEROUS_PROJECT_PATTERNS) {
    if (pattern.test(trimmed)) {
      errors.push('Project name contains invalid characters or patterns');
      break;
    }
  }

  // Sanitize the value
  let sanitized = trimmed;
  if (!allowSpecialChars) {
    sanitized = sanitized.replace(/[^a-zA-Z0-9_-]/g, '-');
  }

  return {
    isValid: errors.length === 0,
    errors: Object.freeze(errors),
    sanitizedValue: sanitized,
  };
}

// @CODE:VALIDATOR-007:COMMAND-OPTIONS | Chain: @REQ:QUAL-001 -> @DESIGN:QUAL-001 -> @TASK:QUAL-001 -> @TEST:UTIL-006
// Related: @SECURITY:INPUT-VALIDATION-001

/**
 * @file Command options validation module
 * @author MoAI Team
 */

import type { ValidationResult } from './types';
import {
  COMMAND_OPTIONS_CONSTANTS,
  DANGEROUS_STRING_PATTERNS,
  VALID_OPTION_KEY_PATTERN,
} from './validation-rules';

/**
 * Validate command options
 * @tags @CODE:VALIDATE-COMMAND-OPTIONS-001:API
 */
export function validateCommandOptions(
  options: Record<string, any>
): ValidationResult {
  const errors: string[] = [];
  const sanitizedOptions: Record<string, any> = {};

  for (const [key, value] of Object.entries(options)) {
    // Validate option key
    if (!VALID_OPTION_KEY_PATTERN.test(key)) {
      errors.push(`Invalid option key: ${key}`);
      continue;
    }

    // Validate option value based on type
    if (typeof value === 'string') {
      if (value.length > COMMAND_OPTIONS_CONSTANTS.MAX_STRING_LENGTH) {
        errors.push(`Option ${key} value is too long`);
        continue;
      }

      if (containsDangerousString(value)) {
        errors.push(`Option ${key} contains dangerous content`);
        continue;
      }

      sanitizedOptions[key] = value.trim();
    } else if (typeof value === 'boolean') {
      sanitizedOptions[key] = value;
    } else if (typeof value === 'number') {
      if (!Number.isFinite(value)) {
        errors.push(`Option ${key} must be a finite number`);
        continue;
      }
      sanitizedOptions[key] = value;
    } else {
      errors.push(`Option ${key} has unsupported type`);
    }
  }

  return {
    isValid: errors.length === 0,
    errors: Object.freeze(errors),
    sanitizedValue: sanitizedOptions,
  };
}

/**
 * Check for dangerous string content
 */
function containsDangerousString(value: string): boolean {
  return DANGEROUS_STRING_PATTERNS.some(pattern => pattern.test(value));
}

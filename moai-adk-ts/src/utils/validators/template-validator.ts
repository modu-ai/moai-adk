// @CODE:VALIDATOR-006:TEMPLATE | Chain: @REQ:QUAL-001 -> @DESIGN:QUAL-001 -> @TASK:QUAL-001 -> @TEST:UTIL-006
// Related: @SECURITY:INPUT-VALIDATION-001

/**
 * @file Template type validation module
 * @author MoAI Team
 */

import type { ValidationResult } from './types';
import { ALLOWED_TEMPLATE_TYPES } from './validation-rules';

/**
 * Validate template type
 * @tags @CODE:VALIDATE-TEMPLATE-TYPE-001:API
 */
export function validateTemplateType(templateType: string): ValidationResult {
  const errors: string[] = [];

  if (!templateType || typeof templateType !== 'string') {
    errors.push('Template type must be a non-empty string');
    return { isValid: false, errors };
  }

  const trimmed = templateType.trim().toLowerCase();

  if (!ALLOWED_TEMPLATE_TYPES.includes(trimmed as any)) {
    errors.push(
      `Invalid template type. Allowed: ${ALLOWED_TEMPLATE_TYPES.join(', ')}`
    );
  }

  return {
    isValid: errors.length === 0,
    errors: Object.freeze(errors),
    sanitizedValue: trimmed,
  };
}

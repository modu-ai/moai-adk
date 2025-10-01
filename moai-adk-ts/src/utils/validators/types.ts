// @CODE:VALIDATOR-008:TYPES | Chain: @SPEC:QUAL-001 -> @SPEC:QUAL-001 -> @CODE:QUAL-001 -> @TEST:UTIL-006
// Related: @CODE:VALIDATION-RESULT-001:DATA

/**
 * @file Common validation types
 * @author MoAI Team
 */

/**
 * Validation result interface
 * @tags @CODE:VALIDATION-RESULT-001:DATA
 */
export interface ValidationResult {
  readonly isValid: boolean;
  readonly errors: readonly string[];
  readonly sanitizedValue?: any;
}

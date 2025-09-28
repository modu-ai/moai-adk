/**
 * @FEATURE:TEMPLATE-VALIDATION-001 Template Context Validation
 *
 * 템플릿 컨텍스트 검증 기능
 * @DESIGN:SEPARATE-CONCERNS-001 TemplateProcessor에서 분리하여 단일 책임 원칙 준수
 */

import type {
  TemplateContext,
  ContextSchema,
  ValidationResult,
} from './template-processor';

/**
 * @API:VALIDATE-CONTEXT-001 컨텍스트 데이터 검증
 *
 * @param context 검증할 컨텍스트
 * @param schema 검증 스키마
 * @returns 검증 결과
 */
export function validateTemplateContext(
  context: TemplateContext,
  schema: ContextSchema
): ValidationResult {
  const errors: string[] = [];
  const warnings: string[] = [];

  try {
    // 필수 필드 검증
    for (const requiredField of schema.required) {
      if (!(requiredField in context)) {
        errors.push(`Required field ${requiredField} is missing`);
      }
    }

    // 유효성 검사 실행
    for (const [field, validator] of Object.entries(schema.validation)) {
      if (field in context) {
        try {
          if (!validator(context[field])) {
            errors.push(`Validation failed for field ${field}`);
          }
        } catch (validationError) {
          errors.push(
            `Validation error for field ${field}: ${validationError}`
          );
        }
      }
    }

    // 경고: 알 수 없는 필드
    const knownFields = new Set([...schema.required, ...schema.optional]);
    for (const field of Object.keys(context)) {
      if (!knownFields.has(field)) {
        warnings.push(`Unknown field ${field} in context`);
      }
    }

    return {
      valid: errors.length === 0,
      errors: Object.freeze(errors),
      warnings: Object.freeze(warnings),
    };
  } catch (error) {
    return {
      valid: false,
      errors: Object.freeze([`Validation error: ${error}`]),
      warnings: Object.freeze(warnings),
    };
  }
}

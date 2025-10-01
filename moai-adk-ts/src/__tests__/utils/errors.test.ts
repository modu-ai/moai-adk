// @TEST:UTIL-006 |
// Related: @CODE:TYPE-SAFETY-001, @CODE:VALID-002:API

/**
 * @file errors.ts test suite
 * @author MoAI Team
 * @tags @TEST:UTIL-006 @SPEC:QUAL-006
 */

import { describe, expect, test } from 'vitest';
import {
  ValidationError,
  InstallationError,
  TemplateError,
  ResourceError,
  PhaseError,
  isValidationError,
  isInstallationError,
  isTemplateError,
  isResourceError,
  isPhaseError,
  toError,
  getErrorMessage,
} from '@/utils/errors';

describe('Custom Error Classes', () => {
  describe('ValidationError', () => {
    test('should create ValidationError with all properties', () => {
      // Given: ValidationError 생성 옵션
      const message = 'Invalid input detected';
      const options = {
        pattern: '../traversal',
        vulnerabilities: ['path-traversal', 'injection'],
        context: { input: 'test', userId: 123 },
      };

      // When: ValidationError 생성
      const error = new ValidationError(message, options);

      // Then: 모든 속성이 올바르게 설정됨
      expect(error.name).toBe('ValidationError');
      expect(error.message).toBe(message);
      expect(error.pattern).toBe(options.pattern);
      expect(error.vulnerabilities).toEqual(options.vulnerabilities);
      expect(error.context).toEqual(options.context);
      expect(error instanceof ValidationError).toBe(true);
      expect(error instanceof Error).toBe(true);
    });

    test('should create ValidationError without options', () => {
      // Given: 옵션 없이 ValidationError 생성
      const message = 'Basic validation error';

      // When: ValidationError 생성
      const error = new ValidationError(message);

      // Then: 기본 속성만 설정됨
      expect(error.name).toBe('ValidationError');
      expect(error.message).toBe(message);
      expect(error.pattern).toBeUndefined();
      expect(error.vulnerabilities).toBeUndefined();
      expect(error.context).toBeUndefined();
    });

    test('should preserve stack trace', () => {
      // Given/When: ValidationError 생성
      const error = new ValidationError('Test error');

      // Then: 스택 트레이스가 존재함
      expect(error.stack).toBeDefined();
      expect(error.stack).toContain('ValidationError');
    });
  });

  describe('InstallationError', () => {
    test('should create InstallationError with all properties', () => {
      // Given: InstallationError 생성 옵션
      const message = 'Installation failed';
      const originalError = new Error('ENOENT: file not found');
      const options = {
        error: originalError,
        errorMessage: 'File not found',
        projectPath: '/test/project',
        phase: 'dependency-install',
        context: { step: 1 },
      };

      // When: InstallationError 생성
      const error = new InstallationError(message, options);

      // Then: 모든 속성이 올바르게 설정됨
      expect(error.name).toBe('InstallationError');
      expect(error.message).toBe(message);
      expect(error.error).toBe(originalError);
      expect(error.errorMessage).toBe(options.errorMessage);
      expect(error.projectPath).toBe(options.projectPath);
      expect(error.phase).toBe(options.phase);
      expect(error.context).toEqual(options.context);
      expect(error instanceof InstallationError).toBe(true);
      expect(error instanceof Error).toBe(true);
    });

    test('should create InstallationError without options', () => {
      // Given: 옵션 없이 InstallationError 생성
      const message = 'Installation failed';

      // When: InstallationError 생성
      const error = new InstallationError(message);

      // Then: 기본 속성만 설정됨
      expect(error.name).toBe('InstallationError');
      expect(error.message).toBe(message);
      expect(error.error).toBeUndefined();
      expect(error.projectPath).toBeUndefined();
      expect(error.phase).toBeUndefined();
    });
  });

  describe('TemplateError', () => {
    test('should create TemplateError with all properties', () => {
      // Given: TemplateError 생성 옵션
      const message = 'Template processing failed';
      const originalError = new Error('Parse error');
      const options = {
        error: originalError,
        templatePath: '/templates/standard',
        context: { templateType: 'standard' },
      };

      // When: TemplateError 생성
      const error = new TemplateError(message, options);

      // Then: 모든 속성이 올바르게 설정됨
      expect(error.name).toBe('TemplateError');
      expect(error.message).toBe(message);
      expect(error.error).toBe(originalError);
      expect(error.templatePath).toBe(options.templatePath);
      expect(error.context).toEqual(options.context);
      expect(error instanceof TemplateError).toBe(true);
      expect(error instanceof Error).toBe(true);
    });

    test('should create TemplateError without options', () => {
      // Given: 옵션 없이 TemplateError 생성
      const message = 'Template error';

      // When: TemplateError 생성
      const error = new TemplateError(message);

      // Then: 기본 속성만 설정됨
      expect(error.name).toBe('TemplateError');
      expect(error.message).toBe(message);
      expect(error.templatePath).toBeUndefined();
    });
  });

  describe('ResourceError', () => {
    test('should create ResourceError with all properties', () => {
      // Given: ResourceError 생성 옵션
      const message = 'Resource not found';
      const originalError = new Error('ENOENT');
      const options = {
        error: originalError,
        resourcePath: '/path/to/resource',
        context: { resourceType: 'template' },
      };

      // When: ResourceError 생성
      const error = new ResourceError(message, options);

      // Then: 모든 속성이 올바르게 설정됨
      expect(error.name).toBe('ResourceError');
      expect(error.message).toBe(message);
      expect(error.error).toBe(originalError);
      expect(error.resourcePath).toBe(options.resourcePath);
      expect(error.context).toEqual(options.context);
      expect(error instanceof ResourceError).toBe(true);
      expect(error instanceof Error).toBe(true);
    });

    test('should create ResourceError without options', () => {
      // Given: 옵션 없이 ResourceError 생성
      const message = 'Resource error';

      // When: ResourceError 생성
      const error = new ResourceError(message);

      // Then: 기본 속성만 설정됨
      expect(error.name).toBe('ResourceError');
      expect(error.message).toBe(message);
      expect(error.resourcePath).toBeUndefined();
    });
  });

  describe('PhaseError', () => {
    test('should create PhaseError with all properties', () => {
      // Given: PhaseError 생성 옵션
      const message = 'Phase execution failed';
      const originalError = new Error('Task failed');
      const options = {
        error: originalError,
        phase: 'build',
        context: { step: 'compilation' },
      };

      // When: PhaseError 생성
      const error = new PhaseError(message, options);

      // Then: 모든 속성이 올바르게 설정됨
      expect(error.name).toBe('PhaseError');
      expect(error.message).toBe(message);
      expect(error.error).toBe(originalError);
      expect(error.phase).toBe(options.phase);
      expect(error.context).toEqual(options.context);
      expect(error instanceof PhaseError).toBe(true);
      expect(error instanceof Error).toBe(true);
    });

    test('should create PhaseError without options', () => {
      // Given: 옵션 없이 PhaseError 생성
      const message = 'Phase error';

      // When: PhaseError 생성
      const error = new PhaseError(message);

      // Then: 기본 속성만 설정됨
      expect(error.name).toBe('PhaseError');
      expect(error.message).toBe(message);
      expect(error.phase).toBeUndefined();
    });
  });
});

describe('Type Guard Functions', () => {
  describe('isValidationError', () => {
    test('should return true for ValidationError', () => {
      // Given: ValidationError 인스턴스
      const error = new ValidationError('test');

      // When/Then: ValidationError로 식별됨
      expect(isValidationError(error)).toBe(true);
    });

    test('should return false for other custom errors', () => {
      // Given: 다른 커스텀 에러들
      const installationError = new InstallationError('test');
      const templateError = new TemplateError('test');
      const resourceError = new ResourceError('test');
      const phaseError = new PhaseError('test');

      // When/Then: ValidationError가 아님
      expect(isValidationError(installationError)).toBe(false);
      expect(isValidationError(templateError)).toBe(false);
      expect(isValidationError(resourceError)).toBe(false);
      expect(isValidationError(phaseError)).toBe(false);
    });

    test('should return false for generic Error', () => {
      // Given: 일반 Error 객체
      const error = new Error('test');

      // When/Then: ValidationError가 아님
      expect(isValidationError(error)).toBe(false);
    });

    test('should return false for null and undefined', () => {
      // When/Then: null과 undefined는 false
      expect(isValidationError(null)).toBe(false);
      expect(isValidationError(undefined)).toBe(false);
    });

    test('should return false for non-error values', () => {
      // When/Then: 에러가 아닌 값은 false
      expect(isValidationError('error string')).toBe(false);
      expect(isValidationError(123)).toBe(false);
      expect(isValidationError({ message: 'test' })).toBe(false);
    });
  });

  describe('isInstallationError', () => {
    test('should return true for InstallationError', () => {
      // Given: InstallationError 인스턴스
      const error = new InstallationError('test');

      // When/Then: InstallationError로 식별됨
      expect(isInstallationError(error)).toBe(true);
    });

    test('should return false for other errors', () => {
      // Given: 다른 에러들
      const validationError = new ValidationError('test');
      const genericError = new Error('test');

      // When/Then: InstallationError가 아님
      expect(isInstallationError(validationError)).toBe(false);
      expect(isInstallationError(genericError)).toBe(false);
      expect(isInstallationError(null)).toBe(false);
    });
  });

  describe('isTemplateError', () => {
    test('should return true for TemplateError', () => {
      // Given: TemplateError 인스턴스
      const error = new TemplateError('test');

      // When/Then: TemplateError로 식별됨
      expect(isTemplateError(error)).toBe(true);
    });

    test('should return false for other errors', () => {
      // Given: 다른 에러들
      const validationError = new ValidationError('test');
      const genericError = new Error('test');

      // When/Then: TemplateError가 아님
      expect(isTemplateError(validationError)).toBe(false);
      expect(isTemplateError(genericError)).toBe(false);
      expect(isTemplateError(undefined)).toBe(false);
    });
  });

  describe('isResourceError', () => {
    test('should return true for ResourceError', () => {
      // Given: ResourceError 인스턴스
      const error = new ResourceError('test');

      // When/Then: ResourceError로 식별됨
      expect(isResourceError(error)).toBe(true);
    });

    test('should return false for other errors', () => {
      // Given: 다른 에러들
      const templateError = new TemplateError('test');
      const genericError = new Error('test');

      // When/Then: ResourceError가 아님
      expect(isResourceError(templateError)).toBe(false);
      expect(isResourceError(genericError)).toBe(false);
      expect(isResourceError(null)).toBe(false);
    });
  });

  describe('isPhaseError', () => {
    test('should return true for PhaseError', () => {
      // Given: PhaseError 인스턴스
      const error = new PhaseError('test');

      // When/Then: PhaseError로 식별됨
      expect(isPhaseError(error)).toBe(true);
    });

    test('should return false for other errors', () => {
      // Given: 다른 에러들
      const resourceError = new ResourceError('test');
      const genericError = new Error('test');

      // When/Then: PhaseError가 아님
      expect(isPhaseError(resourceError)).toBe(false);
      expect(isPhaseError(genericError)).toBe(false);
      expect(isPhaseError(undefined)).toBe(false);
    });
  });
});

describe('Helper Functions', () => {
  describe('toError', () => {
    test('should return Error object as-is', () => {
      // Given: Error 객체
      const originalError = new Error('test error');

      // When: toError 호출
      const result = toError(originalError);

      // Then: 동일한 Error 객체 반환
      expect(result).toBe(originalError);
      expect(result.message).toBe('test error');
    });

    test('should convert string to Error', () => {
      // Given: 문자열
      const errorString = 'Something went wrong';

      // When: toError 호출
      const result = toError(errorString);

      // Then: Error 객체로 변환됨
      expect(result instanceof Error).toBe(true);
      expect(result.message).toBe(errorString);
    });

    test('should convert number to Error', () => {
      // Given: 숫자
      const errorNumber = 404;

      // When: toError 호출
      const result = toError(errorNumber);

      // Then: Error 객체로 변환됨
      expect(result instanceof Error).toBe(true);
      expect(result.message).toBe('404');
    });

    test('should convert null to Error', () => {
      // Given: null
      const errorValue = null;

      // When: toError 호출
      const result = toError(errorValue);

      // Then: Error 객체로 변환됨
      expect(result instanceof Error).toBe(true);
      expect(result.message).toBe('null');
    });

    test('should convert undefined to Error', () => {
      // Given: undefined
      const errorValue = undefined;

      // When: toError 호출
      const result = toError(errorValue);

      // Then: Error 객체로 변환됨
      expect(result instanceof Error).toBe(true);
      expect(result.message).toBe('undefined');
    });

    test('should convert object to Error', () => {
      // Given: 객체
      const errorObject = { code: 'ERR_001', details: 'Failed' };

      // When: toError 호출
      const result = toError(errorObject);

      // Then: Error 객체로 변환됨
      expect(result instanceof Error).toBe(true);
      expect(result.message).toBe('[object Object]');
    });

    test('should preserve custom Error subclass', () => {
      // Given: 커스텀 Error 서브클래스
      const customError = new ValidationError('validation failed');

      // When: toError 호출
      const result = toError(customError);

      // Then: 원본 객체 그대로 반환
      expect(result).toBe(customError);
      expect(result instanceof ValidationError).toBe(true);
    });
  });

  describe('getErrorMessage', () => {
    test('should extract message from Error object', () => {
      // Given: Error 객체
      const error = new Error('test error message');

      // When: getErrorMessage 호출
      const result = getErrorMessage(error);

      // Then: 에러 메시지 반환
      expect(result).toBe('test error message');
    });

    test('should extract message from custom Error', () => {
      // Given: 커스텀 Error 객체
      const error = new ValidationError('validation failed', {
        pattern: '../test',
      });

      // When: getErrorMessage 호출
      const result = getErrorMessage(error);

      // Then: 에러 메시지 반환
      expect(result).toBe('validation failed');
    });

    test('should return string as-is', () => {
      // Given: 문자열
      const errorString = 'error message';

      // When: getErrorMessage 호출
      const result = getErrorMessage(errorString);

      // Then: 동일한 문자열 반환
      expect(result).toBe(errorString);
    });

    test('should convert number to string', () => {
      // Given: 숫자
      const errorNumber = 500;

      // When: getErrorMessage 호출
      const result = getErrorMessage(errorNumber);

      // Then: 문자열로 변환됨
      expect(result).toBe('500');
    });

    test('should convert null to string', () => {
      // Given: null
      const errorValue = null;

      // When: getErrorMessage 호출
      const result = getErrorMessage(errorValue);

      // Then: 문자열로 변환됨
      expect(result).toBe('null');
    });

    test('should convert undefined to string', () => {
      // Given: undefined
      const errorValue = undefined;

      // When: getErrorMessage 호출
      const result = getErrorMessage(errorValue);

      // Then: 문자열로 변환됨
      expect(result).toBe('undefined');
    });

    test('should convert object to string', () => {
      // Given: 객체
      const errorObject = { message: 'test', code: 123 };

      // When: getErrorMessage 호출
      const result = getErrorMessage(errorObject);

      // Then: 문자열로 변환됨
      expect(result).toContain('object');
    });
  });
});

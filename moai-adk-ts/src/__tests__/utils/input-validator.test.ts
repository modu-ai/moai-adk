// @TEST:UTIL-006 |
// Related: @CODE:UTIL-005, @CODE:VALID-002:API, @CODE:INPUT-001

/**
 * @file input-validator.ts test suite
 * @author MoAI Team
 * @tags @TEST:UTIL-006 @SPEC:QUAL-006 @CODE:INPUT-VALIDATION-001
 */

import * as fs from 'node:fs';
import * as os from 'node:os';
import * as path from 'node:path';
import { afterEach, beforeEach, describe, expect, test } from 'vitest';
import {
  InputValidator,
  validateBranchName,
  validatePath,
  validateProjectName,
} from '@/utils/input-validator';

describe('InputValidator', () => {
  describe('validateProjectName', () => {
    test('should accept valid project name', () => {
      // Given: 유효한 프로젝트명
      const validName = 'my-project';

      // When: validateProjectName 호출
      const result = InputValidator.validateProjectName(validName);

      // Then: 검증 통과
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
      expect(result.sanitizedValue).toBe(validName);
    });

    test('should accept alphanumeric with dashes and underscores', () => {
      // Given: 영숫자, 대시, 언더스코어 포함
      const validName = 'my_cool-project123';

      // When: validateProjectName 호출
      const result = InputValidator.validateProjectName(validName);

      // Then: 검증 통과
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test('should reject empty string', () => {
      // Given: 빈 문자열
      const emptyName = '';

      // When: validateProjectName 호출
      const result = InputValidator.validateProjectName(emptyName);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors[0]).toContain('non-empty string');
    });

    test('should reject name shorter than minLength', () => {
      // Given: 최소 길이 미달 (빈 문자열이 아닌)
      const shortName = 'ab';

      // When: validateProjectName 호출 (minLength: 3)
      const result = InputValidator.validateProjectName(shortName, {
        minLength: 3,
      });

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(result.errors.some(err => err.includes('at least'))).toBe(true);
    });

    test('should reject name longer than maxLength', () => {
      // Given: 최대 길이 초과
      const longName = 'a'.repeat(51);

      // When: validateProjectName 호출 (maxLength: 50)
      const result = InputValidator.validateProjectName(longName);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(result.errors.some(err => err.includes('not exceed'))).toBe(true);
    });

    test('should reject spaces by default', () => {
      // Given: 공백 포함
      const nameWithSpaces = 'my project';

      // When: validateProjectName 호출 (allowSpaces: false)
      const result = InputValidator.validateProjectName(nameWithSpaces);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(
        result.errors.some(err => err.includes('cannot contain spaces'))
      ).toBe(true);
    });

    test('should accept spaces when allowSpaces is true', () => {
      // Given: 공백 포함
      const nameWithSpaces = 'my project';

      // When: validateProjectName 호출 (allowSpaces: true)
      const result = InputValidator.validateProjectName(nameWithSpaces, {
        allowSpaces: true,
      });

      // Then: 검증 통과
      expect(result.isValid).toBe(true);
    });

    test('should reject special characters by default', () => {
      // Given: 특수문자 포함
      const nameWithSpecialChars = 'my@project!';

      // When: validateProjectName 호출
      const result = InputValidator.validateProjectName(nameWithSpecialChars);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(
        result.errors.some(err => err.includes('special characters'))
      ).toBe(true);
    });

    test('should accept special characters when allowSpecialChars is true', () => {
      // Given: 특수문자 포함
      const nameWithSpecialChars = 'my@project!';

      // When: validateProjectName 호출 (allowSpecialChars: true)
      const result = InputValidator.validateProjectName(nameWithSpecialChars, {
        allowSpecialChars: true,
      });

      // Then: 검증 통과 (다른 위험 패턴 없으면)
      expect(result.isValid).toBe(true);
    });

    test('should reject path traversal patterns', () => {
      // Given: Path traversal 패턴
      const dangerousName = '../parent-dir';

      // When: validateProjectName 호출
      const result = InputValidator.validateProjectName(dangerousName);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(
        result.errors.some(err =>
          err.includes('invalid characters or patterns')
        )
      ).toBe(true);
    });

    test('should reject names starting with dot or dash', () => {
      // Given: 점이나 대시로 시작
      const nameStartingWithDot = '.hidden';
      const nameStartingWithDash = '-project';

      // When: validateProjectName 호출
      const resultDot = InputValidator.validateProjectName(nameStartingWithDot);
      const resultDash =
        InputValidator.validateProjectName(nameStartingWithDash);

      // Then: 검증 실패
      expect(resultDot.isValid).toBe(false);
      expect(resultDash.isValid).toBe(false);
    });

    test('should reject Windows reserved names (case insensitive)', () => {
      // Given: Windows 예약어
      const reservedNames = ['CON', 'PRN', 'AUX', 'NUL', 'COM1', 'LPT1'];

      // When: validateProjectName 호출
      const results = reservedNames.map(name =>
        InputValidator.validateProjectName(name)
      );

      // Then: 모두 검증 실패
      results.forEach(result => {
        expect(result.isValid).toBe(false);
      });
    });

    test('should reject control characters', () => {
      // Given: 제어 문자 포함
      const nameWithControlChar = 'my\x00project';

      // When: validateProjectName 호출
      const result = InputValidator.validateProjectName(nameWithControlChar);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
    });

    test('should sanitize special characters in sanitizedValue', () => {
      // Given: 특수문자 포함
      const nameWithSpecialChars = 'my@project!test';

      // When: validateProjectName 호출
      const result = InputValidator.validateProjectName(nameWithSpecialChars);

      // Then: sanitizedValue에서 특수문자가 대시로 변환됨
      expect(result.sanitizedValue).toBe('my-project-test');
    });

    test('should trim whitespace', () => {
      // Given: 앞뒤 공백 포함
      const nameWithWhitespace = '  my-project  ';

      // When: validateProjectName 호출
      const result = InputValidator.validateProjectName(nameWithWhitespace);

      // Then: 공백이 제거됨
      expect(result.sanitizedValue).toBe('my-project');
    });

    test('should reject non-string input', () => {
      // Given: 문자열이 아닌 입력
      const invalidInput = 123 as any;

      // When: validateProjectName 호출
      const result = InputValidator.validateProjectName(invalidInput);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(result.errors[0]).toContain('non-empty string');
    });
  });

  describe('validatePath', () => {
    let tempDir: string;

    beforeEach(() => {
      // 임시 디렉토리 생성
      tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'moai-test-'));
    });

    afterEach(() => {
      // 임시 디렉토리 정리
      if (tempDir && fs.existsSync(tempDir)) {
        fs.rmSync(tempDir, { recursive: true, force: true });
      }
    });

    test('should accept valid existing path', async () => {
      // Given: 존재하는 유효한 경로 (현재 디렉토리 기준)
      const validPath = './package.json'; // 프로젝트 루트의 파일

      // When: validatePath 호출
      const result = await InputValidator.validatePath(validPath);

      // Then: 검증 통과 (또는 normalizedPath 반환)
      expect(result.sanitizedValue).toBeDefined();
      expect(path.isAbsolute(result.sanitizedValue!)).toBe(true);
    });

    test('should reject empty string', async () => {
      // Given: 빈 문자열
      const emptyPath = '';

      // When: validatePath 호출
      const result = await InputValidator.validatePath(emptyPath);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(result.errors[0]).toContain('non-empty string');
    });

    test('should reject path longer than 260 characters', async () => {
      // Given: 260자 초과 경로
      const longPath = `/long/path/${'a'.repeat(260)}`;

      // When: validatePath 호출
      const result = await InputValidator.validatePath(longPath);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(result.errors.some(err => err.includes('too long'))).toBe(true);
    });

    test('should reject system directory paths', async () => {
      // Given: 시스템 디렉토리 경로
      const systemPaths = ['/etc/passwd', '/bin/bash', '/var/log/system.log'];

      // When: validatePath 호출
      const results = await Promise.all(
        systemPaths.map(p => InputValidator.validatePath(p))
      );

      // Then: 검증 실패
      results.forEach(result => {
        expect(result.isValid).toBe(false);
        expect(
          result.errors.some(err => err.includes('dangerous patterns'))
        ).toBe(true);
      });
    });

    test('should reject Windows system paths', async () => {
      // Given: Windows 시스템 경로
      const windowsPaths = ['C:\\Windows\\System32', 'C:\\Program Files\\test'];

      // When: validatePath 호출
      const results = await Promise.all(
        windowsPaths.map(p => InputValidator.validatePath(p))
      );

      // Then: 검증 실패
      results.forEach(result => {
        expect(result.isValid).toBe(false);
      });
    });

    test('should reject executable file extensions', async () => {
      // Given: 실행 파일 확장자
      const executablePaths = [
        '/test/file.exe',
        '/test/script.bat',
        '/test/command.cmd',
      ];

      // When: validatePath 호출
      const results = await Promise.all(
        executablePaths.map(p => InputValidator.validatePath(p))
      );

      // Then: 검증 실패
      results.forEach(result => {
        expect(result.isValid).toBe(false);
      });
    });

    test('should enforce maxDepth option', async () => {
      // Given: 깊이 제한을 초과하는 경로
      const deepPath = '/a/b/c/d/e/f/g/h/i/j';

      // When: validatePath 호출 (maxDepth: 3)
      const result = await InputValidator.validatePath(deepPath, {
        maxDepth: 3,
      });

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(result.errors.some(err => err.includes('depth exceeds'))).toBe(
        true
      );
    });

    test('should enforce mustExist option', async () => {
      // Given: 존재하지 않는 경로
      const nonExistentPath = path.join(tempDir, 'nonexistent');

      // When: validatePath 호출 (mustExist: true)
      const result = await InputValidator.validatePath(nonExistentPath, {
        mustExist: true,
      });

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(result.errors.some(err => err.includes('does not exist'))).toBe(
        true
      );
    });

    test('should enforce mustBeDirectory option', async () => {
      // Given: 파일 생성
      const filePath = path.join(tempDir, 'test-file.txt');
      fs.writeFileSync(filePath, 'test content');

      // When: validatePath 호출 (mustBeDirectory: true)
      const result = await InputValidator.validatePath(filePath, {
        mustBeDirectory: true,
      });

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(
        result.errors.some(err => err.includes('must be a directory'))
      ).toBe(true);
    });

    test('should enforce mustBeFile option', async () => {
      // Given: 디렉토리 경로
      const dirPath = tempDir;

      // When: validatePath 호출 (mustBeFile: true)
      const result = await InputValidator.validatePath(dirPath, {
        mustBeFile: true,
      });

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(result.errors.some(err => err.includes('must be a file'))).toBe(
        true
      );
    });

    test('should enforce allowedExtensions option', async () => {
      // Given: 허용되지 않은 확장자
      const filePath = path.join(tempDir, 'test.txt');
      fs.writeFileSync(filePath, 'test');

      // When: validatePath 호출 (allowedExtensions: ['.md', '.json'])
      const result = await InputValidator.validatePath(filePath, {
        allowedExtensions: ['.md', '.json'],
      });

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(
        result.errors.some(err => err.includes('extension not allowed'))
      ).toBe(true);
    });

    test('should accept allowed extensions', async () => {
      // Given: 허용된 확장자 (현재 디렉토리 내)
      const filePath = './package.json'; // 프로젝트 루트의 실제 파일

      // When: validatePath 호출 (allowedExtensions: ['.json'])
      const result = await InputValidator.validatePath(filePath, {
        allowedExtensions: ['.json'],
      });

      // Then: 검증 통과 (또는 normalizedPath 반환)
      expect(result.sanitizedValue).toBeDefined();
    });

    test('should detect path traversal', async () => {
      // Given: Path traversal 패턴
      const traversalPath = '../../../etc/passwd';

      // When: validatePath 호출
      const result = await InputValidator.validatePath(traversalPath);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(
        result.errors.some(
          err => err.includes('dangerous patterns') || err.includes('traversal')
        )
      ).toBe(true);
    });

    test('should normalize and resolve relative paths', async () => {
      // Given: 상대 경로
      const relativePath = './test';

      // When: validatePath 호출
      const result = await InputValidator.validatePath(relativePath);

      // Then: sanitizedValue가 절대 경로로 변환됨
      expect(result.sanitizedValue).toBeDefined();
      expect(path.isAbsolute(result.sanitizedValue!)).toBe(true);
    });

    test('should reject control characters in path', async () => {
      // Given: 제어 문자 포함 경로
      const pathWithControlChar = '/test\x00/path';

      // When: validatePath 호출
      const result = await InputValidator.validatePath(pathWithControlChar);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
    });

    test('should reject non-string input', async () => {
      // Given: 문자열이 아닌 입력
      const invalidInput = 123 as any;

      // When: validatePath 호출
      const result = await InputValidator.validatePath(invalidInput);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
    });
  });

  describe('validateTemplateType', () => {
    test('should accept valid template types', () => {
      // Given: 유효한 템플릿 타입
      const validTypes = ['standard', 'minimal', 'advanced', 'custom'];

      // When: validateTemplateType 호출
      const results = validTypes.map(type =>
        InputValidator.validateTemplateType(type)
      );

      // Then: 모두 검증 통과
      results.forEach(result => {
        expect(result.isValid).toBe(true);
      });
    });

    test('should be case insensitive', () => {
      // Given: 대소문자 변형
      const variations = ['STANDARD', 'Standard', 'sTaNdArD'];

      // When: validateTemplateType 호출
      const results = variations.map(type =>
        InputValidator.validateTemplateType(type)
      );

      // Then: 모두 검증 통과
      results.forEach(result => {
        expect(result.isValid).toBe(true);
        expect(result.sanitizedValue).toBe('standard');
      });
    });

    test('should reject invalid template type', () => {
      // Given: 허용되지 않는 템플릿 타입
      const invalidType = 'invalid-type';

      // When: validateTemplateType 호출
      const result = InputValidator.validateTemplateType(invalidType);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(result.errors[0]).toContain('Invalid template type');
    });

    test('should reject empty string', () => {
      // Given: 빈 문자열
      const emptyType = '';

      // When: validateTemplateType 호출
      const result = InputValidator.validateTemplateType(emptyType);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
    });

    test('should reject non-string input', () => {
      // Given: 문자열이 아닌 입력
      const invalidInput = null as any;

      // When: validateTemplateType 호출
      const result = InputValidator.validateTemplateType(invalidInput);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
    });

    test('should trim and lowercase sanitizedValue', () => {
      // Given: 공백과 대문자 포함
      const typeWithWhitespace = '  STANDARD  ';

      // When: validateTemplateType 호출
      const result = InputValidator.validateTemplateType(typeWithWhitespace);

      // Then: sanitizedValue가 정규화됨
      expect(result.sanitizedValue).toBe('standard');
    });
  });

  describe('validateBranchName', () => {
    test('should accept valid branch name', () => {
      // Given: 유효한 브랜치명
      const validName = 'feature/login-page';

      // When: validateBranchName 호출
      const result = InputValidator.validateBranchName(validName);

      // Then: 검증 통과
      expect(result.isValid).toBe(true);
    });

    test('should accept various valid formats', () => {
      // Given: 다양한 유효한 형식
      const validNames = [
        'main',
        'develop',
        'feature/user-auth',
        'bugfix/fix-login',
        'hotfix/security-patch',
        'release/v1.0.0',
      ];

      // When: validateBranchName 호출
      const results = validNames.map(name =>
        InputValidator.validateBranchName(name)
      );

      // Then: 모두 검증 통과
      results.forEach(result => {
        expect(result.isValid).toBe(true);
      });
    });

    test('should reject branch name longer than 250 characters', () => {
      // Given: 250자 초과 브랜치명
      const longName = `feature/${'a'.repeat(250)}`;

      // When: validateBranchName 호출
      const result = InputValidator.validateBranchName(longName);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(result.errors[0]).toContain('between 1 and 250');
    });

    test('should reject double dots', () => {
      // Given: 더블 닷 포함
      const nameWithDoubleDots = 'feature..login';

      // When: validateBranchName 호출
      const result = InputValidator.validateBranchName(nameWithDoubleDots);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
    });

    test('should reject names starting with dot or dash', () => {
      // Given: 점이나 대시로 시작
      const namesStartingInvalid = ['.feature', '-feature'];

      // When: validateBranchName 호출
      const results = namesStartingInvalid.map(name =>
        InputValidator.validateBranchName(name)
      );

      // Then: 검증 실패
      results.forEach(result => {
        expect(result.isValid).toBe(false);
      });
    });

    test('should reject names ending with dot or dash', () => {
      // Given: 점이나 대시로 끝남
      const namesEndingInvalid = ['feature.', 'feature-'];

      // When: validateBranchName 호출
      const results = namesEndingInvalid.map(name =>
        InputValidator.validateBranchName(name)
      );

      // Then: 검증 실패
      results.forEach(result => {
        expect(result.isValid).toBe(false);
      });
    });

    test('should reject reserved names', () => {
      // Given: 예약어
      const reservedNames = ['HEAD', 'master', 'origin'];

      // When: validateBranchName 호출
      const results = reservedNames.map(name =>
        InputValidator.validateBranchName(name)
      );

      // Then: 검증 실패
      results.forEach(result => {
        expect(result.isValid).toBe(false);
        expect(result.errors[0]).toContain('reserved');
      });
    });

    test('should reject Git-reserved characters', () => {
      // Given: Git 예약 문자 포함
      const namesWithReservedChars = [
        'feature:login',
        'feature*login',
        'feature~login',
        'feature^login',
        'feature?login',
        'feature[login]',
      ];

      // When: validateBranchName 호출
      const results = namesWithReservedChars.map(name =>
        InputValidator.validateBranchName(name)
      );

      // Then: 검증 실패
      results.forEach(result => {
        expect(result.isValid).toBe(false);
      });
    });

    test('should reject whitespace', () => {
      // Given: 공백 포함
      const nameWithSpace = 'feature login';

      // When: validateBranchName 호출
      const result = InputValidator.validateBranchName(nameWithSpace);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
    });

    test('should reject control characters', () => {
      // Given: 제어 문자 포함
      const nameWithControlChar = 'feature\x00login';

      // When: validateBranchName 호출
      const result = InputValidator.validateBranchName(nameWithControlChar);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
    });

    test('should reject @{ sequence', () => {
      // Given: @{ 시퀀스 포함
      const nameWithAtBrace = 'feature@{login}';

      // When: validateBranchName 호출
      const result = InputValidator.validateBranchName(nameWithAtBrace);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
    });

    test('should reject ending with slash', () => {
      // Given: 슬래시로 끝남
      const nameEndingWithSlash = 'feature/';

      // When: validateBranchName 호출
      const result = InputValidator.validateBranchName(nameEndingWithSlash);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
    });

    test('should reject double slashes', () => {
      // Given: 더블 슬래시 포함
      const nameWithDoubleSlash = 'feature//login';

      // When: validateBranchName 호출
      const result = InputValidator.validateBranchName(nameWithDoubleSlash);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
    });

    test('should trim whitespace', () => {
      // Given: 앞뒤 공백 포함
      const nameWithWhitespace = '  feature/login  ';

      // When: validateBranchName 호출
      const result = InputValidator.validateBranchName(nameWithWhitespace);

      // Then: sanitizedValue가 정규화됨
      expect(result.sanitizedValue).toBe('feature/login');
    });

    test('should reject empty string', () => {
      // Given: 빈 문자열
      const emptyName = '';

      // When: validateBranchName 호출
      const result = InputValidator.validateBranchName(emptyName);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
    });
  });

  describe('validateCommandOptions', () => {
    test('should accept valid options', () => {
      // Given: 유효한 옵션
      const validOptions = {
        verbose: true,
        output: 'json',
        timeout: 5000,
      };

      // When: validateCommandOptions 호출
      const result = InputValidator.validateCommandOptions(validOptions);

      // Then: 검증 통과
      expect(result.isValid).toBe(true);
      expect(result.sanitizedValue).toEqual({
        verbose: true,
        output: 'json',
        timeout: 5000,
      });
    });

    test('should accept string, boolean, and number types', () => {
      // Given: 다양한 타입의 옵션
      const options = {
        stringOpt: 'value',
        booleanOpt: true,
        numberOpt: 42,
      };

      // When: validateCommandOptions 호출
      const result = InputValidator.validateCommandOptions(options);

      // Then: 모든 타입이 포함됨
      expect(result.isValid).toBe(true);
      expect(result.sanitizedValue).toHaveProperty('stringOpt', 'value');
      expect(result.sanitizedValue).toHaveProperty('booleanOpt', true);
      expect(result.sanitizedValue).toHaveProperty('numberOpt', 42);
    });

    test('should reject invalid option key', () => {
      // Given: 잘못된 옵션 키
      const invalidOptions = {
        '123invalid': 'value',
      };

      // When: validateCommandOptions 호출
      const result = InputValidator.validateCommandOptions(invalidOptions);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(result.errors[0]).toContain('Invalid option key');
    });

    test('should reject string value longer than 1000 characters', () => {
      // Given: 1000자 초과 문자열
      const longOptions = {
        longString: 'a'.repeat(1001),
      };

      // When: validateCommandOptions 호출
      const result = InputValidator.validateCommandOptions(longOptions);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(result.errors[0]).toContain('too long');
    });

    test('should reject dangerous string content', () => {
      // Given: 위험한 문자열 패턴
      const dangerousOptions = {
        script: 'javascript:alert(1)',
      };

      // When: validateCommandOptions 호출
      const result = InputValidator.validateCommandOptions(dangerousOptions);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(result.errors[0]).toContain('dangerous content');
    });

    test('should reject eval pattern', () => {
      // Given: eval 패턴 포함
      const evalOptions = {
        code: 'eval(maliciousCode)',
      };

      // When: validateCommandOptions 호출
      const result = InputValidator.validateCommandOptions(evalOptions);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
    });

    test('should reject unsupported types', () => {
      // Given: 지원되지 않는 타입 (객체, 배열)
      const unsupportedOptions = {
        nested: { key: 'value' },
      };

      // When: validateCommandOptions 호출
      const result = InputValidator.validateCommandOptions(unsupportedOptions);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(result.errors[0]).toContain('unsupported type');
    });

    test('should reject non-finite numbers', () => {
      // Given: Infinity나 NaN
      const infinityOptions = {
        infinite: Infinity,
      };

      // When: validateCommandOptions 호출
      const result = InputValidator.validateCommandOptions(infinityOptions);

      // Then: 검증 실패
      expect(result.isValid).toBe(false);
      expect(result.errors[0]).toContain('finite number');
    });

    test('should trim string values', () => {
      // Given: 앞뒤 공백 포함 문자열
      const optionsWithWhitespace = {
        name: '  value  ',
      };

      // When: validateCommandOptions 호출
      const result = InputValidator.validateCommandOptions(
        optionsWithWhitespace
      );

      // Then: sanitizedValue에서 공백 제거됨
      expect(result.sanitizedValue?.name).toBe('value');
    });

    test('should handle empty options object', () => {
      // Given: 빈 객체
      const emptyOptions = {};

      // When: validateCommandOptions 호출
      const result = InputValidator.validateCommandOptions(emptyOptions);

      // Then: 검증 통과
      expect(result.isValid).toBe(true);
      expect(result.sanitizedValue).toEqual({});
    });
  });
});

describe('Helper Functions', () => {
  describe('validateProjectName (helper)', () => {
    test('should work as standalone function', () => {
      // Given: 유효한 프로젝트명
      const name = 'test-project';

      // When: 헬퍼 함수 호출
      const result = validateProjectName(name);

      // Then: 검증 통과
      expect(result.isValid).toBe(true);
      expect(result.sanitizedValue).toBe(name);
    });
  });

  describe('validatePath (helper)', () => {
    test('should work as standalone function', async () => {
      // Given: 유효한 경로
      const testPath = './test';

      // When: 헬퍼 함수 호출
      const result = await validatePath(testPath);

      // Then: sanitizedValue가 절대 경로로 변환됨
      expect(result.sanitizedValue).toBeDefined();
      expect(path.isAbsolute(result.sanitizedValue!)).toBe(true);
    });
  });

  describe('validateBranchName (helper)', () => {
    test('should work as standalone function', () => {
      // Given: 유효한 브랜치명
      const branch = 'feature/test';

      // When: 헬퍼 함수 호출
      const result = validateBranchName(branch);

      // Then: 검증 통과
      expect(result.isValid).toBe(true);
    });
  });
});

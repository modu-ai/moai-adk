// @TEST:REFACTOR-003 | SPEC: SPEC-REFACTOR-003.md

/**
 * tag-validator.ts 테스트
 * TAG 검증 로직 테스트
 */

import { beforeEach, describe, test, expect } from 'vitest';
import { TagValidator } from '../../../../src/claude/hooks/tag-enforcer/tag-validator';

describe('@TEST:REFACTOR-003: TAG Validator', () => {
  let validator: TagValidator;

  beforeEach(() => {
    validator = new TagValidator();
  });

  describe('extractTagBlock', () => {
    test('should extract valid TAG block from top of file', () => {
      const content = `/**
 * @DOC:FEATURE:AUTH-001
 * CHAIN: REQ:AUTH-001 -> DESIGN:AUTH-001
 * @IMMUTABLE
 */

export class AuthService {}`;

      const result = validator.extractTagBlock(content);

      expect(result).not.toBeNull();
      expect(result?.content).toContain('@DOC:FEATURE:AUTH-001');
      expect(result?.content).toContain('@IMMUTABLE');
      expect(result?.lineNumber).toBe(1);
    });

    test('should return null if no TAG block exists', () => {
      const content = `// Regular comment
export class Service {}`;

      const result = validator.extractTagBlock(content);

      expect(result).toBeNull();
    });

    test('should ignore non-TAG comment blocks', () => {
      const content = `/**
 * Regular JSDoc comment
 * No TAG here
 */
export class Service {}`;

      const result = validator.extractTagBlock(content);

      expect(result).toBeNull();
    });

    test('should handle TAG block after shebang', () => {
      const content = `#!/usr/bin/env node

/**
 * @DOC:FEATURE:CLI-001
 */

import * as fs from 'fs';`;

      const result = validator.extractTagBlock(content);

      expect(result).not.toBeNull();
      expect(result?.content).toContain('@DOC:FEATURE:CLI-001');
    });
  });

  describe('extractMainTag', () => {
    test('should extract TAG from block content', () => {
      const blockContent = `/**
 * @DOC:CODE:AUTH-001
 * @IMMUTABLE
 */`;

      const tag = validator.extractMainTag(blockContent);

      expect(tag).toBe('@CODE:AUTH-001');
    });

    test('should return UNKNOWN if no TAG found', () => {
      const blockContent = `/**
 * Regular comment
 */`;

      const tag = validator.extractMainTag(blockContent);

      expect(tag).toBe('UNKNOWN');
    });
  });

  describe('normalizeTagBlock', () => {
    test('should normalize TAG block content', () => {
      const blockContent = `/**
 * @DOC:FEATURE:AUTH-001
 *
 * @IMMUTABLE
 */`;

      const normalized = validator.normalizeTagBlock(blockContent);

      expect(normalized).not.toContain('  ');
      expect(normalized.split('\n').every(line => line.trim() !== '')).toBe(true);
    });
  });

  describe('validateCodeFirstTag', () => {
    test('should validate valid TAG block', () => {
      const content = `/**
 * @DOC:FEATURE:AUTH-001
 * CHAIN: REQ:AUTH-001 -> DESIGN:AUTH-001
 * STATUS: active
 * CREATED: 2025-01-15
 * @IMMUTABLE
 */

export class AuthService {}`;

      const result = validator.validateCodeFirstTag(content);

      expect(result.isValid).toBe(true);
      expect(result.violations).toHaveLength(0);
      expect(result.hasTag).toBe(true);
    });

    test('should return warnings for missing @IMMUTABLE', () => {
      const content = `/**
 * @DOC:FEATURE:AUTH-001
 */

export class AuthService {}`;

      const result = validator.validateCodeFirstTag(content);

      expect(result.isValid).toBe(true);
      expect(result.warnings.some(w => w.includes('@IMMUTABLE'))).toBe(true);
    });

    test('should reject invalid TAG category', () => {
      const content = `/**
 * @DOC:INVALID:AUTH-001
 */

export class AuthService {}`;

      const result = validator.validateCodeFirstTag(content);

      expect(result.isValid).toBe(false);
      expect(result.violations.some(v => v.includes('유효하지 않은'))).toBe(true);
    });

    test('should validate without blocking when no TAG exists', () => {
      const content = `export class AuthService {}`;

      const result = validator.validateCodeFirstTag(content);

      expect(result.isValid).toBe(true);
      expect(result.hasTag).toBe(false);
      expect(result.warnings).toHaveLength(1);
    });

    test('should warn about domain ID format', () => {
      const content = `/**
 * @DOC:FEATURE:AUTH_001
 */

export class AuthService {}`;

      const result = validator.validateCodeFirstTag(content);

      expect(result.warnings.some(w => w.includes('도메인 ID 형식'))).toBe(true);
    });

    test('should validate chain references', () => {
      const content = `/**
 * @DOC:FEATURE:AUTH-001
 * CHAIN: REQ:AUTH-001 -> DESIGN:AUTH-001 -> TASK:AUTH-001
 */

export class AuthService {}`;

      const result = validator.validateCodeFirstTag(content);

      expect(result.isValid).toBe(true);
    });

    test('should validate status values', () => {
      const content = `/**
 * @DOC:FEATURE:AUTH-001
 * STATUS: active
 */

export class AuthService {}`;

      const result = validator.validateCodeFirstTag(content);

      expect(result.isValid).toBe(true);
      expect(result.warnings).not.toContain('알 수 없는 STATUS');
    });

    test('should warn about unknown status', () => {
      const content = `/**
 * @DOC:FEATURE:AUTH-001
 * STATUS: unknown
 */

export class AuthService {}`;

      const result = validator.validateCodeFirstTag(content);

      expect(result.warnings.some(w => w.includes('알 수 없는 STATUS'))).toBe(true);
    });
  });
});

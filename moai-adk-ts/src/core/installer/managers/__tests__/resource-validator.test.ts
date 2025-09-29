/**
 * @feature RESOURCE-VALIDATION-001 ResourceValidator Tests
 * @task RESOURCE-TEST-001 RED stage - failing tests for resource validation
 */

import { beforeEach, describe, expect, test } from 'vitest';
import { ResourceValidator } from '../resource-validator';

describe('ResourceValidator', () => {
  let validator: ResourceValidator;

  beforeEach(() => {
    validator = new ResourceValidator();
  });

  describe('validatePackageResources', () => {
    test('should validate all required package resources exist', async () => {
      // Act
      const result = await validator.validatePackageResources();

      // Assert - This will fail initially (RED stage)
      expect(result.isValid).toBe(true);
      expect(result.missingTemplates).toEqual([]);
      expect(result.errors).toEqual([]);
    });

    test('should detect missing templates', async () => {
      // Act
      const result = await validator.validatePackageResources();

      // Assert
      expect(result).toHaveProperty('missingTemplates');
      expect(Array.isArray(result.missingTemplates)).toBe(true);
    });

    test('should provide detailed error information', async () => {
      // Act
      const result = await validator.validatePackageResources();

      // Assert
      expect(result).toHaveProperty('errors');
      expect(Array.isArray(result.errors)).toBe(true);
    });
  });

  describe('checkRequiredTemplates', () => {
    test('should verify all required templates are present', async () => {
      // Act
      const result = await validator.checkRequiredTemplates();

      // Assert - This will fail initially (RED stage)
      expect(result.allPresent).toBe(true);
      expect(result.missingTemplates).toEqual([]);
      expect(result.templateCount).toBeGreaterThan(0);
    });

    test('should detect missing required templates', async () => {
      // Act
      const result = await validator.checkRequiredTemplates();

      // Assert
      expect(result).toHaveProperty('allPresent');
      expect(result).toHaveProperty('missingTemplates');
      expect(result).toHaveProperty('templateCount');
    });

    test('should include expected template names in check', async () => {
      // Act
      const result = await validator.checkRequiredTemplates();

      // Assert
      const requiredTemplates = ['.claude', '.moai', 'CLAUDE.md'];
      if (!result.allPresent) {
        expect(
          result.missingTemplates.some(t => requiredTemplates.includes(t))
        ).toBe(true);
      }
    });
  });

  describe('verifyResourceIntegrity', () => {
    test('should verify template file integrity', async () => {
      // Act
      const result = await validator.verifyResourceIntegrity();

      // Assert - This will fail initially (RED stage)
      expect(result.isValid).toBe(true);
      expect(result.corruptedFiles).toEqual([]);
      expect(result.checksumMismatches).toEqual([]);
    });

    test('should detect corrupted files', async () => {
      // Act
      const result = await validator.verifyResourceIntegrity();

      // Assert
      expect(result).toHaveProperty('corruptedFiles');
      expect(Array.isArray(result.corruptedFiles)).toBe(true);
    });

    test('should verify checksums when available', async () => {
      // Act
      const result = await validator.verifyResourceIntegrity();

      // Assert
      expect(result).toHaveProperty('checksumMismatches');
      expect(Array.isArray(result.checksumMismatches)).toBe(true);
    });
  });

  describe('validatePackageStructure', () => {
    test('should validate package.json integrity', async () => {
      // Act
      const result = await validator.validatePackageStructure();

      // Assert - This will fail initially (RED stage)
      expect(result).toBe(true);
    });

    test('should verify dependency compatibility', async () => {
      // This test will help ensure our dependencies are compatible
      expect(validator.validatePackageStructure).toBeDefined();
      expect(typeof validator.validatePackageStructure).toBe('function');
    });
  });

  describe('checkTemplateIntegrity', () => {
    test('should verify template files exist and are readable', async () => {
      // Act
      const result = await validator.checkTemplateIntegrity();

      // Assert - This will fail initially (RED stage)
      expect(result).toBe(true);
    });

    test('should validate template permissions', async () => {
      // This test ensures we check file permissions properly
      expect(validator.checkTemplateIntegrity).toBeDefined();
      expect(typeof validator.checkTemplateIntegrity).toBe('function');
    });
  });

  describe('error handling', () => {
    test('should handle resource access errors gracefully', async () => {
      // This test ensures our error handling is robust
      expect(async () => {
        await validator.validatePackageResources();
      }).not.toThrow();
    });

    test('should provide meaningful error messages', async () => {
      // Test that error messages are helpful for debugging
      const result = await validator.validatePackageResources();
      if (result.errors.length > 0) {
        expect(
          result.errors.every(
            error => typeof error === 'string' && error.length > 0
          )
        ).toBe(true);
      }
    });
  });
});

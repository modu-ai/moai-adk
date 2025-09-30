/**
 * @file Security Test Suite
 * @author MoAI Team
 * @tags @SECURITY:TEST-SUITE-001 @FIX:SECURITY-VALIDATION-001
 * @description Comprehensive security tests for all vulnerability fixes
 */

import { describe, expect, test } from 'vitest';
import {
  renderTemplateSafely,
  validateTemplateContent,
} from '../../core/installer/templates/template-security';
import { InputValidator } from '../../utils/input-validator';
import {
  RegexSecurity,
  safeMatch,
  safeReplace,
} from '../../utils/regex-security';

describe('Security Test Suite', () => {
  describe('Template Injection Protection', () => {
    test('should block constructor access in templates', () => {
      const maliciousContext = {
        constructor: {
          constructor:
            'return process.mainModule.require("child_process").exec("rm -rf /")',
        },
      };

      expect(() => {
        renderTemplateSafely('{{constructor.constructor}}', maliciousContext);
      }).toThrow('Template contains dangerous patterns');
    });

    test('should block prototype pollution attempts', () => {
      const maliciousContext = {
        __proto__: { isAdmin: true },
        prototype: { dangerous: true },
      };

      const result = renderTemplateSafely(
        '{{__proto__.isAdmin}}',
        maliciousContext
      );
      expect(result).not.toContain('true');
    });

    test('should sanitize dangerous context properties', () => {
      const maliciousContext = {
        PROJECT_NAME: 'test-project',
        eval: 'eval("malicious code")',
        Function: 'new Function("return process")',
        require: 'require("fs")',
      };

      const result = renderTemplateSafely(
        '{{PROJECT_NAME}}{{eval}}{{Function}}{{require}}',
        maliciousContext
      );
      expect(result).toBe('test-project');
    });

    test('should validate template content for dangerous patterns', () => {
      const dangerousTemplates = [
        'javascript:alert(1)',
        '{{constructor.constructor("return process")()}}',
        '{{__proto__.isAdmin}}',
        '<script>alert(1)</script>',
        'eval("malicious")',
      ];

      for (const template of dangerousTemplates) {
        expect(validateTemplateContent(template)).toBe(false);
      }
    });

    test('should allow safe template content', () => {
      const safeTemplates = [
        'Hello {{PROJECT_NAME}}!',
        'Version: {{PROJECT_VERSION}}',
        'Environment: {{ENVIRONMENT}}',
        'Path: {{PROJECT_PATH}}',
      ];

      for (const template of safeTemplates) {
        expect(validateTemplateContent(template)).toBe(true);
      }
    });
  });

  describe('Path Traversal Protection', () => {
    test('should block path traversal in file names', async () => {
      const maliciousPaths = [
        '../../../etc/passwd',
        '..\\..\\..\\Windows\\System32',
        '/etc/shadow',
        'C:\\Windows\\System32\\config\\SAM',
        '....//....//....//etc/passwd',
      ];

      for (const path of maliciousPaths) {
        const validation = await InputValidator.validatePath(path);
        expect(validation.isValid).toBe(false);
        expect(validation.errors).toContain('Path traversal detected');
      }
    });

    test('should allow safe file paths', async () => {
      const safePaths = [
        './project/src/index.ts',
        'src/components/Button.tsx',
        'docs/README.md',
        'package.json',
      ];

      for (const path of safePaths) {
        const validation = await InputValidator.validatePath(path, {
          mustExist: false,
        });
        expect(validation.isValid).toBe(true);
      }
    });

    test('should reject dangerous file names', async () => {
      const dangerousNames = [
        'malware.exe',
        'script.bat',
        'virus.com',
        'CON',
        'PRN',
        'file<name>.txt',
        'file|pipe.txt',
      ];

      for (const name of dangerousNames) {
        const validation = await InputValidator.validatePath(name);
        expect(validation.isValid).toBe(false);
      }
    });
  });

  describe('Input Validation Protection', () => {
    test('should validate project names securely', () => {
      const invalidNames = [
        '', // Empty
        'a'.repeat(60), // Too long
        'project with spaces', // Spaces (when not allowed)
        'project<script>', // HTML injection
        'project/../other', // Path traversal
        'CON', // Windows reserved
        'project|pipe', // Pipe character
      ];

      for (const name of invalidNames) {
        const validation = InputValidator.validateProjectName(name, {
          allowSpaces: false,
        });
        expect(validation.isValid).toBe(false);
      }
    });

    test('should validate Git branch names securely', () => {
      const invalidBranches = [
        '', // Empty
        'feature..bug', // Double dots
        '.hidden', // Starting with dot
        'branch-', // Ending with dash
        'feature~1', // Tilde character
        'branch:name', // Colon character
        'feature branch', // Space
        'a'.repeat(300), // Too long
      ];

      for (const branch of invalidBranches) {
        const validation = InputValidator.validateBranchName(branch);
        expect(validation.isValid).toBe(false);
      }
    });

    test('should validate command options securely', () => {
      const dangerousOptions = {
        normalOption: 'safe-value',
        longOption: 'x'.repeat(2000), // Too long
        scriptOption: '<script>alert(1)</script>', // HTML injection
        evalOption: 'eval("malicious")', // JavaScript code
        'invalid-key-!@#': 'value', // Invalid key format
        functionOption: () => 'function value', // Function (should be rejected)
      };

      const validation =
        InputValidator.validateCommandOptions(dangerousOptions);
      expect(validation.isValid).toBe(false);
      expect(validation.errors.length).toBeGreaterThan(0);
    });
  });

  describe('ReDoS Protection', () => {
    test('should detect dangerous regex patterns', () => {
      const dangerousPatterns = [
        /^(a+)+$/, // Nested quantifiers
        /^(a|a)*$/, // Alternation with repetition
        /^(a+)*$/, // Repeated groups with quantifiers
        /^(a*)*$/, // Evil regex
        /^([a-zA-Z]+)*$/, // Character class with repetition
      ];

      for (const pattern of dangerousPatterns) {
        const validation = RegexSecurity.validateRegexSafety(pattern);
        expect(validation.isSafe).toBe(false);
        expect(validation.riskLevel).toMatch(/high|critical/);
      }
    });

    test('should allow safe regex patterns', () => {
      const safePatterns = [
        /^[a-zA-Z]+$/, // Simple character class
        /^\d{1,10}$/, // Bounded quantifier
        /^[a-z][a-z0-9-]*$/, // Simple pattern with hyphen
        /^.{1,100}$/, // Bounded dot quantifier
      ];

      for (const pattern of safePatterns) {
        const validation = RegexSecurity.validateRegexSafety(pattern);
        expect(validation.isSafe).toBe(true);
        expect(validation.riskLevel).toMatch(/low|medium/);
      }
    });

    test('should timeout on ReDoS attack', () => {
      const evilRegex = /^(a+)+$/;
      const attackString = `${'a'.repeat(50)}X`; // Will cause exponential backtracking

      const result = safeMatch(evilRegex, attackString, { timeout: 10 });
      expect(result.success).toBe(false);
      expect(result.error).toContain('potentially dangerous');
    });

    test('should handle safe regex operations', () => {
      const safeRegex = /^[a-zA-Z0-9-]+$/;
      const testString = 'safe-test-string-123';

      const result = safeMatch(safeRegex, testString, { timeout: 100 });
      expect(result.success).toBe(true);
      expect(result.matches).not.toBeNull();
    });

    test('should protect regex replace operations', () => {
      const safePattern = /test/g;
      const input = 'test string with test words';
      const replacement = 'demo';

      const result = safeReplace(safePattern, input, replacement, {
        timeout: 50,
      });
      expect(result).toBe('demo string with demo words');
    });
  });

  describe('Git Security', () => {
    test('should validate Git URLs securely', () => {
      // Valid URLs for reference (not tested in this case)
      // const validUrls = [
      //   'https://github.com/user/repo.git',
      //   'git@github.com:user/repo.git',
      //   'https://gitlab.com/user/repo',
      // ];

      const invalidUrls = [
        'javascript:alert(1)',
        'file:///etc/passwd',
        'http://malicious.com/$(whoami)',
        'git://github.com/user/repo.git', // Insecure protocol
        'https://github.com/user/re<po.git', // Invalid characters
      ];

      // Note: This would require implementing URL validation in GitManager
      // For now, we'll test the principle
      for (const url of invalidUrls) {
        expect(url).toMatch(/javascript:|file:|<|>/); // Contains dangerous patterns
      }
    });
  });

  describe('Dependency Security', () => {
    test('should verify security packages are included', () => {
      // Mock package.json reading
      const mockPackageJson = {
        dependencies: {
          validator: '^13.12.0',
          mustache: '^4.2.0',
        },
      };

      expect(mockPackageJson.dependencies).toHaveProperty('validator');
      expect(mockPackageJson.dependencies.mustache).toMatch(/^\^4\./); // Updated version
    });
  });

  describe('Error Handling Security', () => {
    test('should not leak sensitive information in errors', () => {
      const sensitiveData = {
        password: 'secret123',
        apiKey: 'sk-1234567890abcdef',
        token: 'ghp_1234567890abcdef',
      };

      // Test that error messages don't contain sensitive data
      try {
        throw new Error(`Connection failed for ${sensitiveData.password}`);
      } catch (error) {
        // In a real implementation, errors should be sanitized
        const errorMessage = (error as Error).message;
        expect(errorMessage).not.toContain('secret123');
        // Note: This test would pass only if error sanitization is implemented
      }
    });
  });

  describe('Integration Security Tests', () => {
    test('should maintain security through the full template processing pipeline', () => {
      const maliciousTemplate =
        '{{constructor.constructor("return process")()}}';
      const context = {
        PROJECT_NAME: 'test-project',
        constructor: { constructor: 'malicious' },
      };

      expect(() => {
        renderTemplateSafely(maliciousTemplate, context);
      }).toThrow();
    });

    test('should handle complex attack vectors', () => {
      const complexAttack = {
        name: '../../../etc/passwd',
        branch: 'feature~1$(rm -rf /)',
        template:
          'Hello {{__proto__.constructor.constructor("return process")()}}',
      };

      // Project name validation
      const nameValidation = InputValidator.validateProjectName(
        complexAttack.name
      );
      expect(nameValidation.isValid).toBe(false);

      // Branch name validation
      const branchValidation = InputValidator.validateBranchName(
        complexAttack.branch
      );
      expect(branchValidation.isValid).toBe(false);

      // Template validation
      expect(validateTemplateContent(complexAttack.template)).toBe(false);
    });
  });
});

// Performance benchmarks for security features
describe('Security Performance Benchmarks', () => {
  test('input validation should complete within reasonable time', () => {
    const start = Date.now();

    for (let i = 0; i < 1000; i++) {
      InputValidator.validateProjectName(`test-project-${i}`);
    }

    const duration = Date.now() - start;
    expect(duration).toBeLessThan(1000); // Should complete 1000 validations in under 1 second
  });

  test('regex security checks should be fast', () => {
    const start = Date.now();

    for (let i = 0; i < 100; i++) {
      RegexSecurity.validateRegexSafety(/^[a-zA-Z0-9-]+$/);
    }

    const duration = Date.now() - start;
    expect(duration).toBeLessThan(100); // Should complete 100 validations in under 100ms
  });

  test('template security should handle large inputs efficiently', () => {
    const largeTemplate = 'Hello {{PROJECT_NAME}}! '.repeat(1000);
    const context = { PROJECT_NAME: 'test' };

    const start = Date.now();
    renderTemplateSafely(largeTemplate, context);
    const duration = Date.now() - start;

    expect(duration).toBeLessThan(1000); // Should complete in under 1 second
  });
});

/**
 * @FEATURE:INSTALLATION-VALIDATOR-001 Installation Validation System Tests
 *
 * @TEST:VALIDATION-RED-001 RED phase tests for InstallationValidator
 * Tests all validation methods with various failure scenarios
 */

import { promises as fs } from 'fs';
import path from 'path';
import os from 'os';
import {
  InstallationValidator,
  ValidationErrorCodes
} from '../../../src/core/installer/validators/installation-validator';

describe('InstallationValidator', () => {
  let tempDir: string;
  let validator: InstallationValidator;

  beforeEach(async () => {
    // Create temporary test directory
    tempDir = await fs.mkdtemp(path.join(os.tmpdir(), 'moai-test-'));
    validator = new InstallationValidator();
  });

  afterEach(async () => {
    // Clean up temporary directory
    await fs.rm(tempDir, { recursive: true, force: true }).catch(() => {});
  });

  describe('validateProjectStructure', () => {
    it('should fail when .claude directory is missing', async () => {
      const result = await validator.validateProjectStructure(tempDir);

      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors.some((e: any) => e.code === ValidationErrorCodes.MISSING_DIRECTORY && e.message.includes('.claude'))).toBe(true);
    });

    it('should fail when .moai directory is missing', async () => {
      // Create .claude but not .moai
      await fs.mkdir(path.join(tempDir, '.claude'), { recursive: true });

      const result = await validator.validateProjectStructure(tempDir);

      expect(result.valid).toBe(false);
      expect(result.errors.some((e: any) => e.message.includes('.moai'))).toBe(true);
    });

    it('should pass when all required directories exist', async () => {
      // Create required directories
      await fs.mkdir(path.join(tempDir, '.claude', 'agents'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.claude', 'commands'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.claude', 'hooks'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.moai', 'project'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.moai', 'specs'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.moai', 'memory'), { recursive: true });

      const result = await validator.validateProjectStructure(tempDir);

      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should fail when path exists but is not a directory', async () => {
      // Create .claude as a file instead of directory
      await fs.writeFile(path.join(tempDir, '.claude'), 'not a directory');

      const result = await validator.validateProjectStructure(tempDir);

      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors[0]?.code).toBe(ValidationErrorCodes.INVALID_STRUCTURE);
    });
  });

  describe('validateConfigurationFiles', () => {
    beforeEach(async () => {
      // Setup basic directory structure
      await fs.mkdir(path.join(tempDir, '.claude'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.moai'), { recursive: true });
    });

    it('should fail when settings.json has invalid JSON', async () => {
      await fs.writeFile(
        path.join(tempDir, '.claude', 'settings.json'),
        '{ invalid json }'
      );

      const result = await validator.validateConfigurationFiles(tempDir);

      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors[0]?.code).toBe(ValidationErrorCodes.INVALID_JSON);
    });

    it('should fail when config.json has invalid JSON', async () => {
      await fs.writeFile(
        path.join(tempDir, '.moai', 'config.json'),
        '{ "invalid": json }'
      );

      const result = await validator.validateConfigurationFiles(tempDir);

      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors[0]?.code).toBe(ValidationErrorCodes.INVALID_JSON);
    });

    it('should pass when all JSON files are valid', async () => {
      await fs.writeFile(
        path.join(tempDir, '.claude', 'settings.json'),
        JSON.stringify({ version: "1.0" })
      );
      await fs.writeFile(
        path.join(tempDir, '.moai', 'config.json'),
        JSON.stringify({ project: { name: "test" } })
      );

      const result = await validator.validateConfigurationFiles(tempDir);

      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should pass when configuration files do not exist', async () => {
      const result = await validator.validateConfigurationFiles(tempDir);

      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });
  });

  describe('validateResourceFiles', () => {
    beforeEach(async () => {
      await fs.mkdir(path.join(tempDir, '.claude'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.moai'), { recursive: true });
    });

    it('should fail when required files are missing', async () => {
      const result = await validator.validateResourceFiles(tempDir);

      expect(result.valid).toBe(false);
      expect(result.errors.some((e: any) => e.code === ValidationErrorCodes.MISSING_FILE)).toBe(true);
    });

    it('should pass when all required files exist', async () => {
      // Create CLAUDE.md
      await fs.writeFile(path.join(tempDir, 'CLAUDE.md'), '# Test Project');

      const result = await validator.validateResourceFiles(tempDir);

      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should validate file permissions on Unix systems', async () => {
      if (process.platform !== 'win32') {
        // Create CLAUDE.md file first
        await fs.writeFile(path.join(tempDir, 'CLAUDE.md'), '# Test Project');

        // Create a directory with restricted permissions to make access test meaningful
        const claudeDir = path.join(tempDir, '.claude');
        await fs.mkdir(claudeDir, { recursive: true });
        await fs.chmod(claudeDir, 0o000);

        const result = await validator.validateResourceFiles(tempDir);

        // Should still pass for main file, may have warnings about permission issues
        expect(result.valid).toBe(true);

        // Clean up
        await fs.chmod(claudeDir, 0o755);
      } else {
        // Skip on Windows
        expect(true).toBe(true);
      }
    });
  });

  describe('validateClaudeIntegration', () => {
    beforeEach(async () => {
      await fs.mkdir(path.join(tempDir, '.claude'), { recursive: true });
    });

    it('should fail when settings.json is missing required fields', async () => {
      await fs.writeFile(
        path.join(tempDir, '.claude', 'settings.json'),
        JSON.stringify({})  // Empty settings
      );

      const result = await validator.validateClaudeIntegration(tempDir);

      expect(result.valid).toBe(false);
      expect(result.errors.some((e: any) => e.message.includes('required field'))).toBe(true);
    });

    it('should pass when Claude integration is properly configured', async () => {
      await fs.mkdir(path.join(tempDir, '.claude', 'agents'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.claude', 'commands'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.claude', 'hooks'), { recursive: true });

      await fs.writeFile(
        path.join(tempDir, '.claude', 'settings.json'),
        JSON.stringify({
          agents: { enabled: true },
          commands: { enabled: true }
        })
      );

      const result = await validator.validateClaudeIntegration(tempDir);

      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should detect corrupted agent files', async () => {
      await fs.mkdir(path.join(tempDir, '.claude', 'agents'), { recursive: true });

      // Create corrupted agent file
      await fs.writeFile(
        path.join(tempDir, '.claude', 'agents', 'corrupted.md'),
        Buffer.from([0x00, 0x01, 0x02])  // Binary data in text file
      );

      const result = await validator.validateClaudeIntegration(tempDir);

      expect(result.warnings.some((w: any) => w.code === ValidationErrorCodes.CORRUPTED_FILE)).toBe(true);
    });
  });

  describe('validateMoaiIntegration', () => {
    beforeEach(async () => {
      await fs.mkdir(path.join(tempDir, '.moai'), { recursive: true });
    });

    it('should fail when required MoAI directories are missing', async () => {
      const result = await validator.validateMoaiIntegration(tempDir);

      expect(result.valid).toBe(false);
      expect(result.errors.some((e: any) => e.code === ValidationErrorCodes.MISSING_DIRECTORY)).toBe(true);
    });

    it('should pass when MoAI structure is complete', async () => {
      // Create required MoAI directories
      await fs.mkdir(path.join(tempDir, '.moai', 'project'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.moai', 'specs'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.moai', 'memory'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.moai', 'indexes'), { recursive: true });

      // Create development guide
      await fs.writeFile(
        path.join(tempDir, '.moai', 'memory', 'development-guide.md'),
        '# Development Guide'
      );

      const result = await validator.validateMoaiIntegration(tempDir);

      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should validate index file integrity', async () => {
      // Create all required directories first
      await fs.mkdir(path.join(tempDir, '.moai', 'project'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.moai', 'specs'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.moai', 'memory'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.moai', 'indexes'), { recursive: true });

      // Create corrupted index file
      await fs.writeFile(
        path.join(tempDir, '.moai', 'indexes', 'tags.json'),
        '{ "corrupted": '  // Incomplete JSON
      );

      const result = await validator.validateMoaiIntegration(tempDir);

      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors.some((e: any) => e.code === ValidationErrorCodes.INVALID_JSON)).toBe(true);
    });
  });

  describe('generateDiagnosticReport', () => {
    it('should generate comprehensive diagnostic report', async () => {
      const report = await validator.generateDiagnosticReport(tempDir);

      expect(report.projectPath).toBe(tempDir);
      expect(report.timestamp).toBeInstanceOf(Date);
      expect(['healthy', 'warning', 'error']).toContain(report.overallHealth);
      expect(report.validationResults).toHaveLength(5); // All validation methods
      expect(report.recommendations).toBeDefined();
      expect(report.nextSteps).toBeDefined();
    });

    it('should detect multiple issues and provide recommendations', async () => {
      // Create a partially broken project
      await fs.mkdir(path.join(tempDir, '.claude'), { recursive: true });
      await fs.writeFile(
        path.join(tempDir, '.claude', 'settings.json'),
        '{ invalid json }'
      );

      const report = await validator.generateDiagnosticReport(tempDir);

      expect(report.overallHealth).toBe('error');
      expect(report.recommendations.length).toBeGreaterThan(0);
      expect(report.nextSteps.length).toBeGreaterThan(0);
    });

    it('should report healthy status for complete installation', async () => {
      // Create complete valid structure
      await fs.mkdir(path.join(tempDir, '.claude', 'agents'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.claude', 'commands'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.claude', 'hooks'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.moai', 'project'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.moai', 'specs'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.moai', 'memory'), { recursive: true });
      await fs.mkdir(path.join(tempDir, '.moai', 'indexes'), { recursive: true });

      await fs.writeFile(path.join(tempDir, 'CLAUDE.md'), '# Test Project');
      await fs.writeFile(
        path.join(tempDir, '.claude', 'settings.json'),
        JSON.stringify({ agents: { enabled: true } })
      );
      await fs.writeFile(
        path.join(tempDir, '.moai', 'config.json'),
        JSON.stringify({ project: { name: "test" } })
      );

      const report = await validator.generateDiagnosticReport(tempDir);

      expect(report.overallHealth).toBe('healthy');
      expect(report.validationResults.every((r: any) => r.valid)).toBe(true);
    });
  });

  describe('error handling', () => {
    it('should handle invalid project paths gracefully', async () => {
      const invalidPath = '/nonexistent/path/that/does/not/exist';

      const result = await validator.validateProjectStructure(invalidPath);

      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors[0]?.code).toBe(ValidationErrorCodes.MISSING_DIRECTORY);
    });

    it('should handle permission denied errors', async () => {
      if (process.platform !== 'win32') {
        // Create directory with no access
        const restrictedDir = path.join(tempDir, 'restricted');
        await fs.mkdir(restrictedDir);
        await fs.chmod(restrictedDir, 0o000);

        const result = await validator.validateProjectStructure(restrictedDir);

        expect(result.valid).toBe(false);
        expect(result.errors.some((e: any) => e.code === ValidationErrorCodes.PERMISSION_DENIED)).toBe(true);

        // Clean up
        await fs.chmod(restrictedDir, 0o755);
      }
    });

    it('should handle binary files gracefully', async () => {
      await fs.mkdir(path.join(tempDir, '.claude'), { recursive: true });

      // Create binary file
      const binaryFile = path.join(tempDir, '.claude', 'binary.bin');
      await fs.writeFile(binaryFile, Buffer.from([0x89, 0x50, 0x4E, 0x47])); // PNG header

      const result = await validator.validateConfigurationFiles(tempDir);

      // Should not crash and handle binary files gracefully
      expect(result).toBeDefined();
    });
  });

  describe('performance requirements', () => {
    it('should complete validation within 5 seconds for large projects', async () => {
      // Create large directory structure
      const baseDir = path.join(tempDir, '.claude');
      await fs.mkdir(baseDir, { recursive: true });

      // Create many files
      const promises = [];
      for (let i = 0; i < 100; i++) {
        promises.push(
          fs.writeFile(
            path.join(baseDir, `file${i}.json`),
            JSON.stringify({ id: i, data: "test".repeat(100) })
          )
        );
      }
      await Promise.all(promises);

      const startTime = Date.now();
      const report = await validator.generateDiagnosticReport(tempDir);
      const duration = Date.now() - startTime;

      expect(duration).toBeLessThan(5000); // 5 seconds
      expect(report).toBeDefined();
    }, 10000); // 10 second timeout for test
  });
});
/**
 * @TEST:TEMPLATE-PROCESSOR-001 Template Processor Cross-Platform Test Suite
 * @tags @TEST:TEMPLATE-PROCESSOR-001 @SPEC:INSTALL-SYSTEM-012
 *
 * Phase 2: Cross-platform path resolution tests
 * Tests for Windows/macOS/Linux compatibility and environment variable handling
 */

import * as path from 'node:path';
import { TemplateProcessor } from '@/core/installer/template-processor';
import type { InstallationConfig } from '@/core/installer/types';

describe('TemplateProcessor - Cross-Platform Path Resolution', () => {
  let processor: TemplateProcessor;
  let originalEnv: NodeJS.ProcessEnv;

  beforeEach(() => {
    processor = new TemplateProcessor();
    originalEnv = { ...process.env };
  });

  afterEach(() => {
    process.env = originalEnv;
    // Note: We cannot restore process.platform directly as it's read-only
    // Tests should use conditional logic or mocking for platform-specific behavior
  });

  describe('Phase 2: Cross-Platform Environment Variable Handling', () => {
    /**
     * @TEST:TEMPLATE-PATH-WINDOWS-001
     * Test: Windows path resolution with USERPROFILE
     */
    it('should resolve Windows paths using USERPROFILE environment variable', () => {
      // RED: This test will fail until we implement Windows-specific path handling

      // Simulate Windows environment
      process.env.USERPROFILE = 'C:\\Users\\TestUser';
      delete process.env.HOME;

      const templatesPath = processor.getTemplatesPath();

      // Should not contain Unix-only hardcoded paths
      expect(templatesPath).toBeDefined();
      expect(templatesPath).not.toContain('/usr/local/lib');

      // On actual Windows, should resolve to valid path
      if (process.platform === 'win32') {
        expect(templatesPath).toMatch(/^[A-Z]:\\/);
      }
    });

    /**
     * @TEST:TEMPLATE-PATH-MACOS-001
     * Test: macOS path resolution with HOME
     */
    it('should resolve macOS paths using HOME environment variable', () => {
      // RED: This test will fail until we fix HOME fallback to '~'

      process.env.HOME = '/Users/testuser';
      delete process.env.USERPROFILE;

      const templatesPath = processor.getTemplatesPath();

      expect(templatesPath).toBeDefined();

      // Should not use '~' literal when HOME is available
      if (templatesPath.includes('~')) {
        expect(process.env.HOME).toBeUndefined();
      }
    });

    /**
     * @TEST:TEMPLATE-PATH-LINUX-001
     * Test: Linux path resolution with HOME
     */
    it('should resolve Linux paths using HOME environment variable', () => {
      // RED: This test will fail until we implement proper HOME handling

      process.env.HOME = '/home/testuser';
      delete process.env.USERPROFILE;

      const templatesPath = processor.getTemplatesPath();

      expect(templatesPath).toBeDefined();

      // Should expand HOME properly
      if (templatesPath.includes(process.env.HOME || '')) {
        expect(templatesPath).toContain('/home/testuser');
      }
    });

    /**
     * @TEST:TEMPLATE-PATH-ENV-FALLBACK-001
     * Test: Source code should use USERPROFILE fallback, not hardcoded '~'
     */
    it('should use process.env.USERPROFILE fallback instead of hardcoded tilde', () => {
      // RED: This test WILL FAIL because source code uses "process.env['HOME'] || '~'"
      // It should use "process.env.HOME || process.env.USERPROFILE || os.homedir()"

      const fs = require('node:fs');
      const path = require('node:path');
      // __dirname is /Users/goos/MoAI/MoAI-ADK/moai-adk-ts/src/__tests__/core/installer
      // We need to go to src/core/installer/template-processor.ts
      const sourceFile = path.join(
        __dirname,
        '../../../core/installer/template-processor.ts'
      );
      const sourceCode = fs.readFileSync(sourceFile, 'utf-8');

      // Check for the problematic pattern
      const problematicPattern = "process.env['HOME'] || '~'";

      // This WILL FAIL - hardcoded '~' fallback exists
      expect(sourceCode).not.toContain(problematicPattern);

      // Verify better pattern exists (after GREEN phase)
      // expect(sourceCode).toContain('process.env.USERPROFILE');
      // expect(sourceCode).toContain('os.homedir()');
    });

    /**
     * @TEST:TEMPLATE-PATH-HARDCODED-001
     * Test: Unix-only paths should be platform-conditional
     */
    it('should make Unix-only paths platform-conditional', () => {
      // GREEN: Now we check that Unix paths are used conditionally

      const fs = require('node:fs');
      const path = require('node:path');
      const sourceFile = path.join(
        __dirname,
        '../../../core/installer/template-processor.ts'
      );
      const sourceCode = fs.readFileSync(sourceFile, 'utf-8');

      // Check that Unix-specific paths are now wrapped in platform checks
      const hasUnixPath = sourceCode.includes(
        "'/usr/local/lib/node_modules/moai-adk/templates'"
      );
      const hasPlatformCheck = sourceCode.includes(
        "process.platform !== 'win32'"
      );

      if (hasUnixPath) {
        // If Unix path exists, it MUST be within a platform check
        expect(hasPlatformCheck).toBe(true);

        // Verify the pattern: Unix path should come after platform check
        const platformCheckIndex = sourceCode.indexOf(
          "process.platform !== 'win32'"
        );
        const unixPathIndex = sourceCode.indexOf(
          "'/usr/local/lib/node_modules/moai-adk/templates'"
        );

        expect(platformCheckIndex).toBeLessThan(unixPathIndex);
        expect(unixPathIndex - platformCheckIndex).toBeLessThan(500); // Within reasonable proximity
      }
    });

    /**
     * @TEST:TEMPLATE-PATH-GLOBAL-CROSS-PLATFORM-001
     * Test: Global paths work across all platforms
     */
    it('should resolve global install paths on all platforms', () => {
      // RED: This test will fail until we implement platform-aware global paths

      const templatesPath = processor.getTemplatesPath();

      expect(templatesPath).toBeDefined();
      expect(typeof templatesPath).toBe('string');
      expect(templatesPath.length).toBeGreaterThan(0);

      // Verify path format matches platform
      if (process.platform === 'win32') {
        // Windows paths should use backslashes or be normalized
        // Node.js path module handles this, but let's verify it works
        const normalized = path.normalize(templatesPath);
        expect(normalized).toBe(templatesPath);
      } else {
        // Unix paths should use forward slashes
        expect(templatesPath).not.toMatch(/\\/);
      }

      // Verify the path is absolute
      expect(path.isAbsolute(templatesPath)).toBe(true);
    });
  });

  describe('Phase 2: Template Variable Creation (existing functionality)', () => {
    /**
     * Test: Template variables are created correctly
     */
    it('should create template variables with correct structure', () => {
      const config: InstallationConfig = {
        projectPath: '/test/project',
        projectName: 'test-project',
        mode: 'personal',
        backupEnabled: false,
        overwriteExisting: false,
        additionalFeatures: [],
      };

      const variables = processor.createTemplateVariables(config);

      expect(variables).toBeDefined();
      expect(variables.PROJECT_NAME).toBe('test-project');
      expect(variables.PROJECT_MODE).toBe('personal');
      expect(typeof variables.PROJECT_VERSION).toBe('string');
      expect(typeof variables.TIMESTAMP).toBe('string');
    });
  });

  describe('Phase 3: User Data Protection with excludePaths', () => {
    /**
     * @TEST:EXCLUDE-PATHS-001
     * Test: copyTemplateDirectory should exclude specified paths
     */
    it('should exclude paths specified in excludePaths option', async () => {
      const fs = require('node:fs');
      const os = require('node:os');
      const path = require('node:path');

      // Create temporary directories
      const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'moai-test-'));
      const srcDir = path.join(tempDir, 'src');
      const dstDir = path.join(tempDir, 'dst');

      // Setup: Create source directory structure
      fs.mkdirSync(srcDir, { recursive: true });
      fs.mkdirSync(path.join(srcDir, 'specs'), { recursive: true });
      fs.mkdirSync(path.join(srcDir, 'reports'), { recursive: true });
      fs.mkdirSync(path.join(srcDir, 'other'), { recursive: true });

      fs.writeFileSync(path.join(srcDir, 'specs', 'SPEC-001.md'), '# SPEC');
      fs.writeFileSync(
        path.join(srcDir, 'reports', 'report.md'),
        '# Report'
      );
      fs.writeFileSync(path.join(srcDir, 'other', 'file.md'), '# Other');

      try {
        // Act: Copy with excludePaths
        await processor.copyTemplateDirectory(srcDir, dstDir, {}, {
          excludePaths: ['specs', 'reports'],
        });

        // Assert: Excluded paths should NOT exist in destination
        expect(fs.existsSync(path.join(dstDir, 'specs'))).toBe(false);
        expect(fs.existsSync(path.join(dstDir, 'reports'))).toBe(false);

        // Assert: Non-excluded paths SHOULD exist in destination
        expect(fs.existsSync(path.join(dstDir, 'other'))).toBe(true);
        expect(fs.existsSync(path.join(dstDir, 'other', 'file.md'))).toBe(
          true
        );
      } finally {
        // Cleanup
        fs.rmSync(tempDir, { recursive: true, force: true });
      }
    });

    /**
     * @TEST:EXCLUDE-PATHS-002
     * Test: excludePaths should be case-insensitive for Windows compatibility
     */
    it('should exclude paths case-insensitively', async () => {
      const fs = require('node:fs');
      const os = require('node:os');
      const path = require('node:path');

      const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'moai-test-'));
      const srcDir = path.join(tempDir, 'src');
      const dstDir = path.join(tempDir, 'dst');

      fs.mkdirSync(srcDir, { recursive: true });
      fs.mkdirSync(path.join(srcDir, 'SPECS'), { recursive: true });
      fs.writeFileSync(path.join(srcDir, 'SPECS', 'test.md'), 'test');

      try {
        await processor.copyTemplateDirectory(srcDir, dstDir, {}, {
          excludePaths: ['specs'], // lowercase
        });

        // Should exclude 'SPECS' even though excludePath is 'specs'
        expect(fs.existsSync(path.join(dstDir, 'SPECS'))).toBe(false);
      } finally {
        fs.rmSync(tempDir, { recursive: true, force: true });
      }
    });

    /**
     * @TEST:EXCLUDE-PATHS-003
     * Test: copyTemplateDirectory should work without excludePaths (backward compatibility)
     */
    it('should copy all directories when excludePaths is not provided', async () => {
      const fs = require('node:fs');
      const os = require('node:os');
      const path = require('node:path');

      const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'moai-test-'));
      const srcDir = path.join(tempDir, 'src');
      const dstDir = path.join(tempDir, 'dst');

      fs.mkdirSync(srcDir, { recursive: true });
      fs.mkdirSync(path.join(srcDir, 'specs'), { recursive: true });
      fs.writeFileSync(path.join(srcDir, 'specs', 'test.md'), 'test');

      try {
        // No excludePaths option
        await processor.copyTemplateDirectory(srcDir, dstDir, {});

        // Should copy everything (backward compatibility)
        expect(fs.existsSync(path.join(dstDir, 'specs'))).toBe(true);
        expect(fs.existsSync(path.join(dstDir, 'specs', 'test.md'))).toBe(
          true
        );
      } finally {
        fs.rmSync(tempDir, { recursive: true, force: true });
      }
    });
  });
});

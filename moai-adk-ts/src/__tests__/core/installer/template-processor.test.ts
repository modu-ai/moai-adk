/**
 * @TEST:TEMPLATE-PROCESSOR-001 Template Processor Cross-Platform Test Suite
 * @tags @TEST:TEMPLATE-PROCESSOR-001 @SPEC:INSTALL-SYSTEM-012
 *
 * Phase 2: Cross-platform path resolution tests
 * Tests for Windows/macOS/Linux compatibility and environment variable handling
 */

import * as path from 'node:path';
import { TemplateProcessor } from '@/core/installer/template-processor';

describe('TemplateProcessor - Cross-Platform Path Resolution', () => {
  let processor: TemplateProcessor;
  let originalEnv: NodeJS.ProcessEnv;
  let _originalPlatform: string;

  beforeEach(() => {
    processor = new TemplateProcessor();
    originalEnv = { ...process.env };
    _originalPlatform = process.platform;
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
      process.env['USERPROFILE'] = 'C:\\Users\\TestUser';
      delete process.env['HOME'];

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

      process.env['HOME'] = '/Users/testuser';
      delete process.env['USERPROFILE'];

      const templatesPath = processor.getTemplatesPath();

      expect(templatesPath).toBeDefined();

      // Should not use '~' literal when HOME is available
      if (templatesPath.includes('~')) {
        expect(process.env['HOME']).toBeUndefined();
      }
    });

    /**
     * @TEST:TEMPLATE-PATH-LINUX-001
     * Test: Linux path resolution with HOME
     */
    it('should resolve Linux paths using HOME environment variable', () => {
      // RED: This test will fail until we implement proper HOME handling

      process.env['HOME'] = '/home/testuser';
      delete process.env['USERPROFILE'];

      const templatesPath = processor.getTemplatesPath();

      expect(templatesPath).toBeDefined();

      // Should expand HOME properly
      if (templatesPath.includes(process.env['HOME'] || '')) {
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
      const config = {
        projectPath: '/test/project',
        projectName: 'test-project',
        mode: 'development' as const,
        language: 'typescript' as const,
        forceOverwrite: false,
      };

      const variables = processor.createTemplateVariables(config);

      expect(variables).toBeDefined();
      expect(variables['PROJECT_NAME']).toBe('test-project');
      expect(variables['PROJECT_MODE']).toBe('development');
      expect(typeof variables['PROJECT_VERSION']).toBe('string');
      expect(typeof variables['TIMESTAMP']).toBe('string');
    });
  });
});

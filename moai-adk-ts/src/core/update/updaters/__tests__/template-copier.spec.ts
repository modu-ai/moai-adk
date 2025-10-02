// @TEST:UPDATE-REFACTOR-001 | SPEC: SPEC-UPDATE-REFACTOR-001.md
// Related: @CODE:UPD-TPL-001

/**
 * @file TemplateCopier Test Suite (Enhanced)
 * @description Test output-styles/alfred directory inclusion
 * @tags @TEST:UPDATE-REFACTOR-001-T004
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { promises as fs } from 'node:fs';
import * as path from 'node:path';
import { TemplateCopier } from '../template-copier.js';

describe('TemplateCopier', () => {
  const mockProjectPath = '/tmp/moai-test-copier-project';
  const mockTemplatePath = '/tmp/moai-test-copier-template';

  beforeEach(async () => {
    await fs.mkdir(mockProjectPath, { recursive: true });
    await fs.mkdir(mockTemplatePath, { recursive: true });
  });

  afterEach(async () => {
    try {
      await fs.rm(mockProjectPath, { recursive: true, force: true });
      await fs.rm(mockTemplatePath, { recursive: true, force: true });
    } catch {
      // Ignore cleanup errors
    }
  });

  describe('copyTemplates', () => {
    // @TEST:UPDATE-REFACTOR-001-T004
    it('should include output-styles/alfred directory', async () => {
      // Given: Output styles directory in template
      const styleDir = path.join(
        mockTemplatePath,
        '.claude/output-styles/alfred'
      );
      await fs.mkdir(styleDir, { recursive: true });
      await fs.writeFile(
        path.join(styleDir, 'alfred-pro.md'),
        '# MoAI Pro Style'
      );
      await fs.writeFile(
        path.join(styleDir, 'beginner-learning.md'),
        '# Beginner Learning Style'
      );

      // Also create required files to avoid errors
      const requiredFiles = [
        '.claude/commands/alfred/1-spec.md',
        '.claude/agents/alfred/spec-builder.md',
        '.claude/hooks/alfred/policy-block.cjs',
        '.moai/memory/development-guide.md',
        '.moai/project/product.md',
        '.moai/project/structure.md',
        '.moai/project/tech.md',
        'CLAUDE.md',
      ];

      for (const file of requiredFiles) {
        const filePath = path.join(mockTemplatePath, file);
        await fs.mkdir(path.dirname(filePath), { recursive: true });
        await fs.writeFile(filePath, `# ${file}`);
      }

      // When: TemplateCopier copies files
      const copier = new TemplateCopier(mockProjectPath);
      const count = await copier.copyTemplates(mockTemplatePath);

      // Then: Output styles should be copied
      const targetStyleDir = path.join(
        mockProjectPath,
        '.claude/output-styles/alfred'
      );
      const styleExists = await fs
        .access(targetStyleDir)
        .then(() => true)
        .catch(() => false);

      expect(styleExists).toBe(true);
      expect(count).toBeGreaterThan(0);

      // Verify style files exist
      const moaiProExists = await fs
        .access(path.join(targetStyleDir, 'alfred-pro.md'))
        .then(() => true)
        .catch(() => false);
      expect(moaiProExists).toBe(true);
    });

    it('should copy all expected directories', async () => {
      // Given: All template directories
      const directories = [
        '.claude/commands/alfred',
        '.claude/agents/alfred',
        '.claude/hooks/alfred',
        '.claude/output-styles/alfred',
      ];

      for (const dir of directories) {
        const dirPath = path.join(mockTemplatePath, dir);
        await fs.mkdir(dirPath, { recursive: true });
        await fs.writeFile(path.join(dirPath, 'test.md'), '# Test');
      }

      // Required files
      const files = [
        '.moai/memory/development-guide.md',
        '.moai/project/product.md',
        '.moai/project/structure.md',
        '.moai/project/tech.md',
        'CLAUDE.md',
      ];

      for (const file of files) {
        const filePath = path.join(mockTemplatePath, file);
        await fs.mkdir(path.dirname(filePath), { recursive: true });
        await fs.writeFile(filePath, `# ${file}`);
      }

      // When: Copy templates
      const copier = new TemplateCopier(mockProjectPath);
      await copier.copyTemplates(mockTemplatePath);

      // Then: All directories should exist
      for (const dir of directories) {
        const targetDir = path.join(mockProjectPath, dir);
        const exists = await fs
          .access(targetDir)
          .then(() => true)
          .catch(() => false);
        expect(exists).toBe(true);
      }
    });
  });
});

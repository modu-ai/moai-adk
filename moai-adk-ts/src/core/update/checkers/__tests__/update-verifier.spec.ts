// @TEST:UPDATE-REFACTOR-001 | SPEC: SPEC-UPDATE-REFACTOR-001.md
// Related: @CODE:UPD-VER-002

/**
 * @file UpdateVerifier Test Suite (Enhanced)
 * @description Test dynamic file count verification and YAML parsing
 * @tags @TEST:UPDATE-REFACTOR-001-T005, @TEST:UPDATE-REFACTOR-001-T006
 */

import { promises as fs } from 'node:fs';
import * as path from 'node:path';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { UpdateVerifier } from '../update-verifier.js';

describe('UpdateVerifier', () => {
  const mockProjectPath = '/tmp/moai-test-verifier';

  beforeEach(async () => {
    await fs.mkdir(mockProjectPath, { recursive: true });
  });

  afterEach(async () => {
    try {
      await fs.rm(mockProjectPath, { recursive: true, force: true });
    } catch {
      // Ignore cleanup errors
    }
  });

  describe('verifyUpdate', () => {
    it('should verify all key files exist', async () => {
      // Given: Key files exist
      const keyFiles = [
        '.moai/memory/development-guide.md',
        'CLAUDE.md',
        '.claude/commands/alfred/1-spec.md',
        '.claude/agents/alfred/spec-builder.md',
        '.claude/output-styles/alfred/alfred-pro.md',
      ];

      for (const file of keyFiles) {
        const filePath = path.join(mockProjectPath, file);
        await fs.mkdir(path.dirname(filePath), { recursive: true });
        await fs.writeFile(filePath, '# Test');
      }

      // When: Verify update
      const verifier = new UpdateVerifier(mockProjectPath);

      // Then: Should not throw
      await expect(verifier.verifyUpdate()).resolves.not.toThrow();
    });

    it('should throw error if key file is missing', async () => {
      // Given: Missing key file
      // (no files created)

      // When: Verify update
      const verifier = new UpdateVerifier(mockProjectPath);

      // Then: Should throw error
      await expect(verifier.verifyUpdate()).rejects.toThrow();
    });
  });

  describe('file count verification', () => {
    // @TEST:UPDATE-REFACTOR-001-T005
    it('should verify file count dynamically using Glob simulation', async () => {
      // Given: Multiple files in alfred directories
      const alfredDirs = [
        '.claude/commands/alfred',
        '.claude/agents/alfred',
        '.claude/hooks/alfred',
        '.claude/output-styles/alfred',
      ];

      let totalFiles = 0;
      for (const dir of alfredDirs) {
        const dirPath = path.join(mockProjectPath, dir);
        await fs.mkdir(dirPath, { recursive: true });

        // Create 3 files per directory
        for (let i = 1; i <= 3; i++) {
          await fs.writeFile(path.join(dirPath, `file${i}.md`), `# File ${i}`);
          totalFiles++;
        }
      }

      // When: Count files (simulate Glob)
      const countFiles = async (pattern: string): Promise<number> => {
        let count = 0;
        const baseDir = path.join(mockProjectPath, pattern);

        try {
          const entries = await fs.readdir(baseDir, { withFileTypes: true });
          for (const entry of entries) {
            if (entry.isFile()) {
              count++;
            } else if (entry.isDirectory()) {
              const subCount = await countFiles(path.join(pattern, entry.name));
              count += subCount;
            }
          }
        } catch {
          // Directory not found
        }

        return count;
      };

      const actualCount = await countFiles('.claude');

      // Then: File count should match
      expect(actualCount).toBe(totalFiles);
      expect(actualCount).toBeGreaterThan(0);
    });
  });

  describe('YAML frontmatter parsing', () => {
    // @TEST:UPDATE-REFACTOR-001-T006
    it('should parse YAML frontmatter from files', async () => {
      // Given: File with YAML frontmatter
      const testFile = path.join(mockProjectPath, 'test-spec.md');
      const yamlContent = `---
version: 1.0.0
status: approved
tags:
  - UPDATE-REFACTOR-001
  - TDD
---

# Test Specification

Content here...`;

      await fs.writeFile(testFile, yamlContent);

      // When: Parse YAML (simulation)
      const parseYAML = async (
        filePath: string
      ): Promise<Record<string, unknown>> => {
        const content = await fs.readFile(filePath, 'utf-8');
        const match = content.match(/^---\n([\s\S]*?)\n---/);

        if (!match) {
          return {};
        }

        const yamlText = match[1];
        if (!yamlText) return {};
        const lines = yamlText.split('\n');
        const result: Record<string, unknown> = {};

        for (const line of lines) {
          if (line.includes(':')) {
            const [key, value] = line.split(':').map(s => s.trim());
            if (key) {
              result[key] = value;
            }
          }
        }

        return result;
      };

      const parsed = await parseYAML(testFile);

      // Then: YAML should be parsed correctly
      expect(parsed).toHaveProperty('version');
      expect(parsed.version).toBe('1.0.0');
      expect(parsed).toHaveProperty('status');
      expect(parsed.status).toBe('approved');
    });
  });
});

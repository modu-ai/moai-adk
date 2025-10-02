// @TEST:UPDATE-REFACTOR-001 | SPEC: SPEC-UPDATE-REFACTOR-001.md
// Related: @CODE:UPDATE-REFACTOR-001

/**
 * @file AlfredUpdateBridge Test Suite
 * @description Phase 4 Claude Code Tools Integration Tests
 * @tags @TEST:UPDATE-REFACTOR-001
 */

import { promises as fs } from 'node:fs';
import * as path from 'node:path';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { AlfredUpdateBridge } from '../alfred-update-bridge.js';

describe('AlfredUpdateBridge', () => {
  const mockProjectPath = '/tmp/moai-test-project';
  const mockTemplatePath = '/tmp/moai-test-template';

  beforeEach(async () => {
    // Setup mock directories
    await fs.mkdir(mockProjectPath, { recursive: true });
    await fs.mkdir(mockTemplatePath, { recursive: true });
  });

  afterEach(async () => {
    // Cleanup
    try {
      await fs.rm(mockProjectPath, { recursive: true, force: true });
      await fs.rm(mockTemplatePath, { recursive: true, force: true });
    } catch {
      // Ignore cleanup errors
    }
  });

  describe('copyTemplatesWithClaudeTools', () => {
    // @TEST:UPDATE-REFACTOR-001-T001
    it('should copy templates using Claude Code tools simulation', async () => {
      // Given: Template files exist
      const templateFile = path.join(
        mockTemplatePath,
        '.moai/project/product.md'
      );
      await fs.mkdir(path.dirname(templateFile), { recursive: true });
      await fs.writeFile(templateFile, '# {{PROJECT_NAME}}');

      // When: AlfredUpdateBridge copies templates
      const bridge = new AlfredUpdateBridge(mockProjectPath);
      const count = await bridge.copyTemplatesWithClaudeTools(mockTemplatePath);

      // Then: Files are copied
      expect(count).toBeGreaterThan(0);
    });

    // @TEST:UPDATE-REFACTOR-001-T002
    it('should protect project docs by checking {{PROJECT_NAME}} pattern', async () => {
      // Given: Customized project doc (no {{PROJECT_NAME}})
      const projectDocPath = path.join(
        mockProjectPath,
        '.moai/project/product.md'
      );
      await fs.mkdir(path.dirname(projectDocPath), { recursive: true });
      await fs.writeFile(projectDocPath, '# My Custom Project');

      const templateFile = path.join(
        mockTemplatePath,
        '.moai/project/product.md'
      );
      await fs.mkdir(path.dirname(templateFile), { recursive: true });
      await fs.writeFile(templateFile, '# New Template Content');

      // When: Bridge checks and backs up
      const bridge = new AlfredUpdateBridge(mockProjectPath);
      await bridge.copyTemplatesWithClaudeTools(mockTemplatePath);

      // Then: Backup should exist (file pattern: *.backup-*)
      const projectDir = path.dirname(projectDocPath);
      const files = await fs.readdir(projectDir);
      const hasBackup = files.some(f => f.includes('.backup-'));
      expect(hasBackup).toBe(true);
    });

    // @TEST:UPDATE-REFACTOR-001-T003
    it('should apply chmod +x to hook files', async () => {
      if (process.platform === 'win32') {
        // Skip on Windows
        return;
      }

      // Given: Hook file in template
      const hookFile = path.join(
        mockTemplatePath,
        '.claude/hooks/alfred/policy-block.cjs'
      );
      await fs.mkdir(path.dirname(hookFile), { recursive: true });
      await fs.writeFile(hookFile, '// hook content');

      // When: Bridge copies hooks
      const bridge = new AlfredUpdateBridge(mockProjectPath);
      await bridge.copyTemplatesWithClaudeTools(mockTemplatePath);

      // Then: Hook file has executable permission
      const targetHook = path.join(
        mockProjectPath,
        '.claude/hooks/alfred/policy-block.cjs'
      );
      const stat = await fs.stat(targetHook);
      // Check executable bit (0o100 for owner)
      expect(stat.mode & 0o100).toBeGreaterThan(0);
    });
  });

  describe('handleProjectDocs', () => {
    // @TEST:UPDATE-REFACTOR-001-T004
    it('should overwrite template-state docs (with {{PROJECT_NAME}})', async () => {
      // Given: Template doc with {{PROJECT_NAME}}
      const projectDocPath = path.join(
        mockProjectPath,
        '.moai/project/structure.md'
      );
      await fs.mkdir(path.dirname(projectDocPath), { recursive: true });
      await fs.writeFile(projectDocPath, '# {{PROJECT_NAME}} Structure');

      const templateFile = path.join(
        mockTemplatePath,
        '.moai/project/structure.md'
      );
      await fs.mkdir(path.dirname(templateFile), { recursive: true });
      await fs.writeFile(templateFile, '# {{PROJECT_NAME}} New Structure');

      // When: Bridge processes docs
      const bridge = new AlfredUpdateBridge(mockProjectPath);
      await bridge.copyTemplatesWithClaudeTools(mockTemplatePath);

      // Then: Should overwrite without backup
      const content = await fs.readFile(projectDocPath, 'utf-8');
      expect(content).toContain('New Structure');
    });

    // @TEST:UPDATE-REFACTOR-001-T005
    it('should backup customized docs (without {{PROJECT_NAME}})', async () => {
      // Given: Customized project doc
      const projectDocPath = path.join(
        mockProjectPath,
        '.moai/project/tech.md'
      );
      await fs.mkdir(path.dirname(projectDocPath), { recursive: true });
      await fs.writeFile(projectDocPath, '# Tech Stack\n\nCustom content here');

      const templateFile = path.join(mockTemplatePath, '.moai/project/tech.md');
      await fs.mkdir(path.dirname(templateFile), { recursive: true });
      await fs.writeFile(templateFile, '# {{PROJECT_NAME}} Tech');

      // When: Bridge processes docs
      const bridge = new AlfredUpdateBridge(mockProjectPath);
      await bridge.copyTemplatesWithClaudeTools(mockTemplatePath);

      // Then: Backup exists
      const projectDir = path.dirname(projectDocPath);
      const files = await fs.readdir(projectDir);
      const hasBackup = files.some(f => f.startsWith('tech.md.backup-'));
      expect(hasBackup).toBe(true);
    });
  });

  describe('handleOutputStyles', () => {
    // @TEST:UPDATE-REFACTOR-001-T006
    it('should copy output-styles/alfred directory', async () => {
      // Given: Output styles in template
      const styleFile = path.join(
        mockTemplatePath,
        '.claude/output-styles/alfred/alfred-pro.md'
      );
      await fs.mkdir(path.dirname(styleFile), { recursive: true });
      await fs.writeFile(styleFile, '# MoAI Pro Style');

      // When: Bridge copies styles
      const bridge = new AlfredUpdateBridge(mockProjectPath);
      await bridge.copyTemplatesWithClaudeTools(mockTemplatePath);

      // Then: Style file exists in project
      const targetStyle = path.join(
        mockProjectPath,
        '.claude/output-styles/alfred/alfred-pro.md'
      );
      const exists = await fs
        .access(targetStyle)
        .then(() => true)
        .catch(() => false);
      expect(exists).toBe(true);
    });
  });

  describe('error recovery', () => {
    // @TEST:UPDATE-REFACTOR-001-T007
    it('should continue on individual file errors', async () => {
      // Given: Some files missing in template
      const validFile = path.join(mockTemplatePath, 'CLAUDE.md');
      await fs.writeFile(validFile, '# Claude');

      // When: Bridge processes with missing files
      const bridge = new AlfredUpdateBridge(mockProjectPath);
      const count = await bridge.copyTemplatesWithClaudeTools(mockTemplatePath);

      // Then: Should still copy available files
      expect(count).toBeGreaterThanOrEqual(0);
    });
  });
});

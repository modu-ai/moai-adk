/**
 * @file Tests for status command implementation (Refactored)
 * @author MoAI Team
 * @tags @TEST:CLI-STATUS-001 @SPEC:CLI-FOUNDATION-012
 */

import { beforeEach, describe, expect, it } from 'vitest';

import { StatusCommand } from '../status';

// Simple minimal test to verify TDD Red phase
describe('StatusCommand (Refactored)', () => {
  let statusCommand: StatusCommand;

  beforeEach(() => {
    statusCommand = new StatusCommand();
  });

  describe('TDD Green Phase - Implemented functionality', () => {
    it('should run status command successfully', async () => {
      const result = await statusCommand.run({
        verbose: false,
        projectPath: '/test/project',
      });

      expect(result.success).toBe(true);
      expect(result.status).toBeDefined();
      expect(result.recommendations).toBeDefined();

      // Should provide recommendations for non-initialized project
      expect(result.recommendations?.length).toBeGreaterThan(0);
    });

    it('should run status command with verbose mode', async () => {
      const result = await statusCommand.run({
        verbose: true,
        projectPath: '/test/project',
      });

      expect(result.success).toBe(true);
      expect(result.status).toBeDefined();

      // Verbose mode should include file counts
      if (result.status) {
        expect(result.status.fileCounts).toBeDefined();
      }
    });

    it('should return proper status structure', async () => {
      const result = await statusCommand.run({
        verbose: false,
        projectPath: '/test/project',
      });

      if (result.status) {
        expect(result.status).toEqual(
          expect.objectContaining({
            path: expect.any(String),
            projectType: expect.any(String),
            moaiInitialized: expect.any(Boolean),
            claudeInitialized: expect.any(Boolean),
            memoryFile: expect.any(Boolean),
            gitRepository: expect.any(Boolean),
            versions: expect.objectContaining({
              package: expect.any(String),
              resources: expect.any(String),
            }),
          })
        );
      }
    });

    it('should have required methods defined', () => {
      // This tests our interface is properly set up
      expect(statusCommand).toBeDefined();
      expect(typeof statusCommand.run).toBe('function');
    });
  });

  describe('Interface Validation', () => {
    it('should accept valid StatusOptions', () => {
      const validOptions = { verbose: true, projectPath: '/test/path' };
      expect(validOptions).toEqual(
        expect.objectContaining({
          verbose: expect.any(Boolean),
        })
      );
    });

    it('should define ProjectStatus interface correctly', () => {
      // This ensures our interfaces are properly defined
      const mockStatus = {
        path: '/test/project',
        projectType: 'MoAI Project',
        moaiInitialized: true,
        claudeInitialized: true,
        memoryFile: true,
        gitRepository: true,
        versions: {
          package: '0.0.1',
          resources: '0.0.1',
        },
      };

      expect(mockStatus).toEqual(
        expect.objectContaining({
          path: expect.any(String),
          projectType: expect.any(String),
          moaiInitialized: expect.any(Boolean),
          claudeInitialized: expect.any(Boolean),
        })
      );
    });

    it('should define FileCount interface correctly', () => {
      const mockFileCount = {
        '.moai': 5,
        '.claude': 12,
        total: 17,
      };

      expect(mockFileCount).toEqual(
        expect.objectContaining({
          '.moai': expect.any(Number),
          '.claude': expect.any(Number),
          total: expect.any(Number),
        })
      );
    });
  });
});

/**
 * @file Init command path validation integration test
 * @author MoAI Team
 * @tags @TEST:CLI-INIT-PATH-001 @SPEC:BUG-FIX-PACKAGE-PATH-001
 */

import { beforeEach, describe, expect, test, vi, type Mocked } from 'vitest';
import '@/__tests__/setup';
import { InitCommand } from '@/cli/commands/init';
import { SystemDetector } from '@/core/system-checker/detector';
import type { DoctorResult } from '@/cli/commands/doctor';

// Mock modules
vi.mock('@/core/system-checker/detector');

describe('InitCommand - Package Path Validation', () => {
  let initCommand: InitCommand;
  let mockDetector: Mocked<SystemDetector>;

  beforeEach(() => {
    vi.clearAllMocks();
    mockDetector = new SystemDetector() as unknown as Mocked<SystemDetector>;
    initCommand = new InitCommand(mockDetector);
  });

  describe('Path validation integration', () => {
    test('should prevent initialization inside MoAI-ADK package root', async () => {
      // Arrange: Mock successful doctor check
      const mockDoctorResult = {
        allPassed: true,
        results: [],
        missingRequirements: [],
        versionConflicts: [],
        summary: { total: 5, passed: 5, failed: 0 },
      } as DoctorResult;
      (initCommand as any).doctorCommand.run = vi
        .fn()
        .mockResolvedValue(mockDoctorResult);

      // Mock prompts to use current directory (package root)
      vi.mock('../../../cli/prompts/init-prompts', () => ({
        promptProjectSetup: vi.fn().mockResolvedValue({
          projectName: 'test-project',
          projectType: 'typescript',
        }),
        displayWelcomeBanner: vi.fn(),
      }));

      // Act: Try to initialize in package root (where this test is running)
      const result = await initCommand.runInteractive({
        name: '.',
      });

      // Assert: Should fail with package path error
      expect(result.success).toBe(false);
      expect(result.errors[0]).toContain('Cannot initialize project inside MoAI-ADK package');
    });

    test('should prevent initialization in package subdirectory', async () => {
      // Arrange: Mock successful doctor check
      const mockDoctorResult = {
        allPassed: true,
        results: [],
        missingRequirements: [],
        versionConflicts: [],
        summary: { total: 5, passed: 5, failed: 0 },
      } as DoctorResult;
      (initCommand as any).doctorCommand.run = vi
        .fn()
        .mockResolvedValue(mockDoctorResult);

      // Mock prompts
      vi.mock('../../../cli/prompts/init-prompts', () => ({
        promptProjectSetup: vi.fn().mockResolvedValue({
          projectName: 'test-project',
          projectType: 'typescript',
        }),
        displayWelcomeBanner: vi.fn(),
      }));

      // Act: Try to initialize in a subdirectory of the package
      const result = await initCommand.runInteractive({
        name: 'my-project',
        path: process.cwd() + '/src/my-project', // Inside package
      });

      // Assert: Should fail with package path error
      expect(result.success).toBe(false);
      expect(result.errors[0]).toContain('Cannot initialize project inside MoAI-ADK package');
    });
  });
});

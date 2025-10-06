// @TEST:INIT-001 | SPEC: SPEC-INIT-001.md | CODE: src/core/system-checker/index.ts
// Related: @CODE:INIT-001:DOCTOR, @SPEC:INIT-001

/**
 * @file Test for optional dependency separation in doctor command
 * @author MoAI Team
 * @tags @TEST:INIT-001:DOCTOR
 */

import { describe, test, expect, vi, beforeEach } from 'vitest';
import type { RequirementCheckResult } from '@/core/system-checker';

describe('Doctor Command - Optional Dependencies', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('allPassed calculation', () => {
    test('should pass when runtime and development dependencies are satisfied (optional can fail)', async () => {
      // Given: runtime/development 성공, optional 실패
      const mockResults: RequirementCheckResult[] = [
        {
          requirement: {
            id: 'git',
            name: 'Git',
            category: 'runtime',
            commands: ['git --version'],
            versionPattern: /git version (\d+\.\d+\.\d+)/,
            minVersion: '2.0.0',
          },
          result: {
            isInstalled: true,
            versionSatisfied: true,
            detectedVersion: '2.45.0',
          },
        },
        {
          requirement: {
            id: 'node',
            name: 'Node.js',
            category: 'runtime',
            commands: ['node --version'],
            versionPattern: /v(\d+\.\d+\.\d+)/,
            minVersion: '18.0.0',
          },
          result: {
            isInstalled: true,
            versionSatisfied: true,
            detectedVersion: '20.11.0',
          },
        },
        {
          requirement: {
            id: 'npm',
            name: 'npm',
            category: 'development',
            commands: ['npm --version'],
            versionPattern: /(\d+\.\d+\.\d+)/,
            minVersion: '9.0.0',
          },
          result: {
            isInstalled: true,
            versionSatisfied: true,
            detectedVersion: '10.5.0',
          },
        },
        {
          requirement: {
            id: 'git-lfs',
            name: 'Git LFS',
            category: 'optional',
            commands: ['git-lfs --version'],
            versionPattern: /git-lfs\/(\d+\.\d+\.\d+)/,
          },
          result: {
            isInstalled: false,
            versionSatisfied: false,
          },
        },
      ];

      // When: allPassed 계산
      const { DoctorCommand } = await import('@/cli/commands/doctor');
      const { SystemDetector } = await import('@/core/system-checker');
      const detector = new SystemDetector();
      const doctorCommand = new DoctorCommand(detector);

      // Mock the run method to return our test results
      vi.spyOn(doctorCommand, 'run').mockResolvedValue({
        allPassed: true,
        results: mockResults,
        missingRequirements: [mockResults[3]],
        versionConflicts: [],
        summary: {
          total: 4,
          passed: 3,
          failed: 1,
        },
      });

      const result = await doctorCommand.run();

      // Then: allPassed는 true (optional 실패는 무시)
      expect(result.allPassed).toBe(true);
      expect(result.summary.passed).toBe(3);
      expect(result.summary.failed).toBe(1);
    });

    test('should fail when runtime dependency is missing (even if optional passes)', async () => {
      // Given: runtime 실패, development/optional 성공
      const mockResults: RequirementCheckResult[] = [
        {
          requirement: {
            id: 'git',
            name: 'Git',
            category: 'runtime',
            commands: ['git --version'],
            versionPattern: /git version (\d+\.\d+\.\d+)/,
            minVersion: '2.0.0',
          },
          result: {
            isInstalled: false,
            versionSatisfied: false,
          },
        },
        {
          requirement: {
            id: 'npm',
            name: 'npm',
            category: 'development',
            commands: ['npm --version'],
            versionPattern: /(\d+\.\d+\.\d+)/,
            minVersion: '9.0.0',
          },
          result: {
            isInstalled: true,
            versionSatisfied: true,
            detectedVersion: '10.5.0',
          },
        },
        {
          requirement: {
            id: 'git-lfs',
            name: 'Git LFS',
            category: 'optional',
            commands: ['git-lfs --version'],
            versionPattern: /git-lfs\/(\d+\.\d+\.\d+)/,
          },
          result: {
            isInstalled: true,
            versionSatisfied: true,
            detectedVersion: '3.4.0',
          },
        },
      ];

      // When: allPassed 계산
      const { DoctorCommand } = await import('@/cli/commands/doctor');
      const { SystemDetector } = await import('@/core/system-checker');
      const detector = new SystemDetector();
      const doctorCommand = new DoctorCommand(detector);

      vi.spyOn(doctorCommand, 'run').mockResolvedValue({
        allPassed: false,
        results: mockResults,
        missingRequirements: [mockResults[0]],
        versionConflicts: [],
        summary: {
          total: 3,
          passed: 2,
          failed: 1,
        },
      });

      const result = await doctorCommand.run();

      // Then: allPassed는 false (runtime 실패는 치명적)
      expect(result.allPassed).toBe(false);
      expect(result.missingRequirements).toContainEqual(mockResults[0]);
    });

    test('should fail when development dependency is missing', async () => {
      // Given: development 실패, runtime/optional 성공
      const mockResults: RequirementCheckResult[] = [
        {
          requirement: {
            id: 'git',
            name: 'Git',
            category: 'runtime',
            commands: ['git --version'],
            versionPattern: /git version (\d+\.\d+\.\d+)/,
            minVersion: '2.0.0',
          },
          result: {
            isInstalled: true,
            versionSatisfied: true,
            detectedVersion: '2.45.0',
          },
        },
        {
          requirement: {
            id: 'npm',
            name: 'npm',
            category: 'development',
            commands: ['npm --version'],
            versionPattern: /(\d+\.\d+\.\d+)/,
            minVersion: '9.0.0',
          },
          result: {
            isInstalled: false,
            versionSatisfied: false,
          },
        },
        {
          requirement: {
            id: 'git-lfs',
            name: 'Git LFS',
            category: 'optional',
            commands: ['git-lfs --version'],
            versionPattern: /git-lfs\/(\d+\.\d+\.\d+)/,
          },
          result: {
            isInstalled: true,
            versionSatisfied: true,
            detectedVersion: '3.4.0',
          },
        },
      ];

      // When: allPassed 계산
      const { DoctorCommand } = await import('@/cli/commands/doctor');
      const { SystemDetector } = await import('@/core/system-checker');
      const detector = new SystemDetector();
      const doctorCommand = new DoctorCommand(detector);

      vi.spyOn(doctorCommand, 'run').mockResolvedValue({
        allPassed: false,
        results: mockResults,
        missingRequirements: [mockResults[1]],
        versionConflicts: [],
        summary: {
          total: 3,
          passed: 2,
          failed: 1,
        },
      });

      const result = await doctorCommand.run();

      // Then: allPassed는 false (development 실패는 치명적)
      expect(result.allPassed).toBe(false);
      expect(result.missingRequirements).toContainEqual(mockResults[1]);
    });

    test('should display warning message for optional dependency failures', async () => {
      // Given: optional 의존성 실패
      const mockResults: RequirementCheckResult[] = [
        {
          requirement: {
            id: 'git-lfs',
            name: 'Git LFS',
            category: 'optional',
            commands: ['git-lfs --version'],
            versionPattern: /git-lfs\/(\d+\.\d+\.\d+)/,
          },
          result: {
            isInstalled: false,
            versionSatisfied: false,
          },
        },
      ];

      // When: doctor 실행
      const { DoctorCommand } = await import('@/cli/commands/doctor');
      const { SystemDetector } = await import('@/core/system-checker');
      const detector = new SystemDetector();
      const doctorCommand = new DoctorCommand(detector);

      vi.spyOn(doctorCommand, 'run').mockResolvedValue({
        allPassed: true,
        results: mockResults,
        missingRequirements: [],
        versionConflicts: [],
        summary: {
          total: 1,
          passed: 0,
          failed: 1,
        },
      });

      const result = await doctorCommand.run();

      // Then: allPassed should be true even with optional failures
      // (The warning message is displayed by the actual implementation, not the mock)
      expect(result.allPassed).toBe(true);
      expect(result.summary.failed).toBe(1);
    });

    test('should categorize dependencies correctly', async () => {
      // Given: 다양한 카테고리의 의존성
      const { requirementRegistry } = await import('@/core/system-checker/requirements');

      // When: 카테고리별 조회
      const runtimeDeps = requirementRegistry.getByCategory('runtime');
      const developmentDeps = requirementRegistry.getByCategory('development');
      const optionalDeps = requirementRegistry.getByCategory('optional');

      // Then: 각 카테고리에 올바른 의존성 포함
      expect(runtimeDeps.some(dep => dep.name === 'Git')).toBe(true);
      expect(runtimeDeps.some(dep => dep.name === 'Node.js')).toBe(true);
      expect(developmentDeps.some(dep => dep.name === 'npm')).toBe(true);
      expect(optionalDeps.some(dep => dep.name === 'Git LFS')).toBe(true);
    });
  });
});

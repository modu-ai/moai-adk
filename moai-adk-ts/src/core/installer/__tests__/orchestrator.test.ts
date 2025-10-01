// @TEST:INST-001 연결: @CODE:INST-001 -> @TEST:INST-001
// Related: @TEST:PHASE-EXECUTOR-001, @TEST:CONTEXT-MANAGER-001

/**
 * @file Installation orchestrator test suite (TDD GREEN phase)
 * @author MoAI Team
 * @tags @TEST:ORCHESTRATOR-001
 */

import { beforeEach, describe, expect, it } from 'vitest';
import { InstallationOrchestrator } from '../orchestrator';
import type { InstallationConfig } from '../types';

describe('@TEST:ORCHESTRATOR-001: InstallationOrchestrator', () => {
  let testConfig: InstallationConfig;
  let orchestrator: InstallationOrchestrator;

  beforeEach(() => {
    testConfig = {
      projectPath: '/test/project',
      projectName: 'test-project',
      mode: 'personal',
      backupEnabled: false,
      overwriteExisting: true,
      additionalFeatures: [],
    };

    orchestrator = new InstallationOrchestrator(testConfig);
  });

  describe('@TEST:ORCHESTRATOR-INIT: Constructor Initialization', () => {
    it('@TEST:ORCHESTRATOR-INIT-001: should initialize with valid configuration', () => {
      expect(orchestrator).toBeDefined();
      expect(orchestrator.getConfig()).toEqual(testConfig);
    });

    it('@TEST:ORCHESTRATOR-INIT-002: should create initial context', () => {
      const context = orchestrator.getContext();
      expect(context).toBeDefined();
      expect(context.config).toEqual(testConfig);
      expect(context.phases).toEqual([]);
      expect(context.allFilesCreated).toEqual([]);
      expect(context.allErrors).toEqual([]);
    });

    it('@TEST:ORCHESTRATOR-INIT-003: should handle personal mode', () => {
      const config = { ...testConfig, mode: 'personal' as const };
      const orch = new InstallationOrchestrator(config);
      expect(orch.getConfig().mode).toBe('personal');
    });

    it('@TEST:ORCHESTRATOR-INIT-004: should handle team mode', () => {
      const config = { ...testConfig, mode: 'team' as const };
      const orch = new InstallationOrchestrator(config);
      expect(orch.getConfig().mode).toBe('team');
    });

    it('@TEST:ORCHESTRATOR-INIT-005: should handle backup enabled', () => {
      const config = { ...testConfig, backupEnabled: true };
      const orch = new InstallationOrchestrator(config);
      expect(orch.getConfig().backupEnabled).toBe(true);
    });
  });

  describe('@TEST:ORCHESTRATOR-CONFIG: Configuration Management', () => {
    it('@TEST:ORCHESTRATOR-CONFIG-001: should return immutable configuration', () => {
      const config1 = orchestrator.getConfig();
      const config2 = orchestrator.getConfig();
      expect(config1).toEqual(config2);
    });

    it('@TEST:ORCHESTRATOR-CONFIG-002: should preserve project path', () => {
      const config = orchestrator.getConfig();
      expect(config.projectPath).toBe('/test/project');
    });

    it('@TEST:ORCHESTRATOR-CONFIG-003: should preserve project name', () => {
      const config = orchestrator.getConfig();
      expect(config.projectName).toBe('test-project');
    });

    it('@TEST:ORCHESTRATOR-CONFIG-004: should preserve additional features', () => {
      const configWithFeatures = {
        ...testConfig,
        additionalFeatures: ['feature1', 'feature2'],
      };
      const orch = new InstallationOrchestrator(configWithFeatures);
      expect(orch.getConfig().additionalFeatures).toEqual([
        'feature1',
        'feature2',
      ]);
    });

    it('@TEST:ORCHESTRATOR-CONFIG-005: should handle empty additional features', () => {
      const config = { ...testConfig, additionalFeatures: [] };
      const orch = new InstallationOrchestrator(config);
      expect(orch.getConfig().additionalFeatures).toEqual([]);
    });

    it('@TEST:ORCHESTRATOR-CONFIG-006: should handle undefined template path', () => {
      const config = { ...testConfig, templatePath: undefined };
      const orch = new InstallationOrchestrator(config);
      expect(orch.getConfig().templatePath).toBeUndefined();
    });

    it('@TEST:ORCHESTRATOR-CONFIG-007: should handle overwrite existing flag', () => {
      const config = { ...testConfig, overwriteExisting: false };
      const orch = new InstallationOrchestrator(config);
      expect(orch.getConfig().overwriteExisting).toBe(false);
    });
  });

  describe('@TEST:ORCHESTRATOR-CONTEXT: Context Management', () => {
    it('@TEST:ORCHESTRATOR-CONTEXT-001: should maintain context state', () => {
      const context = orchestrator.getContext();
      expect(context.startTime).toBeInstanceOf(Date);
      expect(context.phases).toBeInstanceOf(Array);
      expect(context.allFilesCreated).toBeInstanceOf(Array);
      expect(context.allErrors).toBeInstanceOf(Array);
    });

    it('@TEST:ORCHESTRATOR-CONTEXT-002: should have valid start time', () => {
      const context = orchestrator.getContext();
      const now = new Date();
      const timeDiff = Math.abs(now.getTime() - context.startTime.getTime());
      expect(timeDiff).toBeLessThan(1000); // Within 1 second
    });

    it('@TEST:ORCHESTRATOR-CONTEXT-003: should initialize empty phases', () => {
      const context = orchestrator.getContext();
      expect(context.phases).toHaveLength(0);
    });

    it('@TEST:ORCHESTRATOR-CONTEXT-004: should initialize empty files list', () => {
      const context = orchestrator.getContext();
      expect(context.allFilesCreated).toHaveLength(0);
    });

    it('@TEST:ORCHESTRATOR-CONTEXT-005: should initialize empty errors list', () => {
      const context = orchestrator.getContext();
      expect(context.allErrors).toHaveLength(0);
    });

    it('@TEST:ORCHESTRATOR-CONTEXT-006: should maintain context reference', () => {
      const context1 = orchestrator.getContext();
      const context2 = orchestrator.getContext();
      expect(context1).toBe(context2);
    });

    it('@TEST:ORCHESTRATOR-CONTEXT-007: should have context with correct config', () => {
      const context = orchestrator.getContext();
      expect(context.config).toBe(testConfig);
    });
  });

  describe('@TEST:ORCHESTRATOR-EXEC: Installation Execution - Error Handling', () => {
    it('@TEST:ORCHESTRATOR-ERROR-001: should return failure result on error', async () => {
      // Invalid path will cause validation error
      const config = { ...testConfig, projectPath: '' };
      const orch = new InstallationOrchestrator(config);
      const result = await orch.executeInstallation();

      expect(result.success).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
    });

    it('@TEST:ORCHESTRATOR-ERROR-002: should include error messages in result', async () => {
      const config = { ...testConfig, projectPath: '' };
      const orch = new InstallationOrchestrator(config);
      const result = await orch.executeInstallation();

      expect(result.errors).toBeDefined();
      expect(Array.isArray(result.errors)).toBe(true);
      expect(result.errors.some(e => e.includes('failed'))).toBe(true);
    });

    it('@TEST:ORCHESTRATOR-ERROR-003: should return proper error structure', async () => {
      const config = { ...testConfig, projectPath: '' };
      const orch = new InstallationOrchestrator(config);
      const result = await orch.executeInstallation();

      expect(result).toHaveProperty('success');
      expect(result).toHaveProperty('projectPath');
      expect(result).toHaveProperty('filesCreated');
      expect(result).toHaveProperty('errors');
      expect(result).toHaveProperty('nextSteps');
      expect(result).toHaveProperty('config');
      expect(result).toHaveProperty('timestamp');
      expect(result).toHaveProperty('duration');
    });
  });

  describe('@TEST:ORCHESTRATOR-RESULT: Result Building', () => {
    it('@TEST:ORCHESTRATOR-RESULT-001: should include timestamp in result', async () => {
      const config = { ...testConfig, projectPath: '' };
      const orch = new InstallationOrchestrator(config);
      const result = await orch.executeInstallation();

      expect(result.timestamp).toBeInstanceOf(Date);
    });

    it('@TEST:ORCHESTRATOR-RESULT-002: should calculate duration', async () => {
      const config = { ...testConfig, projectPath: '' };
      const orch = new InstallationOrchestrator(config);
      const result = await orch.executeInstallation();

      expect(result.duration).toBeGreaterThanOrEqual(0);
      expect(typeof result.duration).toBe('number');
    });

    it('@TEST:ORCHESTRATOR-RESULT-003: should include configuration in result', async () => {
      const config = { ...testConfig, projectPath: '' };
      const orch = new InstallationOrchestrator(config);
      const result = await orch.executeInstallation();

      expect(result.config).toEqual(config);
    });

    it('@TEST:ORCHESTRATOR-RESULT-004: should include next steps', async () => {
      const config = { ...testConfig, projectPath: '' };
      const orch = new InstallationOrchestrator(config);
      const result = await orch.executeInstallation();

      expect(result.nextSteps).toBeDefined();
      expect(Array.isArray(result.nextSteps)).toBe(true);
    });

    it('@TEST:ORCHESTRATOR-RESULT-005: should list created files', async () => {
      const config = { ...testConfig, projectPath: '' };
      const orch = new InstallationOrchestrator(config);
      const result = await orch.executeInstallation();

      expect(result.filesCreated).toBeDefined();
      expect(Array.isArray(result.filesCreated)).toBe(true);
    });

    it('@TEST:ORCHESTRATOR-RESULT-006: should have valid duration', async () => {
      const config = { ...testConfig, projectPath: '' };
      const orch = new InstallationOrchestrator(config);
      const result = await orch.executeInstallation();

      expect(result.duration).toBeGreaterThanOrEqual(0);
      expect(result.duration).toBeLessThan(10000); // Less than 10 seconds
    });
  });

  describe('@TEST:ORCHESTRATOR-PROGRESS: Progress Reporting', () => {
    it('@TEST:ORCHESTRATOR-PROGRESS-001: should call progress callback', async () => {
      let callCount = 0;
      const progressCallback = () => {
        callCount++;
      };

      const config = { ...testConfig, projectPath: '' };
      const orch = new InstallationOrchestrator(config);
      await orch.executeInstallation(progressCallback);

      expect(callCount).toBeGreaterThanOrEqual(1);
    });

    it('@TEST:ORCHESTRATOR-PROGRESS-002: should provide meaningful progress messages', async () => {
      const messages: string[] = [];
      const progressCallback = (message: string) => {
        messages.push(message);
      };

      const config = { ...testConfig, projectPath: '' };
      const orch = new InstallationOrchestrator(config);
      await orch.executeInstallation(progressCallback);

      expect(messages.length).toBeGreaterThan(0);
      expect(messages.some(m => m.includes('Phase'))).toBe(true);
    });

    it('@TEST:ORCHESTRATOR-PROGRESS-003: should provide progress numbers', async () => {
      const progressData: Array<{ current: number; total: number }> = [];
      const progressCallback = (
        _message: string,
        current: number,
        total: number
      ) => {
        progressData.push({ current, total });
      };

      const config = { ...testConfig, projectPath: '' };
      const orch = new InstallationOrchestrator(config);
      await orch.executeInstallation(progressCallback);

      expect(progressData.length).toBeGreaterThan(0);
      expect(progressData.every(p => p.total > 0)).toBe(true);
    });
  });

  describe('@TEST:ORCHESTRATOR-EDGE: Edge Cases', () => {
    it('@TEST:ORCHESTRATOR-EDGE-001: should handle various project paths', () => {
      const paths = ['/path1', '/path2', '/nested/deep/path'];
      for (const projectPath of paths) {
        const config = { ...testConfig, projectPath };
        const orch = new InstallationOrchestrator(config);
        expect(orch.getConfig().projectPath).toBe(projectPath);
      }
    });

    it('@TEST:ORCHESTRATOR-EDGE-002: should handle various project names', () => {
      const names = ['project1', 'my-app', 'test_project'];
      for (const projectName of names) {
        const config = { ...testConfig, projectName };
        const orch = new InstallationOrchestrator(config);
        expect(orch.getConfig().projectName).toBe(projectName);
      }
    });

    it('@TEST:ORCHESTRATOR-EDGE-003: should handle both personal and team modes', () => {
      const modes: Array<'personal' | 'team'> = ['personal', 'team'];
      for (const mode of modes) {
        const config = { ...testConfig, mode };
        const orch = new InstallationOrchestrator(config);
        expect(orch.getConfig().mode).toBe(mode);
      }
    });

    it('@TEST:ORCHESTRATOR-EDGE-004: should handle backup flags', () => {
      for (const backupEnabled of [true, false]) {
        const config = { ...testConfig, backupEnabled };
        const orch = new InstallationOrchestrator(config);
        expect(orch.getConfig().backupEnabled).toBe(backupEnabled);
      }
    });

    it('@TEST:ORCHESTRATOR-EDGE-005: should handle overwrite flags', () => {
      for (const overwriteExisting of [true, false]) {
        const config = { ...testConfig, overwriteExisting };
        const orch = new InstallationOrchestrator(config);
        expect(orch.getConfig().overwriteExisting).toBe(overwriteExisting);
      }
    });
  });
});

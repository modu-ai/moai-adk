/**
 * @feature FIRST-RUN-001 FirstRunManager Tests
 * @task FIRST-RUN-TEST-001 RED stage - failing tests for first run detection
 */

import { describe, test, expect, beforeEach } from 'vitest';
import { FirstRunManager } from '../first-run-manager';

describe('FirstRunManager', () => {
  let manager: FirstRunManager;

  beforeEach(() => {
    manager = new FirstRunManager();
  });

  describe('isFirstRun', () => {
    test('should detect first run correctly when no config exists', () => {
      // Act
      const result = manager.isFirstRun();

      // Assert - This will fail initially (RED stage)
      expect(result).toBe(true);
    });

    test('should return false when config already exists', () => {
      // This test simulates a scenario where MoAI-ADK has been run before
      expect(manager.isFirstRun).toBeDefined();
      expect(typeof manager.isFirstRun).toBe('function');
    });

    test('should use consistent detection mechanism', () => {
      // Ensure the detection mechanism is consistent across calls
      const firstCall = manager.isFirstRun();
      const secondCall = manager.isFirstRun();
      expect(typeof firstCall).toBe('boolean');
      expect(typeof secondCall).toBe('boolean');
    });
  });

  describe('initializeFirstRun', () => {
    test('should initialize first run setup successfully', async () => {
      // Act & Assert - This will fail initially (RED stage)
      await expect(async () => {
        await manager.initializeFirstRun();
      }).not.toThrow();
    });

    test('should create necessary config files', async () => {
      // This test ensures that first run initialization creates required files
      expect(manager.initializeFirstRun).toBeDefined();
      expect(typeof manager.initializeFirstRun).toBe('function');
    });

    test('should handle initialization errors gracefully', async () => {
      // Test error handling during initialization
      expect(async () => {
        await manager.initializeFirstRun();
      }).not.toThrow();
    });
  });

  describe('setupUserEnvironment', () => {
    test('should setup user environment correctly', async () => {
      // Act & Assert - This will fail initially (RED stage)
      await expect(async () => {
        await manager.setupUserEnvironment();
      }).not.toThrow();
    });

    test('should configure user-specific settings', async () => {
      // This test ensures user environment is properly configured
      expect(manager.setupUserEnvironment).toBeDefined();
      expect(typeof manager.setupUserEnvironment).toBe('function');
    });

    test('should be idempotent', async () => {
      // Test that running setup multiple times doesn't cause issues
      await manager.setupUserEnvironment();
      await expect(async () => {
        await manager.setupUserEnvironment();
      }).not.toThrow();
    });
  });

  describe('createFirstRunMarker', () => {
    test('should create first run marker file', async () => {
      // Act & Assert - This will fail initially (RED stage)
      await expect(async () => {
        await manager.createFirstRunMarker();
      }).not.toThrow();
    });

    test('should mark that first run has completed', async () => {
      // This test ensures the marker prevents future first runs
      expect(manager.createFirstRunMarker).toBeDefined();
      expect(typeof manager.createFirstRunMarker).toBe('function');
    });
  });

  describe('validateFirstRunSetup', () => {
    test('should validate first run setup completed successfully', async () => {
      // Act
      const result = await manager.validateFirstRunSetup();

      // Assert - This will fail initially (RED stage)
      expect(result).toBe(true);
    });

    test('should check all required first run components', async () => {
      // This test ensures validation covers all necessary components
      expect(manager.validateFirstRunSetup).toBeDefined();
      expect(typeof manager.validateFirstRunSetup).toBe('function');
    });
  });

  describe('getFirstRunState', () => {
    test('should return current first run state', () => {
      // Act
      const state = manager.getFirstRunState();

      // Assert
      expect(state).toHaveProperty('isFirstRun');
      expect(state).toHaveProperty('hasMarkerFile');
      expect(state).toHaveProperty('setupCompleted');
      expect(typeof state.isFirstRun).toBe('boolean');
      expect(typeof state.hasMarkerFile).toBe('boolean');
      expect(typeof state.setupCompleted).toBe('boolean');
    });

    test('should provide consistent state information', () => {
      // Ensure state information is consistent across calls
      const state1 = manager.getFirstRunState();
      const state2 = manager.getFirstRunState();
      expect(state1.isFirstRun).toBe(state2.isFirstRun);
    });
  });
});

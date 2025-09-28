/**
 * @feature POST-INSTALL-001 PostInstallManager Tests
 * @task POST-INSTALL-TEST-001 RED stage - failing tests for PostInstall functionality
 */

import { jest, vi } from 'vitest';
import { PostInstallManager } from '../post-install-manager';
import type { ResourceValidator } from '../resource-validator';
import type { FirstRunManager } from '../first-run-manager';
import type { GlobalSetupManager } from '../global-setup-manager';
import type { PostInstallOptions } from '../../types';

// Mock dependencies will be set up in beforeEach

describe('PostInstallManager', () => {
  let manager: PostInstallManager;
  let mockResourceValidator: vi.Mocked<ResourceValidator>;
  let mockFirstRunManager: vi.Mocked<FirstRunManager>;
  let mockGlobalSetupManager: vi.Mocked<GlobalSetupManager>;

  beforeEach(() => {
    // Reset mocks
    vi.clearAllMocks();

    // Create mock instances with proper mock methods
    mockResourceValidator = {
      validatePackageResources: vi.fn(),
      checkRequiredTemplates: vi.fn(),
      verifyResourceIntegrity: vi.fn(),
      validatePackageStructure: vi.fn(),
      checkTemplateIntegrity: vi.fn(),
    } as any;

    mockFirstRunManager = {
      isFirstRun: vi.fn(),
      initializeFirstRun: vi.fn(),
      setupUserEnvironment: vi.fn(),
      createFirstRunMarker: vi.fn(),
      validateFirstRunSetup: vi.fn(),
      getFirstRunState: vi.fn(),
    } as any;

    mockGlobalSetupManager = {
      setupGlobalConfig: vi.fn(),
      registerGlobalCommands: vi.fn(),
      updatePathEnvironment: vi.fn(),
      setupGlobalResources: vi.fn(),
      validateGlobalSetup: vi.fn(),
      cleanupGlobalSetup: vi.fn(),
    } as any;

    // Initialize manager
    manager = new PostInstallManager();

    // Override dependencies with mocks
    (manager as any).resourceValidator = mockResourceValidator;
    (manager as any).firstRunManager = mockFirstRunManager;
    (manager as any).globalSetupManager = mockGlobalSetupManager;
  });

  describe('runPostInstall', () => {
    const validOptions: PostInstallOptions = {
      projectPath: '/test/project',
      setupGlobal: true,
      validateResources: true,
      force: false,
      quiet: false,
    };

    test('should successfully complete post-install process', async () => {
      // Arrange
      mockResourceValidator.validatePackageResources.mockResolvedValue({
        isValid: true,
        missingTemplates: [],
        errors: [],
      });
      mockFirstRunManager.isFirstRun.mockReturnValue(true);
      mockFirstRunManager.initializeFirstRun.mockResolvedValue();
      mockGlobalSetupManager.setupGlobalConfig.mockResolvedValue();

      // Act
      const result = await manager.runPostInstall(validOptions);

      // Assert
      expect(result.success).toBe(true);
      expect(result.resourcesValidated).toBe(true);
      expect(result.globalSetupCompleted).toBe(true);
      expect(result.firstRunSetupCompleted).toBe(true);
      expect(mockResourceValidator.validatePackageResources).toHaveBeenCalled();
      expect(mockFirstRunManager.initializeFirstRun).toHaveBeenCalled();
      expect(mockGlobalSetupManager.setupGlobalConfig).toHaveBeenCalled();
    });

    test('should detect first run correctly', async () => {
      // Arrange
      mockFirstRunManager.isFirstRun.mockReturnValue(true);
      mockResourceValidator.validatePackageResources.mockResolvedValue({
        isValid: true,
        missingTemplates: [],
        errors: [],
      });

      // Act
      const result = await manager.runPostInstall(validOptions);

      // Assert
      expect(result.isFirstRun).toBe(true);
      expect(mockFirstRunManager.isFirstRun).toHaveBeenCalled();
    });

    test('should handle resource validation failure', async () => {
      // Arrange
      mockResourceValidator.validatePackageResources.mockResolvedValue({
        isValid: false,
        missingTemplates: ['.claude', '.moai'],
        errors: ['Template validation failed'],
      });

      // Act
      const result = await manager.runPostInstall(validOptions);

      // Assert
      expect(result.success).toBe(false);
      expect(result.resourcesValidated).toBe(false);
      expect(result.errors).toContain('Resource validation failed');
    });

    test('should skip validation when validateResources is false', async () => {
      // Arrange
      const options = { ...validOptions, validateResources: false };

      // Act
      const result = await manager.runPostInstall(options);

      // Assert
      expect(
        mockResourceValidator.validatePackageResources
      ).not.toHaveBeenCalled();
      expect(result.resourcesValidated).toBe(true); // Should be true when skipped
    });

    test('should skip global setup when setupGlobal is false', async () => {
      // Arrange
      const options = { ...validOptions, setupGlobal: false };
      mockResourceValidator.validatePackageResources.mockResolvedValue({
        isValid: true,
        missingTemplates: [],
        errors: [],
      });

      // Act
      const result = await manager.runPostInstall(options);

      // Assert
      expect(mockGlobalSetupManager.setupGlobalConfig).not.toHaveBeenCalled();
      expect(result.globalSetupCompleted).toBe(true); // Should be true when skipped
    });
  });

  describe('validateResources', () => {
    test('should validate all required resources', async () => {
      // Arrange
      mockResourceValidator.validatePackageResources.mockResolvedValue({
        isValid: true,
        missingTemplates: [],
        errors: [],
      });

      // Act
      const result = await manager.validateResources();

      // Assert
      expect(result).toBe(true);
      expect(mockResourceValidator.validatePackageResources).toHaveBeenCalled();
    });

    test('should return false when resources are missing', async () => {
      // Arrange
      mockResourceValidator.validatePackageResources.mockResolvedValue({
        isValid: false,
        missingTemplates: ['.claude'],
        errors: [],
      });

      // Act
      const result = await manager.validateResources();

      // Assert
      expect(result).toBe(false);
    });
  });

  describe('autoInstallOnFirstRun', () => {
    test('should return true when all required templates exist', async () => {
      // Arrange
      mockResourceValidator.checkRequiredTemplates.mockResolvedValue({
        allPresent: true,
        missingTemplates: [],
        templateCount: 5,
      });

      // Act
      const result = await manager.autoInstallOnFirstRun();

      // Assert
      expect(result).toBe(true);
      expect(mockResourceValidator.checkRequiredTemplates).toHaveBeenCalled();
    });

    test('should return false when required templates are missing', async () => {
      // Arrange
      mockResourceValidator.checkRequiredTemplates.mockResolvedValue({
        allPresent: false,
        missingTemplates: ['.claude', '.moai'],
        templateCount: 3,
      });

      // Act
      const result = await manager.autoInstallOnFirstRun();

      // Assert
      expect(result).toBe(false);
    });

    test('should handle validation errors gracefully', async () => {
      // Arrange
      mockResourceValidator.checkRequiredTemplates.mockRejectedValue(
        new Error('Resource check failed')
      );

      // Act
      const result = await manager.autoInstallOnFirstRun();

      // Assert
      expect(result).toBe(false);
    });
  });

  describe('setupGlobalResources', () => {
    test('should setup global resources successfully', async () => {
      // Arrange
      mockGlobalSetupManager.setupGlobalConfig.mockResolvedValue();
      mockGlobalSetupManager.registerGlobalCommands.mockResolvedValue();

      // Act & Assert - Should not throw when setting up global resources
      await expect(async () => {
        await manager.setupGlobalResources();
      }).not.toThrow();
      expect(mockGlobalSetupManager.setupGlobalConfig).toHaveBeenCalled();
      expect(mockGlobalSetupManager.registerGlobalCommands).toHaveBeenCalled();
    });

    test('should handle global setup errors', async () => {
      // Arrange
      mockGlobalSetupManager.setupGlobalConfig.mockRejectedValue(
        new Error('Global setup failed')
      );

      // Act & Assert
      await expect(manager.setupGlobalResources()).rejects.toThrow(
        'Global setup failed'
      );
    });
  });

  describe('printPostInstallBanner', () => {
    test('should print banner with version information', () => {
      // Arrange
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
      const version = '1.0.0';

      // Act
      manager.printPostInstallBanner(version);

      // Assert
      expect(consoleSpy).toHaveBeenCalled();
      const output = consoleSpy.mock.calls.join('\n');
      expect(output).toContain('MoAI-ADK');
      expect(output).toContain(version);

      consoleSpy.mockRestore();
    });

    test('should not print banner when quiet mode is enabled', () => {
      // Arrange
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      // Act
      manager.printPostInstallBanner('1.0.0', true);

      // Assert
      expect(consoleSpy).not.toHaveBeenCalled();

      consoleSpy.mockRestore();
    });
  });
});

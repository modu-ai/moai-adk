/**
 * @file ConfigManager test suite - RED phase TDD
 * @author MoAI Team
 * @tags @TEST:CONFIG-MANAGER-001 @SPEC:CORE-SYSTEM-013
 */

import fs from 'node:fs';
import { beforeEach, describe, expect, vi, type Mocked } from 'vitest';
import { ConfigManager } from '../config-manager';

// Mock fs and path modules
vi.mock('fs');
const mockFs = fs as Mocked<typeof fs>;

describe('ConfigManager', () => {
  let configManager: ConfigManager;
  let tempDir: string;
  let mockConfig: any;

  beforeEach(() => {
    configManager = new ConfigManager();
    tempDir = '/test/project';
    mockConfig = {
      projectName: 'test-project',
      mode: 'personal' as const,
      runtime: { name: 'node' },
      techStack: ['react', 'typescript'],
      shouldCreatePackageJson: true,
    };
    vi.clearAllMocks();
  });

  describe('createClaudeSettings', () => {
    it('should create Claude settings.json file', async () => {
      const settingsPath = `${tempDir}/.claude/settings.json`;

      mockFs.existsSync.mockReturnValue(false);
      mockFs.mkdirSync.mockReturnValue(undefined);
      mockFs.writeFileSync.mockReturnValue(undefined);

      const result = await configManager.createClaudeSettings(
        settingsPath,
        mockConfig
      );

      expect(result.success).toBe(true);
      expect(result.filePath).toBe(settingsPath);
      expect(mockFs.mkdirSync).toHaveBeenCalled();
      expect(mockFs.writeFileSync).toHaveBeenCalled();
    });

    it('should handle personal mode Claude settings', async () => {
      const personalConfig = { ...mockConfig, mode: 'personal' as const };
      const settingsPath = `${tempDir}/.claude/settings.json`;

      mockFs.existsSync.mockReturnValue(false);
      mockFs.mkdirSync.mockReturnValue(undefined);
      mockFs.writeFileSync.mockReturnValue(undefined);

      const result = await configManager.createClaudeSettings(
        settingsPath,
        personalConfig
      );

      expect(result.success).toBe(true);
      expect(result.settings?.mode).toBe('personal');
    });

    it('should handle team mode Claude settings', async () => {
      const teamConfig = { ...mockConfig, mode: 'team' as const };
      const settingsPath = `${tempDir}/.claude/settings.json`;

      mockFs.existsSync.mockReturnValue(false);
      mockFs.mkdirSync.mockReturnValue(undefined);
      mockFs.writeFileSync.mockReturnValue(undefined);

      const result = await configManager.createClaudeSettings(
        settingsPath,
        teamConfig
      );

      expect(result.success).toBe(true);
      expect(result.settings?.mode).toBe('team');
    });

    it('should handle file creation errors gracefully', async () => {
      const settingsPath = `${tempDir}/.claude/settings.json`;

      mockFs.writeFileSync.mockImplementation(() => {
        throw new Error('Permission denied');
      });

      const result = await configManager.createClaudeSettings(
        settingsPath,
        mockConfig
      );

      expect(result.success).toBe(false);
      expect(result.error).toContain('Permission denied');
    });
  });

  describe('createMoAIConfig', () => {
    it('should create .moai/config.json file', async () => {
      const configPath = `${tempDir}/.moai/config.json`;

      mockFs.existsSync.mockReturnValue(false);
      mockFs.mkdirSync.mockReturnValue(undefined);
      mockFs.writeFileSync.mockReturnValue(undefined);

      const result = await configManager.createMoAIConfig(
        configPath,
        mockConfig
      );

      expect(result.success).toBe(true);
      expect(result.filePath).toBe(configPath);
      expect(result.config?.projectName).toBe('test-project');
    });

    it('should include version and timestamp in MoAI config', async () => {
      const configPath = `${tempDir}/.moai/config.json`;

      mockFs.existsSync.mockReturnValue(false);
      mockFs.mkdirSync.mockReturnValue(undefined);
      mockFs.writeFileSync.mockReturnValue(undefined);

      const result = await configManager.createMoAIConfig(
        configPath,
        mockConfig
      );

      expect(result.config?.version).toBeDefined();
      expect(result.config?.createdAt).toBeInstanceOf(Date);
    });

    it('should handle backup when file exists', async () => {
      const configPath = `${tempDir}/.moai/config.json`;

      mockFs.existsSync.mockReturnValue(true);
      mockFs.readFileSync.mockReturnValue('{"old": "config"}');
      mockFs.mkdirSync.mockReturnValue(undefined);
      mockFs.writeFileSync.mockReturnValue(undefined);
      vi.spyOn(configManager, 'backupConfigFile').mockResolvedValue({
        success: true,
        backupPath: `${configPath}.backup`,
        timestamp: new Date(),
      });

      const result = await configManager.createMoAIConfig(
        configPath,
        mockConfig
      );

      expect(result.success).toBe(true);
      expect(result.backupCreated).toBe(true);
    });
  });

  describe('createPackageJson', () => {
    it('should create package.json for Node.js projects', async () => {
      const packagePath = `${tempDir}/package.json`;

      mockFs.existsSync.mockReturnValue(false);
      mockFs.writeFileSync.mockReturnValue(undefined);

      const result = await configManager.createPackageJson(
        packagePath,
        mockConfig
      );

      expect(result.success).toBe(true);
      expect(result.packageConfig?.name).toBe('test-project');
      expect(result.packageConfig?.scripts).toBeDefined();
    });

    it('should not create package.json for Python-only projects', async () => {
      const packagePath = `${tempDir}/package.json`;
      const pythonConfig = {
        ...mockConfig,
        runtime: { name: 'python' },
        techStack: ['django'],
        shouldCreatePackageJson: false,
      };

      const result = await configManager.createPackageJson(
        packagePath,
        pythonConfig
      );

      expect(result.success).toBe(true);
      expect(result.skipped).toBe(true);
      expect(result.reason).toContain('not needed');
    });

    it('should include TypeScript dependencies when detected', async () => {
      const packagePath = `${tempDir}/package.json`;
      const tsConfig = {
        ...mockConfig,
        techStack: ['typescript', 'react'],
        shouldCreatePackageJson: true,
      };

      mockFs.existsSync.mockReturnValue(false);
      mockFs.writeFileSync.mockReturnValue(undefined);

      const result = await configManager.createPackageJson(
        packagePath,
        tsConfig
      );

      expect(result.packageConfig?.devDependencies?.['typescript']).toBeDefined();
      expect(
        result.packageConfig?.devDependencies?.['@types/node']
      ).toBeDefined();
    });
  });

  describe('validateConfigFile', () => {
    it('should validate correct JSON structure', async () => {
      const filePath = '/test/config.json';
      const validConfig = { projectName: 'test', version: '1.0.0' };

      mockFs.existsSync.mockReturnValue(true);
      mockFs.readFileSync.mockReturnValue(JSON.stringify(validConfig));

      const result = await configManager.validateConfigFile(filePath);

      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should detect invalid JSON', async () => {
      const filePath = '/test/config.json';

      mockFs.existsSync.mockReturnValue(true);
      mockFs.readFileSync.mockReturnValue('invalid json');

      const result = await configManager.validateConfigFile(filePath);

      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Invalid JSON format');
    });

    it('should handle missing files', async () => {
      const filePath = '/test/config.json';

      mockFs.existsSync.mockReturnValue(false);

      const result = await configManager.validateConfigFile(filePath);

      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('File does not exist');
    });
  });

  describe('backupConfigFile', () => {
    it('should create backup with timestamp', async () => {
      const filePath = '/test/config.json';
      const backupContent = '{"test": "data"}';

      mockFs.existsSync.mockReturnValue(true);
      mockFs.readFileSync.mockReturnValue(backupContent);
      mockFs.writeFileSync.mockReturnValue(undefined);

      const result = await configManager.backupConfigFile(filePath);

      expect(result.success).toBe(true);
      expect(result.backupPath).toContain('.backup.');
      expect(mockFs.writeFileSync).toHaveBeenCalledWith(
        expect.stringContaining('.backup.'),
        backupContent,
        'utf-8'
      );
    });

    it('should handle backup creation errors', async () => {
      const filePath = '/test/config.json';

      mockFs.existsSync.mockReturnValue(true);
      mockFs.readFileSync.mockReturnValue('{"test": "data"}');
      mockFs.writeFileSync.mockImplementation(() => {
        throw new Error('Disk full');
      });

      const result = await configManager.backupConfigFile(filePath);

      expect(result.success).toBe(false);
      expect(result.error).toContain('Disk full');
    });
  });

  describe('setupFullProjectConfig', () => {
    it('should create all configuration files', async () => {
      mockFs.existsSync.mockReturnValue(false);
      mockFs.mkdirSync.mockReturnValue(undefined);
      mockFs.writeFileSync.mockReturnValue(undefined);

      const result = await configManager.setupFullProjectConfig(
        tempDir,
        mockConfig
      );

      expect(result.success).toBe(true);
      expect(result.filesCreated).toHaveLength(3); // Claude + MoAI + package.json
      expect(result.claudeSettings).toBeDefined();
      expect(result.moaiConfig).toBeDefined();
      expect(result.packageJson).toBeDefined();
    });

    it('should handle partial failures gracefully', async () => {
      mockFs.existsSync.mockReturnValue(false);
      mockFs.mkdirSync.mockReturnValue(undefined);
      mockFs.writeFileSync.mockImplementation((filePath: any) => {
        if (filePath.includes('settings.json')) {
          throw new Error('Claude settings failed');
        }
      });

      const result = await configManager.setupFullProjectConfig(
        tempDir,
        mockConfig
      );

      expect(result.success).toBe(false);
      expect(result.errors).toContain('Claude settings failed');
    });
  });
});

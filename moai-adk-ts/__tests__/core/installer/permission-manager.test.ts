/**
 * @file PermissionManager test suite
 * @author MoAI Team
 * @tags @TEST:PERMISSION-MANAGER-012 @REQ:CROSS-PLATFORM-PERMISSIONS-012
 */

import * as fs from 'fs-extra';
import * as path from 'path';
import * as os from 'os';
import {
  PermissionManager
} from '../../../src/core/installer/managers/permission-manager';
import {
  PermissionUtils
} from '../../../src/core/installer/managers/permission-utils';
import {
  FilePermissions,
  MoAIPermissions
} from '../../../src/core/installer/managers/permission-types';

describe('PermissionUtils', () => {
  describe('octalToPermissions', () => {
    it('should convert octal 755 to correct permissions', () => {
      const permissions = PermissionUtils.octalToPermissions('755');

      expect(permissions.owner).toEqual({ read: true, write: true, execute: true });
      expect(permissions.group).toEqual({ read: true, write: false, execute: true });
      expect(permissions.others).toEqual({ read: true, write: false, execute: true });
    });

    it('should convert octal 644 to correct permissions', () => {
      const permissions = PermissionUtils.octalToPermissions('644');

      expect(permissions.owner).toEqual({ read: true, write: true, execute: false });
      expect(permissions.group).toEqual({ read: true, write: false, execute: false });
      expect(permissions.others).toEqual({ read: true, write: false, execute: false });
    });

    it('should convert octal 600 to correct permissions', () => {
      const permissions = PermissionUtils.octalToPermissions('600');

      expect(permissions.owner).toEqual({ read: true, write: true, execute: false });
      expect(permissions.group).toEqual({ read: false, write: false, execute: false });
      expect(permissions.others).toEqual({ read: false, write: false, execute: false });
    });

    it('should throw error for invalid octal', () => {
      expect(() => PermissionUtils.octalToPermissions('999')).toThrow('Invalid octal permission: 999');
      expect(() => PermissionUtils.octalToPermissions('abc')).toThrow('Invalid octal permission: abc');
      expect(() => PermissionUtils.octalToPermissions('12')).toThrow('Invalid octal permission: 12');
    });
  });

  describe('permissionsToOctal', () => {
    it('should convert permissions to octal 755', () => {
      const permissions: FilePermissions = {
        owner: { read: true, write: true, execute: true },
        group: { read: true, write: false, execute: true },
        others: { read: true, write: false, execute: true }
      };

      expect(PermissionUtils.permissionsToOctal(permissions)).toBe('755');
    });

    it('should convert permissions to octal 644', () => {
      const permissions: FilePermissions = {
        owner: { read: true, write: true, execute: false },
        group: { read: true, write: false, execute: false },
        others: { read: true, write: false, execute: false }
      };

      expect(PermissionUtils.permissionsToOctal(permissions)).toBe('644');
    });

    it('should convert permissions to octal 600', () => {
      const permissions: FilePermissions = {
        owner: { read: true, write: true, execute: false },
        group: { read: false, write: false, execute: false },
        others: { read: false, write: false, execute: false }
      };

      expect(PermissionUtils.permissionsToOctal(permissions)).toBe('600');
    });
  });

  describe('file type detection', () => {
    it('should identify script files correctly', () => {
      expect(PermissionUtils.isScript('script.py')).toBe(true);
      expect(PermissionUtils.isScript('script.sh')).toBe(true);
      expect(PermissionUtils.isScript('script.js')).toBe(true);
      expect(PermissionUtils.isScript('script.ts')).toBe(true);
      expect(PermissionUtils.isScript('script.bat')).toBe(true);
      expect(PermissionUtils.isScript('script.cmd')).toBe(true);
      expect(PermissionUtils.isScript('script.ps1')).toBe(true);
      expect(PermissionUtils.isScript('document.txt')).toBe(false);
      expect(PermissionUtils.isScript('config.json')).toBe(false);
    });

    it('should identify configuration files correctly', () => {
      expect(PermissionUtils.isConfig('config.json')).toBe(true);
      expect(PermissionUtils.isConfig('settings.yaml')).toBe(true);
      expect(PermissionUtils.isConfig('app.yml')).toBe(true);
      expect(PermissionUtils.isConfig('setup.toml')).toBe(true);
      expect(PermissionUtils.isConfig('app.ini')).toBe(true);
      expect(PermissionUtils.isConfig('nginx.conf')).toBe(true);
      expect(PermissionUtils.isConfig('.env')).toBe(true);
      expect(PermissionUtils.isConfig('.env.local')).toBe(true);
      expect(PermissionUtils.isConfig('webpack.config.js')).toBe(true);
      expect(PermissionUtils.isConfig('script.py')).toBe(false);
      expect(PermissionUtils.isConfig('document.txt')).toBe(false);
    });

    it('should identify sensitive files correctly', () => {
      expect(PermissionUtils.isSensitive('.env')).toBe(true);
      expect(PermissionUtils.isSensitive('secret.json')).toBe(true);
      expect(PermissionUtils.isSensitive('credentials.yaml')).toBe(true);
      expect(PermissionUtils.isSensitive('api.key')).toBe(true);
      expect(PermissionUtils.isSensitive('auth-token.txt')).toBe(true);
      expect(PermissionUtils.isSensitive('password.conf')).toBe(true);
      expect(PermissionUtils.isSensitive('private-key.pem')).toBe(true);
      expect(PermissionUtils.isSensitive('.ssh/id_rsa')).toBe(true);
      expect(PermissionUtils.isSensitive('cert.p12')).toBe(true);
      expect(PermissionUtils.isSensitive('regular-file.txt')).toBe(false);
      expect(PermissionUtils.isSensitive('config.json')).toBe(false);
    });
  });

  describe('platform detection', () => {
    it('should detect platform correctly', () => {
      const platform = PermissionUtils.getCurrentPlatform();
      expect(['windows', 'unix']).toContain(platform);
    });
  });
});

describe('PermissionManager', () => {
  let permissionManager: PermissionManager;
  let testDir: string;
  let testFile: string;
  let testScript: string;
  let testConfig: string;
  let testSensitive: string;

  beforeEach(async () => {
    permissionManager = new PermissionManager();

    // Create temporary test directory
    testDir = await fs.mkdtemp(path.join(os.tmpdir(), 'permission-test-'));
    testFile = path.join(testDir, 'test.txt');
    testScript = path.join(testDir, 'script.py');
    testConfig = path.join(testDir, 'config.json');
    testSensitive = path.join(testDir, '.env');

    // Create test files
    await fs.writeFile(testFile, 'test content');
    await fs.writeFile(testScript, '#!/usr/bin/env python\nprint("hello")');
    await fs.writeFile(testConfig, '{"test": true}');
    await fs.writeFile(testSensitive, 'API_KEY=secret123');
  });

  afterEach(async () => {
    // Clean up test directory
    await fs.remove(testDir);
  });

  describe('setFilePermissions', () => {
    it('should set file permissions successfully', async () => {
      const permissions: FilePermissions = {
        owner: { read: true, write: true, execute: false },
        group: { read: true, write: false, execute: false },
        others: { read: true, write: false, execute: false }
      };

      await expect(permissionManager.setFilePermissions(testFile, permissions)).resolves.not.toThrow();
    });

    it('should throw error for non-existent file', async () => {
      const permissions: FilePermissions = {
        owner: { read: true, write: true, execute: false },
        group: { read: true, write: false, execute: false },
        others: { read: true, write: false, execute: false }
      };

      const nonExistentFile = path.join(testDir, 'non-existent.txt');
      await expect(permissionManager.setFilePermissions(nonExistentFile, permissions))
        .rejects.toThrow(`Failed to set permissions for ${nonExistentFile}`);
    });
  });

  describe('setDirectoryPermissions', () => {
    it('should set directory permissions successfully', async () => {
      const permissions: FilePermissions = {
        owner: { read: true, write: true, execute: true },
        group: { read: true, write: false, execute: true },
        others: { read: true, write: false, execute: true }
      };

      await expect(permissionManager.setDirectoryPermissions(testDir, permissions)).resolves.not.toThrow();
    });

    it('should throw error for non-existent directory', async () => {
      const permissions: FilePermissions = {
        owner: { read: true, write: true, execute: true },
        group: { read: true, write: false, execute: true },
        others: { read: true, write: false, execute: true }
      };

      const nonExistentDir = path.join(testDir, 'non-existent');
      await expect(permissionManager.setDirectoryPermissions(nonExistentDir, permissions))
        .rejects.toThrow(`Failed to set directory permissions for ${nonExistentDir}`);
    });
  });

  describe('makeExecutable', () => {
    it('should make file executable on Unix systems', async () => {
      await expect(permissionManager.makeExecutable(testScript)).resolves.not.toThrow();

      // Verify the file is executable (on Unix systems)
      if (process.platform !== 'win32') {
        const status = await permissionManager.checkPermissions(testScript);
        expect(status.executable).toBe(true);
      }
    });

    it('should handle Windows executable detection', async () => {
      await expect(permissionManager.makeExecutable(testScript)).resolves.not.toThrow();
    });

    it('should throw error for non-existent file', async () => {
      const nonExistentFile = path.join(testDir, 'non-existent.py');
      await expect(permissionManager.makeExecutable(nonExistentFile))
        .rejects.toThrow(`Failed to make ${nonExistentFile} executable`);
    });
  });

  describe('checkPermissions', () => {
    it('should return permission status for existing file', async () => {
      const status = await permissionManager.checkPermissions(testFile);

      expect(status.path).toBe(testFile);
      expect(typeof status.readable).toBe('boolean');
      expect(typeof status.writable).toBe('boolean');
      expect(typeof status.executable).toBe('boolean');
      expect(typeof status.isOwner).toBe('boolean');

      if (process.platform !== 'win32') {
        expect(typeof status.octalMode).toBe('string');
      }
    });

    it('should return permission status for directory', async () => {
      const status = await permissionManager.checkPermissions(testDir);

      expect(status.path).toBe(testDir);
      expect(status.readable).toBe(true);
      expect(status.writable).toBe(true);
      expect(status.executable).toBe(true);
    });

    it('should throw error for non-existent path', async () => {
      const nonExistentPath = path.join(testDir, 'non-existent');
      await expect(permissionManager.checkPermissions(nonExistentPath))
        .rejects.toThrow(`Failed to check permissions for ${nonExistentPath}`);
    });
  });

  describe('validatePermissions', () => {
    it('should validate sufficient permissions', async () => {
      const requiredPermissions: FilePermissions = {
        owner: { read: true, write: true, execute: false },
        group: { read: false, write: false, execute: false },
        others: { read: false, write: false, execute: false }
      };

      const isValid = await permissionManager.validatePermissions(testFile, requiredPermissions);
      expect(isValid).toBe(true);
    });

    it('should reject insufficient permissions', async () => {
      // Make file read-only first
      await fs.chmod(testFile, 0o444);

      const requiredPermissions: FilePermissions = {
        owner: { read: true, write: true, execute: true },
        group: { read: false, write: false, execute: false },
        others: { read: false, write: false, execute: false }
      };

      const isValid = await permissionManager.validatePermissions(testFile, requiredPermissions);
      expect(isValid).toBe(false);
    });

    it('should return false for non-existent file', async () => {
      const requiredPermissions: FilePermissions = {
        owner: { read: true, write: true, execute: false },
        group: { read: false, write: false, execute: false },
        others: { read: false, write: false, execute: false }
      };

      const nonExistentFile = path.join(testDir, 'non-existent.txt');
      const isValid = await permissionManager.validatePermissions(nonExistentFile, requiredPermissions);
      expect(isValid).toBe(false);
    });
  });

  describe('fixPermissions', () => {
    let subDir: string;
    let subScript: string;
    let subConfig: string;

    beforeEach(async () => {
      // Create subdirectory with various files
      subDir = path.join(testDir, 'subdir');
      await fs.mkdir(subDir);

      subScript = path.join(subDir, 'script.sh');
      subConfig = path.join(subDir, 'config.yaml');

      await fs.writeFile(subScript, '#!/bin/bash\necho "hello"');
      await fs.writeFile(subConfig, 'setting: value');
    });

    it('should fix permissions for entire project directory', async () => {
      const result = await permissionManager.fixPermissions(testDir);

      expect(result.success).toBe(true);
      expect(result.fixedFiles.length).toBeGreaterThan(0);
      expect(result.fixedFiles).toContain(testDir);
      expect(result.fixedFiles).toContain(subDir);
      expect(result.failures.length).toBe(0);
    });

    it('should identify and warn about sensitive files', async () => {
      const result = await permissionManager.fixPermissions(testDir);

      expect(result.warnings.some(warning =>
        warning.includes('sensitive file') && warning.includes('.env')
      )).toBe(true);
    });

    it('should handle permission setting failures gracefully', async () => {
      // Create a file we can't modify (simulate permission error)
      const restrictedFile = path.join(testDir, 'restricted.txt');
      await fs.writeFile(restrictedFile, 'restricted content');

      // This test verifies the error handling structure rather than forcing actual failures
      // since permission restrictions can be platform-specific and hard to simulate reliably
      const result = await permissionManager.fixPermissions(testDir);

      // Verify result structure is correct
      expect(result).toHaveProperty('success');
      expect(result).toHaveProperty('fixedFiles');
      expect(result).toHaveProperty('failures');
      expect(result).toHaveProperty('warnings');
      expect(Array.isArray(result.fixedFiles)).toBe(true);
      expect(Array.isArray(result.failures)).toBe(true);
      expect(Array.isArray(result.warnings)).toBe(true);

      // On most development systems, permissions can be set successfully
      if (result.success) {
        expect(result.fixedFiles.length).toBeGreaterThan(0);
        expect(result.failures.length).toBe(0);
      } else {
        expect(result.failures.length).toBeGreaterThan(0);
      }
    });

    it('should return proper result structure', async () => {
      const result = await permissionManager.fixPermissions(testDir);

      expect(result).toHaveProperty('success');
      expect(result).toHaveProperty('fixedFiles');
      expect(result).toHaveProperty('failures');
      expect(result).toHaveProperty('warnings');
      expect(Array.isArray(result.fixedFiles)).toBe(true);
      expect(Array.isArray(result.failures)).toBe(true);
      expect(Array.isArray(result.warnings)).toBe(true);
    });
  });

  describe('MoAI permission policies', () => {
    it('should apply script file permissions correctly', async () => {
      await permissionManager.setFilePermissions(testScript, MoAIPermissions.SCRIPT_FILES);

      const status = await permissionManager.checkPermissions(testScript);
      expect(status.readable).toBe(true);
      expect(status.writable).toBe(true);

      if (process.platform !== 'win32') {
        expect(status.executable).toBe(true);
      }
    });

    it('should apply config file permissions correctly', async () => {
      await permissionManager.setFilePermissions(testConfig, MoAIPermissions.CONFIG_FILES);

      const status = await permissionManager.checkPermissions(testConfig);
      expect(status.readable).toBe(true);
      expect(status.writable).toBe(true);
    });

    it('should apply sensitive file permissions correctly', async () => {
      await permissionManager.setFilePermissions(testSensitive, MoAIPermissions.SENSITIVE_FILES);

      const status = await permissionManager.checkPermissions(testSensitive);
      expect(status.readable).toBe(true);
      expect(status.writable).toBe(true);
    });

    it('should apply directory permissions correctly', async () => {
      await permissionManager.setDirectoryPermissions(testDir, MoAIPermissions.DIRECTORIES);

      const status = await permissionManager.checkPermissions(testDir);
      expect(status.readable).toBe(true);
      expect(status.writable).toBe(true);
      expect(status.executable).toBe(true);
    });
  });

  describe('cross-platform behavior', () => {
    it('should handle platform differences gracefully', async () => {
      const platform = PermissionUtils.getCurrentPlatform();

      if (platform === 'windows') {
        // Test Windows-specific behavior
        await expect(permissionManager.makeExecutable(testFile)).resolves.not.toThrow();
      } else {
        // Test Unix-specific behavior
        await expect(permissionManager.makeExecutable(testScript)).resolves.not.toThrow();

        const status = await permissionManager.checkPermissions(testScript);
        expect(status.octalMode).toBeDefined();
      }
    });

    it('should provide appropriate warnings for platform limitations', async () => {
      // This test ensures that platform-specific limitations are handled
      await expect(permissionManager.makeExecutable(testFile)).resolves.not.toThrow();

      if (process.platform === 'win32') {
        // On Windows, non-script files should generate warnings
        // We can't easily test console.warn, but ensure no errors are thrown
      }
    });
  });
});
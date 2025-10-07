// @TEST:GIT-LOCALE-LOADER-001 | SPEC: SPEC-GIT-LOCALE-001.md
/**
 * @file Locale Loader Test Suite
 * @tags @TEST:GIT-LOCALE-LOADER-001
 * @description .moai/config.json에서 locale 로드 검증
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  getLocaleWithFallback,
  hasConfigFile,
  loadLocaleFromConfig,
} from '../locale-loader';

describe('@TEST:GIT-LOCALE-LOADER-001 - Locale Loader', () => {
  const mockWorkingDir = '/test/project';
  const mockConfigPath = path.join(mockWorkingDir, '.moai', 'config.json');

  // Mock fs module
  vi.mock('node:fs');

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('hasConfigFile', () => {
    it('should return true if config file exists', () => {
      vi.spyOn(fs, 'existsSync').mockReturnValue(true);

      expect(hasConfigFile(mockWorkingDir)).toBe(true);
      expect(fs.existsSync).toHaveBeenCalledWith(mockConfigPath);
    });

    it('should return false if config file does not exist', () => {
      vi.spyOn(fs, 'existsSync').mockReturnValue(false);

      expect(hasConfigFile(mockWorkingDir)).toBe(false);
      expect(fs.existsSync).toHaveBeenCalledWith(mockConfigPath);
    });
  });

  describe('loadLocaleFromConfig', () => {
    it('should load Korean locale from config', () => {
      const mockConfig = {
        project: {
          locale: 'ko',
        },
      };

      vi.spyOn(fs, 'existsSync').mockReturnValue(true);
      vi.spyOn(fs, 'readFileSync').mockReturnValue(JSON.stringify(mockConfig));

      const locale = loadLocaleFromConfig(mockWorkingDir);
      expect(locale).toBe('ko');
    });

    it('should load English locale from config', () => {
      const mockConfig = {
        project: {
          locale: 'en',
        },
      };

      vi.spyOn(fs, 'existsSync').mockReturnValue(true);
      vi.spyOn(fs, 'readFileSync').mockReturnValue(JSON.stringify(mockConfig));

      const locale = loadLocaleFromConfig(mockWorkingDir);
      expect(locale).toBe('en');
    });

    it('should load Japanese locale from config', () => {
      const mockConfig = {
        project: {
          locale: 'ja',
        },
      };

      vi.spyOn(fs, 'existsSync').mockReturnValue(true);
      vi.spyOn(fs, 'readFileSync').mockReturnValue(JSON.stringify(mockConfig));

      const locale = loadLocaleFromConfig(mockWorkingDir);
      expect(locale).toBe('ja');
    });

    it('should load Chinese locale from config', () => {
      const mockConfig = {
        project: {
          locale: 'zh',
        },
      };

      vi.spyOn(fs, 'existsSync').mockReturnValue(true);
      vi.spyOn(fs, 'readFileSync').mockReturnValue(JSON.stringify(mockConfig));

      const locale = loadLocaleFromConfig(mockWorkingDir);
      expect(locale).toBe('zh');
    });

    it('should return en if config file does not exist', () => {
      vi.spyOn(fs, 'existsSync').mockReturnValue(false);

      const locale = loadLocaleFromConfig(mockWorkingDir);
      expect(locale).toBe('en');
    });

    it('should return en if locale is invalid', () => {
      const mockConfig = {
        project: {
          locale: 'invalid',
        },
      };

      vi.spyOn(fs, 'existsSync').mockReturnValue(true);
      vi.spyOn(fs, 'readFileSync').mockReturnValue(JSON.stringify(mockConfig));

      const locale = loadLocaleFromConfig(mockWorkingDir);
      expect(locale).toBe('en');
    });

    it('should return en if locale is missing', () => {
      const mockConfig = {
        project: {},
      };

      vi.spyOn(fs, 'existsSync').mockReturnValue(true);
      vi.spyOn(fs, 'readFileSync').mockReturnValue(JSON.stringify(mockConfig));

      const locale = loadLocaleFromConfig(mockWorkingDir);
      expect(locale).toBe('en');
    });

    it('should return en if config is malformed', () => {
      vi.spyOn(fs, 'existsSync').mockReturnValue(true);
      vi.spyOn(fs, 'readFileSync').mockReturnValue('invalid json');

      const locale = loadLocaleFromConfig(mockWorkingDir);
      expect(locale).toBe('en');
    });

    it('should return en if readFileSync throws error', () => {
      vi.spyOn(fs, 'existsSync').mockReturnValue(true);
      vi.spyOn(fs, 'readFileSync').mockImplementation(() => {
        throw new Error('Permission denied');
      });

      const locale = loadLocaleFromConfig(mockWorkingDir);
      expect(locale).toBe('en');
    });
  });

  describe('getLocaleWithFallback', () => {
    it('should use config locale if available', () => {
      const mockConfig = {
        project: {
          locale: 'ko',
        },
      };

      vi.spyOn(fs, 'existsSync').mockReturnValue(true);
      vi.spyOn(fs, 'readFileSync').mockReturnValue(JSON.stringify(mockConfig));

      const locale = getLocaleWithFallback(mockWorkingDir);
      expect(locale).toBe('ko');
    });

    it('should fallback to environment variable if config has en', () => {
      const mockConfig = {
        project: {
          locale: 'en',
        },
      };

      vi.spyOn(fs, 'existsSync').mockReturnValue(true);
      vi.spyOn(fs, 'readFileSync').mockReturnValue(JSON.stringify(mockConfig));

      // Mock environment variable
      const originalEnv = process.env.MOAI_LOCALE;
      process.env.MOAI_LOCALE = 'ja';

      const locale = getLocaleWithFallback(mockWorkingDir);
      expect(locale).toBe('ja');

      // Restore
      process.env.MOAI_LOCALE = originalEnv;
    });

    it('should use environment variable if config does not exist', () => {
      vi.spyOn(fs, 'existsSync').mockReturnValue(false);

      const originalEnv = process.env.MOAI_LOCALE;
      process.env.MOAI_LOCALE = 'zh';

      const locale = getLocaleWithFallback(mockWorkingDir);
      expect(locale).toBe('zh');

      process.env.MOAI_LOCALE = originalEnv;
    });

    it('should fallback to en if no config and no env', () => {
      vi.spyOn(fs, 'existsSync').mockReturnValue(false);

      const originalEnv = process.env.MOAI_LOCALE;
      delete process.env.MOAI_LOCALE;

      const locale = getLocaleWithFallback(mockWorkingDir);
      expect(locale).toBe('en');

      process.env.MOAI_LOCALE = originalEnv;
    });

    it('should validate environment variable locale', () => {
      vi.spyOn(fs, 'existsSync').mockReturnValue(false);

      const originalEnv = process.env.MOAI_LOCALE;
      process.env.MOAI_LOCALE = 'invalid';

      const locale = getLocaleWithFallback(mockWorkingDir);
      expect(locale).toBe('en'); // Falls back to default

      process.env.MOAI_LOCALE = originalEnv;
    });
  });

  describe('Integration Scenarios', () => {
    it('should handle complete MoAI config structure', () => {
      const mockConfig = {
        _meta: {
          '@CODE:CONFIG-STRUCTURE-001': '@DOC:JSON-CONFIG-001',
        },
        moai: {
          version: '0.1.0',
        },
        project: {
          name: 'test-project',
          mode: 'team',
          locale: 'ko',
          initialized: true,
        },
        constitution: {
          enforce_tdd: true,
        },
      };

      vi.spyOn(fs, 'existsSync').mockReturnValue(true);
      vi.spyOn(fs, 'readFileSync').mockReturnValue(JSON.stringify(mockConfig));

      const locale = loadLocaleFromConfig(mockWorkingDir);
      expect(locale).toBe('ko');
    });

    it('should handle nested project structure', () => {
      const nestedDir = '/test/project/subdir';
      const configPath = path.join(nestedDir, '.moai', 'config.json');

      const mockConfig = {
        project: {
          locale: 'ja',
        },
      };

      vi.spyOn(fs, 'existsSync').mockImplementation(p => p === configPath);
      vi.spyOn(fs, 'readFileSync').mockReturnValue(JSON.stringify(mockConfig));

      const locale = loadLocaleFromConfig(nestedDir);
      expect(locale).toBe('ja');
    });
  });
});

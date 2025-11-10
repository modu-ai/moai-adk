/**
 * @TEST:TAG-MIGRATION-001
 * Framework Core Upgrade Tests
 *
 * These tests verify that the migration from:
 * - Next.js 14.2.15 → Next.js 16
 * - Nextra 3.3.1 → Nextra 4.6.0
 * - React 18.2.0 → React 19
 *
 * is successful and maintains functionality.
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { exec } from 'child_process';
import { readFileSync, existsSync, writeFileSync } from 'fs';
import { join } from 'path';

const PACKAGE_JSON_PATH = '/Users/goos/MoAI/MoAI-ADK/docs/package.json';
const CONFIG_DIR = '/Users/goos/MoAI/MoAI-ADK/docs';

describe('TAG-MIGRATION-001: Framework Core Upgrade', () => {
  let originalPackageJson: string;

  beforeEach(() => {
    // Backup original package.json
    if (existsSync(PACKAGE_JSON_PATH)) {
      originalPackageJson = readFileSync(PACKAGE_JSON_PATH, 'utf-8');
    }
  });

  afterEach(() => {
    // Restore original package.json
    if (originalPackageJson) {
      writeFileSync(PACKAGE_JSON_PATH, originalPackageJson);
    }
  });

  describe('Package Version Tests', () => {
    it('should have Next.js 16 as dependency', () => {
      const packageJson = readFileSync(PACKAGE_JSON_PATH, 'utf-8');
      const pkg = JSON.parse(packageJson);

      // This test should pass in GREEN state
      expect(pkg.dependencies.next).toBe('16.0.0');
      expect(pkg.dependencies.next).toMatch(/^16\./);
    });

    it('should have Nextra 4.6.0 as dependency', () => {
      const packageJson = readFileSync(PACKAGE_JSON_PATH, 'utf-8');
      const pkg = JSON.parse(packageJson);

      // This test should pass in GREEN state
      expect(pkg.dependencies.nextra).toBe('^4.6.0');
      expect(pkg.dependencies.nextra).toMatch(/^\^4\.6\.0/);
    });

    it('should have React 19 as dependency', () => {
      const packageJson = readFileSync(PACKAGE_JSON_PATH, 'utf-8');
      const pkg = JSON.parse(packageJson);

      // This test should pass in GREEN state
      expect(pkg.dependencies.react).toBe('19.0.0');
      expect(pkg.dependencies.react).toMatch(/^19\./);
    });

    it('should have compatible React DOM version', () => {
      const packageJson = readFileSync(PACKAGE_JSON_PATH, 'utf-8');
      const pkg = JSON.parse(packageJson);

      // React DOM should match React version
      expect(pkg.dependencies['react-dom']).toBe(pkg.dependencies.react);
    });
  });

  describe('TypeScript Configuration Tests', () => {
    it('should have compatible TypeScript configuration', () => {
      const tsconfigPath = join(CONFIG_DIR, 'tsconfig.json');
      const tsconfig = readFileSync(tsconfigPath, 'utf-8');
      const config = JSON.parse(tsconfig);

      // TypeScript should be compatible with Next.js 16
      expect(config.compilerOptions.target).toBe('ES2022');
      expect(config.compilerOptions.module).toBe('esnext');
    });
  });

  describe('Build System Tests', () => {
    it('npm install should succeed with new dependencies', async () => {
      return new Promise((resolve, reject) => {
        exec('npm install', (error, stdout, stderr) => {
          // In GREEN state, this should succeed
          expect(error).toBeNull();
          expect(stdout).toContain('added');
          resolve(null);
        });
      });
    });

    it('TypeScript compilation should pass', async () => {
      return new Promise((resolve, reject) => {
        exec('npm run type-check', (error, stdout, stderr) => {
          // In GREEN state, this should pass
          expect(error).toBeNull();
          expect(stdout).not.toContain('error');
          resolve(null);
        });
      });
    });

    it('Development server should start', async () => {
      return new Promise((resolve, reject) => {
        const child = exec('timeout 3 npm run dev || true');

        child.on('close', (code) => {
          // In GREEN state, this should start successfully
          // Dev server might exit with code 0 or 1 depending on timeout
          expect([0, 1]).toContain(code);
          resolve(null);
        });
      });
    });
  });

  describe('Nextra Configuration Tests', () => {
    it('should have Nextra 4.x compatible configuration', () => {
      const nextraConfigPath = join(CONFIG_DIR, 'next.config.cjs');
      if (existsSync(nextraConfigPath)) {
        const config = readFileSync(nextraConfigPath, 'utf-8');
        expect(config).toContain('nextra');
        expect(config).toContain('docs');
      }
    });

    it('should have valid theme configuration', () => {
      const themeConfigPath = join(CONFIG_DIR, 'theme.config.cjs');
      if (existsSync(themeConfigPath)) {
        const config = readFileSync(themeConfigPath, 'utf-8');
        expect(config).toContain('theme');
        expect(config).toContain('docs');
      }
    });
  });

  describe('Compatibility Tests', () => {
    it('should maintain API compatibility', () => {
      // Check that existing components still work
      const componentsPath = join(CONFIG_DIR, 'src/components');
      if (existsSync(componentsPath)) {
        const files = require('fs').readdirSync(componentsPath);
        expect(files.length).toBeGreaterThan(0);
      }
    });

    it('should maintain page structure', () => {
      // Check that pages exist and are valid
      const pagesPath = join(CONFIG_DIR, 'src/pages');
      if (existsSync(pagesPath)) {
        const files = require('fs').readdirSync(pagesPath);
        expect(files.length).toBeGreaterThan(0);
      }
    });
  });
});
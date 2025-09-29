/**
 * @file ProjectDetector test suite - RED phase TDD
 * @author MoAI Team
 * @tags @TEST:PROJECT-DETECTOR-001 @REQ:CORE-SYSTEM-013
 */

import fs from 'node:fs';
import { beforeEach, describe, expect, vi } from 'vitest';
import { ProjectDetector } from '../project-detector';

// Mock fs and path modules
vi.mock('fs');
const mockFs = fs as vi.Mocked<typeof fs>;

describe('ProjectDetector', () => {
  let detector: ProjectDetector;
  let tempDir: string;

  beforeEach(() => {
    detector = new ProjectDetector();
    tempDir = '/test/project';
    vi.clearAllMocks();
  });

  describe('detectProjectType', () => {
    it('should detect Node.js project from package.json', async () => {
      // RED: This test should fail initially
      mockFs.existsSync.mockImplementation((filePath: any) => {
        return filePath.includes('package.json');
      });

      const result = await detector.detectProjectType(tempDir);

      expect(result.type).toBe('nodejs');
      expect(result.language).toBe('javascript');
      expect(result.filesFound).toContain('package.json');
    });

    it('should detect Python project from pyproject.toml', async () => {
      mockFs.existsSync.mockImplementation((filePath: any) => {
        return filePath.includes('pyproject.toml');
      });

      const result = await detector.detectProjectType(tempDir);

      expect(result.type).toBe('python');
      expect(result.language).toBe('python');
      expect(result.filesFound).toContain('pyproject.toml');
    });

    it('should detect TypeScript project from package.json with TypeScript', async () => {
      const packageJson = {
        dependencies: { typescript: '^5.0.0' },
        devDependencies: { '@types/node': '^20.0.0' },
      };

      mockFs.existsSync.mockImplementation((filePath: any) => {
        return filePath.includes('package.json');
      });

      mockFs.readFileSync.mockReturnValue(JSON.stringify(packageJson));

      const result = await detector.detectProjectType(tempDir);

      expect(result.buildTools).toContain('typescript');
      expect(result.language).toBe('javascript'); // Base language from package.json
    });

    it('should return unknown for empty directory', async () => {
      mockFs.existsSync.mockReturnValue(false);

      const result = await detector.detectProjectType(tempDir);

      expect(result.type).toBe('unknown');
      expect(result.language).toBe('unknown');
      expect(result.filesFound).toHaveLength(0);
    });
  });

  describe('analyzePackageJson', () => {
    it('should detect React framework', async () => {
      const packageJson = {
        dependencies: { react: '^18.0.0', 'react-dom': '^18.0.0' },
        devDependencies: { '@types/react': '^18.0.0' },
      };

      mockFs.readFileSync.mockReturnValue(JSON.stringify(packageJson));

      const result = await detector.analyzePackageJson('/test/package.json');

      expect(result.frameworks).toContain('react');
    });

    it('should detect Next.js framework', async () => {
      const packageJson = {
        dependencies: { next: '^13.0.0', react: '^18.0.0' },
      };

      mockFs.readFileSync.mockReturnValue(JSON.stringify(packageJson));

      const result = await detector.analyzePackageJson('/test/package.json');

      expect(result.frameworks).toContain('nextjs');
      expect(result.frameworks).toContain('react');
    });

    it('should detect build tools', async () => {
      const packageJson = {
        devDependencies: {
          vite: '^4.0.0',
          typescript: '^5.0.0',
          webpack: '^5.0.0',
        },
      };

      mockFs.readFileSync.mockReturnValue(JSON.stringify(packageJson));

      const result = await detector.analyzePackageJson('/test/package.json');

      expect(result.buildTools).toContain('vite');
      expect(result.buildTools).toContain('typescript');
      expect(result.buildTools).toContain('webpack');
    });

    it('should handle invalid JSON gracefully', async () => {
      mockFs.readFileSync.mockReturnValue('invalid json');

      const result = await detector.analyzePackageJson('/test/package.json');

      expect(result.frameworks).toHaveLength(0);
      expect(result.buildTools).toHaveLength(0);
    });
  });

  describe('detectLanguageFromFiles', () => {
    it('should detect TypeScript as primary language', async () => {
      // Mock existsSync to return true
      mockFs.existsSync.mockReturnValue(true);

      // Mock file walking to return TypeScript files
      const mockFiles = [
        { path: '/test/src/index.ts', isFile: () => true },
        { path: '/test/src/components/App.tsx', isFile: () => true },
        { path: '/test/src/utils/helper.js', isFile: () => true },
      ];

      // Mock directory scanning
      vi.spyOn(detector as any, 'scanDirectory').mockResolvedValue(mockFiles);

      const result = await detector.detectLanguageFromFiles(tempDir);

      expect(result).toBe('typescript');
    });

    it('should detect Python as primary language', async () => {
      // Mock existsSync to return true
      mockFs.existsSync.mockReturnValue(true);

      const mockFiles = [
        { path: '/test/main.py', isFile: () => true },
        { path: '/test/utils/helper.py', isFile: () => true },
        { path: '/test/tests/test_main.py', isFile: () => true },
      ];

      vi.spyOn(detector as any, 'scanDirectory').mockResolvedValue(mockFiles);

      const result = await detector.detectLanguageFromFiles(tempDir);

      expect(result).toBe('python');
    });

    it('should return unknown for mixed or unclear projects', async () => {
      const mockFiles = [
        { path: '/test/README.md', isFile: () => true },
        { path: '/test/config.txt', isFile: () => true },
      ];

      vi.spyOn(detector as any, 'scanDirectory').mockResolvedValue(mockFiles);

      const result = await detector.detectLanguageFromFiles(tempDir);

      expect(result).toBe('unknown');
    });
  });

  describe('shouldCreatePackageJson', () => {
    it('should return true for Node.js projects', () => {
      const config = { runtime: { name: 'node' }, techStack: [] };

      const result = detector.shouldCreatePackageJson(config);

      expect(result).toBe(true);
    });

    it('should return true for React projects', () => {
      const config = {
        runtime: { name: 'python' },
        techStack: ['react', 'nextjs'],
      };

      const result = detector.shouldCreatePackageJson(config);

      expect(result).toBe(true);
    });

    it('should return false for pure Python projects', () => {
      const config = { runtime: { name: 'python' }, techStack: ['django'] };

      const result = detector.shouldCreatePackageJson(config);

      expect(result).toBe(false);
    });
  });
});

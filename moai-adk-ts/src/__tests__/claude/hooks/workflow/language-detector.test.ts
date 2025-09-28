/**
import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';
 * @file language-detector.test.ts
 * @description Tests for LanguageDetector hook
 */

import { LanguageDetector } from '../../../../claude/hooks/workflow/language-detector';
import { HookInput } from '../../../../claude/hooks/types';
import * as fs from 'fs';
import * as path from 'path';

// Mock filesystem
vi.mock('fs');

const mockFs = fs as vi.Mocked<typeof fs>;

describe('LanguageDetector', () => {
  let languageDetector: LanguageDetector;
  const mockProjectRoot = '/test/project';

  beforeEach(() => {
    languageDetector = new LanguageDetector(mockProjectRoot);

    // Reset all mocks
    vi.clearAllMocks();

    // Mock console.log to avoid output during tests
    vi.spyOn(console, 'log').mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('execute', () => {
    it('should return success when no languages detected', async () => {
      // Mock no files found
      mockFs.existsSync.mockReturnValue(false);

      const result = await languageDetector.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toBeUndefined();
    });

    it('should detect Python projects', async () => {
      mockFs.existsSync.mockImplementation((filePath: string) => {
        return filePath.includes('pyproject.toml');
      });

      const result = await languageDetector.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain('🌐 감지된 언어: python');
      expect(result.message).toContain('test=pytest');
      expect(result.message).toContain('lint=ruff');
      expect(result.message).toContain('format=black');
    });

    it('should detect JavaScript projects', async () => {
      mockFs.existsSync.mockImplementation((filePath: string) => {
        return filePath.includes('package.json');
      });

      const result = await languageDetector.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain('🌐 감지된 언어: javascript');
      expect(result.message).toContain('test=npm test');
      expect(result.message).toContain('lint=eslint');
      expect(result.message).toContain('format=prettier');
    });

    it('should detect TypeScript projects', async () => {
      mockFs.existsSync.mockImplementation((filePath: string) => {
        return filePath.includes('tsconfig.json');
      });

      const result = await languageDetector.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain('typescript');
      expect(result.message).toContain('test=npm test');
      expect(result.message).toContain('lint=eslint');
      expect(result.message).toContain('format=prettier');
    });

    it('should detect Go projects', async () => {
      mockFs.existsSync.mockImplementation((filePath: string) => {
        return filePath.includes('go.mod');
      });

      const result = await languageDetector.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain('go');
      expect(result.message).toContain('test=go test ./...');
      expect(result.message).toContain('lint=golangci-lint');
      expect(result.message).toContain('format=gofmt');
    });

    it('should detect Rust projects', async () => {
      mockFs.existsSync.mockImplementation((filePath: string) => {
        return filePath.includes('Cargo.toml');
      });

      const result = await languageDetector.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain('rust');
      expect(result.message).toContain('test=cargo test');
      expect(result.message).toContain('lint=cargo clippy');
      expect(result.message).toContain('format=rustfmt');
    });

    it('should detect Java projects', async () => {
      mockFs.existsSync.mockImplementation((filePath: string) => {
        return (
          filePath.includes('pom.xml') || filePath.includes('build.gradle')
        );
      });

      const result = await languageDetector.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain('java');
      expect(result.message).toContain('test=gradle test | mvn test');
    });

    it('should detect C# projects', async () => {
      // Mock readdirSync to simulate finding .csproj files
      mockFs.readdirSync.mockImplementation(() => {
        return ['project.csproj'] as any;
      });

      mockFs.statSync.mockImplementation(() => {
        return { isDirectory: () => false, isFile: () => true } as any;
      });

      const result = await languageDetector.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain('csharp');
      expect(result.message).toContain('test=dotnet test');
    });

    it('should detect multiple languages', async () => {
      mockFs.existsSync.mockImplementation((filePath: string) => {
        return (
          filePath.includes('package.json') ||
          filePath.includes('pyproject.toml')
        );
      });

      const result = await languageDetector.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain('python, javascript');
      expect(result.data).toBeDefined();
      expect(result.data.languages).toHaveLength(2);
    });

    it('should include helpful message about TDD usage', async () => {
      mockFs.existsSync.mockImplementation((filePath: string) => {
        return filePath.includes('pyproject.toml');
      });

      const result = await languageDetector.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain(
        '💡 필요 시 /moai:2-build 단계에서 해당 도구를 사용해 TDD를 실행하세요'
      );
    });

    it('should handle file system errors gracefully', async () => {
      mockFs.existsSync.mockImplementation(() => {
        throw new Error('File system error');
      });

      const result = await languageDetector.execute({});

      expect(result.success).toBe(true);
    });
  });

  describe('file detection with patterns', () => {
    beforeEach(() => {
      // Mock readdirSync for pattern-based detection
      mockFs.readdirSync.mockImplementation((dir: string) => {
        const dirStr = dir.toString();
        if (dirStr.includes('test/project')) {
          return ['src', 'tests', 'package.json'] as any;
        }
        if (dirStr.includes('src')) {
          return ['main.py', 'utils.py'] as any;
        }
        return [] as any;
      });

      mockFs.statSync.mockImplementation((path: string) => {
        const pathStr = path.toString();
        if (pathStr.includes('src') || pathStr.includes('tests')) {
          return { isDirectory: () => true, isFile: () => false } as any;
        }
        return { isDirectory: () => false, isFile: () => true } as any;
      });
    });

    it('should detect Python files in directory structure', async () => {
      mockFs.existsSync.mockImplementation((filePath: string) => {
        return !filePath.includes('pyproject.toml'); // Test file-based detection
      });

      const result = await languageDetector.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain('python');
    });

    it('should skip ignored directories', async () => {
      mockFs.readdirSync.mockImplementation((dir: string) => {
        return ['node_modules', '.git', '__pycache__', 'src'] as any;
      });

      // Should not traverse into ignored directories
      const result = await languageDetector.execute({});
      expect(result.success).toBe(true);
    });
  });

  describe('custom mappings', () => {
    it('should load custom mappings from configuration', async () => {
      const customMappings = {
        test_runners: {
          python: 'custom-pytest',
        },
        linters: {
          python: 'custom-linter',
        },
        formatters: {
          python: 'custom-formatter',
        },
      };

      mockFs.existsSync.mockImplementation((filePath: string) => {
        return (
          filePath.includes('pyproject.toml') ||
          filePath.includes('.moai/config/language_mappings.json')
        );
      });

      mockFs.readFileSync.mockImplementation((filePath: string) => {
        if (filePath.toString().includes('language_mappings.json')) {
          return JSON.stringify(customMappings);
        }
        return '';
      });

      const result = await languageDetector.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain('test=custom-pytest');
      expect(result.message).toContain('lint=custom-linter');
      expect(result.message).toContain('format=custom-formatter');
    });

    it('should fallback to defaults when custom mappings fail to load', async () => {
      mockFs.existsSync.mockImplementation((filePath: string) => {
        return (
          filePath.includes('pyproject.toml') ||
          filePath.includes('.moai/config/language_mappings.json')
        );
      });

      mockFs.readFileSync.mockImplementation(() => {
        throw new Error('Failed to read file');
      });

      const result = await languageDetector.execute({});

      expect(result.success).toBe(true);
      expect(result.message).toContain('test=pytest'); // Default mapping
    });
  });

  describe('getLanguages method', () => {
    it('should return structured language information', () => {
      mockFs.existsSync.mockImplementation((filePath: string) => {
        return filePath.includes('pyproject.toml');
      });

      const languages = languageDetector.getLanguages();

      expect(languages).toHaveLength(1);
      expect(languages[0]).toEqual({
        language: 'python',
        confidence: 0.85,
        testRunner: 'pytest',
        linter: 'ruff',
        formatter: 'black',
      });
    });
  });
});

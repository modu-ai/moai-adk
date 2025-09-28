/**
 * @file language-detector.ts
 * @description Language detection hook for MoAI-ADK projects
 * @version 1.0.0
 * @tag @FEATURE:LANGUAGE-DETECT-013
 */

import type { HookInput, HookResult, MoAIHook, LanguageInfo } from '../types';
import * as fs from 'fs';
import * as path from 'path';

/**
 * Language detection mappings
 */
interface LanguageMappings {
  test_runners: Record<string, string>;
  linters: Record<string, string>;
  formatters: Record<string, string>;
}

/**
 * Default tool mappings for each language
 */
const DEFAULT_MAPPINGS: LanguageMappings = {
  test_runners: {
    python: 'pytest',
    javascript: 'npm test',
    typescript: 'npm test',
    go: 'go test ./...',
    rust: 'cargo test',
    java: 'gradle test | mvn test',
    csharp: 'dotnet test',
    cpp: 'ctest | make test',
  },
  linters: {
    python: 'ruff',
    javascript: 'eslint',
    typescript: 'eslint',
    go: 'golangci-lint',
    rust: 'cargo clippy',
    java: 'checkstyle',
    csharp: 'dotnet format',
    cpp: 'clang-tidy',
  },
  formatters: {
    python: 'black',
    javascript: 'prettier',
    typescript: 'prettier',
    go: 'gofmt',
    rust: 'rustfmt',
    java: 'google-java-format',
    csharp: 'dotnet format',
    cpp: 'clang-format',
  },
};

/**
 * Language Detector Hook - TypeScript port of language_detector.py
 */
export class LanguageDetector implements MoAIHook {
  name = 'language-detector';

  private projectRoot: string;

  constructor(projectRoot?: string) {
    this.projectRoot = projectRoot || process.cwd();
  }

  async execute(input: HookInput): Promise<HookResult> {
    try {
      const languages = this.detectProjectLanguages();

      if (languages.length === 0) {
        return { success: true };
      }

      const mappings = this.loadMappings();
      const output = this.generateOutput(languages, mappings);

      return {
        success: true,
        message: output,
        data: {
          languages: languages.map(lang => ({
            language: lang,
            confidence: this.calculateConfidence(lang),
            testRunner: mappings.test_runners[lang] || '-',
            linter: mappings.linters[lang] || '-',
            formatter: mappings.formatters[lang] || '-',
          })),
        },
      };
    } catch (error) {
      // Silent failure to avoid breaking Claude Code session
      return { success: true };
    }
  }

  /**
   * Detect programming languages in the project
   */
  private detectProjectLanguages(): string[] {
    const rootPath = path.resolve(this.projectRoot);
    const languages: string[] = [];

    // Python detection
    if (this.hasFile('pyproject.toml') || this.hasFiles('**/*.py')) {
      languages.push('python');
    }

    // JavaScript/TypeScript detection
    if (this.hasFile('package.json') || this.hasFiles('**/*.{js,jsx}')) {
      languages.push('javascript');
    }

    if (this.hasFiles('**/*.{ts,tsx}') || this.hasFile('tsconfig.json')) {
      if (!languages.includes('typescript')) {
        languages.push('typescript');
      }
    }

    // Go detection
    if (this.hasFile('go.mod') || this.hasFiles('**/*.go')) {
      languages.push('go');
    }

    // Rust detection
    if (this.hasFile('Cargo.toml') || this.hasFiles('**/*.rs')) {
      languages.push('rust');
    }

    // Java detection
    if (
      this.hasFile('pom.xml') ||
      this.hasFile('build.gradle') ||
      this.hasFile('build.gradle.kts') ||
      this.hasFiles('**/*.java')
    ) {
      languages.push('java');
    }

    // C# detection
    if (
      this.hasFiles('**/*.sln') ||
      this.hasFiles('**/*.csproj') ||
      this.hasFiles('**/*.cs')
    ) {
      languages.push('csharp');
    }

    // C++ detection
    if (this.hasFiles('**/*.{c,cpp,cxx}') || this.hasFile('CMakeLists.txt')) {
      languages.push('cpp');
    }

    // Remove duplicates while preserving order
    return Array.from(new Set(languages));
  }

  /**
   * Check if a specific file exists
   */
  private hasFile(filename: string): boolean {
    return fs.existsSync(path.join(this.projectRoot, filename));
  }

  /**
   * Check if files matching pattern exist (simplified implementation)
   */
  private hasFiles(pattern: string): boolean {
    try {
      // Extract extension from pattern
      const extensionMatch = pattern.match(/\*\*\/?\*\.(\w+)/);
      if (!extensionMatch) {
        return false;
      }

      const extension = extensionMatch[1];
      return this.findFilesWithExtension(this.projectRoot, extension);
    } catch {
      return false;
    }
  }

  /**
   * Recursively find files with specific extension
   */
  private findFilesWithExtension(dir: string, extension: string): boolean {
    try {
      const entries = fs.readdirSync(dir, { withFileTypes: true });

      for (const entry of entries) {
        const fullPath = path.join(dir, entry.name);

        if (entry.isDirectory()) {
          // Skip common ignore directories
          if (
            [
              'node_modules',
              '.git',
              '__pycache__',
              '.pytest_cache',
              'dist',
              'build',
            ].includes(entry.name)
          ) {
            continue;
          }

          if (this.findFilesWithExtension(fullPath, extension)) {
            return true;
          }
        } else if (entry.isFile()) {
          if (entry.name.endsWith(`.${extension}`)) {
            return true;
          }
        }
      }

      return false;
    } catch {
      return false;
    }
  }

  /**
   * Load language mappings from configuration
   */
  private loadMappings(): LanguageMappings {
    const mappingPath = path.join(
      this.projectRoot,
      '.moai',
      'config',
      'language_mappings.json'
    );

    try {
      if (fs.existsSync(mappingPath)) {
        const data = fs.readFileSync(mappingPath, 'utf-8');
        return JSON.parse(data) as LanguageMappings;
      }
    } catch (error) {
      // Fall back to defaults
    }

    return DEFAULT_MAPPINGS;
  }

  /**
   * Calculate confidence score for language detection
   */
  private calculateConfidence(language: string): number {
    // Simplified confidence calculation
    // In a real implementation, this would consider file counts, project structure, etc.
    return 0.85;
  }

  /**
   * Generate human-readable output
   */
  private generateOutput(
    languages: string[],
    mappings: LanguageMappings
  ): string {
    const lines: string[] = [];

    lines.push(`🌐 감지된 언어: ${languages.join(', ')}`);

    if (languages.length > 0) {
      lines.push('🔧 권장 도구:');

      for (const lang of languages) {
        const testRunner = mappings.test_runners[lang] || '-';
        const linter = mappings.linters[lang] || '-';
        const formatter = mappings.formatters[lang] || '-';

        lines.push(
          `- ${lang}: test=${testRunner}, lint=${linter}, format=${formatter}`
        );
      }

      lines.push(
        '💡 필요 시 /moai:2-build 단계에서 해당 도구를 사용해 TDD를 실행하세요.'
      );
    }

    return lines.join('\n');
  }

  /**
   * Get detected languages as structured data
   */
  getLanguages(): LanguageInfo[] {
    const languages = this.detectProjectLanguages();
    const mappings = this.loadMappings();

    return languages.map(lang => ({
      language: lang,
      confidence: this.calculateConfidence(lang),
      testRunner: mappings.test_runners[lang],
      linter: mappings.linters[lang],
      formatter: mappings.formatters[lang],
    }));
  }
}

/**
 * CLI entry point for Claude Code compatibility
 */
export async function main(): Promise<void> {
  try {
    const detector = new LanguageDetector();
    const result = await detector.execute({});

    if (result.message) {
      console.log(result.message);
    }
  } catch (error) {
    // Silent failure to avoid breaking Claude Code session
  }
}

// Execute if run directly
if (require.main === module) {
  main().catch(() => {
    // Silent failure
  });
}

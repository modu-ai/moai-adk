/**
 * @FEATURE-FILE-ANALYZER-001: TypeScript FileAnalyzer 구현
 * @연결: @TASK:ANALYSIS-001 → @FEATURE:FILE-ANALYSIS-001 → @API:ANALYZE-FILES
 */

import type { FileChange } from '../commit-helper';

export interface FileAnalysis {
  total_files: number;
  categories: {
    python: number;
    typescript: number;
    docs: number;
    config: number;
    tests: number;
    other: number;
  };
  is_simple_change: boolean;
  has_tests: boolean;
  is_documentation_only: boolean;
  summary?: {
    modified: number;
    added: number;
    deleted: number;
    renamed: number;
  };
}

export class FileAnalyzer {
  classifyFileChange(status: string): string {
    const statusCode = status.trim().charAt(0);

    switch (statusCode) {
      case 'M':
        return 'modified';
      case 'A':
        return 'added';
      case 'D':
        return 'deleted';
      case 'R':
        return 'renamed';
      case '?':
        return 'untracked';
      default:
        return 'unknown';
    }
  }

  analyzeFileChanges(files: FileChange[]): FileAnalysis {
    const categories = {
      python: 0,
      typescript: 0,
      docs: 0,
      config: 0,
      tests: 0,
      other: 0,
    };

    const summary = {
      modified: 0,
      added: 0,
      deleted: 0,
      renamed: 0,
    };

    let hasTests = false;
    let isDocumentationOnly = true;

    for (const file of files) {
      // 파일 카테고리 분류
      const category = this.categorizeFile(file.filename);
      categories[category]++;

      // 변경 유형 분류
      const changeType = file.type as keyof typeof summary;
      if (changeType in summary) {
        summary[changeType]++;
      }

      // 테스트 파일 검사
      if (this.isTestFile(file.filename)) {
        hasTests = true;
      }

      // 문서 전용 변경 검사
      if (!this.isDocumentationFile(file.filename)) {
        isDocumentationOnly = false;
      }
    }

    // 단순 변경 판단: 3개 이하 파일, 단일 카테고리
    const totalFiles = files.length;
    const nonZeroCategories = Object.values(categories).filter(
      count => count > 0
    ).length;
    const isSimpleChange = totalFiles <= 3 && nonZeroCategories <= 2;

    return {
      total_files: totalFiles,
      categories,
      is_simple_change: isSimpleChange,
      has_tests: hasTests,
      is_documentation_only: isDocumentationOnly && totalFiles > 0,
      summary,
    };
  }

  suggestCommitType(analysis: FileAnalysis): string {
    if (analysis.is_documentation_only) {
      return 'docs';
    }

    if (
      analysis.categories.tests > 0 &&
      analysis.categories.tests === analysis.total_files
    ) {
      return 'test';
    }

    if (analysis.categories.config > 0) {
      return 'chore';
    }

    if (analysis.summary?.added && analysis.summary.added > 0) {
      return 'feat';
    }

    if (analysis.summary?.modified && analysis.summary.modified > 0) {
      return 'fix';
    }

    return 'chore';
  }

  private categorizeFile(filename: string): keyof FileAnalysis['categories'] {
    const ext = this.getFileExtension(filename);

    // 테스트 파일
    if (this.isTestFile(filename)) {
      return 'tests';
    }

    // 문서 파일
    if (this.isDocumentationFile(filename)) {
      return 'docs';
    }

    // 설정 파일
    if (this.isConfigFile(filename)) {
      return 'config';
    }

    // 언어별 분류
    switch (ext) {
      case '.py':
        return 'python';
      case '.ts':
      case '.tsx':
      case '.js':
      case '.jsx':
        return 'typescript';
      default:
        return 'other';
    }
  }

  private isTestFile(filename: string): boolean {
    const basename = filename.toLowerCase();
    return (
      basename.includes('test') ||
      basename.includes('spec') ||
      basename.includes('__tests__') ||
      filename.endsWith('.test.ts') ||
      filename.endsWith('.test.js') ||
      filename.endsWith('.spec.ts')
    );
  }

  private isDocumentationFile(filename: string): boolean {
    const ext = this.getFileExtension(filename);
    const basename = filename.toLowerCase();

    return (
      ext === '.md' ||
      ext === '.rst' ||
      ext === '.txt' ||
      basename.includes('readme') ||
      basename.includes('changelog') ||
      basename.includes('docs/')
    );
  }

  private isConfigFile(filename: string): boolean {
    const basename = filename.toLowerCase();
    const configFiles = [
      'package.json',
      'tsconfig.json',
      'jest.config',
      'eslint',
      'prettier',
      '.env',
      'docker',
      'makefile',
      'pyproject.toml',
      'setup.py',
      'requirements.txt',
    ];

    return configFiles.some(config => basename.includes(config));
  }

  private getFileExtension(filename: string): string {
    const lastDot = filename.lastIndexOf('.');
    return lastDot !== -1 ? filename.substring(lastDot) : '';
  }
}

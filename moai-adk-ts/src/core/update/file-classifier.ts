/**
 * @file File classification logic for update strategy
 * @author MoAI Team
 * @tags @FEATURE:UPDATE-STRATEGY-001 | Chain: @REQ:UPDATE-REAL-001 -> @DESIGN:UPDATE-STRATEGY-001 -> @TASK:FILE-CLASSIFIER-001 -> @TEST:FILE-CLASSIFIER-001
 * Related: @SEC:UPDATE-STRATEGY-001, @DOCS:UPDATE-STRATEGY-001
 */

import * as path from 'node:path';
import { type FilePatterns, FileType } from './types.js';

/**
 * File classifier for determining file types during updates
 * @tags @FEATURE:FILE-CLASSIFIER-001
 */
export class FileClassifier {
  private readonly patterns: FilePatterns;

  constructor() {
    this.patterns = this.initializePatterns();
  }

  /**
   * Classify file based on its path and content patterns
   * @param filePath - Relative file path from project root
   * @returns File type classification
   * @tags @API:CLASSIFY-FILE-001
   */
  public classifyFile(filePath: string): FileType {
    const normalizedPath = path.posix.normalize(filePath);

    // Check patterns in order of specificity
    if (this.matchesPatterns(normalizedPath, this.patterns.metadata)) {
      return FileType.METADATA;
    }

    if (this.matchesPatterns(normalizedPath, this.patterns.generated)) {
      return FileType.GENERATED;
    }

    if (this.matchesPatterns(normalizedPath, this.patterns.user)) {
      return FileType.USER;
    }

    if (this.matchesPatterns(normalizedPath, this.patterns.hybrid)) {
      return FileType.HYBRID;
    }

    if (this.matchesPatterns(normalizedPath, this.patterns.template)) {
      return FileType.TEMPLATE;
    }

    // Default to USER for unknown files (safer approach)
    return FileType.USER;
  }

  /**
   * Initialize file pattern definitions
   * @returns File patterns configuration
   * @tags @UTIL:PATTERN-INIT-001
   */
  private initializePatterns(): FilePatterns {
    return {
      // Template files - MoAI-ADK managed
      template: [
        '.claude/agents/moai/**',
        '.claude/commands/moai/**',
        '.claude/hooks/moai/**',
        '.claude/output-styles/**',
        '.moai/memory/development-guide.md',
      ],

      // User files - never touch
      user: [
        '.env*',
        '*.local.*',
        'src/**',
        'tests/**',
        '__tests__/**',
        'spec/**',
        'docs/user/**',
        'README.md',
        'LICENSE*',
        '.gitignore',
      ],

      // Hybrid files - intelligent merge required
      hybrid: [
        'CLAUDE.md',
        '.moai/project/**',
        'package.json',
        'pyproject.toml',
        'Cargo.toml',
        'pom.xml',
        'go.mod',
      ],

      // Generated files - can be recreated
      generated: [
        '.moai/reports/**',
        'node_modules/**',
        'dist/**',
        'build/**',
        '__pycache__/**',
        'target/**',
        '.mypy_cache/**',
      ],

      // Metadata files - special handling
      metadata: [
        '.moai/version.json',
        '.moai/config.json',
        '.claude/settings.local.json',
      ],
    };
  }

  /**
   * Check if file path matches any pattern in list
   * @param filePath - File path to check
   * @param patterns - Glob patterns to match against
   * @returns True if path matches any pattern
   * @tags @UTIL:PATTERN-MATCH-001
   */
  private matchesPatterns(
    filePath: string,
    patterns: readonly string[]
  ): boolean {
    return patterns.some(pattern => {
      // Convert glob pattern to regex for simple matching
      const regexPattern = pattern
        .replace(/\*\*/g, '.*')
        .replace(/\*/g, '[^/]*')
        .replace(/\?/g, '[^/]');

      const regex = new RegExp(`^${regexPattern}$`);
      return regex.test(filePath);
    });
  }
}

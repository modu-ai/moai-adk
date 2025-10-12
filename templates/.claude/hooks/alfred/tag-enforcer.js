#!/usr/bin/env node
'use strict';

/**
 * @CODE:REFACTOR-003 | @CODE:HOOKS-001 |
 * SPEC: .moai/specs/SPEC-HOOKS-001/spec.md
 * TEST: __tests__/claude/hooks/tag-enforcer/
 *
 * TAG Enforcer Hook (Refactored)
 * @CODE:REFACTOR-003:API - 코드 내 TAG 블록 불변성 보장 및 8-Core TAG 체계 검증
 *
 * Pure JavaScript implementation for cross-platform compatibility.
 */

const fs = require('node:fs/promises');
const path = require('node:path');

// ============================================================================
// Type Definitions (JSDoc)
// ============================================================================

/**
 * Hook input from Claude Code
 *
 * @typedef {Object} HookInput
 * @property {string} tool_name - Name of the tool being invoked
 * @property {Object<string, any>} [tool_input] - Tool input parameters
 * @property {Object} [context] - Execution context
 */

/**
 * Hook execution result
 *
 * @typedef {Object} HookResult
 * @property {boolean} success - Whether the hook execution succeeded
 * @property {boolean} [blocked] - Whether the operation was blocked
 * @property {string} [message] - Result message
 * @property {number} [exitCode] - Exit code
 * @property {Object<string, any>} [data] - Additional data
 * @property {string[]} [warnings] - Warning messages
 */

/**
 * TAG block extraction result
 *
 * @typedef {Object} TagBlock
 * @property {string} content - TAG block content
 * @property {number} lineNumber - Starting line number
 */

/**
 * Immutability check result
 *
 * @typedef {Object} ImmutabilityCheck
 * @property {boolean} violated - Whether immutability was violated
 * @property {string} [modifiedTag] - Modified TAG identifier
 * @property {string} [violationDetails] - Details of the violation
 */

/**
 * TAG validation result
 *
 * @typedef {Object} ValidationResult
 * @property {boolean} isValid - Whether TAG is valid
 * @property {string[]} violations - List of violations
 * @property {string[]} warnings - List of warnings
 * @property {boolean} hasTag - Whether TAG block exists
 */

// ============================================================================
// Constants (Inlined from constants.ts)
// ============================================================================

/**
 * Supported programming languages and their file extensions
 */
const SUPPORTED_LANGUAGES = {
  typescript: ['.ts', '.tsx'],
  javascript: ['.js', '.jsx', '.mjs', '.cjs'],
  python: ['.py', '.pyi'],
  java: ['.java'],
  go: ['.go'],
  rust: ['.rs'],
  cpp: ['.cpp', '.hpp', '.cc', '.h', '.cxx', '.hxx'],
  ruby: ['.rb', '.rake', '.gemspec'],
  php: ['.php'],
  csharp: ['.cs'],
  dart: ['.dart'],
  swift: ['.swift'],
  kotlin: ['.kt', '.kts'],
  elixir: ['.ex', '.exs'],
  markdown: ['.md', '.mdx'],
};

/**
 * Paths that should be excluded from TAG enforcement
 */
const EXCLUDED_PATHS = [
  'node_modules',
  '.git',
  'dist',
  'build',
  'test',
  'spec',
  '__test__',
  '__tests__',
];

// ============================================================================
// TAG Patterns (Inlined from tag-patterns.ts)
// ============================================================================

/**
 * Code-First TAG patterns
 */
const CODE_FIRST_PATTERNS = {
  // 전체 TAG 블록 매칭 (파일 최상단)
  TAG_BLOCK: /^\/\*\*\s*([\s\S]*?)\*\//m,

  // 핵심 TAG 라인들
  MAIN_TAG: /^\s*\*\s*@DOC:([A-Z]+):([A-Z0-9_-]+)\s*$/m,
  CHAIN_LINE: /^\s*\*\s*CHAIN:\s*(.+)\s*$/m,
  DEPENDS_LINE: /^\s*\*\s*DEPENDS:\s*(.+)\s*$/m,
  STATUS_LINE: /^\s*\*\s*STATUS:\s*(\w+)\s*$/m,
  CREATED_LINE: /^\s*\*\s*CREATED:\s*(\d{4}-\d{2}-\d{2})\s*$/m,
  IMMUTABLE_MARKER: /^\s*\*\s*@IMMUTABLE\s*$/m,

  // TAG 참조
  TAG_REFERENCE: /@([A-Z]+):([A-Z0-9-]+)/g,
};

/**
 * 8-Core TAG categories
 */
const VALID_CATEGORIES = {
  // Lifecycle (필수 체인)
  lifecycle: ['SPEC', 'REQ', 'DESIGN', 'TASK', 'TEST'],

  // Implementation (선택적)
  implementation: ['FEATURE', 'API', 'FIX'],
};

// ============================================================================
// Utility Functions (Inlined from utils.ts)
// ============================================================================

/**
 * Extract file path from tool input
 *
 * @param {Object<string, any>} toolInput - Tool input object
 * @returns {string|null} File path or null if not found
 */
function extractFilePath(toolInput) {
  return (
    toolInput.file_path ||
    toolInput.filePath ||
    toolInput.path ||
    toolInput.notebook_path ||
    null
  );
}

/**
 * Get all file extensions from supported languages
 *
 * @returns {string[]} Array of all file extensions
 */
function getAllFileExtensions() {
  return Object.values(SUPPORTED_LANGUAGES).flat();
}

// ============================================================================
// CLI Utilities (Inlined from index.ts)
// ============================================================================

/**
 * Parse input from stdin for Claude Code hooks
 *
 * @returns {Promise<HookInput>} Parsed hook input
 */
async function parseClaudeInput() {
  return new Promise((resolve, reject) => {
    let data = '';

    process.stdin.setEncoding('utf8');

    process.stdin.on('data', (chunk) => {
      data += chunk;
    });

    process.stdin.on('end', () => {
      try {
        if (!data.trim()) {
          resolve({
            tool_name: 'Unknown',
            tool_input: {},
            context: {},
          });
          return;
        }

        const parsed = JSON.parse(data);
        resolve(parsed);
      } catch (error) {
        reject(
          new Error(
            `Failed to parse input: ${error instanceof Error ? error.message : 'Unknown error'}`
          )
        );
      }
    });

    process.stdin.on('error', (error) => {
      reject(new Error(`Failed to read stdin: ${error.message}`));
    });
  });
}

/**
 * Output hook result to stdout
 *
 * @param {HookResult} result - Hook execution result
 */
function outputResult(result) {
  if (result.blocked) {
    console.error(`BLOCKED: ${result.message || 'Operation blocked'}`);
    if (result.data?.suggestions) {
      console.error(`\n${result.data.suggestions}`);
    }
    process.exit(result.exitCode || 2);
  } else if (!result.success) {
    console.error(`ERROR: ${result.message || 'Operation failed'}`);
    if (result.warnings && result.warnings.length > 0) {
      console.error(`Warnings:\n${result.warnings.join('\n')}`);
    }
    process.exit(result.exitCode || 1);
  } else {
    if (result.message) {
      console.log(result.message);
    }
    if (result.warnings && result.warnings.length > 0) {
      console.warn(`Warnings:\n${result.warnings.join('\n')}`);
    }
    process.exit(0);
  }
}

// ============================================================================
// TAG Validator Class (Inlined from tag-validator.ts)
// ============================================================================

/**
 * TAG validation class
 */
class TagValidator {
  /**
   * Extract TAG block from file content (top of file only)
   *
   * @param {string} content - File content
   * @returns {TagBlock|null} TAG block or null if not found
   */
  extractTagBlock(content) {
    const lines = content.split('\n');
    let inBlock = false;
    let blockLines = [];
    let startLineNumber = 0;

    for (let i = 0; i < Math.min(lines.length, 30); i++) {
      const line = lines[i]?.trim();

      // Skip empty lines or shebang
      if (!line || line.startsWith('#!')) {
        continue;
      }

      // TAG block start
      if (line.startsWith('/**') && !inBlock) {
        inBlock = true;
        blockLines = [line];
        startLineNumber = i + 1;
        continue;
      }

      // Inside TAG block
      if (inBlock) {
        blockLines.push(line);

        // TAG block end
        if (line.endsWith('*/')) {
          const blockContent = blockLines.join('\n');

          // Check if block contains @TAG
          if (CODE_FIRST_PATTERNS.MAIN_TAG.test(blockContent)) {
            return {
              content: blockContent,
              lineNumber: startLineNumber,
            };
          }

          // No @TAG, reset and continue
          inBlock = false;
          blockLines = [];
          continue;
        }
      }

      // Stop if non-TAG code starts
      if (
        !inBlock &&
        line &&
        !line.startsWith('//') &&
        !line.startsWith('/*')
      ) {
        break;
      }
    }

    return null;
  }

  /**
   * Extract main TAG from block content
   *
   * @param {string} blockContent - TAG block content
   * @returns {string} Main TAG identifier
   */
  extractMainTag(blockContent) {
    const match = CODE_FIRST_PATTERNS.MAIN_TAG.exec(blockContent);
    return match ? `@${match[1]}:${match[2]}` : 'UNKNOWN';
  }

  /**
   * Normalize TAG block for comparison
   *
   * @param {string} blockContent - TAG block content
   * @returns {string} Normalized content
   */
  normalizeTagBlock(blockContent) {
    return blockContent
      .split('\n')
      .map((line) => line.trim())
      .filter((line) => line.length > 0)
      .join('\n');
  }

  /**
   * Validate Code-First TAG
   *
   * @param {string} content - File content
   * @returns {ValidationResult} Validation result
   */
  validateCodeFirstTag(content) {
    const violations = [];
    const warnings = [];
    let hasTag = false;

    // 1. Extract TAG block
    const tagBlock = this.extractTagBlock(content);
    if (!tagBlock) {
      return {
        isValid: true, // No TAG block is OK (recommendation only)
        violations: [],
        warnings: ['파일 최상단에 TAG 블록이 없습니다 (권장사항)'],
        hasTag: false,
      };
    }

    hasTag = true;
    const blockContent = tagBlock.content;

    // 2. Validate main TAG
    const tagMatch = CODE_FIRST_PATTERNS.MAIN_TAG.exec(blockContent);
    if (!tagMatch) {
      violations.push('@TAG 라인이 발견되지 않았습니다');
    } else {
      const [, category, domainId] = tagMatch;

      // Category validation
      const allValidCategories = [
        ...VALID_CATEGORIES.lifecycle,
        ...VALID_CATEGORIES.implementation,
      ];
      const validCategorySet = new Set(allValidCategories);
      if (category && !validCategorySet.has(category)) {
        violations.push(`유효하지 않은 TAG 카테고리: ${category}`);
      }

      // Domain ID format check
      if (domainId && !/^[A-Z0-9]+-\d{3,}$/.test(domainId)) {
        warnings.push(`도메인 ID 형식 권장: ${domainId} -> DOMAIN-001`);
      }
    }

    // 3. Chain validation
    const chainMatch = CODE_FIRST_PATTERNS.CHAIN_LINE.exec(blockContent);
    if (chainMatch) {
      const chainStr = chainMatch[1];
      if (chainStr) {
        const chainTags = chainStr.split(/\s*->\s*/);

        for (const chainTag of chainTags) {
          if (!CODE_FIRST_PATTERNS.TAG_REFERENCE.test(chainTag.trim())) {
            warnings.push(`체인의 TAG 형식을 확인하세요: ${chainTag.trim()}`);
          }
        }
      }
    }

    // 4. Dependencies validation
    const dependsMatch = CODE_FIRST_PATTERNS.DEPENDS_LINE.exec(blockContent);
    if (dependsMatch) {
      const dependsStr = dependsMatch[1];
      if (dependsStr && dependsStr.trim().toLowerCase() !== 'none') {
        const dependsTags = dependsStr.split(/,\s*/);

        for (const dependTag of dependsTags) {
          if (!CODE_FIRST_PATTERNS.TAG_REFERENCE.test(dependTag.trim())) {
            warnings.push(`의존성 TAG 형식을 확인하세요: ${dependTag.trim()}`);
          }
        }
      }
    }

    // 5. Status validation
    const statusMatch = CODE_FIRST_PATTERNS.STATUS_LINE.exec(blockContent);
    if (statusMatch) {
      const status = statusMatch[1]?.toLowerCase();
      if (status && !['active', 'deprecated', 'completed'].includes(status)) {
        warnings.push(`알 수 없는 STATUS: ${status}`);
      }
    }

    // 6. Created date validation
    const createdMatch = CODE_FIRST_PATTERNS.CREATED_LINE.exec(blockContent);
    if (createdMatch) {
      const created = createdMatch[1];
      if (created && !/^\d{4}-\d{2}-\d{2}$/.test(created)) {
        warnings.push(`생성 날짜 형식을 확인하세요: ${created} (YYYY-MM-DD)`);
      }
    }

    // 7. @IMMUTABLE marker recommendation
    if (!CODE_FIRST_PATTERNS.IMMUTABLE_MARKER.test(blockContent)) {
      warnings.push(
        '@IMMUTABLE 마커를 추가하여 TAG 불변성을 보장하는 것을 권장합니다'
      );
    }

    return {
      isValid: violations.length === 0,
      violations,
      warnings,
      hasTag,
    };
  }
}

// ============================================================================
// Code-First TAG Enforcer Hook Implementation
// ============================================================================

/**
 * Code-First TAG Enforcer Hook - Pure JavaScript port
 */
class CodeFirstTAGEnforcer {
  constructor() {
    this.name = 'tag-enforcer';
    this.validator = new TagValidator();
  }

  /**
   * Execute the TAG enforcer hook
   *
   * @param {HookInput} input - Hook input from Claude Code
   * @returns {Promise<HookResult>} Hook execution result
   */
  async execute(input) {
    try {
      // 1. Check if write operation
      if (!this.isWriteOperation(input?.tool_name)) {
        return { success: true };
      }

      const filePath = extractFilePath(input.tool_input || {});
      if (!filePath || !this.shouldEnforceTags(filePath)) {
        return { success: true };
      }

      // 2. Extract old and new content
      const oldContent = await this.getOriginalFileContent(filePath);
      const newContent = this.extractFileContent(input.tool_input || {});

      // 3. Check @IMMUTABLE TAG block modification
      const immutabilityCheck = this.checkImmutability(
        oldContent,
        newContent,
        filePath
      );
      if (immutabilityCheck.violated) {
        return {
          success: false,
          blocked: true,
          message: `🚫 @IMMUTABLE TAG 수정 금지: ${immutabilityCheck.violationDetails}`,
          data: {
            suggestions: this.generateImmutabilityHelp(immutabilityCheck),
          },
          exitCode: 2,
        };
      }

      // 4. Validate new TAG block
      const validation = this.validator.validateCodeFirstTag(newContent);
      if (!validation.isValid) {
        return {
          success: false,
          blocked: true,
          message: `🏷️ Code-First TAG 검증 실패: ${validation.violations.join(', ')}`,
          data: {
            suggestions: this.generateTagSuggestions(filePath, newContent),
          },
          exitCode: 2,
        };
      }

      // 5. Print warnings (non-blocking)
      if (validation.warnings.length > 0) {
        console.error(`⚠️ TAG 개선 권장: ${validation.warnings.join(', ')}`);
      }

      return {
        success: true,
        message: validation.hasTag
          ? `✅ Code-First TAG 검증 완료`
          : `📝 TAG 블록이 없는 파일 (권장사항)`,
      };
    } catch (error) {
      // Don't block on errors, just warn
      console.error(
        `TAG Enforcer 경고: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
      return { success: true };
    }
  }

  /**
   * Check if write operation
   *
   * @param {string} [toolName] - Tool name
   * @returns {boolean} True if write operation
   */
  isWriteOperation(toolName) {
    return (
      !!toolName &&
      ['Write', 'Edit', 'MultiEdit', 'NotebookEdit'].includes(toolName)
    );
  }

  /**
   * Extract file content from tool input
   *
   * @param {Object<string, any>} toolInput - Tool input
   * @returns {string} File content
   */
  extractFileContent(toolInput) {
    if (toolInput.content) return toolInput.content;
    if (toolInput.new_string) return toolInput.new_string;
    if (toolInput.new_source) return toolInput.new_source;
    if (toolInput.edits && Array.isArray(toolInput.edits)) {
      return toolInput.edits.map((edit) => edit.new_string).join('\n');
    }
    return '';
  }

  /**
   * Check if file should be TAG enforced
   *
   * @param {string} filePath - File path
   * @returns {boolean} True if should enforce
   */
  shouldEnforceTags(filePath) {
    const enforceExtensions = getAllFileExtensions();
    const ext = path.extname(filePath);

    // Check excluded paths
    for (const excludedPath of EXCLUDED_PATHS) {
      if (filePath.includes(excludedPath)) {
        return false;
      }
    }

    return enforceExtensions.includes(ext);
  }

  /**
   * Get original file content
   *
   * @param {string} filePath - File path
   * @returns {Promise<string>} File content or empty string for new files
   */
  async getOriginalFileContent(filePath) {
    try {
      return await fs.readFile(filePath, 'utf-8');
    } catch (_error) {
      // New file, return empty string
      return '';
    }
  }

  /**
   * Check @IMMUTABLE TAG block modification
   *
   * @param {string} oldContent - Original content
   * @param {string} newContent - New content
   * @param {string} _filePath - File path (unused)
   * @returns {ImmutabilityCheck} Immutability check result
   */
  checkImmutability(oldContent, newContent, _filePath) {
    // New file, pass
    if (!oldContent) {
      return { violated: false };
    }

    // 1. Find @IMMUTABLE TAG blocks
    const oldTagBlock = this.validator.extractTagBlock(oldContent);
    const newTagBlock = this.validator.extractTagBlock(newContent);

    // No old TAG block, pass
    if (!oldTagBlock) {
      return { violated: false };
    }

    // 2. Check @IMMUTABLE marker
    const wasImmutable = CODE_FIRST_PATTERNS.IMMUTABLE_MARKER.test(
      oldTagBlock.content
    );
    if (!wasImmutable) {
      return { violated: false };
    }

    // 3. Check if @IMMUTABLE TAG block was modified
    if (!newTagBlock) {
      return {
        violated: true,
        modifiedTag: this.validator.extractMainTag(oldTagBlock.content),
        violationDetails: '@IMMUTABLE TAG 블록이 삭제되었습니다',
      };
    }

    // 4. Compare TAG block content (normalized)
    const oldNormalized = this.validator.normalizeTagBlock(oldTagBlock.content);
    const newNormalized = this.validator.normalizeTagBlock(newTagBlock.content);

    if (oldNormalized !== newNormalized) {
      return {
        violated: true,
        modifiedTag: this.validator.extractMainTag(oldTagBlock.content),
        violationDetails: '@IMMUTABLE TAG 블록의 내용이 변경되었습니다',
      };
    }

    return { violated: false };
  }

  /**
   * Generate immutability violation help
   *
   * @param {ImmutabilityCheck} immutabilityCheck - Immutability check result
   * @returns {string} Help text
   */
  generateImmutabilityHelp(immutabilityCheck) {
    const help = [
      '🚫 @IMMUTABLE TAG 수정이 감지되었습니다.',
      '',
      '📋 Code-First TAG 규칙:',
      '• @IMMUTABLE 마커가 있는 TAG 블록은 수정할 수 없습니다',
      '• TAG는 한번 작성되면 불변(immutable)입니다',
      '• 기능 변경 시에는 새로운 TAG를 생성하세요',
      '',
      '✅ 권장 해결 방법:',
      '1. 새로운 TAG ID로 새 기능을 구현하세요',
      '   예: @DOC:FEATURE:AUTH-002',
      '2. 기존 TAG에 @DOC 마커를 추가하세요',
      '3. 새 TAG에서 이전 TAG를 참조하세요',
      '   예: REPLACES: FEATURE:AUTH-001',
      '',
      `🔍 수정 시도된 TAG: ${immutabilityCheck.modifiedTag || 'UNKNOWN'}`,
    ];

    return help.join('\n');
  }

  /**
   * Generate TAG suggestions
   *
   * @param {string} filePath - File path
   * @param {string} _content - File content (unused)
   * @returns {string} Suggestion text
   */
  generateTagSuggestions(filePath, _content) {
    const fileName = path.basename(filePath, path.extname(filePath));

    const suggestions = [
      '📝 Code-First TAG 블록 예시:',
      '',
      '```',
      '/**',
      ` * @DOC:FEATURE:${fileName.toUpperCase()}-001`,
      ` * CHAIN: REQ:${fileName.toUpperCase()}-001 -> DESIGN:${fileName.toUpperCase()}-001 -> TASK:${fileName.toUpperCase()}-001 -> TEST:${fileName.toUpperCase()}-001`,
      ' * DEPENDS: NONE',
      ' * STATUS: active',
      ` * CREATED: ${new Date().toISOString().split('T')[0]}`,
      ' * @IMMUTABLE',
      ' */',
      '```',
      '',
      '🎯 TAG 카테고리 가이드:',
      '• SPEC, REQ, DESIGN, TASK, TEST: 필수 생명주기',
      '• FEATURE, API, FIX: 구현 카테고리',
      '',
      '💡 추가 팁:',
      '• TAG 블록은 파일 최상단에 위치',
      '• @IMMUTABLE 마커로 불변성 보장',
      '• 체인으로 관련 TAG들 연결',
    ];

    return suggestions.join('\n');
  }
}

// ============================================================================
// Main Execution
// ============================================================================

/**
 * Main entry point when run directly
 */
async function main() {
  try {
    const input = await parseClaudeInput();
    const hook = new CodeFirstTAGEnforcer();
    const result = await hook.execute(input);
    outputResult(result);
  } catch (error) {
    console.error(
      `Code-First TAG Enforcer 치명적 오류: ${error instanceof Error ? error.message : 'Unknown error'}`
    );
    process.exit(0);
  }
}

// Execute if run directly
if (require.main === module) {
  main();
}

// Export for testing
module.exports = { CodeFirstTAGEnforcer, TagValidator };

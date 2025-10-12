#!/usr/bin/env node
'use strict';

/**
 * @CODE:HOOK-002 | @CODE:HOOKS-001 |
 * Related: @CODE:HOOK-002:API, @CODE:PREWRITE-001
 * SPEC: .moai/specs/SPEC-HOOKS-001/spec.md
 *
 * Pre-Write Guard Hook
 * 파일 쓰기 전 위험 패턴 검증 및 차단
 *
 * Pure JavaScript implementation for cross-platform compatibility.
 */

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
 * @property {number} [exitCode] - Exit code (0: success, 1: error, 2: blocked)
 * @property {Object<string, any>} [data] - Additional data
 * @property {string[]} [warnings] - Warning messages
 */

// ============================================================================
// Constants (Inlined from constants.ts)
// ============================================================================

/**
 * Sensitive file patterns to protect
 */
const SENSITIVE_KEYWORDS = [
  '.env',
  '/secrets',
  '/.git/',
  '/.ssh',
];

/**
 * Protected paths that should not be modified
 */
const PROTECTED_PATHS = [
  '.moai/memory/',
];

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
// Pre-Write Guard Hook Implementation
// ============================================================================

/**
 * Pre-Write Guard Hook - Pure JavaScript port
 */
class PreWriteGuard {
  constructor() {
    this.name = 'pre-write-guard';
  }

  /**
   * Execute the pre-write guard hook
   *
   * @param {HookInput} input - Hook input from Claude Code
   * @returns {Promise<HookResult>} Hook execution result
   */
  async execute(input) {
    const toolName = input?.tool_name;

    // Only check write operations
    if (!toolName || !['Write', 'Edit', 'MultiEdit'].includes(toolName)) {
      return { success: true };
    }

    const toolInput = input.tool_input || {};
    const filePath = extractFilePath(toolInput);

    if (!this.checkFileSafety(filePath || '')) {
      return {
        success: false,
        blocked: true,
        message: '민감한 파일은 편집할 수 없습니다.',
        exitCode: 2,
      };
    }

    return { success: true };
  }

  /**
   * Check if file is safe to edit
   *
   * @param {string} filePath - File path to check
   * @returns {boolean} True if file is safe to edit
   */
  checkFileSafety(filePath) {
    if (!filePath) {
      return true;
    }

    const pathLower = filePath.toLowerCase();

    // Check sensitive keywords
    for (const keyword of SENSITIVE_KEYWORDS) {
      if (pathLower.includes(keyword)) {
        return false;
      }
    }

    // Check protected paths (except templates)
    const isTemplate = filePath.includes('/templates/.moai/');
    if (!isTemplate) {
      for (const protectedPath of PROTECTED_PATHS) {
        if (filePath.includes(protectedPath)) {
          return false;
        }
      }
    }

    return true;
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
    const hook = new PreWriteGuard();
    const result = await hook.execute(input);
    outputResult(result);
  } catch (error) {
    // Silent failure to avoid breaking Claude Code session
    process.exit(0);
  }
}

// Execute if run directly
if (require.main === module) {
  main();
}

// Export for testing
module.exports = { PreWriteGuard };

#!/usr/bin/env node
'use strict';

/**
 * @CODE:HOOK-001 | @CODE:HOOKS-001 |
 * Related: @CODE:HOOK-001:API, @CODE:POLICY-001
 * SPEC: .moai/specs/SPEC-HOOKS-001/spec.md
 *
 * Policy Blocking Hook - Steering Guard
 * 위험한 프롬프트 차단 및 헌법/지침 무시 방지
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
 * @property {string} [context.working_directory] - Current working directory
 * @property {string} [context.project_root] - Project root directory
 * @property {string} [context.user] - User name
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
 * @property {string[]} [suggestions] - Suggestion messages
 */

// ============================================================================
// Constants (Inlined from constants.ts)
// ============================================================================

/**
 * Read-only tools that bypass policy checks
 */
const READ_ONLY_TOOLS = [
  'Read',
  'Glob',
  'Grep',
  'WebFetch',
  'WebSearch',
  'TodoWrite',
  'BashOutput',
  'mcp__context7__resolve-library-id',
  'mcp__context7__get-library-docs',
  'mcp__ide__getDiagnostics',
  'mcp__ide__executeCode',
];

/**
 * Dangerous commands that should always be blocked
 */
const DANGEROUS_COMMANDS = [
  'rm -rf /',
  'rm -rf --no-preserve-root',
  'sudo rm',
  'dd if=/dev/zero',
  ':(){:|:&};:',
  'mkfs.',
];

/**
 * Command prefixes that are allowed
 */
const ALLOWED_PREFIXES = [
  'git ',
  'python',
  'pytest',
  'npm ',
  'node ',
  'go ',
  'cargo ',
  'poetry ',
  'pnpm ',
  'rg ',
  'ls ',
  'cat ',
  'echo ',
  'which ',
  'make ',
  'moai ',
  'tsx ',
  'moai-ts ',
  'npx ',
  'tsc ',
  'jest ',
  'ts-node ',
  'alfred ',
  'bun ',
];

/**
 * Timeout and threshold constants
 */
const TIMEOUTS = {
  POLICY_BLOCK_SLOW_THRESHOLD: 100,
};

// ============================================================================
// Utility Functions (Inlined from utils.ts)
// ============================================================================

/**
 * Extract command from tool input
 *
 * @param {Object<string, any>} toolInput - Tool input object
 * @returns {string|null} Command string or null if not found
 */
function extractCommand(toolInput) {
  const raw = toolInput.command || toolInput.cmd;

  if (Array.isArray(raw)) {
    return raw.map(String).join(' ');
  }

  if (typeof raw === 'string') {
    return raw.trim();
  }

  return null;
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
// Policy Block Hook Implementation
// ============================================================================

/**
 * Policy Block Hook - Pure JavaScript port
 */
class PolicyBlock {
  constructor() {
    this.name = 'policy-block';
  }

  /**
   * Execute the policy blocking hook
   *
   * @param {HookInput} input - Hook input from Claude Code
   * @returns {Promise<HookResult>} Hook execution result
   */
  async execute(input) {
    const startTime = Date.now();

    // Handle missing tool_name
    if (!input || !input.tool_name) {
      return { success: true };
    }

    // Fast-track: Read-only tools bypass all checks
    if (this.isReadOnlyTool(input.tool_name)) {
      return { success: true };
    }

    // Only process Bash tool invocations
    if (input.tool_name !== 'Bash') {
      return { success: true };
    }

    const command = extractCommand(input.tool_input || {});
    if (!command) {
      return { success: true };
    }

    const commandLower = command.toLowerCase();

    // Check for dangerous commands
    for (const dangerousCommand of DANGEROUS_COMMANDS) {
      if (commandLower.includes(dangerousCommand)) {
        const duration = Date.now() - startTime;
        if (duration > TIMEOUTS.POLICY_BLOCK_SLOW_THRESHOLD) {
          console.error(`[policy-block] Blocked in ${duration}ms`);
        }
        return {
          success: false,
          blocked: true,
          message: `위험 명령이 감지되었습니다 (${dangerousCommand}).`,
          exitCode: 2,
        };
      }
    }

    // Check if command has allowed prefix
    if (!this.isAllowedPrefix(command)) {
      console.error(
        'NOTICE: 등록되지 않은 명령입니다. 필요 시 settings.json 의 allow 목록을 갱신하세요.'
      );
    }

    // Log slow executions
    const duration = Date.now() - startTime;
    if (duration > TIMEOUTS.POLICY_BLOCK_SLOW_THRESHOLD) {
      console.error(
        `[policy-block] Slow execution: ${duration}ms for ${input.tool_name}`
      );
    }

    return { success: true };
  }

  /**
   * Check if command starts with allowed prefix
   *
   * @param {string} command - Command string
   * @returns {boolean} True if command has allowed prefix
   */
  isAllowedPrefix(command) {
    return ALLOWED_PREFIXES.some((prefix) => command.startsWith(prefix));
  }

  /**
   * Check if tool is read-only and can bypass policy checks
   *
   * @param {string} toolName - Tool name
   * @returns {boolean} True if tool is read-only
   */
  isReadOnlyTool(toolName) {
    // Check for MCP tool pattern (mcp__*)
    if (toolName.startsWith('mcp__')) {
      return true;
    }

    // Check against known read-only tools
    return READ_ONLY_TOOLS.includes(toolName);
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
    const hook = new PolicyBlock();
    const result = await hook.execute(input);
    outputResult(result);
  } catch (error) {
    console.error(
      `ERROR policy_block: ${error instanceof Error ? error.message : 'Unknown error'}`
    );
    process.exit(1);
  }
}

// Execute if run directly
if (require.main === module) {
  main();
}

// Export for testing
module.exports = { PolicyBlock };

/**
 * @CODE:HOOK-CLI-001 |
 * Related: @CODE:HOOK-001, @CODE:HOOK-002, @CODE:HOOK-003, @CODE:HOOK-004
 *
 * Claude Code Hook CLI Utilities
 * 훅에서 사용하는 공통 CLI 유틸리티
 */

import type { HookInput, HookResult } from './types';

/**
 * Parse input from stdin for Claude Code hooks
 * Claude Code 훅용 stdin 파싱
 */
export async function parseClaudeInput(): Promise<HookInput> {
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
 * 훅 결과를 stdout으로 출력
 */
export function outputResult(result: HookResult): void {
  if (result.blocked) {
    console.error(`BLOCKED: ${result.message || 'Operation blocked'}`);
    if (result.data?.['suggestions']) {
      console.error(`\n${result.data['suggestions']}`);
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

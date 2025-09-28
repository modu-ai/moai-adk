/**
 * @file index.ts
 * @description MoAI-ADK Claude Code hooks system entry point
 * @version 1.0.0
 * @tag @API:HOOK-SYSTEM-013
 */

import type { HookInput, HookResult, MoAIHook, HookConfig } from './types';
import * as path from 'path';
import * as fs from 'fs';

/**
 * Hook system manager for MoAI-ADK
 */
export class HookSystem {
  private hooks: Map<string, MoAIHook> = new Map();
  private config: HookConfig;

  constructor(config?: Partial<HookConfig>) {
    this.config = {
      enabled: true,
      timeout: 10000, // 10 seconds
      disabledHooks: [],
      security: {
        allowedCommands: [
          'git',
          'npm',
          'node',
          'python',
          'pytest',
          'go',
          'cargo',
        ],
        blockedPatterns: ['rm -rf', 'sudo', 'chmod 777'],
        requireApproval: ['--force', '--hard'],
      },
      ...config,
    };
  }

  /**
   * Register a hook
   */
  registerHook(hook: MoAIHook): void {
    if (this.config.disabledHooks.includes(hook.name)) {
      return;
    }
    this.hooks.set(hook.name, hook);
  }

  /**
   * Execute a specific hook
   */
  async executeHook(hookName: string, input: HookInput): Promise<HookResult> {
    if (!this.config.enabled) {
      return { success: true, message: 'Hooks disabled' };
    }

    const hook = this.hooks.get(hookName);
    if (!hook) {
      return { success: false, message: `Hook ${hookName} not found` };
    }

    try {
      const result = await Promise.race([
        hook.execute(input),
        this.createTimeoutPromise(),
      ]);
      return result;
    } catch (error) {
      return {
        success: false,
        message: `Hook ${hookName} failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
        exitCode: 1,
      };
    }
  }

  /**
   * Execute all registered hooks
   */
  async executeAllHooks(input: HookInput): Promise<HookResult[]> {
    const results: HookResult[] = [];

    for (const [name, hook] of this.hooks) {
      try {
        const result = await this.executeHook(name, input);
        results.push(result);

        // If any hook blocks, stop execution
        if (result.blocked) {
          break;
        }
      } catch (error) {
        results.push({
          success: false,
          message: `Hook ${name} failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
          exitCode: 1,
        });
      }
    }

    return results;
  }

  /**
   * Load configuration from file
   */
  static async loadConfig(projectRoot: string): Promise<HookConfig> {
    const configPath = path.join(projectRoot, '.moai', 'config', 'hooks.json');

    try {
      if (fs.existsSync(configPath)) {
        const configData = fs.readFileSync(configPath, 'utf-8');
        const config = JSON.parse(configData);
        return config as HookConfig;
      }
    } catch (error) {
      // Use default config if loading fails
    }

    return {
      enabled: true,
      timeout: 10000,
      disabledHooks: [],
      security: {
        allowedCommands: [
          'git',
          'npm',
          'node',
          'python',
          'pytest',
          'go',
          'cargo',
        ],
        blockedPatterns: ['rm -rf', 'sudo', 'chmod 777'],
        requireApproval: ['--force', '--hard'],
      },
    };
  }

  /**
   * Get list of registered hooks
   */
  getRegisteredHooks(): string[] {
    return Array.from(this.hooks.keys());
  }

  /**
   * Check if hook is registered
   */
  hasHook(hookName: string): boolean {
    return this.hooks.has(hookName);
  }

  private createTimeoutPromise(): Promise<HookResult> {
    return new Promise((_, reject) => {
      setTimeout(() => {
        reject(
          new Error(`Hook execution timed out after ${this.config.timeout}ms`)
        );
      }, this.config.timeout);
    });
  }
}

/**
 * Parse input from stdin for Claude Code compatibility
 */
export function parseClaudeInput(): Promise<HookInput> {
  return new Promise((resolve, reject) => {
    let data = '';

    process.stdin.on('data', chunk => {
      data += chunk.toString();
    });

    process.stdin.on('end', () => {
      try {
        if (data.trim()) {
          const parsed = JSON.parse(data) as HookInput;
          resolve(parsed);
        } else {
          resolve({});
        }
      } catch (error) {
        reject(
          new Error(
            `Invalid JSON input: ${error instanceof Error ? error.message : 'Unknown error'}`
          )
        );
      }
    });

    process.stdin.on('error', error => {
      reject(error);
    });
  });
}

/**
 * Output result in Claude Code compatible format
 */
export function outputResult(result: HookResult): void {
  if (result.blocked) {
    console.error(`BLOCKED: ${result.message}`);
    process.exit(2);
  } else if (!result.success) {
    console.error(`ERROR: ${result.message}`);
    process.exit(result.exitCode || 1);
  } else if (result.message) {
    console.log(result.message);
  }

  process.exit(0);
}

// Export all hook types
export * from './types';

// Export hook implementations
export * from './security/steering-guard';
export * from './security/policy-block';
export * from './security/pre-write-guard';
export * from './workflow/file-monitor';
export * from './workflow/language-detector';
export * from './workflow/test-runner';
export * from './session/session-notice';

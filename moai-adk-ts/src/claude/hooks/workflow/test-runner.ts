/**
 * @file test-runner.ts
 * @description Test execution and reporting hook for MoAI-ADK
 * @version 1.0.0
 * @tag @TASK:TEST-REPORT-013
 */

import type { HookInput, HookResult, MoAIHook, TestResult } from '../types';
import { spawn } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';

/**
 * Command definition for test execution
 */
interface Command {
  name: string;
  args: string[];
}

/**
 * Test execution timeout in milliseconds
 */
const TIMEOUT_SECONDS = 300000; // 5 minutes

/**
 * Test Runner Hook - TypeScript port of run_tests_and_report.py
 */
export class TestRunner implements MoAIHook {
  name = 'test-runner';

  private projectRoot: string;

  constructor(projectRoot?: string) {
    this.projectRoot = projectRoot || process.cwd();
  }

  async execute(input: HookInput): Promise<HookResult> {
    // Note: This hook is typically disabled by default
    // Tests should be run during /moai:2-build phase only
    return {
      success: true,
      message:
        'Stop Hook: 비활성화됨 - 테스트는 /moai:2-build 단계에서만 실행됩니다.',
    };
  }

  /**
   * Run all detected test commands
   */
  async runTests(): Promise<TestResult[]> {
    const commands = this.collectCommands();
    const results: TestResult[] = [];

    for (const command of commands) {
      const result = await this.runCommand(command);
      results.push(result);
    }

    return results;
  }

  /**
   * Detect and collect available test commands
   */
  private collectCommands(): Command[] {
    const commands: Command[] = [];

    // Detect pytest
    const pytestCommand = this.detectPytest();
    if (pytestCommand) {
      commands.push(pytestCommand);
    }

    // Detect npm test
    const npmCommand = this.detectNpm();
    if (npmCommand) {
      commands.push(npmCommand);
    }

    // Detect go test
    const goCommand = this.detectGo();
    if (goCommand) {
      commands.push(goCommand);
    }

    // Detect cargo test
    const cargoCommand = this.detectCargo();
    if (cargoCommand) {
      commands.push(cargoCommand);
    }

    return commands;
  }

  /**
   * Detect pytest command
   */
  private detectPytest(): Command | null {
    const testsDir = path.join(this.projectRoot, 'tests');
    if (fs.existsSync(testsDir) && this.commandExists('pytest')) {
      return {
        name: 'pytest',
        args: ['python', '-m', 'pytest', '-q'],
      };
    }
    return null;
  }

  /**
   * Detect npm test command
   */
  private detectNpm(): Command | null {
    const packageJson = path.join(this.projectRoot, 'package.json');
    if (fs.existsSync(packageJson) && this.commandExists('npm')) {
      return {
        name: 'npm test',
        args: ['npm', 'test', '--', '--watch=false'],
      };
    }
    return null;
  }

  /**
   * Detect go test command
   */
  private detectGo(): Command | null {
    const goMod = path.join(this.projectRoot, 'go.mod');
    if (fs.existsSync(goMod) && this.commandExists('go')) {
      return {
        name: 'go test',
        args: ['go', 'test', './...'],
      };
    }
    return null;
  }

  /**
   * Detect cargo test command
   */
  private detectCargo(): Command | null {
    const cargoToml = path.join(this.projectRoot, 'Cargo.toml');
    if (fs.existsSync(cargoToml) && this.commandExists('cargo')) {
      return {
        name: 'cargo test',
        args: ['cargo', 'test', '--quiet'],
      };
    }
    return null;
  }

  /**
   * Check if command exists in PATH
   */
  private commandExists(command: string): boolean {
    try {
      const { execSync } = require('child_process');
      execSync(`which ${command}`, { stdio: 'ignore' });
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Run a single test command
   */
  private async runCommand(command: Command): Promise<TestResult> {
    const startTime = Date.now();

    return new Promise(resolve => {
      const proc = spawn(command.args[0], command.args.slice(1), {
        cwd: this.projectRoot,
        stdio: 'pipe',
      });

      let stdout = '';
      let stderr = '';

      proc.stdout?.on('data', data => {
        stdout += data.toString();
      });

      proc.stderr?.on('data', data => {
        stderr += data.toString();
      });

      const timeout = setTimeout(() => {
        proc.kill();
        resolve({
          runner: command.name,
          exitCode: 124,
          stdout: stdout.trim(),
          stderr: `${command.name} timed out after ${TIMEOUT_SECONDS / 1000}s`,
          duration: Date.now() - startTime,
        });
      }, TIMEOUT_SECONDS);

      proc.on('close', code => {
        clearTimeout(timeout);
        resolve({
          runner: command.name,
          exitCode: code || 0,
          stdout: stdout.trim(),
          stderr: stderr.trim(),
          duration: Date.now() - startTime,
        });
      });

      proc.on('error', (error: any) => {
        clearTimeout(timeout);
        resolve({
          runner: command.name,
          exitCode: 1,
          stdout: stdout.trim(),
          stderr: `${command.name} failed: ${error.message}`,
          duration: Date.now() - startTime,
        });
      });
    });
  }

  /**
   * Generate test report from results
   */
  generateReport(results: TestResult[]): string {
    const lines: string[] = [];

    lines.push('🧪 Test Results:');
    lines.push('');

    for (const result of results) {
      const status = result.exitCode === 0 ? '✅' : '❌';
      const duration = (result.duration / 1000).toFixed(2);

      lines.push(`${status} ${result.runner} (${duration}s)`);

      if (result.exitCode !== 0) {
        lines.push(`   Error: ${result.stderr}`);
      }

      if (result.stdout) {
        lines.push(
          `   Output: ${result.stdout.substring(0, 200)}${result.stdout.length > 200 ? '...' : ''}`
        );
      }

      lines.push('');
    }

    const passed = results.filter(r => r.exitCode === 0).length;
    const total = results.length;

    lines.push(`📊 Summary: ${passed}/${total} test suites passed`);

    return lines.join('\n');
  }

  /**
   * Get available test commands
   */
  getAvailableCommands(): Command[] {
    return this.collectCommands();
  }
}

/**
 * CLI entry point for Claude Code compatibility
 */
export async function main(): Promise<void> {
  const testRunner = new TestRunner();
  const result = await testRunner.execute({});

  if (result.message) {
    console.log(result.message);
  }

  process.exit(0);
}

// Execute if run directly
if (require.main === module) {
  main().catch(() => {
    process.exit(0);
  });
}

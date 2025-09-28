/**
 * @file steering-guard.ts
 * @description Steering guard hook for enforcing safety rules with session notifications
 * @version 1.0.0
 * @tag @SEC:STEERING-GUARD-013
 */

import type { HookInput, HookResult, MoAIHook, SecurityPattern } from '../types';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';

/**
 * Banned patterns that should be blocked
 */
const BANNED_PATTERNS: SecurityPattern[] = [
  {
    pattern: /ignore (the )?(claude|constitution|steering|instructions)/i,
    message: '헌법/지침 무시는 허용되지 않습니다.',
    severity: 'critical',
  },
  {
    pattern: /disable (all )?(hooks?|guards?|polic(y|ies))/i,
    message: 'Hook/Guard 해제 요청은 차단되었습니다.',
    severity: 'critical',
  },
  {
    pattern: /rm -rf/i,
    message: '위험한 셸 명령을 프롬프트로 제출할 수 없습니다.',
    severity: 'high',
  },
  {
    pattern: /drop (all )?safeguards/i,
    message: '안전장치 제거 요청은 거부됩니다.',
    severity: 'critical',
  },
  {
    pattern: /clear (all )?(memory|steering)/i,
    message: 'Steering 메모리를 강제 삭제하는 요청은 지원하지 않습니다.',
    severity: 'high',
  },
];

/**
 * Session notification file path
 */
const SESSION_NOTIFIED_FILE = path.join(os.tmpdir(), 'moai_session_notified');

/**
 * Hybrid system status
 */
interface HybridStatus {
  status: 'full_hybrid' | 'python_only' | 'typescript_only' | 'legacy';
  description: string;
}

/**
 * Steering Guard Hook - TypeScript port of steering_guard.py
 */
export class SteeringGuard implements MoAIHook {
  name = 'steering-guard';

  async execute(input: HookInput): Promise<HookResult> {
    // Show session notice on first execution
    this.showSessionNotice();

    const prompt = input.prompt;
    if (!prompt || typeof prompt !== 'string') {
      return { success: true };
    }

    // Check against banned patterns
    for (const pattern of BANNED_PATTERNS) {
      if (pattern.pattern.test(prompt)) {
        return {
          success: false,
          blocked: true,
          message: pattern.message,
          exitCode: 2,
        };
      }
    }

    // Provide lightweight steering context
    return {
      success: true,
      message:
        'Steering Guard: 개발 가이드과 TAG 규칙을 준수하며 작업을 진행합니다.',
    };
  }

  /**
   * Check if this is a MoAI project
   */
  private checkMoAIProject(): boolean {
    const currentDir = process.cwd();
    const moaiPath = path.join(currentDir, '.moai');
    const claudePath = path.join(currentDir, 'CLAUDE.md');

    return fs.existsSync(moaiPath) && fs.existsSync(claudePath);
  }

  /**
   * Check hybrid system status
   */
  private checkHybridSystemStatus(): HybridStatus {
    const currentDir = process.cwd();

    // Check TypeScript project
    const tsProject = path.join(currentDir, 'moai-adk-ts');
    const hasTypeScript =
      fs.existsSync(tsProject) &&
      fs.existsSync(path.join(tsProject, 'package.json'));

    // Check Python bridge
    const pythonBridge = path.join(
      currentDir,
      'src',
      'moai_adk',
      'core',
      'bridge'
    );
    const hasPythonBridge =
      fs.existsSync(pythonBridge) &&
      fs.existsSync(path.join(pythonBridge, 'typescript_bridge.py'));

    if (hasTypeScript && hasPythonBridge) {
      return {
        status: 'full_hybrid',
        description: 'Python + TypeScript 완전 통합 🔗',
      };
    } else if (hasPythonBridge) {
      return {
        status: 'python_only',
        description: 'Python 브릿지 (TypeScript 없음) 🐍',
      };
    } else if (hasTypeScript) {
      return {
        status: 'typescript_only',
        description: 'TypeScript (브릿지 없음) ⚠️',
      };
    } else {
      return {
        status: 'legacy',
        description: '기존 Python 시스템 📦',
      };
    }
  }

  /**
   * Show session notice (first time only)
   */
  private showSessionNotice(): void {
    if (fs.existsSync(SESSION_NOTIFIED_FILE)) {
      return; // Already notified
    }

    if (!this.checkMoAIProject()) {
      return; // Not a MoAI project
    }

    // Check hybrid system status
    const hybridStatus = this.checkHybridSystemStatus();

    // Show notification
    console.error('🚀 MoAI-ADK 하이브리드 프로젝트가 감지되었습니다!');
    console.error(
      '📖 개발 가이드: CLAUDE.md | TRUST 원칙: .moai/memory/development-guide.md'
    );
    console.error(
      '⚡ 하이브리드 워크플로우: /moai:1-spec → /moai:2-build → /moai:3-sync'
    );
    console.error(`🔗 시스템 상태: ${hybridStatus.description}`);
    console.error('🔧 디버깅: /moai:4-debug | 설정 관리: @agent-cc-manager');
    console.error('');

    // Mark as notified
    try {
      fs.writeFileSync(SESSION_NOTIFIED_FILE, 'notified');
    } catch {
      // Ignore write errors
    }
  }
}

/**
 * CLI entry point for Claude Code compatibility
 */
export async function main(): Promise<void> {
  try {
    const { parseClaudeInput, outputResult } = await import('../index');
    const input = await parseClaudeInput();
    const guard = new SteeringGuard();
    const result = await guard.execute(input);
    outputResult(result);
  } catch (error) {
    console.error(
      `ERROR steering_guard: ${error instanceof Error ? error.message : 'Unknown error'}`
    );
    process.exit(1);
  }
}

// Execute if run directly
if (require.main === module) {
  main().catch(error => {
    console.error(
      `ERROR steering_guard: ${error instanceof Error ? error.message : 'Unknown error'}`
    );
    process.exit(1);
  });
}

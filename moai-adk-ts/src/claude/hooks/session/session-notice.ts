/**
 * @file session-notice.ts
 * @description Session start notification hook for MoAI-ADK
 * @version 1.0.0
 * @tag @FEATURE:SESSION-NOTICE-013
 */

import type {
  HookInput,
  HookResult,
  MoAIHook,
  ProjectStatus,
  ConstitutionStatus,
  SpecProgress,
} from '../types';
import { spawn } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';

/**
 * Git information interface
 */
interface GitInfo {
  branch: string;
  commit: string;
  message: string;
  changesCount: number;
}

/**
 * Session Notifier Hook - TypeScript port of session_start_notice.py
 */
export class SessionNotifier implements MoAIHook {
  name = 'session-notice';

  private projectRoot: string;
  private moaiConfigPath: string;

  constructor(projectRoot?: string) {
    this.projectRoot = projectRoot || process.cwd();
    this.moaiConfigPath = path.join(this.projectRoot, '.moai', 'config.json');
  }

  async execute(input: HookInput): Promise<HookResult> {
    try {
      if (this.isMoAIProject()) {
        const status = await this.getProjectStatus();
        const output = await this.generateSessionOutput(status);

        return {
          success: true,
          message: output,
          data: status,
        };
      } else {
        return {
          success: true,
          message: '💡 Run `/moai:0-project` to initialize MoAI-ADK',
        };
      }
    } catch (error) {
      // Silent failure to avoid breaking Claude Code session
      return { success: true };
    }
  }

  /**
   * Get comprehensive project status
   */
  async getProjectStatus(): Promise<ProjectStatus> {
    return {
      projectName: path.basename(this.projectRoot),
      moaiVersion: this.getMoAIVersion(),
      initialized: this.isMoAIProject(),
      constitutionStatus: this.checkConstitutionStatus(),
      pipelineStage: this.getCurrentPipelineStage(),
      specProgress: this.getSpecProgress(),
    };
  }

  /**
   * Check if this is a MoAI project
   */
  private isMoAIProject(): boolean {
    const requiredPaths = [
      path.join(this.projectRoot, '.moai'),
      path.join(this.projectRoot, '.claude', 'commands', 'moai'),
    ];

    return requiredPaths.every(p => fs.existsSync(p));
  }

  /**
   * Check development guide compliance status
   */
  private checkConstitutionStatus(): ConstitutionStatus {
    if (!this.isMoAIProject()) {
      return {
        status: 'not_initialized',
        violations: [],
      };
    }

    const criticalFiles = ['.moai/memory/development-guide.md', 'CLAUDE.md'];

    const violations: string[] = [];

    for (const filePath of criticalFiles) {
      if (!fs.existsSync(path.join(this.projectRoot, filePath))) {
        violations.push(`Missing critical file: ${filePath}`);
      }
    }

    return {
      status: violations.length === 0 ? 'ok' : 'violations_found',
      violations,
    };
  }

  /**
   * Get MoAI-ADK version
   */
  private getMoAIVersion(): string {
    try {
      if (fs.existsSync(this.moaiConfigPath)) {
        const configData = fs.readFileSync(this.moaiConfigPath, 'utf-8');
        const config = JSON.parse(configData);
        return config.project?.version || 'unknown';
      }
    } catch (error) {
      // Fall back to unknown
    }

    return 'unknown';
  }

  /**
   * Get current pipeline stage
   */
  private getCurrentPipelineStage(): string {
    try {
      if (fs.existsSync(this.moaiConfigPath)) {
        const configData = fs.readFileSync(this.moaiConfigPath, 'utf-8');
        const config = JSON.parse(configData);
        return config.pipeline?.current_stage || 'unknown';
      }
    } catch (error) {
      // Fall back to heuristic
    }

    // Fallback heuristic
    const specsDir = path.join(this.projectRoot, '.moai', 'specs');
    if (fs.existsSync(specsDir)) {
      const hasSpecs = fs
        .readdirSync(specsDir)
        .some(dir => fs.existsSync(path.join(specsDir, dir, 'spec.md')));

      if (hasSpecs) {
        return 'implementation';
      }
    }

    if (this.isMoAIProject()) {
      return 'specification';
    }

    return 'initialization';
  }

  /**
   * Get SPEC progress information
   */
  private getSpecProgress(): SpecProgress {
    const specsDir = path.join(this.projectRoot, '.moai', 'specs');

    if (!fs.existsSync(specsDir)) {
      return { total: 0, completed: 0 };
    }

    try {
      const specDirs = fs
        .readdirSync(specsDir)
        .filter(name => fs.statSync(path.join(specsDir, name)).isDirectory())
        .filter(name => name.startsWith('SPEC-'));

      const totalSpecs = specDirs.length;
      let completed = 0;

      // Simple heuristic: completed if has both spec.md and plan.md
      for (const specDir of specDirs) {
        const specPath = path.join(specsDir, specDir, 'spec.md');
        const planPath = path.join(specsDir, specDir, 'plan.md');

        if (fs.existsSync(specPath) && fs.existsSync(planPath)) {
          completed++;
        }
      }

      return { total: totalSpecs, completed };
    } catch (error) {
      return { total: 0, completed: 0 };
    }
  }

  /**
   * Get Git information
   */
  private async getGitInfo(): Promise<GitInfo> {
    const defaultInfo: GitInfo = {
      branch: 'unknown',
      commit: 'unknown',
      message: 'No commit message',
      changesCount: 0,
    };

    try {
      const [branch, commit, message, changesCount] = await Promise.all([
        this.runGitCommand(['rev-parse', '--abbrev-ref', 'HEAD']),
        this.runGitCommand(['rev-parse', 'HEAD']),
        this.runGitCommand(['log', '-1', '--pretty=%s']),
        this.getGitChangesCount(),
      ]);

      return {
        branch: branch || defaultInfo.branch,
        commit: commit || defaultInfo.commit,
        message: message || defaultInfo.message,
        changesCount,
      };
    } catch (error) {
      return defaultInfo;
    }
  }

  /**
   * Get count of Git changes
   */
  private async getGitChangesCount(): Promise<number> {
    try {
      const output = await this.runGitCommand(['status', '--porcelain']);
      if (output) {
        const lines = output
          .trim()
          .split('\n')
          .filter(line => line.trim().length > 0);
        return lines.length;
      }
      return 0;
    } catch (error) {
      return 0;
    }
  }

  /**
   * Run a Git command and return output
   */
  private async runGitCommand(args: string[]): Promise<string | null> {
    return new Promise(resolve => {
      const proc = spawn('git', args, {
        cwd: this.projectRoot,
        stdio: 'pipe',
      });

      let stdout = '';

      proc.stdout?.on('data', data => {
        stdout += data.toString();
      });

      const timeout = setTimeout(() => {
        proc.kill();
        resolve(null);
      }, 2000);

      proc.on('close', code => {
        clearTimeout(timeout);
        if (code === 0) {
          resolve(stdout.trim());
        } else {
          resolve(null);
        }
      });

      proc.on('error', () => {
        clearTimeout(timeout);
        resolve(null);
      });
    });
  }

  /**
   * Generate session output message
   */
  private async generateSessionOutput(status: ProjectStatus): Promise<string> {
    const lines: string[] = [];

    // Show violations if any
    if (status.constitutionStatus.violations.length > 0) {
      lines.push('⚠️  Development guide violations detected:');
      for (const violation of status.constitutionStatus.violations) {
        lines.push(`   • ${violation}`);
      }
      lines.push('');
    }

    // Project information
    lines.push(`🗿 MoAI-ADK 프로젝트: ${status.projectName}`);

    // Git information
    const gitInfo = await this.getGitInfo();
    const shortCommit = gitInfo.commit.substring(0, 7);
    const shortMessage = gitInfo.message.substring(0, 50);
    const ellipsis = gitInfo.message.length > 50 ? '...' : '';

    lines.push(
      `🌿 현재 브랜치: ${gitInfo.branch} (${shortCommit} ${shortMessage}${ellipsis})`
    );

    if (gitInfo.changesCount > 0) {
      lines.push(`📝 변경사항: ${gitInfo.changesCount}개 파일`);
    }

    // SPEC progress
    const remaining = status.specProgress.total - status.specProgress.completed;
    lines.push(
      `📝 SPEC 진행률: ${status.specProgress.completed}/${status.specProgress.total} (미완료 ${remaining}개)`
    );

    // System status
    lines.push('✅ 통합 체크포인트 시스템 사용 가능');

    return lines.join('\n');
  }
}

/**
 * CLI entry point for Claude Code compatibility
 */
export async function main(): Promise<void> {
  try {
    const notifier = new SessionNotifier();
    const result = await notifier.execute({});

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

/**
 * @file Session start notification hook with version check
 * @author MoAI Team
 * @tags @FEATURE:SESSION-NOTICE-002 @REQ:HOOK-SESSION-START-002
 */

import { spawn } from 'node:child_process';
import * as fs from 'node:fs';
import * as path from 'node:path';

/**
 * Project status information
 */
interface ProjectStatus {
  projectName: string;
  moaiVersion: string;
  initialized: boolean;
  constitutionStatus: ConstitutionStatus;
  pipelineStage: string;
  specProgress: SpecProgress;
}

/**
 * Constitution status information
 */
interface ConstitutionStatus {
  status: 'ok' | 'violations_found' | 'not_initialized';
  violations: string[];
}

/**
 * SPEC progress information
 */
interface SpecProgress {
  total: number;
  completed: number;
}

/**
 * Git information
 */
interface GitInfo {
  branch: string;
  commit: string;
  message: string;
  changesCount: number;
}

/**
 * Hook execution input
 */
interface HookInput {
  // Can be extended with additional inputs
  [key: string]: unknown;
}

/**
 * Hook execution result
 */
interface HookResult {
  success: boolean;
  message?: string;
  data?: unknown;
}

/**
 * Version check result
 */
interface VersionCheckResult {
  current: string;
  latest: string | null;
  hasUpdate: boolean;
}

/**
 * Session notification hook
 */
export class SessionNotifier {
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
          message: '💡 Run `/moai:8-project` to initialize MoAI-ADK',
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
  isMoAIProject(): boolean {
    const requiredPaths = [
      path.join(this.projectRoot, '.moai'),
      path.join(this.projectRoot, '.claude', 'commands', 'moai'),
    ];

    return requiredPaths.every((p) => fs.existsSync(p));
  }

  /**
   * Check development guide compliance status
   */
  checkConstitutionStatus(): ConstitutionStatus {
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
   * Get MoAI-ADK version from package.json
   * Falls back to config.json if package.json is unavailable
   */
  getMoAIVersion(): string {
    try {
      // First, try to get version from moai-adk package.json
      const packageJsonPath = path.join(
        this.projectRoot,
        'node_modules',
        'moai-adk',
        'package.json'
      );

      if (fs.existsSync(packageJsonPath)) {
        const packageData = fs.readFileSync(packageJsonPath, 'utf-8');
        const packageJson = JSON.parse(packageData) as { version?: string };
        if (packageJson.version) {
          return packageJson.version;
        }
      }

      // Fallback: check .moai/config.json
      if (fs.existsSync(this.moaiConfigPath)) {
        const configData = fs.readFileSync(this.moaiConfigPath, 'utf-8');
        const config = JSON.parse(configData) as {
          project?: { version?: string };
        };
        const version = config.project?.version;

        // Detect and handle template variables
        if (version && !version.includes('{{') && !version.includes('}}')) {
          return version;
        }
      }
    } catch (error) {
      // Ignore errors
    }
    return 'unknown';
  }

  /**
   * Get current pipeline stage
   */
  getCurrentPipelineStage(): string {
    try {
      if (fs.existsSync(this.moaiConfigPath)) {
        const configData = fs.readFileSync(this.moaiConfigPath, 'utf-8');
        const config = JSON.parse(configData) as {
          pipeline?: { current_stage?: string };
        };
        return config.pipeline?.current_stage || 'unknown';
      }
    } catch (error) {
      // Ignore errors
    }

    // Fallback heuristic
    const specsDir = path.join(this.projectRoot, '.moai', 'specs');
    if (fs.existsSync(specsDir)) {
      const hasSpecs = fs
        .readdirSync(specsDir)
        .some((dir) => fs.existsSync(path.join(specsDir, dir, 'spec.md')));
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
  getSpecProgress(): SpecProgress {
    const specsDir = path.join(this.projectRoot, '.moai', 'specs');

    if (!fs.existsSync(specsDir)) {
      return { total: 0, completed: 0 };
    }

    try {
      const specDirs = fs
        .readdirSync(specsDir)
        .filter((name) => fs.statSync(path.join(specsDir, name)).isDirectory())
        .filter((name) => name.startsWith('SPEC-'));

      const totalSpecs = specDirs.length;
      let completed = 0;

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
  async getGitInfo(): Promise<GitInfo> {
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
  async getGitChangesCount(): Promise<number> {
    try {
      const output = await this.runGitCommand(['status', '--porcelain']);
      if (output) {
        const lines = output
          .trim()
          .split('\n')
          .filter((line) => line.trim().length > 0);
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
  async runGitCommand(args: string[]): Promise<string | null> {
    return new Promise((resolve) => {
      const proc = spawn('git', args, {
        cwd: this.projectRoot,
        stdio: 'pipe',
      });

      let stdout = '';

      proc.stdout?.on('data', (data) => {
        stdout += data.toString();
      });

      const timeout = setTimeout(() => {
        proc.kill();
        resolve(null);
      }, 2000);

      proc.on('close', (code) => {
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
   * Check for latest version from npm registry
   */
  async checkLatestVersion(): Promise<VersionCheckResult | null> {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 2000);

      const response = await fetch(
        'https://registry.npmjs.org/moai-adk/latest',
        {
          signal: controller.signal,
          headers: {
            Accept: 'application/json',
          },
        }
      );

      clearTimeout(timeoutId);

      if (!response.ok) {
        return null;
      }

      const data = (await response.json()) as { version: string };
      const latest = data.version;
      const current = this.getMoAIVersion();

      // Simple version comparison
      const hasUpdate = this.compareVersions(current, latest) < 0;

      return {
        current,
        latest,
        hasUpdate,
      };
    } catch (error) {
      // Silently fail - don't block session start
      return null;
    }
  }

  /**
   * Compare two semantic version strings
   */
  private compareVersions(v1: string, v2: string): number {
    const parts1 = v1.split('.').map(Number);
    const parts2 = v2.split('.').map(Number);

    for (let i = 0; i < Math.max(parts1.length, parts2.length); i++) {
      const num1 = parts1[i] || 0;
      const num2 = parts2[i] || 0;

      if (num1 < num2) return -1;
      if (num1 > num2) return 1;
    }

    return 0;
  }

  /**
   * Generate session output message
   */
  async generateSessionOutput(status: ProjectStatus): Promise<string> {
    const lines: string[] = [];

    // Version check (non-blocking)
    const versionCheck = await this.checkLatestVersion();

    if (status.constitutionStatus.violations.length > 0) {
      lines.push('⚠️  Development guide violations detected:');
      for (const violation of status.constitutionStatus.violations) {
        lines.push(`   • ${violation}`);
      }
      lines.push('');
    }

    lines.push(`🗿 MoAI-ADK 프로젝트: ${status.projectName}`);

    // Version info
    if (versionCheck && versionCheck.latest) {
      if (versionCheck.hasUpdate) {
        lines.push(
          `📦 버전: v${versionCheck.current} → ⚡ v${versionCheck.latest} 업데이트 가능`
        );
      } else {
        lines.push(`📦 버전: v${versionCheck.current} (최신)`);
      }
    } else {
      lines.push(`📦 버전: v${status.moaiVersion}`);
    }

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

    const remaining = status.specProgress.total - status.specProgress.completed;
    lines.push(
      `📝 SPEC 진행률: ${status.specProgress.completed}/${status.specProgress.total} (미완료 ${remaining}개)`
    );

    lines.push('✅ 통합 체크포인트 시스템 사용 가능');

    return lines.join('\n');
  }
}

/**
 * Main entry point for hook execution
 */
export async function main(): Promise<void> {
  try {
    const notifier = new SessionNotifier();
    const result = await notifier.execute({});

    if (result.message) {
      console.log(result.message);
    }
  } catch (error) {
    // Silent failure
  }
}

// Execute if run directly
if (require.main === module) {
  main().catch(() => {
    // Silent failure
  });
}
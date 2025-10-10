/**
 * @CODE:SESSION-NOTICE-001:UTILS |
 * Related: @CODE:SESSION-NOTICE-001
 *
 * Session Notice Utility Functions
 */

import { spawn } from 'node:child_process';
import * as fs from 'node:fs';
import * as path from 'node:path';
import type {
  ConstitutionStatus,
  GitInfo,
  SpecProgress,
  VersionCheckResult,
} from './types';

/**
 * @CODE:INIT-002 | SPEC: .moai/specs/SPEC-INIT-002/spec.md
 * Check if this is a MoAI project
 *
 * Changed from array-based check to explicit variable check for clarity.
 * Updated path from '.claude/commands/moai' to '.claude/commands/alfred'
 * to reflect the new branding (moai → alfred).
 */
export function isMoAIProject(projectRoot: string): boolean {
  const moaiDir = path.join(projectRoot, '.moai');
  const alfredCommands = path.join(projectRoot, '.claude', 'commands', 'alfred');

  return fs.existsSync(moaiDir) && fs.existsSync(alfredCommands);
}

/**
 * Check development guide compliance status
 */
export function checkConstitutionStatus(
  projectRoot: string
): ConstitutionStatus {
  if (!isMoAIProject(projectRoot)) {
    return {
      status: 'not_initialized',
      violations: [],
    };
  }

  const criticalFiles = ['.moai/memory/development-guide.md', 'CLAUDE.md'];
  const violations: string[] = [];

  for (const filePath of criticalFiles) {
    if (!fs.existsSync(path.join(projectRoot, filePath))) {
      violations.push(`Missing critical file: ${filePath}`);
    }
  }

  return {
    status: violations.length === 0 ? 'ok' : 'violations_found',
    violations,
  };
}

/**
 * Get MoAI-ADK plugin version from Claude plugin directory
 * Priority order:
 * 1. ~/.claude/plugins/marketplaces/moai-adk/.claude-plugin/plugin.json (installed plugin)
 * 2. ~/.claude/plugins/cache/moai-adk/.claude-plugin/plugin.json (cached plugin)
 * 3. node_modules/moai-adk/package.json (npm package)
 * 4. 'unknown' (fallback)
 */
export function getMoAIVersion(projectRoot: string): string {
  try {
    // Try to get user home directory
    const homeDir = process.env.HOME || process.env.USERPROFILE;

    if (homeDir) {
      // Check marketplaces plugin (installed version)
      const marketplacePluginPath = path.join(
        homeDir,
        '.claude',
        'plugins',
        'marketplaces',
        'moai-adk',
        '.claude-plugin',
        'plugin.json'
      );

      if (fs.existsSync(marketplacePluginPath)) {
        const pluginData = fs.readFileSync(marketplacePluginPath, 'utf-8');
        const plugin = JSON.parse(pluginData) as { version?: string };
        if (plugin.version) {
          return plugin.version;
        }
      }

      // Check cache plugin (cached version)
      const cachePluginPath = path.join(
        homeDir,
        '.claude',
        'plugins',
        'cache',
        'moai-adk',
        '.claude-plugin',
        'plugin.json'
      );

      if (fs.existsSync(cachePluginPath)) {
        const pluginData = fs.readFileSync(cachePluginPath, 'utf-8');
        const plugin = JSON.parse(pluginData) as { version?: string };
        if (plugin.version) {
          return plugin.version;
        }
      }
    }

    // Fallback: Check node_modules (for development)
    const packageJsonPath = path.join(
      projectRoot,
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
  } catch (_error) {
    // Ignore errors
  }
  return 'unknown';
}

/**
 * Get current pipeline stage
 */
export function getCurrentPipelineStage(projectRoot: string): string {
  try {
    const moaiConfigPath = path.join(projectRoot, '.moai', 'config.json');
    if (fs.existsSync(moaiConfigPath)) {
      const configData = fs.readFileSync(moaiConfigPath, 'utf-8');
      const config = JSON.parse(configData) as {
        pipeline?: { current_stage?: string };
      };
      return config.pipeline?.current_stage || 'unknown';
    }
  } catch (_error) {
    // Ignore errors
  }

  // Fallback heuristic
  const specsDir = path.join(projectRoot, '.moai', 'specs');
  if (fs.existsSync(specsDir)) {
    const hasSpecs = fs
      .readdirSync(specsDir)
      .some(dir => fs.existsSync(path.join(specsDir, dir, 'spec.md')));
    if (hasSpecs) {
      return 'implementation';
    }
  }

  if (isMoAIProject(projectRoot)) {
    return 'specification';
  }

  return 'initialization';
}

/**
 * Get SPEC progress information
 */
export function getSpecProgress(projectRoot: string): SpecProgress {
  const specsDir = path.join(projectRoot, '.moai', 'specs');

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

    for (const specDir of specDirs) {
      const specPath = path.join(specsDir, specDir, 'spec.md');
      const planPath = path.join(specsDir, specDir, 'plan.md');

      if (fs.existsSync(specPath) && fs.existsSync(planPath)) {
        completed++;
      }
    }

    return { total: totalSpecs, completed };
  } catch (_error) {
    return { total: 0, completed: 0 };
  }
}

/**
 * Run a Git command and return output
 */
export async function runGitCommand(
  projectRoot: string,
  args: string[]
): Promise<string | null> {
  return new Promise(resolve => {
    const proc = spawn('git', args, {
      cwd: projectRoot,
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
 * Get count of Git changes
 */
export async function getGitChangesCount(projectRoot: string): Promise<number> {
  try {
    const output = await runGitCommand(projectRoot, ['status', '--porcelain']);
    if (output) {
      const lines = output
        .trim()
        .split('\n')
        .filter(line => line.trim().length > 0);
      return lines.length;
    }
    return 0;
  } catch (_error) {
    return 0;
  }
}

/**
 * Get Git information
 */
export async function getGitInfo(projectRoot: string): Promise<GitInfo> {
  const defaultInfo: GitInfo = {
    branch: 'unknown',
    commit: 'unknown',
    message: 'No commit message',
    changesCount: 0,
  };

  try {
    const [branch, commit, message, changesCount] = await Promise.all([
      runGitCommand(projectRoot, ['rev-parse', '--abbrev-ref', 'HEAD']),
      runGitCommand(projectRoot, ['rev-parse', 'HEAD']),
      runGitCommand(projectRoot, ['log', '-1', '--pretty=%s']),
      getGitChangesCount(projectRoot),
    ]);

    return {
      branch: branch || defaultInfo.branch,
      commit: commit || defaultInfo.commit,
      message: message || defaultInfo.message,
      changesCount,
    };
  } catch (_error) {
    return defaultInfo;
  }
}

/**
 * Check for latest version from npm registry
 */
export async function checkLatestVersion(
  currentVersion: string
): Promise<VersionCheckResult | null> {
  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 2000);

    const response = await fetch('https://registry.npmjs.org/moai-adk/latest', {
      signal: controller.signal,
      headers: {
        Accept: 'application/json',
      },
    });

    clearTimeout(timeoutId);

    if (!response.ok) {
      return null;
    }

    const data = (await response.json()) as { version: string };
    const latest = data.version;

    const hasUpdate = compareVersions(currentVersion, latest) < 0;

    return {
      current: currentVersion,
      latest,
      hasUpdate,
    };
  } catch (_error) {
    return null;
  }
}

/**
 * Compare two semantic version strings
 */
export function compareVersions(v1: string, v2: string): number {
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

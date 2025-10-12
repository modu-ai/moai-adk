#!/usr/bin/env node
'use strict';

/**
 * @CODE:SESSION-NOTICE-001 | @CODE:HOOKS-001 |
 * Related: @CODE:HOOK-003, @CODE:SESSION-001:UI
 * SPEC: .moai/specs/SPEC-HOOKS-001/spec.md
 *
 * Session Notice Hook
 * 프로젝트 상태, SPEC 진행률, Git 정보 등 세션 시작 알림
 *
 * Pure JavaScript implementation for cross-platform compatibility.
 */

const { spawn } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');

// ============================================================================
// Type Definitions (JSDoc)
// ============================================================================

/**
 * Hook input from Claude Code
 *
 * @typedef {Object} HookInput
 * @property {string} [tool_name] - Name of the tool being invoked
 * @property {Object<string, any>} [tool_input] - Tool input parameters
 */

/**
 * Hook execution result
 *
 * @typedef {Object} HookResult
 * @property {boolean} success - Whether the hook execution succeeded
 * @property {string} [message] - Result message
 * @property {Object} [data] - Additional data
 */

/**
 * Project status information
 *
 * @typedef {Object} ProjectStatus
 * @property {string} projectName - Project name
 * @property {string} moaiVersion - MoAI-ADK package version
 * @property {boolean} initialized - Whether MoAI is initialized
 * @property {ConstitutionStatus} constitutionStatus - Constitution status
 * @property {string} pipelineStage - Current pipeline stage
 * @property {SpecProgress} specProgress - SPEC progress
 */

/**
 * Constitution status information
 *
 * @typedef {Object} ConstitutionStatus
 * @property {'ok'|'violations_found'|'not_initialized'} status - Status
 * @property {string[]} violations - List of violations
 */

/**
 * SPEC progress information
 *
 * @typedef {Object} SpecProgress
 * @property {number} total - Total SPECs
 * @property {number} completed - Completed SPECs
 */

/**
 * Git information
 *
 * @typedef {Object} GitInfo
 * @property {string} branch - Current branch
 * @property {string} commit - Current commit hash
 * @property {string} message - Latest commit message
 * @property {number} changesCount - Number of changed files
 */

/**
 * Version check result
 *
 * @typedef {Object} VersionCheckResult
 * @property {string} current - Current version
 * @property {string|null} latest - Latest version
 * @property {boolean} hasUpdate - Whether update is available
 */

// ============================================================================
// Utility Functions (Inlined from utils.ts)
// ============================================================================

/**
 * Check if this is a MoAI project
 *
 * @param {string} projectRoot - Project root directory
 * @returns {boolean} True if MoAI project
 */
function isMoAIProject(projectRoot) {
  const moaiDir = path.join(projectRoot, '.moai');
  const alfredCommands = path.join(
    projectRoot,
    '.claude',
    'commands',
    'alfred'
  );

  return fs.existsSync(moaiDir) && fs.existsSync(alfredCommands);
}

/**
 * Check development guide compliance status
 *
 * @param {string} projectRoot - Project root directory
 * @returns {ConstitutionStatus} Constitution status
 */
function checkConstitutionStatus(projectRoot) {
  if (!isMoAIProject(projectRoot)) {
    return {
      status: 'not_initialized',
      violations: [],
    };
  }

  const criticalFiles = ['.moai/memory/development-guide.md', 'CLAUDE.md'];
  const violations = [];

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
 * Get MoAI-ADK package version
 *
 * @param {string} projectRoot - Project root directory
 * @returns {string} Version string
 */
function getMoAIVersion(projectRoot) {
  try {
    const moaiConfigPath = path.join(projectRoot, '.moai', 'config.json');
    if (fs.existsSync(moaiConfigPath)) {
      const configData = fs.readFileSync(moaiConfigPath, 'utf-8');
      const config = JSON.parse(configData);

      // Priority 1: moai.version (new schema)
      if (config.moai?.version && !config.moai.version.includes('{{')) {
        return config.moai.version;
      }

      // Priority 2: project.version (backward compatibility)
      if (config.project?.version && !config.project.version.includes('{{')) {
        return config.project.version;
      }
    }

    // Priority 3: node_modules/moai-adk/package.json
    const packageJsonPath = path.join(
      projectRoot,
      'node_modules',
      'moai-adk',
      'package.json'
    );
    if (fs.existsSync(packageJsonPath)) {
      const packageData = fs.readFileSync(packageJsonPath, 'utf-8');
      const packageJson = JSON.parse(packageData);
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
 *
 * @param {string} projectRoot - Project root directory
 * @returns {string} Pipeline stage
 */
function getCurrentPipelineStage(projectRoot) {
  try {
    const moaiConfigPath = path.join(projectRoot, '.moai', 'config.json');
    if (fs.existsSync(moaiConfigPath)) {
      const configData = fs.readFileSync(moaiConfigPath, 'utf-8');
      const config = JSON.parse(configData);
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
      .some((dir) => fs.existsSync(path.join(specsDir, dir, 'spec.md')));
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
 *
 * @param {string} projectRoot - Project root directory
 * @returns {SpecProgress} SPEC progress
 */
function getSpecProgress(projectRoot) {
  const specsDir = path.join(projectRoot, '.moai', 'specs');

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
  } catch (_error) {
    return { total: 0, completed: 0 };
  }
}

/**
 * Run a Git command and return output
 *
 * @param {string} projectRoot - Project root directory
 * @param {string[]} args - Git command arguments
 * @returns {Promise<string|null>} Command output or null
 */
async function runGitCommand(projectRoot, args) {
  return new Promise((resolve) => {
    const proc = spawn('git', args, {
      cwd: projectRoot,
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
 * Get count of Git changes
 *
 * @param {string} projectRoot - Project root directory
 * @returns {Promise<number>} Number of changed files
 */
async function getGitChangesCount(projectRoot) {
  try {
    const output = await runGitCommand(projectRoot, ['status', '--porcelain']);
    if (output) {
      const lines = output
        .trim()
        .split('\n')
        .filter((line) => line.trim().length > 0);
      return lines.length;
    }
    return 0;
  } catch (_error) {
    return 0;
  }
}

/**
 * Get Git information
 *
 * @param {string} projectRoot - Project root directory
 * @returns {Promise<GitInfo>} Git information
 */
async function getGitInfo(projectRoot) {
  const defaultInfo = {
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
 * Compare two semantic version strings
 *
 * @param {string} v1 - First version
 * @param {string} v2 - Second version
 * @returns {number} -1 if v1 < v2, 0 if equal, 1 if v1 > v2
 */
function compareVersions(v1, v2) {
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
 * Check for latest version from npm registry
 *
 * @param {string} currentVersion - Current version
 * @returns {Promise<VersionCheckResult|null>} Version check result or null
 */
async function checkLatestVersion(currentVersion) {
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

    const data = await response.json();
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

// ============================================================================
// Message Builder Functions (Inlined from message-builder.ts)
// ============================================================================

/**
 * Build violation warning messages
 *
 * @param {string[]} violations - List of violations
 * @returns {string[]} Warning messages
 */
function buildViolationMessages(violations) {
  if (violations.length === 0) return [];

  const lines = ['⚠️  Development guide violations detected:'];
  for (const violation of violations) {
    lines.push(`   • ${violation}`);
  }
  lines.push('');

  return lines;
}

/**
 * Build version information message
 *
 * @param {ProjectStatus} status - Project status
 * @param {VersionCheckResult|null} versionCheck - Version check result
 * @returns {string} Version message
 */
function buildVersionMessage(status, versionCheck) {
  if (versionCheck?.latest) {
    if (versionCheck.hasUpdate) {
      return `📦 버전: v${versionCheck.current} → ⚡ v${versionCheck.latest} 업데이트 가능`;
    } else {
      return `📦 버전: v${versionCheck.current} (최신)`;
    }
  }

  return `📦 버전: v${status.moaiVersion}`;
}

/**
 * Build Git information message
 *
 * @param {GitInfo} gitInfo - Git information
 * @returns {string[]} Git messages
 */
function buildGitMessage(gitInfo) {
  const shortCommit = gitInfo.commit.substring(0, 7);
  const shortMessage = gitInfo.message.substring(0, 50);
  const ellipsis = gitInfo.message.length > 50 ? '...' : '';

  const lines = [
    `🌿 현재 브랜치: ${gitInfo.branch} (${shortCommit} ${shortMessage}${ellipsis})`,
  ];

  if (gitInfo.changesCount > 0) {
    lines.push(`📝 변경사항: ${gitInfo.changesCount}개 파일`);
  }

  return lines;
}

/**
 * Build SPEC progress message
 *
 * @param {ProjectStatus} status - Project status
 * @returns {string} SPEC progress message
 */
function buildSpecProgressMessage(status) {
  const remaining = status.specProgress.total - status.specProgress.completed;
  return `📝 SPEC 진행률: ${status.specProgress.completed}/${status.specProgress.total} (미완료 ${remaining}개)`;
}

/**
 * Generate session output message
 *
 * @param {ProjectStatus} status - Project status
 * @param {string} projectRoot - Project root directory
 * @returns {Promise<string>} Session output message
 */
async function generateSessionOutput(status, projectRoot) {
  const lines = [];

  // Version check (non-blocking)
  const currentVersion = status.moaiVersion;
  const versionCheck = await checkLatestVersion(currentVersion);

  // Violation warnings
  lines.push(...buildViolationMessages(status.constitutionStatus.violations));

  // Project header
  lines.push(`🗿 MoAI-ADK 프로젝트: ${status.projectName}`);

  // Version info
  lines.push(buildVersionMessage(status, versionCheck));

  // Git info
  const gitInfo = await getGitInfo(projectRoot);
  lines.push(...buildGitMessage(gitInfo));

  // SPEC progress
  lines.push(buildSpecProgressMessage(status));

  // System status
  lines.push('✅ 통합 체크포인트 시스템 사용 가능');

  return lines.join('\n');
}

// ============================================================================
// CLI Utilities
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
          resolve({});
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
  if (result.message) {
    console.log(result.message);
  }
  process.exit(0);
}

// ============================================================================
// Session Notifier Hook Implementation
// ============================================================================

/**
 * Session notification hook - Pure JavaScript port
 */
class SessionNotifier {
  /**
   * @param {string} [projectRoot] - Project root directory
   */
  constructor(projectRoot) {
    this.name = 'session-notice';
    this.projectRoot = projectRoot || process.cwd();
  }

  /**
   * Execute the session notice hook
   *
   * @param {HookInput} _input - Hook input from Claude Code
   * @returns {Promise<HookResult>} Hook execution result
   */
  async execute(_input) {
    try {
      if (isMoAIProject(this.projectRoot)) {
        const status = await this.getProjectStatus();
        const output = await generateSessionOutput(status, this.projectRoot);

        return {
          success: true,
          message: output,
          data: status,
        };
      } else {
        return {
          success: true,
          message: '💡 Run `/alfred:0-project` to initialize MoAI-ADK',
        };
      }
    } catch (_error) {
      return { success: true };
    }
  }

  /**
   * Get project status
   *
   * @returns {Promise<ProjectStatus>} Project status
   */
  async getProjectStatus() {
    return {
      projectName: path.basename(this.projectRoot),
      moaiVersion: getMoAIVersion(this.projectRoot),
      initialized: isMoAIProject(this.projectRoot),
      constitutionStatus: checkConstitutionStatus(this.projectRoot),
      pipelineStage: getCurrentPipelineStage(this.projectRoot),
      specProgress: getSpecProgress(this.projectRoot),
    };
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
    const _input = await parseClaudeInput();
    const hook = new SessionNotifier();
    const result = await hook.execute(_input);
    outputResult(result);
  } catch (_error) {
    // Silent failure
    process.exit(0);
  }
}

// Execute if run directly
if (require.main === module) {
  main();
}

// Export for testing
module.exports = { SessionNotifier };

/**
 * @CODE:SESSION-NOTICE-001:MESSAGE |
 * Related: @CODE:SESSION-NOTICE-001
 *
 * Session Message Builder
 */

import type { GitInfo, ProjectStatus, VersionCheckResult } from './types';
import { checkLatestVersion, getGitInfo } from './utils';

/**
 * Build violation warning messages
 */
function buildViolationMessages(violations: string[]): string[] {
  if (violations.length === 0) return [];

  const lines: string[] = ['⚠️  Development guide violations detected:'];
  for (const violation of violations) {
    lines.push(`   • ${violation}`);
  }
  lines.push('');

  return lines;
}

/**
 * Build version information message
 */
function buildVersionMessage(
  status: ProjectStatus,
  versionCheck: VersionCheckResult | null
): string {
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
 */
function buildGitMessage(gitInfo: GitInfo): string[] {
  const shortCommit = gitInfo.commit.substring(0, 7);
  const shortMessage = gitInfo.message.substring(0, 50);
  const ellipsis = gitInfo.message.length > 50 ? '...' : '';

  const lines: string[] = [
    `🌿 현재 브랜치: ${gitInfo.branch} (${shortCommit} ${shortMessage}${ellipsis})`,
  ];

  if (gitInfo.changesCount > 0) {
    lines.push(`📝 변경사항: ${gitInfo.changesCount}개 파일`);
  }

  return lines;
}

/**
 * Build SPEC progress message
 */
function buildSpecProgressMessage(status: ProjectStatus): string {
  const remaining = status.specProgress.total - status.specProgress.completed;
  return `📝 SPEC 진행률: ${status.specProgress.completed}/${status.specProgress.total} (미완료 ${remaining}개)`;
}

/**
 * Generate session output message
 */
export async function generateSessionOutput(
  status: ProjectStatus,
  projectRoot: string
): Promise<string> {
  const lines: string[] = [];

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

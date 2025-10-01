/**
 * @CODE:SESSION-NOTICE-001 |
 * Related: @CODE:HOOK-003, @CODE:SESSION-001:UI
 *
 * Session Notice Hook
 * 프로젝트 상태, SPEC 진행률, Git 정보 등 세션 시작 알림
 */

import * as path from 'node:path';
import type { HookInput, HookResult, ProjectStatus } from './types';
import * as utils from './utils';
import { generateSessionOutput } from './message-builder';

/**
 * Session notification hook
 */
export class SessionNotifier {
  name = 'session-notice';
  private projectRoot: string;

  constructor(projectRoot?: string) {
    this.projectRoot = projectRoot || process.cwd();
  }

  async execute(input: HookInput): Promise<HookResult> {
    try {
      if (utils.isMoAIProject(this.projectRoot)) {
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
          message: '💡 Run `/moai:8-project` to initialize MoAI-ADK',
        };
      }
    } catch (error) {
      return { success: true };
    }
  }

  async getProjectStatus(): Promise<ProjectStatus> {
    return {
      projectName: path.basename(this.projectRoot),
      moaiVersion: utils.getMoAIVersion(this.projectRoot),
      initialized: utils.isMoAIProject(this.projectRoot),
      constitutionStatus: utils.checkConstitutionStatus(this.projectRoot),
      pipelineStage: utils.getCurrentPipelineStage(this.projectRoot),
      specProgress: utils.getSpecProgress(this.projectRoot),
    };
  }

  // Backward compatibility for tests
  isMoAIProject = () => utils.isMoAIProject(this.projectRoot);
  checkConstitutionStatus = () => utils.checkConstitutionStatus(this.projectRoot);
  getMoAIVersion = () => utils.getMoAIVersion(this.projectRoot);
  getCurrentPipelineStage = () => utils.getCurrentPipelineStage(this.projectRoot);
  getSpecProgress = () => utils.getSpecProgress(this.projectRoot);
  getGitInfo = () => utils.getGitInfo(this.projectRoot);
  getGitChangesCount = () => utils.getGitChangesCount(this.projectRoot);
  checkLatestVersion = () => utils.checkLatestVersion(this.getMoAIVersion());
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

// Re-export types for backward compatibility
export type * from './types';

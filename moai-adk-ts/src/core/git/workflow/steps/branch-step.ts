// @CODE:GIT-004 | SPEC: SPEC-GIT-001.md | TEST: src/core/git/workflow/__tests__/workflow-automation.test.ts
// Related: @CODE:GIT-004:API

/**
 * @file Branch Step
 * @author MoAI Team
 *
 * @fileoverview 브랜치 생성 및 전환 단계
 */

import { GitNamingRules } from '../../constants/index';
import type { GitManager } from '../../git-manager';

/**
 * 브랜치 생성 단계
 */
export class BranchStep {
  constructor(private readonly gitManager: GitManager) {}

  /**
   * SPEC 브랜치 생성
   */
  async createSpecBranch(specId: string): Promise<string> {
    const branchName = GitNamingRules.createSpecBranch(specId);
    await this.gitManager.createBranch(branchName, 'main');
    return branchName;
  }

  /**
   * 릴리스 브랜치 생성
   */
  async createReleaseBranch(version: string): Promise<string> {
    const branchName = `release/${version}`;
    await this.gitManager.createBranch(branchName, 'develop');
    return branchName;
  }
}

// @CODE:GIT-004 | SPEC: SPEC-GIT-001.md | TEST: src/core/git/workflow/__tests__/workflow-automation.test.ts
// Related: @CODE:GIT-004:API

/**
 * @file Commit Step
 * @author MoAI Team
 *
 * @fileoverview 커밋 생성 단계
 */

import { GitCommitTemplates } from '../../constants/index';
import type { GitManager } from '../../git-manager';

/**
 * 커밋 생성 단계
 */
export class CommitStep {
  constructor(private readonly gitManager: GitManager) {}

  /**
   * SPEC 초기화 커밋
   */
  async createSpecInitCommit(specId: string): Promise<string> {
    const message = `${GitCommitTemplates.DOCS}: Initialize ${specId} specification`;
    const result = await this.gitManager.commitChanges(message);
    return result.hash;
  }

  /**
   * 빌드 완료 커밋
   */
  async createBuildCompleteCommit(specId: string): Promise<string> {
    const message = `${GitCommitTemplates.FEATURE}: Complete ${specId} implementation`;
    const result = await this.gitManager.commitChanges(message);
    return result.hash;
  }

  /**
   * 문서 동기화 커밋
   */
  async createSyncCommit(specId: string): Promise<string> {
    const message = `${GitCommitTemplates.DOCS}: Sync ${specId} documentation`;
    const result = await this.gitManager.commitChanges(message);
    return result.hash;
  }

  /**
   * 버전 업데이트 커밋
   */
  async createVersionCommit(version: string): Promise<string> {
    const message = `${GitCommitTemplates.CHORE}: Bump version to ${version}`;
    const result = await this.gitManager.commitChanges(message);
    return result.hash;
  }

  /**
   * TDD 체크포인트 생성
   */
  async createTDDCheckpoints(specId: string): Promise<void> {
    await this.gitManager.createCheckpoint(
      `${specId} TDD RED phase - Tests written`
    );
    await this.gitManager.createCheckpoint(
      `${specId} TDD GREEN phase - Tests passing`
    );
    await this.gitManager.createCheckpoint(
      `${specId} TDD REFACTOR phase - Code optimized`
    );
  }
}

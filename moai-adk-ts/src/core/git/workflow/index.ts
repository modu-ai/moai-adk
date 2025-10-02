// @CODE:GIT-004 | SPEC: SPEC-GIT-001.md | TEST: src/core/git/workflow/__tests__/workflow-automation.test.ts
// Related: @CODE:GIT-004:API, @CODE:GIT-004:DATA

/**
 * @file Git workflow automation orchestrator
 * @author MoAI Team
 *
 * @fileoverview Automated Git workflows for SPEC development
 */

import type { GitConfig } from '../../../types/git';
import type { GitManager } from '../git-manager';
import { BranchStep } from './steps/branch-step';
import { CommitStep } from './steps/commit-step';
import { PullRequestStep } from './steps/pr-step';
import { SpecStructureStep } from './steps/spec-structure-step';
import { SpecWorkflowStage, type WorkflowResult } from './types';

export { SpecWorkflowStage, type WorkflowResult } from './types';

/**
 * MoAI-ADK 자동화 워크플로우 관리
 */
export class WorkflowAutomation {
  private readonly branchStep: BranchStep;
  private readonly commitStep: CommitStep;
  private readonly prStep: PullRequestStep;
  private readonly specStructureStep: SpecStructureStep;

  constructor(
    readonly gitManager: GitManager,
    private readonly config: GitConfig
  ) {
    this.branchStep = new BranchStep(gitManager);
    this.commitStep = new CommitStep(gitManager);
    this.prStep = new PullRequestStep(gitManager);
    this.specStructureStep = new SpecStructureStep();
  }

  /**
   * SPEC 개발 워크플로우 시작
   * /alfred:1-spec 명령어 시뮬레이션
   */
  async startSpecWorkflow(
    specId: string,
    description: string
  ): Promise<WorkflowResult> {
    try {
      // 1. SPEC 브랜치 생성
      const branchName = await this.branchStep.createSpecBranch(specId);

      // 2. SPEC 디렉토리 구조 생성
      await this.specStructureStep.createSpecStructure(specId, description);

      // 3. 초기 커밋
      const commitHash = await this.commitStep.createSpecInitCommit(specId);

      // 4. Team 모드인 경우 Draft PR 생성
      let pullRequestUrl: string | undefined;
      if (this.config.mode === 'team') {
        pullRequestUrl = await this.prStep.createDraftPR(
          specId,
          branchName,
          description
        );
      }

      return {
        success: true,
        stage: SpecWorkflowStage.SPEC,
        branchName,
        commitHash,
        pullRequestUrl: pullRequestUrl || undefined,
        message: `SPEC ${specId} workflow started successfully`,
      };
    } catch (error) {
      return {
        success: false,
        stage: SpecWorkflowStage.INIT,
        pullRequestUrl: undefined,
        message: `Failed to start SPEC workflow: ${(error as Error).message}`,
      };
    }
  }

  /**
   * TDD 빌드 워크플로우 실행
   * /alfred:2-build 명령어 시뮬레이션
   */
  async runBuildWorkflow(specId: string): Promise<WorkflowResult> {
    try {
      // 1-3. TDD RED, GREEN, REFACTOR 체크포인트
      await this.commitStep.createTDDCheckpoints(specId);

      // 4. 빌드 완료 커밋
      const commitHash =
        await this.commitStep.createBuildCompleteCommit(specId);

      return {
        success: true,
        stage: SpecWorkflowStage.BUILD,
        commitHash,
        pullRequestUrl: undefined,
        message: `Build workflow for ${specId} completed successfully`,
      };
    } catch (error) {
      return {
        success: false,
        stage: SpecWorkflowStage.BUILD,
        pullRequestUrl: undefined,
        message: `Build workflow failed: ${(error as Error).message}`,
      };
    }
  }

  /**
   * 문서 동기화 워크플로우
   * /alfred:3-sync 명령어 시뮬레이션
   */
  async runSyncWorkflow(specId: string): Promise<WorkflowResult> {
    try {
      // 1. 문서 동기화 커밋
      const commitHash = await this.commitStep.createSyncCommit(specId);

      // 2. Team 모드인 경우 PR 상태 업데이트 (Draft → Ready)
      // 향후 구현: await this.prStep.updatePRStatus(specId, 'ready');

      return {
        success: true,
        stage: SpecWorkflowStage.SYNC,
        commitHash,
        pullRequestUrl: undefined,
        message: `Sync workflow for ${specId} completed successfully`,
      };
    } catch (error) {
      return {
        success: false,
        stage: SpecWorkflowStage.SYNC,
        pullRequestUrl: undefined,
        message: `Sync workflow failed: ${(error as Error).message}`,
      };
    }
  }

  /**
   * 전체 SPEC 워크플로우 실행
   */
  async runFullSpecWorkflow(
    specId: string,
    description: string
  ): Promise<WorkflowResult[]> {
    const results: WorkflowResult[] = [];

    // 1. SPEC 초기화
    const specResult = await this.startSpecWorkflow(specId, description);
    results.push(specResult);

    if (!specResult.success) {
      return results;
    }

    // 2. 빌드 실행
    const buildResult = await this.runBuildWorkflow(specId);
    results.push(buildResult);

    if (!buildResult.success) {
      return results;
    }

    // 3. 동기화 실행
    const syncResult = await this.runSyncWorkflow(specId);
    results.push(syncResult);

    return results;
  }

  /**
   * 브랜치 정리 워크플로우
   */
  async cleanupBranches(
    _excludeBranches: string[] = ['main', 'develop']
  ): Promise<string[]> {
    try {
      // 향후 구현: Git에서 merged 브랜치 목록 조회 및 삭제
      return [];
    } catch (error) {
      throw new Error(`Branch cleanup failed: ${(error as Error).message}`);
    }
  }

  /**
   * 릴리스 워크플로우
   */
  async createRelease(
    version: string,
    releaseNotes: string
  ): Promise<WorkflowResult> {
    try {
      // 1. 릴리스 브랜치 생성
      const branchName = await this.branchStep.createReleaseBranch(version);

      // 2. 버전 업데이트 커밋
      const commitHash = await this.commitStep.createVersionCommit(version);

      // 3. Team 모드인 경우 릴리스 PR 생성
      let pullRequestUrl: string | undefined;
      if (this.config.mode === 'team') {
        pullRequestUrl = await this.prStep.createReleasePR(
          version,
          branchName,
          releaseNotes
        );
      }

      return {
        success: true,
        stage: SpecWorkflowStage.SYNC,
        branchName,
        commitHash,
        pullRequestUrl: pullRequestUrl || undefined,
        message: `Release ${version} workflow completed successfully`,
      };
    } catch (error) {
      return {
        success: false,
        stage: SpecWorkflowStage.INIT,
        pullRequestUrl: undefined,
        message: `Release workflow failed: ${(error as Error).message}`,
      };
    }
  }
}

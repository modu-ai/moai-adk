// @CODE:GIT-004 |
// Related: @CODE:GIT-004:API

/**
 * @file Git workflow automation
 * @author MoAI Team
 *
 * @fileoverview Automated Git workflows for SPEC development
 */

import type { CreatePullRequestOptions, GitConfig } from '../../types/git';
import { GitCommitTemplates, GitNamingRules } from './constants/index';
import type { GitManager } from './git-manager';

/**
 * SPEC 개발 워크플로우 단계
 */
export enum SpecWorkflowStage {
  INIT = 'init',
  SPEC = 'spec',
  BUILD = 'build',
  SYNC = 'sync',
}

/**
 * 워크플로우 실행 결과
 */
export interface WorkflowResult {
  success: boolean;
  stage: SpecWorkflowStage;
  branchName?: string;
  commitHash?: string;
  pullRequestUrl?: string;
  message: string;
}

/**
 * MoAI-ADK 자동화 워크플로우 관리
 */
export class WorkflowAutomation {
  private gitManager: GitManager;
  private config: GitConfig;

  constructor(gitManager: GitManager, config: GitConfig) {
    this.gitManager = gitManager;
    this.config = config;
  }

  /**
   * SPEC 개발 워크플로우 시작
   * /moai:1-spec 명령어 시뮬레이션
   */
  async startSpecWorkflow(
    specId: string,
    description: string
  ): Promise<WorkflowResult> {
    try {
      // 1. SPEC 브랜치 생성
      const branchName = GitNamingRules.createSpecBranch(specId);
      await this.gitManager.createBranch(branchName, 'main');

      // 2. SPEC 디렉토리 구조 생성
      await this.createSpecStructure(specId, description);

      // 3. 초기 커밋
      const commitMessage = `${GitCommitTemplates.DOCS}: Initialize ${specId} specification`;
      const commitResult = await this.gitManager.commitChanges(commitMessage);

      // 4. Team 모드인 경우 Draft PR 생성
      let pullRequestUrl: string | undefined;
      if (this.config.mode === 'team') {
        pullRequestUrl = await this.createDraftPullRequest(
          specId,
          branchName,
          description
        );
      }

      const result: WorkflowResult = {
        success: true,
        stage: SpecWorkflowStage.SPEC,
        branchName,
        commitHash: commitResult.hash,
        message: `SPEC ${specId} workflow started successfully`,
      };
      if (pullRequestUrl) {
        result.pullRequestUrl = pullRequestUrl;
      }
      return result;
    } catch (error) {
      return {
        success: false,
        stage: SpecWorkflowStage.INIT,
        message: `Failed to start SPEC workflow: ${(error as Error).message}`,
      };
    }
  }

  /**
   * TDD 빌드 워크플로우 실행
   * /moai:2-build 명령어 시뮬레이션
   */
  async runBuildWorkflow(specId: string): Promise<WorkflowResult> {
    try {
      // 1. TDD RED 단계 체크포인트
      await this.gitManager.createCheckpoint(
        `${specId} TDD RED phase - Tests written`
      );

      // 2. TDD GREEN 단계 체크포인트
      await this.gitManager.createCheckpoint(
        `${specId} TDD GREEN phase - Tests passing`
      );

      // 3. TDD REFACTOR 단계 체크포인트
      await this.gitManager.createCheckpoint(
        `${specId} TDD REFACTOR phase - Code optimized`
      );

      // 4. 빌드 완료 커밋
      const buildCommitMessage = `${GitCommitTemplates.FEATURE}: Complete ${specId} implementation`;
      const buildResult =
        await this.gitManager.commitChanges(buildCommitMessage);

      return {
        success: true,
        stage: SpecWorkflowStage.BUILD,
        commitHash: buildResult.hash,
        message: `Build workflow for ${specId} completed successfully`,
      };
    } catch (error) {
      return {
        success: false,
        stage: SpecWorkflowStage.BUILD,
        message: `Build workflow failed: ${(error as Error).message}`,
      };
    }
  }

  /**
   * 문서 동기화 워크플로우
   * /moai:3-sync 명령어 시뮬레이션
   */
  async runSyncWorkflow(specId: string): Promise<WorkflowResult> {
    try {
      // 1. 문서 동기화 커밋
      const syncCommitMessage = `${GitCommitTemplates.DOCS}: Sync ${specId} documentation`;
      const syncResult = await this.gitManager.commitChanges(syncCommitMessage);

      // 2. Team 모드인 경우 PR 상태 업데이트 (Draft → Ready)
      if (this.config.mode === 'team') {
        // GitHub CLI를 사용하여 PR 상태 변경 (실제 구현 시)
        // await this.updatePullRequestStatus(specId, 'ready');
      }

      // 3. 태그 생성 (완료 마킹)
      // const tagName = `${specId}-completed`;
      // await this.gitManager.createTag(tagName, `SPEC ${specId} completed`);

      return {
        success: true,
        stage: SpecWorkflowStage.SYNC,
        commitHash: syncResult.hash,
        message: `Sync workflow for ${specId} completed successfully`,
      };
    } catch (error) {
      return {
        success: false,
        stage: SpecWorkflowStage.SYNC,
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
      // Git에서 merged 브랜치 목록 조회 (실제 구현 시)
      // const mergedBranches = await this.gitManager.getMergedBranches();
      // const branchesToDelete = mergedBranches.filter(branch => !excludeBranches.includes(branch));

      // 브랜치 삭제 (실제 구현 시)
      // for (const branch of branchesToDelete) {
      //   await this.gitManager.deleteBranch(branch);
      // }

      // return branchesToDelete;
      return []; // 플레이스홀더
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
      const releaseBranch = `release/${version}`;
      await this.gitManager.createBranch(releaseBranch, 'develop');

      // 2. 버전 업데이트 커밋
      const versionCommitMessage = `${GitCommitTemplates.CHORE}: Bump version to ${version}`;
      const versionResult =
        await this.gitManager.commitChanges(versionCommitMessage);

      // 3. 릴리스 태그 생성 (실제 구현 시)
      // await this.gitManager.createTag(`v${version}`, releaseNotes);

      // 4. Team 모드인 경우 릴리스 PR 생성
      let pullRequestUrl: string | undefined;
      if (this.config.mode === 'team') {
        pullRequestUrl = await this.createReleasePullRequest(
          version,
          releaseBranch,
          releaseNotes
        );
      }

      const result: WorkflowResult = {
        success: true,
        stage: SpecWorkflowStage.SYNC,
        branchName: releaseBranch,
        commitHash: versionResult.hash,
        message: `Release ${version} workflow completed successfully`,
      };
      if (pullRequestUrl) {
        result.pullRequestUrl = pullRequestUrl;
      }
      return result;
    } catch (error) {
      return {
        success: false,
        stage: SpecWorkflowStage.INIT,
        message: `Release workflow failed: ${(error as Error).message}`,
      };
    }
  }

  // === Private 헬퍼 메서드 ===

  /**
   * SPEC 디렉토리 구조 생성
   */
  private async createSpecStructure(
    specId: string,
    description: string
  ): Promise<void> {
    const fs = await import('fs-extra');
    const path = await import('node:path');

    const specDir = path.join(process.cwd(), '.moai', 'specs', specId);
    await fs.ensureDir(specDir);

    // SPEC 파일들 생성
    const specContent = `# ${specId} Specification

## Description
${description}

## Requirements
- [ ] Requirement 1
- [ ] Requirement 2

## Implementation Plan
- [ ] Step 1
- [ ] Step 2

## Test Plan
- [ ] Test case 1
- [ ] Test case 2
`;

    await fs.writeFile(path.join(specDir, 'spec.md'), specContent);
    await fs.writeFile(
      path.join(specDir, 'plan.md'),
      '# Implementation Plan\n\nTBD\n'
    );
    await fs.writeFile(
      path.join(specDir, 'acceptance.md'),
      '# Acceptance Criteria\n\nTBD\n'
    );
  }

  /**
   * Draft Pull Request 생성
   */
  private async createDraftPullRequest(
    specId: string,
    branchName: string,
    description: string
  ): Promise<string> {
    const prOptions: CreatePullRequestOptions = {
      title: `SPEC ${specId}: ${description}`,
      body: `## SPEC ${specId} Implementation

### Description
${description}

### Checklist
- [x] SPEC documentation created
- [ ] TDD implementation completed
- [ ] Documentation synchronized
- [ ] Tests passing

🤖 Generated with [MoAI-ADK](https://github.com/your-org/moai-adk)`,
      baseBranch: 'main',
      headBranch: branchName,
      draft: true,
      labels: ['spec', 'wip'],
    };

    return await this.gitManager.createPullRequest(prOptions);
  }

  /**
   * 릴리스 Pull Request 생성
   */
  private async createReleasePullRequest(
    version: string,
    branchName: string,
    releaseNotes: string
  ): Promise<string> {
    const prOptions: CreatePullRequestOptions = {
      title: `Release ${version}`,
      body: `## Release ${version}

### Release Notes
${releaseNotes}

### Changes
- Automated version bump to ${version}
- All SPEC implementations completed

🤖 Generated with [MoAI-ADK](https://github.com/your-org/moai-adk)`,
      baseBranch: 'main',
      headBranch: branchName,
      draft: false,
      labels: ['release'],
    };

    return await this.gitManager.createPullRequest(prOptions);
  }
}

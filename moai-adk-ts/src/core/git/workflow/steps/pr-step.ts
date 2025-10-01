// @CODE:GIT-004 | SPEC: SPEC-GIT-001.md | TEST: src/core/git/workflow/__tests__/workflow-automation.test.ts
// Related: @CODE:GIT-004:API

/**
 * @file PR Step
 * @author MoAI Team
 *
 * @fileoverview Pull Request 생성 단계
 */

import type { CreatePullRequestOptions } from '../../../../types/git';
import type { GitManager } from '../../git-manager';

/**
 * PR 생성 단계
 */
export class PullRequestStep {
  constructor(private readonly gitManager: GitManager) {}

  /**
   * Draft Pull Request 생성
   */
  async createDraftPR(
    specId: string,
    branchName: string,
    description: string
  ): Promise<string> {
    const options: CreatePullRequestOptions = {
      title: `SPEC ${specId}: ${description}`,
      body: this.generateSpecPRBody(specId, description),
      baseBranch: 'main',
      headBranch: branchName,
      draft: true,
      labels: ['spec', 'wip'],
    };

    return await this.gitManager.createPullRequest(options);
  }

  /**
   * 릴리스 Pull Request 생성
   */
  async createReleasePR(
    version: string,
    branchName: string,
    releaseNotes: string
  ): Promise<string> {
    const options: CreatePullRequestOptions = {
      title: `Release ${version}`,
      body: this.generateReleasePRBody(version, releaseNotes),
      baseBranch: 'main',
      headBranch: branchName,
      draft: false,
      labels: ['release'],
    };

    return await this.gitManager.createPullRequest(options);
  }

  /**
   * SPEC PR 본문 생성
   */
  private generateSpecPRBody(specId: string, description: string): string {
    return `## SPEC ${specId} Implementation

### Description
${description}

### Checklist
- [x] SPEC documentation created
- [ ] TDD implementation completed
- [ ] Documentation synchronized
- [ ] Tests passing

🤖 Generated with [MoAI-ADK](https://github.com/your-org/moai-adk)`;
  }

  /**
   * 릴리스 PR 본문 생성
   */
  private generateReleasePRBody(version: string, releaseNotes: string): string {
    return `## Release ${version}

### Release Notes
${releaseNotes}

### Changes
- Automated version bump to ${version}
- All SPEC implementations completed

🤖 Generated with [MoAI-ADK](https://github.com/your-org/moai-adk)`;
  }
}

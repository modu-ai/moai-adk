/**
 * @API:GITHUB-INTEGRATION-001 GitHub API Integration
 * @FEATURE:TEAM-MODE-001 팀 모드 GitHub 연동
 *
 * GitHub Integration for Team Mode
 * SPEC-012 Week 2 Track D: Git System Integration
 *
 * @TASK:GITHUB-CLI-INTEGRATION-001 GitHub CLI 통합
 * @DESIGN:TEAM-WORKFLOW-001 팀 워크플로우 설계
 * @API:GITHUB-CLI-001 GitHub CLI 기반 API 래퍼
 *
 * @fileoverview GitHub API integration using GitHub CLI
 */

import { execa } from 'execa';
import type {
  CreateRepositoryOptions,
  CreatePullRequestOptions,
  GitConfig,
} from '../../types/git';
import { GitHubDefaults } from './constants';

/**
 * GitHub CLI를 사용한 GitHub 연동
 */
export class GitHubIntegration {
  constructor(config: GitConfig) {
    if (config.mode !== 'team') {
      throw new Error('GitHub integration is only available in team mode');
    }
    // Config is validated but not stored as it's not used in current implementation
  }

  /**
   * GitHub CLI 설치 확인
   */
  async isGitHubCliAvailable(): Promise<boolean> {
    try {
      await execa('gh', ['--version']);
      return true;
    } catch {
      return false;
    }
  }

  /**
   * GitHub CLI 인증 상태 확인
   */
  async isAuthenticated(): Promise<boolean> {
    try {
      await execa('gh', ['auth', 'status']);
      return true;
    } catch {
      return false;
    }
  }

  /**
   * GitHub 저장소 생성
   */
  async createRepository(options: CreateRepositoryOptions): Promise<void> {
    if (!(await this.isGitHubCliAvailable())) {
      throw new Error(
        'GitHub CLI is not installed. Please install gh CLI first.'
      );
    }

    if (!(await this.isAuthenticated())) {
      throw new Error(
        'GitHub CLI is not authenticated. Please run "gh auth login" first.'
      );
    }

    try {
      const args = ['repo', 'create', options.name];

      if (options.description) {
        args.push('--description', options.description);
      }

      if (options.private) {
        args.push('--private');
      } else {
        args.push('--public');
      }

      if (options.autoInit) {
        args.push('--add-readme');
      }

      if (options.gitignoreTemplate) {
        args.push('--gitignore', options.gitignoreTemplate);
      }

      if (options.licenseTemplate) {
        args.push('--license', options.licenseTemplate);
      }

      await execa('gh', args);
    } catch (error) {
      throw new Error(
        `Failed to create GitHub repository: ${(error as Error).message}`
      );
    }
  }

  /**
   * Pull Request 생성
   */
  async createPullRequest(options: CreatePullRequestOptions): Promise<string> {
    if (!(await this.isGitHubCliAvailable())) {
      throw new Error(
        'GitHub CLI is not installed. Please install gh CLI first.'
      );
    }

    if (!(await this.isAuthenticated())) {
      throw new Error(
        'GitHub CLI is not authenticated. Please run "gh auth login" first.'
      );
    }

    try {
      const args = ['pr', 'create'];

      args.push('--title', options.title);
      args.push('--body', options.body || GitHubDefaults.PR_TEMPLATE);
      args.push('--base', options.baseBranch);

      if (options.headBranch) {
        args.push('--head', options.headBranch);
      }

      if (options.draft) {
        args.push('--draft');
      }

      if (options.assignees && options.assignees.length > 0) {
        args.push('--assignee', options.assignees.join(','));
      }

      if (options.reviewers && options.reviewers.length > 0) {
        args.push('--reviewer', options.reviewers.join(','));
      }

      if (options.labels && options.labels.length > 0) {
        args.push('--label', options.labels.join(','));
      }

      const result = await execa('gh', args);
      return result.stdout.trim();
    } catch (error) {
      throw new Error(
        `Failed to create pull request: ${(error as Error).message}`
      );
    }
  }

  /**
   * 저장소 정보 조회
   */
  async getRepositoryInfo(): Promise<{
    owner: string;
    name: string;
    url: string;
    private: boolean;
  }> {
    try {
      const result = await execa('gh', [
        'repo',
        'view',
        '--json',
        'owner,name,url,isPrivate',
      ]);
      const repoInfo = JSON.parse(result.stdout);

      return {
        owner: repoInfo.owner.login,
        name: repoInfo.name,
        url: repoInfo.url,
        private: repoInfo.isPrivate,
      };
    } catch (error) {
      throw new Error(
        `Failed to get repository info: ${(error as Error).message}`
      );
    }
  }

  /**
   * Issue 생성
   */
  async createIssue(
    title: string,
    body?: string,
    labels?: string[]
  ): Promise<string> {
    try {
      const args = ['issue', 'create', '--title', title];

      if (body) {
        args.push('--body', body);
      } else {
        args.push('--body', GitHubDefaults.ISSUE_TEMPLATE);
      }

      if (labels && labels.length > 0) {
        args.push('--label', labels.join(','));
      }

      const result = await execa('gh', args);
      return result.stdout.trim();
    } catch (error) {
      throw new Error(`Failed to create issue: ${(error as Error).message}`);
    }
  }

  /**
   * 라벨 생성
   */
  async createLabels(): Promise<void> {
    try {
      for (const label of GitHubDefaults.DEFAULT_LABELS) {
        try {
          await execa('gh', [
            'label',
            'create',
            label.name,
            '--description',
            label.description,
            '--color',
            label.color,
          ]);
        } catch {
          // 라벨이 이미 존재하는 경우 무시
        }
      }
    } catch (error) {
      throw new Error(`Failed to create labels: ${(error as Error).message}`);
    }
  }

  /**
   * GitHub Actions 워크플로우 파일 생성
   */
  async createWorkflowFile(
    workflowName: string,
    content: string
  ): Promise<void> {
    const fs = await import('fs-extra');
    const path = await import('path');

    try {
      const workflowDir = path.join(process.cwd(), '.github', 'workflows');
      await fs.ensureDir(workflowDir);

      const workflowFile = path.join(workflowDir, `${workflowName}.yml`);
      await fs.writeFile(workflowFile, content);
    } catch (error) {
      throw new Error(
        `Failed to create workflow file: ${(error as Error).message}`
      );
    }
  }

  /**
   * 기본 GitHub Actions 워크플로우 생성
   */
  async setupDefaultWorkflows(): Promise<void> {
    const ciWorkflow = `name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    runs-on: ubuntu-latest

    strategy:
      matrix:
        node-version: [18.x, 20.x]

    steps:
    - uses: actions/checkout@v3

    - name: Use Node.js \${{ matrix.node-version }}
      uses: actions/setup-node@v3
      with:
        node-version: \${{ matrix.node-version }}
        cache: 'npm'

    - run: npm ci
    - run: npm run build --if-present
    - run: npm test
    - run: npm run lint

    - name: Upload coverage reports
      uses: codecov/codecov-action@v3
      if: matrix.node-version == '20.x'
`;

    const releaseWorkflow = `name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Use Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '20.x'
        cache: 'npm'
        registry-url: 'https://registry.npmjs.org'

    - run: npm ci
    - run: npm run build
    - run: npm test

    - name: Publish to npm
      run: npm publish
      env:
        NODE_AUTH_TOKEN: \${{ secrets.NPM_TOKEN }}

    - name: Create GitHub Release
      uses: actions/create-release@v1
      env:
        GITHUB_TOKEN: \${{ secrets.GITHUB_TOKEN }}
      with:
        tag_name: \${{ github.ref }}
        release_name: Release \${{ github.ref }}
        draft: false
        prerelease: false
`;

    await this.createWorkflowFile('ci', ciWorkflow);
    await this.createWorkflowFile('release', releaseWorkflow);
  }
}

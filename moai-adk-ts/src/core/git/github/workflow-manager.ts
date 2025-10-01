// @CODE:GITHUB-WORKFLOW-001 | SPEC: github-integration
// Related: @CODE:GITHUB-001

/**
 * @file GitHub Actions workflow manager
 * @author MoAI Team
 */

import * as fs from 'fs-extra';
import * as path from 'node:path';
import type { WorkflowTemplate } from './types';

/**
 * GitHub Actions 워크플로우 파일 관리 담당
 */
export class WorkflowManager {
  /**
   * GitHub Actions 워크플로우 파일 생성
   */
  async createWorkflowFile(
    workflowName: string,
    content: string
  ): Promise<void> {
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
   * 기본 CI 워크플로우 템플릿
   */
  private getCIWorkflowTemplate(): WorkflowTemplate {
    return {
      name: 'ci',
      content: `name: CI

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
`,
    };
  }

  /**
   * 기본 Release 워크플로우 템플릿
   */
  private getReleaseWorkflowTemplate(): WorkflowTemplate {
    return {
      name: 'release',
      content: `name: Release

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
`,
    };
  }

  /**
   * 기본 GitHub Actions 워크플로우 생성
   */
  async setupDefaultWorkflows(): Promise<void> {
    const ciWorkflow = this.getCIWorkflowTemplate();
    const releaseWorkflow = this.getReleaseWorkflowTemplate();

    await this.createWorkflowFile(ciWorkflow.name, ciWorkflow.content);
    await this.createWorkflowFile(releaseWorkflow.name, releaseWorkflow.content);
  }
}

// @CODE:GITHUB-ISSUE-001 | SPEC: github-integration
// Related: @CODE:GITHUB-001

/**
 * @file GitHub Issue manager
 * @author MoAI Team
 */

import { execa } from 'execa';
import { GitHubDefaults } from '../constants/index';
import type { AuthChecker } from './auth-checker';

/**
 * GitHub Issue 생성 및 관리 담당
 */
export class IssueManager {
  constructor(private readonly authChecker: AuthChecker) {}

  /**
   * Issue 생성
   */
  async createIssue(
    title: string,
    body?: string,
    labels?: string[]
  ): Promise<string> {
    await this.authChecker.ensureAuthenticated();

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
    await this.authChecker.ensureAuthenticated();

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
}

// @CODE:GITHUB-PR-001 | SPEC: github-integration
// Related: @CODE:GITHUB-001

/**
 * @file GitHub Pull Request manager
 * @author MoAI Team
 */

import { execa } from 'execa';
import type { CreatePullRequestOptions } from '@/types/git';
import { GitHubDefaults } from '../constants/index';
import type { AuthChecker } from './auth-checker';

/**
 * GitHub Pull Request 생성 및 관리 담당
 */
export class PullRequestManager {
  constructor(private readonly authChecker: AuthChecker) {}

  /**
   * Pull Request 생성
   */
  async createPullRequest(options: CreatePullRequestOptions): Promise<string> {
    await this.authChecker.ensureAuthenticated();

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
}

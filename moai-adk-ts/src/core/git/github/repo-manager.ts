// @CODE:GITHUB-REPO-001 | SPEC: github-integration
// Related: @CODE:GITHUB-001

/**
 * @file GitHub Repository manager
 * @author MoAI Team
 */

import { execa } from 'execa';
import type { CreateRepositoryOptions } from '@/types/git';
import type { AuthChecker } from './auth-checker';
import type { RepositoryInfo } from './types';

/**
 * GitHub 저장소 생성 및 정보 조회 담당
 */
export class RepositoryManager {
  constructor(private readonly authChecker: AuthChecker) {}

  /**
   * GitHub 저장소 생성
   */
  async createRepository(options: CreateRepositoryOptions): Promise<void> {
    await this.authChecker.ensureAuthenticated();

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
   * 저장소 정보 조회
   */
  async getRepositoryInfo(): Promise<RepositoryInfo> {
    await this.authChecker.ensureAuthenticated();

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
}

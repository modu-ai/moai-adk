// @CODE:GITHUB-AUTH-001 | SPEC: github-integration
// Related: @CODE:GITHUB-001

/**
 * @file GitHub CLI authentication checker
 * @author MoAI Team
 */

import { execa } from 'execa';
import type { AuthStatus } from './types';

/**
 * GitHub CLI 인증 상태 확인 담당
 */
export class AuthChecker {
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
   * 종합 인증 상태 조회
   */
  async getAuthStatus(): Promise<AuthStatus> {
    const cliAvailable = await this.isGitHubCliAvailable();
    const authenticated = cliAvailable ? await this.isAuthenticated() : false;

    return {
      cliAvailable,
      authenticated,
    };
  }

  /**
   * 인증 전제조건 검증 (CLI 설치 + 인증 완료)
   * @throws Error if prerequisites are not met
   */
  async ensureAuthenticated(): Promise<void> {
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
  }
}

// @CODE:GITHUB-TYPES-001 | SPEC: github-integration
// Related: @CODE:GITHUB-001

/**
 * @file GitHub integration type definitions
 * @author MoAI Team
 */

/**
 * GitHub 저장소 정보
 */
export interface RepositoryInfo {
  readonly owner: string;
  readonly name: string;
  readonly url: string;
  readonly private: boolean;
}

/**
 * GitHub CLI 인증 상태
 */
export interface AuthStatus {
  readonly authenticated: boolean;
  readonly cliAvailable: boolean;
}

/**
 * GitHub 워크플로우 템플릿
 */
export interface WorkflowTemplate {
  readonly name: string;
  readonly content: string;
}

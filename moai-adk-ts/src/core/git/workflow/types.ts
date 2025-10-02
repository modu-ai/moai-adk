// @CODE:GIT-004 | Chain: @SPEC:GIT-001 -> @SPEC:GIT-001 -> @CODE:GIT-004 -> @CODE:GIT-004
// Related: @CODE:GIT-004

/**
 * @file Workflow Types
 * @author MoAI Team
 *
 * @fileoverview 워크플로우 타입 정의
 */

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
  pullRequestUrl: string | undefined;
  message: string;
}

/**
 * 워크플로우 단계 컨텍스트
 */
export interface WorkflowContext {
  specId: string;
  description?: string;
  branchName?: string;
  commitHash?: string;
}

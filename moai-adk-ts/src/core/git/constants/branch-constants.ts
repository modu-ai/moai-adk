// @CODE:REFACTOR-004 연결: @TEST:REFACTOR-004 -> @CODE:REFACTOR-004
/**
 * @file Git Branch Naming Rules
 * @author MoAI Team
 * @tags @CODE:REFACTOR-004 @CODE:GIT-NAMING-RULES-001:DATA
 * @description Git 브랜치 명명 규칙 및 검증 로직
 */

/**
 * Git 브랜치 명명 규칙
 * @tags @CODE:GIT-NAMING-RULES-001:DATA
 */
export const GitNamingRules = {
  FEATURE_PREFIX: 'feature/',
  BUGFIX_PREFIX: 'bugfix/',
  HOTFIX_PREFIX: 'hotfix/',
  SPEC_PREFIX: 'spec/',
  CHORE_PREFIX: 'chore/',

  /**
   * 기능 브랜치명 생성
   * @tags @CODE:CREATE-FEATURE-BRANCH-001:API
   */
  createFeatureBranch: (name: string): string => `feature/${name}`,

  /**
   * SPEC 브랜치명 생성
   * @tags @CODE:CREATE-SPEC-BRANCH-001:API
   */
  createSpecBranch: (specId: string): string => `spec/${specId}`,

  /**
   * 버그픽스 브랜치명 생성
   * @tags @CODE:CREATE-BUGFIX-BRANCH-001:API
   */
  createBugfixBranch: (name: string): string => `bugfix/${name}`,

  /**
   * 핫픽스 브랜치명 생성
   * @tags @CODE:CREATE-HOTFIX-BRANCH-001:API
   */
  createHotfixBranch: (name: string): string => `hotfix/${name}`,

  /**
   * 브랜치명 검증
   * @tags @CODE:VALIDATE-BRANCH-NAME-001:API
   */
  isValidBranchName: (name: string): boolean => {
    // Git 브랜치명 규칙: 알파벳, 숫자, 하이픈, 슬래시 허용
    const pattern = /^[a-zA-Z0-9/\-_.]+$/;
    return (
      pattern.test(name) &&
      !name.startsWith('-') &&
      !name.endsWith('-') &&
      !name.includes('//') &&
      !name.includes('..')
    );
  },
} as const;

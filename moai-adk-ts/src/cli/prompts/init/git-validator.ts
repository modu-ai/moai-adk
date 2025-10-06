// @CODE:INSTALL-001 | SPEC: SPEC-INSTALL-001.md | TEST: __tests__/cli/prompts/git-validator.test.ts
/**
 * @file Git installation validator
 * @author MoAI Team
 * @tags @CODE:INSTALL-001:GIT-VALIDATION
 */

import { execa } from 'execa';

/**
 * Git validation result
 */
export interface GitValidationResult {
  isValid: boolean;
  version?: string;
  error?: string;
}

/**
 * Validate Git installation
 * @returns Validation result with version or error message
 */
export async function validateGitInstallation(): Promise<GitValidationResult> {
  try {
    const result = await execa('git', ['--version']);
    const versionMatch = result.stdout.match(/git version (\d+\.\d+\.\d+)/);
    const version = versionMatch ? versionMatch[1] : 'unknown';

    return {
      isValid: true,
      version,
    };
  } catch {
    return {
      isValid: false,
      error: `❌ Git이 설치되지 않았습니다.

MoAI-ADK는 Git을 필수로 사용합니다.
다음 방법으로 Git을 설치해주세요:

macOS: brew install git
Ubuntu: sudo apt-get install git
Windows: https://git-scm.com/download/win

설치 후 다시 시도해주세요.`,
    };
  }
}

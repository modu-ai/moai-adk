/**
 * @file Validation functions for init prompts
 * @author MoAI Team
 * @tags @CODE:INIT-004 | SPEC: SPEC-INIT-004.md
 * Related: @CODE:INIT-VALIDATORS-001 | Chain: @SPEC:INTERACTIVE-INIT-019 -> @CODE:INTERACTIVE-INIT-019
 */

import { validateGitHubUrl as validateGitHubUrlCore } from '@/utils/git-detector';
import { InputValidator } from '@/utils/input-validator';

/**
 * Validate project name for init prompt
 * @param input User input
 * @returns Validation message or true
 */
export function validateProjectName(input: string): string | true {
  const result = InputValidator.validateProjectName(input, {
    maxLength: 50,
    allowSpaces: false,
  });
  return result.isValid ? true : result.errors.join(', ');
}

/**
 * Validate GitHub URL format (integrated with git-detector)
 * @param input User input
 * @returns Validation message or true
 */
export function validateGitHubUrl(input: string): string | true {
  if (!validateGitHubUrlCore(input)) {
    return 'Please enter a valid GitHub URL (https://github.com/username/repo or git@github.com:username/repo)';
  }
  return true;
}

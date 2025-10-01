/**
 * @file Validation functions for init prompts
 * @author MoAI Team
 * @tags @CODE:INIT-VALIDATORS-001 | Chain: @SPEC:INTERACTIVE-INIT-019 -> @CODE:INTERACTIVE-INIT-019
 * Related: @DOC:INTERACTIVE-INIT-019
 */

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
 * Validate GitHub URL format
 * @param input User input
 * @returns Validation message or true
 */
export function validateGitHubUrl(input: string): string | true {
  const githubRegex = /^https:\/\/github\.com\/[\w-]+\/[\w-]+$/;
  if (!githubRegex.test(input)) {
    return 'Please enter a valid GitHub URL (https://github.com/username/repo)';
  }
  return true;
}

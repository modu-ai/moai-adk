// @CODE:INSTALL-001 | SPEC: SPEC-INSTALL-001.md | TEST: __tests__/cli/prompts/spec-workflow.test.ts
/**
 * @file SPEC Workflow configuration module
 * @author MoAI Team
 * @tags @CODE:INSTALL-001:SPEC-WORKFLOW
 */

import inquirer from 'inquirer';
import type { Answers } from 'inquirer';

/**
 * SPEC configuration
 */
export interface SpecConfig {
  constitution: {
    enforce_spec: boolean;
  };
}

/**
 * Get SPEC workflow prompt for Personal mode
 * @returns Inquirer prompt configuration
 */
export function getSpecWorkflowPersonalPrompt(): Answers {
  return [
    {
      type: 'confirm',
      name: 'enforceSpec',
      message: 'SPEC-First Workflow를 사용할까요? (권장)',
      default: true,
    },
  ];
}

/**
 * Prompt SPEC workflow for Personal mode
 * @returns User answer for enforce_spec
 */
export async function promptSpecWorkflowPersonal(): Promise<{ enforceSpec: boolean }> {
  const prompts = getSpecWorkflowPersonalPrompt();
  const answers = await inquirer.prompt(prompts);
  return { enforceSpec: answers.enforceSpec };
}

/**
 * Get SPEC config for Team mode (always enforced)
 * @returns SPEC config with enforce_spec: true
 */
export function getSpecConfigForTeam(): { enforceSpec: boolean } {
  return { enforceSpec: true };
}

/**
 * Prompt SPEC workflow based on mode
 * @param mode Project mode ('personal' | 'team')
 * @returns User answer or auto-config
 */
export async function promptSpecWorkflow(
  mode: 'personal' | 'team'
): Promise<{ enforceSpec: boolean }> {
  if (mode === 'team') {
    return getSpecConfigForTeam();
  }
  return promptSpecWorkflowPersonal();
}

/**
 * Build SPEC configuration
 * @param _mode Project mode (unused, reserved for future use)
 * @param enforceSpec Whether to enforce SPEC workflow
 * @returns Complete SPEC configuration
 */
export async function buildSpecConfig(
  _mode: 'personal' | 'team',
  enforceSpec: boolean
): Promise<SpecConfig> {
  return {
    constitution: {
      enforce_spec: enforceSpec,
    },
  };
}

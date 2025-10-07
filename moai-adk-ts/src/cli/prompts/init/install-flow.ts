// @CODE:INSTALL-001 | SPEC: SPEC-INSTALL-001.md | TEST: __tests__/cli/prompts/install-flow.test.ts
/**
 * @file Complete installation flow orchestration
 * @author MoAI Team
 * @tags @CODE:INSTALL-001:INSTALL-FLOW
 */

import { collectDeveloperInfo } from './developer-info';
import { validateGitInstallation } from './git-validator';
import { promptAutoPR, promptDraftPR } from './pr-config';
import { promptSpecWorkflow } from './spec-workflow';

/**
 * Installation flow result
 */
export interface InstallFlowResult {
  mode: 'personal' | 'team';
  developer: {
    name: string;
    timestamp: string;
  };
  constitution: {
    enforce_spec: boolean;
  };
  git_strategy?: {
    team: {
      auto_pr?: boolean | undefined;
      draft_pr?: boolean | undefined;
    };
  };
}

/**
 * Config validation result
 */
export interface ConfigValidationResult {
  isValid: boolean;
  warnings: string[];
}

/**
 * Run complete installation prompts
 * @param mode Project mode
 * @returns Complete installation configuration
 */
export async function runInstallPrompts(
  mode: 'personal' | 'team'
): Promise<InstallFlowResult> {
  // Step 1: Validate Git installation
  const gitValidation = await validateGitInstallation();
  if (!gitValidation.isValid) {
    throw new Error(gitValidation.error);
  }

  // Step 2: Collect developer information
  const developerInfo = await collectDeveloperInfo();

  // Step 3: SPEC workflow configuration
  const specConfig = await promptSpecWorkflow(mode);

  const result: InstallFlowResult = {
    mode,
    developer: {
      name: developerInfo.developerName,
      timestamp: developerInfo.timestamp,
    },
    constitution: {
      enforce_spec: specConfig.enforceSpec,
    },
  };

  // Step 4: Team-specific configuration
  if (mode === 'team') {
    const autoPRResult = await promptAutoPR(mode);
    const draftPRResult = await promptDraftPR(autoPRResult.autoPR ?? false);

    result.git_strategy = {
      team: {
        auto_pr: autoPRResult.autoPR,
        draft_pr: draftPRResult.draftPR,
      },
    };
  }

  return result;
}

/**
 * Validate configuration for backward compatibility
 * @param config Old configuration
 * @returns Validation result with warnings
 */
export function validateConfig(config: any): ConfigValidationResult {
  const warnings: string[] = [];

  if (!config.developer) {
    warnings.push('developer 필드가 없습니다');
  }

  return {
    isValid: true,
    warnings,
  };
}

/**
 * Migrate old configuration to new format
 * @param config Old configuration
 * @returns Migrated configuration
 */
export function migrateConfig(config: any): any {
  const migrated = { ...config };

  if (!migrated.constitution) {
    migrated.constitution = {};
  }

  if (migrated.constitution.enforce_spec === undefined) {
    migrated.constitution.enforce_spec = true;
  }

  return migrated;
}

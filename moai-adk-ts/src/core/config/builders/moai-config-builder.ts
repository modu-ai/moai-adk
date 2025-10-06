// @CODE:CFG-MOAI-BUILDER-001 |
// Related: @CODE:CFG-001

/**
 * @file MoAI config builder
 * @author MoAI Team
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { logger } from '../../../utils/winston-logger.js';
import type {
  MoAIConfig,
  MoAIConfigResult,
  ProjectConfigInput,
} from '../types';
import {
  backupConfigFile,
  ensureDirectoryExists,
  writeJsonFile,
} from '../utils/config-file-utils.js';

/**
 * Build .moai/config.json
 */
export async function buildMoAIConfig(
  configPath: string,
  config: ProjectConfigInput,
  version: string
): Promise<MoAIConfigResult> {
  try {
    const configDir = path.dirname(configPath);
    let backupCreated = false;

    ensureDirectoryExists(configDir);

    // Backup existing config if it exists
    if (fs.existsSync(configPath) && config.backup !== false) {
      const backupResult = await backupConfigFile(configPath);
      backupCreated = backupResult.success;
    }

    const moaiConfig: MoAIConfig = {
      _meta: {
        '@CODE:CONFIG-STRUCTURE-001': '@DOC:JSON-CONFIG-001',
        '@SPEC:PROJECT-CONFIG-001': '@SPEC:MOAI-CONFIG-001',
      },
      project: {
        name: config.projectName,
        version,
        mode: config.mode,
        description: `A ${config.projectName} project built with MoAI-ADK`,
        initialized: true,
        created_at: new Date().toISOString(),
        locale: 'ko',
      },
      constitution: {
        enforce_tdd: true,
        require_tags: true,
        test_coverage_target: 85,
        simplicity_threshold: 5,
        principles: {
          simplicity: {
            max_projects: 5,
            notes:
              '기본 권장값. 프로젝트 규모에 따라 .moai/config.json 또는 SPEC/ADR로 근거와 함께 조정하세요.',
          },
        },
      },
      git_strategy: {
        personal: {
          auto_checkpoint: true,
          auto_commit: true,
          branch_prefix: 'feature/',
          checkpoint_interval: 300,
          cleanup_days: 7,
          max_checkpoints: 50,
        },
        team: {
          auto_pr: true,
          develop_branch: 'develop',
          draft_pr: true,
          feature_prefix: 'feature/SPEC-',
          main_branch: 'main',
          use_gitflow: true,
        },
      },
      tags: {
        auto_sync: true,
        storage_type: 'code_scan',
        categories: [
          'REQ',
          'DESIGN',
          'TASK',
          'TEST',
          'FEATURE',
          'API',
          'UI',
          'DATA',
        ],
        code_scan_policy: {
          no_intermediate_cache: true,
          realtime_validation: true,
          scan_tools: ['rg', 'grep'],
          scan_command: "rg '@TAG' -n",
          philosophy: 'TAG의 진실은 코드 자체에만 존재',
        },
      },
      pipeline: {
        available_commands: [
          '/alfred:1-spec',
          '/alfred:2-build',
          '/alfred:3-sync',
          '/alfred:4-debug',
        ],
        current_stage: 'initialized',
      },
    };

    writeJsonFile(configPath, moaiConfig);

    logger.info(`MoAI config created: ${configPath}`);
    return {
      success: true,
      filePath: configPath,
      config: moaiConfig,
      backupCreated,
    };
  } catch (error) {
    const errorMessage =
      error instanceof Error ? error.message : 'Unknown error';
    logger.error('Error creating MoAI config:', errorMessage);
    return {
      success: false,
      error: errorMessage,
    };
  }
}

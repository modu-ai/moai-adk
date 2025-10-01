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
      projectName: config.projectName,
      version,
      mode: config.mode,
      runtime: config.runtime,
      techStack: config.techStack,
      features: {
        tdd: true,
        tagSystem: true,
        gitAutomation: config.mode === 'team',
        documentSync: config.mode === 'team',
      },
      directories: {
        moai: '.moai',
        claude: '.claude',
        specs: '.moai/specs',
        templates: '.moai/templates',
      },
      createdAt: new Date(),
      updatedAt: new Date(),
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

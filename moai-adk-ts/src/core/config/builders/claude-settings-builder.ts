// @CODE:CFG-CLAUDE-BUILDER-001 |
// Related: @CODE:CFG-001

/**
 * @file Claude settings builder
 * @author MoAI Team
 */

import * as path from 'node:path';
import { logger } from '../../../utils/winston-logger.js';
import type {
  ClaudeSettings,
  ClaudeSettingsResult,
  ProjectConfigInput,
} from '../types';
import {
  getCommandShortcuts,
  getEnabledAgents,
  getEnabledCommands,
  getEnabledHooks,
  getHookConfiguration,
} from '../utils/config-helpers.js';
import {
  ensureDirectoryExists,
  writeJsonFile,
} from '../utils/config-file-utils.js';

/**
 * Build Claude Code settings.json
 */
export async function buildClaudeSettings(
  settingsPath: string,
  config: ProjectConfigInput
): Promise<ClaudeSettingsResult> {
  try {
    const settingsDir = path.dirname(settingsPath);
    ensureDirectoryExists(settingsDir);

    const settings: ClaudeSettings = {
      mode: config.mode,
      agents: {
        enabled: getEnabledAgents(config.mode),
        disabled: [],
      },
      commands: {
        enabled: getEnabledCommands(config.mode),
        shortcuts: getCommandShortcuts(),
      },
      hooks: {
        enabled: getEnabledHooks(config.mode),
        configuration: getHookConfiguration(config),
      },
      outputStyles: {
        default: 'study',
        available: ['study', 'professional', 'debug', 'minimal'],
      },
      features: {
        autoSync: config.mode === 'team',
        gitIntegration: true,
        tagTracking: true,
      },
      security: {
        allowedCommands: ['git', 'npm', 'node', 'python'],
        blockedPatterns: ['rm -rf', 'sudo', 'chmod 777'],
        requireApproval: ['--force', '--hard'],
      },
    };

    writeJsonFile(settingsPath, settings);

    logger.info(`Claude settings created: ${settingsPath}`);
    return {
      success: true,
      filePath: settingsPath,
      settings,
    };
  } catch (error) {
    const errorMessage =
      error instanceof Error ? error.message : 'Unknown error';
    logger.error('Error creating Claude settings:', errorMessage);
    return {
      success: false,
      error: errorMessage,
    };
  }
}

/**
 * @file Logger environment-based configuration
 * @author MoAI Team
 * @tags @CONFIG:WINSTON-LOGGER-001
 */

import type winston from 'winston';
import type { LoggerOptions } from '../types/logger.js';

/**
 * Environment-specific logger configuration
 */
interface EnvironmentConfig {
  level: string;
  enableFile: boolean;
  enableConsole: boolean;
}

/**
 * Get logger configuration based on NODE_ENV
 */
export function getLoggerConfig(): EnvironmentConfig {
  const env = process.env['NODE_ENV'] || 'development';

  const configs: Record<string, EnvironmentConfig> = {
    development: { level: 'debug', enableFile: false, enableConsole: true },
    test: { level: 'error', enableFile: false, enableConsole: false },
    production: { level: 'info', enableFile: true, enableConsole: true },
  };

  return configs[env] ?? configs['development']!;
}

/**
 * Merge user options with environment config
 */
export function mergeLoggerOptions(
  userOptions?: LoggerOptions
): Required<Omit<LoggerOptions, 'transports'>> & {
  transports?: winston.transport[];
} {
  const envConfig = getLoggerConfig();
  const logLevel =
    process.env['LOG_LEVEL'] || userOptions?.level || envConfig.level;

  const merged: Required<Omit<LoggerOptions, 'transports'>> & {
    transports?: winston.transport[];
  } = {
    level: logLevel,
    enableFile: userOptions?.enableFile ?? envConfig.enableFile,
    enableConsole: userOptions?.enableConsole ?? envConfig.enableConsole,
    logDir: userOptions?.logDir ?? 'logs',
  };

  // Only include transports if provided
  if (userOptions?.transports) {
    merged.transports = userOptions.transports;
  }

  return merged;
}

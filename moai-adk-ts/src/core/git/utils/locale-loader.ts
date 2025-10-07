// @CODE:GIT-LOCALE-LOADER-001 | SPEC: SPEC-GIT-LOCALE-001.md
/**
 * @file Locale loader from .moai/config.json
 * @author MoAI Team
 * @tags @CODE:GIT-LOCALE-LOADER-001:INFRA
 * @description Load project locale from MoAI configuration
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import type { CommitLocale } from '../constants/commit-message-locales';
import { getValidatedLocale } from '../constants/commit-message-locales';

/**
 * Load locale from .moai/config.json
 * @param workingDir - Working directory path
 * @returns Validated commit locale (defaults to 'en' if not found)
 */
export function loadLocaleFromConfig(workingDir: string): CommitLocale {
  try {
    const configPath = path.join(workingDir, '.moai', 'config.json');

    if (!fs.existsSync(configPath)) {
      return 'en'; // Default to English if config doesn't exist
    }

    const configContent = fs.readFileSync(configPath, 'utf-8');
    const config = JSON.parse(configContent);

    // Extract locale from project.locale
    const locale = config?.project?.locale;

    return getValidatedLocale(locale);
  } catch {
    // Return default locale on any error
    return 'en';
  }
}

/**
 * Check if .moai/config.json exists
 * @param workingDir - Working directory path
 * @returns True if config exists
 */
export function hasConfigFile(workingDir: string): boolean {
  const configPath = path.join(workingDir, '.moai', 'config.json');
  return fs.existsSync(configPath);
}

/**
 * Get locale with fallback chain:
 * 1. .moai/config.json (project.locale)
 * 2. Environment variable (MOAI_LOCALE)
 * 3. Default ('en')
 * @param workingDir - Working directory path
 * @returns Validated commit locale
 */
export function getLocaleWithFallback(workingDir: string): CommitLocale {
  // Try loading from config first
  if (hasConfigFile(workingDir)) {
    const configLocale = loadLocaleFromConfig(workingDir);
    if (configLocale !== 'en') {
      // If explicitly set in config, use it
      return configLocale;
    }
  }

  // Try environment variable
  const envLocale = process.env.MOAI_LOCALE;
  if (envLocale) {
    return getValidatedLocale(envLocale);
  }

  // Default fallback
  return 'en';
}

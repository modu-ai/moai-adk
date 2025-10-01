// @CODE:CFG-FILE-UTILS-001 |
// Related: @CODE:CFG-001

/**
 * @file Configuration file operations utilities
 * @author MoAI Team
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { logger } from '../../../utils/winston-logger.js';
import type { BackupResult, ValidationResult } from '../types';

/**
 * Validate configuration file format and content
 */
export async function validateConfigFile(
  filePath: string
): Promise<ValidationResult> {
  const result: ValidationResult = {
    isValid: true,
    errors: [],
    warnings: [],
    suggestions: [],
  };

  try {
    if (!fs.existsSync(filePath)) {
      result.isValid = false;
      result.errors.push('File does not exist');
      return result;
    }

    const content = fs.readFileSync(filePath, 'utf-8');
    JSON.parse(content); // Validate JSON format

    logger.info(`Configuration file validated: ${filePath}`);
    return result;
  } catch (error) {
    result.isValid = false;
    if (error instanceof SyntaxError) {
      result.errors.push('Invalid JSON format');
    } else {
      result.errors.push(
        error instanceof Error ? error.message : 'Unknown error'
      );
    }
    return result;
  }
}

/**
 * Create backup of configuration file
 */
export async function backupConfigFile(
  filePath: string
): Promise<BackupResult> {
  try {
    if (!fs.existsSync(filePath)) {
      return {
        success: false,
        error: 'File does not exist',
        timestamp: new Date(),
      };
    }

    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const dir = path.dirname(filePath);
    const name = path.basename(filePath, path.extname(filePath));
    const ext = path.extname(filePath);
    const backupPath = path.join(dir, `${name}.backup.${timestamp}${ext}`);

    const content = fs.readFileSync(filePath, 'utf-8');
    fs.writeFileSync(backupPath, content, 'utf-8');

    logger.info(`Backup created: ${backupPath}`);
    return {
      success: true,
      backupPath,
      timestamp: new Date(),
    };
  } catch (error) {
    const errorMessage =
      error instanceof Error ? error.message : 'Unknown error';
    return {
      success: false,
      error: errorMessage,
      timestamp: new Date(),
    };
  }
}

/**
 * Ensure directory exists, create if needed
 */
export function ensureDirectoryExists(dirPath: string): void {
  if (!fs.existsSync(dirPath)) {
    fs.mkdirSync(dirPath, { recursive: true });
  }
}

/**
 * Write JSON file with formatting
 */
export function writeJsonFile(
  filePath: string,
  data: any,
  indent = 2
): void {
  fs.writeFileSync(filePath, JSON.stringify(data, null, indent), 'utf-8');
}

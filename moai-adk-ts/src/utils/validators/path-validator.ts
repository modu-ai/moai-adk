// @CODE:VALIDATOR-003:PATH | Chain: @REQ:QUAL-001 -> @DESIGN:QUAL-001 -> @TASK:QUAL-001 -> @TEST:UTIL-006
// Related: @SECURITY:INPUT-VALIDATION-001

/**
 * @file Path validation module
 * @author MoAI Team
 */

import * as path from 'node:path';
import type { ValidationResult } from './types';
import {
  DANGEROUS_PATH_PATTERNS,
  PATH_VALIDATION_CONSTANTS,
} from './validation-rules';

/**
 * Path validation options
 */
export interface PathValidationOptions {
  readonly mustExist?: boolean;
  readonly mustBeDirectory?: boolean;
  readonly mustBeFile?: boolean;
  readonly allowedExtensions?: readonly string[];
  readonly maxDepth?: number;
}

/**
 * Validate file/directory path
 * @tags @CODE:VALIDATE-PATH-001:API
 */
export async function validatePath(
  inputPath: string,
  options: PathValidationOptions = {}
): Promise<ValidationResult> {
  const {
    mustExist = false,
    mustBeDirectory = false,
    mustBeFile = false,
    allowedExtensions,
    maxDepth = PATH_VALIDATION_CONSTANTS.DEFAULT_MAX_DEPTH,
  } = options;

  const errors: string[] = [];

  // Basic checks
  if (!inputPath || typeof inputPath !== 'string') {
    errors.push('Path must be a non-empty string');
    return { isValid: false, errors };
  }

  const trimmed = inputPath.trim();

  // Path length validation
  if (trimmed.length > PATH_VALIDATION_CONSTANTS.MAX_PATH_LENGTH) {
    errors.push(
      `Path is too long (maximum ${PATH_VALIDATION_CONSTANTS.MAX_PATH_LENGTH} characters)`
    );
  }

  // Dangerous pattern detection
  if (containsDangerousPathPatterns(trimmed)) {
    errors.push('Path contains dangerous patterns');
  }

  // Path depth validation
  const depth = trimmed.split(path.sep).length;
  if (depth > maxDepth) {
    errors.push(`Path depth exceeds maximum (${maxDepth})`);
  }

  // Normalize and resolve path
  let normalizedPath: string;
  try {
    normalizedPath = path.resolve(trimmed);
  } catch (_error) {
    errors.push('Invalid path format');
    return { isValid: false, errors };
  }

  // Path traversal protection
  if (!isPathSafe(normalizedPath, process.cwd())) {
    errors.push('Path traversal detected');
  }

  // File system checks (if required)
  if (mustExist || mustBeDirectory || mustBeFile) {
    const { existsChecks, existsErrors } = await performExistenceChecks(
      normalizedPath,
      {
        mustExist,
        mustBeDirectory,
        mustBeFile,
      }
    );

    if (!existsChecks) {
      errors.push(...existsErrors);
    }
  }

  // Extension validation
  if (allowedExtensions && allowedExtensions.length > 0) {
    const ext = path.extname(normalizedPath).toLowerCase();
    if (!allowedExtensions.includes(ext)) {
      errors.push(
        `File extension not allowed. Allowed: ${allowedExtensions.join(', ')}`
      );
    }
  }

  return {
    isValid: errors.length === 0,
    errors: Object.freeze(errors),
    sanitizedValue: normalizedPath,
  };
}

/**
 * Check for dangerous path patterns
 */
function containsDangerousPathPatterns(inputPath: string): boolean {
  return DANGEROUS_PATH_PATTERNS.some(pattern => pattern.test(inputPath));
}

/**
 * Check if path is safe (within allowed directory)
 */
function isPathSafe(targetPath: string, basePath: string): boolean {
  try {
    const normalizedTarget = path.resolve(targetPath);
    const normalizedBase = path.resolve(basePath);

    return normalizedTarget.startsWith(normalizedBase);
  } catch {
    return false;
  }
}

/**
 * Perform file system existence checks
 */
async function performExistenceChecks(
  normalizedPath: string,
  checks: {
    mustExist: boolean;
    mustBeDirectory: boolean;
    mustBeFile: boolean;
  }
): Promise<{ existsChecks: boolean; existsErrors: string[] }> {
  const { mustExist, mustBeDirectory, mustBeFile } = checks;
  const existsErrors: string[] = [];

  try {
    const { promises: fs } = await import('node:fs');
    const stats = await fs.stat(normalizedPath);

    if (mustBeDirectory && !stats.isDirectory()) {
      existsErrors.push('Path must be a directory');
    }

    if (mustBeFile && !stats.isFile()) {
      existsErrors.push('Path must be a file');
    }
  } catch (_error) {
    if (mustExist) {
      existsErrors.push('Path does not exist');
    }
  }

  return {
    existsChecks: existsErrors.length === 0,
    existsErrors,
  };
}

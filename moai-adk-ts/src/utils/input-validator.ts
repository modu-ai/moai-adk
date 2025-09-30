// @CODE:UTIL-005 | 
// Related: @CODE:VALID-002:API

/**
 * @file Input validation utilities
 * @author MoAI Team
 */

import * as path from 'node:path';

/**
 * Validation result interface
 * @tags @CODE:VALIDATION-RESULT-001:DATA
 */
export interface ValidationResult {
  readonly isValid: boolean;
  readonly errors: readonly string[];
  readonly sanitizedValue?: string;
}

/**
 * Project name validation options
 * @tags @CODE:PROJECT-NAME-OPTIONS-001:DATA
 */
export interface ProjectNameOptions {
  readonly minLength?: number;
  readonly maxLength?: number;
  readonly allowSpaces?: boolean;
  readonly allowSpecialChars?: boolean;
}

/**
 * Path validation options
 * @tags @CODE:PATH-VALIDATION-OPTIONS-001:DATA
 */
export interface PathValidationOptions {
  readonly mustExist?: boolean;
  readonly mustBeDirectory?: boolean;
  readonly mustBeFile?: boolean;
  readonly allowedExtensions?: readonly string[];
  readonly maxDepth?: number;
}

/**
 * Input validator class for CLI parameters and user inputs
 * @tags @SECURITY:INPUT-VALIDATION-001
 */
export class InputValidator {
  /**
   * Validate project name
   * @tags @CODE:VALIDATE-PROJECT-NAME-001:API
   */
  static validateProjectName(
    projectName: string,
    options: ProjectNameOptions = {}
  ): ValidationResult {
    const {
      minLength = 1,
      maxLength = 50,
      allowSpaces = false,
      allowSpecialChars = false,
    } = options;

    const errors: string[] = [];

    // Basic checks
    if (!projectName || typeof projectName !== 'string') {
      errors.push('Project name must be a non-empty string');
      return { isValid: false, errors };
    }

    const trimmed = projectName.trim();

    // Length validation
    if (trimmed.length < minLength) {
      errors.push(`Project name must be at least ${minLength} characters long`);
    }

    if (trimmed.length > maxLength) {
      errors.push(`Project name must not exceed ${maxLength} characters`);
    }

    // Character validation
    if (!allowSpaces && /\s/.test(trimmed)) {
      errors.push('Project name cannot contain spaces');
    }

    // Special character validation
    const specialChars = /[!@#$%^&*(),.?":{}|<>]/;
    if (!allowSpecialChars && specialChars.test(trimmed)) {
      errors.push('Project name cannot contain special characters');
    }

    // Dangerous pattern detection
    const dangerousPatterns = [
      /\.\./, // Path traversal
      /^[.-]/, // Starting with dot or dash
      /[<>:"|?*]/, // File system reserved chars
      /[\x00-\x1f]/, // Control characters
      /^(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])$/i, // Windows reserved names
    ];

    for (const pattern of dangerousPatterns) {
      if (pattern.test(trimmed)) {
        errors.push('Project name contains invalid characters or patterns');
        break;
      }
    }

    // Sanitize the value
    let sanitized = trimmed;
    if (!allowSpecialChars) {
      sanitized = sanitized.replace(/[^a-zA-Z0-9_-]/g, '-');
    }

    return {
      isValid: errors.length === 0,
      errors: Object.freeze(errors),
      sanitizedValue: sanitized,
    };
  }

  /**
   * Validate file/directory path
   * @tags @CODE:VALIDATE-PATH-001:API
   */
  static async validatePath(
    inputPath: string,
    options: PathValidationOptions = {}
  ): Promise<ValidationResult> {
    const {
      mustExist = false,
      mustBeDirectory = false,
      mustBeFile = false,
      allowedExtensions,
      maxDepth = 10,
    } = options;

    const errors: string[] = [];

    // Basic checks
    if (!inputPath || typeof inputPath !== 'string') {
      errors.push('Path must be a non-empty string');
      return { isValid: false, errors };
    }

    const trimmed = inputPath.trim();

    // Path length validation
    if (trimmed.length > 260) {
      // Windows MAX_PATH limitation
      errors.push('Path is too long (maximum 260 characters)');
    }

    // Dangerous pattern detection
    if (InputValidator.containsDangerousPathPatterns(trimmed)) {
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
    if (!InputValidator.isPathSafe(normalizedPath, process.cwd())) {
      errors.push('Path traversal detected');
    }

    // File system checks (if required)
    if (mustExist || mustBeDirectory || mustBeFile) {
      const { existsChecks, existsErrors } =
        await InputValidator.performExistenceChecks(normalizedPath, {
          mustExist,
          mustBeDirectory,
          mustBeFile,
        });

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
   * Validate template type
   * @tags @CODE:VALIDATE-TEMPLATE-TYPE-001:API
   */
  static validateTemplateType(templateType: string): ValidationResult {
    const errors: string[] = [];
    const allowedTypes = ['standard', 'minimal', 'advanced', 'custom'];

    if (!templateType || typeof templateType !== 'string') {
      errors.push('Template type must be a non-empty string');
      return { isValid: false, errors };
    }

    const trimmed = templateType.trim().toLowerCase();

    if (!allowedTypes.includes(trimmed)) {
      errors.push(`Invalid template type. Allowed: ${allowedTypes.join(', ')}`);
    }

    return {
      isValid: errors.length === 0,
      errors: Object.freeze(errors),
      sanitizedValue: trimmed,
    };
  }

  /**
   * Validate Git branch name
   * @tags @CODE:VALIDATE-BRANCH-NAME-001:API
   */
  static validateBranchName(branchName: string): ValidationResult {
    const errors: string[] = [];

    if (!branchName || typeof branchName !== 'string') {
      errors.push('Branch name must be a non-empty string');
      return { isValid: false, errors };
    }

    const trimmed = branchName.trim();

    // Length validation
    if (trimmed.length < 1 || trimmed.length > 250) {
      errors.push('Branch name must be between 1 and 250 characters');
    }

    // Git branch naming rules
    const invalidPatterns = [
      /\.\./, // Double dots
      /^[.-]/, // Starting with dot or dash
      /[.-]$/, // Ending with dot or dash
      /[\x00-\x1f\x7f]/, // Control characters
      /[~^:?*[\]\\]/, // Git-reserved characters
      /\s/, // Whitespace
      /@{/, // @{ sequence
      /\/$/, // Ending with slash
      /\/\//, // Double slashes
    ];

    for (const pattern of invalidPatterns) {
      if (pattern.test(trimmed)) {
        errors.push('Branch name contains invalid characters or patterns');
        break;
      }
    }

    // Git reserved names
    const reservedNames = ['HEAD', 'master', 'origin'];
    if (reservedNames.includes(trimmed)) {
      errors.push('Branch name is reserved');
    }

    return {
      isValid: errors.length === 0,
      errors: Object.freeze(errors),
      sanitizedValue: trimmed,
    };
  }

  /**
   * Validate command options
   * @tags @CODE:VALIDATE-COMMAND-OPTIONS-001:API
   */
  static validateCommandOptions(
    options: Record<string, any>
  ): ValidationResult {
    const errors: string[] = [];
    const sanitizedOptions: Record<string, any> = {};

    for (const [key, value] of Object.entries(options)) {
      // Validate option key
      if (!/^[a-zA-Z][a-zA-Z0-9_-]*$/.test(key)) {
        errors.push(`Invalid option key: ${key}`);
        continue;
      }

      // Validate option value based on type
      if (typeof value === 'string') {
        if (value.length > 1000) {
          errors.push(`Option ${key} value is too long`);
          continue;
        }

        if (InputValidator.containsDangerousString(value)) {
          errors.push(`Option ${key} contains dangerous content`);
          continue;
        }

        sanitizedOptions[key] = value.trim();
      } else if (typeof value === 'boolean') {
        sanitizedOptions[key] = value;
      } else if (typeof value === 'number') {
        if (!Number.isFinite(value)) {
          errors.push(`Option ${key} must be a finite number`);
          continue;
        }
        sanitizedOptions[key] = value;
      } else {
        errors.push(`Option ${key} has unsupported type`);
      }
    }

    return {
      isValid: errors.length === 0,
      errors: Object.freeze(errors),
      sanitizedValue: sanitizedOptions,
    };
  }

  /**
   * Check for dangerous path patterns
   * @tags @UTIL:CHECK-DANGEROUS-PATTERNS-001
   */
  private static containsDangerousPathPatterns(inputPath: string): boolean {
    const dangerousPatterns = [
      /\.\./, // Path traversal
      /[\x00-\x1f]/, // Control characters
      /[<>:"|?*]/, // Windows reserved chars
      /\/etc\//, // Unix system directories
      /\/bin\//,
      /\/usr\/bin\//,
      /\/var\/log\//,
      /C:\\Windows\\/i, // Windows system directories
      /C:\\Program Files\\/i,
      /\.(exe|bat|cmd|scr|pif|com|vbs)$/i, // Executable extensions
    ];

    return dangerousPatterns.some(pattern => pattern.test(inputPath));
  }

  /**
   * Check if path is safe (within allowed directory)
   * @tags @UTIL:CHECK-PATH-SAFETY-001
   */
  private static isPathSafe(targetPath: string, basePath: string): boolean {
    try {
      const normalizedTarget = path.resolve(targetPath);
      const normalizedBase = path.resolve(basePath);

      return normalizedTarget.startsWith(normalizedBase);
    } catch {
      return false;
    }
  }

  /**
   * Check for dangerous string content
   * @tags @UTIL:CHECK-DANGEROUS-STRING-001
   */
  private static containsDangerousString(value: string): boolean {
    const dangerousPatterns = [
      /javascript:/i,
      /eval\s*\(/i,
      /function\s*\(/i,
      /setTimeout\s*\(/i,
      /setInterval\s*\(/i,
      /<script/i,
      /on\w+\s*=/i, // Event handlers
      /\${.*}/, // Template expressions
      /`.*`/, // Template literals
    ];

    return dangerousPatterns.some(pattern => pattern.test(value));
  }

  /**
   * Perform file system existence checks
   * @tags @UTIL:PERFORM-EXISTENCE-CHECKS-001
   */
  private static async performExistenceChecks(
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
}

/**
 * Helper function for quick project name validation
 * @tags @CODE:QUICK-VALIDATE-PROJECT-NAME-001:API
 */
export function validateProjectName(name: string): ValidationResult {
  return InputValidator.validateProjectName(name);
}

/**
 * Helper function for quick path validation
 * @tags @CODE:QUICK-VALIDATE-PATH-001:API
 */
export async function validatePath(
  inputPath: string
): Promise<ValidationResult> {
  return InputValidator.validatePath(inputPath);
}

/**
 * Helper function for quick branch name validation
 * @tags @CODE:QUICK-VALIDATE-BRANCH-NAME-001:API
 */
export function validateBranchName(branchName: string): ValidationResult {
  return InputValidator.validateBranchName(branchName);
}

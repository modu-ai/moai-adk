/**
 * @FEATURE:INSTALLATION-VALIDATOR-001 Installation Verification and Validation System
 *
 * @TASK:VALIDATION-001 Handles comprehensive validation of installation results
 * and verification of all components for proper MoAI-ADK operation.
 */

import { promises as fs } from 'node:fs';
import path from 'node:path';

/**
 * @TASK:VALIDATOR-MAIN-001 Specialized installation validation component
 *
 * Handles comprehensive validation of installation results including
 * resource verification, configuration validation, and system integrity checks.
 *
 * Responsibilities:
 * - Installation completeness verification
 * - Resource integrity validation
 * - Configuration file validation
 * - System integration checks
 * - Error detection and reporting
 */
export class InstallationValidator {
  private readonly requiredDirectories: readonly string[] = [
    '.claude',
    '.claude/agents',
    '.claude/commands',
    '.claude/hooks',
    '.moai',
    '.moai/project',
    '.moai/specs',
    '.moai/memory',
  ] as const;

  private readonly requiredFiles: readonly string[] = ['CLAUDE.md'] as const;

  private readonly configFiles: readonly { path: string; required: boolean }[] =
    [
      { path: '.claude/settings.json', required: false },
      { path: '.moai/config.json', required: false },
    ] as const;

  /**
   * @TASK:VALIDATE-STRUCTURE-001 Validate basic project directory structure
   *
   * Checks for presence of required directories and their proper structure.
   *
   * @param projectPath - Path to the project root
   * @returns Validation result with any errors found
   */
  async validateProjectStructure(
    projectPath: string
  ): Promise<ValidationResult> {
    const errors: ValidationError[] = [];
    const warnings: ValidationWarning[] = [];

    try {
      // Validate all directories concurrently for better performance
      const validationPromises = this.requiredDirectories.map(dirPath =>
        this.validateDirectory(projectPath, dirPath)
      );

      const results = await Promise.allSettled(validationPromises);

      for (const result of results) {
        if (result.status === 'fulfilled' && result.value) {
          errors.push(result.value);
        }
      }

      return this.createValidationResult(errors, warnings);
    } catch (error: any) {
      errors.push(
        this.createError(
          ValidationErrorCodes.MISSING_DIRECTORY,
          `Project path validation failed: ${error.message}`,
          projectPath
        )
      );

      return this.createValidationResult(errors, warnings);
    }
  }

  /**
   * Helper method to validate a single directory
   * @param projectPath - Root project path
   * @param dirPath - Relative directory path to validate
   * @returns ValidationError if directory is invalid, null if valid
   */
  private async validateDirectory(
    projectPath: string,
    dirPath: string
  ): Promise<ValidationError | null> {
    const fullPath = path.join(projectPath, dirPath);

    try {
      const stats = await fs.stat(fullPath);

      if (!stats.isDirectory()) {
        return this.createError(
          ValidationErrorCodes.INVALID_STRUCTURE,
          `Path exists but is not a directory: ${dirPath}`,
          fullPath
        );
      }

      return null;
    } catch (error: any) {
      if (error.code === 'ENOENT') {
        return this.createError(
          ValidationErrorCodes.MISSING_DIRECTORY,
          `Required directory missing: ${dirPath}`,
          fullPath
        );
      } else if (error.code === 'EACCES') {
        return this.createError(
          ValidationErrorCodes.PERMISSION_DENIED,
          `Permission denied accessing: ${dirPath}`,
          fullPath
        );
      } else {
        return this.createError(
          ValidationErrorCodes.INVALID_STRUCTURE,
          `Error accessing directory ${dirPath}: ${error.message}`,
          fullPath
        );
      }
    }
  }

  /**
   * @TASK:VALIDATE-CONFIG-001 Validate configuration files integrity
   *
   * Checks that JSON configuration files are valid and readable.
   *
   * @param projectPath - Path to the project root
   * @returns Validation result with any configuration errors
   */
  async validateConfigurationFiles(
    projectPath: string
  ): Promise<ValidationResult> {
    const errors: ValidationError[] = [];
    const warnings: ValidationWarning[] = [];

    try {
      // Validate all config files concurrently
      const validationPromises = this.configFiles.map(config =>
        this.validateConfigFile(projectPath, config)
      );

      const results = await Promise.allSettled(validationPromises);

      for (const result of results) {
        if (result.status === 'fulfilled' && result.value) {
          errors.push(result.value);
        }
      }

      return this.createValidationResult(errors, warnings);
    } catch (error: any) {
      errors.push(
        this.createError(
          ValidationErrorCodes.INVALID_JSON,
          `Configuration validation failed: ${error.message}`,
          projectPath
        )
      );

      return this.createValidationResult(errors, warnings);
    }
  }

  /**
   * Helper method to validate a single configuration file
   * @param projectPath - Root project path
   * @param config - Configuration file metadata
   * @returns ValidationError if file is invalid, null if valid
   */
  private async validateConfigFile(
    projectPath: string,
    config: { path: string; required: boolean }
  ): Promise<ValidationError | null> {
    const fullPath = path.join(projectPath, config.path);

    try {
      await fs.access(fullPath);

      // File exists, validate JSON
      const content = await fs.readFile(fullPath, 'utf-8');

      try {
        JSON.parse(content);
        return null;
      } catch (jsonError: any) {
        return this.createError(
          ValidationErrorCodes.INVALID_JSON,
          `Invalid JSON in ${config.path}: ${jsonError.message}`,
          fullPath
        );
      }
    } catch (error: any) {
      if (error.code === 'ENOENT') {
        // File doesn't exist - that's OK if not required
        if (config.required) {
          return this.createError(
            ValidationErrorCodes.MISSING_FILE,
            `Required configuration file missing: ${config.path}`,
            fullPath
          );
        }
        return null;
      } else {
        return this.createError(
          ValidationErrorCodes.PERMISSION_DENIED,
          `Cannot access configuration file ${config.path}: ${error.message}`,
          fullPath
        );
      }
    }
  }

  /**
   * @TASK:VALIDATE-RESOURCES-001 Validate resource files presence and integrity
   *
   * Checks for required resource files and their accessibility.
   *
   * @param projectPath - Path to the project root
   * @returns Validation result with resource file errors
   */
  async validateResourceFiles(projectPath: string): Promise<ValidationResult> {
    const errors: ValidationError[] = [];
    const warnings: ValidationWarning[] = [];

    try {
      // Validate all required files concurrently
      const validationPromises = this.requiredFiles.map(fileName =>
        this.validateResourceFile(projectPath, fileName, warnings)
      );

      const results = await Promise.allSettled(validationPromises);

      for (const result of results) {
        if (result.status === 'fulfilled' && result.value) {
          errors.push(result.value);
        }
      }

      return this.createValidationResult(errors, warnings);
    } catch (error: any) {
      errors.push(
        this.createError(
          ValidationErrorCodes.MISSING_FILE,
          `Resource validation failed: ${error.message}`,
          projectPath
        )
      );

      return this.createValidationResult(errors, warnings);
    }
  }

  /**
   * Helper method to validate a single resource file
   * @param projectPath - Root project path
   * @param fileName - Resource file name to validate
   * @param warnings - Array to collect warnings
   * @returns ValidationError if file is invalid, null if valid
   */
  private async validateResourceFile(
    projectPath: string,
    fileName: string,
    warnings: ValidationWarning[]
  ): Promise<ValidationError | null> {
    const fullPath = path.join(projectPath, fileName);

    try {
      const stats = await fs.stat(fullPath);

      if (!stats.isFile()) {
        return this.createError(
          ValidationErrorCodes.MISSING_FILE,
          `Required file is not a file: ${fileName}`,
          fullPath
        );
      }

      // Check file permissions on Unix systems (async to avoid blocking)
      if (process.platform !== 'win32') {
        this.checkFilePermissions(fullPath, fileName, warnings).catch(() => {
          // Ignore permission check errors - they're non-critical
        });
      }

      return null;
    } catch (error: any) {
      if (error.code === 'ENOENT') {
        return this.createError(
          ValidationErrorCodes.MISSING_FILE,
          `Required file missing: ${fileName}`,
          fullPath
        );
      } else {
        return this.createError(
          ValidationErrorCodes.PERMISSION_DENIED,
          `Cannot access file ${fileName}: ${error.message}`,
          fullPath
        );
      }
    }
  }

  /**
   * Helper method to check file permissions asynchronously
   * @param fullPath - Full path to file
   * @param fileName - File name for error messages
   * @param warnings - Array to collect warnings
   */
  private async checkFilePermissions(
    fullPath: string,
    fileName: string,
    warnings: ValidationWarning[]
  ): Promise<void> {
    try {
      await fs.access(fullPath, fs.constants.R_OK);
    } catch (_permError) {
      warnings.push(
        this.createWarning(
          ValidationErrorCodes.PERMISSION_DENIED,
          `File permission issues detected: ${fileName}`,
          fullPath,
          'Check file permissions'
        )
      );
    }
  }

  /**
   * @TASK:VALIDATE-CLAUDE-001 Validate Claude Code specific integration
   *
   * Checks Claude Code agents, commands, and hooks configuration.
   *
   * @param projectPath - Path to the project root
   * @returns Validation result for Claude integration
   */
  async validateClaudeIntegration(
    projectPath: string
  ): Promise<ValidationResult> {
    const errors: ValidationError[] = [];
    const warnings: ValidationWarning[] = [];

    try {
      const claudeDir = path.join(projectPath, '.claude');

      // Check settings.json if it exists
      const settingsPath = path.join(claudeDir, 'settings.json');

      try {
        await fs.access(settingsPath);

        const content = await fs.readFile(settingsPath, 'utf-8');
        const settings = JSON.parse(content);

        // Basic validation of settings structure
        if (typeof settings !== 'object' || settings === null) {
          errors.push({
            code: ValidationErrorCodes.INVALID_JSON,
            message: 'Claude settings.json must contain an object',
            path: settingsPath,
            severity: 'error',
          });
        }

        // Check for some expected fields (minimal validation)
        if (Object.keys(settings).length === 0) {
          errors.push({
            code: ValidationErrorCodes.INTEGRATION_FAILURE,
            message:
              'Claude settings.json appears to be empty - required field missing',
            path: settingsPath,
            severity: 'error',
          });
        }
      } catch (error: any) {
        if (error.code !== 'ENOENT') {
          errors.push({
            code: ValidationErrorCodes.INVALID_JSON,
            message: `Error reading Claude settings.json: ${error.message}`,
            path: settingsPath,
            severity: 'error',
          });
        }
      }

      // Check for corrupted agent files
      const agentsDir = path.join(claudeDir, 'agents');

      try {
        await fs.access(agentsDir);
        const files = await fs.readdir(agentsDir);

        for (const file of files) {
          if (file.endsWith('.md')) {
            const filePath = path.join(agentsDir, file);

            try {
              const content = await fs.readFile(filePath, 'utf-8');

              // Simple check for binary data in text files
              if (
                content.includes('\0') ||
                /[\x00-\x08\x0E-\x1F\x7F]/.test(content)
              ) {
                warnings.push({
                  code: ValidationErrorCodes.CORRUPTED_FILE,
                  message: `Agent file may be corrupted: ${file}`,
                  path: filePath,
                  suggestion: 'Check file encoding and content',
                });
              }
            } catch (_readError) {
              warnings.push({
                code: ValidationErrorCodes.CORRUPTED_FILE,
                message: `Cannot read agent file: ${file}`,
                path: filePath,
                suggestion: 'Check file permissions and encoding',
              });
            }
          }
        }
      } catch (_error: any) {
        // Agents directory doesn't exist or can't be read - that's handled in structure validation
      }

      return {
        valid: errors.length === 0,
        errors: Object.freeze(errors),
        warnings: Object.freeze(warnings),
      };
    } catch (error: any) {
      errors.push({
        code: ValidationErrorCodes.INTEGRATION_FAILURE,
        message: `Claude integration validation failed: ${error.message}`,
        path: projectPath,
        severity: 'error',
      });

      return {
        valid: false,
        errors: Object.freeze(errors),
        warnings: Object.freeze(warnings),
      };
    }
  }

  /**
   * @TASK:VALIDATE-MOAI-001 Validate MoAI system specific integration
   *
   * Checks MoAI directories, memory files, and index integrity.
   *
   * @param projectPath - Path to the project root
   * @returns Validation result for MoAI integration
   */
  async validateMoaiIntegration(
    projectPath: string
  ): Promise<ValidationResult> {
    const errors: ValidationError[] = [];
    const warnings: ValidationWarning[] = [];

    try {
      const moaiDir = path.join(projectPath, '.moai');
      const requiredDirs = ['project', 'specs', 'memory', 'indexes'];

      // Check required directories
      for (const dirName of requiredDirs) {
        const dirPath = path.join(moaiDir, dirName);

        try {
          const stats = await fs.stat(dirPath);

          if (!stats.isDirectory()) {
            errors.push({
              code: ValidationErrorCodes.MISSING_DIRECTORY,
              message: `MoAI directory is not a directory: ${dirName}`,
              path: dirPath,
              severity: 'error',
            });
          }
        } catch (error: any) {
          if (error.code === 'ENOENT') {
            errors.push({
              code: ValidationErrorCodes.MISSING_DIRECTORY,
              message: `Required MoAI directory missing: ${dirName}`,
              path: dirPath,
              severity: 'error',
            });
          }
        }
      }

      // Check development guide if memory directory exists
      const memoryDir = path.join(moaiDir, 'memory');
      try {
        await fs.access(memoryDir);
        const guidePath = path.join(memoryDir, 'development-guide.md');

        try {
          await fs.access(guidePath);
        } catch (_error) {
          // Development guide is optional
        }
      } catch (_error) {
        // Memory directory doesn't exist - handled above
      }

      // Validate index files
      const indexesDir = path.join(moaiDir, 'indexes');
      try {
        await fs.access(indexesDir);
        const files = await fs.readdir(indexesDir);

        for (const file of files) {
          const filePath = path.join(indexesDir, file);

          if (file.endsWith('.json')) {
            try {
              const content = await fs.readFile(filePath, 'utf-8');
              JSON.parse(content);
            } catch (jsonError: any) {
              errors.push({
                code: ValidationErrorCodes.INVALID_JSON,
                message: `Invalid JSON in index file ${file}: ${jsonError.message}`,
                path: filePath,
                severity: 'error',
              });
            }
          } else if (file.endsWith('.db')) {
            // SQLite3 database validation
            try {
              const content = await fs.readFile(filePath);

              // Check SQLite3 magic number (first 16 bytes should start with "SQLite format 3\0")
              const sqliteHeader = 'SQLite format 3\0';
              const fileHeader = content.slice(0, 16).toString('utf8');

              if (!fileHeader.startsWith(sqliteHeader)) {
                errors.push({
                  code: ValidationErrorCodes.INVALID_SQLITE,
                  message: `Invalid SQLite3 database file ${file}: Invalid header`,
                  path: filePath,
                  severity: 'error',
                });
              }
            } catch (sqliteError: any) {
              errors.push({
                code: ValidationErrorCodes.INVALID_SQLITE,
                message: `Failed to validate SQLite3 database ${file}: ${sqliteError.message}`,
                path: filePath,
                severity: 'error',
              });
            }
          }
        }
      } catch (_error: any) {
        // Indexes directory doesn't exist - handled above
      }

      return {
        valid: errors.length === 0,
        errors: Object.freeze(errors),
        warnings: Object.freeze(warnings),
      };
    } catch (error: any) {
      errors.push({
        code: ValidationErrorCodes.INTEGRATION_FAILURE,
        message: `MoAI integration validation failed: ${error.message}`,
        path: projectPath,
        severity: 'error',
      });

      return {
        valid: false,
        errors: Object.freeze(errors),
        warnings: Object.freeze(warnings),
      };
    }
  }

  /**
   * @TASK:DIAGNOSTIC-REPORT-001 Generate comprehensive diagnostic report
   *
   * Runs all validation methods and compiles a detailed report.
   *
   * @param projectPath - Path to the project root
   * @returns Complete diagnostic report with recommendations
   */
  async generateDiagnosticReport(
    projectPath: string
  ): Promise<DiagnosticReport> {
    const timestamp = new Date();
    const validationResults: ValidationResult[] = [];

    // Run all validation methods
    const structureResult = await this.validateProjectStructure(projectPath);
    validationResults.push(structureResult);

    const configResult = await this.validateConfigurationFiles(projectPath);
    validationResults.push(configResult);

    const resourceResult = await this.validateResourceFiles(projectPath);
    validationResults.push(resourceResult);

    const claudeResult = await this.validateClaudeIntegration(projectPath);
    validationResults.push(claudeResult);

    const moaiResult = await this.validateMoaiIntegration(projectPath);
    validationResults.push(moaiResult);

    // Determine overall health
    const hasErrors = validationResults.some(result => !result.valid);
    const hasWarnings = validationResults.some(
      result => result.warnings.length > 0
    );

    let overallHealth: 'healthy' | 'warning' | 'error';
    if (hasErrors) {
      overallHealth = 'error';
    } else if (hasWarnings) {
      overallHealth = 'warning';
    } else {
      overallHealth = 'healthy';
    }

    // Generate recommendations
    const recommendations: string[] = [];
    const nextSteps: string[] = [];

    if (hasErrors) {
      recommendations.push('Fix critical errors before proceeding');
      nextSteps.push(
        'Review error messages and fix missing or corrupted files'
      );
    }

    if (hasWarnings) {
      recommendations.push('Address warnings to improve system reliability');
      nextSteps.push('Check file permissions and resolve any access issues');
    }

    if (overallHealth === 'healthy') {
      recommendations.push('Installation appears to be complete and healthy');
      nextSteps.push('Proceed with normal development workflow');
    }

    return {
      projectPath,
      timestamp,
      overallHealth,
      validationResults: Object.freeze(validationResults),
      recommendations: Object.freeze(recommendations),
      nextSteps: Object.freeze(nextSteps),
    };
  }

  /**
   * Helper method to create validation errors with consistent structure
   * @param code - Error code
   * @param message - Error message
   * @param path - Optional file path
   * @returns Formatted ValidationError
   */
  private createError(
    code: ValidationErrorCode,
    message: string,
    path?: string
  ): ValidationError {
    const error: ValidationError = {
      code,
      message,
      severity: 'error' as const,
    };

    if (path !== undefined) {
      (error as any).path = path;
    }

    return error;
  }

  /**
   * Helper method to create validation warnings with consistent structure
   * @param code - Warning code
   * @param message - Warning message
   * @param path - Optional file path
   * @param suggestion - Optional suggestion for fixing the issue
   * @returns Formatted ValidationWarning
   */
  private createWarning(
    code: ValidationErrorCode,
    message: string,
    path?: string,
    suggestion?: string
  ): ValidationWarning {
    const warning: ValidationWarning = {
      code,
      message,
    };

    if (path !== undefined) {
      (warning as any).path = path;
    }

    if (suggestion !== undefined) {
      (warning as any).suggestion = suggestion;
    }

    return warning;
  }

  /**
   * Helper method to create validation results with frozen arrays
   * @param errors - Array of validation errors
   * @param warnings - Array of validation warnings
   * @returns Formatted ValidationResult
   */
  private createValidationResult(
    errors: ValidationError[],
    warnings: ValidationWarning[]
  ): ValidationResult {
    return {
      valid: errors.length === 0,
      errors: Object.freeze(errors),
      warnings: Object.freeze(warnings),
    };
  }
}

// Error codes for consistent error identification
export const ValidationErrorCodes = {
  MISSING_DIRECTORY: 'MISSING_DIRECTORY',
  MISSING_FILE: 'MISSING_FILE',
  INVALID_JSON: 'INVALID_JSON',
  INVALID_SQLITE: 'INVALID_SQLITE',
  PERMISSION_DENIED: 'PERMISSION_DENIED',
  INVALID_STRUCTURE: 'INVALID_STRUCTURE',
  CORRUPTED_FILE: 'CORRUPTED_FILE',
  INTEGRATION_FAILURE: 'INTEGRATION_FAILURE',
} as const;

export type ValidationErrorCode =
  (typeof ValidationErrorCodes)[keyof typeof ValidationErrorCodes];

// Interface definitions
export interface ValidationResult {
  readonly valid: boolean;
  readonly errors: readonly ValidationError[];
  readonly warnings: readonly ValidationWarning[];
}

export interface ValidationError {
  readonly code: ValidationErrorCode;
  readonly message: string;
  readonly path?: string;
  readonly severity: 'error' | 'warning';
}

export interface ValidationWarning {
  readonly code: ValidationErrorCode;
  readonly message: string;
  readonly path?: string;
  readonly suggestion?: string;
}

export interface DiagnosticReport {
  readonly projectPath: string;
  readonly timestamp: Date;
  readonly overallHealth: 'healthy' | 'warning' | 'error';
  readonly validationResults: readonly ValidationResult[];
  readonly recommendations: readonly string[];
  readonly nextSteps: readonly string[];
}

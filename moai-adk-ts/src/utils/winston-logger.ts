// @CODE:LOG-002 |
// Related: @CODE:LOG-002:API, @CODE:LOG-CFG-001

/**
 * @file Winston-based structured logging system
 * @author MoAI Team
 */

import { existsSync, mkdirSync } from 'node:fs';
import { join } from 'node:path';
import winston from 'winston';
import { mergeLoggerOptions } from '../config/logger-config.js';
import type { LoggerOptions } from '../types/logger.js';

/**
 * Winston-based MoAI Logger with structured logging and sensitive data masking
 * Supports verbose mode for detailed debugging output
 */
export class MoaiLogger {
  private logger: winston.Logger;
  private readonly options: Required<Omit<LoggerOptions, 'transports'>> & {
    transports?: winston.transport[];
  };
  private verboseMode: boolean = false;

  /**
   * Sensitive field patterns to mask (case-insensitive)
   */
  private readonly sensitiveFields = [
    'password',
    'passwd',
    'pwd',
    'token',
    'apiKey',
    'api_key',
    'secret',
    'accessToken',
    'access_token',
    'refreshToken',
    'refresh_token',
    'apiSecret',
    'api_secret',
    'privateKey',
    'private_key',
    'clientSecret',
    'client_secret',
    'authorization',
    'auth',
  ];

  /**
   * Regex patterns for sensitive data in strings (enhanced)
   */
  private readonly sensitivePatterns = [
    /password=\S+/gi,
    /passwd=\S+/gi,
    /pwd=\S+/gi,
    /token=\S+/gi,
    /api[_-]?key=\S+/gi,
    /secret=\S+/gi,
    /bearer\s+\S+/gi,
    /authorization:\s*\S+/gi,
    /auth:\s*\S+/gi,
    /access[_-]?token=\S+/gi,
    /refresh[_-]?token=\S+/gi,
  ];

  constructor(options?: LoggerOptions) {
    this.options = mergeLoggerOptions(options);
    this.logger = this.createWinstonLogger();
  }

  /**
   * Create Winston logger with configured transports
   */
  private createWinstonLogger(): winston.Logger {
    // DI: Use custom transports if provided
    const transports =
      this.options.transports ?? this.createDefaultTransports();

    return winston.createLogger({
      level: this.options.level,
      format: winston.format.combine(
        winston.format.timestamp(),
        winston.format.errors({ stack: true }),
        winston.format.json(),
        this.createMaskingFormat()
      ),
      transports,
    });
  }

  /**
   * Create default transports based on options
   */
  private createDefaultTransports(): winston.transport[] {
    const transports: winston.transport[] = [];

    // Console transport with clean format
    if (this.options.enableConsole) {
      transports.push(
        new winston.transports.Console({
          format: winston.format.combine(
            winston.format.colorize(),
            winston.format.printf(({ level, message }) => {
              // Clean format without timestamp for better UX
              return `${level}: ${message}`;
            })
          ),
        })
      );
    }

    // File transports
    if (this.options.enableFile) {
      try {
        this.ensureLogDirectory();

        transports.push(
          new winston.transports.File({
            filename: join(this.options.logDir, 'error.log'),
            level: 'error',
          })
        );

        transports.push(
          new winston.transports.File({
            filename: join(this.options.logDir, 'combined.log'),
          })
        );
      } catch (error) {
        // Fallback to console only if file logging fails
        // Use process.stderr for initialization errors (logger not yet ready)
        const errorMessage =
          error instanceof Error ? error.message : String(error);
        process.stderr.write(
          `[WINSTON-INIT-WARN] Failed to initialize file logging, using console only: ${errorMessage}\n`
        );
      }
    }

    return transports;
  }

  /**
   * Ensure log directory exists
   */
  private ensureLogDirectory(): void {
    if (!existsSync(this.options.logDir)) {
      mkdirSync(this.options.logDir, { recursive: true });
    }
  }

  /**
   * Create Winston format for masking sensitive data
   */
  private createMaskingFormat(): winston.Logform.Format {
    return winston.format(info => {
      const masked = { ...info };

      // Mask message string
      if (typeof masked.message === 'string') {
        masked.message = this.maskSensitiveString(masked.message);
      }

      // Mask metadata fields
      for (const key of Object.keys(masked)) {
        if (this.isSensitiveField(key)) {
          masked[key] = '***REDACTED***';
        } else if (
          typeof masked[key] === 'object' &&
          masked[key] !== null &&
          !Array.isArray(masked[key])
        ) {
          masked[key] = this.maskSensitiveContext(
            masked[key] as Record<string, unknown>
          );
        }
      }

      return masked;
    })();
  }

  /**
   * Check if field name is sensitive
   */
  private isSensitiveField(fieldName: string): boolean {
    const lowerField = fieldName.toLowerCase();
    return this.sensitiveFields.some(sensitive =>
      lowerField.includes(sensitive)
    );
  }

  /**
   * Mask sensitive patterns in string
   */
  private maskSensitiveString(message: string): string {
    let masked = message;
    for (const pattern of this.sensitivePatterns) {
      masked = masked.replace(pattern, match => {
        const [key] = match.split('=');
        return `${key}=***redacted***`;
      });
    }
    return masked;
  }

  /**
   * Mask sensitive fields in context object
   */
  private maskSensitiveContext(
    context: Record<string, unknown>
  ): Record<string, unknown> {
    const masked: Record<string, unknown> = {};

    for (const [key, value] of Object.entries(context)) {
      if (this.isSensitiveField(key)) {
        masked[key] = '***REDACTED***';
      } else if (
        typeof value === 'object' &&
        value !== null &&
        !Array.isArray(value)
      ) {
        masked[key] = this.maskSensitiveContext(
          value as Record<string, unknown>
        );
      } else {
        masked[key] = value;
      }
    }

    return masked;
  }

  /**
   * Log debug message
   */
  debug(message: string, meta?: Record<string, unknown>): void {
    this.logger.debug(message, meta);
  }

  /**
   * Log info message
   */
  info(message: string, meta?: Record<string, unknown>): void {
    this.logger.info(message, meta);
  }

  /**
   * Log warning message
   */
  warn(message: string, meta?: Record<string, unknown>): void {
    this.logger.warn(message, meta);
  }

  /**
   * Log error message with optional Error object
   */
  error(
    message: string,
    error?: Error | unknown,
    meta?: Record<string, unknown>
  ): void {
    const errorMeta: Record<string, unknown> = { ...meta };

    if (error instanceof Error) {
      errorMeta.error = {
        name: error.name,
        message: error.message,
        stack: error.stack,
      };
    } else if (error !== undefined) {
      errorMeta.error = error;
    }

    this.logger.error(message, errorMeta);
  }

  /**
   * Log with TAG reference for traceability
   */
  logWithTag(
    level: 'debug' | 'info' | 'warn' | 'error',
    tag: string,
    message: string,
    meta?: Record<string, unknown>
  ): void {
    const taggedMeta = { ...meta, tag };
    this.logger.log(level, message, taggedMeta);
  }

  /**
   * Set verbose mode (enables detailed debug output)
   * @param verbose - Enable or disable verbose mode
   */
  setVerbose(verbose: boolean): void {
    this.verboseMode = verbose;
    // Update logger level to debug when verbose is enabled
    if (verbose) {
      this.logger.level = 'debug';
    } else {
      this.logger.level = this.options.level;
    }
  }

  /**
   * Get verbose mode status
   * @returns Current verbose mode status
   */
  isVerbose(): boolean {
    return this.verboseMode;
  }

  /**
   * Log verbose message (only shown in verbose mode)
   * Use this for detailed debugging information
   */
  verbose(message: string, meta?: Record<string, unknown>): void {
    if (this.verboseMode) {
      this.logger.info(message, meta);
    }
  }

  /**
   * Log user-facing message (always shown, regardless of verbose mode)
   * Use this for important user messages like success, errors, warnings
   */
  log(message: string): void {
    // Direct console output for clean user-facing messages
    console.log(message);
  }

  /**
   * Log success message with emoji
   */
  success(message: string): void {
    console.log(message);
  }

  /**
   * Log error message with emoji (always shown)
   */
  errorMessage(message: string): void {
    console.error(message);
  }
}

/**
 * Default logger instance
 */
export const logger = new MoaiLogger();

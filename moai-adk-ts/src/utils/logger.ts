/**
 * @file Structured logging utility with user-friendly output support
 * @author MoAI Team
 * @tags @FEATURE:STRUCTURED-LOGGING-001 @REQ:TRUST-SECURE-001
 */

import chalk from 'chalk';
import { logger as winstonLogger } from './winston-logger.js';

/**
 * Log levels for structured logging
 * @tags @DESIGN:LOG-LEVELS-001
 */
export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

/**
 * Log output format modes
 * @tags @DESIGN:LOG-FORMAT-001
 */
export type LogFormat = 'json' | 'human';

/**
 * Determine the appropriate log format based on environment
 * @returns Log format to use
 * @tags @UTIL:DETECT-FORMAT-001
 */
function getLogFormat(): LogFormat {
  // Use JSON format for development or when LOG_FORMAT is explicitly set to json
  if (
    process.env['NODE_ENV'] === 'development' ||
    process.env['LOG_FORMAT'] === 'json'
  ) {
    return 'json';
  }

  // Use human-readable format for CLI usage and production
  return 'human';
}

/**
 * Check if output should be colorized
 * @returns True if colors should be used
 * @tags @UTIL:SHOULD-COLORIZE-001
 */
function shouldColorize(): boolean {
  return (
    process.stdout.isTTY &&
    process.env['TERM'] !== 'dumb' &&
    process.env['NO_COLOR'] === undefined
  );
}

/**
 * Structured log entry
 * @tags @DESIGN:LOG-ENTRY-001
 */
export interface LogEntry {
  readonly timestamp: string;
  readonly level: LogLevel;
  readonly message: string;
  readonly component?: string;
  readonly tags?: string[];
  readonly context?: Record<string, unknown>;
  readonly error?: {
    readonly name: string;
    readonly message: string;
    readonly stack?: string;
  };
}

/**
 * Structured logger for TRUST compliance
 * @tags @FEATURE:STRUCTURED-LOGGER-001
 */
export class Logger {
  constructor(private readonly component: string = 'moai-adk') {}

  /**
   * Create structured log entry
   * @param level - Log level
   * @param message - Log message
   * @param context - Additional context
   * @returns Structured log entry
   * @tags @UTIL:CREATE-LOG-ENTRY-001
   */
  private createLogEntry(
    level: LogLevel,
    message: string,
    context?: Record<string, unknown>
  ): LogEntry {
    const entry: LogEntry = {
      timestamp: new Date().toISOString(),
      level,
      message: this.maskSensitiveData(message),
      component: this.component,
    };

    if (context) {
      (entry as LogEntry & { context: Record<string, unknown> }).context =
        this.maskSensitiveContext(context);
    }

    return entry;
  }

  /**
   * Mask sensitive data in log messages
   * @param message - Original message
   * @returns Masked message
   * @tags @UTIL:MASK-SENSITIVE-001
   */
  private maskSensitiveData(message: string): string {
    // Mask common sensitive patterns
    return message
      .replace(/password=\S+/gi, 'password=***redacted***')
      .replace(/token=\S+/gi, 'token=***redacted***')
      .replace(/api[_-]?key=\S+/gi, 'api_key=***redacted***')
      .replace(/secret=\S+/gi, 'secret=***redacted***');
  }

  /**
   * Mask sensitive data in context
   * @param context - Original context
   * @returns Masked context
   * @tags @UTIL:MASK-CONTEXT-001
   */
  private maskSensitiveContext(
    context: Record<string, unknown>
  ): Record<string, unknown> {
    const masked = { ...context };
    const sensitiveKeys = [
      'password',
      'token',
      'secret',
      'key',
      'apiKey',
      'accessToken',
    ];

    for (const key of Object.keys(masked)) {
      if (
        sensitiveKeys.some(sensitive => key.toLowerCase().includes(sensitive))
      ) {
        masked[key] = '***redacted***';
      }
    }

    return masked;
  }

  /**
   * Format log entry for human-readable output
   * @param entry - Log entry to format
   * @returns Formatted string
   * @tags @UTIL:FORMAT-HUMAN-001
   */
  private formatHumanReadable(entry: LogEntry): string {
    const colorize = shouldColorize();
    const { level, message, context } = entry;

    // Level indicators
    const levelColors = {
      debug: chalk.gray,
      info: chalk.blue,
      warn: chalk.yellow,
      error: chalk.red,
    };

    const levelColor = levelColors[level];

    // Format the main message
    let output = colorize && levelColor ? levelColor(message) : message;

    // Add context if present and relevant for user display
    if (context) {
      // Only show user-relevant context, not debug tags
      const userContext = Object.entries(context)
        .filter(([key]) => !key.startsWith('tag') && !key.startsWith('@'))
        .map(([key, value]) => `${key}: ${value}`)
        .join(', ');

      if (userContext) {
        output += colorize
          ? chalk.gray(` (${userContext})`)
          : ` (${userContext})`;
      }
    }

    return output;
  }

  /**
   * Output log entry in appropriate format
   * @param entry - Log entry to output
   * @param level - Log level for Winston logger
   * @tags @UTIL:OUTPUT-LOG-001
   */
  private outputLog(
    entry: LogEntry,
    level: 'debug' | 'info' | 'warn' | 'error'
  ): void {
    const format = getLogFormat();

    if (format === 'json') {
      // JSON format: delegate to Winston logger
      winstonLogger[level](entry.message, entry.context);
    } else {
      // Human-readable format - suppress debug logs in human mode
      if (entry.level === 'debug') {
        return; // Skip debug logs in human-readable mode
      }

      const formatted = this.formatHumanReadable(entry);
      // Use Winston logger instead of console.*
      winstonLogger[level](formatted);
    }
  }

  /**
   * Log debug message
   * @param message - Debug message
   * @param context - Additional context
   * @tags @API:LOG-DEBUG-001
   */
  public debug(message: string, context?: Record<string, unknown>): void {
    const entry = this.createLogEntry('debug', message, context);
    this.outputLog(entry, 'debug');
  }

  /**
   * Log info message
   * @param message - Info message
   * @param context - Additional context
   * @tags @API:LOG-INFO-001
   */
  public info(message: string, context?: Record<string, unknown>): void {
    const entry = this.createLogEntry('info', message, context);
    this.outputLog(entry, 'info');
  }

  /**
   * Log warning message
   * @param message - Warning message
   * @param context - Additional context
   * @tags @API:LOG-WARN-001
   */
  public warn(message: string, context?: Record<string, unknown>): void {
    const entry = this.createLogEntry('warn', message, context);
    this.outputLog(entry, 'warn');
  }

  /**
   * Log error message
   * @param message - Error message
   * @param error - Error object
   * @param context - Additional context
   * @tags @API:LOG-ERROR-001
   */
  public error(
    message: string,
    error?: Error,
    context?: Record<string, unknown>
  ): void {
    const baseEntry = this.createLogEntry('error', message, context);
    const entry: LogEntry = error
      ? {
          ...baseEntry,
          error: {
            name: error.name,
            message: error.message,
            ...(error.stack && { stack: error.stack }),
          },
        }
      : baseEntry;
    this.outputLog(entry, 'error');
  }

  /**
   * Log a simple message without JSON formatting (always human-readable)
   * @param message - Message to log
   * @param level - Log level for styling
   * @tags @API:LOG-SIMPLE-001
   */
  public simple(message: string, level: LogLevel = 'info'): void {
    const colorize = shouldColorize();

    const levelColors = {
      debug: chalk.gray,
      info: chalk.blue,
      warn: chalk.yellow,
      error: chalk.red,
    };

    const levelColor = levelColors[level];
    const output = colorize && levelColor ? levelColor(message) : message;
    winstonLogger.info(output);
  }
}

/**
 * Default logger instance
 * @tags @SINGLETON:LOGGER-001
 */
export const logger = new Logger();

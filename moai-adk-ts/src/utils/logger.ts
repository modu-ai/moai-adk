/**
 * @file Structured logging utility
 * @author MoAI Team
 * @tags @FEATURE:STRUCTURED-LOGGING-001 @REQ:TRUST-SECURE-001
 */

/**
 * Log levels for structured logging
 * @tags @DESIGN:LOG-LEVELS-001
 */
export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

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
   * Log debug message
   * @param message - Debug message
   * @param context - Additional context
   * @tags @API:LOG-DEBUG-001
   */
  public debug(message: string, context?: Record<string, unknown>): void {
    const entry = this.createLogEntry('debug', message, context);
    console.debug(JSON.stringify(entry));
  }

  /**
   * Log info message
   * @param message - Info message
   * @param context - Additional context
   * @tags @API:LOG-INFO-001
   */
  public info(message: string, context?: Record<string, unknown>): void {
    const entry = this.createLogEntry('info', message, context);
    console.log(JSON.stringify(entry));
  }

  /**
   * Log warning message
   * @param message - Warning message
   * @param context - Additional context
   * @tags @API:LOG-WARN-001
   */
  public warn(message: string, context?: Record<string, unknown>): void {
    const entry = this.createLogEntry('warn', message, context);
    console.warn(JSON.stringify(entry));
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
    console.error(JSON.stringify(entry));
  }
}

/**
 * Default logger instance
 * @tags @SINGLETON:LOGGER-001
 */
export const logger = new Logger();

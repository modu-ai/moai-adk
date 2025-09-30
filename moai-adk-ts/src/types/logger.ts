// @DATA:LOG-TYPES-001 | Chain: @REQ:LOG-001 -> @DESIGN:LOG-001 -> @TASK:LOG-001 -> @TEST:LOG-001
// Related: @FEATURE:LOG-001

/**
 * @file Logger type definitions
 * @author MoAI Team
 */

import type winston from 'winston';

/**
 * Log levels for structured logging
 * @tags @DESIGN:LOG-LEVELS-001
 */
export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

/**
 * Logger configuration options
 */
export interface LoggerOptions {
  level?: string;
  enableFile?: boolean;
  enableConsole?: boolean;
  logDir?: string;
  transports?: winston.transport[]; // DI: Allow custom transports
}

/**
 * Log context with optional TAG and metadata
 */
export interface LogContext {
  tag?: string;
  tags?: string[];
  userId?: string;
  operation?: string;
  [key: string]: unknown;
}

/**
 * Structured log entry
 */
export interface LogEntry {
  readonly timestamp: string;
  readonly level: string;
  readonly message: string;
  readonly tag?: string;
  readonly tags?: string[];
  readonly context?: Record<string, unknown>;
  readonly error?: {
    readonly name: string;
    readonly message: string;
    readonly stack?: string;
  };
}
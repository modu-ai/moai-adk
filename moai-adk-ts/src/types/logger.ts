/**
 * @file Logger type definitions
 * @author MoAI Team
 * @tags @DESIGN:WINSTON-LOGGER-001
 */

import type winston from 'winston';

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
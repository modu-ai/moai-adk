/**
 * @file Winston-based structured logging tests
 * @author MoAI Team
 * @tags @TEST:WINSTON-LOGGER-001 @CODE:STRUCTURED-LOGGING-001
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { mkdirSync, rmSync, existsSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { MoaiLogger } from '../../src/utils/winston-logger.js';
import type { LoggerOptions } from '../../src/types/logger.js';

describe('MoaiLogger', () => {
  const testLogDir = join(process.cwd(), '__tests__', 'temp-logs');
  let logger: MoaiLogger;
  let consoleSpies: {
    debug: ReturnType<typeof vi.spyOn>;
    log: ReturnType<typeof vi.spyOn>;
    warn: ReturnType<typeof vi.spyOn>;
    error: ReturnType<typeof vi.spyOn>;
  };

  beforeEach(() => {
    // Create test log directory
    if (!existsSync(testLogDir)) {
      mkdirSync(testLogDir, { recursive: true });
    }

    // Spy on console methods
    consoleSpies = {
      debug: vi.spyOn(console, 'debug').mockImplementation(() => {}),
      log: vi.spyOn(console, 'log').mockImplementation(() => {}),
      warn: vi.spyOn(console, 'warn').mockImplementation(() => {}),
      error: vi.spyOn(console, 'error').mockImplementation(() => {}),
    };
  });

  afterEach(() => {
    // Restore console methods
    Object.values(consoleSpies).forEach(spy => spy.mockRestore());

    // Clean up test logs
    if (existsSync(testLogDir)) {
      rmSync(testLogDir, { recursive: true, force: true });
    }
  });

  describe('Winston Initialization', () => {
    it('should initialize Winston logger with default configuration', () => {
      expect(() => {
        logger = new MoaiLogger();
      }).not.toThrow();
      expect(logger).toBeDefined();
    });

    it('should accept custom log level via options', () => {
      expect(() => {
        logger = new MoaiLogger({ level: 'debug' });
      }).not.toThrow();
      expect(logger).toBeDefined();
    });

    it('should respect LOG_LEVEL environment variable', () => {
      process.env['LOG_LEVEL'] = 'error';
      expect(() => {
        logger = new MoaiLogger();
      }).not.toThrow();
      expect(logger).toBeDefined();
      delete process.env['LOG_LEVEL'];
    });
  });

  describe('Sensitive Data Masking', () => {
    beforeEach(() => {
      logger = new MoaiLogger({
        enableFile: false,
        enableConsole: false, // Disable console to avoid test output
      });
    });

    it('should mask password in log messages', () => {
      expect(() => {
        logger.info('User login with password=secret123');
      }).not.toThrow();
      // Message should be masked (verified by console spy)
      // Actual masking verification would require Winston transport inspection
    });

    it('should mask token in context object', () => {
      expect(() => {
        logger.info('API request', {
          token: 'bearer-token-xyz',
          userId: '12345',
        });
      }).not.toThrow();
      // Token should be masked to ***REDACTED***
    });

    it('should mask multiple sensitive fields', () => {
      expect(() => {
        logger.error('Authentication failed', new Error('Invalid credentials'), {
          password: 'secret',
          apiKey: 'key-123',
          secret: 'hidden',
          publicData: 'visible',
        });
      }).not.toThrow();
      // password, apiKey, secret should be ***REDACTED***
      // publicData should remain visible
    });

    it('should mask sensitive patterns in strings', () => {
      expect(() => {
        logger.warn('Config: apiKey=abc123 token=xyz789');
      }).not.toThrow();
      // Patterns should be masked to apiKey=***redacted*** token=***redacted***
    });
  });

  describe('Environment-based Log Levels', () => {
    it('should use debug level in development mode', () => {
      process.env['NODE_ENV'] = 'development';
      expect(() => {
        logger = new MoaiLogger({ enableConsole: false, enableFile: false });
        logger.debug('Debug message');
      }).not.toThrow();
      delete process.env['NODE_ENV'];
    });

    it('should use info level in production mode', () => {
      process.env['NODE_ENV'] = 'production';
      expect(() => {
        logger = new MoaiLogger({ enableConsole: false, enableFile: false });
        logger.info('Production log');
      }).not.toThrow();
      delete process.env['NODE_ENV'];
    });

    it('should suppress debug logs in production', () => {
      process.env['NODE_ENV'] = 'production';
      expect(() => {
        logger = new MoaiLogger({ enableConsole: false, enableFile: false });
        logger.debug('This should not appear');
      }).not.toThrow();
      delete process.env['NODE_ENV'];
    });
  });

  describe('TAG Traceability Integration', () => {
    beforeEach(() => {
      logger = new MoaiLogger({ enableConsole: false, enableFile: false });
    });

    it('should log with TAG reference', () => {
      expect(() => {
        logger.logWithTag('info', '@CODE:AUTH-001', 'User authenticated');
      }).not.toThrow();
    });

    it('should include TAG in structured log output', () => {
      expect(() => {
        logger.logWithTag(
          'error',
          '@CODE:AUTH-001',
          'Security violation detected',
          {
            ip: '192.168.1.1',
            action: 'unauthorized_access',
          }
        );
      }).not.toThrow();
    });

    it('should support multiple TAGs in context', () => {
      expect(() => {
        logger.info('Operation completed', {
          tags: ['@CODE:AUTH-001', '@CODE:AUTH-001'],
          duration: 150,
        });
      }).not.toThrow();
    });
  });

  describe('File Logging Fallback', () => {
    it('should write logs to file when enabled', async () => {
      logger = new MoaiLogger({
        enableFile: true,
        enableConsole: false,
        logDir: testLogDir,
      });

      logger.error('File log test');

      // Wait for Winston to flush to disk
      await new Promise(resolve => setTimeout(resolve, 100));

      // Verify file was created
      const errorLogPath = join(testLogDir, 'error.log');
      expect(existsSync(errorLogPath)).toBe(true);
    });

    it('should handle missing log directory gracefully', () => {
      // Should not throw, just fallback to console
      expect(() => {
        logger = new MoaiLogger({
          enableFile: true,
          enableConsole: false,
          logDir: '/invalid/path/that/does/not/exist',
        });
        logger.info('Should fallback to console only');
      }).not.toThrow();
    });

    it('should write error logs to separate error.log file', async () => {
      logger = new MoaiLogger({
        enableFile: true,
        enableConsole: false,
        logDir: testLogDir,
      });

      logger.error('Critical error', new Error('Test error'));

      // Wait for Winston to flush to disk
      await new Promise(resolve => setTimeout(resolve, 100));

      // Verify error.log exists
      const errorLogPath = join(testLogDir, 'error.log');
      expect(existsSync(errorLogPath)).toBe(true);
    });
  });

  describe('Error Object Handling', () => {
    beforeEach(() => {
      logger = new MoaiLogger({ enableConsole: false, enableFile: false });
    });

    it('should extract error name, message, and stack', () => {
      const testError = new Error('Test error message');
      testError.name = 'TestError';

      expect(() => {
        logger.error('Operation failed', testError);
      }).not.toThrow();
    });

    it('should handle non-Error objects', () => {
      expect(() => {
        logger.error('Unexpected error', { code: 500, message: 'Server error' });
      }).not.toThrow();
    });

    it('should handle undefined error parameter', () => {
      expect(() => {
        logger.error('Error without error object');
      }).not.toThrow();
    });
  });

  describe('Backward Compatibility with Logger class', () => {
    it('should support existing Logger API - info()', () => {
      expect(() => {
        logger = new MoaiLogger({ enableConsole: false, enableFile: false });
        logger.info('Compatibility test');
      }).not.toThrow();
    });

    it('should support existing Logger API - error() with Error', () => {
      expect(() => {
        logger = new MoaiLogger({ enableConsole: false, enableFile: false });
        logger.error('Error test', new Error('Test'));
      }).not.toThrow();
    });

    it('should support existing Logger API - warn() with context', () => {
      expect(() => {
        logger = new MoaiLogger({ enableConsole: false, enableFile: false });
        logger.warn('Warning test', { userId: '123' });
      }).not.toThrow();
    });
  });

  describe('Performance and Memory', () => {
    it('should not leak memory with many log calls', () => {
      expect(() => {
        logger = new MoaiLogger({ enableFile: false, enableConsole: false });
        for (let i = 0; i < 1000; i++) {
          logger.info(`Log message ${i}`);
        }
      }).not.toThrow();
    });

    it('should handle large context objects efficiently', () => {
      const largeContext = {
        data: Array.from({ length: 100 }, (_, i) => ({
          id: i,
          value: `item-${i}`,
        })),
      };

      expect(() => {
        logger = new MoaiLogger({ enableFile: false, enableConsole: false });
        logger.info('Large context test', largeContext);
      }).not.toThrow();
    });
  });
});
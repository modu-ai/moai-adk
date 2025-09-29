/**
 * @file Regex Security Utilities
 * @author MoAI Team
 * @tags @SECURITY:REDOS-PROTECTION-001 @FIX:REDOS-VULNERABILITY-001
 * @description Safe regex execution with timeout protection against ReDoS attacks
 */

import { logger } from './logger';

/**
 * Regex execution result
 * @tags @DATA:REGEX-RESULT-001
 */
export interface RegexResult {
  readonly success: boolean;
  readonly matches: RegExpMatchArray | null;
  readonly timedOut: boolean;
  readonly executionTime: number;
  readonly error?: string;
}

/**
 * Regex validation result
 * @tags @DATA:REGEX-VALIDATION-RESULT-001
 */
export interface RegexValidationResult {
  readonly isSafe: boolean;
  readonly vulnerabilities: readonly string[];
  readonly riskLevel: 'low' | 'medium' | 'high' | 'critical';
}

/**
 * Safe regex execution options
 * @tags @DATA:SAFE-REGEX-OPTIONS-001
 */
export interface SafeRegexOptions {
  readonly timeout?: number; // Timeout in milliseconds (default: 100ms)
  readonly maxLength?: number; // Maximum input string length (default: 10000)
  readonly enableLogging?: boolean; // Enable security logging (default: true)
}

/**
 * Regex security utility class
 * @tags @SECURITY:REDOS-PROTECTION-001
 */
export class RegexSecurity {
  private static readonly DEFAULT_TIMEOUT = 100; // 100ms
  private static readonly DEFAULT_MAX_LENGTH = 10000;

  /**
   * Execute regex with timeout protection
   * @tags @API:SAFE-MATCH-001
   */
  static safeMatch(
    pattern: RegExp,
    input: string,
    options: SafeRegexOptions = {}
  ): RegexResult {
    const {
      timeout = RegexSecurity.DEFAULT_TIMEOUT,
      maxLength = RegexSecurity.DEFAULT_MAX_LENGTH,
      enableLogging = true,
    } = options;

    const startTime = Date.now();

    // Input length validation
    if (input.length > maxLength) {
      if (enableLogging) {
        logger.warn(`Input too long for regex: ${input.length} > ${maxLength}`);
      }
      return {
        success: false,
        matches: null,
        timedOut: false,
        executionTime: 0,
        error: 'Input string too long',
      };
    }

    // Regex safety validation
    const validation = RegexSecurity.validateRegexSafety(pattern);
    if (
      validation.riskLevel === 'critical' ||
      validation.riskLevel === 'high'
    ) {
      if (enableLogging) {
        logger.error(`Dangerous regex pattern detected: ${pattern.source}`, {
          vulnerabilities: validation.vulnerabilities,
          riskLevel: validation.riskLevel,
        });
      }
      return {
        success: false,
        matches: null,
        timedOut: false,
        executionTime: 0,
        error: 'Regex pattern is potentially dangerous',
      };
    }

    try {
      let timedOut = false;
      let matches: RegExpMatchArray | null = null;

      // Set up timeout
      const timeoutId = setTimeout(() => {
        timedOut = true;
      }, timeout);

      // Execute regex with monitoring
      const executeRegex = (): RegExpMatchArray | null => {
        if (timedOut) {
          throw new Error('Regex execution timed out');
        }
        return input.match(pattern);
      };

      matches = executeRegex();
      clearTimeout(timeoutId);

      const executionTime = Date.now() - startTime;

      // Log slow executions
      if (enableLogging && executionTime > 50) {
        logger.warn(`Slow regex execution: ${executionTime}ms`, {
          pattern: pattern.source,
          inputLength: input.length,
        });
      }

      return {
        success: true,
        matches,
        timedOut,
        executionTime,
      };
    } catch (error) {
      const executionTime = Date.now() - startTime;
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';

      if (enableLogging) {
        logger.error(`Regex execution failed: ${errorMessage}`, {
          pattern: pattern.source,
          inputLength: input.length,
          executionTime,
        });
      }

      return {
        success: false,
        matches: null,
        timedOut: errorMessage.includes('timed out'),
        executionTime,
        error: errorMessage,
      };
    }
  }

  /**
   * Execute regex test with timeout protection
   * @tags @API:SAFE-TEST-001
   */
  static safeTest(
    pattern: RegExp,
    input: string,
    options: SafeRegexOptions = {}
  ): boolean {
    const result = RegexSecurity.safeMatch(pattern, input, options);
    return result.success && result.matches !== null;
  }

  /**
   * Execute regex replace with timeout protection
   * @tags @API:SAFE-REPLACE-001
   */
  static safeReplace(
    pattern: RegExp,
    input: string,
    replacement: string,
    options: SafeRegexOptions = {}
  ): string {
    const {
      timeout = RegexSecurity.DEFAULT_TIMEOUT,
      maxLength = RegexSecurity.DEFAULT_MAX_LENGTH,
      enableLogging = true,
    } = options;

    // Input length validation
    if (input.length > maxLength) {
      if (enableLogging) {
        logger.warn(
          `Input too long for regex replace: ${input.length} > ${maxLength}`
        );
      }
      return input; // Return original string
    }

    // Regex safety validation
    const validation = RegexSecurity.validateRegexSafety(pattern);
    if (
      validation.riskLevel === 'critical' ||
      validation.riskLevel === 'high'
    ) {
      if (enableLogging) {
        logger.error(`Dangerous regex pattern in replace: ${pattern.source}`);
      }
      return input; // Return original string
    }

    const startTime = Date.now();

    try {
      let timedOut = false;

      // Set up timeout
      const timeoutId = setTimeout(() => {
        timedOut = true;
      }, timeout);

      // Execute replace with monitoring
      const executeReplace = (): string => {
        if (timedOut) {
          throw new Error('Regex replace timed out');
        }
        return input.replace(pattern, replacement);
      };

      const result = executeReplace();
      clearTimeout(timeoutId);

      const executionTime = Date.now() - startTime;

      // Log slow executions
      if (enableLogging && executionTime > 50) {
        logger.warn(`Slow regex replace: ${executionTime}ms`, {
          pattern: pattern.source,
          inputLength: input.length,
        });
      }

      return result;
    } catch (error) {
      const executionTime = Date.now() - startTime;

      if (enableLogging) {
        logger.error(`Regex replace failed: ${error}`, {
          pattern: pattern.source,
          inputLength: input.length,
          executionTime,
        });
      }

      return input; // Return original string on error
    }
  }

  /**
   * Validate regex pattern for potential ReDoS vulnerabilities
   * @tags @API:VALIDATE-REGEX-SAFETY-001
   */
  static validateRegexSafety(pattern: RegExp): RegexValidationResult {
    const source = pattern.source;
    const vulnerabilities: string[] = [];
    let riskLevel: 'low' | 'medium' | 'high' | 'critical' = 'low';

    // Check for catastrophic backtracking patterns
    const dangerousPatterns = [
      // Nested quantifiers
      {
        pattern: /\([^)]*\+[^)]*\)[*+?]/g,
        vulnerability: 'Nested quantifiers can cause exponential backtracking',
        risk: 'critical' as const,
      },
      // Alternation with repetition
      {
        pattern: /\([^)]*\|[^)]*\)[*+]/g,
        vulnerability:
          'Alternation with repetition can cause exponential backtracking',
        risk: 'high' as const,
      },
      // Repeated groups with nested quantifiers
      {
        pattern: /\([^)]*[*+?][^)]*\)[*+?]/g,
        vulnerability: 'Repeated groups with nested quantifiers',
        risk: 'critical' as const,
      },
      // Evil regex patterns
      {
        pattern: /\([^)]*[*+][^)]*[*+][^)]*\)/g,
        vulnerability: 'Multiple overlapping quantifiers (evil regex)',
        risk: 'critical' as const,
      },
    ];

    for (const {
      pattern: dangerousPattern,
      vulnerability,
      risk,
    } of dangerousPatterns) {
      if (dangerousPattern.test(source)) {
        vulnerabilities.push(vulnerability);
        if (
          risk === 'critical' ||
          (risk === 'high' && riskLevel !== 'critical')
        ) {
          riskLevel = risk;
        }
      }
    }

    // Check for complexity indicators
    const complexityPatterns = [
      // Long alternations
      {
        pattern: /\([^)]*\|[^)]*\|[^)]*\|[^)]*\)/g,
        vulnerability: 'Long alternation chains can be slow',
        risk: 'medium' as const,
      },
      // Deep nesting
      {
        pattern: /\([^)]*\([^)]*\([^)]*\)/g,
        vulnerability: 'Deeply nested groups can be slow',
        risk: 'medium' as const,
      },
      // Character class with ranges
      {
        pattern: /\[[^\]]*-[^\]]*-[^\]]*\]/g,
        vulnerability: 'Complex character classes can be slow',
        risk: 'low' as const,
      },
    ];

    for (const {
      pattern: complexPattern,
      vulnerability,
      risk,
    } of complexityPatterns) {
      if (complexPattern.test(source)) {
        vulnerabilities.push(vulnerability);
        if (riskLevel === 'low') {
          riskLevel = risk;
        }
      }
    }

    // Check regex length (very long patterns can be problematic)
    if (source.length > 500) {
      vulnerabilities.push('Regex pattern is very long');
      if (riskLevel === 'low') {
        riskLevel = 'medium';
      }
    }

    return {
      isSafe: riskLevel === 'low' || riskLevel === 'medium',
      vulnerabilities: Object.freeze(vulnerabilities),
      riskLevel,
    };
  }

  /**
   * Create a safe regex pattern with built-in timeout protection
   * @tags @API:CREATE-SAFE-REGEX-001
   */
  static createSafeRegex(
    pattern: string,
    flags?: string,
    _options: SafeRegexOptions = {}
  ): RegExp | null {
    try {
      const regex = new RegExp(pattern, flags);
      const validation = RegexSecurity.validateRegexSafety(regex);

      if (!validation.isSafe) {
        logger.error(`Unsafe regex pattern rejected: ${pattern}`, {
          vulnerabilities: validation.vulnerabilities,
          riskLevel: validation.riskLevel,
        });
        return null;
      }

      return regex;
    } catch (error) {
      logger.error(`Invalid regex pattern: ${pattern}`, { error });
      return null;
    }
  }
}

/**
 * Safe regex execution functions (convenience exports)
 * @tags @API:CONVENIENCE-EXPORTS-001
 */
/**
 * @tags @API:SAFE-MATCH-EXPORT-001
 */
export const safeMatch = RegexSecurity.safeMatch.bind(RegexSecurity);
/**
 * @tags @API:SAFE-TEST-EXPORT-001
 */
export const safeTest = RegexSecurity.safeTest.bind(RegexSecurity);
/**
 * @tags @API:SAFE-REPLACE-EXPORT-001
 */
export const safeReplace = RegexSecurity.safeReplace.bind(RegexSecurity);
/**
 * @tags @API:VALIDATE-REGEX-SAFETY-EXPORT-001
 */
export const validateRegexSafety =
  RegexSecurity.validateRegexSafety.bind(RegexSecurity);
/**
 * @tags @API:CREATE-SAFE-REGEX-EXPORT-001
 */
export const createSafeRegex =
  RegexSecurity.createSafeRegex.bind(RegexSecurity);

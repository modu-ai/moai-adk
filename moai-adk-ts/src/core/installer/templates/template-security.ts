// @CODE:INST-006 |
// Related: @CODE:INST-006:API

/**
 * @file Template security validation
 * @author MoAI Team
 */

import { logger } from '../../../utils/logger';

/**
 * Allowed template context property names (whitelist)
 */
const ALLOWED_CONTEXT_KEYS = new Set([
  // Project metadata
  'PROJECT_NAME',
  'PROJECT_VERSION',
  'PROJECT_DESCRIPTION',
  'PROJECT_AUTHOR',
  'PROJECT_LICENSE',
  'PROJECT_REPOSITORY',

  // Environment variables
  'ENVIRONMENT',
  'NODE_ENV',
  'NODE_VERSION',
  'NPM_VERSION',

  // Path variables
  'PROJECT_PATH',
  'SOURCE_PATH',
  'BUILD_PATH',
  'DIST_PATH',

  // Template variables
  'TEMPLATE_NAME',
  'TEMPLATE_VERSION',
  'GENERATED_DATE',
  'GENERATED_TIME',

  // Language-specific
  'LANGUAGE',
  'FRAMEWORK',
  'PACKAGE_MANAGER',

  // Git information
  'GIT_BRANCH',
  'GIT_COMMIT',
  'GIT_AUTHOR',

  // Boolean flags
  'IS_TYPESCRIPT',
  'IS_JAVASCRIPT',
  'IS_PYTHON',
  'IS_DEVELOPMENT',
  'IS_PRODUCTION',
  'IS_TEST',
]);

/**
 * Dangerous property names that should never be allowed
 */
const DANGEROUS_PROPERTIES = new Set([
  'constructor',
  'prototype',
  '__proto__',
  '__defineGetter__',
  '__defineSetter__',
  '__lookupGetter__',
  '__lookupSetter__',
  'hasOwnProperty',
  'isPrototypeOf',
  'propertyIsEnumerable',
  'valueOf',
  'toString',
  'eval',
  'Function',
  'require',
  'process',
  'global',
  'Buffer',
  'setImmediate',
  'clearImmediate',
  'setTimeout',
  'clearTimeout',
  'setInterval',
  'clearInterval',
]);

/**
 * Safe template context interface
 */
export interface SafeTemplateContext {
  readonly [key: string]: string | number | boolean;
}

/**
 * Context sanitization result
 */
export interface ContextSanitizationResult {
  readonly sanitizedContext: SafeTemplateContext;
  readonly removedKeys: readonly string[];
  readonly warnings: readonly string[];
}

/**
 * Sanitize template context to prevent injection attacks
 *
 * @param context - Raw template context
 * @returns Sanitized context with dangerous properties removed
 */
export function sanitizeTemplateContext(
  context: Record<string, any>
): ContextSanitizationResult {
  const sanitizedContext: Record<string, string | number | boolean> = {};
  const removedKeys: string[] = [];
  const warnings: string[] = [];

  for (const [key, value] of Object.entries(context)) {
    // Check for dangerous property names
    if (DANGEROUS_PROPERTIES.has(key)) {
      removedKeys.push(key);
      warnings.push(`Removed dangerous property: ${key}`);
      continue;
    }

    // Check for dangerous property access patterns
    if (
      key.includes('__') ||
      key.includes('constructor') ||
      key.includes('prototype')
    ) {
      removedKeys.push(key);
      warnings.push(`Removed suspicious property: ${key}`);
      continue;
    }

    // Whitelist validation
    if (!ALLOWED_CONTEXT_KEYS.has(key)) {
      removedKeys.push(key);
      warnings.push(`Property '${key}' not in whitelist`);
      continue;
    }

    // Sanitize values
    const sanitizedValue = sanitizeValue(value);
    if (sanitizedValue !== null) {
      sanitizedContext[key] = sanitizedValue;
    } else {
      removedKeys.push(key);
      warnings.push(`Invalid value type for property: ${key}`);
    }
  }

  // Log security actions
  if (removedKeys.length > 0) {
    logger.warn(
      `Template security: removed ${removedKeys.length} dangerous properties`,
      {
        removedKeys,
        originalKeyCount: Object.keys(context).length,
        sanitizedKeyCount: Object.keys(sanitizedContext).length,
      }
    );
  }

  return {
    sanitizedContext: Object.freeze(sanitizedContext),
    removedKeys: Object.freeze(removedKeys),
    warnings: Object.freeze(warnings),
  };
}

/**
 * Sanitize individual context values
 *
 * @param value - Value to sanitize
 * @returns Sanitized value or null if invalid
 */
function sanitizeValue(value: any): string | number | boolean | null {
  // Only allow primitive types
  if (typeof value === 'string') {
    // Remove potentially dangerous string content
    return value
      .replace(/\${.*}/g, '') // Remove ${...} expressions
      .replace(/`.*`/g, '') // Remove template literals
      .replace(/eval\s*\(/gi, '') // Remove eval calls
      .replace(/function\s*\(/gi, '') // Remove function definitions
      .replace(/=>\s*{/g, '') // Remove arrow functions
      .replace(/javascript:/gi, '') // Remove javascript: URIs
      .slice(0, 1000); // Limit string length
  }

  if (typeof value === 'number') {
    // Ensure finite numbers only
    return Number.isFinite(value) ? value : null;
  }

  if (typeof value === 'boolean') {
    return value;
  }

  // Reject all other types (objects, arrays, functions, etc.)
  return null;
}

/**
 * Validate template content for dangerous patterns
 *
 * @param templateContent - Template content to validate
 * @returns True if safe, false if dangerous patterns detected
 */
export function validateTemplateContent(templateContent: string): boolean {
  const dangerousPatterns = [
    // JavaScript execution patterns
    /javascript:/gi,
    /eval\s*\(/gi,
    /Function\s*\(/gi,
    /setTimeout\s*\(/gi,
    /setInterval\s*\(/gi,

    // Constructor access patterns
    /constructor\s*\./gi,
    /prototype\s*\./gi,
    /__proto__/gi,

    // Process/global access
    /process\s*\./gi,
    /global\s*\./gi,
    /require\s*\(/gi,

    // Template injection attempts
    /\{\{\s*constructor/gi,
    /\{\{\s*__proto__/gi,
    /\{\{\s*prototype/gi,
  ];

  for (const pattern of dangerousPatterns) {
    if (pattern.test(templateContent)) {
      logger.error(`Dangerous pattern detected in template: ${pattern.source}`);
      return false;
    }
  }

  return true;
}

/**
 * Create safe Mustache rendering function with context sanitization
 *
 * @param template - Template string
 * @param context - Template context
 * @returns Safely rendered template
 */
export function renderTemplateSafely(
  template: string,
  context: Record<string, any>
): string {
  // Validate template content first
  if (!validateTemplateContent(template)) {
    throw new Error('Template contains dangerous patterns');
  }

  // Sanitize context
  const { sanitizedContext, warnings } = sanitizeTemplateContext(context);

  // Log warnings if any
  if (warnings.length > 0) {
    logger.warn('Template context sanitization warnings:', warnings);
  }

  try {
    // Use dynamic import to prevent static analysis issues
    const Mustache = require('mustache');

    // Disable Mustache functions to prevent code execution
    const oldWriter = Mustache.Writer;
    Mustache.Writer = () => {
      const writer = new oldWriter();
      writer.compileFn = null; // Disable function compilation
      return writer;
    };

    // Render with sanitized context only
    const result = Mustache.render(template, sanitizedContext);

    // Restore original writer
    Mustache.Writer = oldWriter;

    return result;
  } catch (error) {
    logger.error('Safe template rendering failed:', error);
    throw new Error(
      `Template rendering failed: ${error instanceof Error ? error.message : 'Unknown error'}`
    );
  }
}

// @CODE:VALIDATOR-005:RULES | Chain: @SPEC:QUAL-001 -> @SPEC:QUAL-001 -> @CODE:QUAL-001 -> @TEST:UTIL-006
// Related: @CODE:INPUT-VALIDATION-001

/**
 * @file Common validation rules and patterns
 * @author MoAI Team
 */

/**
 * Dangerous path patterns for security validation
 */
export const DANGEROUS_PATH_PATTERNS = [
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
] as const;

/**
 * Dangerous project name patterns
 */
export const DANGEROUS_PROJECT_PATTERNS = [
  /\.\./, // Path traversal
  /^[.-]/, // Starting with dot or dash
  /[<>:"|?*]/, // File system reserved chars
  /[\x00-\x1f]/, // Control characters
  /^(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])$/i, // Windows reserved names
] as const;

/**
 * Dangerous string content patterns
 */
export const DANGEROUS_STRING_PATTERNS = [
  /javascript:/i,
  /eval\s*\(/i,
  /function\s*\(/i,
  /setTimeout\s*\(/i,
  /setInterval\s*\(/i,
  /<script/i,
  /on\w+\s*=/i, // Event handlers
  /\${.*}/, // Template expressions
  /`.*`/, // Template literals
] as const;

/**
 * Git branch invalid patterns
 */
export const INVALID_BRANCH_PATTERNS = [
  /\.\./, // Double dots
  /^[.-]/, // Starting with dot or dash
  /[.-]$/, // Ending with dot or dash
  /[\x00-\x1f\x7f]/, // Control characters
  /[~^:?*[\]\\]/, // Git-reserved characters
  /\s/, // Whitespace
  /@{/, // @{ sequence
  /\/$/, // Ending with slash
  /\/\//, // Double slashes
] as const;

/**
 * Git reserved branch names
 */
export const RESERVED_BRANCH_NAMES = ['HEAD', 'master', 'origin'] as const;

/**
 * Project name special characters pattern
 */
export const SPECIAL_CHARS_PATTERN = /[!@#$%^&*(),.?":{}|<>]/;

/**
 * Valid option key pattern
 */
export const VALID_OPTION_KEY_PATTERN = /^[a-zA-Z][a-zA-Z0-9_-]*$/;

/**
 * Allowed template types
 */
export const ALLOWED_TEMPLATE_TYPES = [
  'standard',
  'minimal',
  'advanced',
  'custom',
] as const;

/**
 * Path validation constants
 */
export const PATH_VALIDATION_CONSTANTS = {
  MAX_PATH_LENGTH: 260, // Windows MAX_PATH limitation
  DEFAULT_MAX_DEPTH: 10,
} as const;

/**
 * Project name validation constants
 */
export const PROJECT_NAME_CONSTANTS = {
  DEFAULT_MIN_LENGTH: 1,
  DEFAULT_MAX_LENGTH: 50,
} as const;

/**
 * Branch name validation constants
 */
export const BRANCH_NAME_CONSTANTS = {
  MIN_LENGTH: 1,
  MAX_LENGTH: 250,
} as const;

/**
 * Command options validation constants
 */
export const COMMAND_OPTIONS_CONSTANTS = {
  MAX_STRING_LENGTH: 1000,
} as const;

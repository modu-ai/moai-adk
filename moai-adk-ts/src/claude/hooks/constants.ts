/**
 * @CODE:HOOKS-REFACTOR-001 |
 * SPEC: ../.moai/specs/SPEC-HOOKS-REFACTOR-001/spec.md |
 * TEST: __tests__/constants.test.ts
 *
 * Hook System Constants
 * 훅 시스템에서 사용하는 모든 상수 중앙 관리
 */

/**
 * Supported programming languages and their file extensions
 */
export const SUPPORTED_LANGUAGES = {
  typescript: ['.ts', '.tsx'],
  javascript: ['.js', '.jsx', '.mjs', '.cjs'],
  python: ['.py', '.pyi'],
  java: ['.java'],
  go: ['.go'],
  rust: ['.rs'],
  cpp: ['.cpp', '.hpp', '.cc', '.h', '.cxx', '.hxx'],
  ruby: ['.rb', '.rake', '.gemspec'],
  php: ['.php'],
  csharp: ['.cs'],
  dart: ['.dart'],
  swift: ['.swift'],
  kotlin: ['.kt', '.kts'],
  elixir: ['.ex', '.exs'],
  markdown: ['.md', '.mdx'],
} as const;

/**
 * Read-only tools that bypass policy checks
 */
export const READ_ONLY_TOOLS = [
  'Read',
  'Glob',
  'Grep',
  'WebFetch',
  'WebSearch',
  'TodoWrite',
  'BashOutput',
  'mcp__context7__resolve-library-id',
  'mcp__context7__get-library-docs',
  'mcp__ide__getDiagnostics',
  'mcp__ide__executeCode',
] as const;

/**
 * Timeout and threshold constants
 */
export const TIMEOUTS = {
  TAG_BLOCK_SEARCH_LIMIT: 30,
  GIT_COMMAND_TIMEOUT: 2000,
  NPM_REGISTRY_TIMEOUT: 2000,
  POLICY_BLOCK_SLOW_THRESHOLD: 100,
} as const;

/**
 * Dangerous commands that should always be blocked
 */
export const DANGEROUS_COMMANDS = [
  'rm -rf /',
  'rm -rf --no-preserve-root',
  'sudo rm',
  'dd if=/dev/zero',
  ':(){:|:&};:',
  'mkfs.',
] as const;

/**
 * Paths that should be excluded from TAG enforcement
 */
export const EXCLUDED_PATHS = [
  'node_modules',
  '.git',
  'dist',
  'build',
  'test',
  'spec',
  '__test__',
  '__tests__',
] as const;

/**
 * Sensitive file patterns to protect
 */
export const SENSITIVE_KEYWORDS = [
  '.env',
  '/secrets',
  '/.git/',
  '/.ssh',
] as const;

/**
 * Protected paths that should not be modified
 */
export const PROTECTED_PATHS = ['.moai/memory/'] as const;

/**
 * Command prefixes that are allowed
 */
export const ALLOWED_PREFIXES = [
  'git ',
  'python',
  'pytest',
  'npm ',
  'node ',
  'go ',
  'cargo ',
  'poetry ',
  'pnpm ',
  'rg ',
  'ls ',
  'cat ',
  'echo ',
  'which ',
  'make ',
  'moai ',
  'tsx ',
  'moai-ts ',
  'npx ',
  'tsc ',
  'jest ',
  'ts-node ',
  'alfred ',
  'bun ',
] as const;

/**
 * @CODE:BUILD-001 | Chain: @SPEC:BUILD-001 -> @SPEC:BUILD-001 -> @CODE:BUILD-001 -> @TEST:BUILD-001
 * Related: @CODE:BUILD-001:API
 *
 * TSUP Configuration for Claude Code Hooks
 * templates/.claude/hooks/alfred/ts/*.ts → templates/.claude/hooks/alfred/*.cjs
 */

import { defineConfig } from 'tsup';

export default defineConfig({
  entry: {
    'policy-block': 'src/claude/hooks/policy-block.ts',
    'pre-write-guard': 'src/claude/hooks/pre-write-guard.ts',
    'session-notice': 'src/claude/hooks/session-notice/index.ts',
    'tag-enforcer': 'src/claude/hooks/tag-enforcer.ts'
  },
  format: ['cjs'],
  target: 'node18',
  outDir: 'templates/.claude/hooks/alfred',
  outExtension: () => ({ js: '.cjs' }),
  sourcemap: false,
  clean: false, // CJS 파일 보존
  dts: false,
  splitting: false,
  bundle: true,
  minify: false, // 디버깅 용이성을 위해 minify 비활성화
  treeshake: true,
  external: [
    // Node.js 내장 모듈
    'fs',
    'fs/promises',
    'path',
    'child_process',
    'os'
  ],
  esbuildOptions(options) {
    options.banner = {
      js: '// Auto-generated from TypeScript source - DO NOT EDIT DIRECTLY'
    };
  }
});

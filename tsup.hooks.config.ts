/**
 * @CODE:BUILD-HOOKS-001 | SPEC: SPEC-BUILD-001.md
 *
 * TSUP Configuration for MoAI-ADK Claude Code Hooks
 * src/hooks/*.ts → .claude-plugin/hooks/scripts/*.cjs
 */

import { defineConfig } from 'tsup';

export default defineConfig({
  entry: {
    'policy-block': 'src/hooks/policy-block/index.ts',
    'pre-write-guard': 'src/hooks/pre-write-guard/index.ts',
    'session-notice': 'src/hooks/session-notice/index.ts',
    'tag-enforcer': 'src/hooks/tag-enforcer/index.ts'
  },
  format: ['cjs'],
  target: 'node18',
  outDir: '.claude-plugin/hooks/scripts',
  outExtension: () => ({ js: '.cjs' }),
  sourcemap: false,
  clean: false, // 기존 CJS 파일 보존
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

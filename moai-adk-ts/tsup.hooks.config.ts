import { defineConfig } from 'tsup';

export default defineConfig({
  entry: {
    'claude/hooks/security/pre-write-guard': 'src/claude/hooks/security/pre-write-guard.ts',
    'claude/hooks/security/policy-block': 'src/claude/hooks/security/policy-block.ts',
    'claude/hooks/security/steering-guard': 'src/claude/hooks/security/steering-guard.ts',
    'claude/hooks/session/session-notice': 'src/claude/hooks/session/session-notice.ts',
    'claude/hooks/workflow/file-monitor': 'src/claude/hooks/workflow/file-monitor.ts',
    'claude/hooks/workflow/language-detector': 'src/claude/hooks/workflow/language-detector.ts',
  },
  format: ['cjs'],
  target: 'node18',
  outDir: 'dist',
  clean: false, // main build already cleans
  sourcemap: false, // hooks don't need sourcemaps
  minify: false, // keep readable for debugging
  splitting: false, // each hook should be standalone
  bundle: true, // bundle dependencies for standalone execution
  external: [], // don't externalize anything for hooks
  platform: 'node',
  shims: true, // add import.meta.url and __dirname shims
});
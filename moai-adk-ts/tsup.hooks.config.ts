import { defineConfig } from 'tsup';

export default defineConfig({
  entry: {
    'session-notice': 'src/claude/hooks/session/session-notice.ts',
  },
  format: ['cjs'],
  target: 'node18',
  outDir: '../.claude/hooks/moai',
  outExtension: () => ({ js: '.cjs' }), // Use .cjs extension for CommonJS in ES module project
  sourcemap: false,
  clean: false,
  dts: false,
  splitting: false,
  bundle: true,
  minify: false,
  treeshake: true,
  external: [
    'fs',
    'path',
    'util',
    'child_process',
  ],
});
import { defineConfig } from 'tsup';

export default defineConfig({
  entry: {
    'scripts/commit-helper': 'src/scripts/commit-helper.ts',
    'scripts/trust-principles-checker': 'src/scripts/trust-principles-checker.ts'
  },
  format: ['cjs'],
  target: 'node18',
  outDir: 'dist',
  sourcemap: false,
  clean: false, // Don't clean since we might have other builds
  dts: false,
  splitting: false,
  bundle: true, // Bundle dependencies for standalone execution
  minify: false,
  treeshake: true,
  banner: {
    js: '#!/usr/bin/env node'
  },
  external: [
    // Bundle most dependencies except these system-level ones
    'fs',
    'path',
    'util',
    'child_process'
  ]
});
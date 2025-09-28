import { defineConfig } from 'tsup';

export default defineConfig({
  entry: [
    'src/cli/index.ts',
    'src/index.ts'
  ],
  format: ['esm'],
  dts: true,
  splitting: false,
  sourcemap: true,
  clean: true,
  target: 'node18',
  outDir: 'dist',
  shims: true,
  banner: {
    js: '#!/usr/bin/env node'
  }
});
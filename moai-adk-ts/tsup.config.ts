import { defineConfig } from 'tsup';

export default defineConfig({
  entry: {
    'index': 'src/index.ts',
    'cli/index': 'src/cli/index.ts',
    'scripts/sync-analyzer': 'src/scripts/sync-analyzer.ts'
  },
  format: ['cjs', 'esm'],
  target: 'node18',
  outDir: 'dist',
  sourcemap: true,
  clean: true,
  dts: true,
  splitting: false,
  bundle: true,
  minify: process.env.NODE_ENV === 'production',
  treeshake: true,
  external: [
    'chalk',
    'commander',
    'inquirer',
    'semver',
    'execa',
    'fs-extra',
    'simple-git',
    'yaml',
    'mustache',
    'mime-types',
    'chokidar'
  ]
});
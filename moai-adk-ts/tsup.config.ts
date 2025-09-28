import { defineConfig } from 'tsup';

export default defineConfig({
  entry: {
    'index': 'src/index.ts',
    'cli/index': 'src/cli/index.ts'
  },
  format: ['cjs', 'esm'],
  target: 'node18',
  outDir: 'dist',
  sourcemap: true,
  clean: true,
  dts: true,
  splitting: false,
  bundle: true,
  minify: false,
  treeshake: true,
  external: [
    'chalk',
    'commander',
    'inquirer',
    'semver',
    'execa'
  ]
});
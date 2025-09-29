/// <reference types="vitest" />
import { defineConfig } from 'vitest/config';
import tsconfigPaths from 'vite-tsconfig-paths';
import path from 'path';

export default defineConfig({
  plugins: [tsconfigPaths()],

  test: {
    // Environment
    environment: 'node',

    // Test files
    include: [
      'src/**/*.{test,spec}.ts',
      '__tests__/**/*.{test,spec}.ts'
    ],
    exclude: [
      'node_modules',
      'dist',
      '**/*setup*',
      '__tests__/_disabled/**',
      // 누락된 소스 파일이 있는 테스트들 임시 비활성화
      'src/__tests__/claude/**',
      'src/scripts/__tests__/**',
      'src/__tests__/cli/commands/doctor-advanced.test.ts'
    ],

    // Timeout
    testTimeout: 10000,

    // Coverage
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov', 'html'],
      reportsDirectory: 'coverage',
      include: ['src/**/*.ts'],
      exclude: [
        'src/**/*.d.ts',
        'src/**/*.test.ts',
        'src/**/*.spec.ts'
      ],
      thresholds: {
        branches: 80,
        functions: 80,
        lines: 80,
        statements: 80
      }
    },

    // Setup
    setupFiles: ['src/__tests__/setup.ts'],

    // Globals (enables Jest-like global functions)
    globals: true,

    // Types
    types: ['vitest/globals'],

    // Pool options for better performance
    pool: 'threads',
    poolOptions: {
      threads: {
        singleThread: false
      }
    }
  },

  // Path resolution
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
      '@/cli': path.resolve(__dirname, 'src/cli'),
      '@/core': path.resolve(__dirname, 'src/core'),
      '@/utils': path.resolve(__dirname, 'src/utils'),
      '@/types': path.resolve(__dirname, 'src/types')
    }
  },

  // ESM support
  esbuild: {
    target: 'node18'
  }
});
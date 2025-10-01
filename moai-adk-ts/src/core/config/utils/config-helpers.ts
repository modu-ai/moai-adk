// @CODE:CFG-HELPERS-001 |
// Related: @CODE:CFG-001

/**
 * @file Configuration helper utilities
 * @author MoAI Team
 */

import type { ProjectConfigInput } from '../types';

/**
 * Get enabled agents based on project mode
 */
export function getEnabledAgents(mode: string): string[] {
  const baseAgents = ['spec-builder', 'code-builder', 'doc-syncer'];
  if (mode === 'team') {
    return [...baseAgents, 'git-manager', 'debug-helper'];
  }
  return baseAgents;
}

/**
 * Get enabled commands based on project mode
 */
export function getEnabledCommands(mode: string): string[] {
  const baseCommands = [
    '/moai:8-project',
    '/moai:1-spec',
    '/moai:2-build',
    '/moai:3-sync',
  ];
  if (mode === 'team') {
    return [...baseCommands, '/moai:4-debug'];
  }
  return baseCommands;
}

/**
 * Get command shortcuts mapping
 */
export function getCommandShortcuts(): Record<string, string> {
  return {
    spec: '/moai:1-spec',
    build: '/moai:2-build',
    sync: '/moai:3-sync',
  };
}

/**
 * Get enabled hooks based on project mode
 */
export function getEnabledHooks(mode: string): string[] {
  const baseHooks = ['steering-guard', 'pre-write-guard', 'file-monitor'];
  if (mode === 'team') {
    return [...baseHooks, 'test-runner', 'policy-block'];
  }
  return baseHooks;
}

/**
 * Get hook configuration based on project config
 */
export function getHookConfiguration(
  config: ProjectConfigInput
): Record<string, any> {
  return {
    'steering-guard': {
      enabled: true,
      checkPatterns: ['*.ts', '*.js', '*.py'],
    },
    'file-monitor': {
      watchPaths: ['src/', 'tests/'],
      extensions: config.techStack.includes('typescript')
        ? ['.ts', '.tsx']
        : ['.js', '.jsx'],
    },
  };
}

/**
 * Get package.json scripts based on project config
 */
export function getPackageScripts(
  config: ProjectConfigInput
): Record<string, string> {
  const scripts: Record<string, string> = {
    build: 'tsc',
    dev: 'tsx src/index.ts',
    test: 'jest',
    lint: 'eslint src/',
    format: 'prettier --write src/',
  };

  if (config.techStack.includes('typescript')) {
    scripts['type-check'] = 'tsc --noEmit';
  }

  if (config.techStack.includes('react')) {
    scripts.dev = 'vite';
    scripts.build = 'vite build';
  }

  return scripts;
}

/**
 * Get package.json dependencies based on tech stack
 */
export function getPackageDependencies(
  config: ProjectConfigInput
): Record<string, string> {
  const deps: Record<string, string> = {};

  if (config.techStack.includes('react')) {
    deps.react = '^18.0.0';
    deps['react-dom'] = '^18.0.0';
  }

  if (config.techStack.includes('nextjs')) {
    deps.next = '^13.0.0';
  }

  if (config.techStack.includes('express')) {
    deps.express = '^4.18.0';
  }

  return deps;
}

/**
 * Get package.json devDependencies based on tech stack
 */
export function getPackageDevDependencies(
  config: ProjectConfigInput
): Record<string, string> {
  const devDeps: Record<string, string> = {
    jest: '^29.0.0',
    eslint: '^8.0.0',
    prettier: '^3.0.0',
  };

  if (config.techStack.includes('typescript')) {
    devDeps.typescript = '^5.0.0';
    devDeps['@types/node'] = '^20.0.0';
    devDeps.tsx = '^4.0.0';
  }

  if (config.techStack.includes('react')) {
    devDeps['@types/react'] = '^18.0.0';
    devDeps['@types/react-dom'] = '^18.0.0';
    devDeps.vite = '^4.0.0';
  }

  return devDeps;
}

// @TEST:TAG-MIGRATION-001
import { test, expect } from 'vitest';
import { readFileSync, existsSync } from 'fs';
import { execSync } from 'child_process';
import { join } from 'path';

// Test 1: package.json should have bun as packageManager
test('package.json should have bun as packageManager', () => {
  const packageJson = JSON.parse(readFileSync(join('docs', 'package.json'), 'utf-8'));
  expect(packageJson.packageManager).toBe('bun@1.1.0');
});

// Test 2: biome.json configuration should exist
test('biome.json configuration should exist', () => {
  expect(existsSync(join('docs', 'biome.json'))).toBe(true);
});

// Test 3: biome should be in devDependencies
test('biome should be in devDependencies', () => {
  const packageJson = JSON.parse(readFileSync(join('docs', 'package.json'), 'utf-8'));
  expect(packageJson.devDependencies).toBeDefined();
  expect(packageJson.devDependencies.biome).toBeDefined();
});

// Test 4: bun install should succeed
test('bun install should succeed', () => {
  try {
    execSync('cd docs && bun install', { stdio: 'pipe', timeout: 30000 });
    expect(true).toBe(true);
  } catch (error) {
    throw new Error(`bun install failed: ${error.message}`);
  }
});

// Test 5: bun biome check should pass
test('bun biome check should pass', () => {
  try {
    execSync('cd docs && bun x @biomejs/biome check .', { stdio: 'pipe', timeout: 60000 });
    expect(true).toBe(true);
  } catch (error) {
    throw new Error(`bun biome check failed: ${error.message}`);
  }
});

// Test 6: bun next build should work
test('bun next build should work', () => {
  try {
    execSync('cd docs && bun run build', { stdio: 'pipe', timeout: 120000 });
    expect(true).toBe(true);
  } catch (error) {
    throw new Error(`bun run build failed: ${error.message}`);
  }
});

// Test 7: TypeScript compilation with bun should work
test('TypeScript compilation with bun should work', () => {
  try {
    execSync('cd docs && bun run type-check', { stdio: 'pipe', timeout: 30000 });
    expect(true).toBe(true);
  } catch (error) {
    throw new Error(`bun run type-check failed: ${error.message}`);
  }
});

// Test 8: Scripts should use bun commands
test('scripts should use bun commands where appropriate', () => {
  const packageJson = JSON.parse(readFileSync(join('docs', 'package.json'), 'utf-8'));

  // CI script should use bun commands
  if (packageJson.scripts.ci) {
    expect(packageJson.scripts.ci).toContain('bun run');
  }

  // Check that npm references are removed
  Object.values(packageJson.scripts).forEach(script => {
    if (typeof script === 'string') {
      expect(script).not.toContain('npm run');
      expect(script).not.toContain('npm ');
    }
  });
});
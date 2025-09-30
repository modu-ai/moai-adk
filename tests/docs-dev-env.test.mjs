import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
import path from 'node:path';

import { ensureDevEnv, buildVitepressArgs } from '../scripts/docs-dev.mjs';

const require = createRequire(import.meta.url);

const resetEnv = (previousValue) => {
  if (typeof previousValue === 'undefined') {
    delete process.env.NODE_ENV;
  } else {
    process.env.NODE_ENV = previousValue;
  }
};

test('ensureDevEnv sets NODE_ENV to development when missing', () => {
  const input = {};
  const result = ensureDevEnv(input);

  assert.equal(result.NODE_ENV, 'development');
  assert.equal(typeof input.NODE_ENV, 'undefined');
});

test('ensureDevEnv keeps existing keys intact', () => {
  const input = { FOO: 'bar' };
  const result = ensureDevEnv(input);

  assert.equal(result.FOO, 'bar');
});

test('ensureDevEnv forces process env to development', () => {
  const previousNodeEnv = process.env.NODE_ENV;
  process.env.NODE_ENV = 'production';

  ensureDevEnv(process.env);

  assert.equal(process.env.NODE_ENV, 'development');
  resetEnv(previousNodeEnv);
});

test('buildVitepressArgs prefixes CLI invocation and preserves flags', () => {
  const args = buildVitepressArgs(['--host', '0.0.0.0']);
  const packageJsonPath = require.resolve('vitepress/package.json');
  const expectedBin = path.join(path.dirname(packageJsonPath), 'bin', 'vitepress.js');

  assert.deepEqual(args.slice(0, 3), [expectedBin, 'dev', 'docs']);
  assert.deepEqual(args.slice(3), ['--host', '0.0.0.0']);
});

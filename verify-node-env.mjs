#!/usr/bin/env node
/**
 * Verify that NODE_ENV is correctly set to 'development'
 * by the docs-dev.mjs launcher
 */

import { spawn } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const projectRoot = path.dirname(fileURLToPath(import.meta.url));
const scriptPath = path.join(projectRoot, 'scripts', 'docs-dev.mjs');

console.log('🧪 Verifying NODE_ENV setting via docs-dev.mjs launcher\n');

// Start the launcher with a modified script that prints NODE_ENV and exits
const testScript = `
import { ensureDevEnv } from '${scriptPath}';
const env = ensureDevEnv(process.env);
console.log('NODE_ENV=' + env.NODE_ENV);
process.exit(0);
`;

const child = spawn('node', ['--input-type=module', '-e', testScript], {
  stdio: ['pipe', 'pipe', 'pipe'],
  env: { ...process.env, NODE_ENV: 'production' } // Start with production to test override
});

let output = '';

child.stdout.on('data', (data) => {
  output += data.toString();
});

child.stderr.on('data', (data) => {
  console.error('stderr:', data.toString());
});

child.on('close', (code) => {
  if (code === 0) {
    const match = output.match(/NODE_ENV=(\w+)/);
    if (match && match[1] === 'development') {
      console.log('✅ SUCCESS: NODE_ENV is correctly set to "development"');
      console.log('✅ The launcher correctly overrides any existing NODE_ENV value');
      console.log('\n📝 Original NODE_ENV was "production", launcher forced "development"');
      process.exit(0);
    } else {
      console.log('❌ FAILURE: NODE_ENV is not "development"');
      console.log('Output:', output);
      process.exit(1);
    }
  } else {
    console.log('❌ Test script failed with code:', code);
    process.exit(1);
  }
});
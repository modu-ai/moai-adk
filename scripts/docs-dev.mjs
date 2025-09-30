#!/usr/bin/env node
import { spawn } from 'node:child_process';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import process from 'node:process';

const require = createRequire(import.meta.url);
const entryPath = fileURLToPath(import.meta.url);

export function ensureDevEnv(baseEnv = process.env) {
  const nextEnv = { ...baseEnv, NODE_ENV: 'development' };
  if (baseEnv === process.env) {
    process.env.NODE_ENV = 'development';
  }
  return nextEnv;
}

function resolveVitepressBin() {
  const packageJsonPath = require.resolve('vitepress/package.json');
  return path.join(path.dirname(packageJsonPath), 'bin', 'vitepress.js');
}

export function buildVitepressArgs(argv = process.argv.slice(2)) {
  return [resolveVitepressBin(), 'dev', 'docs', ...argv];
}

export function run() {
  const child = spawn(process.execPath, buildVitepressArgs(), {
    stdio: 'inherit',
    env: ensureDevEnv(process.env)
  });

  child.on('close', (code, signal) => {
    if (signal) {
      process.kill(process.pid, signal);
      return;
    }
    process.exit(code ?? 0);
  });

  child.on('error', (error) => {
    console.error('VitePress 개발 서버 실행에 실패했습니다.', error);
    process.exit(1);
  });
}

if (process.argv[1] && path.resolve(process.argv[1]) === entryPath) {
  run();
}

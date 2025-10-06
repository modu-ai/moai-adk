// @TEST:DOCS-001 | SPEC: .moai/specs/SPEC-DOCS-001/spec.md
// VitePress 빌드 테스트 - Phase 1 (5개 핵심 페이지)

import { describe, test, expect } from 'vitest';
import { execa } from 'execa';
import { existsSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const docsRoot = join(__dirname, '../../../docs');

describe('@TEST:DOCS-001 - VitePress Build Tests', () => {
  test('should have VitePress config file', () => {
    const configPath = join(docsRoot, '.vitepress/config.mts');

    // GREEN: config.mts 파일 존재 확인
    expect(existsSync(configPath)).toBe(true);
  });

  test('should build VitePress successfully', async () => {
    try {
      // RED: VitePress config 없어 빌드 실패 예상
      const { exitCode } = await execa('bun', ['run', 'docs:build'], {
        cwd: join(__dirname, '../../'),
        timeout: 60000,
      });

      expect(exitCode).toBe(0);
    } catch (error) {
      // 현재 예상되는 에러: config.ts 없음
      throw new Error(`VitePress build failed: ${error}`);
    }
  }, 60000);

  test('should have no build errors', async () => {
    try {
      const { stdout } = await execa('bun', ['run', 'docs:build'], {
        cwd: join(__dirname, '../../'),
        timeout: 60000,
      });

      // GREEN: VitePress 빌드 성공 확인
      expect(stdout).toContain('build complete');
    } catch (error) {
      // 빌드 실패 시 에러 메시지 확인
      throw error;
    }
  }, 60000);
});

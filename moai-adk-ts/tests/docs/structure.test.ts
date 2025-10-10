// @TEST:DOCS-001 | SPEC: SPEC-DOCS-001.md
/**
 * 문서 구조 재편 테스트
 *
 * 검증 항목:
 * 1. 9개 카테고리 디렉토리 존재
 * 2. 각 카테고리 index.md 존재
 * 3. VitePress 제거 확인
 * 4. README.md 크기 제한 (< 5KB)
 * 5. Jekyll 설정 존재
 */

import { describe, it, expect } from 'vitest';
import { existsSync, statSync } from 'fs';
import { resolve } from 'path';

const PROJECT_ROOT = resolve(__dirname, '../..');
const DOCS_ROOT = resolve(PROJECT_ROOT, '../docs');

describe('@TEST:DOCS-001 - 문서 구조 재편', () => {
  describe('VitePress 제거 검증', () => {
    it('.vitepress 디렉토리가 존재하지 않아야 함', () => {
      const vitepressDir = resolve(DOCS_ROOT, '.vitepress');
      expect(existsSync(vitepressDir)).toBe(false);
    });

    it('package.json에 vitepress 의존성이 없어야 함', () => {
      const packageJsonPath = resolve(PROJECT_ROOT, 'package.json');
      const packageJson = JSON.parse(
        require('fs').readFileSync(packageJsonPath, 'utf-8')
      );

      const allDeps = {
        ...packageJson.dependencies,
        ...packageJson.devDependencies,
      };

      expect(allDeps).not.toHaveProperty('vitepress');
      expect(allDeps).not.toHaveProperty('vue');
    });

    it('package.json에 VitePress 스크립트가 없어야 함 (docs:api 제외)', () => {
      const packageJsonPath = resolve(PROJECT_ROOT, 'package.json');
      const packageJson = JSON.parse(
        require('fs').readFileSync(packageJsonPath, 'utf-8')
      );

      const { scripts } = packageJson;

      // docs:api는 typedoc만 사용하므로 허용
      expect(scripts['docs:dev']).toBeUndefined();
      expect(scripts['docs:build']).toBeUndefined();
      expect(scripts['docs:preview']).toBeUndefined();
    });
  });

  describe('9개 카테고리 디렉토리 검증', () => {
    const categories = [
      'getting-started',
      'concepts',
      'alfred',
      'cli',
      'api',
      'guides',
      'agents',
      'examples',
      'contributing',
    ];

    categories.forEach((category) => {
      it(`${category} 디렉토리가 존재해야 함`, () => {
        const categoryDir = resolve(DOCS_ROOT, category);
        expect(existsSync(categoryDir)).toBe(true);
      });

      it(`${category}/index.md 파일이 존재해야 함`, () => {
        const indexFile = resolve(DOCS_ROOT, category, 'index.md');
        expect(existsSync(indexFile)).toBe(true);
      });
    });
  });

  describe('메인 허브 검증', () => {
    it('docs/index.md 파일이 존재해야 함', () => {
      const mainHub = resolve(DOCS_ROOT, 'index.md');
      expect(existsSync(mainHub)).toBe(true);
    });

    it('docs/index.md에 9개 카테고리 링크가 포함되어야 함', () => {
      const mainHub = resolve(DOCS_ROOT, 'index.md');
      const content = require('fs').readFileSync(mainHub, 'utf-8');

      const categories = [
        'getting-started',
        'concepts',
        'alfred',
        'cli',
        'api',
        'guides',
        'agents',
        'examples',
        'contributing',
      ];

      categories.forEach((category) => {
        expect(content).toContain(category);
      });
    });
  });

  describe('Jekyll 설정 검증', () => {
    it('docs/_config.yml 파일이 존재해야 함', () => {
      const configFile = resolve(DOCS_ROOT, '_config.yml');
      expect(existsSync(configFile)).toBe(true);
    });

    it('_config.yml에 필수 설정이 포함되어야 함', () => {
      const configFile = resolve(DOCS_ROOT, '_config.yml');
      const content = require('fs').readFileSync(configFile, 'utf-8');

      expect(content).toContain('theme:');
      expect(content).toContain('title:');
      expect(content).toContain('description:');
    });
  });

  describe('README.md 검증', () => {
    it('README.md 파일이 존재해야 함', () => {
      const readmePath = resolve(PROJECT_ROOT, '../README.md');
      expect(existsSync(readmePath)).toBe(true);
    });

    it('README.md 크기가 5KB 이하여야 함', () => {
      const readmePath = resolve(PROJECT_ROOT, '../README.md');
      const stats = statSync(readmePath);
      const sizeInKB = stats.size / 1024;

      expect(sizeInKB).toBeLessThan(5);
    });

    it('README.md에 docs 링크가 포함되어야 함', () => {
      const readmePath = resolve(PROJECT_ROOT, '../README.md');
      const content = require('fs').readFileSync(readmePath, 'utf-8');

      expect(content).toContain('docs/');
    });
  });
});

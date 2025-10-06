// @TEST:DOCS-001 | SPEC: .moai/specs/SPEC-DOCS-001/spec.md
// 링크 검증 테스트 - Phase 1 (5개 핵심 페이지)

import { describe, test, expect } from 'vitest';
import { existsSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const docsRoot = join(__dirname, '../../../docs');

describe('@TEST:DOCS-001 - Link Validation Tests (Phase 1)', () => {
  test('should have index.md homepage', () => {
    const indexPath = join(docsRoot, 'index.md');

    // RED: index.md 파일이 아직 없어 실패 예상
    expect(existsSync(indexPath)).toBe(true);
  });

  test('should have getting-started.md guide', () => {
    const gettingStartedPath = join(docsRoot, 'guide/getting-started.md');

    // RED: guide/getting-started.md 파일이 아직 없어 실패 예상
    expect(existsSync(gettingStartedPath)).toBe(true);
  });

  test('should have what-is-moai-adk.md guide', () => {
    const whatIsPath = join(docsRoot, 'guide/what-is-moai-adk.md');

    // RED: guide/what-is-moai-adk.md 파일이 아직 없어 실패 예상
    expect(existsSync(whatIsPath)).toBe(true);
  });

  test('should have spec-first-tdd.md concept', () => {
    const specFirstPath = join(docsRoot, 'concepts/spec-first-tdd.md');

    // RED: concepts/spec-first-tdd.md 파일이 아직 없어 실패 예상
    expect(existsSync(specFirstPath)).toBe(true);
  });

  test('should have faq.md guide', () => {
    const faqPath = join(docsRoot, 'guide/faq.md');

    // RED: guide/faq.md 파일이 아직 없어 실패 예상
    expect(existsSync(faqPath)).toBe(true);
  });

  test('should have all Phase 1 pages (5 files)', () => {
    const phase1Pages = [
      join(docsRoot, 'index.md'),
      join(docsRoot, 'guide/getting-started.md'),
      join(docsRoot, 'guide/what-is-moai-adk.md'),
      join(docsRoot, 'concepts/spec-first-tdd.md'),
      join(docsRoot, 'guide/faq.md'),
    ];

    const existingPages = phase1Pages.filter(page => existsSync(page));

    // RED: 5개 페이지 모두 없어 실패 예상
    expect(existingPages.length).toBe(5);
  });
});

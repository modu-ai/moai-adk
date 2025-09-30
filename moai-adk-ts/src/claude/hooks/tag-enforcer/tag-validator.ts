// @CODE:REFACTOR-003 | SPEC: SPEC-REFACTOR-003.md | TEST: __tests__/claude/hooks/tag-enforcer/tag-validator.test.ts

/**
 * TAG Validator Module
 * @CODE:REFACTOR-003:DOMAIN - TAG 블록 검증 및 추출 로직
 *
 * SPEC-REFACTOR-003 요구사항:
 * - TAG 블록 추출 (파일 최상단)
 * - TAG 정규화
 * - TAG 유효성 검증
 */

import { CODE_FIRST_PATTERNS, VALID_CATEGORIES } from './tag-patterns';
import type { TagBlock, ValidationResult } from './types';

/**
 * TAG 검증 클래스
 */
export class TagValidator {
  /**
   * TAG 블록 추출 (파일 최상단에서만)
   */
  extractTagBlock(content: string): TagBlock | null {
    const lines = content.split('\n');
    let inBlock = false;
    let blockLines: string[] = [];
    let startLineNumber = 0;

    for (let i = 0; i < Math.min(lines.length, 30); i++) {
      const line = lines[i].trim();

      // 빈 줄이나 shebang 무시
      if (!line || line.startsWith('#!')) {
        continue;
      }

      // TAG 블록 시작
      if (line.startsWith('/**') && !inBlock) {
        inBlock = true;
        blockLines = [line];
        startLineNumber = i + 1;
        continue;
      }

      // TAG 블록 내부
      if (inBlock) {
        blockLines.push(line);

        // TAG 블록 종료
        if (line.endsWith('*/')) {
          const blockContent = blockLines.join('\n');

          // @TAG가 포함된 블록인지 확인
          if (CODE_FIRST_PATTERNS.MAIN_TAG.test(blockContent)) {
            return {
              content: blockContent,
              lineNumber: startLineNumber
            };
          }

          // @TAG가 없으면 리셋하고 계속
          inBlock = false;
          blockLines = [];
          continue;
        }
      }

      // TAG 블록이 아닌 코드 시작되면 중단
      if (!inBlock && line && !line.startsWith('//') && !line.startsWith('/*')) {
        break;
      }
    }

    return null;
  }

  /**
   * TAG 블록에서 메인 TAG 추출
   */
  extractMainTag(blockContent: string): string {
    const match = CODE_FIRST_PATTERNS.MAIN_TAG.exec(blockContent);
    return match ? `@${match[1]}:${match[2]}` : 'UNKNOWN';
  }

  /**
   * TAG 블록 정규화 (비교용)
   */
  normalizeTagBlock(blockContent: string): string {
    return blockContent
      .split('\n')
      .map(line => line.trim())
      .filter(line => line.length > 0)
      .join('\n');
  }

  /**
   * Code-First TAG 유효성 검증
   */
  validateCodeFirstTag(content: string): ValidationResult {
    const violations: string[] = [];
    const warnings: string[] = [];
    let hasTag = false;

    // 1. TAG 블록 추출
    const tagBlock = this.extractTagBlock(content);
    if (!tagBlock) {
      return {
        isValid: true, // TAG 블록이 없어도 차단하지 않음 (권장사항)
        violations: [],
        warnings: ['파일 최상단에 TAG 블록이 없습니다 (권장사항)'],
        hasTag: false
      };
    }

    hasTag = true;
    const blockContent = tagBlock.content;

    // 2. 메인 TAG 검증
    const tagMatch = CODE_FIRST_PATTERNS.MAIN_TAG.exec(blockContent);
    if (!tagMatch) {
      violations.push('@TAG 라인이 발견되지 않았습니다');
    } else {
      const [, category, domainId] = tagMatch;

      // 카테고리 유효성 검사
      const allValidCategories = [...VALID_CATEGORIES.lifecycle, ...VALID_CATEGORIES.implementation];
      if (!allValidCategories.includes(category)) {
        violations.push(`유효하지 않은 TAG 카테고리: ${category}`);
      }

      // 도메인 ID 형식 검사 (하이픈 권장, 언더스코어 사용 시 경고)
      if (!/^[A-Z0-9]+-\d{3,}$/.test(domainId)) {
        // 하이픈이 아닌 다른 구분자를 사용하거나 형식이 올바르지 않으면 경고
        warnings.push(`도메인 ID 형식 권장: ${domainId} -> DOMAIN-001`);
      }
    }

    // 3. 체인 검증
    const chainMatch = CODE_FIRST_PATTERNS.CHAIN_LINE.exec(blockContent);
    if (chainMatch) {
      const chainStr = chainMatch[1];
      const chainTags = chainStr.split(/\s*->\s*/);

      for (const chainTag of chainTags) {
        if (!CODE_FIRST_PATTERNS.TAG_REFERENCE.test(chainTag.trim())) {
          warnings.push(`체인의 TAG 형식을 확인하세요: ${chainTag.trim()}`);
        }
      }
    }

    // 4. 의존성 검증
    const dependsMatch = CODE_FIRST_PATTERNS.DEPENDS_LINE.exec(blockContent);
    if (dependsMatch) {
      const dependsStr = dependsMatch[1];
      if (dependsStr.trim().toLowerCase() !== 'none') {
        const dependsTags = dependsStr.split(/,\s*/);

        for (const dependTag of dependsTags) {
          if (!CODE_FIRST_PATTERNS.TAG_REFERENCE.test(dependTag.trim())) {
            warnings.push(`의존성 TAG 형식을 확인하세요: ${dependTag.trim()}`);
          }
        }
      }
    }

    // 5. 상태 검증
    const statusMatch = CODE_FIRST_PATTERNS.STATUS_LINE.exec(blockContent);
    if (statusMatch) {
      const status = statusMatch[1].toLowerCase();
      if (!['active', 'deprecated', 'completed'].includes(status)) {
        warnings.push(`알 수 없는 STATUS: ${status}`);
      }
    }

    // 6. 생성 날짜 검증
    const createdMatch = CODE_FIRST_PATTERNS.CREATED_LINE.exec(blockContent);
    if (createdMatch) {
      const created = createdMatch[1];
      if (!/^\d{4}-\d{2}-\d{2}$/.test(created)) {
        warnings.push(`생성 날짜 형식을 확인하세요: ${created} (YYYY-MM-DD)`);
      }
    }

    // 7. @IMMUTABLE 마커 권장
    if (!CODE_FIRST_PATTERNS.IMMUTABLE_MARKER.test(blockContent)) {
      warnings.push('@IMMUTABLE 마커를 추가하여 TAG 불변성을 보장하는 것을 권장합니다');
    }

    return {
      isValid: violations.length === 0,
      violations,
      warnings,
      hasTag
    };
  }
}

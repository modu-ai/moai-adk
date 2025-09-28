/**
 * @FEATURE:TAG-VALIDATOR-001 16-Core TAG Validator
 *
 * Python tag_system/validator.py의 완전 포팅
 * Primary Chain 검증 및 무결성 검사
 *
 * @TASK:TAG-VALIDATOR-001 TAG 검증 엔진 TypeScript 구현
 * @DESIGN:PYTHON-PORTING-001 Python 구현과 완전 호환
 */

import type { TagMatch } from './tag-parser';

/**
 * Chain 검증 결과
 */
export interface ChainValidationResult {
  readonly isValid: boolean;
  readonly completenessScore: number;
  readonly missingLinks: readonly string[];
  readonly chainType: string;
}

/**
 * 검증 오류 정보
 */
export interface ValidationError {
  readonly message: string;
  readonly tag?: TagMatch;
}

/**
 * 깨진 참조 정보
 */
export interface BrokenReference {
  readonly sourceIdentifier: string;
  readonly brokenReference: string;
  readonly reason: string;
}

/**
 * 일관성 위반 정보
 */
export interface ConsistencyViolation {
  readonly identifier: string;
  readonly category: string;
  readonly issueType: string;
  readonly expected: string;
  readonly actual: string;
}

/**
 * 16-Core TAG 검증 엔진
 *
 * Python TagValidator 클래스의 완전 포팅
 * TRUST 원칙 적용:
 * - Test First: 테스트 요구사항만 충족
 * - Readable: 명확한 검증 로직
 * - Unified: 검증 책임만 담당
 * - Secured: 입력 검증 및 안전한 순환 참조 검출
 * - Trackable: @TAG 시스템으로 추적 가능한 검증
 */
export class TagValidator {
  // 16-Core TAG 체계 정의 (Python과 동일)
  private readonly primaryChain = ['REQ', 'DESIGN', 'TASK', 'TEST'];
  private readonly steeringChain = ['VISION', 'STRUCT', 'TECH', 'ADR'];
  private readonly implementationChain = ['FEATURE', 'API', 'UI', 'DATA'];
  private readonly qualityChain = ['PERF', 'SEC', 'DOCS', 'TAG'];

  private readonly allCategories: Record<string, readonly string[]>;

  constructor() {
    // 모든 카테고리 매핑 초기화 (Python과 동일 구조)
    this.allCategories = {
      PRIMARY: this.primaryChain,
      STEERING: this.steeringChain,
      IMPLEMENTATION: this.implementationChain,
      QUALITY: this.qualityChain,
    };
  }

  /**
   * Primary Chain 검증
   * Python: validate_primary_chain(tags)
   */
  validatePrimaryChain(tags: TagMatch[]): ChainValidationResult {
    const primaryTags = tags.filter(tag =>
      this.primaryChain.includes(tag.category)
    );
    const foundCategories = new Set(primaryTags.map(tag => tag.category));

    const missingLinks = this.primaryChain.filter(
      cat => !foundCategories.has(cat)
    );
    const completenessScore = foundCategories.size / this.primaryChain.length;

    return {
      isValid: missingLinks.length === 0,
      completenessScore,
      missingLinks: Object.freeze(missingLinks),
      chainType: 'PRIMARY',
    };
  }

  /**
   * 순환 참조 검사
   * Python: detect_circular_references(tags)
   */
  detectCircularReferences(tags: TagMatch[]): TagMatch[][] {
    const circularRefs: TagMatch[][] = [];

    // 태그별 참조 매핑 구축 (Python과 동일 로직)
    const tagRefs: Record<string, string[]> = {};
    const tagMap: Record<string, TagMatch> = {};

    for (const tag of tags) {
      const key = `${tag.category}:${tag.identifier}`;
      tagMap[key] = tag;
      tagRefs[key] = tag.references ? [...tag.references] : [];
    }

    // DFS로 순환 참조 검색 (Python과 동일 알고리즘)
    const visited = new Set<string>();
    const recursionStack = new Set<string>();

    const dfs = (tagKey: string, path: TagMatch[]): void => {
      if (recursionStack.has(tagKey)) {
        // 순환 참조 발견
        const cycleStartIndex = path.findIndex(
          tag => `${tag.category}:${tag.identifier}` === tagKey
        );
        if (cycleStartIndex >= 0) {
          const cycle = path.slice(cycleStartIndex);
          if (cycle.length > 0) {
            circularRefs.push(cycle);
          }
        }
        return;
      }

      if (visited.has(tagKey)) {
        return;
      }

      visited.add(tagKey);
      recursionStack.add(tagKey);

      if (tagKey in tagMap) {
        const currentPath = [...path, tagMap[tagKey]!];
        const refs = tagRefs[tagKey] || [];
        for (const ref of refs) {
          if (ref in tagMap) {
            dfs(ref, currentPath);
          }
        }
      }

      recursionStack.delete(tagKey);
    };

    // 모든 TAG에 대해 DFS 실행
    for (const tag of tags) {
      const tagKey = `${tag.category}:${tag.identifier}`;
      if (!visited.has(tagKey)) {
        dfs(tagKey, []);
      }
    }

    return circularRefs;
  }

  /**
   * 고아 TAG 검색
   * Python: find_orphaned_tags(tags)
   */
  findOrphanedTags(tags: TagMatch[]): TagMatch[] {
    const orphanedTags: TagMatch[] = [];
    const allReferences = new Set<string>();

    // 모든 참조 수집 (Python과 동일 로직)
    for (const tag of tags) {
      if (tag.references) {
        for (const ref of tag.references) {
          allReferences.add(ref);
        }
      }
    }

    // 참조되지 않고 참조도 하지 않는 TAG 찾기
    for (const tag of tags) {
      const tagKey = `${tag.category}:${tag.identifier}`;
      const hasNoReferences = !tag.references || tag.references.length === 0;
      const isNotReferenced = !allReferences.has(tagKey);

      if (hasNoReferences && isNotReferenced) {
        orphanedTags.push(tag);
      }
    }

    return orphanedTags;
  }

  /**
   * 명명 일관성 검사
   * Python: check_naming_consistency(tags)
   */
  checkNamingConsistency(tags: TagMatch[]): ConsistencyViolation[] {
    const violations: ConsistencyViolation[] = [];

    for (const tag of tags) {
      // 대문자-하이픈 패턴 검사 (Python과 동일)
      if (!this.isConsistentNaming(tag.identifier)) {
        violations.push({
          identifier: tag.identifier,
          category: tag.category,
          issueType: 'naming_inconsistency',
          expected: 'UPPERCASE-WITH-HYPHENS',
          actual: tag.identifier,
        });
      }
    }

    return violations;
  }

  /**
   * TAG 커버리지 계산
   * Python: calculate_tag_coverage(tags)
   */
  calculateTagCoverage(tags: TagMatch[]): Record<string, number> {
    const coverage: Record<string, number> = {};

    for (const [groupName, categories] of Object.entries(this.allCategories)) {
      const foundCategories = new Set<string>();

      for (const tag of tags) {
        if (categories.includes(tag.category)) {
          foundCategories.add(tag.category);
        }
      }

      const coverageRatio = foundCategories.size / categories.length;
      coverage[groupName] = coverageRatio;
    }

    return coverage;
  }

  /**
   * 참조 무결성 검사
   * Python: validate_reference_integrity(tags)
   */
  validateReferenceIntegrity(tags: TagMatch[]): BrokenReference[] {
    const brokenRefs: BrokenReference[] = [];
    const existingTags = new Set<string>();

    // 존재하는 TAG 수집 (Python과 동일 로직)
    for (const tag of tags) {
      const tagKey = `${tag.category}:${tag.identifier}`;
      existingTags.add(tagKey);
    }

    // 참조 무결성 검사
    for (const tag of tags) {
      if (tag.references) {
        for (const ref of tag.references) {
          if (!existingTags.has(ref)) {
            brokenRefs.push({
              sourceIdentifier: tag.identifier,
              brokenReference: ref,
              reason: 'Referenced tag does not exist',
            });
          }
        }
      }
    }

    return brokenRefs;
  }

  /**
   * 명명 일관성 검사 (Private)
   * Python: _is_consistent_naming(identifier)
   */
  private isConsistentNaming(identifier: string): boolean {
    // 대문자-하이픈-숫자 패턴 허용 (Python과 동일 패턴)
    const pattern = /^[A-Z][A-Z0-9-]*[A-Z0-9]$/;
    return pattern.test(identifier);
  }
}

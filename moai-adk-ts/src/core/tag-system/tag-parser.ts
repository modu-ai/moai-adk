/**
 * @FEATURE:TAG-PARSER-001  TAG Parser
 *
 * Python tag_system/parser.py의 완전 포팅
 *  TAG 시스템의 TAG 추출 및 분류 엔진
 *
 * @TASK:TAG-PARSER-001 TAG 파싱 엔진 TypeScript 구현
 * @DESIGN:PYTHON-PORTING-001 Python 구현과 완전 호환
 */

/**
 *  TAG 카테고리 분류
 */
export enum TagCategory {
  PRIMARY = 'PRIMARY',
  STEERING = 'STEERING',
  IMPLEMENTATION = 'IMPLEMENTATION',
  QUALITY = 'QUALITY',
}

/**
 * TAG 매칭 결과
 */
export interface TagMatch {
  readonly category: string;
  readonly identifier: string;
  readonly description: string | null;
  readonly references: readonly string[];
}

/**
 * TAG 위치 정보
 */
export interface TagPosition {
  readonly lineNumber: number;
  readonly column: number;
}

/**
 * TAG 체인
 */
export interface TagChain {
  readonly links: readonly TagMatch[];
}

/**
 * 중복 TAG 정보
 */
export interface DuplicateTagInfo {
  readonly category: string;
  readonly identifier: string;
  readonly positions: readonly TagPosition[];
}

/**
 *  TAG 파싱 엔진
 *
 * Python TagParser 클래스의 완전 포팅
 * TRUST 원칙 적용:
 * - Test First: 테스트가 요구하는 최소 기능만 구현
 * - Readable: 명시적이고 이해하기 쉬운 코드
 * - Unified: 단일 책임 - TAG 파싱만 담당
 * - Secured: 입력 검증 및 안전한 정규표현식 사용
 * - Trackable: @TAG 시스템으로 추적 가능한 구현
 */
export class TagParser {
  // 상수:  TAG 체계 정의 (Python과 동일)
  private static readonly PRIMARY_CATEGORIES = [
    'REQ',
    'DESIGN',
    'TASK',
    'TEST',
  ];
  private static readonly STEERING_CATEGORIES = [
    'VISION',
    'STRUCT',
    'TECH',
    'ADR',
  ];
  private static readonly IMPLEMENTATION_CATEGORIES = [
    'FEATURE',
    'API',
    'UI',
    'DATA',
  ];
  private static readonly QUALITY_CATEGORIES = ['PERF', 'SEC', 'DOCS', 'TAG'];

  // 정규표현식 패턴 상수 (Python과 동일)
  private static readonly TAG_PATTERN =
    /(?<!@)@([A-Z]+):([A-Z0-9-]+)(?:\s+([^\n\r@]+))?/g;
  private static readonly CHAIN_PATTERN =
    /@[A-Z]+:[A-Z0-9-]+(?:\s*→\s*@[A-Z]+:[A-Z0-9-]+)+/g;
  private static readonly INDIVIDUAL_TAG_PATTERN = /@([A-Z]+):([A-Z0-9-]+)/g;

  private readonly tagCategories: Record<TagCategory, readonly string[]>;

  constructor() {
    // TAG 카테고리 매핑 초기화 (Python과 동일 구조)
    this.tagCategories = {
      [TagCategory.PRIMARY]: TagParser.PRIMARY_CATEGORIES,
      [TagCategory.STEERING]: TagParser.STEERING_CATEGORIES,
      [TagCategory.IMPLEMENTATION]: TagParser.IMPLEMENTATION_CATEGORIES,
      [TagCategory.QUALITY]: TagParser.QUALITY_CATEGORIES,
    };
  }

  /**
   *  TAG 카테고리 반환
   * Python: get_tag_categories()
   */
  getTagCategories(): Record<TagCategory, readonly string[]> {
    return { ...this.tagCategories };
  }

  /**
   * 텍스트에서 TAG 추출
   * Python: extract_tags(content)
   */
  extractTags(content: string): TagMatch[] {
    const tags: TagMatch[] = [];
    const regex = new RegExp(TagParser.TAG_PATTERN);
    let match: RegExpExecArray | null;

    while ((match = regex.exec(content)) !== null) {
      const category = match[1]!;
      const identifier = match[2]!;
      const description = match[3] || null;

      //  TAG 검증 (Python과 동일)
      if (this.isValidTagCategory(category)) {
        tags.push({
          category,
          identifier,
          description,
          references: [],
        });
      }
    }

    return tags;
  }

  /**
   * TAG 체인 파싱
   * Python: parse_tag_chains(content)
   */
  parseTagChains(content: string): TagChain[] {
    const chains: TagChain[] = [];
    const chainRegex = new RegExp(TagParser.CHAIN_PATTERN);
    let match: RegExpExecArray | null;

    while ((match = chainRegex.exec(content)) !== null) {
      const fullMatch = match[0];
      // 개별 TAG들을 추출
      const individualTagRegex = new RegExp(TagParser.INDIVIDUAL_TAG_PATTERN);
      const links: TagMatch[] = [];
      let tagMatch: RegExpExecArray | null;

      while ((tagMatch = individualTagRegex.exec(fullMatch)) !== null) {
        const category = tagMatch[1]!;
        const identifier = tagMatch[2]!;

        if (this.isValidTagCategory(category)) {
          links.push({
            category,
            identifier,
            description: null,
            references: [],
          });
        }
      }

      if (links.length > 0) {
        chains.push({ links: Object.freeze(links) });
      }
    }

    return chains;
  }

  /**
   * TAG 형식 검증
   * Python: validate_tag_format(tag_string)
   */
  validateTagFormat(tagString: string): boolean {
    const regex = new RegExp(`^${TagParser.TAG_PATTERN.source}$`);
    const match = regex.exec(tagString.trim());

    if (!match) {
      return false;
    }

    const category = match[1]!;
    return this.isValidTagCategory(category);
  }

  /**
   * 위치 정보와 함께 TAG 추출
   * Python: extract_tags_with_positions(content)
   */
  extractTagsWithPositions(content: string): Array<[TagMatch, TagPosition]> {
    const results: Array<[TagMatch, TagPosition]> = [];
    const lines = content.split('\n');

    lines.forEach((line, lineIndex) => {
      const regex = new RegExp(TagParser.TAG_PATTERN);
      let match: RegExpExecArray | null;

      while ((match = regex.exec(line)) !== null) {
        const category = match[1]!;
        const identifier = match[2]!;
        const description = match[3] || null;

        if (this.isValidTagCategory(category)) {
          const tag: TagMatch = {
            category,
            identifier,
            description,
            references: [],
          };

          const position: TagPosition = {
            lineNumber: lineIndex + 1, // 1-based line numbers
            column: match.index!,
          };

          results.push([tag, position]);
        }
      }
    });

    return results;
  }

  /**
   * 중복 TAG 검색
   * Python: find_duplicate_tags(content)
   */
  findDuplicateTags(content: string): DuplicateTagInfo[] {
    const tagPositions: Record<string, TagPosition[]> = {};
    const tagsWithPos = this.extractTagsWithPositions(content);

    // TAG별 위치 수집 (Python과 동일 로직)
    for (const [tag, position] of tagsWithPos) {
      const key = `${tag.category}:${tag.identifier}`;
      if (!tagPositions[key]) {
        tagPositions[key] = [];
      }
      tagPositions[key].push(position);
    }

    // 중복 TAG 식별 (Python과 동일 로직)
    const duplicates: DuplicateTagInfo[] = [];
    for (const [key, positions] of Object.entries(tagPositions)) {
      if (positions.length > 1) {
        const [category, identifier] = key.split(':', 2);
        duplicates.push({
          category: category!,
          identifier: identifier!,
          positions: Object.freeze(positions),
        });
      }
    }

    return duplicates;
  }

  /**
   * TAG 카테고리 유효성 검사
   * Python: _is_valid_tag_category(category)
   */
  private isValidTagCategory(category: string): boolean {
    const allCategories = Object.values(this.tagCategories).flat();
    return allCategories.includes(category);
  }
}

/**
 * @TAG:API:AI-SEARCH-PATTERNS-001
 * @CHAIN: REQ:TAG-SYSTEM-001 -> DESIGN:CODE-FIRST-001 -> TASK:AI-SEARCH-001 -> @TAG:API:AI-SEARCH-PATTERNS-001
 * @DEPENDS: @TAG:API:CODE-FIRST-TYPES-001, @TAG:API:CODE-FIRST-PARSER-001
 * @STATUS: active
 * @CREATED: 2025-09-29
 * @IMMUTABLE
 */

/**
 * AI 검색 패턴 시스템
 *
 * Code-First TAG 시스템에서 AI가 코드 파일에서 직접 TAG를 읽어올 수 있도록
 * 최적화된 ripgrep 기반 검색 패턴을 제공합니다.
 *
 * 핵심 원칙:
 * - 코드가 유일한 진실의 원천 (No Index Files)
 * - ripgrep 기반 고성능 검색
 * - AI 컨텍스트 구축 최적화
 * - TAG 관계 분석 자동화
 */

import { AISearchPattern, TagCategory, TagBlock } from './code-first-types.js';

/**
 * 핵심 TAG 검색 패턴들
 *
 * AI가 코드베이스에서 TAG를 효율적으로 찾기 위한
 * ripgrep 패턴 모음
 */
export class AISearchPatterns {
  /**
   * 모든 TAG 블록을 찾는 기본 패턴
   */
  static readonly FIND_ALL_TAGS: AISearchPattern = {
    name: "find_all_tags",
    pattern: "@TAG:[A-Z]+:[A-Z0-9-]+",
    fileTypes: ["ts", "js", "py", "java", "go", "rs", "cpp", "cs", "md"],
    options: ["-n", "--color=never"]
  };

  /**
   * 특정 카테고리의 TAG들을 찾는 패턴
   */
  static createCategoryPattern(category: TagCategory): AISearchPattern {
    return {
      name: `find_${category.toLowerCase()}_tags`,
      pattern: `@TAG:${category}:[A-Z0-9-]+`,
      fileTypes: ["ts", "js", "py", "java", "go", "rs", "cpp", "cs", "md"],
      options: ["-n", "--color=never", "-A", "10"]
    };
  }

  /**
   * TAG 체인을 찾는 패턴
   */
  static readonly FIND_TAG_CHAINS: AISearchPattern = {
    name: "find_tag_chains",
    pattern: "@CHAIN:\\s*([A-Z]+:[A-Z0-9-]+\\s*->\\s*)+[A-Z]+:[A-Z0-9-]+",
    fileTypes: ["ts", "js", "py", "java", "go", "rs", "cpp", "cs", "md"],
    options: ["-n", "--color=never", "-A", "5"]
  };

  /**
   * TAG 의존성을 찾는 패턴
   */
  static readonly FIND_TAG_DEPENDENCIES: AISearchPattern = {
    name: "find_tag_dependencies",
    pattern: "@DEPENDS:\\s*(@TAG:[A-Z]+:[A-Z0-9-]+(,\\s*)?)+",
    fileTypes: ["ts", "js", "py", "java", "go", "rs", "cpp", "cs", "md"],
    options: ["-n", "--color=never", "-A", "5"]
  };

  /**
   * 불변 TAG를 찾는 패턜
   */
  static readonly FIND_IMMUTABLE_TAGS: AISearchPattern = {
    name: "find_immutable_tags",
    pattern: "@IMMUTABLE",
    fileTypes: ["ts", "js", "py", "java", "go", "rs", "cpp", "cs", "md"],
    options: ["-n", "--color=never", "-B", "10", "-A", "2"]
  };

  /**
   * 완전한 TAG 블록을 찾는 패턴 (주석 시작부터 끝까지)
   */
  static readonly FIND_COMPLETE_TAG_BLOCKS: AISearchPattern = {
    name: "find_complete_tag_blocks",
    pattern: "/\\*\\*[\\s\\S]*?@TAG:[A-Z]+:[A-Z0-9-]+[\\s\\S]*?\\*/",
    fileTypes: ["ts", "js", "java", "cs", "cpp"],
    options: ["-U", "--multiline", "-n", "--color=never"]
  };

  /**
   * Python docstring TAG 블록 패턴
   */
  static readonly FIND_PYTHON_TAG_BLOCKS: AISearchPattern = {
    name: "find_python_tag_blocks",
    pattern: '"""[\\s\\S]*?@TAG:[A-Z]+:[A-Z0-9-]+[\\s\\S]*?"""',
    fileTypes: ["py"],
    options: ["-U", "--multiline", "-n", "--color=never"]
  };

  /**
   * 특정 도메인 ID의 TAG들을 찾는 패턴
   */
  static createDomainPattern(domainId: string): AISearchPattern {
    return {
      name: `find_domain_${domainId.toLowerCase()}`,
      pattern: `@TAG:[A-Z]+:${domainId.toUpperCase()}`,
      fileTypes: ["ts", "js", "py", "java", "go", "rs", "cpp", "cs", "md"],
      options: ["-n", "--color=never", "-B", "5", "-A", "15"]
    };
  }

  /**
   * 모든 코어 패턴들
   */
  static readonly CORE_PATTERNS: AISearchPattern[] = [
    AISearchPatterns.FIND_ALL_TAGS,
    AISearchPatterns.FIND_TAG_CHAINS,
    AISearchPatterns.FIND_TAG_DEPENDENCIES,
    AISearchPatterns.FIND_IMMUTABLE_TAGS,
    AISearchPatterns.FIND_COMPLETE_TAG_BLOCKS,
    AISearchPatterns.FIND_PYTHON_TAG_BLOCKS
  ];
}

/**
 * AI 컨텍스트 빌더
 *
 * 코드베이스 전체의 TAG를 스캔하여 AI가 이해할 수 있는
 * 컨텍스트 정보를 구축합니다.
 */
export class AIContextBuilder {
  /**
   * 전체 프로젝트의 TAG 맵을 구축
   *
   * @param rootPath 프로젝트 루트 경로
   * @returns TAG 맵과 관계 정보
   */
  async buildProjectTagMap(rootPath: string): Promise<{
    tags: Map<string, TagBlock>;
    chains: Map<string, string[]>;
    dependencies: Map<string, string[]>;
  }> {
    const tags = new Map<string, TagBlock>();
    const chains = new Map<string, string[]>();
    const dependencies = new Map<string, string[]>();

    // TODO: 실제 구현은 CodeFirstTagParser와 연동
    // 여기서는 인터페이스만 정의

    return { tags, chains, dependencies };
  }

  /**
   * 특정 TAG와 관련된 컨텍스트 구축
   *
   * @param tagId TAG ID (예: "FEATURE:AUTH-001")
   * @param rootPath 프로젝트 루트 경로
   * @returns 관련 TAG들과 체인 정보
   */
  async buildTagContext(tagId: string, rootPath: string): Promise<{
    primaryTag: TagBlock;
    relatedTags: TagBlock[];
    chainTags: TagBlock[];
    dependentTags: TagBlock[];
  }> {
    // TODO: 실제 구현 필요
    throw new Error("Not implemented yet");
  }

  /**
   * AI가 읽기 좋은 형태의 TAG 요약 생성
   *
   * @param tags TAG 목록
   * @returns 마크다운 형식의 요약
   */
  generateTagSummaryForAI(tags: TagBlock[]): string {
    const summary = [`# 프로젝트 TAG 요약 (Code-First)`];

    // 카테고리별 분류
    const categorized = new Map<TagCategory, TagBlock[]>();

    for (const tag of tags) {
      if (!categorized.has(tag.category)) {
        categorized.set(tag.category, []);
      }
      categorized.get(tag.category)!.push(tag);
    }

    // 생명주기 TAG들 우선 표시
    const lifecycleOrder: TagCategory[] = [
      TagCategory.SPEC, TagCategory.REQ, TagCategory.DESIGN,
      TagCategory.TASK, TagCategory.TEST
    ];

    for (const category of lifecycleOrder) {
      const categoryTags = categorized.get(category);
      if (categoryTags && categoryTags.length > 0) {
        summary.push(`\n## ${category} 태그들`);
        for (const tag of categoryTags) {
          summary.push(`- **${tag.tag}** (${tag.filePath}:${tag.lineNumber})`);
          if (tag.chain && tag.chain.length > 0) {
            summary.push(`  - 체인: ${tag.chain.join(' -> ')}`);
          }
          if (tag.depends && tag.depends.length > 0) {
            summary.push(`  - 의존성: ${tag.depends.join(', ')}`);
          }
        }
      }
    }

    // 구현 TAG들
    const implCategories = [TagCategory.FEATURE, TagCategory.API, TagCategory.FIX];

    for (const category of implCategories) {
      const categoryTags = categorized.get(category);
      if (categoryTags && categoryTags.length > 0) {
        summary.push(`\n## ${category} 태그들`);
        for (const tag of categoryTags) {
          summary.push(`- **${tag.tag}** (${tag.filePath}:${tag.lineNumber})`);
        }
      }
    }

    return summary.join('\n');
  }
}

/**
 * TAG 검색 실행기
 *
 * ripgrep를 사용하여 실제 검색을 수행하고 결과를 파싱합니다.
 */
export class TagSearchExecutor {
  /**
   * 패턴을 사용하여 실제 검색 수행
   *
   * @param pattern 검색 패턴
   * @param rootPath 검색할 루트 경로
   * @returns 검색 결과
   */
  async executeSearch(pattern: AISearchPattern, rootPath: string): Promise<{
    matches: Array<{
      file: string;
      line: number;
      content: string;
    }>;
    totalMatches: number;
  }> {
    // TODO: 실제 ripgrep 명령어 실행
    // 여기서는 인터페이스만 정의

    return {
      matches: [],
      totalMatches: 0
    };
  }

  /**
   * 여러 패턴을 병렬로 실행
   *
   * @param patterns 검색할 패턴들
   * @param rootPath 검색할 루트 경로
   * @returns 통합된 검색 결과
   */
  async executeMultipleSearches(
    patterns: AISearchPattern[],
    rootPath: string
  ): Promise<Map<string, any>> {
    // TODO: 병렬 검색 실행 구현
    return new Map();
  }
}

/**
 * AI가 사용할 수 있는 고수준 TAG 검색 API
 */
export class AITagSearchAPI {
  private contextBuilder = new AIContextBuilder();
  private searchExecutor = new TagSearchExecutor();

  /**
   * AI가 프로젝트 전체 컨텍스트를 이해할 수 있도록
   * 모든 TAG 정보를 제공
   *
   * @param rootPath 프로젝트 루트 경로
   * @returns AI용 컨텍스트 정보
   */
  async getProjectContext(rootPath: string): Promise<string> {
    const tagMap = await this.contextBuilder.buildProjectTagMap(rootPath);
    const tags = Array.from(tagMap.tags.values());
    return this.contextBuilder.generateTagSummaryForAI(tags);
  }

  /**
   * 특정 기능이나 모듈과 관련된 TAG들을 찾아 반환
   *
   * @param keyword 검색 키워드 (예: "auth", "user", "api")
   * @param rootPath 프로젝트 루트 경로
   * @returns 관련 TAG 정보
   */
  async findRelatedTags(keyword: string, rootPath: string): Promise<TagBlock[]> {
    // TODO: 키워드 기반 TAG 검색 구현
    return [];
  }

  /**
   * TAG 체인의 완성도 분석
   *
   * @param rootPath 프로젝트 루트 경로
   * @returns 체인별 완성도 리포트
   */
  async analyzeChainCompleteness(rootPath: string): Promise<{
    chains: Array<{
      chainId: string;
      completeness: number;
      missingSteps: TagCategory[];
    }>;
    overallCompleteness: number;
  }> {
    // TODO: 체인 완성도 분석 구현
    return {
      chains: [],
      overallCompleteness: 0
    };
  }
}

// 기본 API 인스턴스 내보내기
export const aiTagSearch = new AITagSearchAPI();
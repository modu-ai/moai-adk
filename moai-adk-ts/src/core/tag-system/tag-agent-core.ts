/**
 * @FEATURE:TAG-AGENT-001 | Chain: @REQ:TAG-AGENT-001 → @DESIGN:TAG-AGENT-001 → @TASK:TAG-AGENT-001 → @FEATURE:TAG-AGENT-001
 * @PERF:TAG-AGENT-001: 고성능 TAG 관리 및 검색 최적화
 * @SEC:TAG-AGENT-001: TAG 입력 검증 및 보안 강화
 *
 * TAG Agent 핵심 구현체
 *
 *  TAG 시스템의 모든 관리 기능을 제공하는 전문 에이전트 코어입니다.
 * - TAG 생성, 검증, 체인 무결성 관리
 * - 지능적 중복 방지 및 재사용 제안
 * - JSONL 인덱스 자동 관리 및 성능 최적화
 * - 고아 TAG 방지 및 순환 참조 해결
 */

import { promises as fs } from 'node:fs';
import { join } from 'node:path';
import { CodeFirstTagParser } from './code-first-parser.js';
import { TagParser } from './tag-parser.js';
import type {
  TagSearchQuery,
  TagStatistics,
  TagValidationResult,
} from './types.js';
import type { TagBlock } from './code-first-types.js';

/**
 * TAG Agent 요청 인터페이스
 */
export interface TagAgentRequest {
  action: 'create' | 'search' | 'validate' | 'repair' | 'index' | 'stats';
  domain?: string;
  keyword?: string;
  category?: string;
  description?: string;
  relatedFiles?: string[];
  parentTags?: string[];
  options?: Record<string, any>;
}

/**
 * TAG Agent 응답 인터페이스
 */
export interface TagAgentResponse {
  success: boolean;
  message: string;
  data?: any;
  createdTags?: string[];
  suggestedTags?: string[];
  warnings?: string[];
  performance?: {
    executionTime: number;
    indexSize: number;
    searchSpeed: number;
  };
}

/**
 * TAG 재사용 제안
 */
export interface TagReuseRecommendation {
  existingTag: string;
  similarity: number;
  reason: string;
  shouldReuse: boolean;
}

/**
 * TAG 체인 수리 결과
 */
export interface TagChainRepairResult {
  repairedChains: number;
  orphanedTagsFixed: number;
  circularReferencesResolved: number;
  newConnections: string[];
}

/**
 * @FEATURE:TAG-AGENT-001: TAG Agent 핵심 구현체
 *
 * MoAI-ADK의  TAG 시스템을 완전히 관리하는 전문 에이전트입니다.
 * Claude Code 에이전트에서 호출되어 모든 TAG 관리 작업을 자동화합니다.
 */
export class TagAgentCore {
  private readonly codeFirstParser: CodeFirstTagParser;
  private readonly tagParser: TagParser;
  private readonly projectRoot: string;

  constructor(projectRoot: string) {
    this.projectRoot = projectRoot;

    // NOTE: [v0.0.3+] CODE-FIRST TAG 시스템
    // - tags.json 인덱스 캐시 제거
    // - 코드 직접 스캔 (rg/grep) 기반 실시간 검증
    // - 단일 진실 소스(코드)로 동기화 문제 완전 해결
    this.codeFirstParser = new CodeFirstTagParser({
      scanExtensions: ['.ts', '.tsx', '.js', '.jsx', '.py', '.md'],
      excludeDirectories: ['node_modules', '.git', 'dist', 'build', '__pycache__'],
      enforceImmutability: true,
      enablePerformanceTracking: true,
    });

    this.tagParser = new TagParser();
  }

  /**
   * @API:TAG-AGENT-001: TAG Agent 메인 처리 함수
   *
   * Claude Code 에이전트에서 호출하는 주요 진입점입니다.
   */
  async processRequest(request: TagAgentRequest): Promise<TagAgentResponse> {
    const startTime = performance.now();

    try {
      // NOTE: 코드 직접 스캔 방식 - tags.json 로드 제거

      let result: any;

      switch (request.action) {
        case 'create':
          result = await this.createTagChain(request);
          break;
        case 'search':
          result = await this.searchTags(request);
          break;
        case 'validate':
          result = await this.validateTagSystem();
          break;
        case 'repair':
          result = await this.repairTagChains();
          break;
        case 'index':
          // NOTE: 인덱싱 불필요 - 코드 직접 스캔 방식
          result = { message: 'No indexing needed - using code-first approach' };
          break;
        case 'stats':
          result = await this.generateStatistics();
          break;
        default:
          throw new Error(`Unsupported action: ${request.action}`);
      }

      const executionTime = performance.now() - startTime;

      return {
        success: true,
        message: `TAG ${request.action} completed successfully`,
        data: result,
        performance: {
          executionTime,
          indexSize: 0, // 인덱스 없음
          searchSpeed: executionTime,
        },
      };
    } catch (error) {
      return {
        success: false,
        message: `TAG ${request.action} failed: ${error instanceof Error ? error.message : String(error)}`,
        warnings: [
          error instanceof Error ? error.stack || error.message : String(error),
        ],
      };
    }
  }

  /**
   * @API:TAG-CREATE-001: TAG 체인 생성 가이드 제공 (CODE-FIRST)
   *
   * CODE-FIRST 철학: TAG는 코드에만 존재합니다.
   * 이 메서드는 새로운 TAG를 생성하는 대신, 코드에 작성할 TAG 템플릿을 제공합니다.
   */
  private async createTagChain(request: TagAgentRequest): Promise<{
    createdTags: string[];
    suggestedReuse: TagReuseRecommendation[];
    chainIntegrity: boolean;
  }> {
    if (!request.domain) {
      throw new Error('Domain is required for TAG creation');
    }

    // 1. 기존 유사 TAG 검색 (코드 스캔)
    const existingSimilar = await this.findSimilarTags(request.domain);
    const reuseRecommendations: TagReuseRecommendation[] = [];

    for (const similar of existingSimilar) {
      const similarity = this.calculateSimilarity(request.domain, similar.tag);
      if (similarity > 0.7) {
        reuseRecommendations.push({
          existingTag: similar.tag,
          similarity,
          reason: `Found in ${similar.filePath}`,
          shouldReuse: similarity > 0.8,
        });
      }
    }

    // 2. 새로운 TAG ID 생성
    const baseId = this.generateTagId(request.domain);
    const nextId = await this.getNextAvailableId(baseId);

    // 3. Primary Chain TAG IDs 생성 (가이드용)
    const createdTags: string[] = [];
    const primaryChain = ['REQ', 'DESIGN', 'TASK', 'TEST'];

    for (const category of primaryChain) {
      const tagId = `@${category}:${nextId}`;
      createdTags.push(tagId);
    }

    // NOTE: 실제 TAG 생성은 코드에 직접 작성해야 합니다.
    // 이 메서드는 TAG 템플릿만 제공합니다.

    return {
      createdTags,
      suggestedReuse: reuseRecommendations,
      chainIntegrity: true, // 템플릿 제공만 하므로 항상 true
    };
  }

  /**
   * @API:TAG-SEARCH-001: 지능적 TAG 검색 (CODE-FIRST)
   *
   * 코드를 직접 스캔하여 TAG 검색 (tags.json 인덱스 불필요)
   */
  private async searchTags(request: TagAgentRequest): Promise<{
    matches: TagBlock[];
    suggestions: string[];
    reuseRecommendations: TagReuseRecommendation[];
  }> {
    if (!request.keyword) {
      throw new Error('Keyword is required for TAG search');
    }

    // 1. 코드 전체 스캔
    const allTagBlocks = await this.codeFirstParser.scanDirectory(this.projectRoot);

    // 2. 키워드로 필터링
    const keyword = request.keyword.toUpperCase();
    const matches = allTagBlocks.filter(block => {
      // TAG ID 매칭
      if (block.tag.includes(keyword)) return true;
      // 도메인 ID 매칭
      if (block.domainId.includes(keyword)) return true;
      // 카테고리 매칭
      if (block.category.includes(keyword)) return true;
      // 파일 경로 매칭
      if (request.keyword.includes('/') && block.filePath.includes(request.keyword)) return true;
      return false;
    });

    // 3. 재사용 제안 생성
    const reuseRecommendations: TagReuseRecommendation[] = [];
    for (const block of matches) {
      const similarity = this.calculateSimilarity(request.keyword, block.tag);
      if (similarity > 0.5) {
        reuseRecommendations.push({
          existingTag: block.tag,
          similarity,
          reason: `Found in ${block.filePath}`,
          shouldReuse: similarity > 0.7,
        });
      }
    }

    // 4. 검색 제안 생성
    const suggestions = this.generateSearchSuggestions(request.keyword);

    return {
      matches,
      suggestions,
      reuseRecommendations,
    };
  }

  /**
   * @API:TAG-VALIDATE-001: TAG 시스템 전체 검증 (CODE-FIRST)
   *
   * 코드를 직접 스캔하여 TAG 검증
   */
  private async validateTagSystem(): Promise<{
    totalTags: number;
    validTags: number;
    invalidTags: string[];
    orphanedTags: string[];
    brokenChains: string[];
    circularReferences: string[];
  }> {
    // 코드 전체 스캔
    const allTagBlocks = await this.codeFirstParser.scanDirectory(this.projectRoot);

    const invalidTags: string[] = [];
    const orphanedTags: string[] = [];
    const brokenChains: string[] = [];
    const circularReferences: string[] = [];

    // TAG 맵 구축 (검증용)
    const tagMap = new Map<string, TagBlock>();
    for (const block of allTagBlocks) {
      tagMap.set(block.tag, block);
    }

    for (const block of allTagBlocks) {
      // 형식 검증
      if (!this.tagParser.validateTagFormat(block.tag)) {
        invalidTags.push(block.tag);
        continue;
      }

      // 체인 연결 검사
      if (block.chain) {
        for (const chainTag of block.chain) {
          if (!tagMap.has(chainTag)) {
            brokenChains.push(`${block.tag} -> ${chainTag}`);
          }
        }
      }

      // 의존성 검사
      if (block.depends) {
        for (const depTag of block.depends) {
          if (!tagMap.has(depTag)) {
            brokenChains.push(`${block.tag} depends on ${depTag} (not found)`);
          }
        }
      }

      // 고아 TAG 검사 (REQ가 아닌데 체인이 없는 경우)
      if (!block.chain || block.chain.length === 0) {
        if (!block.category.includes('REQ')) {
          orphanedTags.push(block.tag);
        }
      }
    }

    return {
      totalTags: allTagBlocks.length,
      validTags: allTagBlocks.length - invalidTags.length,
      invalidTags,
      orphanedTags,
      brokenChains,
      circularReferences,
    };
  }

  /**
   * @API:TAG-REPAIR-001: TAG 체인 수리 가이드 제공 (CODE-FIRST)
   *
   * CODE-FIRST 철학: TAG는 코드에만 존재하므로, 자동 수리 불가.
   * 대신 문제가 있는 TAG 목록과 수정 가이드를 제공합니다.
   */
  private async repairTagChains(): Promise<TagChainRepairResult> {
    const validation = await this.validateTagSystem();

    // 수리 가이드만 제공
    const newConnections: string[] = [];

    for (const brokenChain of validation.brokenChains) {
      newConnections.push(`Fix needed: ${brokenChain}`);
    }

    for (const orphanTag of validation.orphanedTags) {
      newConnections.push(`Orphaned TAG: ${orphanTag} - Add chain in code`);
    }

    return {
      repairedChains: 0, // 자동 수리 불가
      orphanedTagsFixed: 0, // 자동 수리 불가
      circularReferencesResolved: 0,
      newConnections,
    };
  }

  /**
   * @API:TAG-INDEX-001: 인덱스 최적화
   *
   * NOTE: [v0.0.3+] tags.json 캐싱 제거 - 코드 직접 스캔 방식으로 전환
   * 이 함수는 하위 호환성을 위해 유지하지만 실제로는 코드 스캔만 수행
   */
  private async optimizeIndexes(): Promise<{
    sizeBefore: number;
    sizeAfter: number;
    optimizationRatio: number;
    indexingSpeed: number;
  }> {
    const sizeBefore = await this.getIndexSize();
    const startTime = performance.now();

    // NOTE: 메인 TAG 데이터베이스 최적화 제거 (코드 스캔으로 전환)
    // await this.tagManager.save();

    // 2. 분산 인덱스 재구축
    await this.rebuildDistributedIndexes();

    // 3. 성능 지표 측정
    const sizeAfter = await this.getIndexSize();
    const indexingSpeed = performance.now() - startTime;
    const optimizationRatio = sizeBefore > 0 ? ((sizeBefore - sizeAfter) / sizeBefore) * 100 : 0;

    return {
      sizeBefore,
      sizeAfter,
      optimizationRatio,
      indexingSpeed,
    };
  }

  /**
   * @API:TAG-STATS-001: TAG 시스템 통계 생성 (CODE-FIRST)
   *
   * 코드를 직접 스캔하여 통계 생성
   */
  private async generateStatistics(): Promise<
    TagStatistics & {
      performance: {
        indexSize: number;
        searchSpeed: number;
        memoryUsage: number;
      };
      health: {
        integrityScore: number;
        qualityGate: 'healthy' | 'warning' | 'critical';
      };
    }
  > {
    // 코드 전체 스캔
    const searchStartTime = performance.now();
    const allTagBlocks = await this.codeFirstParser.scanDirectory(this.projectRoot);
    const searchSpeed = performance.now() - searchStartTime;

    const validation = await this.validateTagSystem();

    // 카테고리별 통계
    const byType: any = {};
    const byCategory: any = {};
    const byStatus: any = {};

    for (const block of allTagBlocks) {
      // 타입별 카운트
      byType[block.category] = (byType[block.category] || 0) + 1;

      // 상태별 카운트
      byStatus[block.status] = (byStatus[block.status] || 0) + 1;
    }

    // 건강도 지표
    const integrityScore = validation.totalTags > 0
      ? (validation.validTags / validation.totalTags) * 100
      : 100;
    let qualityGate: 'healthy' | 'warning' | 'critical';

    if (integrityScore >= 95 && validation.brokenChains.length === 0) {
      qualityGate = 'healthy';
    } else if (integrityScore >= 85) {
      qualityGate = 'warning';
    } else {
      qualityGate = 'critical';
    }

    return {
      total: validation.totalTags,
      byType,
      byCategory,
      byStatus,
      orphanedTags: validation.orphanedTags.length,
      circularReferences: validation.circularReferences.length,
      performance: {
        indexSize: 0, // 인덱스 없음
        searchSpeed,
        memoryUsage: process.memoryUsage().heapUsed,
      },
      health: {
        integrityScore,
        qualityGate,
      },
    };
  }

  // Helper Methods

  /**
   * 유사 TAG 검색 (CODE-FIRST)
   */
  private async findSimilarTags(keyword: string): Promise<TagBlock[]> {
    const allTagBlocks = await this.codeFirstParser.scanDirectory(this.projectRoot);
    return allTagBlocks.filter(block => {
      const similarity = this.calculateSimilarity(keyword, block.tag);
      return similarity > 0.3;
    });
  }

  /**
   * 문자열 유사도 계산 (Levenshtein distance 기반)
   */
  private calculateSimilarity(str1: string, str2: string): number {
    const matrix: number[][] = Array(str2.length + 1)
      .fill(0)
      .map(() => Array(str1.length + 1).fill(0));

    for (let i = 0; i <= str1.length; i++) matrix[0]![i] = i;
    for (let j = 0; j <= str2.length; j++) matrix[j]![0] = j;

    for (let j = 1; j <= str2.length; j++) {
      for (let i = 1; i <= str1.length; i++) {
        const indicator = str1[i - 1] === str2[j - 1] ? 0 : 1;
        const row = matrix[j]!;
        const prevRow = matrix[j - 1]!;
        row[i] = Math.min(
          row[i - 1]! + 1, // insertion
          prevRow[i]! + 1, // deletion
          prevRow[i - 1]! + indicator // substitution
        );
      }
    }

    const maxLength = Math.max(str1.length, str2.length);
    const lastRow = matrix[str2.length]!;
    return maxLength === 0
      ? 1
      : (maxLength - lastRow[str1.length]!) / maxLength;
  }

  /**
   * TAG ID 생성
   */
  private generateTagId(domain: string): string {
    const normalizedDomain = domain.toUpperCase().replace(/[^A-Z0-9]/g, '-');
    return normalizedDomain;
  }

  /**
   * 다음 사용 가능한 ID 찾기 (CODE-FIRST)
   */
  private async getNextAvailableId(baseId: string): Promise<string> {
    const allTagBlocks = await this.codeFirstParser.scanDirectory(this.projectRoot);
    const existingIds = new Set<string>();

    // 모든 TAG에서 도메인 ID 추출
    for (const block of allTagBlocks) {
      existingIds.add(block.domainId);
    }

    // 다음 사용 가능한 ID 찾기
    let counter = 1;
    let candidateId = `${baseId}-${counter.toString().padStart(3, '0')}`;

    while (existingIds.has(candidateId)) {
      counter++;
      candidateId = `${baseId}-${counter.toString().padStart(3, '0')}`;
    }

    return candidateId;
  }

  /**
   * TAG 체인 검증 (CODE-FIRST)
   */
  private async validateTagChain(_tags: string[]): Promise<TagValidationResult> {
    // NOTE: CODE-FIRST 방식에서는 체인 검증이 validateTagSystem에서 처리됨
    return {
      isValid: true,
      errors: [],
      warnings: ['Use validateTagSystem() for comprehensive validation'],
    };
  }

  /**
   * @deprecated 인덱스 불필요 - 코드 직접 스캔 방식으로 전환
   */
  private async updateDistributedIndexes(_tags?: string[]): Promise<void> {
    // NO-OP: 인덱스 불필요
  }

  /**
   * @deprecated 인덱스 불필요 - 코드 직접 스캔 방식으로 전환
   */
  private async rebuildDistributedIndexes(): Promise<void> {
    // NO-OP: 인덱스 불필요
  }

  /**
   * 검색 제안 생성
   */
  private generateSearchSuggestions(keyword: string): string[] {
    const suggestions: string[] = [];

    // 카테고리 기반 제안
    const categories = Object.values(this.tagParser.getTagCategories()).flat();
    for (const category of categories) {
      if (
        category.toLowerCase().includes(keyword.toLowerCase()) ||
        keyword.toLowerCase().includes(category.toLowerCase())
      ) {
        suggestions.push(category);
      }
    }

    // 도메인 기반 제안
    const commonDomains = ['AUTH', 'USER', 'API', 'UI', 'DATA', 'PERF', 'SEC'];
    for (const domain of commonDomains) {
      if (this.calculateSimilarity(keyword.toUpperCase(), domain) > 0.3) {
        suggestions.push(domain);
      }
    }

    return suggestions.slice(0, 5); // 최대 5개 제안
  }

  /**
   * 가장 유사한 TAG 찾기
   */
  private async findMostSimilarTag(tagId: string): Promise<TagEntry | null> {
    const allTags = (await this.tagManager.search({})).tags;
    let bestMatch: TagEntry | null = null;
    let bestSimilarity = 0;

    for (const tag of allTags) {
      const similarity = this.calculateSimilarity(tagId, tag.id);
      if (similarity > bestSimilarity && similarity > 0.5) {
        bestSimilarity = similarity;
        bestMatch = tag;
      }
    }

    return bestMatch;
  }

  /**
   * TAG ID에서 도메인 추출
   */
  private extractDomain(tagId: string): string {
    const match = tagId.match(/@[A-Z]+:([A-Z0-9-]+)/);
    return match?.[1] ?? '';
  }

  /**
   * 도메인과 타입에 맞는 부모 TAG 찾기
   */
  private async findParentTagForDomain(
    domain: string,
    tagType: string
  ): Promise<TagEntry | null> {
    const parentTypes: Record<string, string> = {
      DESIGN: 'REQ',
      TASK: 'DESIGN',
      TEST: 'TASK',
      FEATURE: 'TASK',
      API: 'FEATURE',
      UI: 'FEATURE',
      DATA: 'FEATURE',
    };

    const parentType = parentTypes[tagType];
    if (!parentType) return null;

    const parentTagId = `@${parentType}:${domain}`;
    return await this.tagManager.getTag(parentTagId);
  }
}

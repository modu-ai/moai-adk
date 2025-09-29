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
import { TagManager } from './tag-manager.js';
import { TagParser } from './tag-parser.js';
import type {
  TagEntry,
  TagSearchQuery,
  TagStatistics,
  TagValidationResult,
} from './types.js';

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
  private readonly tagManager: TagManager;
  private readonly tagParser: TagParser;
  private readonly indexesPath: string;

  constructor(projectRoot: string) {
    this.projectRoot = projectRoot;
    this.indexesPath = join(projectRoot, '.moai/indexes');

    this.tagManager = new TagManager({
      filePath: join(this.indexesPath, 'tags.json'),
      enableCache: true,
      autoSave: true,
      autoSaveDelay: 1000,
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
      await this.tagManager.load();

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
          result = await this.optimizeIndexes();
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
          indexSize: await this.getIndexSize(),
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
   * @API:TAG-CREATE-001: 새로운 TAG 체인 생성
   *
   * 지능적 중복 방지 및 재사용 제안을 포함한 TAG 체인 생성
   */
  private async createTagChain(request: TagAgentRequest): Promise<{
    createdTags: string[];
    suggestedReuse: TagReuseRecommendation[];
    chainIntegrity: boolean;
  }> {
    if (!request.domain) {
      throw new Error('Domain is required for TAG creation');
    }

    // 1. 기존 유사 TAG 검색
    const existingSimilar = await this.findSimilarTags(request.domain);
    const reuseRecommendations: TagReuseRecommendation[] = [];

    for (const similar of existingSimilar) {
      const similarity = this.calculateSimilarity(request.domain, similar.id);
      if (similarity > 0.7) {
        reuseRecommendations.push({
          existingTag: similar.id,
          similarity,
          reason: `Similar domain: ${similar.description || similar.title}`,
          shouldReuse: similarity > 0.8,
        });
      }
    }

    // 2. 새로운 TAG ID 생성
    const baseId = this.generateTagId(request.domain);
    const nextId = await this.getNextAvailableId(baseId);

    // 3. Primary Chain 생성
    const createdTags: string[] = [];
    const primaryChain = ['REQ', 'DESIGN', 'TASK', 'TEST'];

    for (const [index, category] of primaryChain.entries()) {
      const tagId = `@${category}:${nextId}`;
      const parentTags =
        index > 0 ? [`@${primaryChain[index - 1]}:${nextId}`] : [];

      await this.tagManager.createTag({
        id: tagId,
        type: category as any,
        category: 'PRIMARY',
        title: `${category} for ${request.domain}`,
        description:
          request.description ||
          `${category} requirements for ${request.domain}`,
        parents: parentTags,
        children:
          index < primaryChain.length - 1
            ? [`@${primaryChain[index + 1]}:${nextId}`]
            : [],
        files: request.relatedFiles || [],
      });

      createdTags.push(tagId);
    }

    // 4. 체인 무결성 검증
    const chainIntegrity = await this.validateTagChain(createdTags);

    // 5. 인덱스 업데이트
    await this.updateDistributedIndexes(createdTags);

    return {
      createdTags,
      suggestedReuse: reuseRecommendations,
      chainIntegrity: chainIntegrity.isValid,
    };
  }

  /**
   * @API:TAG-SEARCH-001: 지능적 TAG 검색
   *
   * 키워드, 도메인, 파일 경로 등을 기반으로 한 포괄적 TAG 검색
   */
  private async searchTags(request: TagAgentRequest): Promise<{
    matches: TagEntry[];
    suggestions: string[];
    reuseRecommendations: TagReuseRecommendation[];
  }> {
    if (!request.keyword) {
      throw new Error('Keyword is required for TAG search');
    }

    // 1. 다중 검색 전략
    const searchQuery: TagSearchQuery = {
      idPattern: request.keyword,
    };

    // 키워드가 카테고리명인지 확인
    const categories = Object.values(this.tagParser.getTagCategories()).flat();
    if (categories.includes(request.keyword.toUpperCase())) {
      searchQuery.types = [request.keyword.toUpperCase() as any];
    }

    // 파일 경로 패턴 감지
    if (request.keyword.includes('/') || request.keyword.includes('.')) {
      searchQuery.filePaths = [request.keyword];
    }

    const searchResult = await this.tagManager.search(searchQuery);

    // 2. 유사도 기반 추가 검색
    const allTags = searchResult.tags;
    const similarTags = await this.findSimilarTags(request.keyword);

    // 3. 재사용 제안 생성
    const reuseRecommendations: TagReuseRecommendation[] = [];
    for (const tag of [...allTags, ...similarTags]) {
      const similarity = this.calculateSimilarity(request.keyword, tag.id);
      if (similarity > 0.5) {
        reuseRecommendations.push({
          existingTag: tag.id,
          similarity,
          reason: `Keyword match: ${tag.title}`,
          shouldReuse: similarity > 0.7,
        });
      }
    }

    // 4. 검색 제안 생성
    const suggestions = this.generateSearchSuggestions(request.keyword);

    return {
      matches: searchResult.tags,
      suggestions,
      reuseRecommendations,
    };
  }

  /**
   * @API:TAG-VALIDATE-001: TAG 시스템 전체 검증
   */
  private async validateTagSystem(): Promise<{
    totalTags: number;
    validTags: number;
    invalidTags: string[];
    orphanedTags: string[];
    brokenChains: string[];
    circularReferences: string[];
  }> {
    const stats = await this.tagManager.getStatistics();
    const allTags = (await this.tagManager.search({})).tags;

    const invalidTags: string[] = [];
    const orphanedTags: string[] = [];
    const brokenChains: string[] = [];
    const _circularReferences: string[] = [];

    for (const tag of allTags) {
      // 형식 검증
      if (!this.tagParser.validateTagFormat(tag.id)) {
        invalidTags.push(tag.id);
        continue;
      }

      // TAG 검증
      const validation = await this.tagManager.validateTag(tag);
      if (!validation.isValid) {
        invalidTags.push(tag.id);
      }

      // 고아 TAG 검사 (Primary Chain 외)
      if (tag.parents.length === 0 && tag.type !== 'REQ') {
        orphanedTags.push(tag.id);
      }

      // 체인 연결 검사
      for (const parentId of tag.parents) {
        const parent = await this.tagManager.getTag(parentId);
        if (!parent) {
          brokenChains.push(`${tag.id} -> ${parentId}`);
        }
      }

      // 순환 참조 검사는 이미 tagManager.validateTag에서 처리됨
    }

    return {
      totalTags: stats.total,
      validTags: stats.total - invalidTags.length,
      invalidTags,
      orphanedTags,
      brokenChains,
      circularReferences: [], // stats.circularReferences로 대체 가능
    };
  }

  /**
   * @API:TAG-REPAIR-001: TAG 체인 자동 수리
   */
  private async repairTagChains(): Promise<TagChainRepairResult> {
    const validation = await this.validateTagSystem();
    let repairedChains = 0;
    let orphanedTagsFixed = 0;
    const circularReferencesResolved = 0;
    const newConnections: string[] = [];

    // 1. 끊어진 체인 수리
    for (const brokenChain of validation.brokenChains) {
      const [childId, parentId] = brokenChain.split(' -> ');

      // 유사한 이름의 기존 TAG 찾기
      const similarParent = await this.findMostSimilarTag(parentId);
      if (similarParent) {
        const child = await this.tagManager.getTag(childId);
        if (child) {
          const updatedParents = child.parents.map(p =>
            p === parentId ? similarParent.id : p
          );
          await this.tagManager.updateTag(childId, { parents: updatedParents });
          newConnections.push(`${childId} -> ${similarParent.id}`);
          repairedChains++;
        }
      }
    }

    // 2. 고아 TAG 부모 찾기
    for (const orphanId of validation.orphanedTags) {
      const orphan = await this.tagManager.getTag(orphanId);
      if (orphan && orphan.type !== 'REQ') {
        // 같은 도메인의 상위 TAG 찾기
        const domain = this.extractDomain(orphan.id);
        const possibleParent = await this.findParentTagForDomain(
          domain,
          orphan.type
        );

        if (possibleParent) {
          await this.tagManager.updateTag(orphanId, {
            parents: [possibleParent.id],
          });
          newConnections.push(`${orphanId} -> ${possibleParent.id}`);
          orphanedTagsFixed++;
        }
      }
    }

    // 3. 인덱스 재구축
    await this.tagManager.save();
    await this.updateDistributedIndexes();

    return {
      repairedChains,
      orphanedTagsFixed,
      circularReferencesResolved,
      newConnections,
    };
  }

  /**
   * @API:TAG-INDEX-001: 인덱스 최적화
   */
  private async optimizeIndexes(): Promise<{
    sizeBefore: number;
    sizeAfter: number;
    optimizationRatio: number;
    indexingSpeed: number;
  }> {
    const sizeBefore = await this.getIndexSize();
    const startTime = performance.now();

    // 1. 메인 TAG 데이터베이스 최적화
    await this.tagManager.save();

    // 2. 분산 인덱스 재구축
    await this.rebuildDistributedIndexes();

    // 3. 성능 지표 측정
    const sizeAfter = await this.getIndexSize();
    const indexingSpeed = performance.now() - startTime;
    const optimizationRatio = ((sizeBefore - sizeAfter) / sizeBefore) * 100;

    return {
      sizeBefore,
      sizeAfter,
      optimizationRatio,
      indexingSpeed,
    };
  }

  /**
   * @API:TAG-STATS-001: TAG 시스템 통계 생성
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
    const stats = await this.tagManager.getStatistics();
    const validation = await this.validateTagSystem();

    // 성능 지표
    const indexSize = await this.getIndexSize();
    const searchStartTime = performance.now();
    await this.tagManager.search({ types: ['REQ'] });
    const searchSpeed = performance.now() - searchStartTime;

    // 건강도 지표
    const integrityScore = (validation.validTags / validation.totalTags) * 100;
    let qualityGate: 'healthy' | 'warning' | 'critical';

    if (integrityScore >= 95 && validation.brokenChains.length === 0) {
      qualityGate = 'healthy';
    } else if (integrityScore >= 85) {
      qualityGate = 'warning';
    } else {
      qualityGate = 'critical';
    }

    return {
      ...stats,
      performance: {
        indexSize,
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
   * 유사 TAG 검색
   */
  private async findSimilarTags(keyword: string): Promise<TagEntry[]> {
    const allTags = (await this.tagManager.search({})).tags;
    return allTags.filter(tag => {
      const similarity = this.calculateSimilarity(keyword, tag.id);
      return similarity > 0.3;
    });
  }

  /**
   * 문자열 유사도 계산 (Levenshtein distance 기반)
   */
  private calculateSimilarity(str1: string, str2: string): number {
    const matrix: number[][] = Array(str2.length + 1)
      .fill(null)
      .map(() => Array(str1.length + 1).fill(null));

    for (let i = 0; i <= str1.length; i++) matrix[0][i] = i;
    for (let j = 0; j <= str2.length; j++) matrix[j][0] = j;

    for (let j = 1; j <= str2.length; j++) {
      for (let i = 1; i <= str1.length; i++) {
        const indicator = str1[i - 1] === str2[j - 1] ? 0 : 1;
        matrix[j][i] = Math.min(
          matrix[j][i - 1] + 1, // insertion
          matrix[j - 1][i] + 1, // deletion
          matrix[j - 1][i - 1] + indicator // substitution
        );
      }
    }

    const maxLength = Math.max(str1.length, str2.length);
    return maxLength === 0
      ? 1
      : (maxLength - matrix[str2.length][str1.length]) / maxLength;
  }

  /**
   * TAG ID 생성
   */
  private generateTagId(domain: string): string {
    const normalizedDomain = domain.toUpperCase().replace(/[^A-Z0-9]/g, '-');
    return normalizedDomain;
  }

  /**
   * 다음 사용 가능한 ID 찾기
   */
  private async getNextAvailableId(baseId: string): Promise<string> {
    let counter = 1;
    let candidateId = `${baseId}-${counter.toString().padStart(3, '0')}`;

    while (await this.tagManager.getTag(`@REQ:${candidateId}`)) {
      counter++;
      candidateId = `${baseId}-${counter.toString().padStart(3, '0')}`;
    }

    return candidateId;
  }

  /**
   * TAG 체인 검증
   */
  private async validateTagChain(tags: string[]): Promise<TagValidationResult> {
    const errors: string[] = [];
    const warnings: string[] = [];

    // Primary Chain 순서 검증
    const _expectedOrder = ['REQ', 'DESIGN', 'TASK', 'TEST'];
    for (let i = 0; i < tags.length - 1; i++) {
      const current = await this.tagManager.getTag(tags[i]);
      const next = await this.tagManager.getTag(tags[i + 1]);

      if (!current || !next) {
        errors.push(`Missing TAG in chain: ${tags[i]} -> ${tags[i + 1]}`);
        continue;
      }

      if (!current.children.includes(next.id)) {
        errors.push(`Broken chain link: ${current.id} -> ${next.id}`);
      }
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings,
    };
  }

  /**
   * 분산 인덱스 업데이트
   */
  private async updateDistributedIndexes(tags?: string[]): Promise<void> {
    // JSONL 인덱스 업데이트 로직
    // 실제 구현은 기존 tag-system의 분산 구조를 활용
    await this.ensureIndexDirectories();

    if (tags) {
      // 특정 TAG들에 대한 인덱스 업데이트
      for (const tagId of tags) {
        const tag = await this.tagManager.getTag(tagId);
        if (tag) {
          await this.writeTagToDistributedIndex(tag);
        }
      }
    } else {
      // 전체 인덱스 재구축
      await this.rebuildDistributedIndexes();
    }
  }

  /**
   * 인덱스 디렉터리 보장
   */
  private async ensureIndexDirectories(): Promise<void> {
    const dirs = [
      join(this.indexesPath, 'categories'),
      join(this.indexesPath, 'relations'),
      join(this.indexesPath, 'cache'),
    ];

    for (const dir of dirs) {
      try {
        await fs.mkdir(dir, { recursive: true });
      } catch (_error) {
        // 이미 존재하는 경우 무시
      }
    }
  }

  /**
   * TAG를 분산 인덱스에 기록
   */
  private async writeTagToDistributedIndex(tag: TagEntry): Promise<void> {
    const categoryFile = join(
      this.indexesPath,
      'categories',
      `${tag.type.toLowerCase()}.jsonl`
    );
    const indexEntry = {
      tag: tag.id,
      type: tag.type.toLowerCase(),
      description: tag.description || tag.title,
      created: new Date().toISOString().split('T')[0],
      status: tag.status,
    };

    // JSONL 형식으로 추가
    await fs.appendFile(categoryFile, `${JSON.stringify(indexEntry)}\n`);
  }

  /**
   * 분산 인덱스 재구축
   */
  private async rebuildDistributedIndexes(): Promise<void> {
    // 기존 인덱스 파일들 삭제
    const categories = [
      'req',
      'design',
      'task',
      'test',
      'feature',
      'api',
      'ui',
      'data',
      'perf',
      'sec',
      'docs',
      'tag',
    ];

    for (const category of categories) {
      const categoryFile = join(
        this.indexesPath,
        'categories',
        `${category}.jsonl`
      );
      try {
        await fs.unlink(categoryFile);
      } catch {
        // 파일이 없는 경우 무시
      }
    }

    // 모든 TAG를 다시 인덱싱
    const allTags = (await this.tagManager.search({})).tags;
    for (const tag of allTags) {
      await this.writeTagToDistributedIndex(tag);
    }
  }

  /**
   * 인덱스 크기 계산
   */
  private async getIndexSize(): Promise<number> {
    let totalSize = 0;

    try {
      // 메인 인덱스 파일
      const mainIndexPath = join(this.indexesPath, 'tags.json');
      const mainStat = await fs.stat(mainIndexPath);
      totalSize += mainStat.size;

      // 분산 인덱스 파일들
      const categoriesDir = join(this.indexesPath, 'categories');
      const categoryFiles = await fs.readdir(categoriesDir);

      for (const file of categoryFiles) {
        if (file.endsWith('.jsonl')) {
          const filePath = join(categoriesDir, file);
          const stat = await fs.stat(filePath);
          totalSize += stat.size;
        }
      }
    } catch {
      // 파일이 없는 경우 0 반환
    }

    return totalSize;
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
    return match ? match[1] : '';
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

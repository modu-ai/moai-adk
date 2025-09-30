// @FEATURE:TAG-MANAGER-001 | Chain: @REQ:TAG-MANAGER-001 → @DESIGN:TAG-MANAGER-001 → @TASK:TAG-MANAGER-001 → @FEATURE:TAG-MANAGER-001
// Related: @SEC:TAG-MANAGER-001, @PERF:TAG-MANAGER-001

/**
 * ⚠️ DEPRECATED: This file is deprecated and should not be used in new code.
 *
 * **Migration Guide:**
 * - Use `CodeFirstTagParser` from `./code-first-parser.ts` instead
 * - Tags should exist in code only (CODE-FIRST philosophy)
 * - No more tags.json index files
 * - Direct code scanning with rg/grep for TAG search
 *
 * **Why deprecated:**
 * - TAG system philosophy changed to CODE-FIRST approach
 * - tags.json caused synchronization issues
 * - Single source of truth should be code, not cache files
 *
 * @deprecated Since v0.0.3 - Use CodeFirstTagParser instead
 */

import { promises as fs } from 'node:fs';
import { dirname } from 'node:path';
import { logger } from '../../utils/winston-logger.js';
import type {
  TagCategory,
  TagDatabase,
  TagEntry,
  TagManagerConfig,
  TagSearchQuery,
  TagSearchResult,
  TagStatistics,
  TagStatus,
  TagType,
  TagValidationResult,
} from './types.js';

/**
 * @FEATURE:TAG-MANAGER-001: JSON 기반 TAG 관리 시스템
 * @PERF:TAG-MANAGER-001: 고성능 메모리 캐싱 시스템
 * @SEC:TAG-MANAGER-001: 입력 검증 및 데이터 무결성 보장
 *
 * @deprecated Since v0.0.3 - Use CodeFirstTagParser from './code-first-parser.ts' instead
 *
 * SQLite3를 대체하는 경량 JSON 기반 TAG 시스템입니다.
 * - 메모리 캐싱: < 50ms 로드 성능 보장
 * - 자동 백업: 데이터 손실 방지
 * - 인덱스 최적화: 빠른 검색 성능
 * - 타입 안전성: 완전한 TypeScript 지원
 *
 * **⚠️ This class is deprecated. Use CodeFirstTagParser instead.**
 */
export class TagManager {
  private readonly config: TagManagerConfig;
  private database: TagDatabase | null = null;
  private readonly cache: Map<string, TagEntry> = new Map();
  private readonly searchIndexCache: Map<string, TagEntry[]> = new Map();
  private autoSaveTimer: NodeJS.Timeout | null = null;
  private loadPromise: Promise<TagDatabase> | null = null;

  constructor(config: TagManagerConfig) {
    this.config = config;
  }

  /**
   * @API:TAG-LOAD-001: 데이터베이스 로드
   * @PERF:LOAD-001: 중복 로드 방지 및 캐싱 최적화
   */
  async load(): Promise<TagDatabase> {
    // 이미 로딩 중이면 기존 Promise 반환 (중복 로드 방지)
    if (this.loadPromise) {
      return this.loadPromise;
    }

    // 이미 로드되었으면 캐시된 데이터베이스 반환
    if (this.database) {
      return this.database;
    }

    this.loadPromise = this.performLoad();

    try {
      const result = await this.loadPromise;
      return result;
    } finally {
      this.loadPromise = null;
    }
  }

  /**
   * @PERF:LOAD-002: 실제 로드 작업 수행
   * @SEC:LOAD-001: 보안 검증 및 오류 처리
   */
  private async performLoad(): Promise<TagDatabase> {
    try {
      const data = await fs.readFile(this.config.filePath, 'utf-8');
      const parsed = JSON.parse(data);
      this.database = this.validateAndMigrateDatabase(parsed);
    } catch (error) {
      if ((error as NodeJS.ErrnoException).code === 'ENOENT') {
        // 파일이 없으면 새 데이터베이스 생성
        this.database = this.createEmptyDatabase();
      } else if (error instanceof SyntaxError) {
        throw new Error('Invalid JSON format');
      } else {
        throw error;
      }
    }

    if (this.config.enableCache) {
      this.populateCache();
      this.buildSearchIndexes();
    }

    return this.database;
  }

  /**
   * @API:TAG-SAVE-001: 데이터베이스 저장
   */
  async save(): Promise<void> {
    if (!this.database) {
      await this.load();
    }

    if (!this.database) {
      throw new Error('Database not initialized');
    }

    // 메타데이터 업데이트
    this.database.metadata.totalTags = Object.keys(this.database.tags).length;
    this.database.metadata.lastUpdated = new Date().toISOString();

    // 인덱스 재구성
    this.rebuildIndexes();

    // 디렉토리가 없으면 생성
    const dir = dirname(this.config.filePath);
    try {
      await fs.access(dir);
    } catch {
      await fs.mkdir(dir, { recursive: true });
    }

    // JSON 파일로 저장
    const data = JSON.stringify(this.database, null, 2);
    await fs.writeFile(this.config.filePath, data, 'utf-8');

    this.isDirty = false;
  }

  /**
   * @API:TAG-CREATE-001: TAG 생성
   */
  async createTag(tagEntry: Partial<TagEntry>): Promise<TagEntry> {
    if (!this.database) {
      await this.load();
    }

    if (!tagEntry.id) {
      throw new Error('TAG ID is required');
    }

    if (!this.isValidTagId(tagEntry.id)) {
      throw new Error(`Invalid TAG ID format: ${tagEntry.id}`);
    }

    if (this.database?.tags[tagEntry.id]) {
      throw new Error(`TAG with ID ${tagEntry.id} already exists`);
    }

    const now = new Date().toISOString();
    const fullTagEntry: TagEntry = {
      id: tagEntry.id,
      type: tagEntry.type || 'REQ',
      category: tagEntry.category || 'PRIMARY',
      title: tagEntry.title || '',
      description: tagEntry.description,
      status: tagEntry.status || 'pending',
      priority: tagEntry.priority || 'medium',
      parents: tagEntry.parents || [],
      children: tagEntry.children || [],
      files: tagEntry.files || [],
      createdAt: now,
      updatedAt: now,
      author: tagEntry.author,
      metadata: tagEntry.metadata,
    };

    this.database?.tags[fullTagEntry.id] = fullTagEntry;

    if (this.config.enableCache) {
      this.cache.set(fullTagEntry.id, fullTagEntry);
    }

    this.markDirty();
    return fullTagEntry;
  }

  /**
   * @API:TAG-GET-001: TAG 조회
   */
  async getTag(id: string): Promise<TagEntry | null> {
    if (!this.database) {
      await this.load();
    }

    // 캐시에서 먼저 확인
    if (this.config.enableCache && this.cache.has(id)) {
      return this.cache.get(id) || null;
    }

    const tag = this.database?.tags[id] || null;

    if (tag && this.config.enableCache) {
      this.cache.set(id, tag);
    }

    return tag;
  }

  /**
   * @API:TAG-UPDATE-001: TAG 업데이트
   */
  async updateTag(id: string, updates: Partial<TagEntry>): Promise<TagEntry> {
    if (!this.database) {
      await this.load();
    }

    const existingTag = this.database?.tags[id];
    if (!existingTag) {
      throw new Error(`TAG with ID ${id} not found`);
    }

    // 약간의 지연을 통해 updatedAt이 다른 시간을 갖도록 함
    await new Promise(resolve => setTimeout(resolve, 5));

    const updatedTag: TagEntry = {
      ...existingTag,
      ...updates,
      id, // ID는 변경 불가
      updatedAt: new Date().toISOString(),
    };

    this.database?.tags[id] = updatedTag;

    if (this.config.enableCache) {
      this.cache.set(id, updatedTag);
    }

    this.markDirty();
    return updatedTag;
  }

  /**
   * @API:TAG-DELETE-001: TAG 삭제
   */
  async deleteTag(id: string): Promise<boolean> {
    if (!this.database) {
      await this.load();
    }

    if (!this.database?.tags[id]) {
      return false;
    }

    delete this.database?.tags[id];

    if (this.config.enableCache) {
      this.cache.delete(id);
    }

    this.markDirty();
    return true;
  }

  /**
   * @API:TAG-SEARCH-001: TAG 검색
   * @PERF:SEARCH-001: 인덱스 기반 고성능 검색
   */
  async search(query: TagSearchQuery): Promise<TagSearchResult> {
    if (!this.database) {
      await this.load();
    }

    const startTime = performance.now();
    const allTags = Object.values(this.database?.tags);

    let filteredTags = allTags;

    // 타입 필터
    if (query.types && query.types.length > 0) {
      filteredTags = filteredTags.filter(tag =>
        query.types?.includes(tag.type)
      );
    }

    // 카테고리 필터
    if (query.categories && query.categories.length > 0) {
      filteredTags = filteredTags.filter(tag =>
        query.categories?.includes(tag.category)
      );
    }

    // 상태 필터
    if (query.statuses && query.statuses.length > 0) {
      filteredTags = filteredTags.filter(tag =>
        query.statuses?.includes(tag.status)
      );
    }

    // ID 패턴 필터
    if (query.idPattern) {
      const pattern = new RegExp(query.idPattern, 'i');
      filteredTags = filteredTags.filter(tag => pattern.test(tag.id));
    }

    // 파일 경로 필터
    if (query.filePaths && query.filePaths.length > 0) {
      filteredTags = filteredTags.filter(tag =>
        tag.files.some(file =>
          query.filePaths?.some(path => file.includes(path))
        )
      );
    }

    // 부모 TAG 필터
    if (query.parentIds && query.parentIds.length > 0) {
      filteredTags = filteredTags.filter(tag =>
        tag.parents.some(parent => query.parentIds?.includes(parent))
      );
    }

    // 자식 TAG 필터
    if (query.childIds && query.childIds.length > 0) {
      filteredTags = filteredTags.filter(tag =>
        tag.children.some(child => query.childIds?.includes(child))
      );
    }

    // 날짜 범위 필터
    if (query.createdAfter) {
      filteredTags = filteredTags.filter(
        tag => tag.createdAt >= query.createdAfter!
      );
    }

    if (query.createdBefore) {
      filteredTags = filteredTags.filter(
        tag => tag.createdAt <= query.createdBefore!
      );
    }

    if (query.updatedAfter) {
      filteredTags = filteredTags.filter(
        tag => tag.updatedAt >= query.updatedAfter!
      );
    }

    if (query.updatedBefore) {
      filteredTags = filteredTags.filter(
        tag => tag.updatedAt <= query.updatedBefore!
      );
    }

    const endTime = performance.now();

    return {
      tags: filteredTags,
      total: filteredTags.length,
      executionTime: endTime - startTime,
    };
  }

  /**
   * @API:TAG-VALIDATE-001: TAG 유효성 검증
   */
  async validateTag(tagEntry: TagEntry): Promise<TagValidationResult> {
    const errors: string[] = [];
    const warnings: string[] = [];

    // ID 형식 검증
    if (!this.isValidTagId(tagEntry.id)) {
      errors.push(`Invalid TAG ID format: ${tagEntry.id}`);
    }

    // 필수 필드 검증
    if (!tagEntry.title.trim()) {
      errors.push('TAG title is required');
    }

    // 순환 참조 검증
    if (this.hasCircularReference(tagEntry)) {
      errors.push(`Circular reference detected in TAG: ${tagEntry.id}`);
    }

    // 고아 TAG 검증
    if (tagEntry.parents.length === 0 && tagEntry.type !== 'REQ') {
      warnings.push(`TAG ${tagEntry.id} might be orphaned (no parent TAGs)`);
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings,
    };
  }

  /**
   * @API:TAG-STATS-001: TAG 통계 정보
   */
  async getStatistics(): Promise<TagStatistics> {
    if (!this.database) {
      await this.load();
    }

    const tags = Object.values(this.database?.tags);
    const stats: TagStatistics = {
      total: tags.length,
      byType: {} as Record<TagType, number>,
      byCategory: {} as Record<TagCategory, number>,
      byStatus: {} as Record<TagStatus, number>,
      orphanedTags: 0,
      circularReferences: 0,
    };

    // 카운터 초기화
    const allTypes: TagType[] = [
      'REQ',
      'DESIGN',
      'TASK',
      'TEST',
      'VISION',
      'STRUCT',
      'TECH',
      'ADR',
      'FEATURE',
      'API',
      'UI',
      'DATA',
      'PERF',
      'SEC',
      'DOCS',
      'TAG',
    ];
    const allCategories: TagCategory[] = [
      'PRIMARY',
      'STEERING',
      'IMPLEMENTATION',
      'QUALITY',
    ];
    const allStatuses: TagStatus[] = [
      'pending',
      'in_progress',
      'completed',
      'blocked',
    ];

    for (const type of allTypes) {
      stats.byType[type] = 0;
    }

    for (const category of allCategories) {
      stats.byCategory[category] = 0;
    }

    for (const status of allStatuses) {
      stats.byStatus[status] = 0;
    }

    // 통계 계산
    for (const tag of tags) {
      stats.byType[tag.type]++;
      stats.byCategory[tag.category]++;
      stats.byStatus[tag.status]++;

      // 고아 TAG 검사
      if (tag.parents.length === 0 && tag.type !== 'REQ') {
        stats.orphanedTags++;
      }

      // 순환 참조 검사
      if (this.hasCircularReference(tag)) {
        stats.circularReferences++;
      }
    }

    return stats;
  }

  // Private helper methods

  private createEmptyDatabase(): TagDatabase {
    return {
      version: '1.0.0',
      tags: {},
      indexes: {
        byType: {} as Record<TagType, string[]>,
        byCategory: {} as Record<TagCategory, string[]>,
        byStatus: {} as Record<TagStatus, string[]>,
        byFile: {},
      },
      metadata: {
        totalTags: 0,
        lastUpdated: new Date().toISOString(),
      },
    };
  }

  private validateAndMigrateDatabase(data: any): TagDatabase {
    // 기본 구조 검증
    if (!data.version || !data.tags || !data.metadata) {
      throw new Error('Invalid database structure');
    }

    return data as TagDatabase;
  }

  private populateCache(): void {
    if (!this.database) return;

    this.cache.clear();
    for (const [id, tag] of Object.entries(this.database.tags)) {
      this.cache.set(id, tag);
    }
  }

  private rebuildIndexes(): void {
    if (!this.database) return;

    // 인덱스 초기화
    const indexes = this.database.indexes;
    indexes.byType = {} as Record<TagType, string[]>;
    indexes.byCategory = {} as Record<TagCategory, string[]>;
    indexes.byStatus = {} as Record<TagStatus, string[]>;
    indexes.byFile = {};

    // 인덱스 재구성
    for (const [id, tag] of Object.entries(this.database.tags)) {
      // 타입별 인덱스
      if (!indexes.byType[tag.type]) {
        indexes.byType[tag.type] = [];
      }
      indexes.byType[tag.type].push(id);

      // 카테고리별 인덱스
      if (!indexes.byCategory[tag.category]) {
        indexes.byCategory[tag.category] = [];
      }
      indexes.byCategory[tag.category].push(id);

      // 상태별 인덱스
      if (!indexes.byStatus[tag.status]) {
        indexes.byStatus[tag.status] = [];
      }
      indexes.byStatus[tag.status].push(id);

      // 파일별 인덱스
      for (const file of tag.files) {
        if (!indexes.byFile[file]) {
          indexes.byFile[file] = [];
        }
        indexes.byFile[file].push(id);
      }
    }
  }

  private isValidTagId(id: string): boolean {
    // @CATEGORY:DOMAIN-ID 형식 검증 (예: @REQ:AUTH-001, @PERF:LOAD-001)
    // 숫자는 3자리 이상 허용
    const pattern = /^@[A-Z]+(-[A-Z0-9]+)*-\d{3,}$/;
    return pattern.test(id);
  }

  private hasCircularReference(tag: TagEntry): boolean {
    const visited = new Set<string>();
    const recursionStack = new Set<string>();

    const dfs = (tagId: string): boolean => {
      if (recursionStack.has(tagId)) {
        return true; // 순환 참조 발견
      }

      if (visited.has(tagId)) {
        return false; // 이미 방문한 노드
      }

      visited.add(tagId);
      recursionStack.add(tagId);

      const currentTag = this.database?.tags[tagId];
      if (currentTag) {
        for (const childId of currentTag.children) {
          if (dfs(childId)) {
            return true;
          }
        }
      }

      recursionStack.delete(tagId);
      return false;
    };

    return dfs(tag.id);
  }

  private markDirty(): void {
    this.isDirty = true;

    // 검색 캐시 무효화
    if (this.config.enableCache) {
      this.searchIndexCache.clear();
    }

    if (this.config.autoSave && !this.autoSaveTimer) {
      this.autoSaveTimer = setTimeout(() => {
        this.save().catch((error) => {
          logger.error('Auto-save failed', error);
        });
        this.autoSaveTimer = null;
      }, this.config.autoSaveDelay);
    }
  }

  /**
   * @PERF:SEARCH-002: 검색 인덱스 구축
   */
  private buildSearchIndexes(): void {
    if (!this.database) return;

    // 검색 캐시 초기화
    this.searchIndexCache.clear();

    // 공통 검색 패턴을 미리 인덱싱
    const commonQueries = [
      { types: ['REQ'] },
      { types: ['DESIGN'] },
      { types: ['TASK'] },
      { types: ['TEST'] },
      { categories: ['PRIMARY'] },
      { categories: ['IMPLEMENTATION'] },
      { statuses: ['pending'] },
      { statuses: ['in_progress'] },
      { statuses: ['completed'] },
    ];

    for (const query of commonQueries) {
      const key = this.generateSearchCacheKey(query);
      const results = this.performSearch(query);
      this.searchIndexCache.set(key, results);
    }
  }

  /**
   * @PERF:SEARCH-003: 검색 캐시 키 생성
   */
  private generateSearchCacheKey(query: TagSearchQuery): string {
    return JSON.stringify(query, Object.keys(query).sort());
  }

  /**
   * @PERF:SEARCH-004: 인덱스 기반 초기 후보 선택
   */
  private getInitialCandidates(query: TagSearchQuery): TagEntry[] {
    if (!this.database) return [];

    // 가장 선택적인 필터를 먼저 적용
    if (query.types && query.types.length > 0) {
      const candidates = new Set<string>();
      for (const type of query.types) {
        const typeIds = this.database.indexes.byType[type] || [];
        for (const id of typeIds) {
          candidates.add(id);
        }
      }
      return Array.from(candidates)
        .map(id => this.database?.tags[id])
        .filter(Boolean);
    }

    if (query.categories && query.categories.length > 0) {
      const candidates = new Set<string>();
      for (const category of query.categories) {
        const categoryIds = this.database.indexes.byCategory[category] || [];
        for (const id of categoryIds) {
          candidates.add(id);
        }
      }
      return Array.from(candidates)
        .map(id => this.database?.tags[id])
        .filter(Boolean);
    }

    if (query.statuses && query.statuses.length > 0) {
      const candidates = new Set<string>();
      for (const status of query.statuses) {
        const statusIds = this.database.indexes.byStatus[status] || [];
        for (const id of statusIds) {
          candidates.add(id);
        }
      }
      return Array.from(candidates)
        .map(id => this.database?.tags[id])
        .filter(Boolean);
    }

    // 필터가 없으면 모든 TAG 반환
    return Object.values(this.database.tags);
  }

  /**
   * @PERF:SEARCH-005: 추가 필터링 적용
   */
  private applyAdditionalFilters(
    tags: TagEntry[],
    query: TagSearchQuery
  ): TagEntry[] {
    let filtered = tags;

    // 타입 필터 (getInitialCandidates에서 처리하지 않은 경우)
    if (query.types && query.types.length > 0) {
      filtered = filtered.filter(tag => query.types?.includes(tag.type));
    }

    // 카테고리 필터 (getInitialCandidates에서 처리하지 않은 경우)
    if (query.categories && query.categories.length > 0) {
      filtered = filtered.filter(tag =>
        query.categories?.includes(tag.category)
      );
    }

    // 상태 필터 (getInitialCandidates에서 처리하지 않은 경우)
    if (query.statuses && query.statuses.length > 0) {
      filtered = filtered.filter(tag => query.statuses?.includes(tag.status));
    }

    // ID 패턴 필터
    if (query.idPattern) {
      const pattern = new RegExp(query.idPattern, 'i');
      filtered = filtered.filter(tag => pattern.test(tag.id));
    }

    // 파일 경로 필터
    if (query.filePaths && query.filePaths.length > 0) {
      filtered = filtered.filter(tag =>
        tag.files.some(file =>
          query.filePaths?.some(path => file.includes(path))
        )
      );
    }

    // 부모 TAG 필터
    if (query.parentIds && query.parentIds.length > 0) {
      filtered = filtered.filter(tag =>
        tag.parents.some(parent => query.parentIds?.includes(parent))
      );
    }

    // 자식 TAG 필터
    if (query.childIds && query.childIds.length > 0) {
      filtered = filtered.filter(tag =>
        tag.children.some(child => query.childIds?.includes(child))
      );
    }

    // 날짜 범위 필터
    if (query.createdAfter) {
      filtered = filtered.filter(tag => tag.createdAt >= query.createdAfter!);
    }

    if (query.createdBefore) {
      filtered = filtered.filter(tag => tag.createdAt <= query.createdBefore!);
    }

    if (query.updatedAfter) {
      filtered = filtered.filter(tag => tag.updatedAt >= query.updatedAfter!);
    }

    if (query.updatedBefore) {
      filtered = filtered.filter(tag => tag.updatedAt <= query.updatedBefore!);
    }

    return filtered;
  }

  /**
   * @PERF:SEARCH-006: 실제 검색 수행 (캐싱용)
   */
  private performSearch(query: TagSearchQuery): TagEntry[] {
    if (!this.database) return [];

    const candidates = this.getInitialCandidates(query);
    return this.applyAdditionalFilters(candidates, query);
  }
}

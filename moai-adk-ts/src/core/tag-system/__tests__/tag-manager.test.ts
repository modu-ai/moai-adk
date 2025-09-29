// @TEST-TAG-MANAGER-001: TagManager 핵심 기능 테스트
// 연결: @REQ-TAG-JSON-001 → @DESIGN-TAG-TYPES-001 → @TASK-TAG-MANAGER-001

import { promises as fs } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { TagManager } from '../tag-manager.js';
import type {
  TagCategory,
  TagEntry,
  TagManagerConfig,
  TagSearchQuery,
  TagStatus,
  TagType,
} from '../types.js';

describe('@TEST-TAG-MANAGER-001: TagManager Core Functionality', () => {
  let tagManager: TagManager;
  let tempFilePath: string;
  let config: TagManagerConfig;

  beforeEach(async () => {
    // @TEST-SETUP-001: 임시 테스트 환경 설정
    const tempDir = await fs.mkdtemp(join(tmpdir(), 'tag-manager-test-'));
    tempFilePath = join(tempDir, 'tags.json');

    config = {
      filePath: tempFilePath,
      autoSave: false,
      autoSaveDelay: 1000,
      enableBackup: false,
      maxBackups: 5,
      enableCache: true,
      cacheTtl: 300000, // 5분
    };

    tagManager = new TagManager(config);
  });

  afterEach(async () => {
    // @TEST-CLEANUP-001: 테스트 환경 정리
    try {
      await fs.unlink(tempFilePath);
    } catch {
      // 파일이 없을 수 있음
    }
  });

  describe('@TEST-HAPPY-001: Happy Path Scenarios', () => {
    it('@TEST-LOAD-EMPTY-001: should initialize with empty database when file does not exist', async () => {
      // @TEST-LOAD-EMPTY-001: 파일이 없을 때 빈 데이터베이스로 초기화
      const database = await tagManager.load();

      expect(database.version).toBe('1.0.0');
      expect(database.tags).toEqual({});
      expect(database.metadata.totalTags).toBe(0);
      expect(Object.keys(database.tags)).toHaveLength(0);
    });

    it('@TEST-CREATE-TAG-001: should create a new TAG entry', async () => {
      // @TEST-CREATE-TAG-001: 새로운 TAG 엔트리 생성
      const tagEntry: Partial<TagEntry> = {
        id: '@REQ-AUTH-001',
        type: 'REQ' as TagType,
        category: 'PRIMARY' as TagCategory,
        title: '사용자 인증 요구사항',
        description: '사용자 로그인 및 권한 관리 요구사항',
        status: 'pending' as TagStatus,
        priority: 'high',
        parents: [],
        children: [],
        files: ['src/auth/login.ts'],
      };

      const createdTag = await tagManager.createTag(tagEntry);

      expect(createdTag.id).toBe('@REQ-AUTH-001');
      expect(createdTag.type).toBe('REQ');
      expect(createdTag.category).toBe('PRIMARY');
      expect(createdTag.title).toBe('사용자 인증 요구사항');
      expect(createdTag.status).toBe('pending');
      expect(createdTag.createdAt).toBeDefined();
      expect(createdTag.updatedAt).toBeDefined();
    });

    it('@TEST-GET-TAG-001: should retrieve an existing TAG by ID', async () => {
      // @TEST-GET-TAG-001: 기존 TAG를 ID로 조회
      const tagEntry: Partial<TagEntry> = {
        id: '@DESIGN-AUTH-001',
        type: 'DESIGN' as TagType,
        category: 'PRIMARY' as TagCategory,
        title: '인증 시스템 설계',
      };

      await tagManager.createTag(tagEntry);
      const retrievedTag = await tagManager.getTag('@DESIGN-AUTH-001');

      expect(retrievedTag).toBeDefined();
      expect(retrievedTag?.id).toBe('@DESIGN-AUTH-001');
      expect(retrievedTag?.type).toBe('DESIGN');
      expect(retrievedTag?.title).toBe('인증 시스템 설계');
    });

    it('@TEST-UPDATE-TAG-001: should update an existing TAG', async () => {
      // @TEST-UPDATE-TAG-001: 기존 TAG 업데이트
      const tagEntry: Partial<TagEntry> = {
        id: '@TASK-AUTH-001',
        type: 'TASK' as TagType,
        category: 'PRIMARY' as TagCategory,
        title: '인증 구현 작업',
        status: 'pending' as TagStatus,
      };

      await tagManager.createTag(tagEntry);

      const updatedTag = await tagManager.updateTag('@TASK-AUTH-001', {
        status: 'in_progress' as TagStatus,
        title: '인증 구현 작업 (진행중)',
      });

      expect(updatedTag.status).toBe('in_progress');
      expect(updatedTag.title).toBe('인증 구현 작업 (진행중)');
      expect(updatedTag.updatedAt).not.toBe(updatedTag.createdAt);
    });

    it('@TEST-DELETE-TAG-001: should delete an existing TAG', async () => {
      // @TEST-DELETE-TAG-001: 기존 TAG 삭제
      const tagEntry: Partial<TagEntry> = {
        id: '@TEST-AUTH-001',
        type: 'TEST' as TagType,
        category: 'PRIMARY' as TagCategory,
        title: '인증 테스트',
      };

      await tagManager.createTag(tagEntry);
      const deleted = await tagManager.deleteTag('@TEST-AUTH-001');

      expect(deleted).toBe(true);

      const retrievedTag = await tagManager.getTag('@TEST-AUTH-001');
      expect(retrievedTag).toBeNull();
    });

    it('@TEST-SAVE-LOAD-001: should save and load database to/from JSON file', async () => {
      // @TEST-SAVE-LOAD-001: JSON 파일로 저장 및 로드
      const tagEntry: Partial<TagEntry> = {
        id: '@FEATURE-AUTH-001',
        type: 'FEATURE' as TagType,
        category: 'IMPLEMENTATION' as TagCategory,
        title: '인증 기능 구현',
      };

      await tagManager.createTag(tagEntry);
      await tagManager.save();

      // 새로운 TagManager 인스턴스로 로드 테스트
      const newTagManager = new TagManager(config);
      const loadedDatabase = await newTagManager.load();

      expect(loadedDatabase.tags['@FEATURE-AUTH-001']).toBeDefined();
      expect(loadedDatabase.tags['@FEATURE-AUTH-001'].title).toBe(
        '인증 기능 구현'
      );
      expect(loadedDatabase.metadata.totalTags).toBe(1);
    });
  });

  describe('@TEST-SEARCH-001: Search Functionality', () => {
    beforeEach(async () => {
      // @TEST-SEARCH-SETUP-001: 검색 테스트용 샘플 데이터 생성
      const sampleTags: Partial<TagEntry>[] = [
        {
          id: '@REQ-USER-001',
          type: 'REQ',
          category: 'PRIMARY',
          title: '사용자 관리 요구사항',
          status: 'completed',
        },
        {
          id: '@DESIGN-USER-001',
          type: 'DESIGN',
          category: 'PRIMARY',
          title: '사용자 시스템 설계',
          status: 'in_progress',
        },
        {
          id: '@API-USER-001',
          type: 'API',
          category: 'IMPLEMENTATION',
          title: '사용자 API',
          status: 'pending',
        },
      ];

      for (const tag of sampleTags) {
        await tagManager.createTag(tag);
      }
    });

    it('@TEST-SEARCH-BY-TYPE-001: should search TAGs by type', async () => {
      // @TEST-SEARCH-BY-TYPE-001: 타입별 TAG 검색
      const query: TagSearchQuery = {
        types: ['REQ', 'DESIGN'],
      };

      const result = await tagManager.search(query);

      expect(result.tags).toHaveLength(2);
      expect(result.total).toBe(2);
      expect(result.tags.map(tag => tag.type)).toEqual(
        expect.arrayContaining(['REQ', 'DESIGN'])
      );
    });

    it('@TEST-SEARCH-BY-CATEGORY-001: should search TAGs by category', async () => {
      // @TEST-SEARCH-BY-CATEGORY-001: 카테고리별 TAG 검색
      const query: TagSearchQuery = {
        categories: ['PRIMARY'],
      };

      const result = await tagManager.search(query);

      expect(result.tags).toHaveLength(2);
      expect(result.tags.every(tag => tag.category === 'PRIMARY')).toBe(true);
    });

    it('@TEST-SEARCH-BY-STATUS-001: should search TAGs by status', async () => {
      // @TEST-SEARCH-BY-STATUS-001: 상태별 TAG 검색
      const query: TagSearchQuery = {
        statuses: ['completed', 'in_progress'],
      };

      const result = await tagManager.search(query);

      expect(result.tags).toHaveLength(2);
      expect(result.tags.map(tag => tag.status)).toEqual(
        expect.arrayContaining(['completed', 'in_progress'])
      );
    });
  });

  describe('@TEST-PERFORMANCE-001: Performance Requirements', () => {
    it('@TEST-PERF-LOAD-001: should load database in under 50ms', async () => {
      // @TEST-PERF-LOAD-001: 50ms 이내 데이터베이스 로드 성능 요구사항
      const startTime = performance.now();
      await tagManager.load();
      const endTime = performance.now();

      const loadTime = endTime - startTime;
      expect(loadTime).toBeLessThan(50); // 50ms 미만
    });

    it('@TEST-PERF-SEARCH-001: should search 1000 TAGs in under 50ms', async () => {
      // @TEST-PERF-SEARCH-001: 1000개 TAG 검색을 50ms 이내에 처리
      // 1000개 TAG 생성
      const tags: Partial<TagEntry>[] = [];
      for (let i = 1; i <= 1000; i++) {
        tags.push({
          id: `@REQ-PERF-${i.toString().padStart(3, '0')}`,
          type: 'REQ',
          category: 'PRIMARY',
          title: `Performance Test Requirement ${i}`,
          status: 'pending',
        });
      }

      for (const tag of tags) {
        await tagManager.createTag(tag);
      }

      const startTime = performance.now();
      const result = await tagManager.search({ types: ['REQ'] });
      const endTime = performance.now();

      const searchTime = endTime - startTime;
      expect(searchTime).toBeLessThan(50); // 50ms 미만
      expect(result.total).toBe(1000);
    });
  });

  describe('@TEST-EDGE-001: Edge Cases', () => {
    it('@TEST-EDGE-DUPLICATE-001: should handle duplicate TAG ID creation', async () => {
      // @TEST-EDGE-DUPLICATE-001: 중복 TAG ID 생성 처리
      const tagEntry: Partial<TagEntry> = {
        id: '@REQ-DUPLICATE-001',
        type: 'REQ',
        category: 'PRIMARY',
        title: '중복 테스트',
      };

      await tagManager.createTag(tagEntry);

      await expect(tagManager.createTag(tagEntry)).rejects.toThrow(
        'TAG with ID @REQ-DUPLICATE-001 already exists'
      );
    });

    it('@TEST-EDGE-INVALID-ID-001: should handle invalid TAG ID format', async () => {
      // @TEST-EDGE-INVALID-ID-001: 잘못된 TAG ID 형식 처리
      const invalidTagEntry: Partial<TagEntry> = {
        id: 'INVALID-TAG-ID',
        type: 'REQ',
        category: 'PRIMARY',
        title: '잘못된 형식',
      };

      await expect(tagManager.createTag(invalidTagEntry)).rejects.toThrow(
        'Invalid TAG ID format'
      );
    });

    it('@TEST-EDGE-NONEXISTENT-001: should handle operations on non-existent TAGs', async () => {
      // @TEST-EDGE-NONEXISTENT-001: 존재하지 않는 TAG 작업 처리
      const result = await tagManager.getTag('@NONEXISTENT-001');
      expect(result).toBeNull();

      await expect(
        tagManager.updateTag('@NONEXISTENT-001', { status: 'completed' })
      ).rejects.toThrow('TAG with ID @NONEXISTENT-001 not found');

      const deleteResult = await tagManager.deleteTag('@NONEXISTENT-001');
      expect(deleteResult).toBe(false);
    });
  });

  describe('@TEST-ERROR-001: Error Handling', () => {
    it('@TEST-ERROR-FILE-PERMISSION-001: should handle file permission errors', async () => {
      // @TEST-ERROR-FILE-PERMISSION-001: 파일 권한 오류 처리
      const readOnlyConfig = {
        ...config,
        filePath: '/root/readonly-tags.json', // 권한 없는 경로
      };

      const readOnlyTagManager = new TagManager(readOnlyConfig);

      await expect(readOnlyTagManager.save()).rejects.toThrow();
    });

    it('@TEST-ERROR-MALFORMED-JSON-001: should handle malformed JSON files', async () => {
      // @TEST-ERROR-MALFORMED-JSON-001: 잘못된 JSON 파일 처리
      await fs.writeFile(tempFilePath, '{ invalid json }');

      await expect(tagManager.load()).rejects.toThrow('Invalid JSON format');
    });
  });

  describe('@TEST-MEMORY-CACHE-001: Memory Cache Functionality', () => {
    it('@TEST-CACHE-HIT-001: should use memory cache for repeated reads', async () => {
      // @TEST-CACHE-HIT-001: 반복 읽기 시 메모리 캐시 활용
      const tagEntry: Partial<TagEntry> = {
        id: '@CACHE-TEST-001',
        type: 'TEST',
        category: 'QUALITY',
        title: '캐시 테스트',
      };

      await tagManager.createTag(tagEntry);

      // 첫 번째 조회 (캐시 미스)
      const firstRead = await tagManager.getTag('@CACHE-TEST-001');

      // 두 번째 조회 (캐시 히트)
      const secondRead = await tagManager.getTag('@CACHE-TEST-001');

      expect(firstRead).toEqual(secondRead);
      expect(firstRead?.id).toBe('@CACHE-TEST-001');
    });

    it('@TEST-CACHE-INVALIDATION-001: should invalidate cache on updates', async () => {
      // @TEST-CACHE-INVALIDATION-001: 업데이트 시 캐시 무효화
      const tagEntry: Partial<TagEntry> = {
        id: '@CACHE-INVALIDATE-001',
        type: 'TASK',
        category: 'PRIMARY',
        title: '캐시 무효화 테스트',
        status: 'pending',
      };

      await tagManager.createTag(tagEntry);
      await tagManager.getTag('@CACHE-INVALIDATE-001'); // 캐시에 저장

      // 업데이트 (캐시 무효화)
      await tagManager.updateTag('@CACHE-INVALIDATE-001', {
        status: 'completed',
      });

      const updatedTag = await tagManager.getTag('@CACHE-INVALIDATE-001');
      expect(updatedTag?.status).toBe('completed');
    });
  });
});

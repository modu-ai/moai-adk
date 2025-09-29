/**
 * @API:TAG-SYSTEM-001  TAG System Module Exports
 * @FEATURE:TAG-UNIFIED-001 통합 TAG 시스템 진입점
 *
 * TAG 시스템 통합 모듈 export
 * Python tag_system/__init__.py의 완전 포팅
 *
 * @TASK:TAG-EXPORTS-001 TAG 시스템 모듈 export 관리
 * @DESIGN:MODULE-STRUCTURE-001 모듈 구조 설계
 * @API:TAG-UNIFIED-001 통합 TAG API 구조 정의
 */

// Core TAG 파싱 시스템
export {
  TagParser,
  TagCategory,
  TagMatch,
  TagPosition,
  TagChain,
  DuplicateTagInfo,
} from './tag-parser';

// TAG 검증 시스템
export {
  TagValidator,
  ChainValidationResult,
  ValidationError,
  BrokenReference,
  ConsistencyViolation,
} from './tag-validator';

// TAG 데이터베이스 시스템 (임시 비활성화 - tag-database 모듈 누락)
// export {
//   TagDatabase,
//   TagSearchFilter,
//   TagRecord,
//   IndexingResult,
//   PrimaryChainAnalysis,
//   TagStatistics,
//   TableSchema,
//   FileIndexRequest,
//   getDefaultTagDatabase,
//   createTagDatabase,
// } from './tag-database';

// 편의 함수들 (tag-database 비활성화)
export function createTagSystem(_dbPath?: string) {
  const parser = new TagParser();
  const validator = new TagValidator();
  // const database = dbPath ? createTagDatabase(dbPath) : getDefaultTagDatabase();

  return {
    parser,
    validator,
    // database,
    async initialize() {
      // await database.initialize();
    },
    async close() {
      // await database.close();
    },
  };
}

/**
 *  TAG 카테고리 상수
 */
export const TAG_CATEGORIES = {
  PRIMARY: ['REQ', 'DESIGN', 'TASK', 'TEST'],
  STEERING: ['VISION', 'STRUCT', 'TECH', 'ADR'],
  IMPLEMENTATION: ['FEATURE', 'API', 'UI', 'DATA'],
  QUALITY: ['PERF', 'SEC', 'DOCS', 'TAG'],
} as const;

/**
 * 기본 TAG 시스템 인스턴스
 * 즉시 사용 가능한 편의 export
 */
export const defaultTagSystem = createTagSystem();

// @FEATURE-TAG-TYPES-001: JSON 기반 TAG 시스템 타입 정의
// 연결: @REQ-TAG-JSON-001 → @DESIGN-TAG-TYPES-001 → @TASK-TAG-TYPES-001

/**
 * 16-Core TAG 카테고리 정의
 */
export type TagCategory =
  | 'PRIMARY'     // @REQ, @DESIGN, @TASK, @TEST
  | 'STEERING'    // @VISION, @STRUCT, @TECH, @ADR
  | 'IMPLEMENTATION' // @FEATURE, @API, @UI, @DATA
  | 'QUALITY';    // @PERF, @SEC, @DOCS, @TAG

/**
 * TAG 타입 정의 (16-Core 시스템)
 */
export type TagType =
  // Primary Chain
  | 'REQ' | 'DESIGN' | 'TASK' | 'TEST'
  // Steering
  | 'VISION' | 'STRUCT' | 'TECH' | 'ADR'
  // Implementation
  | 'FEATURE' | 'API' | 'UI' | 'DATA'
  // Quality
  | 'PERF' | 'SEC' | 'DOCS' | 'TAG';

/**
 * TAG 상태 정의
 */
export type TagStatus = 'pending' | 'in_progress' | 'completed' | 'blocked';

/**
 * TAG 우선순위 정의
 */
export type TagPriority = 'critical' | 'high' | 'medium' | 'low';

/**
 * 개별 TAG 엔트리 정의
 */
export interface TagEntry {
  /** TAG ID (예: @REQ-AUTH-001) */
  id: string;

  /** TAG 타입 */
  type: TagType;

  /** TAG 카테고리 */
  category: TagCategory;

  /** TAG 제목/설명 */
  title: string;

  /** 상세 설명 */
  description?: string;

  /** 현재 상태 */
  status: TagStatus;

  /** 우선순위 */
  priority: TagPriority;

  /** 부모 TAG ID들 */
  parents: string[];

  /** 자식 TAG ID들 */
  children: string[];

  /** 관련 파일 경로들 */
  files: string[];

  /** 생성일시 */
  createdAt: string;

  /** 수정일시 */
  updatedAt: string;

  /** 작성자 */
  author?: string;

  /** 추가 메타데이터 */
  metadata?: Record<string, unknown>;
}

/**
 * TAG 데이터베이스 스키마 (JSON 파일)
 */
export interface TagDatabase {
  /** 버전 정보 */
  version: string;

  /** TAG 엔트리 맵 (TAG ID → TagEntry) */
  tags: Record<string, TagEntry>;

  /** 인덱스 정보 */
  indexes: {
    /** 타입별 인덱스 */
    byType: Record<TagType, string[]>;

    /** 카테고리별 인덱스 */
    byCategory: Record<TagCategory, string[]>;

    /** 상태별 인덱스 */
    byStatus: Record<TagStatus, string[]>;

    /** 파일별 인덱스 */
    byFile: Record<string, string[]>;
  };

  /** 메타데이터 */
  metadata: {
    /** 총 TAG 개수 */
    totalTags: number;

    /** 마지막 업데이트 시간 */
    lastUpdated: string;

    /** 체크섬 (무결성 검증용) */
    checksum?: string;
  };
}

/**
 * TAG 검색 쿼리 조건
 */
export interface TagSearchQuery {
  /** TAG ID 패턴 */
  idPattern?: string;

  /** TAG 타입 필터 */
  types?: TagType[];

  /** 카테고리 필터 */
  categories?: TagCategory[];

  /** 상태 필터 */
  statuses?: TagStatus[];

  /** 파일 경로 필터 */
  filePaths?: string[];

  /** 부모 TAG 필터 */
  parentIds?: string[];

  /** 자식 TAG 필터 */
  childIds?: string[];

  /** 생성일 범위 */
  createdAfter?: string;
  createdBefore?: string;

  /** 수정일 범위 */
  updatedAfter?: string;
  updatedBefore?: string;
}

/**
 * TAG 검색 결과
 */
export interface TagSearchResult {
  /** 검색된 TAG들 */
  tags: TagEntry[];

  /** 총 결과 개수 */
  total: number;

  /** 검색 실행 시간 (밀리초) */
  executionTime: number;
}

/**
 * TAG 유효성 검증 결과
 */
export interface TagValidationResult {
  /** 유효성 여부 */
  isValid: boolean;

  /** 오류 메시지들 */
  errors: string[];

  /** 경고 메시지들 */
  warnings: string[];
}

/**
 * TAG 관리 설정
 */
export interface TagManagerConfig {
  /** JSON 파일 경로 */
  filePath: string;

  /** 자동 저장 활성화 */
  autoSave: boolean;

  /** 자동 저장 지연시간 (밀리초) */
  autoSaveDelay: number;

  /** 백업 활성화 */
  enableBackup: boolean;

  /** 백업 파일 개수 */
  maxBackups: number;

  /** 메모리 캐시 활성화 */
  enableCache: boolean;

  /** 캐시 TTL (밀리초) */
  cacheTtl: number;
}

/**
 * TAG 통계 정보
 */
export interface TagStatistics {
  /** 총 TAG 개수 */
  total: number;

  /** 타입별 통계 */
  byType: Record<TagType, number>;

  /** 카테고리별 통계 */
  byCategory: Record<TagCategory, number>;

  /** 상태별 통계 */
  byStatus: Record<TagStatus, number>;

  /** 고아 TAG 개수 (부모-자식 연결이 끊어진 TAG) */
  orphanedTags: number;

  /** 순환 참조 TAG 개수 */
  circularReferences: number;
}
/**
 * @TAG:API:CODE-FIRST-TYPES-001
 * @CHAIN: REQ:TAG-SYSTEM-001 -> DESIGN:CODE-FIRST-001 -> TASK:TYPES-001 -> @TAG:API:CODE-FIRST-TYPES-001
 * @DEPENDS: NONE
 * @STATUS: active
 * @CREATED: 2025-09-29
 * @IMMUTABLE
 */

/**
 * Code-First TAG 시스템 타입 정의
 *
 * 핵심 원칙:
 * - 코드가 유일한 진실의 원천
 * - TAG는 파일 주석에만 존재
 * - @IMMUTABLE 마커로 불변성 보장
 * - 8-Core 단순화 (기존 16-Core 대비)
 */

/**
 * 8-Core TAG 카테고리 (16-Core에서 단순화)
 */
export enum TagCategory {
  // Lifecycle (생명주기 - 필수 체인)
  SPEC = "SPEC",        // 명세 작성
  REQ = "REQ",          // 요구사항 정의
  DESIGN = "DESIGN",    // 아키텍처 설계
  TASK = "TASK",        // 구현 작업
  TEST = "TEST",        // 테스트 검증

  // Implementation (구현 - 선택적)
  FEATURE = "FEATURE",  // 비즈니스 기능
  API = "API",         // 인터페이스
  FIX = "FIX"          // 버그 수정
}

/**
 * TAG 상태
 */
export enum TagStatus {
  ACTIVE = "active",        // 활성 상태
  DEPRECATED = "deprecated", // 폐기 예정
  COMPLETED = "completed"   // 완료 상태
}

/**
 * Code-First TAG 블록 구조
 *
 * 예시:
 * ```
 * /**
 *  * @TAG:FEATURE:AUTH-001
 *  * @CHAIN: REQ:AUTH-001 -> DESIGN:AUTH-001 -> TASK:AUTH-001 -> TEST:AUTH-001
 *  * @DEPENDS: FEATURE:USER-001, API:SESSION-001
 *  * @STATUS: active
 *  * @CREATED: 2025-09-29
 *  * @IMMUTABLE
 *  *\/
 * ```
 */
export interface TagBlock {
  /** 메인 TAG ID (@TAG:CATEGORY:DOMAIN-ID) */
  readonly tag: string;

  /** TAG 카테고리 */
  readonly category: TagCategory;

  /** 도메인 ID (예: AUTH-001) */
  readonly domainId: string;

  /** TAG 체인 (REQ -> DESIGN -> TASK -> TEST) */
  readonly chain?: string[];

  /** 의존성 TAG들 */
  readonly depends?: string[];

  /** TAG 상태 */
  readonly status: TagStatus;

  /** 생성 날짜 */
  readonly created: string;

  /** 불변성 마커 */
  readonly immutable: boolean;

  /** 파일 경로 */
  readonly filePath: string;

  /** 파일 내 라인 번호 */
  readonly lineNumber: number;
}

/**
 * TAG 체인 구조 (Primary Chain)
 */
export interface TagChain {
  /** 체인 ID */
  readonly chainId: string;

  /** 체인을 구성하는 TAG들 (순서 보장) */
  readonly tags: string[];

  /** 체인 완성도 (0-100%) */
  readonly completeness: number;
}

/**
 * TAG 의존성 그래프
 */
export interface TagDependencyGraph {
  /** 전체 TAG 목록 */
  readonly tags: Map<string, TagBlock>;

  /** 의존성 관계 (TAG -> 의존하는 TAG들) */
  readonly dependencies: Map<string, string[]>;

  /** 역 의존성 관계 (TAG -> 이 TAG에 의존하는 TAG들) */
  readonly dependents: Map<string, string[]>;
}

/**
 * TAG 검색 결과
 */
export interface TagSearchResult {
  /** 매칭된 TAG */
  readonly tag: TagBlock;

  /** 검색어와의 유사도 (0-1) */
  readonly relevance: number;

  /** 매칭된 컨텍스트 */
  readonly context: string;
}

/**
 * TAG 파싱 결과
 */
export interface TagParseResult {
  /** 파싱 성공 여부 */
  readonly success: boolean;

  /** 파싱된 TAG 블록 */
  readonly tagBlock?: TagBlock;

  /** 파싱 오류 메시지 */
  readonly error?: string;

  /** 경고 메시지들 */
  readonly warnings: string[];
}

/**
 * TAG 검증 결과
 */
export interface TagValidationResult {
  /** 검증 통과 여부 */
  readonly valid: boolean;

  /** 오류 목록 */
  readonly errors: string[];

  /** 경고 목록 */
  readonly warnings: string[];

  /** 수정 제안들 */
  readonly suggestions: string[];
}

/**
 * TAG 불변성 검사 결과
 */
export interface ImmutabilityCheckResult {
  /** 불변성 위반 여부 */
  readonly violated: boolean;

  /** 변경 시도된 TAG */
  readonly modifiedTag?: string;

  /** 위반 상세 정보 */
  readonly violationDetails?: string;
}

/**
 * AI 검색 패턴
 */
export interface AISearchPattern {
  /** 검색 패턴 이름 */
  readonly name: string;

  /** ripgrep 패턴 */
  readonly pattern: string;

  /** 파일 타입 필터 */
  readonly fileTypes?: string[];

  /** 추가 옵션 */
  readonly options?: string[];
}

/**
 * TAG 마이그레이션 결과
 */
export interface MigrationResult {
  /** 마이그레이션 성공 여부 */
  readonly success: boolean;

  /** 처리된 TAG 수 */
  readonly processedCount: number;

  /** 성공적으로 변환된 TAG 수 */
  readonly convertedCount: number;

  /** 실패한 TAG 목록 */
  readonly failures: string[];

  /** 경고 메시지들 */
  readonly warnings: string[];
}

/**
 * 성능 메트릭
 */
export interface PerformanceMetrics {
  /** 검색 응답 시간 (ms) */
  readonly searchTime: number;

  /** 파싱 시간 (ms) */
  readonly parseTime: number;

  /** 메모리 사용량 (bytes) */
  readonly memoryUsage: number;

  /** 스캔된 파일 수 */
  readonly filesScanned: number;

  /** 발견된 TAG 수 */
  readonly tagsFound: number;
}

/**
 * Code-First TAG 설정
 */
export interface CodeFirstTagConfig {
  /** 스캔할 파일 확장자들 */
  readonly scanExtensions: string[];

  /** 제외할 디렉토리들 */
  readonly excludeDirectories: string[];

  /** TAG 블록 시작 패턴 */
  readonly tagBlockStartPattern: string;

  /** TAG 블록 종료 패턴 */
  readonly tagBlockEndPattern: string;

  /** 불변성 강제 여부 */
  readonly enforceImmutability: boolean;

  /** 성능 모니터링 활성화 여부 */
  readonly enablePerformanceTracking: boolean;
}
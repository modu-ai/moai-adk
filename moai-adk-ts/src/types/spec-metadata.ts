/**
 * SPEC 메타데이터 타입 정의
 * @file spec-metadata.ts
 * @description SPEC 문서의 YAML frontmatter 메타데이터 구조 정의
 *
 * @REQ:SPEC-METADATA-001 SPEC 메타데이터 구조 정의 요구사항
 * @DESIGN:YAML-FRONTMATTER-001 YAML frontmatter 기반 메타데이터 설계
 * @TASK:TYPE-DEFINITION-001 TypeScript 타입 정의 구현
 * @TEST:SPEC-METADATA-TEST-001 메타데이터 타입 정의 테스트 (covered by enum validation)
 */

/**
 * SPEC 상태 열거형
 */
export enum SpecStatus {
  /** 초안 - SPEC 작성 중, 아직 구현 시작하지 않음 */
  DRAFT = 'draft',

  /** 활성 - 구현 진행 중, /moai:2-build 단계 실행 중 */
  ACTIVE = 'active',

  /** 완료 - 구현 완료 및 테스트 통과, /moai:3-sync로 문서 동기화 완료 */
  COMPLETED = 'completed',

  /** 폐기 - 더 이상 사용하지 않음, 다른 SPEC으로 대체됨 */
  DEPRECATED = 'deprecated',
}

/**
 * SPEC 우선순위 열거형
 */
export enum SpecPriority {
  /** 높음 - 즉시 처리 필요, 비즈니스 크리티컬 */
  HIGH = 'high',

  /** 중간 - 일반적인 개발 우선순위 */
  MEDIUM = 'medium',

  /** 낮음 - 추후 처리 가능, 개선 사항 */
  LOW = 'low',
}

/**
 * SPEC 메타데이터 인터페이스
 * YAML frontmatter에서 사용되는 메타데이터 구조
 */
export interface SpecMetadata {
  /** SPEC 고유 식별자 (예: "SPEC-014") */
  spec_id: string;

  /** SPEC 현재 상태 */
  status: SpecStatus;

  /** SPEC 우선순위 */
  priority: SpecPriority;

  /** 의존하는 SPEC ID 목록 (선택사항) */
  dependencies?: string[];

  /** 관련된 SPEC ID 목록 (선택사항) */
  related_specs?: string[];

  /** 분류 태그 (선택사항) */
  tags?: string[];

  /** 생성 날짜 (선택사항) */
  created?: string;

  /** 마지막 업데이트 날짜 (선택사항) */
  updated?: string;

  /** 목표 완료 날짜 (선택사항) */
  target_completion?: string;

  /** 담당자 (Team 모드에서 사용, 선택사항) */
  assignee?: string;

  /** 리뷰어 목록 (Team 모드에서 사용, 선택사항) */
  reviewers?: string[];

  /** 팀 이름 (Team 모드에서 사용, 선택사항) */
  team?: string;
}

/**
 * SPEC 메타데이터 검증 결과
 */
export interface SpecMetadataValidation {
  /** 검증 성공 여부 */
  isValid: boolean;

  /** 오류 메시지 목록 */
  errors: string[];

  /** 경고 메시지 목록 */
  warnings: string[];
}

/**
 * SPEC 의존성 정보
 */
export interface SpecDependency {
  /** 의존 대상 SPEC ID */
  spec_id: string;

  /** 의존 대상 SPEC 상태 */
  status: SpecStatus;

  /** 의존성 만족 여부 */
  satisfied: boolean;
}

/**
 * SPEC 의존성 그래프 노드
 */
export interface SpecDependencyNode {
  /** SPEC ID */
  spec_id: string;

  /** SPEC 메타데이터 */
  metadata: SpecMetadata;

  /** 직접 의존성 목록 */
  dependencies: SpecDependency[];

  /** 이 SPEC에 의존하는 SPEC 목록 */
  dependents: string[];
}

/**
 * SPEC 상태 전환 규칙
 */
export const STATUS_TRANSITIONS: Record<SpecStatus, SpecStatus[]> = {
  [SpecStatus.DRAFT]: [SpecStatus.ACTIVE, SpecStatus.DEPRECATED],
  [SpecStatus.ACTIVE]: [SpecStatus.COMPLETED, SpecStatus.DEPRECATED],
  [SpecStatus.COMPLETED]: [SpecStatus.DEPRECATED],
  [SpecStatus.DEPRECATED]: [], // 폐기된 SPEC은 상태 변경 불가
};

/**
 * 메타데이터 기본값
 */
export const DEFAULT_METADATA: Partial<SpecMetadata> = {
  status: SpecStatus.DRAFT,
  priority: SpecPriority.MEDIUM,
  dependencies: [],
  related_specs: [],
  tags: [],
};

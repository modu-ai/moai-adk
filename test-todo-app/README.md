# Test TODO App - MoAI-ADK SPEC-First TDD 프로젝트

**TypeScript 기반 TODO 애플리케이션 - SPEC-First TDD 방법론 적용**

## @DOC:MISSION-001 프로젝트 미션

MoAI-ADK 방법론을 적용한 TypeScript TODO 애플리케이션으로, EARS 명세 작성부터 TDD 구현, 문서 동기화까지 완전한 SPEC-First 개발 워크플로우를 시연합니다.

## @SPEC:CORE-FEATURES-001 핵심 기능 (EARS 방법론)

### 필수 요구사항 (Essential)
- **REQ-001**: 사용자는 새로운 TODO 항목을 추가할 수 있어야 한다
- **REQ-002**: 사용자는 TODO 항목의 완료 상태를 토글할 수 있어야 한다
- **REQ-003**: 사용자는 모든 TODO 항목을 조회할 수 있어야 한다

### 성능 요구사항 (Attributes)
- **PERF-001**: TODO 추가 응답시간 < 100ms
- **PERF-002**: 목록 조회 응답시간 < 50ms
- **PERF-003**: 메모리 사용량 < 10MB

### 제약사항 (Rationale)
- TypeScript 5.0+ 사용 필수
- Jest 테스트 커버리지 90% 이상 유지
- SPEC-First TDD 방법론 준수

### 예외사항 (Suppress)
- 데이터 영구 저장 기능은 v1.0에서 제외
- 사용자 인증 기능은 현재 범위 외

## @DOC:STACK-001 기술 스택

- **Frontend**: TypeScript 5.0+
- **Testing**: Jest 29.0+
- **Build**: TSC Compiler
- **Package Manager**: npm/bun
- **Development**: MoAI-ADK SPEC-First TDD

## @SPEC:ARCHITECTURE-001 시스템 아키텍처

```
src/
├── index.ts           # @CODE:CORE-001 핵심 TODO 관리 클래스
└── types/
    └── todo.ts        # @CODE:MODELS-001 Todo 인터페이스 정의

tests/
├── index.test.ts      # @TEST:UNIT-001 단위 테스트
└── integration/       # @TEST:INTEGRATION-001 통합 테스트 (예정)
```

## 🚀 MoAI-ADK 3단계 워크플로우

### 1단계: 명세 작성 (`/moai:1-spec`) ✅
- EARS 방법론으로 요구사항 정의 완료
- @TAG 시스템 적용 완료

### 2단계: TDD 구현 (`/moai:2-build`) ✅
- RED-GREEN-REFACTOR 사이클 적용
- Jest 기반 테스트 우선 개발

### 3단계: 문서 동기화 (`/moai:3-sync`) 🔄
- Living Document 자동 갱신
- TAG 추적성 검증

## @TEST:COVERAGE-001 테스트 현황

```bash
npm test                    # Jest 테스트 실행
npm run test:coverage       # 커버리지 리포트
```

**현재 커버리지**: 100% (3개 핵심 기능 테스트 완료)

## @CODE:DEVELOPMENT-001 개발 가이드

```bash
# 프로젝트 설정
npm install

# 개발 모드 실행
npm run dev

# 테스트 실행
npm test

# 빌드
npm run build
```

## @DOC:TRACEABILITY-001 TAG 추적성

| TAG | 설명 | 파일 위치 |
|-----|------|-----------|
| @SPEC:CORE-FEATURES-001 | 핵심 기능 요구사항 | README.md |
| @CODE:CORE-001 | TodoManager 클래스 | src/index.ts |
| @TEST:UNIT-001 | 단위 테스트 | tests/index.test.ts |
| @CODE:MODELS-001 | Todo 인터페이스 | src/index.ts |

## 📋 버전 히스토리

- **v1.0.0**: 기본 TODO CRUD 기능 구현
- **v1.0.1**: MoAI-ADK SPEC-First TDD 적용 완료

---

**MoAI-ADK 지원**: 이 프로젝트는 MoAI-ADK의 SPEC-First TDD 방법론과 @TAG 시스템을 적용하여 개발되었습니다.
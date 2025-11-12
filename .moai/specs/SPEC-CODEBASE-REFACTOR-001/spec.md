---
id: CODEBASE-REFACTOR-001
version: 0.1.0
status: draft
created: 2025-11-05
updated: 2025-11-05
priority: high
category: refactor
labels:
  - codebase
  - tag-chains
  - organization
  - maintainability
---


## HISTORY

### v0.1.0 (2025-11-05)
- **INITIAL**: 코드베이스 종합 리팩토링 SPEC 작성
- **SCOPE**: TAG 체인 복구, 코드 조직화, 유지보수성 개선
- **CONTEXT**: TAG 분석 리포트 기반 체계적 리팩토링 필요
- **TASKS**:
  1. 246개 손상된 TAG 체인 복구
  2. 고아 TAG 정리 및 재구성
  3. 코드-테스트-SPEC 연결 강화
  4. 모듈별 책임 재정의 및 개선

---

## Environment (환경)

### 현재 상태 (TAG 분석 기준)

**TAG 현황**:
- 전체 도메인: 246개
- 완전 체인: 0개
- 부분 체인: 0개
- 손상 체인: 246개 (100%)

**고아 TAG 분포**:
- CODE without SPEC: 27개
- CODE without TEST: 27개
- TEST without CODE: 121개
- SPEC without CODE: 55개

**프로젝트 구조**:
- 언어: Python (TypeScript 지원 감소)
- 패키지 관리: uv/pip
- 테스트: pytest
- 문서: Markdown 기반

**기존 리팩토링 현황**:
- SKILL-REFACTOR-001: 완료 (Skills 표준화)
- REFACTOR-001: 완료 (Git Manager 분리)
- UPDATE-REFACTOR-001~003: 진행 중

---

## Assumptions (가정)

### 기술적 가정

1. **TAG 시스템**: 현재 모든 TAG 체인이 손상되었으므로, 체계적 재구성 필요
2. **코드 조직**: 기능별로 코드가 분산되어 있어 통합이 필요
3. **테스트 커버리지**: 121개의 테스트만 존재하므로 확장 필요
4. **문서화**: SPEC은 존재하나 CODE와 연결이 부족

### 비즈니스 가정

1. **유지보수성**: TAG 체인 복구가 장기 유지보수성 향상에 필수
2. **개발 효율**: 코드 조직화가 개발 속도 및 품질 향상에 기여
3. **협업**: 표준화된 구조가 팀 협업 효율성 증대
4. **품질 보증**: TAG 기반 추적성이 품질 보증 강화

---

## Requirements (요구사항)

### Ubiquitous Requirements (기본 요구사항)

- 시스템은 모든 TAG 체인을 복구해야 한다 (SPEC ↔ CODE ↔ TEST)
- 시스템은 고아 TAG를 정리하고 재구성해야 한다
- 시스템은 코드 조직을 기능별로 개선해야 한다
- 시스템은 테스트 커버리지를 85% 이상으로 확보해야 한다

### Event-driven Requirements (이벤트 기반)

- WHEN 고아 CODE TAG가 발견되면, 시스템은 해당 SPEC과 TEST를 생성하거나 연결해야 한다
- WHEN 고아 TEST TAG가 발견되면, 시스템은 해당 CODE와 SPEC을 생성하거나 연결해야 한다
- WHEN 고아 SPEC TAG가 발견되면, 시스템은 해당 CODE와 TEST를 생성해야 한다
- WHEN 모듈 경계가 불분명하면, 시스템은 책임 분리를 개선해야 한다

### State-driven Requirements (상태 기반)

- WHILE TAG 체인 복구 중일 때, 시스템은 기존 기능을 유지해야 한다
- WHILE 코드 재구성 중일 때, 시스템은 API 호환성을 보장해야 한다
- WHILE 테스트를 확장 중일 때, 시스템은 모든 테스트가 통과하도록 해야 한다

### Constraints (제약사항)

- IF TAG ID가 중복되면, 새로운 ID 체계로 마이그레이션해야 한다
- IF CODE가 없는 SPEC이라면, 해당 SPEC의 우선순위를 재평가해야 한다
- IF TEST가 없는 CODE라면, 반드시 테스트를 생성해야 한다
- 모든 변경사항은 TRUST 5 원칙을 준수해야 한다

---


### 핵심 TAG 체인


### 관련 기존 TAG


---

## Specifications (상세 명세)

### 1. TAG 체인 복구 전략

#### 1.1 고아 TAG 처리 원칙

**CODE without SPEC/TEST (27개)**:
```
...
```

**TEST without CODE (121개)**:
```
...
```

**SPEC without CODE (55개)**:
```
...
```

#### 1.2 우선순위 분류

**높은 우선순위** (즉시 복구):
- 핵심 기능 관련 TAG (LDE, CORE, INSTALLER)
- 보안 관련 TAG (SEC, AUTH)
- 품질 관련 TAG (TEST, COVERAGE)

**중간 우선순위** (계획적 복구):
- 문서화 관련 TAG (DOCS, DOC-TAG)
- 유틸리티 관련 TAG (UTILS, TOOLS)

**낮은 우선순위** (필요시 복구):
- 실험적 기능 TAG (EXPERIMENTAL)
- 폐기 예정 기능 TAG (DEPRECATED)

### 2. 코드 조직화 전략

#### 2.1 모듈별 책임 재정의

**Core 모듈 (src/moai_adk/core/)**:
- 프로젝트 관리: project/, config/
- Git 통합: git/ (이미 리팩토링 완료)
- 버전 관리: version/, update/
- 템플릿: template/, backup/

**Utils 모듈 (src/moai_adk/utils/)**:
- 공통 유틸리티: common/, helpers/
- 검증기: validators/, checkers/
- 로깅: logging/, monitoring/

**CLI 모듈 (src/moai_adk/cli/)**:
- 명령어: commands/
- 인터페이스: ui/, prompts/
- 설정: cli-config/

**Domain 모듈 (src/moai_adk/domain/)**:
- 언어 감지: languages/
- 환경 설정: environments/
- 플랫폼: platforms/

#### 2.2 의존성 개선

**현재 문제점**:
- 순환 의존성 존재
- 모듈 경계 불분명
- 중복 코드 발생

**개선 방안**:
- 단방향 의존성 강제
- 인터페이스 기반 설계
- 공통 모듈 추출

### 3. 테스트 전략

#### 3.1 테스트 커버리지 확장

**목표 커버리지**: 85% 이상

**확대 대상**:
- 27개 CODE에 대한 단위 테스트
- 55개 SPEC에 대한 통합 테스트
- 핵심 모듈에 대한 E2E 테스트

#### 3.2 테스트 구조 개선

```
tests/
├── unit/           # 단위 테스트
│   ├── core/       # Core 모듈 테스트
│   ├── utils/      # Utils 모듈 테스트
│   ├── cli/        # CLI 모듈 테스트
│   └── domain/     # Domain 모듈 테스트
├── integration/    # 통합 테스트
├── e2e/           # E2E 테스트
└── fixtures/      # 테스트 데이터
```

### 4. 마이그레이션 절차

#### Phase 1: TAG 분석 및 정리 (1주일)
1. 전체 TAG 현황 상세 분석
2. 우선순위별 그룹화
3. 복구 가능성 평가
4. 마이그레이션 계획 수립

#### Phase 2: 핵심 기능 복구 (2주일)
1. 높은 우선순위 TAG 복구
2. CORE/LDE/INSTALLER 모듈 연결
3. 기본 테스트 커버리지 확보
4. API 호환성 검증

#### Phase 3: 전체 기능 확장 (3주일)
1. 중간/낮은 우선순위 TAG 복구
2. 코드 조직화 진행
3. 테스트 커버리지 85% 달성
4. 문서화 완성

#### Phase 4: 최종 검증 및 안정화 (1주일)
1. 전체 TAG 체인 검증
2. 성능 테스트
3. 회귀 테스트
4. 문서 최종화

### 5. 품질 게이트

#### TAG 품질

- TAG 체인 완성도: 100%
- 고아 TAG: 0개
- TAG 중복: 없음
- TAG 추적성: 완전

#### 코드 품질

- 테스트 커버리지: ≥85%
- 모듈 복잡도: ≤10
- 순환 의존성: 0개
- 코드 중복: ≤5%

#### 문서 품질

- SPEC-CODE 연결: 100%
- CODE-TEST 연결: 100%
- 문서 최신성: 100%
- API 문서: 완비

---

## 성공 지표

### 정량적 지표

- TAG 체인 복구: 0개 → 100% (246개 복구)
- 테스트 커버리지: 현재 수준 → 85% 이상
- 모듈화: 현재 구조 → 명확한 책임 분리
- 문서화: 현재 수준 → 100% 연결

### 정성적 지표

- 코드 가독성 향상
- 유지보수성 개선
- 개발 속도 향상
- 협업 효율성 증대

---

## 리스크 및 완화 방안

### 리스크 1: TAG 체인 복구 복잡성

**완화 방안**:
- 점진적 복구 전략 (우선순위별)
- 자동화 스크립트 활용
- 철저한 검증 절차

### 리스크 2: 코드 호환성 깨짐

**완화 방안**:
- API 호환성 테스트 강화
- 점진적 리팩토링 진행
- 롤백 계획 수립

### 리스크 3: 개발 속도 저하

**완화 방안**:
- 명확한 일정 관리
- 자동화 도구 활용
- 팀 동기화 강화

---

## 다음 단계

1. `/alfred:2-run CODEBASE-REFACTOR-001` - TDD 기반 리팩토링 시작
2. Phase별 순차 진행 (분석 → 복구 → 확장 → 안정화)
3. 주간 진행 상황 검증 및 조정
4. `/alfred:3-sync` - TAG 체인 검증 및 문서 동기화

---

**작성일**: 2025-11-05
**버전**: 0.1.0
**상태**: Draft
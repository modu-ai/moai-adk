---
spec_id: SPEC-CORE-001
title: Foundation Methodologies - Implementation Plan
created: 2026-02-03
status: Planned
tags: [foundation, ears, languages, trust5, domain-patterns]
---

# SPEC-CORE-001: Foundation Methodologies - 구현 계획

## 1. 구현 전략 개요

### 접근 방식

이 모듈은 순수 데이터/로직 모듈로 외부 의존성이 없다. 따라서 DDD의 ANALYZE 단계에서 기존 Python 구현(`foundation/`)을 참조하되, Go의 관용적 패턴(table-driven tests, iota enums, unexported 레지스트리)을 활용하여 재설계한다.

### 아키텍처 방향

- **데이터 중심 설계**: 모든 정의는 Go 구조체 리터럴로 컴파일 타임에 초기화
- **레지스트리 패턴**: 각 도메인별로 unexported 슬라이스 레지스트리를 두고, exported 함수로 조회
- **인터페이스 최소화**: 순수 데이터 모듈이므로 인터페이스보다 구체 타입 우선
- **JSON 호환**: 모든 타입에 json 구조체 태그 필수 적용

### 리스크 및 대응

| 리스크 | 영향 | 대응 방안 |
|--------|------|----------|
| 언어 정의 누락 | LSP 연동 실패 | Python 소스 대조 + 16개 언어 전수 테스트 |
| EARS 패턴 검증 부정확 | SPEC 생성 오류 | 정규식 + 테스트 케이스 50개 이상 |
| TRUST 5 가중치 불균형 | 품질 점수 왜곡 | 기존 Python 가중치 참조 + 설정 가능 구조 |
| Go 패키지 순환 참조 | 빌드 실패 | foundation -> trust 단방향 참조만 허용 |

---

## 2. 마일스톤

### Primary Goal: EARS 패턴 엔진 (ears.go)

**우선순위**: High
**예상 LOC**: ~200

#### 태스크 분해

| # | 태스크 | 설명 |
|---|--------|------|
| 1.1 | 타입 정의 | `EARSPatternType`, `EARSPattern`, `EARSValidationResult` 구조체 정의 |
| 1.2 | 패턴 레지스트리 | 5가지 표준 패턴 + Complex 패턴 데이터 초기화 |
| 1.3 | 조회 함수 | `GetAllPatterns()`, `GetPattern()` 구현 |
| 1.4 | 패턴 검증 | `ValidateEARSRequirement()` 정규식 기반 검증 로직 |
| 1.5 | 패턴 탐지 | `DetectPatternType()` 키워드 기반 분류 로직 |
| 1.6 | 단위 테스트 | 정상 케이스, 에지 케이스, 오류 케이스 테스트 작성 |

#### 기술적 접근

- EARS 5 패턴은 `var patterns = []EARSPattern{...}`로 패키지 레벨 초기화
- 검증 로직은 `regexp` 패키지로 영문/한국어 패턴 매칭
- 복합 패턴은 State + Event 조합을 감지하여 `Complex` 타입으로 분류
- 모호한 표현("should", "might", "usually")은 명시적으로 거부

---

### Secondary Goal: 언어 생태계 정의 (langs.go)

**우선순위**: High
**예상 LOC**: ~300

#### 태스크 분해

| # | 태스크 | 설명 |
|---|--------|------|
| 2.1 | 타입 정의 | `LanguageDefinition` 구조체 + JSON 태그 |
| 2.2 | 언어 레지스트리 | 18개 언어 정의 데이터 초기화 |
| 2.3 | 확장자 인덱스 | 파일 확장자 -> 언어 매핑 `map[string]*LanguageDefinition` 구축 |
| 2.4 | ID 인덱스 | 언어 ID -> 정의 매핑 `map[string]*LanguageDefinition` 구축 |
| 2.5 | 조회 함수 | `GetAllLanguages()`, `GetLanguageByID()`, `GetLanguageByExtension()` |
| 2.6 | 유틸리티 함수 | `GetSupportedExtensions()` 전체 확장자 목록 반환 |
| 2.7 | 단위 테스트 | 18개 언어 전수 테스트, 확장자 매핑 테스트, 알 수 없는 언어 오류 테스트 |

#### 기술적 접근

- 언어 정의는 `var languages = []LanguageDefinition{...}`로 패키지 레벨 초기화
- `init()` 함수에서 확장자 인덱스와 ID 인덱스를 한 번만 구축
- C와 C++의 공유 확장자(.h)는 C에 우선 매핑하되, 별도 해결 함수 제공 고려
- 대소문자 무시 매칭: `.R`과 `.r` 모두 R 언어로 인식

---

### Tertiary Goal: 도메인 패턴 라이브러리 (backend.go, frontend.go, database.go, testing.go, devops.go)

**우선순위**: Medium
**예상 LOC**: ~300

#### 태스크 분해

| # | 태스크 | 설명 |
|---|--------|------|
| 3.1 | 공통 타입 정의 | `DomainPattern` 구조체 + 공통 조회 인터페이스 |
| 3.2 | 백엔드 패턴 | API, 인증, 마이크로서비스, 아키텍처 패턴 정의 |
| 3.3 | 프론트엔드 패턴 | 컴포넌트, 상태관리, 렌더링, 스타일링 패턴 정의 |
| 3.4 | 데이터베이스 패턴 | DB 유형, ORM, 마이그레이션, 캐싱 패턴 정의 |
| 3.5 | 테스팅 패턴 | 테스트 레벨, 전략, 커버리지, 테스트 더블 정의 |
| 3.6 | DevOps 패턴 | CI/CD, 컨테이너, 인프라, 모니터링 패턴 정의 |
| 3.7 | 통합 조회 함수 | `GetPatternsByCategory()`, `GetPatternByID()` 구현 |
| 3.8 | 단위 테스트 | 도메인별, 카테고리별 조회 테스트 |

#### 기술적 접근

- 5개 도메인 파일이 공통 `DomainPattern` 타입을 공유
- 각 파일은 `var xxxPatterns = []DomainPattern{...}` 형태로 패턴 등록
- 통합 레지스트리는 `init()`에서 모든 도메인 패턴을 하나의 맵에 병합
- `GetPatternsByCategory(domain, category)`: 2단계 필터링

---

### Final Goal: TRUST 5 원칙 시스템 (trust/)

**우선순위**: Medium
**예상 LOC**: ~200

#### 태스크 분해

| # | 태스크 | 설명 |
|---|--------|------|
| 4.1 | 원칙 타입 정의 | `TRUSTPrinciple`, `Severity`, `Automation`, `Phase` enum 타입 |
| 4.2 | 원칙 정의 데이터 | 5가지 원칙별 상세 정의 (설명, 핵심 지표, 가중치) |
| 4.3 | 체크리스트 항목 정의 | 원칙별 5-10개 체크리스트 항목 데이터 |
| 4.4 | 조회 함수 | `GetAllPrinciples()`, `GetChecklist()`, `GetChecklistForPhase()` |
| 4.5 | 점수 계산 엔진 | `CalculatePrincipleScore()`, `CalculateTRUST5Score()` |
| 4.6 | 등급 산정 | 점수 -> 등급(A/B/C/D/F) 변환 로직 |
| 4.7 | 단위 테스트 | 점수 계산 정확성, 등급 경계값, 빈 결과 처리 테스트 |

#### 기술적 접근

- `trust/` 서브패키지는 `foundation` 패키지와 분리하여 관심사 분리
- 원칙별 가중치는 기본값 동일(각 20%)이되, 커스터마이징 가능한 구조
- 등급 산정: A(90+), B(80+), C(70+), D(60+), F(60 미만)
- 체크리스트 항목은 `Phase` 태그로 plan/run/sync 필터링 지원

---

## 3. 구현 순서 및 의존성

```
[1. ears.go] ─────────────────────┐
                                   │
[2. langs.go] ────────────────────┤
                                   ├──> [통합 테스트]
[3. backend.go ~ devops.go] ──────┤
                                   │
[4. trust/principles.go] ─────────┤
[4. trust/checklist.go] ──────────┘
```

모든 파일은 독립적으로 구현 가능하며, 서로 간에 의존성이 없다. 병렬 구현이 가능하지만, 태스크 우선순위에 따라 순차 진행을 권장한다.

---

## 4. 테스트 전략

### 단위 테스트

모든 파일은 대응하는 `*_test.go` 파일을 가진다:

| 구현 파일 | 테스트 파일 | 주요 테스트 항목 |
|----------|-----------|----------------|
| ears.go | ears_test.go | 패턴 조회, 검증 성공/실패, 복합 패턴, 모호한 표현 거부 |
| langs.go | langs_test.go | 18개 언어 전수 조회, 확장자 매핑, 알 수 없는 언어 오류 |
| backend.go | backend_test.go | 백엔드 패턴 조회, 카테고리 필터링 |
| frontend.go | frontend_test.go | 프론트엔드 패턴 조회, 카테고리 필터링 |
| database.go | database_test.go | 데이터베이스 패턴 조회, 카테고리 필터링 |
| testing.go | testing_test.go | 테스팅 패턴 조회, 카테고리 필터링 |
| devops.go | devops_test.go | DevOps 패턴 조회, 카테고리 필터링 |
| trust/principles.go | trust/principles_test.go | 원칙 조회, JSON 직렬화 |
| trust/checklist.go | trust/checklist_test.go | 체크리스트 필터링, 점수 계산, 등급 산정 |

### 테스트 패턴

- **Table-driven tests**: Go 관용적 테이블 기반 테스트 패턴 사용
- **Parallel execution**: `t.Parallel()` 적용으로 테스트 병렬 실행
- **Subtests**: `t.Run(name, func)` 패턴으로 테스트 케이스 구조화
- **Benchmark tests**: `GetLanguageByExtension` 등 조회 함수의 성능 벤치마크

### 커버리지 목표

- **전체 커버리지**: 95%+ (순수 데이터/로직 모듈이므로 높은 목표 설정)
- **분기 커버리지**: 90%+ (오류 경로 포함)

---

## 5. 품질 게이트

### TRUST 5 체크리스트 (이 SPEC 적용)

| 원칙 | 항목 | 기준 |
|------|------|------|
| Tested | 단위 테스트 커버리지 | 95%+ |
| Tested | 벤치마크 테스트 | 조회 함수 < 1ms |
| Readable | godoc 주석 | 모든 exported 함수에 godoc 주석 |
| Readable | 패키지 문서 | doc.go 파일에 패키지 설명 |
| Unified | gofumpt 포맷팅 | 전체 파일 포맷팅 준수 |
| Unified | golangci-lint | 린트 오류 0개 |
| Secured | 입력 검증 | 모든 공개 함수에 nil/empty 입력 검증 |
| Trackable | 커밋 메시지 | conventional commit 형식 |

---

## 6. 참고 사항

### Python 소스 참조

기존 Python 구현의 다음 파일을 참조하여 데이터 완전성을 보장한다:

- `foundation/ears.py`: EARS 패턴 정의
- `foundation/langs.py`: 언어 생태계 정의
- `foundation/backend.py` ~ `foundation/devops.py`: 도메인 패턴
- `foundation/trust5.py`: TRUST 5 원칙 및 체크리스트

### Go 관용 패턴 적용

- `iota`를 활용한 열거형 대신 문자열 상수 사용 (JSON 직렬화 편의)
- `unexported` 레지스트리 + `exported` 조회 함수 패턴
- `init()` 함수에서 인덱스 구축 (한 번만 실행)
- `sync.Once`는 이 모듈에서 불필요 (패키지 레벨 `init()`으로 충분)

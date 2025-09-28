---
name: moai:2-build
description: 구현할 SPEC ID (예: SPEC-001) 또는 all로 모든 SPEC 구현: 언어별 최적화된 TDD 구현 (Red-Green-Refactor) with JSON-based TAG system
argument-hint: "SPEC-ID | all"
tools: Read, Write, Edit, MultiEdit, Bash, Task, WebFetch, Grep, Glob, TodoWrite
---

# MoAI-ADK 2단계: 언어별 최적화된 TDD 구현 (Red-Green-Refactor)

**TDD 구현 대상**: $ARGUMENTS

## 🔍 STEP 1: SPEC 분석 및 구현 계획 수립

먼저 지정된 SPEC을 분석하여 구현 계획을 수립하고 사용자 확인을 받습니다.

### SPEC 분석 진행

1. **SPEC 문서 분석**
   - 요구사항 추출 및 복잡도 평가
   - 기술적 제약사항 확인
   - 의존성 및 영향 범위 분석

2. **구현 전략 수립**
   - 프로젝트 언어 감지 및 최적화된 구현 전략
   - TDD 접근 방식 결정 (언어별 도구 선택)
   - 예상 작업 범위 및 시간 산정

3. **구현 계획 보고**
   - 단계별 구현 계획 제시
   - 잠재적 위험 요소 식별
   - 품질 게이트 체크포인트 설정

### 사용자 확인 단계

구현 계획 검토 후 다음 중 선택하세요:
- **"진행"** 또는 **"시작"**: 계획대로 TDD 구현 시작
- **"수정 [내용]"**: 계획 수정 요청
- **"중단"**: 구현 작업 중단

---

## 🚀 STEP 2: TDD 구현 실행 (사용자 승인 후)

사용자 승인 후 code-builder 에이전트가 **언어별 최적화**된 Red-Green-Refactor 사이클과 TRUST 원칙 검증을 지원합니다.

## 🔗 언어별 TDD 최적화

### 프로젝트 언어 감지 및 최적 라우팅

TDD 구현은 전용 스크립트를 통해 수행됩니다:

```bash
# TDD 구현 실행 (Red-Green-Refactor)
tsx .moai/scripts/tdd-runner.ts --spec-id="$ARGUMENTS" --phase=all

# 단계별 실행
tsx .moai/scripts/tdd-runner.ts --spec-id=SPEC-001 --phase=red    # 실패 테스트 작성
tsx .moai/scripts/tdd-runner.ts --spec-id=SPEC-001 --phase=green  # 최소 구현
tsx .moai/scripts/tdd-runner.ts --spec-id=SPEC-001 --phase=refactor # 코드 개선
```

### TDD 도구 매핑

| SPEC 타입 | 구현 언어 | 테스트 프레임워크 | 성능 목표 | 커버리지 목표 |
|-----------|-----------|-------------------|-----------|---------------|
| **CLI/시스템** | TypeScript | Vitest + tsx | < 18ms | 95%+ |
| **API/백엔드** | TypeScript | Vitest + SuperTest | < 50ms | 90%+ |
| **프론트엔드** | TypeScript | Vitest + Testing Library | < 100ms | 85%+ |
| **데이터 처리** | TypeScript | Vitest + Mock | < 200ms | 85%+ |
| **사용자 Python 프로젝트** | Python 지원 | pytest 도구 | 사용자 정의 | 사용자 정의 |

## 🚀 최적화된 에이전트 협업 구조

- **Phase 1**: `code-builder` 에이전트가 전체 TDD 사이클(Red-Green-Refactor)을 일괄 처리합니다.
- **Phase 2**: `git-manager` 에이전트가 TDD 완료 후 모든 커밋을 한 번에 처리합니다.
- **단일 책임 원칙**: code-builder는 전체 TDD 구현, git-manager는 Git 작업 일괄 처리
- **배치 처리**: 단계별 중단 없이 연속적인 TDD 사이클 실행
- **에이전트 간 호출 금지**: 각 에이전트는 독립적으로 실행, 커맨드 레벨에서만 순차 호출

## 🔄 2단계 워크플로우 실행 순서

### Phase 1: 분석 및 계획 단계

**SPEC 분석기**가 다음을 수행:

1. **SPEC 문서 로딩**: 지정된 SPEC ID 또는 all 모드에 따른 문서 분석
2. **복잡도 평가**: 구현 범위, 기술적 제약사항, 의존성 분석
3. **언어별 구현 전략**: 프로젝트 언어별 최적화 방안 제시
4. **구현 계획 생성**: 단계별 TDD 접근 방식 및 예상 작업량 산정
5. **사용자 승인 대기**: 계획 검토 및 피드백 수집

### Phase 2: TDD 구현 단계 (승인 후)

`code-builder` 에이전트가 사용자 승인 후 **연속적으로** 수행:

1. **RED**: 실패하는 테스트 작성 및 확인
2. **GREEN**: 최소 구현으로 테스트 통과 확인
3. **REFACTOR**: 코드 품질 개선 및 TRUST 원칙 검증
4. **품질 검증**: 린터, 테스트 커버리지, 보안 검사 일괄 실행

### Phase 3: Git 작업 (git-manager)

`git-manager` 에이전트가 TDD 완료 후 **한 번에** 수행:

1. **체크포인트 생성**: TDD 시작 전 백업 포인트
2. **구조화된 커밋**: RED→GREEN→REFACTOR 단계별 커밋 생성
3. **최종 동기화**: 모드별 Git 전략 적용 및 원격 동기화


## 📋 STEP 1 실행 가이드: SPEC 분석 및 계획 수립

### 1. SPEC 문서 분석

다음을 우선적으로 실행하여 SPEC을 분석합니다:

```bash
# SPEC 문서 분석 스크립트 실행
tsx .moai/scripts/spec-analyzer.ts --target="$ARGUMENTS" --mode=analysis
```

#### 분석 체크리스트

- [ ] **요구사항 명확성**: SPEC의 기능 요구사항이 구체적인가?
- [ ] **기술적 제약**: 성능, 호환성, 보안 요구사항 확인
- [ ] **의존성 분석**: 기존 코드와의 연결점 및 영향 범위
- [ ] **복잡도 평가**: 구현 난이도 및 예상 작업량

### 2. 구현 전략 결정

#### TypeScript 구현 기준

| SPEC 특성 | 구현 언어 | 이유 |
|-----------|-----------|------|
| CLI/시스템 도구 | TypeScript | 고성능 (18ms), 타입 안전성, JSON TAG 시스템 통합 |
| API/백엔드 | TypeScript | Node.js 생태계, Express/Fastify 호환성 |
| 프론트엔드 | TypeScript | React/Vue 네이티브 지원 |
| 데이터 처리 | TypeScript | 고성능 비동기 처리, 타입 안전성 |
| 사용자 Python 프로젝트 | Python 도구 지원 | MoAI-ADK가 Python 프로젝트 개발 도구 제공 |

#### TDD 접근 방식

- **Bottom-up**: 유틸리티 → 서비스 → API
- **Top-down**: API → 서비스 → 유틸리티
- **Middle-out**: 핵심 로직 → 양방향 확장

### 3. 구현 계획 보고서 생성

다음 형식으로 계획을 제시합니다:

```
## 구현 계획 보고서: [SPEC-ID]

### 📊 분석 결과
- **복잡도**: [낮음/중간/높음]
- **예상 작업시간**: [시간 산정]
- **주요 기술 도전**: [기술적 어려움]

### 🎯 구현 전략
- **선택 언어**: [Python/TypeScript + 이유]
- **TDD 접근법**: [Bottom-up/Top-down/Middle-out]
- **핵심 모듈**: [주요 구현 대상]

### 🚨 위험 요소
- **기술적 위험**: [예상 문제점]
- **의존성 위험**: [외부 의존성 이슈]
- **일정 위험**: [지연 가능성]

### ✅ 품질 게이트
- **테스트 커버리지**: [목표 %]
- **성능 목표**: [구체적 지표]
- **보안 체크포인트**: [검증 항목]

---
**승인 요청**: 위 계획으로 진행하시겠습니까?
("진행", "수정 [내용]", "중단" 중 선택)
```

---

## 🚀 STEP 2 실행 가이드: TDD 구현 (승인 후)

사용자가 **"진행"** 또는 **"시작"**을 선택한 경우에만 다음을 실행합니다:

```bash
# TDD 구현 시작
tsx .moai/scripts/tdd-runner.ts --spec-id="$ARGUMENTS" --phase=all --approved=true
```

### TDD 단계별 가이드

1. **RED**: Given/When/Then 구조로 실패 테스트 작성. 언어별 테스트 파일 규칙을 따르고, 실패 로그를 간단히 기록합니다.
2. **GREEN**: 테스트를 통과시키는 최소한의 구현만 추가합니다. 최적화는 REFACTOR 단계로 미룹니다.
3. **REFACTOR**: 중복 제거, 명시적 네이밍, 구조화 로깅/예외 처리 보강. 필요 시 추가 커밋으로 분리합니다.

> 헌법 Article I은 기본 권장치만 제공하므로, `simplicity_threshold`를 초과하는 구조가 필요하다면 SPEC 또는 ADR에 근거를 남기고 진행하세요.

## 에이전트 역할 분리

### code-builder 전담 영역

- TDD Red-Green-Refactor 코드 구현
- 테스트 작성 및 실행
- TRUST 5원칙 검증
- 코드 품질 체크
- 언어별 린터/포매터 실행

### git-manager 전담 영역

- 모든 Git 커밋 작업 (add, commit, push)
- TDD 단계별 체크포인트 생성
- 모드별 커밋 전략 적용
- 깃 브랜치/태그 관리
- 원격 동기화 처리

## 품질 게이트 체크리스트

- 테스트 커버리지 ≥ `.moai/config.json.test_coverage_target` (기본 85%)
- 린터/포매터 통과 (`ruff`, `eslint --fix`, `gofmt` 등)
- 구조화 로깅 또는 관측 도구 호출 존재 확인
- 16-Core @TAG 업데이트 필요 변경 사항 메모 (다음 단계에서 doc-syncer가 사용)

## 다음 단계

- TDD 구현 완료 후 `/moai:3-sync`로 문서 동기화 진행
- 모든 Git 작업은 git-manager 에이전트가 전담하여 일관성 보장
- 에이전트 간 직접 호출 없이 커맨드 레벨 오케스트레이션만 사용

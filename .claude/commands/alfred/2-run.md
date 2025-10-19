---
name: alfred:2-run
description: "구현할 SPEC ID (예: SPEC-001) 또는 all로 모든 SPEC 구현 - 계획된 작업 실행 (TDD 구현, 프로토타입, 문서화 등)"
argument-hint: "SPEC-ID - 구현할 SPEC ID (예: SPEC-001) 또는 all로 모든 SPEC 구현"
allowed-tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Bash(python3:*)
  - Bash(pytest:*)
  - Bash(npm:*)
  - Bash(node:*)
  - Bash(git:*)
  - Task
  - WebFetch
  - Grep
  - Glob
  - TodoWrite
---

# ⚒️ MoAI-ADK 2단계: 계획 실행 (Run) - 유연한 구현 전략

## 🎯 커맨드 목적

SPEC 문서를 분석하여 계획된 작업을 실행합니다. TDD 구현뿐만 아니라 프로토타입 제작, 문서화 작업 등 다양한 실행 시나리오를 지원합니다.

**실행 대상**: $ARGUMENTS

## 💡 실행 철학: "계획(Plan) → 실행(Run) → 동기화(Sync)"

`/alfred:2-run`은 단순히 코드를 "빌드"하는 것이 아니라, **계획된 작업을 수행**하는 범용 커맨드입니다.

### 3가지 주요 시나리오

#### 시나리오 1: TDD 구현 (주 사용 방식) ⭐
```bash
/alfred:2-run SPEC-AUTH-001
→ RED → GREEN → REFACTOR
→ 테스트 주도 개발로 고품질 코드 구현
```

#### 시나리오 2: 프로토타입 제작
```bash
/alfred:2-run SPEC-PROTO-001
→ 빠른 검증을 위한 프로토타입 구현
→ 최소한의 테스트로 신속한 피드백
```

#### 시나리오 3: 문서화 작업
```bash
/alfred:2-run SPEC-DOCS-001
→ 문서 작성 및 샘플 코드 생성
→ API 문서, 튜토리얼, 가이드 등
```

> **표준 2단계 워크플로우** (자세한 내용: `CLAUDE.md` - "Alfred 커맨드 실행 패턴" 참조)

## 📋 실행 흐름

1. **SPEC 분석**: 요구사항 추출 및 복잡도 평가
2. **실행 전략 수립**: 언어별 최적화된 접근법 결정 (TDD, 프로토타입, 문서화 등)
3. **사용자 확인**: 실행 계획 검토 및 승인
4. **작업 실행**: 승인된 계획에 따라 작업 수행
5. **Git 작업**: git-manager를 통한 단계별 커밋 생성

## 🔗 연관 에이전트

- **Phase 1**: implementation-planner (📋 테크니컬 아키텍트) - SPEC 분석 및 실행 전략 수립
- **Phase 2**: tdd-implementer (🔬 시니어 개발자) - 실행 작업 전담
- **Phase 2.5**: quality-gate (🛡️ 품질 보증 엔지니어) - TRUST 원칙 검증 (자동)
- **Phase 3**: git-manager (🚀 릴리스 엔지니어) - Git 커밋 전담

## 💡 사용 예시

사용자가 다음과 같이 커맨드를 실행할 수 있습니다:
- `/alfred:2-run SPEC-001` - 특정 SPEC 실행
- `/alfred:2-run all` - 모든 SPEC 일괄 실행
- `/alfred:2-run SPEC-003 --test` - 테스트만 실행

## 🔍 STEP 1: SPEC 분석 및 실행 계획 수립

먼저 지정된 SPEC을 분석하여 실행 계획을 수립하고 사용자 확인을 받습니다.

**implementation-planner 에이전트가 자동으로 필요한 문서를 로드하여 분석합니다.**

### 🔍 코드베이스 탐색 (권장)

**기존 코드 구조를 파악하거나 유사 패턴을 찾아야 하는 경우** Explore 에이전트를 먼저 활용합니다:

```
Task tool 호출 (Explore 에이전트):
- subagent_type: "Explore"
- description: "기존 코드 구조 및 패턴 탐색"
- prompt: "SPEC-$ARGUMENTS와 관련된 기존 코드를 탐색해주세요:
          - 유사한 기능 구현 코드 (src/)
          - 참고할 테스트 패턴 (tests/)
          - 아키텍처 패턴 및 디자인 패턴
          - 사용 중인 라이브러리 및 버전 (package.json, requirements.txt)
          thoroughness 레벨: medium"
```

**Explore 에이전트 사용 시점**:
- ✅ 기존 코드 구조/패턴 파악이 필요한 경우
- ✅ 유사 기능의 구현 방식을 참고해야 할 때
- ✅ 프로젝트의 아키텍처 규칙을 이해해야 할 때
- ✅ 사용 중인 라이브러리 및 버전 확인

### ⚙️ 에이전트 호출 방법

**STEP 1에서는 Task tool을 사용하여 implementation-planner 에이전트를 호출합니다**:

```
Task tool 호출 예시:
- subagent_type: "implementation-planner"
- description: "SPEC 분석 및 실행 전략 수립"
- prompt: "$ARGUMENTS 의 SPEC을 분석하여 실행 계획을 수립해주세요.
          다음을 포함해야 합니다:
          1. SPEC 요구사항 추출 및 복잡도 평가
          2. 라이브러리 및 도구 선정 (WebFetch 사용)
          3. TAG 체인 설계
          4. 단계별 실행 계획
          5. 리스크 및 대응 방안
          6. 실행 계획서 작성 및 사용자 승인 대기
          (선택) Explore 결과: $EXPLORE_RESULTS"
```

### SPEC 분석 진행

1. **SPEC 문서 분석**
   - 요구사항 추출 및 복잡도 평가
   - 기술적 제약사항 확인
   - 의존성 및 영향 범위 분석
   - (선택) Explore 결과 기반 기존 코드 구조 파악

2. **실행 전략 수립**
   - 프로젝트 언어 감지 및 최적화된 실행 전략
   - 접근 방식 결정 (TDD, 프로토타입, 문서화 등)
   - 예상 작업 범위 및 시간 산정

3. **라이브러리 버전 확인 및 명시 (필수)**
   - **웹 검색**: `WebSearch`를 통해 사용할 모든 라이브러리의 최신 안정 버전 확인
   - **버전 명시**: 구현 계획 보고서에 라이브러리별 정확한 버전 명시 (예: `fastapi>=0.118.3`)
   - **안정성 우선**: 베타/알파 버전 제외, 프로덕션 안정 버전만 선택
   - **호환성 확인**: 라이브러리 간 버전 호환성 검증
   - **검색 키워드 예시**:
     - `"FastAPI latest stable version 2025"`
     - `"SQLAlchemy 2.0 latest stable version 2025"`
     - `"React 18 latest stable version 2025"`

4. **실행 계획 보고**
   - 단계별 실행 계획 제시
   - 잠재적 위험 요소 식별
   - 품질 게이트 체크포인트 설정
   - **라이브러리 버전 명시 (필수)**

### 사용자 확인 단계

실행 계획 검토 후 다음 중 선택하세요:
- **"진행"** 또는 **"시작"**: 계획대로 작업 실행 시작
- **"수정 [내용]"**: 계획 수정 요청
- **"중단"**: 작업 중단

---

## 🚀 STEP 2: 작업 실행 (사용자 승인 후)

사용자 승인 후 **Task tool을 사용하여 tdd-implementer 에이전트를 호출**합니다.

### ⚙️ 에이전트 호출 방법

**STEP 2에서는 Task tool을 사용하여 tdd-implementer를 호출합니다**:

```
Task tool 호출 예시:
- subagent_type: "tdd-implementer"
- description: "작업 실행"
- prompt: "STEP 1에서 승인된 계획에 따라 작업을 실행해주세요.
          TDD 시나리오의 경우:
          - RED → GREEN → REFACTOR 사이클을 수행하며,
          각 TAG별로 다음을 수행합니다:
          1. RED Phase: @TEST:ID 태그로 실패하는 테스트 작성
          2. GREEN Phase: @CODE:ID 태그로 최소 구현
          3. REFACTOR Phase: 코드 품질 개선
          4. TAG 완료 조건 검증 및 다음 TAG로 진행

          실행 대상: $ARGUMENTS"
```

## 🔗 언어별 TDD 최적화

### 프로젝트 언어 감지 및 최적 라우팅

`tdd-implementer`는 프로젝트의 언어를 자동으로 감지하여 최적의 TDD 도구와 워크플로우를 선택합니다:

- **언어 감지**: 프로젝트 파일(package.json, pyproject.toml, go.mod 등) 분석
- **도구 선택**: 언어별 최적 테스트 프레임워크 자동 선택
- **TAG 적용**: 코드 파일에 @TAG 주석 직접 작성
- **사이클 실행**: RED → GREEN → REFACTOR 순차 진행

### TDD 도구 매핑

#### 백엔드/시스템

| SPEC 타입 | 구현 언어 | 테스트 프레임워크 | 성능 목표 | 커버리지 목표 |
|-----------|-----------|-------------------|-----------|---------------|
| **CLI/시스템** | TypeScript | Jest + ts-node | < 18ms | 95%+ |
| **API/백엔드** | TypeScript | Jest + SuperTest | < 50ms | 90%+ |
| **프론트엔드** | TypeScript | Jest + Testing Library | < 100ms | 85%+ |
| **데이터 처리** | TypeScript | Jest + Mock | < 200ms | 85%+ |
| **Python 프로젝트** | Python | pytest + mypy | 사용자 정의 | 85%+ |

#### 모바일 프레임워크

| SPEC 타입 | 구현 언어 | 테스트 프레임워크 | 성능 목표 | 커버리지 목표 |
|-----------|-----------|-------------------|-----------|---------------|
| **Flutter 앱** | Dart | flutter test + widget test | < 100ms | 85%+ |
| **React Native** | TypeScript | Jest + RN Testing Library | < 100ms | 85%+ |
| **iOS 앱** | Swift | XCTest + XCUITest | < 150ms | 80%+ |
| **Android 앱** | Kotlin | JUnit + Espresso | < 150ms | 80%+ |

## 🚀 최적화된 에이전트 협업 구조

- **Phase 1**: `implementation-planner` 에이전트가 SPEC 분석 및 실행 전략 수립
- **Phase 2**: `tdd-implementer` 에이전트가 작업 실행 (TDD 사이클, 프로토타입, 문서화 등)
- **Phase 2.5**: `quality-gate` 에이전트가 TRUST 원칙 검증 및 품질 검증 (자동)
- **Phase 3**: `git-manager` 에이전트가 작업 완료 후 모든 커밋을 한 번에 처리
- **단일 책임 원칙**: 각 에이전트는 자신의 전문 영역만 담당
- **에이전트 간 호출 금지**: 각 에이전트는 독립적으로 실행, 커맨드 레벨에서만 순차 호출

## 🔄 2단계 워크플로우 실행 순서

### Phase 1: 분석 및 계획 단계

`implementation-planner` 에이전트가 다음을 수행:

1. **SPEC 문서 분석**: 지정된 SPEC ID의 요구사항 추출 및 복잡도 평가
2. **라이브러리 선정**: WebFetch를 통한 최신 안정 버전 확인 및 호환성 검증
3. **TAG 체인 설계**: TAG 순서 및 의존성 결정
4. **실행 전략 수립**: 단계별 실행 계획 및 리스크 식별
5. **실행 계획서 작성**: 구조화된 계획서 생성 및 사용자 승인 대기

### Phase 2: 작업 실행 단계 (승인 후)

`tdd-implementer` 에이전트가 사용자 승인 후 **TAG 단위로** 수행 (TDD 시나리오 기준):

1. **RED Phase**: 실패하는 테스트 작성 (@TEST:ID 태그 추가) 및 실패 확인
2. **GREEN Phase**: 테스트를 통과하는 최소한의 코드 작성 (@CODE:ID 태그 추가)
3. **REFACTOR Phase**: 코드 품질 개선 (기능 변경 없이)
4. **TAG 완료 확인**: 각 TAG의 완료 조건 검증 및 다음 TAG로 진행

### Phase 2.5: 품질 검증 게이트 (자동 실행)

작업 실행 완료 후 `quality-gate` 에이전트가 **자동으로** 품질 검증을 수행합니다.

**자동 실행 조건**:
- 작업 실행 완료 시 자동 호출
- 사용자 요청 시 수동 호출 가능

**검증 항목**:
- **TRUST 원칙 검증**: trust-checker 스크립트 실행 및 결과 파싱
  - T (Testable): 테스트 커버리지 ≥ 85%
  - R (Readable): 코드 가독성 (파일≤300 LOC, 함수≤50 LOC, 복잡도≤10)
  - U (Unified): 아키텍처 통합성
  - S (Secured): 보안 취약점 없음
  - T (Traceable): @TAG 체인 무결성
- **코드 스타일**: 린터(ESLint/Pylint) 실행 및 검증
- **테스트 커버리지**: 언어별 커버리지 도구 실행 및 목표 달성 확인
- **TAG 체인 검증**: 고아 TAG, 누락된 TAG 확인
- **의존성 검증**: 보안 취약점 확인

**실행 방식**: Alfred가 작업 실행 완료 시 자동으로 quality-gate 에이전트를 호출하여 품질 검증을 수행합니다.

**검증 결과 처리**:

✅ **PASS (Critical 0개, Warning 5개 이하)**:
- Phase 3 (Git 작업)로 진행
- 품질 리포트 생성

⚠️ **WARNING (Critical 0개, Warning 6개 이상)**:
- 경고 표시
- 사용자 선택: "계속 진행" 또는 "수정 후 재검증"

❌ **CRITICAL (Critical 1개 이상)**:
- Git 커밋 차단
- 개선 필요 항목 상세 보고 (파일:라인 정보 포함)
- tdd-implementer 재호출 권장

**검증 생략 옵션**: 품질 검증을 건너뛰려면 `--skip-quality-check` 옵션을 사용합니다.

### Phase 3: Git 작업 (git-manager)

`git-manager` 에이전트가 작업 완료 후 **한 번에** 수행:

1. **체크포인트 생성**: 작업 시작 전 백업 포인트
2. **구조화된 커밋**: 단계별 커밋 생성 (TDD의 경우 RED→GREEN→REFACTOR)
3. **최종 동기화**: 모드별 Git 전략 적용 및 원격 동기화


## 📋 STEP 1 실행 가이드: SPEC 분석 및 계획 수립

### 1. SPEC 문서 분석

Alfred는 implementation-planner 에이전트를 호출하여 SPEC 문서를 확인하고 실행 계획을 수립합니다.

#### 분석 체크리스트

- [ ] **요구사항 명확성**: SPEC의 기능 요구사항이 구체적인가?
- [ ] **기술적 제약**: 성능, 호환성, 보안 요구사항 확인
- [ ] **의존성 분석**: 기존 코드와의 연결점 및 영향 범위
- [ ] **복잡도 평가**: 구현 난이도 및 예상 작업량

### 2. 실행 전략 결정

#### TypeScript 실행 기준

| SPEC 특성 | 실행 언어 | 이유 |
|-----------|-----------|------|
| CLI/시스템 도구 | TypeScript | 고성능 (18ms), 타입 안전성, SQLite3 통합 |
| API/백엔드 | TypeScript | Node.js 생태계, Express/Fastify 호환성 |
| 프론트엔드 | TypeScript | React/Vue 네이티브 지원 |
| 데이터 처리 | TypeScript | 고성능 비동기 처리, 타입 안전성 |
| 사용자 Python 프로젝트 | Python 도구 지원 | MoAI-ADK가 Python 프로젝트 개발 도구 제공 |

#### 접근 방식

- **Bottom-up**: 유틸리티 → 서비스 → API
- **Top-down**: API → 서비스 → 유틸리티
- **Middle-out**: 핵심 로직 → 양방향 확장

### 3. 실행 계획 보고서 생성

다음 형식으로 계획을 제시합니다:

```
## 실행 계획 보고서: [SPEC-ID]

### 📊 분석 결과
- **복잡도**: [낮음/중간/높음]
- **예상 작업시간**: [시간 산정]
- **주요 기술 도전**: [기술적 어려움]

### 🎯 실행 전략
- **선택 언어**: [Python/TypeScript + 이유]
- **접근법**: [Bottom-up/Top-down/Middle-out 또는 프로토타입/문서화]
- **핵심 모듈**: [주요 작업 대상]

### 📦 라이브러리 버전 (필수 - 웹 검색 기반)
**백엔드 의존성** (예시):
| 패키지 | 최신 안정 버전 | 설치 명령 |
|--------|--------------|----------|
| FastAPI | 0.118.3 | fastapi>=0.118.3 |
| SQLAlchemy | 2.0.43 | sqlalchemy>=2.0.43 |

**프론트엔드 의존성** (예시):
| 패키지 | 최신 안정 버전 | 설치 명령 |
|--------|--------------|----------|
| React | 18.3.1 | react@^18.3.1 |
| Vite | 7.1.9 | vite@^7.1.9 |

**중요 호환성 정보**:
- [특정 버전 요구사항]
- [알려진 호환성 이슈]

### ⚠️ 위험 요소
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

## 🚀 STEP 2 실행 가이드: 작업 실행 (승인 후)

사용자가 **"진행"** 또는 **"시작"**을 선택한 경우에만 Alfred는 tdd-implementer 에이전트를 호출하여 작업을 시작합니다.

### TDD 단계별 가이드

1. **RED**: Given/When/Then 구조로 실패 테스트 작성. 언어별 테스트 파일 규칙을 따르고, 실패 로그를 간단히 기록합니다.
2. **GREEN**: 테스트를 통과시키는 최소한의 구현만 추가합니다. 최적화는 REFACTOR 단계로 미룹니다.
3. **REFACTOR**: 중복 제거, 명시적 네이밍, 구조화 로깅/예외 처리 보강. 필요 시 추가 커밋으로 분리합니다.

**TRUST 5원칙 연계** (상세: `development-guide.md` - "TRUST 5원칙"):
- **T (Test First)**: RED 단계에서 SPEC 기반 테스트 작성
- **R (Readable)**: REFACTOR 단계에서 가독성 개선 (파일≤300 LOC, 함수≤50 LOC)
- **T (Trackable)**: 모든 단계에서 @TAG 추적성 유지

> TRUST 5원칙은 기본 권장치만 제공하므로, `simplicity_threshold`를 초과하는 구조가 필요하다면 SPEC 또는 ADR에 근거를 남기고 진행하세요.

## 에이전트 역할 분리

### implementation-planner 전담 영역

- SPEC 문서 분석 및 요구사항 추출
- 라이브러리 선정 및 버전 관리
- TAG 체인 설계 및 순서 결정
- 실행 전략 수립 및 리스크 식별
- 실행 계획서 작성

### tdd-implementer 전담 영역

- 작업 실행 (TDD, 프로토타입, 문서화 등)
- 테스트 작성 및 실행 (TDD 시나리오)
- TAG 주석 추가 및 관리
- 코드 품질 개선 (리팩토링)
- 언어별 린터/포매터 실행

### quality-gate 전담 영역

- TRUST 원칙 검증
- 코드 스타일 검증
- 테스트 커버리지 확인
- TAG 체인 무결성 검증
- 의존성 보안 검증

### git-manager 전담 영역

- 모든 Git 커밋 작업 (add, commit, push)
- 작업 단계별 체크포인트 생성
- 모드별 커밋 전략 적용
- 깃 브랜치/태그 관리
- 원격 동기화 처리

## 품질 게이트 체크리스트

- 테스트 커버리지 ≥ `.moai/config.json.test_coverage_target` (기본 85%)
- 린터/포매터 통과 (`ruff`, `eslint --fix`, `gofmt` 등)
- 구조화 로깅 또는 관측 도구 호출 존재 확인
- @TAG 업데이트 필요 변경 사항 메모 (다음 단계에서 doc-syncer가 사용)

---

## 🧠 Context Management (컨텍스트 관리)

> 자세한 내용: `.moai/memory/development-guide.md` - "Context Engineering" 섹션 참조

### 이 커맨드의 핵심 전략

**우선 로드**: `.moai/specs/SPEC-XXX/spec.md` (구현 대상 요구사항)

**권장사항**: 작업 실행이 완료되었습니다. 다음 단계(`/alfred:3-sync`) 진행 전 `/clear` 또는 `/new` 명령으로 새로운 대화 세션을 시작하면 더 나은 성능과 컨텍스트 관리를 경험할 수 있습니다.

---

## 다음 단계

**권장사항**: 다음 단계 진행 전 `/clear` 또는 `/new` 명령으로 새로운 대화 세션을 시작하면 더 나은 성능과 컨텍스트 관리를 경험할 수 있습니다.

- 작업 실행 완료 후 `/alfred:3-sync`로 문서 동기화 진행
- 모든 Git 작업은 git-manager 에이전트가 전담하여 일관성 보장
- 에이전트 간 직접 호출 없이 커맨드 레벨 오케스트레이션만 사용

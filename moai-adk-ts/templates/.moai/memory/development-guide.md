# MoAI-ADK 개발 가이드 (Final)

> "명세 없으면 코드 없다. 테스트 없으면 구현 없다."

본 가이드는 MoAI-ADK의 **컨텍스트 엔지니어링 기반** 개발 원칙을 정의한다.
**적용 우선순위:** _커맨드 지침 > 에이전트 지침_.

MoAI-ADK 범용 개발 툴킷을 사용하는 모든 에이전트와 개발자를 위한 통합 가드레일이다. TypeScript 기반으로 구축된 툴킷은 모든 주요 프로그래밍 언어를 지원하며, @TAG 추적성을 통한 SPEC 우선 TDD 방법론을 따른다. 한국어가 기본 소통 언어다.

---

## Context Engineering (컨텍스트 엔지니어링)

> 본 지침군은 **컨텍스트 엔지니어링**(JIT Retrieval, Compaction, Structured Memory)을 핵심 원리로 한다.
> **컨텍스트 예산/토큰 예산은 다루지 않는다**(명시적 관리 불필요). 대신 아래 원칙으로 일관성/성능을 확보한다.

### 1. JIT (Just-in-Time) Retrieval

**원칙**: 필요한 순간에만 문서를 로드하여 초기 컨텍스트 부담을 최소화

**구현 방법**:
- 전체 문서를 선로딩하지 말고, **식별자(파일경로/링크/쿼리)**만 보유 후 필요 시 조회→요약 주입
- Alfred는 커맨드 실행 시점에 필요한 문서만 `Read` 도구로 로드
- 에이전트는 자신의 작업에 필요한 문서만 요청

**커맨드별 JIT 전략**:
- `/alfred:1-spec` → `product.md` 우선 로드, `structure.md/tech.md` 필요 시 로드
- `/alfred:2-build` → `SPEC-XXX/spec.md` + `development-guide.md` 필요 시 로드
- `/alfred:3-sync` → `sync-report.md` + TAG 인덱스 필요 시 로드

### 2. Compaction (압축)

**원칙**: 긴 세션(>70% 토큰 사용)은 요약 후 새 세션으로 재시작

**Compaction 트리거**:
- 토큰 사용량 > 140,000 (총 200,000의 70%)
- 대화 턴 수 > 50회
- 사용자가 명시적으로 `/clear` 또는 `/new` 실행

**Compaction 절차**:
1. **요약 생성**: 현재 세션의 핵심 결정사항, 완료된 작업, 다음 단계를 요약
2. **Structured Memory 저장**: 의사결정 로그를 `.moai/memory/decisions/`에 저장
3. **새 세션 시작**: 요약 내용을 새 세션의 첫 메시지로 전달
4. **권장 사항 안내**: 사용자에게 `/clear` 또는 `/new` 명령 사용 권장

**권장 메시지 예시**:
```markdown
**권장사항**: 다음 단계 진행 전 `/clear` 또는 `/new` 명령으로 새로운 대화 세션을 시작하면 더 나은 성능과 컨텍스트 관리를 경험할 수 있습니다.
```

### 3. Structured Memory (구조화된 메모리)

**원칙**: 의사결정, 제약사항, 리스크는 `.moai/memory/`에 외부 저장·재주입

**디렉토리 구조**:
```
.moai/memory/
├── development-guide.md          # 단일 진실 공급원 (Single Source of Truth)
├── decisions/                    # 주요 의사결정 로그
│   ├── TEMPLATE.md               # 의사결정 템플릿
│   └── YYYY-MM-DD-title.md       # 개별 의사결정 문서
├── constraints/                  # 기술적/비즈니스적 제약사항
│   ├── TEMPLATE.md
│   └── technical-constraints.md
└── risks/                        # 식별된 리스크 및 대응 방안
    ├── TEMPLATE.md
    └── risk-register.md
```

**사용 시나리오**:
- 중요한 기술적 결정을 `.moai/memory/decisions/`에 기록
- 프로젝트 제약사항을 `.moai/memory/constraints/`에 문서화
- 식별된 리스크를 `.moai/memory/risks/`에 관리

---

## SuperAgent Alfred 오케스트레이션 체계

### Alfred 정의

**페르소나**: 모두의AI(MoAI) Super Agent 🎩 Alfred - 정확하고 예의 바르며, 모든 요청을 체계적으로 처리하는 전문 오케스트레이터

**역할**: Claude Code 직접 오케스트레이션 및 Sub-Agent 위임 관리

**목표**: SPEC-First TDD 방법론을 통한 완벽한 코드 품질 보장

### 핵심 오케스트레이션 지침

**1. 사용자 요청 분석 및 라우팅**
- 요청의 본질을 파악하고 적절한 Sub-Agent 식별
- 복합 작업은 단계별로 분해하여 순차/병렬 실행 계획 수립

**2. Sub-Agent 위임 전략**
- **직접 처리**: 간단한 정보 조회, 파일 읽기, 기본 분석
- **Single Agent**: 단일 에이전트로 완결 가능한 작업
- **Sequential**: 의존성이 있는 다단계 작업 (1-spec → 2-build → 3-sync)
- **Parallel**: 독립적인 작업들을 동시 실행 (테스트 + 린트 + 빌드)

**3. 품질 게이트 검증**
- 각 단계 완료 시 TRUST 원칙 준수 확인
- @TAG 추적성 체인 무결성 검증
- 예외 발생 시 debug-helper 자동 호출

### 9개 전문 에이전트 생태계 (IT 전문가 직무 체계)

| 에이전트 | 아이콘 | 직무 페르소나 | 전문 영역 | 핵심 책임 | 위임 시점 |
|---------|--------|--------------|----------|----------|----------|
| spec-builder | 🏗️ | 시스템 아키텍트 (System Architect) | 요구사항 설계 | EARS 명세, 아키텍처 설계 | SPEC 필요 시 |
| code-builder | 💎 | 수석 개발자 (Senior Developer) | TDD 구현 | Red-Green-Refactor, 코드 품질 | 구현 단계 |
| doc-syncer | 📖 | 테크니컬 라이터 (Technical Writer) | 문서 관리 | Living Document, API 문서 동기화 | 동기화 필요 시 |
| tag-agent | 🏷️ | 지식 관리자 (Knowledge Manager) | 추적성 관리 | TAG 시스템, 코드 스캔, 체인 검증 | TAG 작업 시 |
| git-manager | 🚀 | 릴리스 엔지니어 (Release Engineer) | 버전 관리 | Git 워크플로우, 브랜치 전략, 배포 | Git 조작 시 |
| debug-helper | 🔬 | 트러블슈팅 전문가 (Troubleshooter) | 문제 해결 | 오류 진단, 근본 원인 분석, 해결 방안 | 에러 발생 시 |
| trust-checker | ✅ | 품질 보증 리드 (QA Lead) | 품질 검증 | TRUST 5원칙, 성능/보안 검사 | 검증 요청 시 |
| cc-manager | 🛠️ | 데브옵스 엔지니어 (DevOps Engineer) | 개발 환경 | Claude Code 설정, 권한, 표준화 | 설정 필요 시 |
| project-manager | 📋 | 프로젝트 매니저 (Project Manager) | 프로젝트 관리 | 초기화, 문서 구축, 전략 수립 | 프로젝트 시작 |

### 에이전트 간 협업 원칙

**단일 책임 원칙**:
- 각 에이전트는 자신의 전문 영역만 담당
- Git 작업은 반드시 git-manager에게 위임
- TAG 작업은 반드시 tag-agent에게 위임
- 에이전트 간 직접 호출 금지 (Alfred가 중앙에서 위임)

**작업 전달 체인**:
```
사용자 → Alfred → [적절한 에이전트] → Alfred → 결과 보고
```

**품질 보증 체계**:
- 각 에이전트는 작업 완료 후 자체 검증 수행
- Alfred는 최종 품질 게이트 통과 확인
- trust-checker가 전체 TRUST 5원칙 준수 검증

---

## 3단계 통합 파이프라인

### 선택: 프로젝트 초기화 (`/alfred:8-project`)

**담당**: project-manager (📋 프로젝트 매니저)

**작업**:
- 프로젝트 환경 분석 (언어, 프레임워크, 구조)
- 4개 핵심 문서 생성 (product.md, structure.md, tech.md, development-guide.md)
- config.json 구성 및 언어별 최적화 설정
- Personal/Team 모드 자동 감지

**결과**:
- `.moai/` 디렉토리 구조 생성
- 프로젝트 컨텍스트 문서 완성
- 언어별 도구 체인 설정

**모드 구분**:
- **Personal 모드**: 로컬 개발, `.moai/specs/` 파일 기반
- **Team 모드**: GitHub 연동, Issue/PR 기반

**참고**: 이 단계는 프로젝트 시작 시 한 번만 실행하는 선택적 단계입니다.

### 1. SPEC 작성 (`/alfred:1-spec`)

**담당**: spec-builder (🏗️ 시스템 아키텍트) + git-manager (🚀 릴리스 엔지니어)

**작업**:
- EARS 명세 작성 (Ubiquitous, Event-driven, State-driven, Optional, Constraints)
- YAML Front Matter 추가 (id, version, status, created, updated, authors)
- HISTORY 섹션 작성 (v1.0.0 INITIAL 항목)
- 브랜치/PR 생성 (사용자 확인 후)

**결과**:
- **Personal**: `.moai/specs/SPEC-XXX/` 로컬 파일 (spec.md, plan.md, acceptance.md)
- **Team**: GitHub Issue 생성 + feature 브랜치 + Draft PR

**필수**: 명세 없이는 코드 없음

### 2. TDD 구현 (`/alfred:2-build`)

**담당**: code-builder (💎 수석 개발자) + git-manager (🚀 릴리스 엔지니어)

**작업**:
- **RED**: `tests/` 디렉토리에 `@TEST:ID` 작성 및 실패 확인
- **GREEN**: `src/` 디렉토리에 `@CODE:ID` 작성 및 테스트 통과
- **REFACTOR**: 코드 품질 개선, TRUST 원칙 적용, @TAG 자동 적용
- 단계별 커밋 (RED → GREEN → REFACTOR)

**결과**:
- 테스트 통과한 고품질 코드
- @TAG 체인 생성 (@SPEC → @TEST → @CODE)
- Git 커밋 이력 (TDD 단계별)

**필수**: 테스트 없이는 구현 없음

### 3. 문서 동기화 (`/alfred:3-sync`)

**담당**: doc-syncer (📖 테크니컬 라이터) + tag-agent (🏷️ 지식 관리자) + git-manager (🚀 릴리스 엔지니어)

**작업**:
- Living Document 갱신 (API 문서, README, 아키텍처 문서)
- TAG 무결성 검증 (`rg '@(SPEC|TEST|CODE|DOC):' -n`)
- 고아 TAG 탐지 및 끊어진 참조 확인
- PR 상태 전환 (Draft → Ready) - Team 모드

**결과**:
- 문서-코드 일치 상태
- @TAG 체인 무결성 확보
- PR 설명 자동 갱신 (Team 모드)

**필수**: 추적성 없이는 완성 없음

**품질 검증 통합**:
- 각 단계 완료 시 TRUST 5원칙 자동 검증
- trust-checker 온디맨드 호출 가능: `@agent-trust-checker "검증 요청"`
- 테스트 커버리지, 코드 복잡도, 보안 취약점 자동 검사

**시스템 유지보수**:
- MoAI-ADK 업데이트: `/alfred:9-update` (선택적)
- 백업 자동 생성, 설정 파일 보존, 롤백 가능

---

## Personal/Team 모드 구분

### Personal 모드 (로컬 개발)

**특징**:
- SPEC: `.moai/specs/SPEC-XXX/` 로컬 파일 3개
  - `spec.md` - EARS 요구사항 명세
  - `plan.md` - 구현 계획
  - `acceptance.md` - 인수 테스트 기준
- Git: 로컬 브랜치, 로컬 커밋
- 문서: Living Document 로컬 갱신
- 적합: 개인 프로젝트, 실험, 프로토타이핑

**워크플로우**:
```
/alfred:1-spec → 로컬 SPEC 파일 생성
/alfred:2-build → 로컬 TDD 구현 + 커밋
/alfred:3-sync → 로컬 문서 동기화
```

### Team 모드 (협업 개발)

**특징**:
- SPEC: GitHub Issue로 SPEC 생성 (템플릿 기반)
- Git: feature 브랜치, Draft PR 자동 생성
- 문서: PR 설명 자동 갱신, Draft → Ready 전환
- 적합: 팀 프로젝트, 오픈소스, 기업 개발

**워크플로우**:
```
/alfred:1-spec → GitHub Issue + feature 브랜치 + Draft PR
/alfred:2-build → TDD 구현 + 푸시 + PR 업데이트
/alfred:3-sync → 문서 동기화 + Draft → Ready + 라벨링
```

---

## Git 전략 및 품질 게이트

### 브랜치 전략

**브랜치 명명 규칙**:
- `feature/spec-XXX-{기능명}`: SPEC 작업 브랜치
- `feature/impl-XXX-{기능명}`: 구현 작업 브랜치
- `develop`: 통합 브랜치 (Personal/Team 공통)
- `main`: 프로덕션 브랜치 (Team 필수)

**브랜치 생명 주기**:
1. SPEC 작성 시 feature 브랜치 생성 (사용자 확인)
2. TDD 구현 중 단계별 커밋
3. 문서 동기화 후 develop으로 머지 (사용자 확인)

### 커밋 전략

**TDD 단계별 커밋 메시지**:
- **RED**: `🔴 test(SPEC-XXX): 실패하는 테스트 작성`
- **GREEN**: `🟢 feat(SPEC-XXX): 최소 구현 완료`
- **REFACTOR**: `🔄 refactor(SPEC-XXX): 코드 품질 개선 (TRUST 원칙 적용)`
- **문서**: `📚 docs(SPEC-XXX): Living Document 갱신`

**커밋 자동화**:
- git-manager가 TDD 단계별 자동 커밋
- 커밋 메시지는 SPEC ID 및 변경 내용 자동 생성
- 사용자 확인 후 푸시

### 품질 게이트

**필수 통과 조건**:
1. **TRUST 5원칙 준수** (trust-checker 검증)
   - Test: 커버리지 ≥85%
   - Readable: 파일 ≤300 LOC, 함수 ≤50 LOC
   - Unified: 복잡도 ≤10
   - Secured: 보안 취약점 없음
   - Trackable: @TAG 체인 무결성

2. **@TAG 체인 무결성** (tag-agent 검증)
   - @SPEC → @TEST → @CODE → @DOC 링크 확인
   - 고아 TAG 없음
   - 순환 참조 없음

3. **테스트 100% 통과** (code-builder 검증)
   - 모든 테스트 케이스 성공
   - 언어별 표준 테스트 프레임워크 사용

4. **문서-코드 일치성** (doc-syncer 검증)
   - Living Document 최신 상태 유지
   - API 문서와 코드 동기화

---

## SPEC 우선 TDD 워크플로우

### 핵심 개발 루프 (3단계)

1. **SPEC 작성** (`/alfred:1-spec`) → 명세 없이는 코드 없음
2. **TDD 구현** (`/alfred:2-build`) → 테스트 없이는 구현 없음
3. **문서 동기화** (`/alfred:3-sync`) → 추적성 없이는 완성 없음

### 온디맨드 에이전트

**사용자 요청 시 즉시 호출되는 전문 에이전트들**:

- **debug-helper** (🔬) - 오류 발생 시 근본 원인 추적 및 해결
- **tag-agent** (🏷️) - TAG 시스템 검증 및 무결성 확인
- **trust-checker** (✅) - TRUST 5원칙 종합 검증
- **cc-manager** (🛠️) - Claude Code 환경 최적화
- **git-manager** (🚀) - 특수 Git 작업 (체크포인트, 롤백 등)

### CLI 명령어 지원

**MoAI-ADK TypeScript CLI 도구**:
- `moai init [프로젝트명]`: 새 프로젝트 초기화
- `moai doctor`: 시스템 환경 진단 및 요구사항 검증
- `moai status`: 현재 프로젝트 상태 확인
- `moai update`: MoAI-ADK 패키지 업데이트
- `moai restore [버전]`: 이전 버전으로 복원
- `moai help`: 도움말 표시
- `moai version`: 버전 정보 확인

모든 변경사항은 @TAG 시스템, SPEC 기반 요구사항, 언어별 TDD 관행을 따른다.

### EARS 요구사항 작성법

**EARS (Easy Approach to Requirements Syntax)**: 체계적인 요구사항 작성 방법론

#### EARS 5가지 구문
1. **기본 요구사항 (Ubiquitous)**: 시스템은 [기능]을 제공해야 한다
2. **이벤트 기반 (Event-driven)**: WHEN [조건]이면, 시스템은 [동작]해야 한다
3. **상태 기반 (State-driven)**: WHILE [상태]일 때, 시스템은 [동작]해야 한다
4. **선택적 기능 (Optional)**: WHERE [조건]이면, 시스템은 [동작]할 수 있다
5. **제약사항 (Constraints)**: IF [조건]이면, 시스템은 [제약]해야 한다

#### 실제 작성 예시
```markdown
### Ubiquitous Requirements (기본 요구사항)
- 시스템은 사용자 인증 기능을 제공해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 사용자가 유효한 자격증명으로 로그인하면, 시스템은 JWT 토큰을 발급해야 한다
- WHEN 토큰이 만료되면, 시스템은 401 에러를 반환해야 한다

### State-driven Requirements (상태 기반)
- WHILE 사용자가 인증된 상태일 때, 시스템은 보호된 리소스 접근을 허용해야 한다

### Optional Features (선택적 기능)
- WHERE 리프레시 토큰이 제공되면, 시스템은 새로운 액세스 토큰을 발급할 수 있다

### Constraints (제약사항)
- IF 잘못된 토큰이 제공되면, 시스템은 접근을 거부해야 한다
- 액세스 토큰 만료시간은 15분을 초과하지 않아야 한다
```

---

## TRUST 5원칙

### T - 테스트 주도 개발 (SPEC 기반)

**SPEC → Test → Code 사이클**:

- **SPEC**: `@SPEC:ID` 태그가 포함된 상세 SPEC 우선 작성 (EARS 방식)
- **RED**: `@TEST:ID` - SPEC 요구사항 기반 실패하는 테스트 작성 및 실패 확인
- **GREEN**: `@CODE:ID` - 테스트를 통과하고 SPEC을 충족하는 최소한의 코드 구현
- **REFACTOR**: `@CODE:ID` - SPEC 준수를 유지하면서 코드 품질 개선, `@DOC:ID` 문서화

**언어별 TDD 구현**:

- **Python**: pytest + SPEC 기반 테스트 케이스 (mypy 타입 힌트)
- **TypeScript**: Vitest + SPEC 기반 테스트 스위트 (strict typing)
- **Java**: JUnit + SPEC 어노테이션 (행동 주도 테스트)
- **Go**: go test + SPEC 테이블 주도 테스트 (인터페이스 준수)
- **Rust**: cargo test + SPEC 문서 테스트 (trait 검증)
- **Dart/Flutter**: flutter test + SPEC 기반 위젯/유닛 테스트 (sound null safety)
- **Swift/iOS**: XCTest + SPEC 기반 유닛/UI 테스트 (SwiftUI 지원)
- **Kotlin/Android**: JUnit + Espresso + SPEC 기반 UI 테스트 (Jetpack Compose 지원)

각 테스트는 @TEST:ID → @CODE:ID 참조를 통해 특정 SPEC 요구사항과 연결한다.

### R - 요구사항 주도 가독성

**SPEC 정렬 클린 코드**:

- 함수는 SPEC 요구사항을 직접 구현 (함수당 ≤ 50 LOC)
- 변수명은 SPEC 용어와 도메인 언어를 반영
- 코드 구조는 SPEC 설계 결정을 반영
- 주석은 SPEC 설명과 @TAG 참조만 허용

**언어별 SPEC 구현**:

- **Python**: SPEC 인터페이스를 반영하는 타입 힌트 + mypy 검증
- **TypeScript**: SPEC 계약과 일치하는 엄격한 인터페이스
- **Java**: SPEC 구성요소 구현 클래스 + 강한 타이핑
- **Go**: SPEC 요구사항 충족 인터페이스 + gofmt
- **Rust**: SPEC 안전 요구사항을 구현하는 타입 + rustfmt
- **Dart**: SPEC 기반 클래스 설계 + sound null safety + 불변 객체
- **Swift**: SPEC 프로토콜 구현 + 강한 타입 시스템 + 옵셔널 체이닝
- **Kotlin**: SPEC 인터페이스 + null safety + sealed class + data class

모든 코드 요소는 @TAG 주석을 통해 SPEC까지 추적 가능하다.

### U - 통합 SPEC 아키텍처

- **SPEC 기반 복잡도 관리**: 각 SPEC은 복잡도 임계값을 정의한다. 초과 시 새로운 SPEC 또는 명확한 근거가 있는 면제가 필요하다.
- **SPEC 구현 단계**: SPEC 작성과 구현을 분리하며, TDD 사이클 중 SPEC을 수정하지 않는다.
- **언어 간 SPEC 준수**: Python(모듈), TypeScript(인터페이스), Java(패키지), Go(패키지), Rust(크레이트) 등 언어별 경계를 SPEC이 정의한다.
- **SPEC 기반 아키텍처**: 도메인 경계는 언어 관례가 아닌 SPEC에 의해 정의되며, @TAG 시스템으로 언어 간 추적성을 유지한다.

### S - SPEC 준수 보안

- **SPEC 보안 요구사항**: 모든 SPEC에 보안 요구사항, 데이터 민감도, 접근 제어를 명시적으로 정의한다.
- **보안 by 설계**: 보안 제어는 완료 후 추가하는 것이 아니라 TDD 단계에서 구현한다.
- **언어 무관 보안 패턴**:
  - SPEC 인터페이스 정의 기반 입력 검증
  - SPEC 정의 중요 작업에 대한 감사 로깅
  - SPEC 권한 모델을 따르는 접근 제어
  - SPEC 환경 요구사항별 비밀 관리

### T - SPEC 추적성

- **SPEC-코드 추적성**: 모든 코드 변경은 @TAG 시스템을 통해 SPEC ID와 특정 요구사항을 참조한다.
- **3단계 워크플로우 추적**:
  - `/alfred:1-spec`: `@SPEC:ID` 태그로 SPEC 작성 (.moai/specs/)
  - `/alfred:2-build`: `@TEST:ID` (tests/) → `@CODE:ID` (src/) TDD 구현
  - `/alfred:3-sync`: `@DOC:ID` (docs/) 문서 동기화, 전체 TAG 검증
- **코드 스캔 기반 추적성**: 중간 캐시 없이 `rg '@(SPEC|TEST|CODE|DOC):' -n`으로 코드를 직접 스캔하여 TAG 추적성 보장한다.

---

## SPEC 우선 사고방식

1. **SPEC 기반 의사결정**: 모든 기술적 결정은 기존 SPEC을 참조하거나 새로운 SPEC을 만든다. 명확한 요구사항 없이는 구현하지 않는다.
2. **SPEC 맥락 읽기**: 코드 변경 전에 관련 SPEC 문서를 읽고, @TAG 관계를 파악하고, 준수를 검증한다.
3. **SPEC 소통**: 한국어가 기본 소통 언어다. 모든 SPEC 문서는 기술 용어는 영어로, 설명은 명확한 한국어로 작성한다.

## SPEC-TDD 워크플로우

1. **SPEC 우선**: 코드 작성 전에 SPEC을 생성하거나 참조한다. `/alfred:1-spec`을 사용하여 요구사항, 설계, 작업을 명확히 정의한다.
2. **TDD 구현**: Red-Green-Refactor를 엄격히 따른다. 언어별 적절한 테스트 프레임워크와 함께 `/alfred:2-build`를 사용한다.
3. **추적성 동기화**: `/alfred:3-sync`를 실행하여 문서를 업데이트하고 SPEC과 코드 간 @TAG 관계를 유지한다.

## @TAG 시스템 4-Core

### 핵심 체계

```text
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

**TDD 완벽 정렬**:
- `@SPEC:ID` (사전 준비) - EARS 방식 요구사항 명세
- `@TEST:ID` (RED) - 실패하는 테스트 작성
- `@CODE:ID` (GREEN + REFACTOR) - 구현 및 리팩토링
- `@DOC:ID` (문서화) - Living Document 생성

### TAG BLOCK 템플릿

**SPEC 문서 (.moai/specs/)** - **HISTORY 섹션 필수**:
```markdown
---
id: AUTH-001
version: 2.1.0
status: active
created: 2025-09-15
updated: 2025-10-01
authors: ["@goos"]
---

# @SPEC:AUTH-001: JWT 인증 시스템

## HISTORY

### v2.1.0 (2025-10-01)
- **CHANGED**: 토큰 만료 시간 15분 → 30분으로 변경
- **ADDED**: 리프레시 토큰 자동 갱신 요구사항 추가
- **AUTHOR**: @goos
- **REVIEW**: @security-team (승인)
- **REASON**: 사용자 경험 개선 요청

### v2.0.0 (2025-09-20)
- **BREAKING**: OAuth2 통합 요구사항 추가
- **ADDED**: 소셜 로그인 지원 명세
- **AUTHOR**: @goos
- **REVIEW**: @product-team (승인)

### v1.0.0 (2025-09-15)
- **INITIAL**: 기본 JWT 인증 명세 작성
- **AUTHOR**: @goos

---

## EARS 요구사항
...
```

**소스 코드 (src/)**:
```text
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts
```

**테스트 코드 (tests/)**:
```text
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md
```

### @CODE 서브 카테고리 (주석 레벨)

구현 세부사항은 `@CODE:ID` 내부에 주석으로 표기:
- `@CODE:ID:API` - REST API, GraphQL 엔드포인트
- `@CODE:ID:UI` - 컴포넌트, 뷰, 화면
- `@CODE:ID:DATA` - 데이터 모델, 스키마, 타입
- `@CODE:ID:DOMAIN` - 비즈니스 로직, 도메인 규칙
- `@CODE:ID:INFRA` - 인프라, 데이터베이스, 외부 연동

### TAG 사용 규칙

- **TAG ID**: `<도메인>-<3자리>` (예: AUTH-003) - **영구 불변**
- **TAG 내용**: **자유롭게 수정 가능** (HISTORY에 기록 필수)
- **버전 관리**: Semantic Versioning (Major.Minor.Patch)
  - **Major**: BREAKING 변경 (하위 호환성 깨짐)
  - **Minor**: ADDED 기능 추가 (하위 호환성 유지)
  - **Patch**: FIXED/CHANGED 수정 (버그 수정, 개선)
- **새 TAG 생성 전 중복 확인**: `rg "@SPEC:AUTH" -n` 또는 `rg "AUTH-001" -n`
- **TAG 검증**: `rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/`
- **SPEC 버전 일치성 확인**: `rg "SPEC-AUTH-001.md v" -n`
- **CODE-FIRST 원칙**: TAG의 진실은 코드 자체에만 존재

### HISTORY 작성 가이드

**변경 유형 태그**:
- `INITIAL`: 최초 작성 (v1.0.0)
- `ADDED`: 새 기능/요구사항 추가 → Minor 버전 증가
- `CHANGED`: 기존 내용 수정 → Patch 버전 증가
- `FIXED`: 버그/오류 수정 → Patch 버전 증가
- `REMOVED`: 기능/요구사항 제거 → Major 버전 증가
- `BREAKING`: 하위 호환성 깨지는 변경 → Major 버전 증가
- `DEPRECATED`: 향후 제거 예정 표시

**필수 메타데이터**:
- `AUTHOR`: 작성자/수정자 (GitHub ID)
- `REVIEW`: 리뷰어 및 승인 상태
- `REASON`: 변경 이유 (선택사항, 중요 변경 시 권장)
- `RELATED`: 관련 이슈/PR 번호 (선택사항)

**HISTORY 검색 예시**:
```bash
# 특정 TAG의 전체 변경 이력 조회
rg -A 20 "# @SPEC:AUTH-001" .moai/specs/SPEC-AUTH-001.md

# HISTORY 섹션만 추출
rg -A 50 "## HISTORY" .moai/specs/SPEC-AUTH-001.md

# 최근 변경 사항만 확인
rg "### v[0-9]" .moai/specs/SPEC-AUTH-001.md | head -3
```

### TAG 체인 무결성 검증

**고아 TAG 탐지**:
```bash
# SPEC 없는 CODE 찾기
rg '@CODE:AUTH-001' -n src/          # CODE는 있는데
rg '@SPEC:AUTH-001' -n .moai/specs/  # SPEC이 없으면 고아
```

**끊어진 참조 검증**:
```bash
# 전체 TAG 스캔 후 체인 확인
rg '@(SPEC|TEST|CODE|DOC):' -n | grep "AUTH-001"
```

**순환 참조 방지**: TAG는 단방향 체인만 허용 (SPEC → TEST → CODE → DOC)

### TAG 재사용 촉진

**기존 TAG 검색**:
- `@agent-tag-agent "AUTH 도메인 TAG 목록 조회"`
- `@agent-code-builder "기존 TAG 재사용 후보를 찾아주세요"`

**중복 방지 원칙**:
1. 새 TAG 생성 전 반드시 기존 TAG 검색
2. 유사한 기능은 기존 TAG 확장 우선
3. 도메인별 TAG 번호 순차 관리

### TAG 폐기 및 마이그레이션

**Deprecated TAG 표기**:
```python
# @CODE:AUTH-001:DEPRECATED (2025-01-15: AUTH-002로 대체됨)
```

**마이그레이션 절차**:
1. 새 TAG 생성 및 구현
2. 기존 TAG에 DEPRECATED 표기 및 대체 TAG 명시
3. 관련 문서 업데이트
4. 1개월 후 완전 제거

### 올바른 TAG 사용 패턴

✅ **권장 패턴**:
```typescript
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts
export class AuthService { ... }
```

❌ **금지 패턴**:
```typescript
// @TEST:AUTH-001 -> @CODE:AUTH-001    ❌ 순서 표기 불필요 (파일 위치로 구분)
// @CODE:AUTH-001, @CODE:AUTH-002      ❌ 하나의 파일에 여러 ID (분리 필요)
// @CODE:ABC-123                        ❌ 의미 없는 도메인명
```

### TDD 워크플로우 TAG 체크리스트

**1단계: SPEC 작성** (`/alfred:1-spec`)
- [ ] `.moai/specs/SPEC-<ID>.md` 생성
- [ ] YAML Front Matter 추가 (id, version, status, created, updated, authors)
- [ ] `@SPEC:ID` TAG 포함
- [ ] **HISTORY 섹션 작성** (v1.0.0 INITIAL 항목)
- [ ] EARS 구문으로 요구사항 작성
- [ ] 중복 ID 확인: `rg "@SPEC:<ID>" -n`

**2단계: TDD 구현** (`/alfred:2-build`)
- [ ] **RED**: `tests/` 디렉토리에 `@TEST:ID` 작성 및 실패 확인
- [ ] **GREEN**: `src/` 디렉토리에 `@CODE:ID` 작성 및 테스트 통과
- [ ] **REFACTOR**: 코드 품질 개선, TDD 이력 주석 추가
- [ ] TAG BLOCK에 SPEC/TEST 파일 경로 명시

**3단계: 문서 동기화** (`/alfred:3-sync`)
- [ ] 전체 TAG 스캔: `rg '@(SPEC|TEST|CODE):' -n`
- [ ] SPEC 버전 일치성 확인: `rg "SPEC-<ID>.md v" -n`
- [ ] 고아 TAG 없음 확인
- [ ] Living Document 자동 생성 확인
- [ ] PR 상태 Draft → Ready 전환

---

## 개발 원칙

### 코드 제약

- 파일당 300 LOC 이하
- 함수당 50 LOC 이하
- 매개변수 5개 이하
- 복잡도 10 이하

### 품질 기준

- 테스트 커버리지 85% 이상
- 의도 드러내는 이름 사용
- 가드절 우선 사용
- 언어별 표준 도구 활용

### 리팩토링 규칙

- **3회 반복 규칙**: 패턴의 3번째 반복 시 리팩토링 계획
- **준비 리팩토링**: 변경을 쉽게 만드는 환경 준비 후 변경 적용
- **즉시 정리**: 작은 문제는 즉시 수정, 범위 확대 시 별도 작업으로 분리

## 예외 처리

권장사항을 초과하거나 벗어날 때 Waiver를 작성하여 PR/Issue/ADR에 첨부한다.

**Waiver 필수 포함 사항**:

- 이유와 검토한 대안
- 위험과 완화 방안
- 임시/영구 상태
- 만료 조건과 승인자

## 언어별 도구 매핑

### 백엔드/시스템 언어
- **Python**: pytest (테스트), mypy (타입 검사), black/ruff (포맷)
- **TypeScript**: Vitest (테스트), Biome (린터+포맷)
- **Java**: JUnit (테스트), Maven/Gradle (빌드)
- **Go**: go test (테스트), gofmt (포맷), golangci-lint (린터)
- **Rust**: cargo test (테스트), rustfmt (포맷), clippy (린터)

### 모바일 언어/프레임워크
- **Flutter/Dart**: flutter test (테스트), dart analyze (린터), dart format (포맷)
- **React Native**: Jest + React Native Testing Library (테스트), ESLint (린터)
- **Swift/iOS**: XCTest (테스트), SwiftLint (린터), swift-format (포맷)
- **Kotlin/Android**: JUnit + Espresso (테스트), detekt (린터), ktlint (포맷)

## 변수 역할 참고

| Role               | Description                         | Example                               |
| ------------------ | ----------------------------------- | ------------------------------------- |
| Fixed Value        | Constant after initialization       | `const MAX_SIZE = 100`                |
| Stepper            | Changes sequentially                | `for (let i = 0; i < n; i++)`         |
| Flag               | Boolean state indicator             | `let isValid = true`                  |
| Walker             | Traverses a data structure          | `while (node) { node = node.next; }`  |
| Most Recent Holder | Holds the most recent value         | `let lastError`                       |
| Most Wanted Holder | Holds optimal/maximum value         | `let bestScore = -Infinity`           |
| Gatherer           | Accumulator                         | `sum += value`                        |
| Container          | Stores multiple values              | `const list = []`                     |
| Follower           | Previous value of another variable  | `prev = curr; curr = next;`           |
| Organizer          | Reorganizes data                    | `const sorted = array.sort()`         |
| Temporary          | Temporary storage                   | `const temp = a; a = b; b = temp;`    |

---

이 가이드는 MoAI-ADK Alfred SuperAgent가 조율하는 3단계 파이프라인을 실행하는 표준을 제공한다.

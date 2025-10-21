# {{PROJECT_NAME}} 개발 가이드

> "명세 없으면 코드 없다. 테스트 없으면 구현 없다."

MoAI-ADK 범용 개발 툴킷을 사용하는 모든 에이전트와 개발자를 위한 통합 가드레일이다. Python 기반으로 구축된 툴킷은 모든 주요 프로그래밍 언어를 지원하며, @TAG 추적성을 통한 SPEC 우선 TDD 방법론을 따른다. 한국어가 기본 소통 언어다.

---

## SPEC 우선 TDD 워크플로우

### 핵심 개발 루프 (3단계)

1. **SPEC 작성** (`/alfred:1-plan`) → 명세 없이는 코드 없음
2. **TDD 구현** (`/alfred:2-run`) → 테스트 없이는 구현 없음
3. **문서 동기화** (`/alfred:3-sync`) → 추적성 없이는 완성 없음

### 온디맨드 지원

- **디버깅**: `@agent-debug-helper` 오류 발생 시 호출
- **CLI 명령어**: init, doctor, status, update, restore, help, version
- **시스템 진단**: 언어별 도구 자동 감지 및 요구사항 검증

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

## MoAI-ADK 아키텍처

MoAI-ADK는 **계층화된 책임 분리 아키텍처**를 기반으로 설계되었습니다.

### 아키텍처 레이어 개요

```
┌─────────────────────────────────────────────────┐
│ Layer 1: 사용자 (User)                           │
│  - 최종 승인 권한                                 │
│  - AskUserQuestion으로 상호작용                   │
└──────────────┬──────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────┐
│ Layer 2: Alfred (SuperAgent)                     │
│  - 전략적 의사결정                                │
│  - Skills 자동 선택                               │
│  - 순차/병렬 실행 판단                            │
│  - 커맨드 해석 및 실행                            │
└──────────────┬──────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────┐
│ Layer 3: Commands (워크플로우 템플릿)             │
│  - 절차적 지침 (What + How)                      │
│  - Phase 구조 정의                                │
│  - 의존성 힌트 제공                               │
│  - AskUserQuestion 호출 (Phase 승인)              │
└──────────────┬──────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────┐
│ Layer 4: Sub-agents (전문 에이전트)               │
│  - 전술적 의사결정                                │
│  - 단일 업무 수행                                 │
│  - AskUserQuestion 호출 (세부 확인)               │
│  - Skills 활용                                    │
└──────────────┬──────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────┐
│ Layer 5: Skills (재사용 가능 기능 모듈)           │
│  - 도메인 로직 캡슐화                             │
│  - 패턴화된 작업 수행                             │
│  - Alfred가 자동 선택/호출                        │
└──────────────┬──────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────┐
│ Layer 6: Tools (Claude Code 기본 도구)           │
│  - Read, Write, Edit (파일 조작)                 │
│  - Bash (명령 실행)                               │
│  - Grep, Glob (검색)                              │
│  - AskUserQuestion (상호작용)                     │
└─────────────────────────────────────────────────┘
```

### 각 레이어의 책임과 경계

#### Layer 1: 사용자 (User)

**책임**:
- 최종 의사결정 (Phase 승인, 세부 작업 확인)
- 요구사항 제공
- 결과 검토 및 피드백

**의사결정 범위**:
- 전체 워크플로우 진행 여부 (커맨드 레벨)
- 위험한 작업 승인 여부 (Sub-agent 레벨)

**인터페이스**:
- 입력: 자연어 명령 (`/alfred:1-plan "새 기능"`)
- 출력: AskUserQuestion 응답 ("진행", "수정", "중단")

---

#### Layer 2: Alfred (SuperAgent)

**책임**:
- **전략적 의사결정**: 어떤 Sub-agent를 호출할지, 어떤 Skills를 사용할지
- **커맨드 해석**: 커맨드 파일을 읽고 워크플로우 실행
- **자동 판단**: Skills 선택, 순차/병렬 실행 결정
- **결과 통합**: 여러 에이전트의 결과를 취합하여 사용자에게 보고

**의사결정 범위**:
- Skills 자동 선택 (CLAUDE.md의 "Alfred 지능형 오케스트레이션" 참조)
- 순차/병렬 실행 판단 (의존성 분석 기반)
- 에이전트 간 작업 조율

**인터페이스**:
- 입력: 커맨드 지침, 사용자 요청
- 출력: Sub-agent 호출 (Task tool), Skills 호출 (Skill tool)

**핵심 원칙**:
- **중앙 조율**: Alfred만이 에이전트 간 작업을 조율 (에이전트 간 직접 호출 금지)
- **자동 판단**: 명시적 지침이 없어도 상황 기반 최적 선택
- **투명성**: 모든 의사결정을 사용자에게 보고

---

#### Layer 3: Commands (워크플로우 템플릿)

**책임**:
- **절차 정의**: Phase 구조 및 단계별 작업 명시
- **의존성 힌트**: Alfred에게 실행 순서 힌트 제공
- **AskUserQuestion 호출**: Phase 전환 시 사용자 승인
- **예시 제공**: 구체적인 코드 예시 및 사용법

**의사결정 범위**:
- 워크플로우 구조 (Phase 1 → Phase 2)
- Sub-agent 호출 순서 제안
- Skills 자동 활성화 조건 힌트

**인터페이스**:
- 입력: 사용자 명령 (slash command)
- 출력: Alfred에게 주는 지침 (Markdown 문서)

**핵심 원칙**:
- **선언적 + 절차적**: What (목표) + How (방법) 모두 명시
- **커맨드 우선순위**: 커맨드 지침은 에이전트 지침보다 상위
- **상세 가이드**: 예시 코드 및 상세 설명 포함

---

#### Layer 4: Sub-agents (전문 에이전트)

**책임**:
- **전술적 의사결정**: 어떻게 구현할지 (구체적 방법)
- **단일 업무 수행**: 자신의 전문 영역만 담당
- **AskUserQuestion 호출**: 세부 작업 확인 (파일 덮어쓰기 등)
- **Skills 활용**: Alfred가 제공한 Skills 사용

**의사결정 범위**:
- 구현 세부사항 (파일 구조, 코드 스타일 등)
- 위험한 작업 확인 (덮어쓰기, 삭제 등)
- 작업 완료 여부 판단

**인터페이스**:
- 입력: Alfred의 Task 호출 (prompt)
- 출력: 작업 결과 보고 (output)

**핵심 원칙**:
- **단일 책임**: 각 에이전트는 자신의 전문 영역만 담당
- **자율성**: 내부 구현은 에이전트가 자율적으로 결정
- **협업 금지**: 에이전트 간 직접 호출 금지 (Alfred 경유)

---

#### Layer 5: Skills (재사용 가능 기능 모듈)

**책임**:
- **도메인 로직 캡슐화**: TAG 스캔, TRUST 검증 등
- **패턴화된 작업 수행**: 반복적이고 규칙 기반 작업
- **빠른 실행**: Haiku 모델 사용 (비용 절감, 속도 향상)

**의사결정 범위**:
- 검증 기준 판단 (TRUST 5원칙 준수 여부 등)
- 결과 심각도 판정 (PASS, WARNING, CRITICAL)

**인터페이스**:
- 입력: Alfred의 Skill 호출
- 출력: 검증 결과 (JSON 또는 Markdown 보고서)

**핵심 원칙**:
- **자동 활성화**: Alfred가 상황 기반 자동 선택
- **경량화**: Agent보다 가볍고 빠름
- **재사용성**: 여러 커맨드/에이전트에서 공통 사용

---

#### Layer 6: Tools (Claude Code 기본 도구)

**책임**:
- 파일 시스템 조작 (Read, Write, Edit)
- 명령 실행 (Bash)
- 검색 (Grep, Glob)
- 사용자 상호작용 (AskUserQuestion)

**핵심 원칙**:
- 모든 레이어가 사용 가능
- 부작용 최소화 (readonly 우선)

---

### 정보 흐름 (Information Flow)

#### 순방향 흐름 (Top-Down)

```
사용자 요청: "/alfred:1-plan 새 기능"
    ↓
Alfred 분석:
    - 커맨드: /alfred:1-plan 파일 읽기
    - 의도: SPEC 작성 요청
    ↓
커맨드 지침 해석:
    - Phase 1: 의도 파악 (AskUserQuestion)
    - Phase 2: spec-builder 호출
    ↓
AskUserQuestion (Layer 1):
    - "어떤 기능을 만드시겠습니까?"
    - 사용자 응답: "인증 기능"
    ↓
Alfred 실행 판단:
    - spec-builder 호출 (Task tool)
    - 입력: "인증 기능 SPEC 작성"
    ↓
spec-builder 실행:
    - .moai/project/ 문서 분석
    - SPEC 초안 작성
    - AskUserQuestion: "파일 덮어쓸까요?"
    ↓
Alfred Skills 자동 선택:
    - 조건: SPEC 파일 생성됨
    - 선택: moai-alfred-spec-metadata
    - 실행: Skill 호출
    ↓
결과 반환
```

#### 역방향 흐름 (Bottom-Up)

```
Skills 실행 결과:
    - moai-alfred-spec-metadata: "7개 필수 필드 모두 존재"
    ↓
Sub-agent 결과 통합:
    - spec-builder: "SPEC 작성 완료 + 검증 통과"
    ↓
Alfred 결과 통합:
    - SPEC 파일 경로: .moai/specs/SPEC-AUTH-001/spec.md
    - 검증 결과: PASS
    - 다음 단계: /alfred:2-run AUTH-001
    ↓
사용자에게 보고:
    - 작업 완료 요약
    - 다음 단계 제안
```

---

### 의사결정 경계 (Decision Boundary)

각 레이어의 의사결정 범위를 명확히 구분하여 책임을 분리합니다.

| 의사결정 유형       | 레이어     | 예시                                         |
| ------------------- | ---------- | -------------------------------------------- |
| **최종 승인**       | User       | "Phase 2 진행할까요?", "파일 덮어쓸까요?"    |
| **전략적 판단**     | Alfred     | "어떤 Skills를 호출할까?", "순차/병렬 실행?" |
| **워크플로우 구조** | Commands   | "Phase 1 → Phase 2 순서", "의존성 힌트"      |
| **전술적 판단**     | Sub-agents | "어떤 파일 구조?", "어떤 코드 스타일?"       |
| **검증 판단**       | Skills     | "TRUST 원칙 준수?", "TAG 체인 무결성?"       |

**핵심 원칙**:
1. **상위 레이어 우선**: 충돌 시 상위 레이어 의사결정 우선
2. **명확한 경계**: 각 레이어는 자신의 범위 내에서만 판단
3. **위임 금지**: 자신의 책임을 다른 레이어에 위임 금지

---

### 실전 예시: /alfred:1-plan 실행 흐름

```
1. User → Alfred
   사용자: "/alfred:1-plan 새 기능"

2. Alfred → Commands
   Alfred: /alfred:1-plan.md 읽기

3. Commands → Alfred
   커맨드 지침: "Phase 1에서 AskUserQuestion으로 의도 파악"

4. Alfred → User (AskUserQuestion)
   Alfred: "어떤 기능을 만드시겠습니까?"
   User: "인증 기능"

5. Alfred → Sub-agent
   Alfred: Task(spec-builder, prompt="인증 기능 SPEC 작성")

6. Sub-agent → User (AskUserQuestion)
   spec-builder: "기존 파일 덮어쓸까요?"
   User: "덮어쓰기"

7. Sub-agent → Tools
   spec-builder: Write(.moai/specs/SPEC-AUTH-001/spec.md)

8. Alfred → Skills (자동 판단)
   Alfred: 조건 체크 → SPEC 생성됨
   Alfred: Skill(moai-alfred-spec-metadata)

9. Skills → Alfred
   moai-alfred-spec-metadata: "검증 PASS"

10. Alfred → User
    Alfred: "작업 완료 보고 + 다음 단계 제안"
```

**각 레이어의 역할**:
- **User**: 승인 (2회)
- **Alfred**: 조율 (커맨드 해석, Sub-agent 호출, Skills 자동 선택)
- **Commands**: 절차 제공 (Phase 구조, AskUserQuestion 타이밍)
- **Sub-agent**: 전문 작업 (SPEC 작성)
- **Skills**: 검증 (메타데이터 확인)
- **Tools**: 실행 (파일 쓰기)

---

## Context Engineering (컨텍스트 엔지니어링)

MoAI-ADK는 Anthropic의 "Effective Context Engineering for AI Agents" 원칙을 기반으로 효율적인 컨텍스트 관리를 구현합니다.

### 1. JIT (Just-in-Time) Retrieval

**원칙**: 필요한 순간에만 문서를 로드하여 초기 컨텍스트 부담을 최소화

**Alfred의 JIT 전략**:

| 커맨드           | 필수 로드        | 선택적 로드               | 로드 타이밍       |
| ---------------- | ---------------- | ------------------------- | ----------------- |
| `/alfred:1-plan` | product.md       | structure.md, tech.md     | SPEC 후보 발굴 시 |
| `/alfred:2-run`  | SPEC-XXX/spec.md | development-guide.md      | TDD 구현 시작 시  |
| `/alfred:3-sync` | sync-report.md   | TAG 체인 검증 (`rg` 스캔) | 문서 동기화 시    |

**구현 방법**:
- Alfred는 커맨드 실행 시점에 필요한 문서만 `Read` 도구로 로드
- 에이전트는 자신의 작업에 필요한 문서만 요청
- CLAUDE.md의 "메모리 전략" 섹션에 명시된 5개 문서는 항상 로드

### Context Engineering 체크리스트

**커맨드 설계 시**:
- [ ] JIT: 필요한 문서만 로드하는가?
- [ ] 선택적 로드: 조건부로 문서를 로드하는가?

**에이전트 설계 시**:
- [ ] 최소 도구: 필요한 도구만 YAML frontmatter에 선언했는가?
- [ ] 명확한 역할: 단일 책임 원칙을 준수하는가?

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
- **Ruby**: RSpec + SPEC 기반 BDD 테스트 (행동 명세 우선)

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
- **Ruby**: SPEC 행동을 반영하는 duck typing + RuboCop 검증

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
  - `/alfred:1-plan`: `@SPEC:ID` 태그로 SPEC 작성 (.moai/specs/)
  - `/alfred:2-run`: `@TEST:ID` (tests/) → `@CODE:ID` (src/) TDD 구현
  - `/alfred:3-sync`: `@DOC:ID` (docs/) 문서 동기화, 전체 TAG 검증
- **코드 스캔 기반 추적성**: 중간 캐시 없이 `rg '@(SPEC|TEST|CODE|DOC):' -n`으로 코드를 직접 스캔하여 TAG 추적성 보장한다.

---

## SPEC 우선 사고방식

1. **SPEC 기반 의사결정**: 모든 기술적 결정은 기존 SPEC을 참조하거나 새로운 SPEC을 만든다. 명확한 요구사항 없이는 구현하지 않는다.
2. **SPEC 맥락 읽기**: 코드 변경 전에 관련 SPEC 문서를 읽고, @TAG 관계를 파악하고, 준수를 검증한다.
3. **SPEC 소통**: 한국어가 기본 소통 언어다. 모든 SPEC 문서는 기술 용어는 영어로, 설명은 명확한 한국어로 작성한다.

## SPEC-TDD 워크플로우

1. **SPEC 우선**: 코드 작성 전에 SPEC을 생성하거나 참조한다. `/alfred:1-plan`을 사용하여 요구사항, 설계, 작업을 명확히 정의한다.
2. **TDD 구현**: Red-Green-Refactor를 엄격히 따른다. 언어별 적절한 테스트 프레임워크와 함께 `/alfred:2-run`를 사용한다.
3. **추적성 동기화**: `/alfred:3-sync`를 실행하여 문서를 업데이트하고 SPEC과 코드 간 @TAG 관계를 유지한다.

## @TAG 시스템

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

> **📋 SPEC 메타데이터 표준 (SSOT)**: `spec-metadata.md`

**모든 SPEC 문서는 YAML Front Matter + HISTORY 섹션을 포함**해야 합니다:
- **필수 필드 7개**: id, version, status, created, updated, author, priority
- **선택 필드 9개**: category, labels, depends_on, blocks, related_specs, related_issue, scope
- **HISTORY 섹션**: 모든 버전 변경 이력 기록 (필수)

**전체 템플릿, 필드 상세 설명, 검증 방법**: `spec-metadata.md` 참조

**간단한 참조 예시**:
```yaml
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-09-15
updated: 2025-09-15
author: @Goos
priority: high
---

# @SPEC:AUTH-001: JWT 인증 시스템

## HISTORY
### v0.0.1 (2025-09-15)
- **INITIAL**: JWT 기반 인증 시스템 명세 작성
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

- **TAG ID**: `<도메인>-<3자리>` (예: `AUTH-003`) - **영구 불변**
- **디렉토리 명명 규칙**: `.moai/specs/SPEC-{ID}/` (필수)
  - ✅ **올바른 예**: `SPEC-AUTH-001/`, `SPEC-REFACTOR-001/`, `SPEC-UPDATE-REFACTOR-001/`
  - ❌ **잘못된 예**: `AUTH-001/`, `SPEC-001-auth/`, `SPEC-AUTH-001-jwt/`
  - **복합 도메인**: 하이픈으로 연결 가능 (예: `UPDATE-REFACTOR-001`)
  - **경고**: 하이픈 3개 이상 연결 시 단순화 권장
- **TAG 내용**: 자유롭게 수정 가능 (HISTORY에 기록 필수)
- **버전 관리**: Semantic Versioning (v0.0.1 → v0.1.0 → v1.0.0)
  - 상세 버전 체계: `spec-metadata.md#버전-체계` 참조
- **새 TAG 생성 전 중복 확인**: `rg "@SPEC:{ID}" -n .moai/specs/` (필수)
- **TAG 검증**: `rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/`
- **SPEC 버전 일치성 확인**: `rg "SPEC-{ID}.md v" -n`
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

- **Python**: pytest (테스트), mypy (타입 검사), black (포맷)
- **TypeScript**: Vitest (테스트), Biome (린터+포맷)
- **Java**: JUnit (테스트), Maven/Gradle (빌드)
- **Go**: go test (테스트), gofmt (포맷)
- **Rust**: cargo test (테스트), rustfmt (포맷)
- **Ruby**: RSpec (테스트), RuboCop (린터+포맷), Bundler (패키지 관리)

## 변수 역할 참고

| Role               | Description                        | Example                              |
| ------------------ | ---------------------------------- | ------------------------------------ |
| Fixed Value        | Constant after initialization      | `const MAX_SIZE = 100`               |
| Stepper            | Changes sequentially               | `for (let i = 0; i < n; i++)`        |
| Flag               | Boolean state indicator            | `let isValid = true`                 |
| Walker             | Traverses a data structure         | `while (node) { node = node.next; }` |
| Most Recent Holder | Holds the most recent value        | `let lastError`                      |
| Most Wanted Holder | Holds optimal/maximum value        | `let bestScore = -Infinity`          |
| Gatherer           | Accumulator                        | `sum += value`                       |
| Container          | Stores multiple values             | `const list = []`                    |
| Follower           | Previous value of another variable | `prev = curr; curr = next;`          |
| Organizer          | Reorganizes data                   | `const sorted = array.sort()`        |
| Temporary          | Temporary storage                  | `const temp = a; a = b; b = temp;`   |

---

이 가이드는 MoAI-ADK 3단계 파이프라인을 실행하는 표준을 제공한다.

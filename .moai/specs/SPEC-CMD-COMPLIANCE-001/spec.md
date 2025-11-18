# SPEC-CMD-COMPLIANCE-001: Zero Direct Tool Usage 준수

**프로젝트**: MoAI-ADK v0.26.0
**작성일**: 2025-11-19
**상태**: Draft
**우선순위**: High
**범위**: `.claude/commands/moai/` 명령어 아키텍처

---

## 📋 개요

MoAI-ADK의 Claude Code 커맨드들이 "Zero Direct Tool Usage" 원칙을 준수하도록 아키텍처를 정리하는 프로젝트입니다. 현재 6개 커맨드 중 3개가 원칙 위반 상태이며, 이를 에이전트 위임 패턴으로 전환합니다.

**주요 목표**:
- ✅ 프로덕션 커맨드 100% 에이전트 위임 준수
- 📦 패키지 템플릿과 로컬 프로젝트 SSOT 유지
- 📚 명확한 예외 패턴 문서화
- 🔍 자동 검증 파이프라인 구축

---

## 🎯 Zero Direct Tool Usage 원칙

### 핵심 규칙

| 규칙 | 설명 |
|------|------|
| **✅ 허용** | `Task()`, `AskUserQuestion()`, `Skill()` |
| **❌ 금지** | `Read()`, `Write()`, `Edit()`, `Bash()`, `Grep()`, `Glob()`, `TodoWrite()` |
| **🎯 목표** | 모든 작업을 전문 에이전트에게 위임 |

### 이점

```
직접 도구 사용 (Before)          에이전트 위임 (After)
├─ 컨텍스트 분산                ├─ 명확한 책임 분리
├─ 토큰 낭비                    ├─ 토큰 80-85% 절약
├─ 유지보수 어려움              ├─ 일관된 패턴
└─ 테스트 복잡함                └─ 재사용 가능한 에이전트
```

---

## 🔍 현재 상태 분석

### 명령어 준수 현황

| 커맨드 | 상태 | 위반 도구 수 | 조치 |
|--------|------|------------|------|
| `/moai:0-project` | ✅ 준수 | 0 | 없음 |
| `/moai:1-plan` | ❌ 위반 | 9 | **필수 수정** |
| `/moai:2-run` | ✅ 준수 | 0 | 없음 |
| `/moai:3-sync` | ❌ 위반 | 10 | **필수 수정** |
| `/moai:9-feedback` | ✅ 예외 | - | 도구 특화 (허용) |
| `/moai:99-release` | ⚠️ 예외 | - | **예외 문서화** |

### 위반 상세 분석

**1. `/moai:1-plan` (CRITICAL)**
```yaml
allowed-tools:
  - Read           # ❌ 위반
  - Write          # ❌ 위반
  - Edit           # ❌ 위반
  - MultiEdit      # ❌ 위반
  - Grep           # ❌ 위반
  - Glob           # ❌ 위반
  - TodoWrite      # ❌ 위반
  - Bash(git:*)    # ❌ 위반
  - Bash(gh:*)     # ❌ 위반
  - AskUserQuestion # ✅ 허용
  - Skill          # ✅ 허용
```

**2. `/moai:3-sync` (CRITICAL)**
```yaml
allowed-tools:
  - Read           # ❌ 위반
  - Write          # ❌ 위반
  - Edit           # ❌ 위반
  - MultiEdit      # ❌ 위반
  - Bash(git:*)    # ❌ 위반
  - Bash(gh:*)     # ❌ 위반
  - Bash(uv:*)     # ❌ 위반
  - Grep           # ❌ 위반
  - Glob           # ❌ 위반
  - TodoWrite      # ❌ 위반
  - Task           # ✅ 허용
```

**3. `/moai:99-release` (로컬 전용)**
- 패키지와 함께 배포되지 않음
- 메인테이너 전용 (GoosLab only)
- 직접 접근 필요 (PyPI 릴리스)
- → **예외 패턴으로 문서화 필요**

---

## 📐 EARS 요구사항

### 1️⃣ Ubiquitous (항상 참)

**Statement**: 모든 Claude Code 커맨드는 `Task()`, `AskUserQuestion()`, `Skill()` 만으로 구성되어야 한다.

**Rationale**: 에이전트 위임을 통해 아키텍처 명확성, 토큰 효율성, 유지보수성 확보

**Acceptance**:
```yaml
# ✅ CORRECT
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill

# ❌ WRONG
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
```

---

### 2️⃣ Event-Driven (이벤트 발생)

**Statement**: WHEN 커맨드가 실행되면, THEN 모든 작업을 전문 에이전트에게 위임해야 한다.

**Examples**:

| 이벤트 | 위임 대상 | 방식 |
|--------|----------|------|
| 사용자가 `/moai:1-plan` 실행 | `spec-builder` | `Task(subagent_type="spec-builder", prompt="...")` |
| 코드 탐색 필요 | `Explore` | `Task(subagent_type="Explore", prompt="...")` |
| Git 작업 필요 | `git-manager` | `Task(subagent_type="git-manager", prompt="...")` |
| 문서 동기화 | `doc-syncer` | `Task(subagent_type="doc-syncer", prompt="...")` |
| 품질 검증 | `quality-gate` | `Task(subagent_type="quality-gate", prompt="...")` |

**Acceptance**:
```
Given: 파일 탐색이 필요한 커맨드
When: 커맨드 실행
Then: Explore 에이전트에게 Task() 위임
```

---

### 3️⃣ Unwanted (방지해야 할 조건)

**Statement**: IF 커맨드가 직접 도구를 사용하면, THEN 자동 검증에서 실패해야 한다.

**Detection Rules**:
```python
# ❌ 감지되면 안 되는 패턴
FORBIDDEN_PATTERNS = [
    r'^allowed-tools:.*Read',
    r'^allowed-tools:.*Write',
    r'^allowed-tools:.*Edit',
    r'^allowed-tools:.*Bash',
    r'^allowed-tools:.*Grep',
    r'^allowed-tools:.*Glob',
    r'^allowed-tools:.*TodoWrite'
]
```

**Acceptance**:
```
Given: 커맨드가 직접 도구를 사용
When: 자동 검증 실행
Then: FAIL 상태로 전환 및 에러 메시지 표시
```

---

### 4️⃣ State-Driven (상태 기반)

**Statement**: WHILE 프로덕션 환경에서 커맨드가 실행되는 동안, THEN 엄격한 에이전트 위임 규칙이 적용되어야 한다.

**State Transitions**:
```
프로덕션 커맨드 로드
    ↓
allowed-tools 검증
    ↓
(위반 발견) ← 차단 및 에러
(준수) ↓
에이전트 위임 실행
    ↓
작업 완료
```

**Scope**:
- 프로덕션 커맨드: `/moai:0-project`, `/moai:1-plan`, `/moai:2-run`, `/moai:3-sync`
- 로컬 커맨드 (예외): `/moai:9-feedback`, `/moai:99-release`

---

### 5️⃣ Optional (선택적)

**Statement**: WHERE 로컬 전용 도구인 경우, THEN 예외 근거를 명확하게 문서화할 수 있다.

**Exception Criteria**:
1. 패키지와 함께 배포되지 않음
2. 메인테이너 또는 특정 역할 전용
3. 직접 시스템 접근이 필수
4. 일반 사용자에게 영향 없음

**Exception Pattern Template**:
```markdown
---
⚠️ **EXCEPTION: Local-Only Development Tool**

이 커맨드는 Zero Direct Tool Usage 원칙의 예외입니다:

**예외 사유**:
1. 로컬 개발 전용 (패키지 배포 안됨)
2. [특정 역할] 전용 사용
3. [직접 접근이 필요한 이유]
4. [일반 사용자 영향 없음]

**프로덕션 커맨드**: 패키지와 함께 배포되는 커맨드는
엄격한 에이전트 위임 원칙 준수 필수
---
```

---

## ✅ 수용 기준

### 기준 1: allowed-tools 준수

```yaml
# ✅ PASS
allowed-tools:
  - Task
  - AskUserQuestion

# ❌ FAIL
allowed-tools:
  - Task
  - Read              ← 위반!
  - AskUserQuestion
```

---

### 기준 2: 에이전트 위임 패턴

모든 지시사항이 다음 중 하나를 사용해야 함:

```python
# Pattern 1: Task 위임
Task(
    subagent_type="agent-name",
    description="작업 설명",
    prompt="상세 지시사항"
)

# Pattern 2: 사용자 상호작용
AskUserQuestion(
    questions=[
        {
            "question": "질문?",
            "header": "헤더",
            "multiSelect": false,
            "options": [...]
        }
    ]
)

# Pattern 3: 특정 스킬 호출
Skill("skill-name")
```

---

### 기준 3: 패키지 템플릿 동기화

```
로컬 프로젝트 파일:     .claude/commands/moai/1-plan.md
패키지 템플릿 파일:     src/moai_adk/templates/.claude/commands/moai/1-plan.md

검증: diff 결과 동일해야 함
```

---

### 기준 4: 예외 문서화

로컬 전용 커맨드는 상단에 다음 포함:

```markdown
---
⚠️ **EXCEPTION: Local-Only Development Tool**
[예외 사유 4가지]
---
```

---

## 📊 성공 지표

| 지표 | 목표 | 측정 방법 |
|------|------|----------|
| **프로덕션 준수율** | 100% (4/6) | allowed-tools 검증 |
| **예외 문서화율** | 100% (2/6) | 문서 완료 확인 |
| **템플릿 동기화율** | 100% | diff 비교 |
| **자동 검증** | PASS | CI/CD 파이프라인 |
| **테스트 커버리지** | 90%+ | pytest 실행 |

---

## 🔗 관련 문서

- **CLAUDE.md**: Zero Direct Tool Usage 원칙 개요
- **git-workflow-detailed.md**: Git 워크플로우
- **agent-delegation.md**: 에이전트 위임 패턴
- **token-efficiency.md**: 토큰 효율성 전략

---

**생성일**: 2025-11-19
**상태**: Draft → Ready for TDD Implementation
**다음 단계**: `/moai:2-run SPEC-CMD-COMPLIANCE-001` 실행
# Alfred Command Best Practices - MoAI-ADK 참조 패턴

> **문서 목적**: `/alfred:2-run`의 우수한 에이전트 위임 패턴을 참조 구현으로 문서화하고, 다른 Alfred 명령 개발의 기준 제시

## 📋 개요

MoAI-ADK의 모든 Alfred 명령은 **에이전트 중심 아키텍처**를 따릅니다:

```
Command (Orchestration)
    ↓ Read context, Ask user
    ↓ Task(subagent) delegation
Agent (Execution)
    ↓ Uses tools directly
Skill (Knowledge)
    ↓ Reusable domain knowledge
```

---

## ✅ 참조 구현: `/alfred:2-run`

### 개요

`/alfred:2-run SPEC-XXX` 명령은 TDD 방식의 완전한 구현 사이클을 관리합니다:

1. SPEC 분석 및 계획 수립 (implementation-planner)
2. TDD 기반 코드 개발 (tdd-implementer)
3. 품질 검증 (quality-gate)
4. 버전 관리 (git-manager)

### 구조 분석

#### Phase 1: 분석 및 계획 (Line 80-150)

```markdown
Step 1.1: SPEC 문서 읽기 (직접 도구 사용 - 정당함)
    Tool: Read
    Reason: Alfred의 컨텍스트 준비

Step 1.2: 구현 계획 에이전트 위임 (Task 사용 - 정당함)
    Task(subagent_type="implementation-planner")
    Reason: 복잡한 분석 및 전략 수립 필요

Step 1.3: 사용자 승인 (AskUserQuestion - 정당함)
    Reason: 중요한 결정에 대한 사용자 의견 수집
```

**핵심**: 컨텍스트 준비는 Alfred가 직접, 복잡한 분석은 에이전트 위임

#### Phase 2: 작업 실행 (Line 151-380)

```markdown
Step 2.1: 진행 상황 추적 (TodoWrite 초기화 - Alfred 오케스트레이션)
    Reason: 프로젝트 관리는 Alfred의 책임

Step 2.2: 코드 탐색 (선택적 Explore 에이전트)
    Task(subagent_type="Explore")
    Reason: 복잡한 코드 탐색 자동화

Step 2.3: TDD 구현 (tdd-implementer 위임)
    Task(subagent_type="tdd-implementer")
    Reason: 복잡한 개발 로직은 전문 에이전트에게

Step 2.4: 품질 검증 (quality-gate 위임)
    Task(subagent_type="quality-gate")
    Reason: TRUST 5 원칙 검증은 전문 에이전트에게
```

**핵심**: 모든 복잡한 작업을 특화된 에이전트에게 위임

#### Phase 3: Git 작업 (Line 381-420)

```markdown
Step 3.1: 커밋 및 푸시 (git-manager 위임)
    Task(subagent_type="git-manager")
    Reason: Git 작업의 복잡성 (충돌, 브랜치 등)

Step 3.2: 단순 검증 (직접 Bash - 정당함)
    Bash: git log -1 --oneline
    Reason: 결과 확인용 단순 명령 (<10자)
```

**핵심**: 복잡한 Git 작업은 위임, 검증만 직접 수행

#### Phase 4: 다음 단계 (Line 421-450)

```markdown
Step 4.1: 사용자 선택
    AskUserQuestion
    Reason: 다음 작업 방향을 사용자와 협의
```

---

## 🎯 우수한 점 분석

### 1. 100% 에이전트 위임 (복잡한 작업)

```
✅ 모든 복잡한 작업이 에이전트에게 할당됨:
  - 분석 → implementation-planner
  - 개발 → tdd-implementer
  - 검증 → quality-gate
  - Git → git-manager
```

**이점**:
- 각 에이전트가 특화된 Skills를 로드
- 책임 분리가 명확
- 테스트가 독립적으로 가능

### 2. 정당한 직접 도구 사용

```
✅ 직접 도구 사용 기준 (모두 충족):
  - Read: SPEC 문서 읽기 (컨텍스트 준비)
  - Bash: git log 확인 (단순 검증, <10자)
  - Hook: spec_status_hooks.py (재사용 가능, <100ms)
```

**기준**:
- 컨텍스트 준비용 Read
- 단순 검증용 Bash (<10자)
- 경량 인프라 훅 (<100ms)

### 3. 스크립트 생성 제로

```
✅ 모든 스크립트는 재사용 가능한 인프라:
  - .claude/hooks/alfred/spec_status_hooks.py (기존)
  - 임시 스크립트 생성 없음
```

**이점**:
- 임시 파일 정리 불필요
- 재사용 가능한 도구 구축
- 코드 중복 제거

### 4. 에이전트 Skill 활용

```
✅ 각 에이전트는 특화된 Skills 로드:
  - tdd-implementer: moai-lang-python, moai-essentials-debug
  - quality-gate: moai-foundation-tags, moai-trust-5-principles
  - git-manager: moai-domain-git (또는 Bash 직접 사용)
```

**이점**:
- 지식 중앙화
- 중복 제거
- 유지보수 단순화

### 5. 명확한 책임 분리

```
Commands (Orchestration)
    ↓ Read context (컨텍스트 준비)
    ↓ Task() delegation (작업 위임)
    ↓ AskUserQuestion (사용자 의견)

Agents (Execution)
    ↓ Read/Write/Edit (파일 조작)
    ↓ Bash (명령 실행)
    ↓ Load Skills (지식 활용)

Skills (Knowledge)
    ↓ Best practices (모범 사례)
    ↓ Patterns (패턴)
    ↓ Checklists (체크리스트)
```

---

## 📐 일반화된 Alfred 명령 템플릿

모든 Alfred 명령은 다음 구조를 따라야 합니다:

### 기본 구조

```markdown
# /alfred:X-name 명령

## Phase 1: 컨텍스트 및 계획 (10-20% 토큰)
1. Read: 필요한 문서/설정 읽기 (Alfred가 직접)
2. Analyze: 복잡한 분석 필요 시 Task 위임
3. Plan: AskUserQuestion으로 사용자 승인
4. Initialize: TodoWrite로 진행 상황 추적

## Phase 2: 작업 실행 (60-80% 토큰)
1. Task(subagent_type="specialist-agent")
   └─ 에이전트가 도구 직접 사용
   └─ Skills 자동 로드
   └─ 진행 상황 자체 업데이트

## Phase 3: 최종화 (10-20% 토큰)
1. Git 작업: Task(subagent_type="git-manager")
2. 검증: Bash로 단순 확인
3. 사용자 선택: AskUserQuestion으로 다음 단계

## 금지 사항
❌ Alfred가 복잡한 작업 직접 실행
❌ 임시 스크립트 생성
❌ 에이전트 책임 분산
```

---

## 🔄 다른 명령 적용 사례

### `/alfred:1-plan` (SPEC 작성)

```markdown
Phase 1: SPEC 요구사항 읽기 (Read)
Phase 2: spec-builder 에이전트에게 SPEC 생성 위임 (Task)
Phase 3: 사용자 승인 및 브랜치 생성 (git-manager 위임)
```

**Skill 활용**:
- spec-builder → moai-alfred-spec-authoring
- git-manager → moai-domain-git 또는 Bash

### `/alfred:3-sync` (문서 동기화)

```markdown
Phase 1: 변경 사항 분석 (Read git diff)
Phase 2: doc-syncer 에이전트에게 문서화 위임 (Task)
Phase 3: PR 생성 및 QA 검증 (git-manager, quality-gate 위임)
```

**Skill 활용**:
- doc-syncer → moai-docs-generation, moai-docs-validation
- quality-gate → moai-trust-5-principles

### `/alfred:0-project` (프로젝트 초기화)

```markdown
Phase 1: 프로젝트 설정 수집 (AskUserQuestion)
Phase 2: project-manager 에이전트에게 프로젝트 생성 위임 (Task)
Phase 3: 초기 파일 생성 및 커밋 (git-manager 위임)
```

**Skill 활용**:
- project-manager → moai-project-config-manager
- git-manager → moai-domain-git

---

## ✅ 검증 체크리스트

### Alfred 명령 작성 체크리스트

새로운 Alfred 명령을 작성할 때 다음을 확인하세요:

- [ ] **Phase 1: 컨텍스트 준비**
  - [ ] Read 또는 Glob로 필요한 문서 읽기
  - [ ] 복잡한 분석은 Task로 위임
  - [ ] AskUserQuestion으로 사용자 승인

- [ ] **Phase 2: 작업 실행**
  - [ ] 복잡한 작업은 모두 Task 위임
  - [ ] 에이전트가 적절한 Skills 로드
  - [ ] 임시 스크립트 생성 없음

- [ ] **Phase 3: 최종화**
  - [ ] Git 작업 → git-manager 위임
  - [ ] 단순 검증만 직접 Bash 사용
  - [ ] 사용자에게 다음 단계 제시

- [ ] **코드 품질**
  - [ ] Skill 재사용 극대화
  - [ ] 에이전트 책임 분리 명확
  - [ ] 에이전트 매핑 문서화

### 명령 리뷰 기준

```
코드 라인 분석:
- Read/Glob: 5-10% (컨텍스트 준비)
- Task(): 70-80% (에이전트 위임)
- Bash: 5-10% (단순 검증)
- AskUserQuestion: 5-10% (사용자 상호작용)
```

---

## 📚 추가 참고 자료

### CLAUDE.md 참조
- "Commands → Agents → Skills Architecture" (L254)
- "Skill Reuse Pattern" (L594)

### 에이전트 정의
- `.claude/agents/tdd-implementer.md`
- `.claude/agents/quality-gate.md`
- `.claude/agents/git-manager.md`
- `.claude/agents/implementation-planner.md`

### Skills 활용
- 55+ Skills in `.claude/skills/`
- Skill("skill-name") 로드 패턴

---

## 🎯 결론

`/alfred:2-run`은 MoAI-ADK의 **모범 구현 사례**입니다.

**핵심 원칙**:
1. ✅ 컨텍스트 준비는 Alfred가 직접
2. ✅ 복잡한 작업은 에이전트에게 위임
3. ✅ 각 에이전트는 특화된 Skills 로드
4. ✅ 책임 분리가 명확
5. ✅ 스크립트는 재사용 가능한 인프라만

이 패턴을 따르면:
- 코드가 유지보수하기 쉬움
- Skill 중복 제거
- 테스트가 독립적
- 확장성이 우수함
- 에러 추적이 명확함

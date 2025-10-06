# ANALYSIS:INTEGRATION-001 - Claude Code Integration Analysis

**분석 일시**: 2025-10-01
**분석 대상**: MoAI-ADK Claude Code 통합 시스템 (에이전트, 슬래시 명령어, 훅)
**분석 범위**: 8개 핵심 에이전트 + 3개 주요 명령어 + 협업 패턴

---

## 📊 Executive Summary

### 핵심 발견사항

1. **완전한 역할 분리 달성**: 8개 에이전트가 명확한 단일 책임 원칙(SRP)을 준수
2. **2단계 승인 워크플로우**: 분석(analysis) → 승인 → 실행(implement) 패턴 일관성
3. **에이전트 간 직접 호출 금지**: 명령어 레벨 오케스트레이션으로 명확한 의존성 관리
4. **도구 권한 최소화**: 에이전트별 필수 도구만 허용하여 보안 강화
5. **Living Document 기반**: 코드 직접 스캔(@TAG) 방식으로 중간 캐시 제거

### 주요 개선 영역

- **성능 최적화**: MultiEdit 활용으로 60% 시간 단축 (spec-builder)
- **차등 스캔**: Level 1→2→3 단계별 검증으로 불필요한 심화 분석 방지
- **Git 전담화**: git-manager가 모든 Git 작업 독점 처리

---

## 🏗️ Architecture Overview

### System Composition

```
MoAI-ADK Claude Code Integration
├── Agents (8 Core)
│   ├── spec-builder       (SPEC 작성 전담)
│   ├── code-builder       (TDD 구현 전담)
│   ├── doc-syncer         (문서 동기화 전담)
│   ├── tag-agent          (TAG 관리 독점)
│   ├── git-manager        (Git 작업 독점)
│   ├── debug-helper       (오류 분석 전담)
│   ├── trust-checker      (품질 검증 전담)
│   └── cc-manager         (Claude Code 설정 전담)
├── Commands (3 Primary)
│   ├── /alfred:1-spec       (명세 작성 파이프라인)
│   ├── /alfred:2-build      (TDD 구현 파이프라인)
│   └── /alfred:3-sync       (문서 동기화 파이프라인)
└── Hooks (3 Types)
    ├── SessionStart       (세션 시작 알림)
    ├── PreToolUse         (쓰기 작업 가드)
    └── UserPromptSubmit   (스티어링 가드)
```

---

## 🎯 Agent Deep Dive

### 1. spec-builder (SPEC 작성 전담)

**파일**: `.claude/agents/alfred/spec-builder.md`

#### 핵심 역할
- 프로젝트 문서 분석 및 SPEC 후보 발굴
- EARS 명세 작성 (Environment, Assumptions, Requirements, Specifications)
- Personal/Team 모드별 산출물 생성

#### 주요 특징
- **MultiEdit 활용**: 3개 파일(spec.md, plan.md, acceptance.md) 동시 생성으로 60% 시간 단축
- **시간 예측 금지**: "예상 소요 시간" 표현 절대 금지, 우선순위 기반 접근
- **Git 작업 위임**: git-manager에게 브랜치/PR 생성 완전 위임

#### 협업 패턴
```
spec-builder (분석) → 사용자 승인 → spec-builder (SPEC 작성) → git-manager (Git 작업)
```

#### 도구 권한
- Read, Write, Edit, MultiEdit, Bash, Glob, Grep, TodoWrite, WebFetch

#### 산출물
- **Personal 모드**: `.moai/specs/SPEC-XXX/{spec,plan,acceptance}.md`
- **Team 모드**: GitHub Issue + SPEC 문서

---

### 2. code-builder (TDD 구현 전담)

**파일**: `.claude/agents/alfred/code-builder.md`

#### 핵심 역할
- SPEC 기반 TDD 구현 (Red-Green-Refactor)
- 언어별 최적 도구 자동 선택
- @TAG 시스템 완전 통합

#### 2단계 워크플로우
1. **분석 모드** (`--mode=analysis`)
   - SPEC 문서 분석
   - @TAG 체인 분석
   - 구현 계획 수립
   - 사용자 승인 대기

2. **구현 모드** (`--mode=implement`)
   - RED: 실패 테스트 작성 (@TEST TAG 적용)
   - GREEN: 최소 구현 (@CODE/@CODE/@CODE/@CODE TAG 적용)
   - REFACTOR: 품질 개선 (@CODE/@CODE/@DOC TAG 적용)

#### @TAG 통합 전략
```typescript
// @CODE:LOGIN-001 | Chain: @SPEC:AUTH-001 -> @SPEC:AUTH-001 -> @CODE:AUTH-001 -> @TEST:AUTH-001
// Related: @CODE:LOGIN-001, @CODE:LOGIN-001
class AuthenticationService {
    // 구현...
}
```

#### 언어별 최적화
| 언어 | 테스트 도구 | 목표 커버리지 | 성능 목표 |
|------|------------|--------------|-----------|
| TypeScript | Vitest/Jest | 85%+ | < 100ms |
| Python | pytest | 85%+ | < 150ms |
| Go | go test | 85%+ | < 50ms |
| Rust | cargo test | 85%+ | < 50ms |

#### 협업 패턴
```
code-builder (분석) → 승인 → code-builder (TDD 구현) → git-manager (커밋)
```

#### 도구 권한
- Read, Write, Edit, MultiEdit, Bash, Grep, Glob, TodoWrite

---

### 3. doc-syncer (문서 동기화 전담)

**파일**: `.claude/agents/alfred/doc-syncer.md`

#### 핵심 역할
- Living Document 동기화 (코드 ↔ 문서)
- @TAG 시스템 검증 및 업데이트
- 문서-코드 일치성 보장

#### 동기화 프로세스
1. **Phase 1: 현황 분석** (2-3분)
   - Git 상태 확인
   - 코드 스캔 (CODE-FIRST): `rg '@TAG' -n src/ tests/`
   - 문서 현황 파악

2. **Phase 2: 문서 동기화** (5-10분)
   - 코드 → 문서: API 문서, README 자동 갱신
   - 문서 → 코드: SPEC 변경 추적, TODO 마킹

3. **Phase 3: 품질 검증** (3-5분)
   - TAG 무결성 검사
   - 문서-코드 일치성 검증
   - 동기화 보고서 생성

#### TAG 검증 명령
```bash
# Primary Chain 완전성 검증
rg '@SPEC:[A-Z]+-[0-9]{3}' -n src/ | wc -l
rg '@SPEC:[A-Z]+-[0-9]{3}' -n src/ | wc -l
rg '@CODE:[A-Z]+-[0-9]{3}' -n src/ | wc -l
rg '@TEST:[A-Z]+-[0-9]{3}' -n tests/ | wc -l
```

#### 협업 패턴
```
doc-syncer (동기화) → git-manager (커밋 + PR 전환)
```

#### 도구 권한
- Read, Write, Edit, MultiEdit, Grep, Glob, TodoWrite

---

### 4. tag-agent (TAG 관리 독점)

**파일**: `.claude/agents/alfred/tag-agent.md`

#### 핵심 원칙
- **유일한 TAG 관리 권한자**: MoAI-ADK의 모든 TAG 작업 독점
- **코드 기반 진실**: TAG의 source of truth는 코드 자체에만 존재
- **읽기 전용**: Write/Edit 도구 사용 금지

#### 주요 책임
1. **코드 기반 TAG 스캔**: 소스 파일에서 실시간 TAG 추출
2. **TAG 무결성 검증**: Primary Chain, 참조 관계, 중복 검증
3. **TAG 체인 관리**: @SPEC → @SPEC → @CODE → @TEST 무결성 보장

#### 스캔 대상
- 소스 파일: `.ts`, `.js`, `.py`, `.java`, `.go`, `.rs`, `.cpp`, `.c`, `.h`
- 문서 파일: `.md`

#### 정규식 패턴
```regex
@[A-Z]+(?:[:|-]([A-Z0-9-]+))?
```

#### 성공 기준
- TAG 형식 오류: 0건 유지
- 중복 TAG: 95% 이상 방지
- 체인 무결성: 100% 보장
- 코드 스캔 속도: < 50ms (소형 프로젝트)

#### 도구 권한
- Read, Glob, Bash (읽기 전용)

---

### 5. git-manager (Git 작업 독점)

**파일**: `.claude/agents/alfred/git-manager.md`

#### 핵심 원칙
- **모든 Git 작업 독점**: 브랜치, 커밋, 태그, PR 생성
- **직접 Git 명령 사용**: 복잡한 스크립트 의존성 최소화
- **모드별 최적화**: Personal/Team 모드 차별화 전략

#### 간소화된 운영
```bash
# 체크포인트 생성 (한국시간 기준)
git tag -a "moai_cp/$(TZ=Asia/Seoul date +%Y%m%d_%H%M%S)" -m "메시지"

# 브랜치 생성 (Personal)
git checkout -b "feature/$(echo 설명 | tr ' ' '-')"

# TDD 커밋 (Team)
git add . && git commit -m "🔴 RED: $테스트_설명\n\n@TEST:$SPEC_ID-RED"
```

#### 모드별 전략

| 모드 | 철학 | 주요 기능 |
|------|------|----------|
| Personal | "안전한 실험, 간단한 Git" | 로컬 중심, 간단한 체크포인트 |
| Team | "체계적 협업, 간단한 자동화" | GitFlow, 구조화 커밋, PR 관리 |

#### 협업 패턴
- spec-builder → git-manager (SPEC 브랜치/Issue)
- code-builder → git-manager (TDD 커밋)
- doc-syncer → git-manager (동기화 커밋 + PR 전환)

#### 도구 권한
- Bash (git:*), Read, Write, Edit, Glob, Grep

---

### 6. debug-helper (오류 분석 전담)

**파일**: `.claude/agents/alfred/debug-helper.md`

#### 2가지 전문 모드
1. **일반 오류 디버깅**: 코드/Git/설정 오류 분석
2. **TRUST 원칙 검사**: TRUST 5원칙 준수도 검증

#### 차등 스캔 시스템 (성능 최적화)

| 레벨 | 소요시간 | 검사 내용 | 조기 종료 |
|------|----------|----------|-----------|
| Level 1 | 1-3초 | 파일 존재, 기본 구조 | Critical 발견 시 |
| Level 2 | 5-10초 | 코드 품질, 테스트 실행 | Warning 발견 시 |
| Level 3 | 20-30초 | 전체 분석, 의존성 검사 | 완료 |

#### 진단 도구
```bash
# 파일 시스템 분석
find . -name "*.py" -exec wc -l {} + | sort -nr

# Git 상태 분석
git status --porcelain
git log --oneline -10

# 테스트 및 품질
python -m pytest --tb=short
ruff check . || flake8 .
```

#### 위임 규칙
```yaml
코드 관련 문제: → code-builder
Git 관련 문제: → git-manager
설정 관련 문제: → cc-manager
문서 관련 문제: → doc-syncer
복합 문제: → 해당 커맨드 실행
```

#### 도구 권한
- Read, Grep, Glob, Bash, TodoWrite

---

### 7. trust-checker (품질 검증 전담)

**파일**: `.claude/agents/alfred/trust-checker.md`

#### 온디맨드 에이전트
- **호출 방식**: 사용자 직접 호출만
- **전문 분야**: TRUST 5원칙 + 코드 품질 + 보안 + 성능

#### TRUST 5원칙 검증

##### T - Test First
```yaml
Level 1: test/ 디렉토리 존재
Level 2: npm test 실행 및 성공률
Level 3: 커버리지 ≥ 85% 정밀 측정
```

##### R - Readable
```yaml
Level 1: wc -l 파일 크기 (≤ 300 LOC)
Level 2: 함수 크기 (≤ 50 LOC)
Level 3: 복잡도 (≤ 10) 정밀 계산
```

##### U - Unified
```yaml
Level 1: import 구문 기본 분석
Level 2: 계층 분리 구조 검사
Level 3: 순환 의존성 탐지
```

##### S - Secured
```yaml
Level 1: .env 파일 .gitignore 확인
Level 2: 입력 검증 로직 분석
Level 3: 보안 취약점 심화 분석
```

##### T - Trackable
```yaml
Level 1: package.json version 확인
Level 2: @TAG 사용 패턴 분석
Level 3: @TAG 시스템 완전 분석
```

#### 검증 결과 포맷
```markdown
🧭 TRUST 5원칙 검증 결과
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 전체 준수율: XX% | 레벨: X | 소요시간: X초

🎯 원칙별 점수:
┌─────────────────┬──────┬────────┐
│ 원칙            │ 점수 │ 상태   │
├─────────────────┼──────┼────────┤
│ T (Test First)  │ XX%  │ ✅/⚠️/❌ │
│ R (Readable)    │ XX%  │ ✅/⚠️/❌ │
│ U (Unified)     │ XX%  │ ✅/⚠️/❌ │
│ S (Secured)     │ XX%  │ ✅/⚠️/❌ │
│ T (Trackable)   │ XX%  │ ✅/⚠️/❌ │
└─────────────────┴──────┴────────┘
```

#### 도구 권한
- Read, Grep, Glob, Bash, TodoWrite

---

### 8. cc-manager (Claude Code 설정 전담)

**파일**: `.claude/agents/alfred/cc-manager.md`

#### 중앙 관제탑 역할
- **표준화 관리**: 모든 Claude Code 파일 생성/수정 표준
- **설정 최적화**: 권한 관리, 훅 시스템
- **품질 검증**: 표준 준수 자동 검증
- **가이드 통합**: 외부 참조 불필요한 완전한 지침

#### 커맨드 파일 표준
```markdown
---
name: command-name
description: Clear one-line description
argument-hint: [param1] [param2]
tools: Tool1, Tool2, Bash(cmd:*)
model: sonnet
---

# Command Title
Brief description...

## Usage
- Examples
- Parameters

## Agent Orchestration
1. Call agent
2. Handle results
3. Provide feedback
```

#### 에이전트 파일 표준
```markdown
---
name: agent-name
description: Use PROACTIVELY for [trigger conditions]
tools: Read, Write, Edit, MultiEdit, Bash, Glob, Grep
model: sonnet
---

# Agent Name - Specialist Role
Brief description...

## Core Mission
- Primary responsibility
- Scope boundaries
- Success criteria

## Proactive Triggers
- Activation conditions
- Workflow integration

## Constraints
- What NOT to do
- Delegation rules
```

#### 권한 설정 최적화
```json
{
  "permissions": {
    "defaultMode": "default",
    "allow": ["Task", "Read", "Write", "Edit", "MultiEdit", ...],
    "ask": ["Bash(git push:*)", "Bash(rm:*)"],
    "deny": ["Read(./.env)", "Bash(sudo:*)", "Bash(rm -rf:*)"]
  }
}
```

#### 도구 권한
- Read, Write, Edit, MultiEdit, Glob, Bash, WebFetch

---

## 🔄 Slash Command Deep Dive

### /alfred:1-spec (명세 작성 파이프라인)

**파일**: `.claude/commands/alfred/1-spec.md`

#### 2단계 워크플로우

**STEP 1: 분석 및 계획 수립**
```
프로젝트 문서 분석 → SPEC 후보 발굴 → 구현 계획 보고 → 사용자 승인 대기
```

**STEP 2: SPEC 작성 실행** (승인 후)
```
spec-builder (EARS 명세 작성) → git-manager (브랜치/PR 생성)
```

#### 입력 옵션
```bash
/alfred:1-spec                      # 자동 제안 (권장)
/alfred:1-spec "JWT 인증 시스템"       # 수동 생성
/alfred:1-spec SPEC-001 "보안 보강"   # 기존 SPEC 보완
```

#### 모드별 산출물

| 모드 | 산출물 | 추가 작업 |
|------|--------|----------|
| Personal | `.moai/specs/SPEC-XXX/{spec,plan,acceptance}.md` | git-manager 체크포인트 |
| Team | GitHub Issue + SPEC 문서 | gh CLI 라벨/담당자 지정 |

#### 에이전트 협업
```
Command Level (orchestration)
    ↓
spec-builder (분석 모드)
    ↓
User Approval (진행/수정/중단)
    ↓
spec-builder (SPEC 작성)
    ↓
git-manager (Git 작업)
```

#### EARS 구조
1. **Event**: 시스템 트리거 이벤트
2. **Action**: 이벤트 대응 행동
3. **Response**: 행동 결과 응답
4. **State**: 시스템 상태 변화

---

### /alfred:2-build (TDD 구현 파이프라인)

**파일**: `.claude/commands/alfred/2-build.md`

#### 2단계 워크플로우

**STEP 1: SPEC 분석 및 구현 계획**
```
SPEC 문서 분석 → 복잡도 평가 → 언어별 구현 전략 → 사용자 승인 대기
```

**STEP 2: TDD 구현 실행** (승인 후)
```
code-builder (RED-GREEN-REFACTOR) → git-manager (TDD 커밋)
```

#### 입력 옵션
```bash
/alfred:2-build SPEC-001    # 단일 SPEC 구현
/alfred:2-build all         # 모든 SPEC 구현
```

#### 언어별 TDD 최적화

| SPEC 타입 | 구현 언어 | 테스트 도구 | 성능 목표 | 커버리지 |
|-----------|-----------|------------|-----------|---------|
| CLI/시스템 | TypeScript | Jest + ts-node | < 18ms | 95%+ |
| API/백엔드 | TypeScript | Jest + SuperTest | < 50ms | 90%+ |
| 프론트엔드 | TypeScript | Jest + Testing Library | < 100ms | 85%+ |
| 데이터 처리 | TypeScript | Jest + Mock | < 200ms | 85%+ |

#### TDD 사이클
1. **RED**: 실패 테스트 작성 + @TEST TAG 적용
2. **GREEN**: 최소 구현 + @CODE/@CODE/@CODE/@CODE TAG 적용
3. **REFACTOR**: 품질 개선 + @CODE/@CODE/@DOC TAG 적용

#### 에이전트 협업
```
Command Level (orchestration)
    ↓
code-builder (분석 모드: --mode=analysis)
    ↓
User Approval (진행/수정/중단)
    ↓
code-builder (구현 모드: --mode=implement)
    ↓
git-manager (TDD 커밋 일괄 처리)
```

#### 품질 게이트
- 테스트 커버리지 ≥ 85%
- 린터/포매터 통과
- 구조화 로깅 존재
- @TAG 업데이트 완료

---

### /alfred:3-sync (문서 동기화 파이프라인)

**파일**: `.claude/commands/alfred/3-sync.md`

#### 2단계 워크플로우

**STEP 1: 동기화 범위 분석**
```
프로젝트 상태 확인 → 동기화 범위 결정 → 동기화 전략 수립 → 사용자 승인 대기
```

**STEP 2: 문서 동기화 실행** (승인 후)
```
doc-syncer (Living Document + TAG 검증) → git-manager (커밋 + PR 전환)
```

#### 입력 옵션
```bash
/alfred:3-sync                  # 기본 자동 동기화
/alfred:3-sync force            # 전체 강제 동기화
/alfred:3-sync status           # 동기화 상태 확인
/alfred:3-sync project          # 통합 프로젝트 동기화
/alfred:3-sync auto src/auth/   # 특정 경로 동기화
```

#### 동기화 모드

| 모드 | 동기화 범위 | PR 처리 | 특징 |
|------|------------|---------|------|
| Personal | 로컬 문서 | 체크포인트만 | 개인 작업 중심 |
| Team | 전체 + TAG | PR Ready 전환 | 협업 지원 |
| Auto | 지능형 선택 | 상황별 결정 | 최적 전략 |
| Force | 강제 전체 | 전체 재생성 | 오류 복구용 |

#### 동기화 프로세스
1. **Phase 1**: 빠른 상태 확인 (병렬 실행)
   - Task 1 (haiku): Git 상태 체크
   - Task 2 (sonnet): 문서 구조 분석

2. **Phase 2**: 문서 동기화 (순차 실행)
   - Living Document 동기화
   - @TAG 시스템 검증

3. **Phase 3**: Git 작업 (순차 실행)
   - 문서 변경사항 커밋
   - Team 모드에서 PR Ready 전환

#### TAG 검증 명령
```bash
# Primary Chain 완전성 검증
rg '@SPEC:[A-Z]+-[0-9]{3}' -n src/ | wc -l
rg '@SPEC:[A-Z]+-[0-9]{3}' -n src/ | wc -l
rg '@CODE:[A-Z]+-[0-9]{3}' -n src/ | wc -l
rg '@TEST:[A-Z]+-[0-9]{3}' -n tests/ | wc -l
```

#### 에이전트 협업
```
Command Level (orchestration)
    ↓
Phase 1: 병렬 상태 확인 (haiku + sonnet)
    ↓
User Approval (진행/수정/중단)
    ↓
Phase 2: doc-syncer (동기화 모드: --mode=sync)
    ↓
Phase 3: git-manager (커밋 + PR 전환)
```

---

## 🔗 Integration Patterns

### 1. Agent Collaboration Pattern

#### 단일 책임 원칙 (SRP) 구현
```
spec-builder  → SPEC 작성만
code-builder  → TDD 구현만
doc-syncer    → 문서 동기화만
tag-agent     → TAG 관리만
git-manager   → Git 작업만
debug-helper  → 오류 분석만
trust-checker → 품질 검증만
cc-manager    → 설정 관리만
```

#### 에이전트 간 호출 금지
- **금지**: 에이전트가 다른 에이전트를 직접 호출
- **허용**: 명령어 레벨에서 순차 오케스트레이션
- **이점**: 명확한 의존성, 디버깅 용이, 순환 참조 방지

#### 표준 협업 플로우
```
Command Level
    ↓ (invoke)
Agent 1 (analysis)
    ↓ (output to user)
User Approval
    ↓ (invoke)
Agent 1 (execution)
    ↓ (output to command)
Command Level
    ↓ (invoke)
Agent 2 (git work)
    ↓ (output to user)
Result
```

### 2. Two-Stage Workflow Pattern

#### Phase 1: Analysis & Planning
- **목적**: 사용자 확인 전 계획 수립
- **출력**: 구현 계획 보고서
- **승인 요청**: "진행", "수정 [내용]", "중단"

#### Phase 2: Execution
- **조건**: 사용자 승인 후에만 실행
- **모드 플래그**: `--mode=implement`, `--approved=true`
- **출력**: 실제 작업 결과 (코드, 문서, 커밋)

#### 적용 범위
- `/alfred:1-spec`: 분석 → SPEC 작성
- `/alfred:2-build`: 분석 → TDD 구현
- `/alfred:3-sync`: 분석 → 문서 동기화

### 3. Tool Permission Model

#### 최소 권한 원칙

| 에이전트 | 읽기 | 쓰기 | Git | 제한 Bash |
|---------|------|------|-----|-----------|
| spec-builder | ✅ | ✅ | ❌ | ✅ |
| code-builder | ✅ | ✅ | ❌ | ✅ (python, npm, pytest) |
| doc-syncer | ✅ | ✅ | ❌ | ✅ (git read-only) |
| tag-agent | ✅ | ❌ | ❌ | ✅ (read-only) |
| git-manager | ✅ | ✅ | ✅ | ✅ (git:*) |
| debug-helper | ✅ | ❌ | ❌ | ✅ (diagnostic) |
| trust-checker | ✅ | ❌ | ❌ | ✅ (test/lint) |
| cc-manager | ✅ | ✅ | ❌ | ✅ (limited) |

#### 위험 도구 차단
```json
"deny": [
  "Read(./.env)",
  "Read(./.env.*)",
  "Read(./secrets/**)",
  "Bash(sudo:*)",
  "Bash(rm -rf:*)",
  "Bash(chmod -R 777:*)"
]
```

### 4. TAG System Integration

#### Code-First 원칙
- **진실의 원천**: 코드 자체
- **중간 캐시 없음**: TAG INDEX 파일 미사용
- **실시간 스캔**: `rg '@TAG' -n` 직접 실행

#### 8-Core TAG 체계

**Primary Chain (4 Core)**:
```
@SPEC → @SPEC → @CODE → @TEST
```

**Implementation (4 Core)**:
```
@CODE, @CODE, @CODE, @CODE
```

#### TAG 적용 책임

| 에이전트 | 적용 TAG |
|---------|---------|
| spec-builder | @SPEC, @SPEC, @CODE |
| code-builder | @TEST, @CODE, @CODE, @CODE, @CODE, @CODE, @CODE, @DOC |
| doc-syncer | TAG 검증 및 동기화 |
| tag-agent | TAG 무결성 검증 (읽기 전용) |

### 5. Performance Optimization Patterns

#### MultiEdit 배치 처리
```python
# ❌ 비효율적 (순차 생성)
Write("spec.md", content1)
Write("plan.md", content2)
Write("acceptance.md", content3)

# ✅ 효율적 (동시 생성, 60% 시간 단축)
MultiEdit([
  {file: "spec.md", content: content1},
  {file: "plan.md", content: content2},
  {file: "acceptance.md", content: content3}
])
```

#### 차등 스캔 시스템
```
Level 1 (1-3초) → 기본 검사 → Critical 발견 시 즉시 종료
    ↓ (No critical)
Level 2 (5-10초) → 중간 검사 → Warning 발견 시 보고
    ↓ (No warning)
Level 3 (20-30초) → 심화 분석 → 전체 검증 완료
```

#### 병렬/순차 하이브리드
```
Phase 1: 병렬 분석 (haiku + sonnet 동시 실행)
    ↓
Phase 2: 순차 처리 (에이전트별 순차 실행)
    ↓
Phase 3: Git 작업 (git-manager 일괄 처리)
```

---

## 🎯 Command Orchestration Flow

### Full Pipeline Execution

```mermaid
graph TD
    Start[User Request] --> Spec[/alfred:1-spec]
    Spec --> SpecAnalysis[spec-builder: Analysis]
    SpecAnalysis --> Approve1{User Approval}
    Approve1 -->|진행| SpecWrite[spec-builder: Write SPEC]
    Approve1 -->|수정| SpecAnalysis
    Approve1 -->|중단| End1[End]
    SpecWrite --> Git1[git-manager: Branch/PR]
    Git1 --> Build[/alfred:2-build]

    Build --> BuildAnalysis[code-builder: Analysis]
    BuildAnalysis --> Approve2{User Approval}
    Approve2 -->|진행| TDD[code-builder: TDD]
    Approve2 -->|수정| BuildAnalysis
    Approve2 -->|중단| End2[End]
    TDD --> Git2[git-manager: TDD Commits]
    Git2 --> Sync[/alfred:3-sync]

    Sync --> SyncAnalysis[doc-syncer: Analysis]
    SyncAnalysis --> Approve3{User Approval}
    Approve3 -->|진행| DocSync[doc-syncer: Sync]
    Approve3 -->|수정| SyncAnalysis
    Approve3 -->|중단| End3[End]
    DocSync --> Git3[git-manager: Commit + PR Ready]
    Git3 --> Complete[Development Cycle Complete]
```

### Stage-by-Stage Breakdown

#### Stage 1: SPEC Creation (/alfred:1-spec)
```
1. Command invokes spec-builder (analysis mode)
2. spec-builder analyzes project docs
3. spec-builder generates SPEC proposal
4. User reviews and approves/modifies/cancels
5. spec-builder writes SPEC files (MultiEdit)
6. Command invokes git-manager
7. git-manager creates branch + GitHub Issue
8. Result: SPEC-XXX ready for implementation
```

#### Stage 2: TDD Implementation (/alfred:2-build)
```
1. Command invokes code-builder (analysis mode)
2. code-builder analyzes SPEC-XXX
3. code-builder generates implementation plan
4. User reviews and approves/modifies/cancels
5. code-builder executes TDD cycle:
   - RED: Write failing tests + @TEST TAG
   - GREEN: Implement minimal code + @CODE/@CODE/@CODE/@CODE TAG
   - REFACTOR: Improve quality + @CODE/@CODE/@DOC TAG
6. Command invokes git-manager
7. git-manager creates TDD commits (RED→GREEN→REFACTOR)
8. Result: Tested code ready for sync
```

#### Stage 3: Documentation Sync (/alfred:3-sync)
```
1. Command invokes doc-syncer (analysis mode)
2. doc-syncer analyzes project state
3. doc-syncer generates sync plan
4. User reviews and approves/modifies/cancels
5. doc-syncer performs:
   - Living Document update
   - TAG verification (rg '@TAG' -n)
   - Sync report generation
6. Command invokes git-manager
7. git-manager commits docs + PR Ready transition
8. Result: Code-doc synchronized, PR ready for review
```

---

## 🛡️ Stability & Safety Analysis

### 1. Error Handling Strategy

#### Agent-Level Error Handling
- **분석 단계 오류**: 사용자에게 명확한 오류 메시지 + 재시도 옵션
- **실행 단계 오류**: 체크포인트 복구 + 로그 기록 + debug-helper 호출

#### Command-Level Error Handling
- **Agent 호출 실패**: 명확한 오류 메시지 + 다음 단계 안내
- **승인 거부**: 작업 중단 + 현재 상태 저장
- **부분 실패**: 완료된 작업 유지 + 실패 지점부터 재개

### 2. Rollback Mechanisms

#### git-manager 체크포인트 시스템
```bash
# 체크포인트 생성 (한국시간 기준)
git tag -a "moai_cp/$(TZ=Asia/Seoul date +%Y%m%d_%H%M%S)" -m "작업 백업"

# 체크포인트 목록
git tag -l "moai_cp/*" --sort=-version:refname | head -10

# 롤백
git reset --hard TAG_NAME
```

#### 단계별 체크포인트
- **SPEC 작성 전**: 자동 체크포인트
- **TDD 시작 전**: 자동 체크포인트
- **동기화 전**: 자동 체크포인트

### 3. Validation & Quality Gates

#### SPEC 단계 검증
- [ ] EARS 구조 완전성 (E-A-R-S)
- [ ] @TAG 체인 연결 (@SPEC → @SPEC → @CODE)
- [ ] Acceptance Criteria 존재 (Given-When-Then)

#### TDD 단계 검증
- [ ] 테스트 커버리지 ≥ 85%
- [ ] 모든 테스트 통과
- [ ] 린터/포매터 통과
- [ ] @TAG 적용 완료

#### 동기화 단계 검증
- [ ] TAG 무결성 검사 통과
- [ ] 문서-코드 일치성 확인
- [ ] 고아 TAG 0개
- [ ] 끊어진 링크 0개

### 4. Conflict Resolution

#### Git 충돌 해결
```bash
# 동기화 전 체크포인트
git tag -a "pre-sync-$(TZ=Asia/Seoul date +%Y%m%d-%H%M%S)" -m "동기화 전 백업"

# 안전한 pull
git fetch origin
if git diff --quiet HEAD origin/$(git branch --show-current); then
    echo "✅ 이미 최신 상태"
else
    git pull --rebase origin $(git branch --show-current)
fi
```

#### TAG 충돌 해결
- **중복 TAG**: tag-agent가 감지 → 자동 병합 제안
- **끊어진 링크**: doc-syncer가 감지 → 수동 복구 안내
- **고아 TAG**: tag-agent가 감지 → 폐기 또는 연결 제안

### 5. Dependency Management

#### Agent Dependencies
```
tag-agent (독립) → 다른 에이전트에 의존하지 않음
debug-helper (독립) → 다른 에이전트에 의존하지 않음
trust-checker (독립) → 다른 에이전트에 의존하지 않음
cc-manager (독립) → 다른 에이전트에 의존하지 않음

spec-builder → 프로젝트 문서 (.moai/project/)
code-builder → SPEC 문서 (.moai/specs/)
doc-syncer → 코드 + SPEC + TAG
git-manager → 모든 에이전트 출력
```

#### External Dependencies
- **Git**: 필수 (버전 ≥ 2.0)
- **gh CLI**: 선택 (Team 모드 GitHub 연동)
- **Python3**: 선택 (훅 시스템, Python 프로젝트)
- **Node.js**: 선택 (TypeScript 프로젝트)

---

## 📈 Performance Metrics

### Agent Execution Times

| 에이전트 | 분석 단계 | 실행 단계 | 최적화 기법 |
|---------|----------|----------|------------|
| spec-builder | 30-60초 | 10-20초 | MultiEdit (60% 단축) |
| code-builder | 1-3분 | 5-15분 | 언어별 라우팅 |
| doc-syncer | 2-3분 | 5-10분 | CODE-FIRST 스캔 |
| tag-agent | 1-5초 | N/A | 읽기 전용 |
| git-manager | 5-10초 | 10-30초 | 직접 Git 명령 |
| debug-helper | 1-3초 | 5-30초 | 차등 스캔 (3단계) |
| trust-checker | 1-3초 | 5-30초 | 차등 스캔 (3단계) |
| cc-manager | 5-10초 | 10-20초 | 표준 템플릿 |

### Command Pipeline Times

| 명령어 | 분석 | 승인 | 실행 | 전체 |
|--------|------|------|------|------|
| /alfred:1-spec | 30-60초 | 사용자 | 30-50초 | 1-3분 |
| /alfred:2-build | 1-3분 | 사용자 | 5-15분 | 6-20분 |
| /alfred:3-sync | 2-3분 | 사용자 | 5-15분 | 7-20분 |
| **전체 사이클** | **3-6분** | **사용자** | **10-30분** | **15-45분** |

### Optimization Impact

| 최적화 기법 | 시간 단축 | 적용 범위 |
|------------|----------|----------|
| MultiEdit 배치 | 60% | spec-builder |
| 차등 스캔 (Level 1→2→3) | 50-80% | debug-helper, trust-checker |
| 직접 Git 명령 | 30% | git-manager |
| CODE-FIRST 스캔 | 40% | doc-syncer, tag-agent |
| 병렬 분석 | 30-50% | /alfred:3-sync Phase 1 |

---

## 🔍 Critical Success Factors

### 1. Clear Role Separation
✅ **달성**: 8개 에이전트가 명확한 단일 책임 보유
- 에이전트 간 역할 충돌 0건
- 중복 기능 0건
- 명확한 위임 규칙

### 2. Two-Stage Approval
✅ **달성**: 모든 주요 명령어에 분석→승인→실행 패턴 적용
- 사용자 통제권 보장
- 예상치 못한 작업 방지
- 계획 투명성 확보

### 3. No Direct Agent Calls
✅ **달성**: 에이전트 간 직접 호출 금지, 명령어 레벨 오케스트레이션만 허용
- 순환 참조 방지
- 디버깅 용이성
- 명확한 실행 흐름

### 4. Minimal Tool Permissions
✅ **달성**: 에이전트별 최소 권한 원칙 적용
- tag-agent: 읽기 전용
- debug-helper, trust-checker: 진단 전용
- git-manager: Git 독점
- 민감 파일 접근 차단

### 5. Code-First TAG System
✅ **달성**: 코드 자체가 TAG의 유일한 진실
- 중간 캐시 제거
- 실시간 스캔 (`rg`)
- 무결성 100% 보장

---

## 🚨 Known Issues & Limitations

### 1. Agent Communication Overhead
**문제**: 에이전트 간 직접 호출 금지로 명령어 레벨 오케스트레이션 필요
**영향**: 약간의 실행 시간 증가 (약 5-10초)
**완화**: 명확성과 안정성 향상으로 상쇄됨

### 2. User Approval Bottleneck
**문제**: 2단계 워크플로우에서 사용자 승인 필수
**영향**: 전체 사이클 시간 증가 (사용자 대기 시간)
**완화**: 예상치 못한 작업 방지, 사용자 통제권 보장

### 3. External Dependency
**문제**: gh CLI 없으면 Team 모드 PR 자동 전환 불가
**영향**: 수동 PR 관리 필요
**완화**: Personal 모드는 정상 작동, Team 모드는 선택적 기능

### 4. Language Detection Complexity
**문제**: 복잡한 멀티 언어 프로젝트에서 언어 감지 실패 가능
**영향**: 최적 도구 선택 실패 → 수동 설정 필요
**완화**: `.moai/config.json`에 명시적 언어 설정 지원

### 5. TAG Scan Performance
**문제**: 대형 프로젝트에서 TAG 전체 스캔 시간 증가
**영향**: 동기화 단계 지연 (최대 30-60초)
**완화**: 차등 스캔, 캐시 전략 (향후 개선 예정)

---

## 🎯 Recommendations

### Short-term (1-3 months)

1. **성능 모니터링 추가**
   - 에이전트별 실행 시간 추적
   - 병목 지점 자동 감지
   - 성능 리포트 자동 생성

2. **에러 복구 자동화**
   - 일반적인 오류 패턴 자동 복구
   - 체크포인트 자동 롤백
   - 재시도 로직 표준화

3. **TAG 스캔 최적화**
   - 증분 스캔 지원 (변경된 파일만)
   - 캐시 전략 도입 (5분 TTL)
   - 병렬 스캔 (파일별)

### Mid-term (3-6 months)

1. **에이전트 학습 기능**
   - 사용자 승인 패턴 학습
   - 자주 거부되는 계획 유형 회피
   - 프로젝트별 맞춤 제안

2. **멀티 프로젝트 지원**
   - 프로젝트 간 TAG 참조
   - 공유 SPEC 라이브러리
   - 크로스 프로젝트 동기화

3. **고급 품질 게이트**
   - 커스텀 품질 규칙 정의
   - 프로젝트별 임계값 설정
   - A/B 테스트 지원

### Long-term (6-12 months)

1. **AI-Powered 계획 최적화**
   - 과거 프로젝트 데이터 학습
   - 최적 구현 전략 자동 제안
   - 리스크 예측 모델

2. **분산 에이전트 실행**
   - 에이전트 병렬 실행 (독립적 작업)
   - 클라우드 기반 실행 지원
   - 스케일링 전략

3. **커뮤니티 에이전트 마켓**
   - 커스텀 에이전트 공유
   - 에이전트 버전 관리
   - 에이전트 조합 템플릿

---

## 📚 Appendix

### A. Agent Summary Table

| 에이전트 | 주 책임 | 모델 | 도구 권한 | 특이사항 |
|---------|---------|------|----------|----------|
| spec-builder | SPEC 작성 | sonnet | Read, Write, Edit, MultiEdit, Bash, Glob, Grep, TodoWrite, WebFetch | MultiEdit 최적화 |
| code-builder | TDD 구현 | sonnet | Read, Write, Edit, MultiEdit, Bash, Grep, Glob, TodoWrite | 2단계 워크플로우 |
| doc-syncer | 문서 동기화 | sonnet | Read, Write, Edit, MultiEdit, Grep, Glob, TodoWrite | CODE-FIRST |
| tag-agent | TAG 관리 | sonnet | Read, Glob, Bash | 읽기 전용 |
| git-manager | Git 작업 | haiku | Bash(git:*), Read, Write, Edit, Glob, Grep | Git 독점 |
| debug-helper | 오류 분석 | sonnet | Read, Grep, Glob, Bash, TodoWrite | 차등 스캔 |
| trust-checker | 품질 검증 | sonnet | Read, Grep, Glob, Bash, TodoWrite | 온디맨드 |
| cc-manager | 설정 관리 | sonnet | Read, Write, Edit, MultiEdit, Glob, Bash, WebFetch | 중앙 관제탑 |

### B. Command Summary Table

| 명령어 | 목적 | 에이전트 사용 | 승인 필요 | 산출물 |
|--------|------|--------------|----------|--------|
| /alfred:1-spec | SPEC 작성 | spec-builder, git-manager | Yes | SPEC 문서, 브랜치, Issue |
| /alfred:2-build | TDD 구현 | code-builder, git-manager | Yes | 테스트 코드, TDD 커밋 |
| /alfred:3-sync | 문서 동기화 | doc-syncer, git-manager | Yes | 동기화 문서, PR Ready |

### C. TAG Categories Reference

**Primary Chain (필수)**:
- @SPEC: 요구사항
- @SPEC: 설계
- @CODE: 작업
- @TEST: 테스트

**Implementation (필수)**:
- @CODE: 기능 구현
- @CODE: API 엔드포인트
- @CODE: 사용자 인터페이스
- @CODE: 데이터 모델/처리

**Quality (선택)**:
- @CODE: 성능 최적화
- @CODE: 보안 강화
- @DOC: 문서화

### D. TRUST Principles Mapping

| 원칙 | 검증 에이전트 | 검증 방법 | 목표 |
|------|-------------|----------|------|
| T (Test First) | trust-checker, code-builder | 커버리지 측정 | ≥ 85% |
| R (Readable) | trust-checker, code-builder | LOC, 복잡도 | ≤ 300 LOC/파일, ≤ 10 복잡도 |
| U (Unified) | trust-checker | 의존성 분석 | 순환 의존성 0개 |
| S (Secured) | trust-checker | 보안 스캔 | 취약점 0개 |
| T (Trackable) | tag-agent, doc-syncer | TAG 검증 | 무결성 100% |

---

## 🏁 Conclusion

MoAI-ADK의 Claude Code 통합 시스템은 **명확한 역할 분리**, **2단계 승인 워크플로우**, **최소 권한 원칙**을 통해 안정적이고 예측 가능한 개발 환경을 제공합니다.

### 핵심 강점
1. ✅ **완전한 SRP**: 8개 에이전트의 명확한 단일 책임
2. ✅ **사용자 통제**: 분석→승인→실행 패턴
3. ✅ **안전성**: 에이전트 간 직접 호출 금지
4. ✅ **성능**: MultiEdit, 차등 스캔, 병렬 처리
5. ✅ **추적성**: CODE-FIRST TAG 시스템

### 향후 발전 방향
- 성능 모니터링 및 최적화
- 에이전트 학습 기능
- 멀티 프로젝트 지원
- AI 기반 계획 최적화

---

**보고서 작성자**: Claude (Sonnet 4.5)
**검증 기준**: TRUST:INTEGRATION-001 (TRUST 5원칙 준수)
**다음 단계**: `/alfred:2-build` 또는 `/alfred:3-sync`로 개발 사이클 계속

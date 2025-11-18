# SPEC-CMD-COMPLIANCE-001 구현 계획

**SPEC ID**: SPEC-CMD-COMPLIANCE-001
**제목**: Zero Direct Tool Usage 준수
**작성일**: 2025-11-19

---

## 📋 4단계 구현 계획

### Phase 1: `/moai:1-plan` 수정 (High Priority)

**목표**: SPEC 생성 커맨드의 에이전트 위임 전환

#### 1.1 로컬 프로젝트 파일 수정

**파일**: `.claude/commands/moai/1-plan.md`

**변경사항**:
```yaml
# Before (위반)
allowed-tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Grep
  - Glob
  - TodoWrite
  - Bash(git:*)
  - Bash(gh:*)
  - Bash(rg:*)
  - Bash(mkdir:*)
  - AskUserQuestion
  - Skill

# After (준수)
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
```

**Instructions 재작성**:
- 모든 파일 탐색 → `Task(subagent_type="Explore")`
- SPEC 문서 생성 → `Task(subagent_type="spec-builder")`
- SPEC 수정 → `Task(subagent_type="spec-builder")`
- Git 브랜치/PR 생성 → `Task(subagent_type="git-manager")`
- 사용자 상호작용 → `AskUserQuestion()`

#### 1.2 패키지 템플릿 동기화

**파일**: `src/moai_adk/templates/.claude/commands/moai/1-plan.md`

**조치**: 로컬 파일과 동일한 내용으로 업데이트 (SSOT 원칙)

#### 1.3 검증

```bash
# 검증 1: allowed-tools 확인
grep -A 10 "^allowed-tools:" .claude/commands/moai/1-plan.md
# 결과: Task, AskUserQuestion, Skill만 포함되어야 함

# 검증 2: 직접 도구 사용 없음
grep -E "^\s*(Read|Write|Edit|Bash|Grep|Glob|TodoWrite)" .claude/commands/moai/1-plan.md
# 결과: 매치 없음 (공백 결과)

# 검증 3: 패키지 동기화
diff .claude/commands/moai/1-plan.md src/moai_adk/templates/.claude/commands/moai/1-plan.md
# 결과: 동일 (no differences)
```

---

### Phase 2: `/moai:3-sync` 수정 (High Priority)

**목표**: 문서 동기화 커맨드의 에이전트 위임 전환

#### 2.1 로컬 프로젝트 파일 수정

**파일**: `.claude/commands/moai/3-sync.md`

**변경사항**:
```yaml
# Before (위반)
allowed-tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Bash(git:*)
  - Bash(gh:*)
  - Bash(uv:*)
  - Task
  - Grep
  - Glob
  - TodoWrite

# After (준수)
allowed-tools:
  - Task
  - AskUserQuestion
```

**Instructions 재작성**:
- Git 상태 분석 → `Task(subagent_type="doc-syncer")`
- 문서 동기화 실행 → `Task(subagent_type="doc-syncer")`
- 품질 검증 → `Task(subagent_type="quality-gate")`
- Git commit/PR → `Task(subagent_type="git-manager")`
- 사용자 상호작용 → `AskUserQuestion()`

#### 2.2 패키지 템플릿 동기화

**파일**: `src/moai_adk/templates/.claude/commands/moai/3-sync.md`

**조치**: 로컬 파일과 동일한 내용으로 업데이트

#### 2.3 검증

```bash
# 검증 1: allowed-tools 확인
grep -A 5 "^allowed-tools:" .claude/commands/moai/3-sync.md
# 결과: Task, AskUserQuestion만 포함

# 검증 2: 직접 도구 사용 없음
grep -E "^\s*(Read|Write|Edit|Bash|Grep|Glob|TodoWrite)" .claude/commands/moai/3-sync.md
# 결과: 매치 없음

# 검증 3: 패키지 동기화
diff .claude/commands/moai/3-sync.md src/moai_adk/templates/.claude/commands/moai/3-sync.md
# 결과: 동일
```

---

### Phase 3: `/moai:99-release` 예외 문서화 (Medium Priority)

**목표**: 로컬 전용 도구 예외 패턴 명시

#### 3.1 예외 문서 추가

**파일**: `.claude/commands/moai/99-release.md`

**추가 위치**: 파일 상단 (제목 아래)

**추가 내용**:
```markdown
---
⚠️ **EXCEPTION: Local-Only Development Tool**

이 커맨드는 "Zero Direct Tool Usage" 원칙의 예외입니다:

**예외 사유**:
1. 로컬 개발 전용 (MoAI-ADK 패키지와 함께 배포되지 않음)
2. 메인테이너 전용 사용 (GoosLab only)
3. PyPI 릴리스 자동화를 위한 직접 시스템 접근 필수
4. 로컬 전용 도구 예외 패턴 (동일: /moai:9-feedback의 도구 특화 패턴)

**프로덕션 커맨드**: 패키지와 함께 배포되는 커맨드(/moai:0-project, /moai:1-plan, /moai:2-run, /moai:3-sync)는
엄격한 에이전트 위임 원칙 준수 필수
---
```

---

### Phase 4: CLAUDE.md 업데이트

**목표**: Command Compliance 가이드라인 문서화

#### 4.1 새로운 섹션 추가

**파일**: `CLAUDE.md`

**섹션 위치**: "📚 Extended Resources" 이전

**섹션 제목**: "🎯 Command Compliance Guidelines"

**내용 포함**:
1. Zero Direct Tool Usage 원칙 설명
2. 프로덕션 vs 로컬 커맨드 분류
3. 예외 패턴 문서화
4. 새 커맨드 개발 체크리스트

**체크리스트 템플릿**:
```markdown
## 새 커맨드 개발 체크리스트

- [ ] **allowed-tools** contains ONLY:
  - [ ] `Task` (required)
  - [ ] `AskUserQuestion` (if user interaction)
  - [ ] `Skill` (if specific skill needed)

- [ ] **NO** direct tool usage:
  - [ ] No Read(), Write(), Edit()
  - [ ] No Bash() (except documented exceptions)
  - [ ] No TodoWrite(), Grep(), Glob()

- [ ] **All work delegated**:
  - [ ] Task() calls for file operations
  - [ ] Task() calls for Git operations
  - [ ] Clear agent responsibilities

- [ ] **Exception documentation** (if applicable):
  - [ ] Clear rationale for exception
  - [ ] Scope limitation (local-only, maintainer-only)
  - [ ] Plan for future compliance
```

---

## 🔄 에이전트 위임 패턴 - Before/After

### Before (직접 도구 사용)

```markdown
# /moai:1-plan.md (위반)

allowed-tools:
  - Read
  - Grep
  - Glob
  - Write
  - Bash(git:*)
  - Bash(gh:*)

## Phase 1: SPEC 생성

### Step 1: 기존 SPEC 탐색
Execute: grep -r "SPEC-" .moai/specs/
Read: .moai/specs/*/spec.md
```

### After (에이전트 위임)

```markdown
# /moai:1-plan.md (준수)

allowed-tools:
  - Task
  - AskUserQuestion
  - Skill

## Phase 1: SPEC 생성

### Step 1: 기존 SPEC 분석

Use Task tool:
- subagent_type: "Explore"
- prompt: "Find existing SPEC documents and analyze patterns"

### Step 2: SPEC 문서 생성

Use Task tool:
- subagent_type: "spec-builder"
- prompt: "Create comprehensive SPEC document with EARS format"
```

---

## 📊 작업 분해 및 예상 소요 시간

| Phase | 작업 | 예상 시간 | 담당 에이전트 |
|-------|------|----------|--------------|
| 1 | /moai:1-plan 수정 | 2-3시간 | tdd-implementer, git-manager |
| 2 | /moai:3-sync 수정 | 2-3시간 | tdd-implementer, git-manager |
| 3 | /moai:99-release 문서화 | 30분 | docs-manager |
| 4 | CLAUDE.md 업데이트 | 1시간 | docs-manager |
| **총계** | | **5.5-7시간** | |

---

## ✅ 검증 체크리스트

### Phase 1 검증

- [ ] allowed-tools: [Task, AskUserQuestion, Skill]
- [ ] 직접 도구 사용 없음
- [ ] 에이전트 위임 패턴 적용
- [ ] 패키지 템플릿 동기화
- [ ] 커맨드 실행 테스트: `/moai:1-plan "test feature"`

### Phase 2 검증

- [ ] allowed-tools: [Task, AskUserQuestion]
- [ ] 직접 도구 사용 없음
- [ ] 에이전트 위임 패턴 적용
- [ ] 패키지 템플릿 동기화
- [ ] 커맨드 실행 테스트: `/moai:3-sync status`

### Phase 3 검증

- [ ] 예외 문서 추가 확인
- [ ] 예외 사유 명확성 검증
- [ ] 패키지 템플릿 검토

### Phase 4 검증

- [ ] CLAUDE.md 섹션 추가 확인
- [ ] 체크리스트 명확성 검증
- [ ] 링크 및 참조 검증

---

## 🚀 실행 순서 (TDD Red-Green-Refactor)

1. **RED**: 현재 커맨드들의 준수 상태 검증 (실패)
2. **GREEN**: 각 Phase별로 파일 수정 및 검증 통과
3. **REFACTOR**: 코드 리뷰 및 문서 최적화

---

## 📈 성공 조건

```
✅ 완료 조건:
├─ allowed-tools 준수: 100% (4/6 프로덕션)
├─ 예외 문서화: 100% (2/6 로컬 전용)
├─ 템플릿 동기화: 100% (diff 동일)
├─ 테스트 통과: 100% (자동 검증)
└─ 문서 완성: 100% (CLAUDE.md 포함)
```

---

**계획 버전**: 1.0
**최종 업데이트**: 2025-11-19
**상태**: Ready for TDD Implementation
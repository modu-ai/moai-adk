# Alfred 명령어 일관성 분석

> **문서 목적**: 모든 Alfred 명령어(`0-project`, `1-plan`, `2-run`, `3-sync`)의 패턴 일관성 검증 및 개선 권장사항

---

## 📊 분석 개요

**분석 범위**:
- `/alfred:0-project` - 프로젝트 초기화
- `/alfred:1-plan` - SPEC 작성
- `/alfred:2-run` - TDD 구현 (참조: alfred-command-best-practices.md)
- `/alfred:3-sync` - 문서 동기화

**검증 항목**:
1. 복잡한 작업이 에이전트에게 위임되는가?
2. 직접 도구 사용이 정당화되는가?
3. 스크립트 생성이 없는가?
4. Skill이 명시적으로 로드되는가?

---

## 🎯 2-run 패턴 (참조 구현)

`/alfred:2-run`은 MoAI-ADK의 **모범 구현**으로, 모든 명령어가 따라야 할 패턴입니다.

### 핵심 원칙

1. **100% 에이전트 위임**
   - 모든 복잡한 작업 → 에이전트 위임
   - implementation-planner, tdd-implementer, quality-gate, git-manager

2. **정당화된 직접 도구 사용**
   - Read: SPEC 문서 읽기 (컨텍스트 준비, ~5%)
   - Bash: git log 확인 (단순 검증, <10자, ~5%)
   - 나머지 80-90%: 에이전트 위임

3. **스크립트 생성 제로**
   - 임시 스크립트 없음
   - 기존 재사용 가능한 훅만 사용

4. **Skill 재사용 최대화**
   - 각 에이전트가 명시적으로 Skill 로드
   - 지식 중앙화로 유지보수 단순화

---

## ✅ `/alfred:1-plan` 분석

### 구조

```
Phase 1A: Explore agent (선택적 코드베이스 탐색)
   ↓ (탐색 결과)
Phase 1B: spec-builder agent (SPEC 계획 수립)
   ↓ (사용자 승인)
Phase 2: spec-builder agent (SPEC 파일 생성)
   ├─ spec.md (메인 스펙)
   ├─ plan.md (구현 계획)
   └─ acceptance.md (수용 기준)
   ↓
Phase 3: 선택적 git-manager 위임
   ├─ tag-agent (SPEC ID 검증)
   └─ git-manager (브랜치/PR 생성)
```

### 직접 도구 사용

**Read 도구**:
- `.moai/config/config.json` 읽기 (Line 589-596)
- 정당성: ✅ 컨텍스트 준비 (git_strategy 확인)

**Bash 도구**:
- `git branch --list` 검증 (Line 733-746)
- `git ls-remote` 검증 (Line 745)
- `gh issue list`, `gh pr list` 검증 (Line 746-747)
- 정당성: ✅ 단순 검증 (<10자 명령어)

### 에이전트 위임

| 에이전트 | 책임 | 라인 | 상태 |
|---------|------|------|------|
| Explore | 코드베이스 탐색 | 165-188 | Optional |
| spec-builder | SPEC 분석 및 생성 | 197-248, 414-461 | Required |
| tag-agent | SPEC ID 검증 | 383-408 | Required |
| git-manager | 브랜치/PR 생성 | 640-674 | Conditional |

### Skill 활용

```yaml
✅ 명시적 로드:
  - moai-foundation-specs (849)
  - moai-foundation-ears (849)
  - moai-alfred-spec-metadata-validation (853)
  - moai-alfred-tag-scanning (406)
```

### 평가: ⭐ 우수 (8.5/10)

**강점**:
- Agent 위임이 명확하고 체계적
- Skills 명시적으로 로드됨
- 3단계 승인 프로세스 명확
- 2-run 패턴과 완벽하게 일치

**개선 기회**:
- Phase B에서 Explore 결과 전달 방식 명확화 가능
- 에러 복구 절차 추가 고려

---

## ✅ `/alfred:3-sync` 분석

### 구조

```
Phase 1: 분석 & 계획
  ├─ 전제조건 검증 (1.1)
  ├─ 프로젝트 상태 분석 (1.2)
  ├─ tag-agent 위임 (1.3)
  ├─ doc-syncer 위임 (1.4)
  └─ 사용자 승인 (1.5)
   ↓ (승인됨)
Phase 2: 문서 동기화 실행
  ├─ 안전 백업 생성 (2.1)
  ├─ doc-syncer 위임 (2.2)
  └─ quality-gate 위임 (2.3)
   ↓
Phase 3: Git 작업 & PR
  ├─ git-manager 커밋 (3.1)
  ├─ PR 전환 (3.2)
  └─ PR 자동 병합 (3.3, 선택적)
   ↓
Phase 4: 완료 & 다음 단계
```

### 직접 도구 사용

**Read 도구**:
- `.moai/config.json` 읽기 (Line 160-164)
- 정당성: ✅ Git 전략 확인

**Bash 도구**:
- `git status --porcelain` (Line 156)
- `git diff --name-only` (Line 157)
- `git rev-parse --is-inside-work-tree` (Line 140)
- `which python3` (Line 144)
- 정당성: ✅ 프로젝트 상태 분석

**재사용 가능한 훅**:
```bash
# Line 434-445
python3 .claude/hooks/alfred/spec_status_hooks.py batch_update
python3 .claude/hooks/alfred/spec_status_hooks.py validate_completion SPEC_ID
python3 .claude/hooks/alfred/spec_status_hooks.py status_update SPEC_ID
```
- 정당성: ✅ 기존 인프라 훅 (재사용 가능, 경량)

### 에이전트 위임

| 에이전트 | 책임 | 라인 | 상태 |
|---------|------|------|------|
| tag-agent | TAG 검증 | 182-213 | Required |
| doc-syncer | 동기화 계획 & 실행 | 216-251, 333-389 | Required |
| quality-gate | 품질 검증 | 393-424 | Required |
| git-manager | 커밋/PR | 462-516, 533-561 | Required |

### Skill 활용

```yaml
✅ 명시적 로드:
  - moai-alfred-tag-scanning (79)
  - moai-alfred-git-workflow (82)
  - moai-alfred-trust-validation (80)
```

### 평가: ⭐ 우수 (8/10)

**강점**:
- Agent 위임이 명확
- 훅 사용이 정당화됨
- 4단계 프로세스가 체계적
- 2-run 패턴 준수

**개선 기회**:
- Phase 2.4의 훅 호출을 명시적으로 doc-syncer에 위임 고려
- 백업 위치 표준화

---

## ⚠️ `/alfred:0-project` 분석

### 구조

```
Phase 1: 명령어 라우팅
  └─ project-manager agent (모드 분석)
   ↓ (모드 결정)
Phase 2: 모드 실행
  └─ project-manager agent (언어-우선 초기화)
   ↓
Phase 2.5: 컨텍스트 저장 ❌❌❌
  └─ Python 코드 직접 작성 (Alfred 패턴 위반)
   ↓
Phase 3: 완료 & 다음 단계
```

### 주요 문제: Phase 2.5

**현재 구현 (Line 177-254)**:

```python
# ❌ WRONG: Alfred가 직접 Python 코드 작성
try:
    from moai_adk.core.context_manager import ContextManager
    context_mgr = ContextManager(project_root)
except ImportError:
    context_mgr = None

phase_data = {
    "phase": "0-project",
    "timestamp": ...,
    "outputs": {...},
    "files_created": [],
    "next_phase": "1-plan"
}

files_created_absolute = []
for rel_path in files_created_relative:
    try:
        abs_path = validate_and_convert_path(rel_path, project_root)
        files_created_absolute.append(abs_path)
    except (ValueError, FileNotFoundError) as e:
        print(f"Warning: Could not validate path {rel_path}: {e}")

if context_mgr:
    try:
        saved_path = context_mgr.save_phase_result(phase_data)
        print(f"Phase context saved to: {saved_path}")
    except (IOError, OSError) as e:
        print(f"Warning: Failed to save phase context: {e}")
```

### 위반 사항

| 항목 | 현재 | 문제 |
|------|------|------|
| ContextManager 임포트 | 직접 실행 | ❌ Agent가 담당해야 함 |
| 데이터 구조 조작 | 직접 작성 | ❌ Agent가 담당해야 함 |
| 파일 경로 변환 | validate_and_convert_path 호출 | ❌ Agent 함수 호출 |
| try/except 처리 | 직접 작성 | ❌ Agent가 담당해야 함 |
| 에러 로깅 | print 문 | ❌ Agent가 담당해야 함 |

### 올바른 패턴

```yaml
# ✅ CORRECT PATTERN
---
name: alfred:0-project
---

## Phase 2.5: Context 저장

Use Task tool:
- subagent_type: "project-manager"
- description: "Save phase context using ContextManager"
- prompt: |
    You are the project-manager agent.

    **Task**: Save phase execution context.

    **Context Extraction**:
    1. Extract project metadata from .moai/config.json
       - project_name, mode, language, tech_stack
    2. List files created in this phase
       - .moai/config/config.json
       - .moai/project/product.md
       - .moai/project/structure.md
       - CLAUDE.md

    **Context Saving**:
    1. Initialize ContextManager with project_root
    2. Build phase_data structure:
       - phase: "0-project"
       - status: "completed"
       - outputs: (from config)
       - files_created: (validated absolute paths)
       - next_phase: "1-plan"
    3. Validate file paths (convert relative → absolute)
    4. Save phase result using ContextManager.save_phase_result()
    5. Handle errors gracefully (warn but continue)

    **Output**: Confirmation of saved context path or warning
```

### 직접 도구 사용

**Read 도구**:
- `.moai/config/config.json` 존재 확인 (Line 71)
- 정당성: ✅ 모드 결정을 위한 컨텍스트

**Bash 도구**:
- `ls -la .moai/` 등 (Line 14에 allowed-tools에 명시)
- 정당성: ✅ 구조 검증

### Skill 활용

```yaml
✅ 명시적 로드:
  - moai-project-language-initializer (Line 309)
  - moai-project-config-manager (Line 310)
  - moai-project-template-optimizer (Line 311)
  - moai-project-batch-questions (Line 312)
```

### 평가: ⚠️ 개선 필요 (6.5/10)

**강점**:
- Phase 1, 2, 3는 정상 구조
- Skills 명시적으로 로드됨
- 에이전트 위임 기본 구조 있음

**문제**:
- ❌ Phase 2.5에서 Python 코드 직접 작성
- ❌ ContextManager 직접 사용
- ❌ 데이터 구조 변환 로직이 명령어에 있음
- ❌ 에러 처리가 명령어 수준에 있음

**영향**:
- 테스트 불가능 (Alfred는 테스트 불가)
- 유지보수 어려움 (로직이 분산)
- 에러 추적 어려움
- Agent가 책임을 이행하지 못함

---

## 📋 전체 평가 표

| 명령어 | Agent 위임 | 직접 도구 | 스크립트 | Skill | 에러 처리 | 전체 평가 |
|--------|----------|---------|---------|-------|---------|----------|
| **1-plan** | ✅ 100% | ✅ 정당 | ❌ 없음 | ✅ 명시 | ✅ 우수 | **8.5/10** ⭐ |
| **2-run** | ✅ 100% | ✅ 정당 | ❌ 없음 | ✅ 명시 | ✅ 우수 | **9/10** ⭐⭐ |
| **3-sync** | ✅ 95% | ✅ 정당 | ✅ 정당 | ✅ 명시 | ✅ 우수 | **8/10** ⭐ |
| **0-project** | ⚠️ 70% | ⚠️ 문제 | N/A | ✅ 명시 | ❌ 위반 | **6.5/10** ⚠️ |

---

## 💡 개선 방안

### Priority 1: `/alfred:0-project` Phase 2.5 리팩토링

**목표**: Alfred 패턴 100% 준수

**변경 내용**:
1. Phase 2.5의 모든 Python 로직을 project-manager agent에 위임
2. project-manager agent가 ContextManager 관리
3. 명령어에서는 Task() 호출만 수행

**예상 효과**:
- 테스트 가능성: ❌ → ✅
- 유지보수성: ⚠️ → ✅
- 에러 추적: ⚠️ → ✅
- 평가: 6.5/10 → 8.5/10

### Priority 2: 일관성 검증 자동화

**제안**:
1. 모든 명령어에 일관성 체크마크 추가
2. 검증 체크리스트 CLAUDE.md에 포함
3. 새 명령어 작성 시 템플릿 제공

### Priority 3: Skill 로딩 명시화 (선택적)

**제안**:
1. 각 에이전트에 Skill 로딩 명시 추가
2. Agent prompt에서 명시적으로 Skill() 호출 지정
3. 예시:
   ```markdown
   SKILL CALLS:
   - Skill("moai-foundation-tags") for TAG validation
   - Skill("moai-alfred-git-workflow") for Git operations
   ```

---

## 📊 현재 상태 체크리스트

### 구현 완료
- ✅ 1-plan: 패턴 준수 (8.5/10)
- ✅ 2-run: 참조 구현 (9/10)
- ✅ 3-sync: 패턴 준수 (8/10)

### 개선 필요
- ⚠️ 0-project: Phase 2.5 리팩토링 필요 (6.5/10)

### 다음 Phase
1. Phase 5 (선택적): 훅 스크립트 평가
2. Phase 4 완료 후: 0-project 개선안 작성
3. Phase 4 완료 후: 일관성 가이드 문서 작성

---

## 📚 참고 자료

- `.moai/docs/patterns/alfred-command-best-practices.md` - 2-run 참조 구현
- `.moai/docs/patterns/agent-skill-mapping.md` - Agent Skill 매핑
- `/src/moai_adk/templates/CLAUDE.md` - MoAI-ADK 철학

---

**작성 일자**: 2025-11-12
**문서 버전**: 1.0
**상태**: Complete - Phase 4 분석 완료

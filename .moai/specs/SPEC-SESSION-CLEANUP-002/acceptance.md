# Acceptance Criteria: SPEC-SESSION-CLEANUP-002

<!-- @ACCEPTANCE:SESSION-CLEANUP-002 -->

---

## Overview

이 문서는 SPEC-SESSION-CLEANUP-002의 구현 완료를 검증하기 위한 8개의 상세한 테스트 시나리오를 정의합니다. 모든 시나리오는 **Given-When-Then** 형식을 따릅니다.

**테스트 범위**:
- Group 1: AskUserQuestion 호출 검증 (4 tests)
- Group 2: 세션 정리 검증 (2 tests)
- Group 3: 품질 검증 (2 tests)

**테스트 환경**:
- MoAI-ADK v0.8.0 (target version)
- Python 3.11+
- Claude Code CLI
- Branch: `feature/SPEC-SESSION-CLEANUP-002`

---

## Test Group 1: AskUserQuestion 호출 검증

### TEST-001: `/alfred:0-project` 완료 후 AskUserQuestion 호출 검증

**@TAG**: `@TEST:IMPL-001`

**Priority**: MUST

**Given**:
- 사용자가 `/alfred:0-project` 커맨드를 실행한다
- 프로젝트 초기화가 성공적으로 완료된다
- `.moai/project/product.md`, `structure.md`, `tech.md` 파일이 생성된다

**When**:
- 커맨드 종료 직전 (Success Criteria 체크 후)
- Alfred가 Final Step 섹션을 실행한다

**Then**:
- AskUserQuestion tool이 정확히 1번 호출된다 (batched design)
- 다음 3개 옵션이 제시된다:
  1. `"📋 스펙 작성 진행"` - description: `"/alfred:1-plan 실행"`
  2. `"🔍 프로젝트 구조 검토"` - description: `"현재 상태 확인"`
  3. `"🔄 새 세션 시작"` - description: `"/clear 실행"`
- Question 텍스트: `"프로젝트 초기화가 완료되었습니다. 다음으로 뭘 하시겠습니까?"`
- Header 텍스트: `"다음 단계"`
- Prose 제안 (예: "You can now run...") 이 **출력되지 않는다**

**Validation Method**:
```bash
# 1. Run command
/alfred:0-project

# 2. Verify AskUserQuestion call (check logs or output)
# Expected: AskUserQuestion invoked with 3 options

# 3. Check for prose patterns (should be 0)
rg "You can now run" src/moai_adk/templates/.claude/commands/alfred-0-project.md
# Expected: No matches
```

**Definition of Done**:
- ✅ AskUserQuestion 호출 확인
- ✅ 3개 옵션 모두 올바르게 표시
- ✅ Prose 패턴 검색 결과 0건

---

### TEST-002: `/alfred:1-plan` 완료 후 AskUserQuestion 호출 검증

**@TAG**: `@TEST:IMPL-002`

**Priority**: MUST

**Given**:
- 사용자가 `/alfred:1-plan` 커맨드를 실행한다
- SPEC 문서 3개 파일이 생성된다 (spec.md, plan.md, acceptance.md)
- `.moai/specs/SPEC-XXX-001/` 디렉토리가 존재한다

**When**:
- 커맨드 종료 직전 (SPEC 생성 완료 후)
- Alfred가 Final Step 섹션을 실행한다

**Then**:
- AskUserQuestion tool이 정확히 1번 호출된다
- 다음 3개 옵션이 제시된다:
  1. `"🚀 구현 진행"` - description: `"/alfred:2-run SPEC-XXX-001 실행"`
  2. `"✏️ SPEC 수정"` - description: `"현재 SPEC 재작업"`
  3. `"🔄 새 세션 시작"` - description: `"/clear 실행"`
- Question 텍스트: `"SPEC 작성이 완료되었습니다. 다음으로 뭘 하시겠습니까?"`
- Header 텍스트: `"다음 단계"`
- Prose 제안이 **출력되지 않는다**

**Validation Method**:
```bash
# 1. Run command
/alfred:1-plan "Test Feature"

# 2. Verify AskUserQuestion call
# Expected: AskUserQuestion invoked with 3 options

# 3. Check for prose patterns
rg "You can now run" src/moai_adk/templates/.claude/commands/alfred-1-plan.md
# Expected: No matches
```

**Definition of Done**:
- ✅ AskUserQuestion 호출 확인
- ✅ 3개 옵션 모두 올바르게 표시
- ✅ SPEC ID가 옵션 description에 올바르게 포함됨

---

### TEST-003: `/alfred:2-run` 완료 후 AskUserQuestion 호출 검증

**@TAG**: `@TEST:IMPL-003`

**Priority**: MUST

**Given**:
- 사용자가 `/alfred:2-run SPEC-XXX-001` 커맨드를 실행한다
- TDD 구현이 완료된다 (RED → GREEN → REFACTOR)
- 모든 테스트가 통과한다

**When**:
- 커맨드 종료 직전 (구현 완료 후)
- Alfred가 Final Step 섹션을 실행한다

**Then**:
- AskUserQuestion tool이 정확히 1번 호출된다
- 다음 3개 옵션이 제시된다:
  1. `"📚 문서 동기화"` - description: `"/alfred:3-sync 실행"`
  2. `"🧪 추가 테스트/검증"` - description: `"테스트 재실행"`
  3. `"🔄 새 세션 시작"` - description: `"/clear 실행"`
- Question 텍스트: `"구현이 완료되었습니다. 다음으로 뭘 하시겠습니까?"`
- Header 텍스트: `"다음 단계"`
- Prose 제안이 **출력되지 않는다**

**Validation Method**:
```bash
# 1. Run command
/alfred:2-run SPEC-XXX-001

# 2. Verify AskUserQuestion call
# Expected: AskUserQuestion invoked with 3 options

# 3. Check for prose patterns
rg "You can now run" src/moai_adk/templates/.claude/commands/alfred-2-run.md
# Expected: No matches
```

**Definition of Done**:
- ✅ AskUserQuestion 호출 확인
- ✅ 3개 옵션 모두 올바르게 표시
- ✅ 구현 완료 후에만 호출됨 (테스트 실패 시 호출 안 됨)

---

### TEST-004: `/alfred:3-sync` 완료 후 AskUserQuestion 호출 검증

**@TAG**: `@TEST:IMPL-004`

**Priority**: MUST

**Given**:
- 사용자가 `/alfred:3-sync` 커맨드를 실행한다
- 문서 동기화가 완료된다 (README.md, CHANGELOG.md 업데이트)
- @TAG 체인이 검증된다

**When**:
- 커맨드 종료 직전 (동기화 완료 후)
- Alfred가 Final Step 섹션을 실행한다

**Then**:
- AskUserQuestion tool이 정확히 1번 호출된다
- 다음 3개 옵션이 제시된다:
  1. `"📋 다음 기능 계획"` - description: `"/alfred:1-plan 실행"`
  2. `"🔀 PR 병합"` - description: `"main 브랜치로 병합"`
  3. `"✅ 세션 완료"` - description: `"작업 종료"`
- Question 텍스트: `"문서 동기화가 완료되었습니다. 다음으로 뭘 하시겠습니까?"`
- Header 텍스트: `"다음 단계"`
- Prose 제안이 **출력되지 않는다**

**Validation Method**:
```bash
# 1. Run command
/alfred:3-sync

# 2. Verify AskUserQuestion call
# Expected: AskUserQuestion invoked with 3 options

# 3. Check for prose patterns
rg "You can now run" src/moai_adk/templates/.claude/commands/alfred-3-sync.md
# Expected: No matches
```

**Definition of Done**:
- ✅ AskUserQuestion 호출 확인
- ✅ 3개 옵션 모두 올바르게 표시
- ✅ "세션 완료" 옵션이 포함됨

---

## Test Group 2: 세션 정리 검증

### TEST-005: 세션 요약 생성 (Markdown 형식 출력)

**@TAG**: `@TEST:IMPL-005`

**Priority**: SHOULD

**Given**:
- 사용자가 `/alfred:3-sync` 커맨드를 완료한다
- TodoWrite에 3개 이상의 `completed` 작업이 있다
- Git commit이 2개 이상 생성되었다

**When**:
- 사용자가 AskUserQuestion에서 "✅ 세션 완료" 옵션을 선택한다
- Alfred가 세션 정리 프로토콜을 실행한다

**Then**:
- 세션 요약이 **직접 Markdown 형식**으로 출력된다 (Bash wrapping 없음)
- 다음 섹션들이 포함된다:
  1. `## 🎊 세션 요약` (헤더)
  2. `### 완료된 작업` (TodoWrite에서 추출)
  3. `### Git 통계` (커밋 수, 변경된 파일, 라인 변경)
  4. `### 다음 권장 작업` (optional)
- 모든 섹션이 Markdown 표준을 따른다
- Bash 명령어로 wrap 되어 있지 **않다** (예: `cat << 'EOF'` 사용 금지)

**Validation Method**:
```bash
# 1. Complete a full workflow
/alfred:0-project
/alfred:1-plan "Test"
/alfred:2-run SPEC-TEST-001
/alfred:3-sync

# 2. Select "세션 완료" option
# Expected: Session summary displayed in Markdown

# 3. Verify format
# Expected sections:
# - ## 🎊 세션 요약
# - ### 완료된 작업
# - ### Git 통계
# - ### 다음 권장 작업

# 4. Verify no Bash wrapping
rg "cat <<" src/moai_adk/templates/.claude/agents/agent-alfred.md
# Expected: No matches in session summary section
```

**Expected Output Example**:
```markdown
## 🎊 세션 요약

### 완료된 작업
- ✅ 프로젝트 초기화 완료
- ✅ SPEC-TEST-001 작성 완료
- ✅ 사용자 인증 구현 완료
- ✅ 테스트 작성 및 통과
- ✅ 문서 동기화 완료

### Git 통계
- 📝 생성된 커밋: 5개
- 📂 변경된 파일: 12개
- ➕ 추가된 라인: +450
- ➖ 삭제된 라인: -120

### 다음 권장 작업
1. PR 생성 및 리뷰 요청
2. 다음 SPEC 작성 시작
```

**Definition of Done**:
- ✅ 세션 요약이 Markdown 형식으로 출력됨
- ✅ 모든 필수 섹션 포함 (완료된 작업, Git 통계)
- ✅ Bash wrapping 없음

---

### TEST-006: TodoWrite 정리 (completed 작업 추출)

**@TAG**: `@TEST:IMPL-006`

**Priority**: SHOULD

**Given**:
- 커맨드 실행 중 TodoWrite에 5개 작업이 있다
- 그 중 3개가 `completed` 상태이다
- 2개가 `pending` 또는 `in_progress` 상태이다

**When**:
- 커맨드 종료 직전 (AskUserQuestion 호출 직전)
- TodoWrite Cleanup Logic이 실행된다

**Then**:
- 3개의 `completed` 작업이 session_context에 저장된다
- session_context 구조:
  ```python
  {
      "completed_tasks": [
          {"content": "Task 1", "status": "completed"},
          {"content": "Task 2", "status": "completed"},
          {"content": "Task 3", "status": "completed"}
      ],
      "command": "/alfred:X-command",
      "timestamp": "2025-10-30T..."
  }
  ```
- `pending` 및 `in_progress` 작업은 **포함되지 않는다**

**Validation Method**:
```bash
# 1. Run a command with multiple tasks
/alfred:2-run SPEC-XXX-001

# 2. Observe TodoWrite updates during execution
# - Task 1: pending → in_progress → completed
# - Task 2: pending → in_progress → completed
# - Task 3: pending → in_progress (still running)

# 3. Verify session_context at completion
# Expected: Only completed tasks (Task 1, Task 2) extracted
```

**Pseudocode Verification**:
```python
# This logic should be in agent-alfred.md
completed_tasks = [task for task in todos if task.status == "completed"]
assert len(completed_tasks) == 3
assert all(task["status"] == "completed" for task in completed_tasks)
```

**Definition of Done**:
- ✅ `completed` 작업만 추출됨
- ✅ session_context에 올바르게 저장됨
- ✅ `pending` / `in_progress` 작업은 제외됨

---

## Test Group 3: 품질 검증

### TEST-007: Prose 패턴 검색 (검출 0건)

**@TAG**: `@TEST:IMPL-007`

**Priority**: MUST

**Given**:
- 4개 커맨드 파일 수정이 완료된다
- Phase 2A가 완료된다

**When**:
- 다음 grep 검색을 실행한다:
  ```bash
  rg "You can now run" src/moai_adk/templates/.claude/commands/alfred-*.md
  ```

**Then**:
- 검색 결과가 **0건**이다
- 추가 검색 (다른 prose 패턴):
  ```bash
  rg "Next, you can|You may now|Now you can" src/moai_adk/templates/.claude/commands/alfred-*.md
  ```
- 모든 검색 결과가 **0건**이다

**Validation Method**:
```bash
# 1. Search for common prose patterns
rg "You can now run" src/moai_adk/templates/.claude/commands/alfred-*.md
rg "Next, you can" src/moai_adk/templates/.claude/commands/alfred-*.md
rg "You may now" src/moai_adk/templates/.claude/commands/alfred-*.md
rg "Now you can" src/moai_adk/templates/.claude/commands/alfred-*.md

# Expected: No matches for all searches
```

**Prohibited Patterns**:
- ❌ "You can now run `/alfred:X`..."
- ❌ "Next, you can proceed to..."
- ❌ "You may now execute..."
- ❌ "Now you can start..."

**Allowed Patterns**:
- ✅ AskUserQuestion with options
- ✅ "Use AskUserQuestion tool" (in comments)

**Definition of Done**:
- ✅ 모든 prose 패턴 검색 결과 0건
- ✅ 4개 커맨드 파일 모두 clean

---

### TEST-008: Batched 디자인 (호출 횟수 = 1)

**@TAG**: `@TEST:IMPL-008`

**Priority**: MUST

**Given**:
- 각 커맨드 완료 시 AskUserQuestion을 호출한다
- 3개 옵션을 제시해야 한다

**When**:
- 커맨드 종료 시 Final Step 섹션이 실행된다

**Then**:
- AskUserQuestion이 정확히 **1번만** 호출된다 (batched design)
- 1회 호출에 3개 옵션이 모두 포함된다
- Sequential 호출 (3번 호출)이 **아니다**

**Validation Method**:
```bash
# 1. Run any command
/alfred:0-project

# 2. Monitor AskUserQuestion calls
# Expected: 1 call with 3 options in "questions" array

# 3. Verify code structure in template files
rg "AskUserQuestion\(" src/moai_adk/templates/.claude/commands/alfred-*.md -A 15

# Expected: Single AskUserQuestion call with:
# questions=[
#     {
#         "question": "...",
#         "options": [
#             {"label": "Option 1", ...},
#             {"label": "Option 2", ...},
#             {"label": "Option 3", ...}
#         ]
#     }
# ]
```

**Anti-pattern (Prohibited)**:
```python
# ❌ WRONG: Sequential calls (3 times)
AskUserQuestion(questions=[{"question": "Option 1?", ...}])
AskUserQuestion(questions=[{"question": "Option 2?", ...}])
AskUserQuestion(questions=[{"question": "Option 3?", ...}])
```

**Correct Pattern**:
```python
# ✅ CORRECT: Batched call (1 time)
AskUserQuestion(
    questions=[
        {
            "question": "다음으로 뭘 하시겠습니까?",
            "options": [
                {"label": "Option 1", ...},
                {"label": "Option 2", ...},
                {"label": "Option 3", ...}
            ]
        }
    ]
)
```

**Definition of Done**:
- ✅ 각 커맨드에서 AskUserQuestion 1번만 호출
- ✅ 1회 호출에 3개 옵션 포함
- ✅ Sequential 호출 패턴 없음

---

## Test Execution Summary

### Test Results Table

| Test ID    | Description                               | Priority | Status      |
| ---------- | ----------------------------------------- | -------- | ----------- |
| TEST-001   | `/alfred:0-project` AskUserQuestion 검증  | MUST     | ⏳ Pending  |
| TEST-002   | `/alfred:1-plan` AskUserQuestion 검증     | MUST     | ⏳ Pending  |
| TEST-003   | `/alfred:2-run` AskUserQuestion 검증      | MUST     | ⏳ Pending  |
| TEST-004   | `/alfred:3-sync` AskUserQuestion 검증     | MUST     | ⏳ Pending  |
| TEST-005   | 세션 요약 생성 (Markdown)                 | SHOULD   | ⏳ Pending  |
| TEST-006   | TodoWrite 정리 (completed 추출)           | SHOULD   | ⏳ Pending  |
| TEST-007   | Prose 패턴 검색 (0건)                     | MUST     | ⏳ Pending  |
| TEST-008   | Batched 디자인 (1회 호출)                 | MUST     | ⏳ Pending  |

**Note**: 상태는 구현 중 업데이트됩니다 (⏳ Pending → ✅ Pass / ❌ Fail)

---

## Quality Metrics

### Success Criteria

- **MUST 테스트 통과율**: 100% (6/6 tests)
- **SHOULD 테스트 통과율**: 80% 이상 (2/2 tests)
- **Prose 패턴 검출**: 0건
- **Batched 디자인 준수**: 100% (4/4 commands)

### Performance Criteria

- AskUserQuestion 호출 시간: < 500ms
- 세션 요약 생성 시간: < 1000ms
- TodoWrite 상태 변경: < 100ms

---

## Test Environment

### Required Tools
- Claude Code CLI (latest version)
- MoAI-ADK v0.8.0 (target version)
- Python 3.11+
- ripgrep (rg) for pattern search

### Setup Steps
1. Clone repository: `git clone https://github.com/modu-ai/moai-adk.git`
2. Checkout branch: `git checkout feature/SPEC-SESSION-CLEANUP-002`
3. Install dependencies: `uv sync --all-extras`
4. Verify environment: `python --version`, `rg --version`

---

## Definition of Done

**전체 SPEC-SESSION-CLEANUP-002 구현이 완료되려면**:

1. ✅ **모든 MUST 테스트 통과** (TEST-001, 002, 003, 004, 007, 008)
2. ✅ **80% 이상의 SHOULD 테스트 통과** (TEST-005, 006)
3. ✅ **Prose 패턴 검색 결과 0건**
4. ✅ **4개 커맨드 파일 모두 Final Step 섹션 추가 완료**
5. ✅ **agent-alfred.md에 세션 정리 로직 추가 완료**
6. ✅ **CHANGELOG.md 업데이트 완료**
7. ✅ **CLAUDE.md 검증 완료**
8. ✅ **Git commit 생성 완료** (Alfred co-authorship 포함)

**검증 방법**:
```bash
# Run full test suite
./scripts/test_session_cleanup.sh

# Expected output:
# TEST-001: ✅ PASS
# TEST-002: ✅ PASS
# TEST-003: ✅ PASS
# TEST-004: ✅ PASS
# TEST-005: ✅ PASS
# TEST-006: ✅ PASS
# TEST-007: ✅ PASS
# TEST-008: ✅ PASS
#
# Total: 8/8 tests passed (100%)
```

---

**End of Acceptance Criteria: SPEC-SESSION-CLEANUP-002**

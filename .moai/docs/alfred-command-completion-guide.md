# Alfred Command Completion Pattern - Implementation Guide

<!-- @DOC:PATTERN-GUIDELINES -->

---

## Overview

This guide provides comprehensive instructions for implementing the **Alfred Command Completion Pattern** across all MoAI-ADK commands. The pattern ensures consistent user experience by using `AskUserQuestion` tool to guide users through workflow transitions.

**Key Principles**:
- ✅ ALWAYS use `AskUserQuestion` at command completion
- ❌ NEVER suggest next steps in prose (e.g., "You can now run...")
- ✅ Provide 3-4 clear options with structured format
- ✅ Use user's `conversation_language` for all text
- ✅ Apply batched design (1 call with multiple questions when applicable)

---

## Pattern Components

### 1. AskUserQuestion Structure

**Required Format**:
```python
AskUserQuestion(
    questions=[
        {
            "question": "<question text in user's language>",
            "header": "<section header in user's language>",
            "options": [
                {
                    "label": "<emoji> <option name>",
                    "description": "<brief description>"
                },
                # ... 2-3 more options
            ]
        }
    ]
)
```

**Field Requirements**:
- `question`: Full question text in `conversation_language`
- `header`: Section header (e.g., "다음 단계", "Next Steps")
- `options`: Array of 3-4 option objects
  - `label`: Short name with emoji prefix
  - `description`: Explanation of what happens when selected

---

## Command-Specific Implementation

### Pattern 1: `/alfred:0-project` Completion

**Context**: After project initialization completes

**Implementation Location**:
- File: `src/moai_adk/templates/.claude/commands/alfred/0-project.md`
- Section: Add at the end, after success criteria

**Code Template**:
```python
# After all project initialization tasks complete
AskUserQuestion(
    questions=[
        {
            "question": "프로젝트 초기화가 완료되었습니다. 다음으로 뭘 하시겠습니까?",
            "header": "다음 단계",
            "options": [
                {
                    "label": "📋 스펙 작성 진행",
                    "description": "/alfred:1-plan 실행하여 기능 명세 작성"
                },
                {
                    "label": "🔍 프로젝트 구조 검토",
                    "description": "생성된 파일과 디렉토리 구조 확인"
                },
                {
                    "label": "🔄 새 세션 시작",
                    "description": "/clear로 현재 세션 종료"
                }
            ]
        }
    ]
)
```

**Localization Examples**:

**English**:
```python
"question": "Project initialization completed. What would you like to do next?",
"header": "Next Steps",
"options": [
    {"label": "📋 Start Planning", "description": "Run /alfred:1-plan to create specifications"},
    {"label": "🔍 Review Structure", "description": "Check generated files and directories"},
    {"label": "🔄 New Session", "description": "End current session with /clear"}
]
```

**Japanese**:
```python
"question": "プロジェクトの初期化が完了しました。次に何をしますか？",
"header": "次のステップ",
"options": [
    {"label": "📋 仕様書作成", "description": "/alfred:1-plan を実行して仕様を作成"},
    {"label": "🔍 構造確認", "description": "生成されたファイルとディレクトリを確認"},
    {"label": "🔄 新しいセッション", "description": "/clear で現在のセッションを終了"}
]
```

---

### Pattern 2: `/alfred:1-plan` Completion

**Context**: After SPEC creation completes

**Implementation Location**:
- File: `src/moai_adk/templates/.claude/commands/alfred/1-plan.md`
- Section: Add at the end, after SPEC validation

**Code Template**:
```python
# Extract SPEC ID from created spec
spec_id = extract_spec_id_from_path()  # e.g., "SPEC-AUTH-001"

AskUserQuestion(
    questions=[
        {
            "question": f"SPEC 작성이 완료되었습니다 ({spec_id}). 다음으로 뭘 하시겠습니까?",
            "header": "다음 단계",
            "options": [
                {
                    "label": "🚀 구현 진행",
                    "description": f"/alfred:2-run {spec_id} 실행하여 TDD 구현"
                },
                {
                    "label": "✏️ SPEC 수정",
                    "description": "현재 SPEC 문서를 수정하고 재검증"
                },
                {
                    "label": "🔄 새 세션 시작",
                    "description": "/clear로 현재 세션 종료"
                }
            ]
        }
    ]
)
```

**Dynamic SPEC ID Handling**:
```python
def extract_spec_id_from_path():
    """Extract SPEC ID from .moai/specs/SPEC-XXX-001/ directory"""
    import glob
    spec_dirs = glob.glob('.moai/specs/SPEC-*-*/')
    if spec_dirs:
        # Return the most recently created SPEC
        latest_spec = max(spec_dirs, key=os.path.getctime)
        return os.path.basename(latest_spec.rstrip('/'))
    return "SPEC-XXX-001"  # Fallback
```

---

### Pattern 3: `/alfred:2-run` Completion

**Context**: After TDD implementation completes (RED → GREEN → REFACTOR)

**Implementation Location**:
- File: `src/moai_adk/templates/.claude/commands/alfred/2-run.md`
- Section: Add at the end, after test validation

**Code Template**:
```python
# After all TDD cycles complete and tests pass
AskUserQuestion(
    questions=[
        {
            "question": "구현이 완료되었습니다. 다음으로 뭘 하시겠습니까?",
            "header": "다음 단계",
            "options": [
                {
                    "label": "📚 문서 동기화",
                    "description": "/alfred:3-sync 실행하여 문서 업데이트"
                },
                {
                    "label": "🧪 추가 테스트/검증",
                    "description": "추가 테스트 시나리오 실행 및 커버리지 확인"
                },
                {
                    "label": "🔄 새 세션 시작",
                    "description": "/clear로 현재 세션 종료"
                }
            ]
        }
    ]
)
```

**Pre-condition Check**:
```python
def validate_implementation_complete():
    """Verify all conditions before showing AskUserQuestion"""
    checks = {
        "all_tests_passed": run_tests(),
        "coverage_threshold": check_coverage(min_threshold=80),
        "linting_passed": run_linter(),
        "commits_created": check_git_commits()
    }

    if not all(checks.values()):
        raise ImplementationIncompleteError(
            f"Cannot complete command: {checks}"
        )
    return True
```

---

### Pattern 4: `/alfred:3-sync` Completion

**Context**: After documentation sync completes

**Implementation Location**:
- File: `src/moai_adk/templates/.claude/commands/alfred/3-sync.md`
- Section: Add at the end, after TAG validation

**Code Template**:
```python
# After documentation sync and TAG verification
AskUserQuestion(
    questions=[
        {
            "question": "문서 동기화가 완료되었습니다. 다음으로 뭘 하시겠습니까?",
            "header": "다음 단계",
            "options": [
                {
                    "label": "📋 다음 기능 계획",
                    "description": "/alfred:1-plan 실행하여 새 SPEC 작성"
                },
                {
                    "label": "🔀 PR 병합",
                    "description": "Pull Request를 main 브랜치로 병합"
                },
                {
                    "label": "✅ 세션 완료",
                    "description": "작업을 종료하고 세션 요약 확인"
                }
            ]
        }
    ]
)
```

**Option 3 Handler** (Session Summary):
```python
def handle_session_completion():
    """Generate session summary when user selects '세션 완료'"""
    summary = generate_session_summary()
    print(summary)
    # Optionally save to .moai/reports/session-summary-{timestamp}.md
```

---

## TodoWrite Cleanup Protocol

### When to Clean TodoWrite

**Timing**: Immediately **BEFORE** calling `AskUserQuestion`

**Steps**:
1. Extract all `completed` tasks from TodoWrite
2. Store completed tasks in session context
3. Optionally reset TodoWrite for next command

**Implementation**:
```python
def cleanup_todowrite_before_completion():
    """Extract and store completed tasks"""
    # Pseudo-code (actual implementation depends on TodoWrite API)
    completed_tasks = [
        task for task in todowrite.get_all()
        if task.status == "completed"
    ]

    session_context = {
        "command": current_command_name,  # e.g., "/alfred:1-plan"
        "completed_tasks": completed_tasks,
        "timestamp": datetime.now().isoformat(),
        "spec_id": current_spec_id if applicable else None
    }

    # Store for session summary
    save_session_context(session_context)

    return session_context
```

---

## Session Summary Generation

### Trigger Conditions

Generate session summary when user selects:
- "🔄 새 세션 시작" (New Session)
- "✅ 세션 완료" (Session Complete)

### Summary Template

**Markdown Format**:
```markdown
## 🎊 세션 요약

### 완료된 작업
- ✅ [Task 1 content]
- ✅ [Task 2 content]
- ✅ [Task 3 content]

### Git 통계
- 📝 생성된 커밋: [N]개
- 📂 변경된 파일: [N]개
- ➕ 추가된 라인: +[N]
- ➖ 삭제된 라인: -[N]

### 생성된 문서
- 📋 SPEC: `.moai/specs/[SPEC-ID]/`
- 📄 구현 파일: `[file paths]`
- 🧪 테스트 파일: `[test file paths]`

### 다음 권장 작업
1. [Recommendation 1]
2. [Recommendation 2]
3. [Recommendation 3]
```

**Implementation**:
```python
def generate_session_summary():
    """Generate comprehensive session summary"""
    # Get completed tasks from session context
    context = load_session_context()

    # Git statistics
    git_stats = get_git_statistics()

    # Build markdown report
    summary = f"""
## 🎊 세션 요약

### 완료된 작업
{format_completed_tasks(context['completed_tasks'])}

### Git 통계
- 📝 생성된 커밋: {git_stats['commit_count']}개
- 📂 변경된 파일: {git_stats['files_changed']}개
- ➕ 추가된 라인: +{git_stats['insertions']}
- ➖ 삭제된 라인: -{git_stats['deletions']}

### 생성된 문서
{format_created_documents(context)}

### 다음 권장 작업
{generate_recommendations(context)}
"""
    return summary

def get_git_statistics():
    """Get Git stats since session start"""
    # Run: git log --shortstat --since="[session_start_time]"
    import subprocess
    result = subprocess.run(
        ['git', 'log', '--shortstat', '--oneline'],
        capture_output=True,
        text=True
    )
    # Parse output and return stats
    return parse_git_shortstat(result.stdout)
```

---

## Error Handling & Fallbacks

### Scenario 1: AskUserQuestion Fails

**Error**: Tool invocation timeout or error

**Fallback Strategy**:
```python
try:
    AskUserQuestion(questions=[...])
except ToolInvocationError as e:
    # Log error
    logger.error(f"AskUserQuestion failed: {e}")

    # Output fallback message
    print("""
⚠️ 다음 단계 선택을 불러올 수 없습니다.

수동으로 다음 커맨드를 실행하세요:
- `/alfred:1-plan` - 스펙 작성 시작
- `/clear` - 새 세션 시작

에러 로그: .moai/logs/session-error-{timestamp}.log
""")
```

---

### Scenario 2: TodoWrite Incomplete Tasks

**Error**: Not all tasks marked as `completed` when command tries to finish

**Validation Check**:
```python
def validate_todowrite_before_completion():
    """Ensure all tasks are completed"""
    pending_tasks = [
        task for task in todowrite.get_all()
        if task.status in ["pending", "in_progress"]
    ]

    if pending_tasks:
        raise IncompleteTasksError(
            f"Cannot complete command with {len(pending_tasks)} "
            f"incomplete tasks: {[t.content for t in pending_tasks]}"
        )
```

**Error Message**:
```markdown
⚠️ 미완료 작업이 있습니다:
- [ ] [pending/in_progress]: [Task content]

먼저 모든 작업을 완료하거나 취소하세요.
```

---

### Scenario 3: Invalid User Selection

**Error**: User provides invalid option number

**Handling**:
```python
def handle_user_selection(user_input, options):
    """Validate and execute user selection"""
    try:
        selection = int(user_input)
        if 1 <= selection <= len(options):
            return execute_option(options[selection - 1])
        else:
            raise ValueError("Out of range")
    except (ValueError, TypeError):
        print("""
❌ 잘못된 선택입니다. 1-{} 중에서 선택해주세요.
""".format(len(options)))
        # Re-invoke AskUserQuestion
        return ask_user_again(options)
```

---

## Testing & Validation

### Unit Tests

**Test 1: AskUserQuestion Structure Validation**
```python
def test_askuserquestion_structure():
    """Verify all commands have valid AskUserQuestion calls"""
    command_files = [
        "0-project.md",
        "1-plan.md",
        "2-run.md",
        "3-sync.md"
    ]

    for cmd_file in command_files:
        content = read_file(f".claude/commands/alfred/{cmd_file}")

        # Check AskUserQuestion exists
        assert "AskUserQuestion" in content

        # Parse Python code block
        code_block = extract_python_code(content)

        # Validate structure
        assert "questions" in code_block
        assert "options" in code_block
        assert len(code_block["options"]) in [3, 4]
```

**Test 2: Prose Pattern Detection**
```python
def test_no_prose_suggestions():
    """Ensure no prose next-step suggestions exist"""
    prohibited_patterns = [
        r"You can now run",
        r"Next, you should",
        r"To continue, please run",
        r"다음으로 .* 실행하세요"
    ]

    command_files = glob.glob(".claude/commands/alfred/*.md")

    for cmd_file in command_files:
        content = read_file(cmd_file)
        for pattern in prohibited_patterns:
            matches = re.findall(pattern, content, re.IGNORECASE)
            assert len(matches) == 0, f"Found prose pattern in {cmd_file}: {matches}"
```

**Test 3: TodoWrite Cleanup Verification**
```python
def test_todowrite_cleanup():
    """Verify TodoWrite is cleaned up before AskUserQuestion"""
    # Mock TodoWrite with completed tasks
    todowrite.add_task("Task 1", status="completed")
    todowrite.add_task("Task 2", status="completed")

    # Trigger cleanup
    context = cleanup_todowrite_before_completion()

    # Verify all tasks captured
    assert len(context["completed_tasks"]) == 2
    assert all(t.status == "completed" for t in context["completed_tasks"])
```

---

### Integration Tests

**Test Scenario 1**: Full workflow execution
```python
def test_full_workflow_with_askuserquestion():
    """Test complete workflow: 0-project → 1-plan → 2-run → 3-sync"""

    # Step 1: /alfred:0-project
    run_command("/alfred:0-project")
    assert_askuserquestion_called()
    simulate_user_selection(option=1)  # Select "📋 스펙 작성 진행"

    # Step 2: /alfred:1-plan
    run_command("/alfred:1-plan")
    assert_askuserquestion_called()
    simulate_user_selection(option=1)  # Select "🚀 구현 진행"

    # Step 3: /alfred:2-run SPEC-XXX-001
    run_command(f"/alfred:2-run {spec_id}")
    assert_askuserquestion_called()
    simulate_user_selection(option=1)  # Select "📚 문서 동기화"

    # Step 4: /alfred:3-sync
    run_command("/alfred:3-sync")
    assert_askuserquestion_called()
    simulate_user_selection(option=3)  # Select "✅ 세션 완료"

    # Verify session summary generated
    assert_session_summary_exists()
```

---

## Language Localization Guide

### Supported Languages

| Language | Code | conversation_language Value |
|----------|------|----------------------------|
| English  | en   | `en`                       |
| Korean   | ko   | `ko`                       |
| Japanese | ja   | `ja`                       |
| Chinese  | zh   | `zh`                       |
| Spanish  | es   | `es`                       |

### Translation Mapping

**Command**: `/alfred:0-project`

| Language | Question Text | Header | Option 1 | Option 2 | Option 3 |
|----------|--------------|--------|----------|----------|----------|
| English  | Project initialization completed. What would you like to do next? | Next Steps | 📋 Start Planning | 🔍 Review Structure | 🔄 New Session |
| Korean   | 프로젝트 초기화가 완료되었습니다. 다음으로 뭘 하시겠습니까? | 다음 단계 | 📋 스펙 작성 진행 | 🔍 프로젝트 구조 검토 | 🔄 새 세션 시작 |
| Japanese | プロジェクトの初期化が完了しました。次に何をしますか？ | 次のステップ | 📋 仕様書作成 | 🔍 構造確認 | 🔄 新しいセッション |

**Command**: `/alfred:1-plan`

| Language | Question Text | Option 1 | Option 2 | Option 3 |
|----------|--------------|----------|----------|----------|
| English  | SPEC creation completed ({spec_id}). What's next? | 🚀 Start Implementation | ✏️ Revise SPEC | 🔄 New Session |
| Korean   | SPEC 작성이 완료되었습니다 ({spec_id}). 다음으로 뭘 하시겠습니까? | 🚀 구현 진행 | ✏️ SPEC 수정 | 🔄 새 세션 시작 |
| Japanese | SPEC作成が完了しました ({spec_id})。次は？ | 🚀 実装開始 | ✏️ SPEC修正 | 🔄 新しいセッション |

### Dynamic Language Loading

**Implementation**:
```python
def get_localized_text(key, lang=None):
    """Get localized text from language files"""
    if lang is None:
        lang = get_conversation_language()  # From .moai/config.json

    translations = load_translations(f".moai/i18n/{lang}.json")
    return translations.get(key, translations["en"][key])  # Fallback to English

# Example usage:
AskUserQuestion(
    questions=[
        {
            "question": get_localized_text("alfred_0_project.question"),
            "header": get_localized_text("common.next_steps"),
            "options": [
                {
                    "label": get_localized_text("alfred_0_project.option1.label"),
                    "description": get_localized_text("alfred_0_project.option1.description")
                },
                # ... more options
            ]
        }
    ]
)
```

---

## Checklist for Implementers

### Pre-Implementation
- [ ] Read this guide completely
- [ ] Understand AskUserQuestion tool API
- [ ] Review user's `conversation_language` setting
- [ ] Identify command completion point (where to insert pattern)

### During Implementation
- [ ] Add AskUserQuestion call at command end
- [ ] Provide exactly 3-4 options
- [ ] Use user's language for all text
- [ ] Add emojis to option labels
- [ ] Implement TodoWrite cleanup logic
- [ ] Add session summary generation (for "세션 완료" option)

### Post-Implementation
- [ ] Run unit tests (structure validation, prose detection)
- [ ] Run integration tests (full workflow)
- [ ] Test with different `conversation_language` values
- [ ] Verify no prose suggestions remain
- [ ] Test error handling scenarios
- [ ] Update CLAUDE.md if pattern changed

---

## Common Pitfalls & Solutions

### Pitfall 1: Hardcoded English Text

**Problem**:
```python
"question": "Project initialization completed. What's next?"  # ❌ Hardcoded
```

**Solution**:
```python
question_text = get_localized_text("alfred_0_project.question")  # ✅ Localized
```

---

### Pitfall 2: Prose Suggestions Remaining

**Problem**:
```markdown
프로젝트가 준비되었습니다. 이제 `/alfred:1-plan`을 실행하세요.  # ❌ Prose
```

**Solution**:
```python
# Use AskUserQuestion instead
AskUserQuestion(questions=[...])  # ✅ Structured interaction
```

---

### Pitfall 3: Too Many Options

**Problem**:
```python
"options": [
    {"label": "Option 1", ...},
    {"label": "Option 2", ...},
    {"label": "Option 3", ...},
    {"label": "Option 4", ...},
    {"label": "Option 5", ...},  # ❌ Too many (5 > 4)
]
```

**Solution**: Limit to 3-4 options. Merge similar options if needed.

---

### Pitfall 4: Missing TodoWrite Cleanup

**Problem**: Forgetting to extract completed tasks before AskUserQuestion

**Solution**:
```python
# ALWAYS cleanup TodoWrite first
session_context = cleanup_todowrite_before_completion()

# THEN call AskUserQuestion
AskUserQuestion(questions=[...])
```

---

## Reference Links

**Related Documentation**:
- CLAUDE.md: "⚡ Alfred Command Completion Pattern" section
- SPEC-SESSION-CLEANUP-001: spec.md, plan.md, acceptance.md
- .moai/memory/language-config-schema.md: Language configuration reference

**Tool Documentation**:
- `moai-alfred-ask-user-questions` skill: AskUserQuestion API
- TodoWrite tool: Task tracking API

**Command Files** (implementation targets):
- `src/moai_adk/templates/.claude/commands/alfred/0-project.md`
- `src/moai_adk/templates/.claude/commands/alfred/1-plan.md`
- `src/moai_adk/templates/.claude/commands/alfred/2-run.md`
- `src/moai_adk/templates/.claude/commands/alfred/3-sync.md`

---

## Change Log

| Version | Date       | Changes                                    |
|---------|------------|--------------------------------------------|
| 1.0.0   | 2025-10-30 | Initial guide creation for SPEC-SESSION-CLEANUP-001 |

---

**Generated by**: tdd-implementer (SPEC-SESSION-CLEANUP-001)
**Language**: English (technical documentation standard)
**Target Audience**: MoAI-ADK developers implementing Alfred commands

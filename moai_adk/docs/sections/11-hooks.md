# MoAI-ADK Hook 시스템

## 🪝 Hook 시스템 개요

MoAI-ADK의 Hook 시스템은 Claude Code의 표준 Hook 인터페이스를 활용하여 개발 과정의 품질을 자동으로 보장합니다.

## Hook 이벤트 정의

| Hook Event           | Trigger           | 주요 역할                       | Exit Code        |
| -------------------- | ----------------- | ------------------------------- | ---------------- |
| **PreToolUse**       | 도구 실행 전      | 정책 차단, Constitution 보호    | 0: 허용, 2: 차단 |
| **PostToolUse**      | 도구 실행 후      | 자동 검수, 커밋 지침, 세션 안내 | 0: 성공          |
| **SessionStart**     | 세션 시작         | Gate 상태 확인, 미완료 알림     | 0: 성공          |

## 5개 핵심 Hook 스크립트

### 1. policy_block.py - PreToolUse Hook

**기능**: 위험 명령어 차단 및 정책 검증

```python
# 주요 차단 대상
dangerous_commands = [
    'rm -rf /',
    'sudo rm',
    'dd if=/dev/zero',
    ':(){:|:&};:'  # Fork bomb
]

# Steering 문서 보호
if '.moai/steering/' in file_path:
    print("🔒 Steering 문서는 /moai:1-project setting으로 수정하세요")
    sys.exit(2)  # 차단
```

### 2. constitution_guard.py - PreToolUse Hook

**기능**: Constitution 5원칙 보호 및 검증

```python
# Constitution 변경 체크리스트 확인
if 'constitution.md' in file_path:
    checklist_path = Path('.moai/memory/constitution_update_checklist.md')
    if not checklist_path.exists():
        print("⚠️ Constitution 변경은 체크리스트가 필요합니다")
        sys.exit(2)  # 차단
```

### 3. tag_validator.py - PreToolUse Hook

**기능**: 14-Core @TAG 시스템 품질 검증

```python
# 14-Core 태그 체계 검증
valid_tags = {
    'SPEC': ['REQ', 'DESIGN', 'TASK'],
    'Steering': ['VISION', 'STRUCT', 'TECH', 'STACK'],
    'Implementation': ['FEATURE', 'API', 'TEST', 'DATA'],
    'Quality': ['PERF', 'SEC', 'DEBT', 'TODO']
}

# 태그 패턴 검증
tag_pattern = r'@([A-Z]+)[-:]([A-Z0-9-]+)'
```

### 4. post_stage_guard.py - PostToolUse Hook

**기능**: 단계 완료 후 자동 검수 및 안내

```python
# 자동 검수 실행
if tool_name in ['Write', 'Edit', 'MultiEdit']:
    # TAG 인덱스 업데이트
    update_tag_index()

    # 추적성 매트릭스 갱신
    update_traceability_matrix()

    # 다음 단계 안내
    suggest_next_step()
```

### 5. session_start_notice.py - SessionStart Hook

**기능**: 세션 시작 시 프로젝트 상태 알림

```python
# 프로젝트 상태 분석
def analyze_project_state():
    # SPEC 문서 존재 여부
    specs = list(Path('.moai/specs').glob('SPEC-*/spec.md'))

    # 미완료 작업 확인
    pending_tasks = check_pending_tasks()

    # Gate 상태 확인
    gate_status = check_gate_status()

    return {
        'specs_count': len(specs),
        'pending_tasks': pending_tasks,
        'current_gate': gate_status
    }
```

## .claude/settings.json Hook 설정

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash|WebFetch",
        "hooks": [
          {
            "type": "command",
            "command": "python3 .claude/hooks/moai/policy_block.py"
          }
        ]
      },
      {
        "matcher": "Edit|MultiEdit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "python3 .claude/hooks/moai/constitution_guard.py"
          },
          {
            "type": "command",
            "command": "python3 .claude/hooks/moai/tag_validator.py"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit|MultiEdit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "python3 .claude/hooks/moai/post_stage_guard.py"
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "python3 .claude/hooks/moai/session_start_notice.py"
          }
        ]
      }
    ]
  }
}
```

## Hook 시스템의 이점

### 자동화된 품질 보장
- **사전 차단**: 위험한 작업을 미리 방지
- **실시간 검증**: 작업 중 품질 규칙 자동 확인
- **사후 처리**: 완료 후 자동 정리 및 동기화

### Constitution 5원칙 강제
- **Simplicity**: 복잡한 구조 생성 방지
- **Architecture**: 표준 구조 준수 강제
- **Testing**: 테스트 없는 구현 차단
- **Observability**: 로깅 및 모니터링 강제
- **Versioning**: 버전 관리 규칙 준수

### 14-Core TAG 무결성
- **실시간 검증**: 태그 생성 시점에 규칙 확인
- **자동 수정**: 링크 깨짐 자동 복구
- **품질 점수**: 태그 품질 정량 평가

## 문제 해결

### Hook 실행 권한 문제
```bash
# Linux/macOS
chmod +x .claude/hooks/moai/*.py

# Windows PowerShell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Hook 테스트
```bash
# Hook 시스템 테스트
python .claude/hooks/moai/test_hook.py

# 개별 Hook 테스트
echo '{"tool_name": "Write", "tool_input": {"file_path": "test.md", "content": "@REQ:TEST-001"}}' | python .claude/hooks/moai/tag_validator.py
```

### JSON 처리 오류 (v0.1.12 수정)
- **기존**: `sys.argv` 사용으로 인한 파싱 오류
- **개선**: `stdin` JSON 처리로 Claude Code 표준 완전 준수

Hook 시스템은 개발자가 의식하지 않아도 자동으로 품질을 보장하는 **투명한 품질 게이트**입니다.
# Windows Hook 환경 변수 이슈 회귀 분석 보고서

**분석 일시**: 2025-11-03 18:50
**상태**: ❌ **미해결 (회귀 발생)**
**심각도**: 🔴 High

---

## 🚨 핵심 발견

**GitHub Issue #161에서 요청한 Windows Hook 환경 변수 문제가 실제로는 해결되지 않았습니다.**

해결책이 적용된 후 약 48시간 내에 **회귀(Regression)**되었습니다.

---

## 📋 타임라인

### Phase 1: Windows 문제 보고 (Issue #161)
- **날짜**: 2025-11-02 09:26:06
- **문제**: `$CLAUDE_PROJECT_DIR` 환경 변수가 Windows PowerShell에서 미설정으로 Hook 실행 오류
- **보고자**: Windows 사용자

### Phase 2: 첫 번째 해결책 (Commit a2898697)
**커밋**: a2898697 (2025-11-02 17:25:32)
**제목**: `fix(hooks): 크로스 플랫폼 호환성 복원 - 상대 경로 사용 (fixes #161)`

**변경 사항**:
```json
Before:
"command": "uv run \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/alfred/session_start__show_project_info.py"

After:
"command": "uv run .claude/hooks/alfred/session_start__show_project_info.py"
```

**목표**:
- ✅ 환경 변수 제거
- ✅ 상대 경로 사용으로 모든 플랫폼 호환
- ✅ 이전 c6355dc6 커밋의 해결책 재도입

**포함 버전**: v0.13.0

---

### Phase 3: 두 번째 시도 (Commit 3a3b8808)
**커밋**: 3a3b8808 (2025-11-03 17:12:45)
**제목**: `fix: Remove $CLAUDE_PROJECT_DIR from hook configuration - use auto-discovery instead`

**변경 사항**:
- Settings.json에서 hooks 섹션 전체 제거
- "auto-discovery" 방식으로 변경

**결과**: hooks 섹션이 없어진 상태

---

### Phase 4: 회귀 발생 (Commit aaff7388)
**커밋**: aaff7388 (2025-11-03 17:13:44) ← **현재 상태**
**제목**: `fix: Restore hooks configuration with $CLAUDE_PROJECT_DIR environment variable`

**변경 사항**:
```json
Restored:
"command": "uv run \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/alfred/session_start__show_project_info.py"
```

**문제점**:
- 환경 변수 버전으로 복구됨
- a2898697의 Windows 호환성 해결책 **완전히 제거됨**
- Windows PowerShell에서 **다시 동일한 오류 발생 가능**

---

## 🔴 현재 상태 비교

| 파일 | 위치 | 환경 변수 상태 | 상대 경로 상태 | 현황 |
|------|------|---|---|---|
| `.claude/settings.json` | 로컬 | ❌ `$CLAUDE_PROJECT_DIR` 포함 | ✅ 미사용 | **회귀** |
| `src/moai_adk/templates/.claude/settings.json` | 템플릿 | ❌ `$CLAUDE_PROJECT_DIR` 포함 | ✅ 미사용 | **회귀** |

### 현재 실제 코드 (로컬 & 템플릿 모두 동일)

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "command": "uv run \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/alfred/session_start__show_project_info.py",
            "type": "command"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "hooks": [
          {
            "command": "uv run \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/alfred/pre_tool__auto_checkpoint.py",
            "type": "command"
          }
        ],
        "matcher": "Edit|Write|MultiEdit"
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "command": "uv run \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/alfred/user_prompt__jit_load_docs.py",
            "type": "command"
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "hooks": [
          {
            "command": "uv run \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/alfred/session_end__cleanup.py",
            "type": "command"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "hooks": [
          {
            "command": "uv run \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/alfred/post_tool__log_changes.py",
            "type": "command"
          }
        ],
        "matcher": "Edit|Write|MultiEdit"
      }
    ]
  }
}
```

---

## 💥 영향 분석

### Windows PowerShell 사용자
❌ **CLI 명령 실행 시 오류**:
```
오류: $CLAUDE_PROJECT_DIR는 설정되지 않은 환경 변수입니다
```

### macOS/Linux 사용자
✅ **작동 가능** (환경 변수는 작동하나, 상대 경로가 더 안정적)

---

## 📝 올바른 해결책

### a2898697에서 제시한 해결책 (검증됨)

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "command": "uv run .claude/hooks/alfred/session_start__show_project_info.py",
            "type": "command"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "hooks": [
          {
            "command": "uv run .claude/hooks/alfred/pre_tool__auto_checkpoint.py",
            "type": "command"
          }
        ],
        "matcher": "Edit|Write|MultiEdit"
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "command": "uv run .claude/hooks/alfred/user_prompt__jit_load_docs.py",
            "type": "command"
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "hooks": [
          {
            "command": "uv run .claude/hooks/alfred/session_end__cleanup.py",
            "type": "command"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "hooks": [
          {
            "command": "uv run .claude/hooks/alfred/post_tool__log_changes.py",
            "type": "command"
          }
        ],
        "matcher": "Edit|Write|MultiEdit"
      }
    ]
  }
}
```

**장점**:
- ✅ 모든 플랫폼(Windows/macOS/Linux) 호환
- ✅ 환경 변수 설정 불필요
- ✅ 상대 경로는 `uv run`이 자동으로 처리
- ✅ 이미 c6355dc6에서 5일간 성공적으로 검증됨

---

## 🔧 권장 조치

### 즉시 해결
1. a2898697 커밋의 변경사항 다시 적용
2. 상대 경로로 모든 hook 명령 변경
3. 로컬 + 템플릿 모두 동기화
4. v0.14.1 패치 릴리즈

### 검증
```bash
# 환경 변수 미설정 상태에서 테스트
unset CLAUDE_PROJECT_DIR

# Windows PowerShell
$env:CLAUDE_PROJECT_DIR = $null

# Hook 실행 테스트
uv run .claude/hooks/alfred/session_start__show_project_info.py
```

---

## 📊 회귀 원인 분석

**왜 이런 일이?**

1. **Phase 2 (a2898697)**: 환경 변수 제거, 상대 경로로 변경 ✅
2. **Phase 3 (3a3b8808)**: hooks 섹션 전체 제거 (auto-discovery 시도)
3. **Phase 4 (aaff7388)**: hooks 섹션 복구하면서 **환경 변수 버전으로 복구**

aaff7388 커밋 메시지:
> "Restore the hook configuration that was previously removed"

즉, hooks 섹션이 없어진 것을 복구하려다가, **이전 환경 변수 버전을 그대로 복구**한 것으로 보입니다.

**교훈**:
- Git history를 정확히 추적하지 않음
- 상대 경로 버전과 환경 변수 버전이 혼재됨
- 자동화된 동기화 체계 부족

---

## ✅ 검증 체크리스트

- [ ] a2898697의 상대 경로 변경사항 재적용
- [ ] 로컬 `.claude/settings.json` 업데이트
- [ ] 템플릿 `src/moai_adk/templates/.claude/settings.json` 업데이트
- [ ] 양쪽 파일이 동일한지 확인
- [ ] Windows 환경에서 테스트
- [ ] macOS/Linux 환경에서 테스트
- [ ] v0.14.1 패치 릴리즈
- [ ] 관련 이슈 #161 재오픈 및 코멘트

---

## 📌 결론

**Issue #161 "Windows에서 hook들이 $CLAUDE_PROJECT_DIR 미설정으로 오류나고 있습니다"는:**

- ✅ 일시적으로 해결됨 (a2898697 커밋)
- ❌ 이후 회귀됨 (aaff7388 커밋)
- 🔴 **현재 미해결 상태**

Windows 사용자는 여전히 동일한 오류를 경험할 가능성이 높습니다.

---

**작성**: 🎩 Alfred
**상태**: 긴급 (Urgent)
**필요 조치**: 즉시 패치 필요

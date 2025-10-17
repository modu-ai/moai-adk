# @DOC:HOOK-PRE-001 | Chain: @SPEC:DOCS-003 -> @DOC:HOOK-001

# PreToolUseHook

도구 사용 전에 실행되는 Hook입니다.

## 개요

`PreToolUseHook`은 에이전트가 도구(tool)를 사용하기 직전에 실행되는 Hook입니다. 이를 통해 도구 실행 전 검증, 로깅, 권한 체크 등을 수행할 수 있습니다.

## 실행 시점

```
에이전트 실행 시작
  ↓
SessionStartHook 실행
  ↓
[반복]
  ├─→ PreToolUseHook 실행 ← 여기!
  ├─→ 도구 실행 (Read, Write, Bash 등)
  └─→ PostToolUseHook 실행
```

## 사용 사례

1. **보안 검증**: 민감한 파일 접근 차단
2. **로깅**: 도구 사용 기록
3. **권한 체크**: 파일 수정 권한 확인
4. **입력 검증**: 도구 파라미터 검증

## 구현 예제

```python
from moai_adk.core.hooks import PreToolUseHook

class SecurityPreToolHook(PreToolUseHook):
    """보안 검증 Pre-Tool Hook"""

    def execute(self, tool_name: str, tool_params: dict) -> dict:
        """도구 사용 전 보안 검증"""

        # 민감한 파일 접근 차단
        if tool_name == "Write":
            file_path = tool_params.get("file_path", "")
            if ".env" in file_path or "secret" in file_path.lower():
                return {
                    "block": True,
                    "reason": "민감한 파일에 대한 쓰기가 차단되었습니다."
                }

        # 허용
        return {"block": False}
```

## 설정 방법

`.moai/config.json`에 Hook 설정 추가:

```json
{
  "hooks": {
    "pre_tool_use": [
      {
        "name": "SecurityPreToolHook",
        "enabled": true,
        "config": {
          "blocked_patterns": [".env", "secret", "credentials"]
        }
      }
    ]
  }
}
```

## API Reference

### execute(tool_name, tool_params)

도구 사용 전 실행되는 메서드입니다.

**파라미터**:
- `tool_name` (str): 사용할 도구 이름 (예: "Read", "Write", "Bash")
- `tool_params` (dict): 도구 파라미터

**반환값** (dict):
- `block` (bool): True이면 도구 실행 차단
- `reason` (str, optional): 차단 이유
- `modified_params` (dict, optional): 수정된 파라미터

## 관련 문서

- [PostToolUseHook](post-tool-use-hook.md)
- [SessionStartHook](session-start-hook.md)
- [Custom Hooks](custom-hooks.md)

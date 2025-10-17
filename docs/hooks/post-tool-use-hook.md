# @DOC:HOOK-POST-001 | Chain: @SPEC:DOCS-003 -> @DOC:HOOK-001

# PostToolUseHook

도구 사용 후에 실행되는 Hook입니다.

## 개요

`PostToolUseHook`은 에이전트가 도구(tool)를 사용한 직후에 실행되는 Hook입니다. 이를 통해 결과 검증, 로깅, 자동 커밋, 알림 등을 수행할 수 있습니다.

## 실행 시점

```
에이전트 실행 시작
  ↓
SessionStartHook 실행
  ↓
[반복]
  ├─→ PreToolUseHook 실행
  ├─→ 도구 실행 (Read, Write, Bash 등)
  └─→ PostToolUseHook 실행 ← 여기!
```

## 사용 사례

1. **자동 커밋**: 파일 수정 후 Git 커밋
2. **결과 검증**: 도구 실행 결과 검증
3. **로깅**: 실행 결과 기록
4. **알림**: 중요 작업 완료 알림
5. **백업**: 파일 수정 후 자동 백업

## 구현 예제

```python
from moai_adk.core.hooks import PostToolUseHook

class AutoCommitPostToolHook(PostToolUseHook):
    """파일 수정 후 자동 커밋 Hook"""

    def execute(self, tool_name: str, tool_params: dict, result: dict) -> None:
        """도구 사용 후 자동 커밋"""

        # Write 도구 사용 후 자동 커밋
        if tool_name == "Write" and result.get("success"):
            file_path = tool_params.get("file_path", "")

            # Git 커밋
            import subprocess
            subprocess.run([
                "git", "add", file_path
            ])
            subprocess.run([
                "git", "commit", "-m",
                f"Auto-commit: Updated {file_path}"
            ])

            print(f"✅ Auto-committed: {file_path}")
```

## 설정 방법

`.moai/config.json`에 Hook 설정 추가:

```json
{
  "hooks": {
    "post_tool_use": [
      {
        "name": "AutoCommitPostToolHook",
        "enabled": true,
        "config": {
          "auto_commit": true,
          "commit_message_template": "Auto-commit: {action} {file}"
        }
      }
    ]
  }
}
```

## API Reference

### execute(tool_name, tool_params, result)

도구 사용 후 실행되는 메서드입니다.

**파라미터**:
- `tool_name` (str): 사용한 도구 이름 (예: "Read", "Write", "Bash")
- `tool_params` (dict): 도구 파라미터
- `result` (dict): 도구 실행 결과
  - `success` (bool): 성공 여부
  - `output` (any): 실행 결과
  - `error` (str, optional): 에러 메시지

**반환값**: None (부수 효과만 발생)

## 실전 예제: Git Checkpoint Hook

TDD 단계별 자동 체크포인트 생성:

```python
from moai_adk.core.hooks import PostToolUseHook

class GitCheckpointHook(PostToolUseHook):
    """TDD 단계별 Git 체크포인트 생성"""

    def execute(self, tool_name: str, tool_params: dict, result: dict) -> None:
        """TDD 단계 감지 후 체크포인트 생성"""

        if tool_name != "Write" or not result.get("success"):
            return

        file_path = tool_params.get("file_path", "")

        # 테스트 파일 작성 후 -> RED 체크포인트
        if "test_" in file_path:
            self._create_checkpoint("tdd-red", f"RED: {file_path}")

        # 구현 파일 작성 후 -> GREEN 체크포인트
        elif file_path.endswith(".py") and "test" not in file_path:
            self._create_checkpoint("tdd-green", f"GREEN: {file_path}")

    def _create_checkpoint(self, type: str, message: str):
        """체크포인트 생성"""
        # git-manager 호출
        pass
```

## 관련 문서

- [PreToolUseHook](pre-tool-use-hook.md)
- [SessionStartHook](session-start-hook.md)
- [Custom Hooks](custom-hooks.md)
- [Git Strategy](../configuration/personal-vs-team.md)

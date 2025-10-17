# @DOC:HOOK-SESSION-001 | Chain: @SPEC:DOCS-003 -> @DOC:HOOK-001

# SessionStartHook

세션 시작 시 실행되는 Hook입니다.

## 사용 예시

```python
from moai_adk.hooks import SessionStartHook

class MySessionStartHook(SessionStartHook):
    def execute(self, context):
        print("세션 시작!")
```

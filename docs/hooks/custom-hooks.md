# @DOC:HOOK-CUSTOM-001 | Chain: @SPEC:DOCS-003 -> @DOC:HOOK-001

# Custom Hooks

커스텀 Hook 작성 가이드입니다.

## 기본 구조

```python
from moai_adk.hooks import BaseHook

class CustomHook(BaseHook):
    def execute(self, context):
        # 커스텀 로직
        pass
```

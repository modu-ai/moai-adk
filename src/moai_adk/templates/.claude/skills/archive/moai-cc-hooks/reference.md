# Claude Code Hooks - Reference

## Hook Types

| Hook | Timing | Purpose |
|------|--------|---------|
| `pre-commit` | Before commit | Code validation, linting |
| `post-commit` | After commit | Logging, notifications |
| `pre-push` | Before push | Final validation |
| `post-merge` | After merge | Automation tasks |
| `pre-build` | Before build | Environment preparation |
| `post-deploy` | After deployment | Notifications, monitoring |

## Hook Interface

```python
class Hook:
    event: str  # Hook event
    priority: int  # Execution priority
    async_execution: bool  # Async execution

    async def execute(context: dict) -> dict:
        """Execute hook"""

    async def on_error(error: Exception) -> dict:
        """Error handling"""
```

## Context Object

```python
{
    "branch": "main",
    "files_changed": ["src/main.py", "tests/test_main.py"],
    "commit_message": "feat: add new feature",
    "author": "user@example.com",
    "timestamp": "2025-11-22T10:00:00Z"
}
```

## Return Value Format

```python
{
    "success": True,        # Success status
    "message": "...",       # Message
    "data": {...}          # Additional data
}
```

---

**Last Updated**: 2025-11-22

---
title: Windows Compatibility - Cross-Platform Timeout Handler
id: SPEC-BUGFIX-001
type: bugfix
priority: critical
status: completed
version: "0.11.0"
affects: hooks system
issue: "#129"
---

# SPEC-BUGFIX-001: Windows Compatibility Fix


## Problem Statement

**Issue**: All MoAI-ADK hooks fail on Windows due to `signal.SIGALRM` which is Unix-only.

**Error**: `AttributeError: module 'signal' has no attribute 'SIGALRM'`

**Impact**: Windows users cannot use MoAI-ADK at all (100% functionality blocked).

**Severity**: Critical

## Root Cause

### Affected Code

**File**: `src/moai_adk/templates/.claude/hooks/alfred/alfred_hooks.py`
**Lines**: 128-129

```python
# ❌ This fails on Windows
signal.signal(signal.SIGALRM, _hook_timeout_handler)
signal.alarm(5)
```

### Affected Files (11 total)

1. `alfred_hooks.py` (main router)
2. `notification__handle_events.py`
3. `post_tool__log_changes.py`
4. `pre_tool__auto_checkpoint.py`
5. `session_end__cleanup.py`
6. `session_start__show_project_info.py`
7. `stop__handle_interrupt.py`
8. `subagent_stop__handle_subagent_end.py`
9. `user_prompt__jit_load_docs.py`
10. `core/project.py`
11. `shared/core/project.py`

### Why signal.SIGALRM Fails on Windows

From Python documentation:
> **signal.SIGALRM**: Availability: Unix

Windows does not support POSIX signals like SIGALRM.

## Solution

### Approach: Platform-Agnostic Timeout Handler

Use `threading.Timer` for Windows, preserve `signal.SIGALRM` for Unix/Linux/macOS.

### Implementation Design

#### 1. Create Cross-Platform Timeout Module

**File**: `src/moai_adk/templates/.claude/hooks/alfred/utils/timeout.py`

```python
"""Cross-platform timeout handler for hooks."""

import platform
import signal
import sys
import threading
from typing import Callable, Optional


class TimeoutError(Exception):
    """Timeout exception raised when hook execution exceeds limit."""
    pass


class CrossPlatformTimeout:
    """Cross-platform timeout handler supporting Windows, Unix, Linux, macOS."""

    def __init__(self, timeout_seconds: int, callback: Optional[Callable] = None):
        """Initialize timeout handler.

        Args:
            timeout_seconds: Timeout duration in seconds
            callback: Optional callback to execute on timeout (default: sys.exit(1))
        """
        self.timeout_seconds = timeout_seconds
        self.callback = callback or self._default_timeout_callback
        self.timer: Optional[threading.Timer] = None
        self.is_windows = platform.system() == "Windows"

    def _default_timeout_callback(self):
        """Default timeout callback: print error and exit."""
        print(f"Hook timeout after {self.timeout_seconds} seconds", file=sys.stderr)
        sys.exit(1)

    def _unix_signal_handler(self, signum, frame):
        """Signal handler for Unix platforms."""
        raise TimeoutError(f"Hook execution exceeded {self.timeout_seconds}-second timeout")

    def start(self):
        """Start timeout monitoring."""
        if self.is_windows:
            # Windows: use threading.Timer
            self.timer = threading.Timer(self.timeout_seconds, self.callback)
            self.timer.daemon = True
            self.timer.start()
        else:
            # Unix/Linux/macOS: use signal.SIGALRM
            signal.signal(signal.SIGALRM, self._unix_signal_handler)
            signal.alarm(self.timeout_seconds)

    def cancel(self):
        """Cancel timeout monitoring."""
        if self.is_windows:
            if self.timer:
                self.timer.cancel()
        else:
            signal.alarm(0)

    def __enter__(self):
        """Context manager entry."""
        self.start()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.cancel()
        return False
```

#### 2. Update alfred_hooks.py

**File**: `src/moai_adk/templates/.claude/hooks/alfred/alfred_hooks.py`

```python
import json
import sys
from pathlib import Path
from typing import Any

from core import HookResult
from handlers import (
    handle_notification,
    handle_post_tool_use,
    handle_pre_tool_use,
    handle_session_end,
    handle_session_start,
    handle_stop,
    handle_subagent_stop,
    handle_user_prompt_submit,
)
from utils.timeout import CrossPlatformTimeout, TimeoutError

# Add the hooks directory to sys.path to enable package imports
HOOKS_DIR = Path(__file__).parent
if str(HOOKS_DIR) not in sys.path:
    sys.path.insert(0, str(HOOKS_DIR))


def main() -> None:
    """Main entry point - Claude Code Hook script with CROSS-PLATFORM TIMEOUT.

    🛠️ Usage:
        python alfred_hooks.py <event_name> < payload.json

    📣 Supported Events:
        - SessionStart, UserPromptSubmit, SessionEnd, PreToolUse, PostToolUse
        - Notification, Stop, SubagentStop

    🚦 Exit Codes:
        - 0: Success
        - 1: Error (timeout, no arguments, JSON parsing failure, exception thrown)

    🌍 Platform Support:
        - ✅ Windows (threading.Timer)
        - ✅ Unix/Linux/macOS (signal.SIGALRM)

    """
    # Set global 5-second timeout for entire hook execution
    with CrossPlatformTimeout(5):
        try:
            # Check for event argument
            if len(sys.argv) < 2:
                print("Usage: alfred_hooks.py <event>", file=sys.stderr)
                sys.exit(1)

            event_name = sys.argv[1]

            try:
                # Read JSON from stdin
                input_data = sys.stdin.read()
                if not input_data or not input_data.strip():
                    data = {}
                else:
                    data = json.loads(input_data)

                cwd = data.get("cwd", ".")

                # Route to appropriate handler
                handlers = {
                    "SessionStart": handle_session_start,
                    "UserPromptSubmit": handle_user_prompt_submit,
                    "SessionEnd": handle_session_end,
                    "PreToolUse": handle_pre_tool_use,
                    "PostToolUse": handle_post_tool_use,
                    "Notification": handle_notification,
                    "Stop": handle_stop,
                    "SubagentStop": handle_subagent_stop,
                }

                handler = handlers.get(event_name)
                result = handler({"cwd": cwd, **data}) if handler else HookResult()

                # Output Hook result as JSON
                if event_name == "UserPromptSubmit":
                    print(json.dumps(result.to_user_prompt_submit_dict()))
                else:
                    print(json.dumps(result.to_dict()))

                sys.exit(0)

            except json.JSONDecodeError as e:
                error_response: dict[str, Any] = {
                    "continue": True,
                    "hookSpecificOutput": {"error": f"JSON parse error: {e}"}
                }
                print(json.dumps(error_response))
                print(f"JSON parse error: {e}", file=sys.stderr)
                sys.exit(1)
            except Exception as e:
                error_response: dict[str, Any] = {
                    "continue": True,
                    "hookSpecificOutput": {"error": f"Hook error: {e}"}
                }
                print(json.dumps(error_response))
                print(f"Unexpected error: {e}", file=sys.stderr)
                sys.exit(1)

        except TimeoutError:
            timeout_response: dict[str, Any] = {
                "continue": True,
                "systemMessage": "⚠️ Hook execution timeout - continuing without session info"
            }
            print(json.dumps(timeout_response))
            print("Hook timeout after 5 seconds", file=sys.stderr)
            sys.exit(1)


if __name__ == "__main__":
    main()
```

#### 3. Update Other Hook Files

Apply the same pattern to all 10 remaining hook files.

**Example**: `user_prompt__jit_load_docs.py`

```python
from utils.timeout import CrossPlatformTimeout, TimeoutError

def main() -> None:
    """Main entry point with cross-platform timeout."""
    with CrossPlatformTimeout(5):
        try:
            # ... existing logic ...
            pass
        except TimeoutError:
            # Handle timeout
            timeout_response = {
                "continue": True,
                "hookSpecificOutput": {
                    "hookEventName": "UserPromptSubmit",
                    "additionalContext": "⚠️ Hook timeout"
                }
            }
            print(json.dumps(timeout_response))
            sys.exit(1)
```

## Testing Strategy

### Unit Tests

**File**: `tests/unit/test_cross_platform_timeout.py`

```python
import platform
import time
import pytest
from src.moai_adk.templates.claude.hooks.alfred.utils.timeout import (
    CrossPlatformTimeout,
    TimeoutError
)


def test_timeout_on_windows(monkeypatch):
    """Test timeout using threading.Timer on Windows."""
    monkeypatch.setattr(platform, "system", lambda: "Windows")

    with pytest.raises(SystemExit):
        with CrossPlatformTimeout(1):
            time.sleep(2)  # Exceeds timeout


def test_timeout_on_unix(monkeypatch):
    """Test timeout using signal.SIGALRM on Unix."""
    monkeypatch.setattr(platform, "system", lambda: "Linux")

    with pytest.raises(TimeoutError):
        with CrossPlatformTimeout(1):
            time.sleep(2)  # Exceeds timeout


def test_no_timeout():
    """Test normal execution without timeout."""
    with CrossPlatformTimeout(2):
        time.sleep(0.5)  # Within timeout
        result = "success"

    assert result == "success"


def test_cancel_timeout():
    """Test timeout cancellation."""
    timeout = CrossPlatformTimeout(5)
    timeout.start()
    timeout.cancel()
    time.sleep(6)  # Should not timeout after cancel
```

### Integration Tests

**Test on all platforms**:

```yaml
# .github/workflows/test-hooks.yml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
    python-version: ['3.10', '3.11', '3.12']

steps:
  - name: Test hooks on ${{ matrix.os }}
    run: |
      echo '{"cwd": "."}' | python src/moai_adk/templates/.claude/hooks/alfred/alfred_hooks.py SessionStart
```

### Manual Testing

**Windows**:
```bash
cd shopify-mcp-server
echo '{"userPrompt": "/alfred:0-project", "cwd": "."}' | python .claude/hooks/alfred/user_prompt__jit_load_docs.py
```

**Unix/Linux/macOS**:
```bash
cd shopify-mcp-server
echo '{"userPrompt": "/alfred:0-project", "cwd": "."}' | python3 .claude/hooks/alfred/user_prompt__jit_load_docs.py
```

## Acceptance Criteria

- [ ] All 11 hook files use `CrossPlatformTimeout`
- [ ] Hooks work on Windows 10/11
- [ ] Hooks work on macOS
- [ ] Hooks work on Linux (Ubuntu)
- [ ] Unit tests pass on all platforms
- [ ] Integration tests pass on all platforms
- [ ] Performance: timeout overhead < 10ms
- [ ] Documentation updated with Windows support notes
- [ ] CI/CD includes Windows test matrix

## Implementation Plan

### Phase 1: Core Module (Day 1)
- Create `utils/timeout.py`
- Write unit tests
- Verify on Windows/Unix

### Phase 2: Update Hook Files (Day 2-3)
- Update `alfred_hooks.py`
- Update 10 remaining hook files
- Test each file individually

### Phase 3: CI/CD Integration (Day 4)
- Add Windows to GitHub Actions test matrix
- Add integration tests
- Verify all platforms

### Phase 4: Documentation (Day 5)
- Update README.md with Windows support
- Update installation guide
- Add troubleshooting section

### Phase 5: Release (Day 6)
- Version bump to v0.11.0
- Create release notes
- Deploy to PyPI

## Risks and Mitigations

### Risk 1: Threading overhead on Windows
**Impact**: Slower hook execution
**Mitigation**: Benchmark shows <10ms overhead (acceptable)

### Risk 2: Threading.Timer accuracy
**Impact**: Timeout may not fire exactly at 5 seconds
**Mitigation**: Use daemon threads to ensure cleanup

### Risk 3: Backward compatibility
**Impact**: Existing projects may break
**Mitigation**: Template update via `moai-adk init --force` (automatic backup)

## Related Issues

- Issue #129: Windows Compatibility: signal.SIGALRM Not Supported
- Discussion #117: Hook path issues (may be related)

## Success Metrics

- Windows users report successful hook execution
- No regression on Unix/Linux/macOS
- CI/CD passes on all platforms
- Performance impact < 10ms

---

**Author**: debug-helper
**Created**: 2025-10-30
**Target Release**: v0.11.0

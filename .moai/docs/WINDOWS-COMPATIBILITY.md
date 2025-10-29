# Windows Compatibility Guide

## SessionStart Hook Issue (Issue #107)

### Problem

Windows 환경에서 SessionStart 훅 실행 시 시스템이 멈추는 문제가 발생합니다. 훅 자체는 정상적으로 실행되지만, Claude Code가 사용자 입력을 받지 못하는 상태가 됩니다.

### Solution (v0.7.1+)

MoAI-ADK는 Windows를 자동으로 감지하여 최소한의 세션 정보만 반환하도록 구현되었습니다. 이를 통해 Claude Code를 정상적으로 사용할 수 있습니다.

### What Changed

#### Windows
```
🚀 MoAI-ADK Session Started (Windows mode - limited info due to platform issue #107)
```

#### macOS/Linux
```
🚀 MoAI-ADK Session Started
   🗿 MoAI-ADK Ver: 0.7.1
   🐍 Language: python
   🌿 Branch: main (a1b2c3d)
   📝 Changes: 5
   🔨 Last: feat: Add Windows compatibility
   📋 SPEC Progress: 5/10 (50%)
```

### Impact

| 기능 | Windows | macOS/Linux |
|------|---------|-------------|
| **Claude Code 실행** | ✅ 정상 | ✅ 정상 |
| **세션 정보 표시** | ⚠️ 최소 정보 | ✅ 전체 정보 |
| **Git 정보** | ❌ 생략 | ✅ 표시 |
| **SPEC 진행률** | ❌ 생략 | ✅ 표시 |
| **체크포인트** | ❌ 생략 | ✅ 표시 |

---

## Verification

### 1. Update MoAI-ADK

```bash
uv tool upgrade moai-adk
```

### 2. Verify Platform Detection

```bash
python -c "from .claude.hooks.alfred.core.platform import get_platform; print(get_platform())"
```

Expected output: `windows`

### 3. Test Hook Execution

```bash
echo '{"cwd": ".", "phase": "compact"}' | uv run .claude/hooks/alfred/alfred_hooks.py SessionStart
```

Expected output:
```json
{"continue": true, "systemMessage": "🚀 MoAI-ADK Session Started (Windows mode - limited info due to platform issue #107)"}
```

훅이 즉시 응답을 반환하고 시스템이 멈추지 않아야 합니다.

---

## Technical Details

### Root Cause

이는 **MoAI-ADK 문제가 아니라 Claude Code의 Windows subprocess 관리 버그**입니다.

- **Upstream Issue**: https://github.com/anthropics/claude-code/issues/9542
- **Tracking Issue**: https://github.com/modu-ai/MoAI-ADK/issues/107

디버그 로그에서 훅은 정상적으로 실행되고 유효한 JSON을 반환하지만, Claude Code가 subprocess에서 제어권을 반환하지 못합니다:

```
[DEBUG] Successfully parsed and validated hook JSON output
{"continue":true,"systemMessage":"🚀 MoAI-ADK Session Started..."}
```

### Solution Architecture

#### 1. Platform Detection Layer

**File**: `.claude/hooks/alfred/core/platform.py`

```python
def get_platform() -> PlatformType:
    """Detect current OS: windows, darwin, linux, or unknown"""
    platform = sys.platform.lower()

    if platform.startswith("win") or platform in ["cygwin", "msys"]:
        return "windows"
    elif platform == "darwin":
        return "darwin"
    elif platform.startswith("linux"):
        return "linux"
    else:
        return "unknown"

def is_windows() -> bool:
    """Convenience function for Windows detection"""
    return get_platform() == "windows"
```

#### 2. Graceful Degradation

**File**: `.claude/hooks/alfred/handlers/session.py`

```python
def handle_session_start(payload: HookPayload) -> HookResult:
    # WORKAROUND: Windows subprocess hang (Issue #107)
    if is_windows():
        return HookResult(
            continue_execution=True,
            system_message="🚀 MoAI-ADK Session Started (Windows mode - limited info due to platform issue #107)"
        )

    # ... full implementation for Unix systems ...
```

#### 3. Comprehensive Testing

**File**: `.claude/hooks/alfred/test_platform.py`

- 12 unit tests covering all platforms
- Mock-based testing for cross-platform verification
- Integration tests for actual system detection

### Test Results

```
============================= test session starts ==============================
test_platform.py::TestPlatformDetection::test_get_platform_windows_win32 PASSED
test_platform.py::TestPlatformDetection::test_get_platform_windows_variants PASSED
test_platform.py::TestPlatformDetection::test_get_platform_macos PASSED
test_platform.py::TestPlatformDetection::test_get_platform_linux PASSED
test_platform.py::TestPlatformDetection::test_get_platform_unknown PASSED
test_platform.py::TestPlatformDetection::test_is_windows_true PASSED
test_platform.py::TestPlatformDetection::test_is_windows_false_on_unix PASSED
test_platform.py::TestPlatformDetection::test_platform_type_annotation PASSED
test_platform.py::TestPlatformDetection::test_case_insensitivity PASSED
test_platform.py::TestPlatformDetectionIntegration::test_actual_platform_detection PASSED
test_platform.py::TestPlatformDetectionIntegration::test_is_windows_returns_bool PASSED
test_platform.py::TestPlatformDetectionIntegration::test_platform_consistency PASSED

============================== 12 passed in 0.06s ==============================
```

---

## Monitoring Upstream Fix

이는 **임시 우회 방법**입니다. Claude Code가 Windows subprocess 문제를 수정하면, 플랫폼 감지 로직을 제거하고 Windows에서도 전체 기능을 복원할 예정입니다.

### Tracking

- **Upstream Issue**: https://github.com/anthropics/claude-code/issues/9542
- **MoAI-ADK Issue**: https://github.com/modu-ai/MoAI-ADK/issues/107

### Expected Timeline

- **Current**: Workaround active (v0.7.1+)
- **When fixed**: Remove `is_windows()` check in session.py
- **Result**: Full session status automatically restored on Windows

---

## Report Issues

v0.7.1 이후에도 Windows에서 멈춤 현상이 발생하는 경우:

1. **디버그 로그 활성화**:
   ```bash
   set CLAUDE_DEBUG=1
   ```

2. **로그 캡처**: 멈춤 발생 시점의 디버그 로그를 저장합니다.

3. **이슈 제보**: https://github.com/modu-ai/MoAI-ADK/issues/new
   - 디버그 로그 첨부
   - Windows 버전 명시 (Windows 10/11, build number)
   - Git Bash/PowerShell/CMD 등 사용 중인 셸 명시

---

## Related Tags

- CODE:WINDOWS-HOOK-COMPAT-001 - Platform detection implementation
- TEST:WINDOWS-HOOK-COMPAT-001 - Unit tests for platform detection

---

**Last Updated**: 2025-10-29
**Status**: Workaround Active
**Target Resolution**: Upstream fix in Claude Code
**Documentation Version**: v0.7.1

# GitHub 이슈 분석 보고서

**분석 날짜**: 2025-10-31
**분석 대상 저장소**: https://github.com/modu-ai/moai-adk
**분석 범위**: 모든 열린 이슈 (4개)

---

## 📊 요약

| 이슈 | 제목 | 심각도 | 상태 | 영향범위 |
|------|------|--------|------|---------|
| #155 | 구현계획이 없는데 승인을 자주 요청 | Medium | Open | UX |
| #154 | hook error이 계속됩니다 (v0.12.1) | **High** | Open | 전체 Hook 시스템 |
| #153 | SessionStart:compact hook error | **High** | Open | Hook 시스템 |
| #152 | backup 안내문구 불일치 | Medium | Open | 초기화 프로세스 |
| **Disc #117** | **Hook 파일 경로 오류** | **High** | Open | **Hook 배포/초기화** |

---

## 🔴 Critical Issues

### Issue #154 & #153: Hook ImportError (HIGH PRIORITY)

**이슈 번호**: #154, #153
**제목**: MoAI-ADK Hook 실행 오류
**심각도**: 🔴 **HIGH** (모든 Hook 기능 동작 불가)
**상태**: Open

#### 💥 오류 증상

```
ImportError: cannot import name 'HookResult' from 'core' (unknown location)
```

사용자가 `/compact` 또는 `/alfred:0-project` 같은 Claude Code 기능을 사용할 때 Hook 초기화 시 발생합니다.

#### 📍 영향받는 파일

**로컬 프로젝트 Hook**:
- `.claude/hooks/alfred/alfred_hooks.py` (Line 63, 83, 202)

**패키지 템플릿 Hook** (동일 문제):
- `src/moai_adk/templates/.claude/hooks/alfred/alfred_hooks.py` (Line 63, 83, 202)

#### 🔍 근본 원인 분석

##### 원인 1: HookResult Import (Line 61-63)

```python
# Line 61
from utils.timeout import CrossPlatformTimeout, TimeoutError as PlatformTimeoutError

# Line 63
from core import HookResult  # ✅ 올바른 import
```

**상태**: ✅ `HookResult`는 `core/__init__.py`에서 정의되어 있고 export됨 (Line 164)
**원인**: 동적 import 경로 문제 - sys.path 설정이 import 시점보다 뒤에 실행됨

##### 원인 2: HookTimeoutError 정의 누락 (Line 83)

```python
# Line 81-83
def _hook_timeout_handler(signum, frame):
    """Signal handler for global hook timeout"""
    raise HookTimeoutError("Hook execution exceeded 5-second timeout")  # ❌ 정의되지 않음
```

**문제**:
- `HookTimeoutError` 클래스가 정의되지 않음
- 사용 가능한 것은 `PlatformTimeoutError` (utils.timeout.TimeoutError의 별칭)

**영향**:
- SIGALRM 신호가 발생할 때 프로그램 크래시
- 5초 이상 걸리는 hook 실행 시 모두 실패

##### 원인 3: timeout 변수 미초기화 (Line 202)

```python
# Line 125-126
signal.signal(signal.SIGALRM, _hook_timeout_handler)
signal.alarm(5)

# Line 202 (finally 블록)
finally:
    timeout.cancel()  # ❌ timeout 변수가 정의되지 않음
```

**문제**:
- `timeout` 변수가 어디에도 초기화되지 않음
- Line 61에서 import한 `CrossPlatformTimeout` 클래스와는 별개
- 이 코드는 Windows에서도 실행 불가

**영향**:
- `finally` 블록에서 AttributeError 발생
- Hook이 정상 종료되지 않음

##### 원인 4: Windows 비호환 구현 (Line 125-126)

```python
# Line 125-126 (Unix/POSIX 전용)
signal.signal(signal.SIGALRM, _hook_timeout_handler)
signal.alarm(5)
```

**문제**:
- `signal.SIGALRM`은 Unix/POSIX에만 존재
- Line 61에서 import한 `CrossPlatformTimeout`은 Windows 지원
- 하지만 실제 코드는 signal 직접 사용

**영향**:
- Windows 사용자의 모든 Hook 실행 실패
- AttributeError: module 'signal' has no attribute 'SIGALRM'

#### 💡 제안 해결 방안

**Fix 1: sys.path 설정 위치 조정**

```python
# 현재 (wrong)
from utils.timeout import ...  # Line 61 - import 먼저
from core import HookResult    # Line 63 - HookResult import

# Add sys.path
HOOKS_DIR = Path(__file__).parent
sys.path.insert(0, str(HOOKS_DIR))  # Line 77 - 너무 뒤

# 수정 (correct)
import sys
from pathlib import Path

# sys.path 먼저 설정
HOOKS_DIR = Path(__file__).parent
if str(HOOKS_DIR) not in sys.path:
    sys.path.insert(0, str(HOOKS_DIR))

# 그 후 import
from utils.timeout import CrossPlatformTimeout, TimeoutError as PlatformTimeoutError
from core import HookResult
```

**Fix 2: CrossPlatformTimeout 사용으로 통일**

```python
# Line 81-126 전체 교체
def main() -> None:
    """Main entry point with cross-platform timeout"""
    try:
        timeout = CrossPlatformTimeout(5)
        timeout.start()

        try:
            # Check for event argument
            if len(sys.argv) < 2:
                print("Usage: alfred_hooks.py <event>", file=sys.stderr)
                sys.exit(1)

            event_name = sys.argv[1]
            # ... 나머지 코드 ...

        except PlatformTimeoutError:
            # Hook 타임아웃 처리
            timeout_response: dict[str, Any] = {
                "continue": True,
                "systemMessage": "⚠️ Hook execution timeout - continuing without session info",
            }
            print(json.dumps(timeout_response))
            print("Hook timeout after 5 seconds", file=sys.stderr)
            sys.exit(1)

    finally:
        # timeout.cancel() 또는 context manager 사용
        pass
```

**Fix 3: Context Manager 패턴 (권장)**

```python
def main() -> None:
    """Main entry point with cross-platform timeout"""
    try:
        with CrossPlatformTimeout(5):
            # 모든 로직 여기에
            if len(sys.argv) < 2:
                print("Usage: alfred_hooks.py <event>", file=sys.stderr)
                sys.exit(1)
            # ...

    except PlatformTimeoutError:
        timeout_response: dict[str, Any] = {
            "continue": True,
            "systemMessage": "⚠️ Hook execution timeout",
        }
        print(json.dumps(timeout_response))
        sys.exit(1)
```

#### 🧪 테스트 방법

```bash
# 로컬 테스트
python .claude/hooks/alfred/alfred_hooks.py SessionStart < test-payload.json

# Windows 호환성 테스트
# Windows 머신에서 위 명령어 실행

# 타임아웃 테스트 (5초 이상 실행되는 hook)
python .claude/hooks/alfred/alfred_hooks.py SessionStart
```

#### 📋 파일 동기화 필요

| 파일 | 위치 |
|------|------|
| ❌ `.claude/hooks/alfred/alfred_hooks.py` | 로컬 |
| ❌ `src/moai_adk/templates/.claude/hooks/alfred/alfred_hooks.py` | 패키지 템플릿 |

**둘 다 동일한 오류가 있으므로 동시에 수정 필요**

---

## 🔴 Additional Critical Issue

### GitHub Discussion #117: Hook 파일 경로 오류 (HIGH PRIORITY)

**출처**: GitHub Discussions #117 (https://github.com/modu-ai/moai-adk/discussions/117)
**제목**: Hook Configuration Error in moai-adk
**심각도**: 🔴 **HIGH** (Hook 배포 시 반복적 오류)
**상태**: Open (부분 해결, 영구적 수정 필요)

#### 💥 오류 증상

```
Failed to spawn: .claude/hooks/alfred/alfred_hooks.py
No such file or directory (os error 2)
```

**전체 경로 예시**:
```
/Users/ip9202/develop/vibe/jeju-tourlist/apps/web/./.claude/hooks/alfred/alfred_hooks.py
```

**발생 빈도**: 하루에 6회 이상 반복

#### 🔍 근본 원인 분석

##### 원인 1: 상대 경로 vs 절대 경로 혼용

**문제**:
- `.claude/settings.json`에 절대 경로로 Hook 경로가 저장됨
- 사용자 환경에 따라 경로가 변경될 때 Hook을 찾을 수 없음
- 특히 프로젝트 이동 또는 클론 후 경로가 무효화됨

**영향받는 파일**:
- `.claude/settings.json` (Hook 경로 설정)
- Hook 초기화 로직 (절대 경로 사용)

##### 원인 2: Template 복사 실패

**문제**:
- `moai-adk init` 실행 시 `.claude/hooks/` 디렉토리를 복사하지 못함
- 또는 부분적으로만 복사되어 필요한 파일이 누락됨

**원인 가능성**:
1. 디렉토리 권한 문제
2. 패키지 템플릿 경로 오류
3. 플랫폼별 경로 처리 오류 (Windows vs Unix)

##### 원인 3: 상대 경로 처리 오류

**현재 시스템**:
- Claude Code는 절대 경로로 Hook 스크립트를 실행
- 상대 경로 `./.claude/hooks/alfred/alfred_hooks.py`에서 `./` 중복 발생

**경로 문제**:
```
# 정상
.claude/hooks/alfred/alfred_hooks.py

# 오류 (중복된 ./)
./.claude/hooks/alfred/alfred_hooks.py

# 절대 경로 (환경 의존)
/Users/ip9202/develop/.../apps/web/.claude/hooks/alfred/alfred_hooks.py
```

##### 원인 4: 플랫폼별 경로 차이

**Windows**:
- 환경 변수 처리 미흡
- 백슬래시 vs 포워드 슬래시 혼용

**macOS/Linux**:
- 상대적으로 잘 작동하지만 절대 경로 문제 발생

#### 💡 제안 해결 방안

**Fix 1: 절대 경로 → 상대 경로로 변경**

```json
// .claude/settings.json (현재 - 절대 경로)
{
  "hooks": {
    "on": true,
    "path": "/Users/ip9202/develop/vibe/jeju-tourlist/apps/web/.claude/hooks/alfred/alfred_hooks.py"
  }
}

// 수정안 (상대 경로)
{
  "hooks": {
    "on": true,
    "path": ".claude/hooks/alfred/alfred_hooks.py"
  }
}
```

**Fix 2: Hook 경로 정규화 함수 추가**

```python
# src/moai_adk/core/hooks/path_resolver.py (새로 생성)

from pathlib import Path

def resolve_hook_path(project_path: Path) -> Path:
    """Resolve hook path with platform-specific handling."""
    hook_file = project_path / ".claude" / "hooks" / "alfred" / "alfred_hooks.py"

    if not hook_file.exists():
        raise FileNotFoundError(f"Hook file not found: {hook_file}")

    # Return relative path for portability
    return Path(".") / ".claude" / "hooks" / "alfred" / "alfred_hooks.py"
```

**Fix 3: moai-adk init에서 경로 설정**

```python
# src/moai_adk/cli/commands/init.py

def generate_settings_json(project_path: Path) -> dict:
    """Generate settings.json with relative hook paths."""
    return {
        "hooks": {
            "on": True,
            "path": ".claude/hooks/alfred/alfred_hooks.py"  # 상대 경로
        }
    }
```

**Fix 4: 디버깅 및 진단 명령어**

```bash
# moai doctor (진단 명령)
$ moai doctor

✓ Project structure
✓ Hook file exists: .claude/hooks/alfred/alfred_hooks.py
✓ settings.json syntax valid
⚠ Hook path is absolute (should be relative)
  → Run: moai fix-hooks
```

#### 📋 파일 동기화 필요

| 파일 | 현황 | 수정 필요 |
|------|------|---------|
| `src/moai_adk/cli/commands/init.py` | settings.json 생성 로직 | ✅ 경로를 상대 경로로 변경 |
| `src/moai_adk/templates/.claude/settings.json` | 템플릿 파일 | ✅ 상대 경로로 업데이트 |
| 사용자 프로젝트 `.claude/settings.json` | 이미 절대 경로로 설정됨 | ✅ migration 스크립트 필요 |

---

## 🟡 Medium Priority Issues

### Issue #155: ExitPlanMode 과도 호출 (MEDIUM PRIORITY)

**이슈 번호**: #155
**제목**: 구현계획이 없는데 승인을 자주 요청하네요
**심각도**: 🟡 **MEDIUM** (UX 저하)
**상태**: Open

#### 💬 사용자 피드백

사용자가 계획(план)을 작성하지 않았는데도 승인을 자주 요청한다고 불평합니다.

#### 🔍 근본 원인 분석

**분석 결과**: 코드베이스 내에서 `ExitPlanMode` 호출을 직접 찾을 수 없음

**가능한 원인들**:

1. **Claude Code의 Plan Mode 자동 동작**
   - Claude Code의 Plan mode에서 자동으로 사용자에게 계획 승인을 요청하는 동작
   - 이것은 MoAI-ADK 코드가 아닌 Claude Code의 기본 동작

2. **서브에이전트의 불필요한 계획 생성**
   - Sub-agent들이 간단한 작업도 계획 단계를 거치도록 설정됨
   - 특히 `/alfred:1-plan`이나 `/alfred:2-run` 명령어에서 발생

3. **Task tool 호출 시 자동 계획 생성**
   - Task tool 사용 시 Claude Code가 자동으로 계획을 생성하고 승인 대기

#### 💡 개선 방안

**Option 1: Plan Mode 비활성화**
- 간단한 읽기/쓰기 작업에서는 Plan mode 불필요
- Task tool 사용 전에 `ExitPlanMode` 호출 고려

**Option 2: Sub-agent 최적화**
- 간단한 작업은 직접 수행 (Task tool 미사용)
- 복잡한 작업만 계획 단계 포함

**Option 3: 사용자 설정**
- `.moai/config.json`에 `require_plan_confirmation` 옵션 추가
- 사용자가 자동 승인 여부 선택 가능

#### 📌 우선순위

현재 코드에서 명확한 원인을 찾을 수 없으므로, 이 이슈는:
- 사용자 환경에서의 Claude Code 동작과 관련
- MoAI-ADK 코드 수정이 필요하지 않을 수 있음
- 사용자와의 협력 탐구 필요

---

### Issue #152: 백업 안내 문구 불일치 (MEDIUM PRIORITY)

**이슈 번호**: #152
**제목**: backup 안내문구
**심각도**: 🟡 **MEDIUM** (사용자 혼란 + 정보 손실)
**상태**: Open

#### 💬 사용자 불만 사항

1. **백업 경로 불일치**: `moai-adk init . `의 안내 문구와 실제 백업 위치가 다름
2. **문서 덮어쓰기**: `/alfred:0-project` 실행 후 사용자가 수정한 내용이 모두 삭제됨
3. **불명확한 안내**: 백업과 새 생성의 차이가 명확하지 않음

#### 📍 영향받는 파일

**CLI 명령어**:
- `src/moai_adk/cli/commands/init.py` (Line 170, 258)

**Backup 유틸리티**:
- `src/moai_adk/core/project/backup_utils.py` (Line 6 주석)

**Alfred 명령어**:
- `src/moai_adk/templates/.claude/commands/alfred/0-project.md` (Line 298-417)

#### 🔍 근본 원인 분석

##### 원인 1: 백업 경로 표시 불일치

**init.py (Line 170, 258)**:
```python
console.print("   Backup will be created at .moai-backups/{timestamp}/\n")
# ...
console.print("   Previous files are backed up in [cyan].moai-backups/{timestamp}/[/cyan]")
```

**backup_utils.py (Line 6)**:
```python
# Backup path: .moai-backups/backup/ (v0.4.2)
```

**문제**:
- CLI에서는 `.moai-backups/{timestamp}/` 형태로 표시
- 코멘트에는 `.moai-backups/backup/`으로 기술
- 실제 구현은 어느 쪽인지 명확하지 않음

**영향**:
- 사용자가 백업 파일을 찾을 수 없음
- 혼란스러운 안내

##### 원인 2: /alfred:0-project에서 문서 덮어쓰기

**현상**:
- 사용자가 `.moai/project/product.md`를 수정
- `/alfred:0-project` 실행
- 모든 수정 사항이 초기화됨

**근본 원인** (CLAUDE.md Line 318-327):

```yaml
**Backup existence conditions**:
- `.moai-backups/` directory exists
- `.moai/project/*.md` file exists in the latest backup folder
- `optimized: false` in `config.json`

**Select user if backup exists**:
Call `AskUserQuestion` to display a TUI with options:
- **Merge**: Merge backup contents and latest template (recommended)
- **New**: Ignore the backup and start a new interview
- **Skip**: Keep current file
```

**예상 동작**:
- 백업이 있을 때는 사용자에게 선택 옵션 제공
- "Merge" 선택 시 기존 콘텐츠 보존

**실제 동작**:
- 백업 머지 플로우가 제대로 작동하지 않음
- 사용자 정의 내용이 템플릿 기본값으로 덮어써짐

##### 원인 3: 안내 문구의 모호함

**현재 안내문** (init.py Line 230-231, 256-261):

```
✅ Initialization Completed Successfully!
📊 Summary: [프로젝트 정보]
⚠️  Configuration Notice:
  All template files have been force overwritten
  Previous files are backed up in .moai-backups/{timestamp}/
```

**문제**:
- "force overwritten"은 강압적으로 들림
- "Previous files are backed up"은 복원 방법을 암시하지 않음
- 사용자가 백업이 자동 복원되는 줄 착각할 수 있음

#### 💡 제안 해결 방안

**Fix 1: 백업 경로 통일 및 명확화**

```python
# backup_utils.py 업데이트
# Backup path structure:
# - Legacy path (v0.3.x): .moai-backups/{timestamp}/
# - Current path (v0.4.2+): .moai-backups/{timestamp}/ (동일하지만 내용 변경)

# 코멘트 명확화
BACKUP_DIR_FORMAT = ".moai-backups/{timestamp}/"
# 예: .moai-backups/20251031-143022/
```

**Fix 2: init.py의 안내 문구 개선**

```python
# 현재 (부정적)
console.print("  All template files have been [bold]force overwritten[/bold]")
console.print("  Previous files are backed up in [cyan].moai-backups/{timestamp}/[/cyan]")

# 개선 (긍정적)
console.print("  [yellow]ℹ️  Template Updates[/yellow]")
console.print("  New template features have been installed")
console.print("  Your custom files are backed up in [cyan].moai-backups/{timestamp}/[/cyan]")
console.print("\n  [cyan]To merge your customizations:[/cyan]")
console.print("  Run [bold]/alfred:0-project[/bold] to restore custom content")
```

**Fix 3: /alfred:0-project의 백업 머지 검증**

CLAUDE.md의 Phase 1.1 (Line 318-447) 구현 확인:

```markdown
### 1.1 Backup merge workflow (when user selects "Merge")

**STEP 1: Read backup file**
- ✅ Read backup files

**STEP 2: Detect template defaults**
- ❓ 구현되어 있는가?

**STEP 3: Extract user customization**
- ❓ 사용자 커스텀만 추출되는가?

**STEP 4: Merge Strategy**
- ❓ 머지 로직이 올바른가?
```

**현재 상태**: 문서에는 명시되어 있지만 실제 구현에서 제대로 동작하지 않음

#### 🧪 테스트 방법

```bash
# 1단계: 프로젝트 초기화
moai-adk init test-project

# 2단계: 문서 수정
echo "## Custom Section\nMy custom content" >> test-project/.moai/project/product.md

# 3단계: 재초기화
cd test-project
moai-adk init . --force

# 4단계: /alfred:0-project 실행
# Claude Code에서 /alfred:0-project 실행

# 5단계: 확인
# product.md에 "Custom Section"이 보존되어 있는가?
```

---

## 📈 우선순위 및 영향도 종합 분석

| 순번 | 이슈 | 심각도 | 영향범위 | 작업량 | 우선순위 |
|------|------|--------|---------|--------|---------|
| 1 | #154/#153 Hook Error | 🔴 HIGH | 전체 시스템 | 중간 | **1순위** |
| 1b | **Disc #117 Hook 경로** | 🔴 HIGH | **Hook 배포** | **중간** | **1순위** |
| 2 | #152 백업 문구 | 🟡 MEDIUM | 초기화 프로세스 | 중간 | **2순위** |
| 3 | #155 ExitPlanMode | 🟡 MEDIUM | UX | 낮음 | **3순위** |

### 작업 계획

**Phase 1A (긴급 수정 - Hook 시스템 전체)**:
- Fix #154/#153: Hook 타임아웃 메커니즘 교체 (ImportError 해결)
- Fix Disc #117: Hook 경로 설정을 상대 경로로 변경
- 예상 시간: 3-4시간
- 테스트: 단위 테스트 + E2E 테스트 + 플랫폼별 테스트
- **의존성**: 두 이슈를 함께 해결해야 Hook 시스템 완전 복구

**Phase 1B (마이그레이션 스크립트)**:
- 기존 사용자 프로젝트의 `.claude/settings.json` 절대 경로 → 상대 경로 변환
- 예상 시간: 1-2시간
- 실행: `moai migrate-hook-paths` 명령어 제공

**Phase 2 (문서화 개선)**:
- Fix #152: 백업 안내 문구 개선 및 머지 로직 검증
- 예상 시간: 2-3시간

**Phase 3 (UX 개선)**:
- #155: Plan mode 로직 분석 및 개선
- 예상 시간: 1-2시간 (또는 Claude Code 설정 변경 권장)

---

## 🔗 참고 자료

### 관련 SPEC 문서
- `.moai/specs/SPEC-BUGFIX-001/` - Hook 타임아웃 관련
- `.moai/specs/SPEC-INIT-003/` - 초기화 및 백업 관련

### 관련 보고서
- `.moai/reports/hooks-*.md` - Hook 시스템 이전 분석 보고서

### 코드 레퍼런스
- `CrossPlatformTimeout` 구현: `.claude/hooks/alfred/utils/timeout.py` (Line 26-114)
- HookResult 정의: `.claude/hooks/alfred/shared/core/__init__.py` (Line 24-164)

---

**분석 완료일**: 2025-10-31
**다음 단계**: Issue 우선순위에 따라 Fix SPEC 작성 및 구현
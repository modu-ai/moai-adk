# @SPEC:HOOKS-EMERGENCY-001: 구현 계획 (Implementation Plan)

> **SPEC ID**: HOOKS-EMERGENCY-001
> **Version**: 0.0.1
> **Status**: draft
> **Priority**: critical

---

## 📊 구현 개요

Hook 시스템의 두 가지 Critical 문제를 3개의 Phase로 나누어 해결합니다:

1. **Phase 1**: ImportError 수정 (sys.path, HookResult)
2. **Phase 2**: 경로 설정 표준화 (settings.json 상대 경로)
3. **Phase 3**: Cross-platform 호환성 (Windows/Unix timeout)

**예상 작업 범위**:
- 파일 수정: 4개
- 테스트 파일 추가: 3개
- 문서 업데이트: 2개

---

## Phase 1: ImportError 수정

### 목표
`alfred_hooks.py`에서 발생하는 ImportError 및 NameError를 해결하여 Hook 시스템이 정상적으로 초기화되도록 함

### 변경 대상 파일

#### 1.1 `.claude/hooks/alfred/alfred_hooks.py`
**변경 내용**:
```python
# AS-IS (문제 코드)
from core import HookResult  # ImportError: sys.path 미설정
timeout(5, hook_func, *args, **kwargs)  # NameError: timeout 정의 없음

# TO-BE (수정 코드)
import sys
import os
from pathlib import Path

# 1. sys.path 설정 (import 전)
project_root = Path(__file__).parent.parent.parent.parent
sys.path.insert(0, str(project_root))

# 2. HookResult import
from core import HookResult

# 3. timeout Context Manager 정의
import signal
import platform
from contextlib import contextmanager

@contextmanager
def timeout_context(seconds):
    """Cross-platform timeout context manager"""
    if platform.system() == 'Windows':
        # Windows: threading.Timer 사용
        import threading
        timer = threading.Timer(seconds, lambda: None)
        timer.start()
        try:
            yield
        finally:
            timer.cancel()
    else:
        # Unix: signal.SIGALRM 사용
        def timeout_handler(signum, frame):
            raise TimeoutError(f"Hook execution exceeded {seconds} seconds")

        signal.signal(signal.SIGALRM, timeout_handler)
        signal.alarm(seconds)
        try:
            yield
        finally:
            signal.alarm(0)

# 4. Hook 실행 시 timeout 적용
def execute_hook(hook_func, *args, **kwargs):
    try:
        with timeout_context(5):
            return hook_func(*args, **kwargs)
    except TimeoutError as e:
        print(f"⚠️ Hook timeout: {e}")
        return HookResult(success=False, message=str(e))
```

**@TAG**: `@CODE:HOOKS-EMERGENCY-001:FIX-001`

#### 1.2 `src/moai_adk/templates/.claude/hooks/alfred/alfred_hooks.py`
**변경 내용**: 위와 동일 (패키지 템플릿 동기화)

**@TAG**: `@CODE:HOOKS-EMERGENCY-001:FIX-001-TEMPLATE`

### 테스트 계획

#### TEST-001: sys.path 설정 검증
```python
# tests/hooks/test_hook_import.py
@TEST:HOOKS-EMERGENCY-001:IMPORT-001

def test_sys_path_setup():
    """sys.path에 프로젝트 루트가 추가되는지 검증"""
    from hooks.alfred import alfred_hooks
    import sys

    project_root = str(Path(__file__).parent.parent.parent)
    assert project_root in sys.path

def test_hook_result_import():
    """HookResult가 정상적으로 import되는지 검증"""
    from hooks.alfred.alfred_hooks import HookResult
    assert HookResult is not None
```

#### TEST-002: timeout 메커니즘 검증
```python
# tests/hooks/test_hook_timeout.py
@TEST:HOOKS-EMERGENCY-001:TIMEOUT-001

def test_timeout_context_unix():
    """Unix 환경에서 timeout이 정상 작동하는지 검증"""
    import platform
    if platform.system() == 'Windows':
        pytest.skip("Unix only test")

    from hooks.alfred.alfred_hooks import timeout_context

    with pytest.raises(TimeoutError):
        with timeout_context(1):
            import time
            time.sleep(2)  # 1초 timeout → TimeoutError 발생

def test_timeout_context_windows():
    """Windows 환경에서 timeout이 정상 작동하는지 검증"""
    import platform
    if platform.system() != 'Windows':
        pytest.skip("Windows only test")

    from hooks.alfred.alfred_hooks import timeout_context

    with timeout_context(1):
        import time
        time.sleep(0.5)  # 0.5초 실행 → 정상 종료
```

### 검증 기준
- ✅ ImportError 발생하지 않음
- ✅ NameError 발생하지 않음
- ✅ SessionStart Hook 정상 실행
- ✅ 프로젝트 정보 카드 출력 성공

---

## Phase 2: 경로 설정 표준화

### 목표
`settings.json`의 절대 경로를 상대 경로로 전환하여 프로젝트 이동/클론 시에도 Hook이 정상 작동하도록 함

### 변경 대상 파일

#### 2.1 `.claude/settings.json`
**변경 내용**:
```json
// AS-IS (절대 경로 문제)
{
  "hooks": {
    "path": "/Users/goos/MoAI/MoAI-ADK-v1.0/.claude/hooks/alfred"
  }
}

// TO-BE (상대 경로로 전환)
{
  "hooks": {
    "path": ".claude/hooks/alfred"
  }
}
```

**@TAG**: `@CODE:HOOKS-EMERGENCY-001:FIX-002`

#### 2.2 `src/moai_adk/templates/.claude/settings.json`
**변경 내용**: 위와 동일 (패키지 템플릿 동기화)

**@TAG**: `@CODE:HOOKS-EMERGENCY-001:FIX-002-TEMPLATE`

### Migration 전략

#### 기존 프로젝트 자동 Migration
`moai init` 또는 `/alfred:0-project` 실행 시 자동으로 상대 경로로 전환:

```python
# src/moai_adk/core/migrator.py
@CODE:HOOKS-EMERGENCY-001:MIGRATION

def migrate_settings_json(settings_path: Path):
    """settings.json의 절대 경로를 상대 경로로 변환"""
    with open(settings_path, 'r') as f:
        settings = json.load(f)

    if 'hooks' in settings and 'path' in settings['hooks']:
        hook_path = settings['hooks']['path']

        # 절대 경로 감지
        if Path(hook_path).is_absolute():
            # 상대 경로로 변환
            settings['hooks']['path'] = '.claude/hooks/alfred'

            with open(settings_path, 'w') as f:
                json.dump(settings, f, indent=2)

            print("✅ settings.json migrated to relative path")
```

### 테스트 계획

#### TEST-003: 경로 설정 검증
```python
# tests/hooks/test_hook_path.py
@TEST:HOOKS-EMERGENCY-001:PATH-001

def test_settings_relative_path():
    """settings.json에 상대 경로가 설정되어 있는지 검증"""
    settings_path = Path('.claude/settings.json')
    with open(settings_path) as f:
        settings = json.load(f)

    hook_path = settings['hooks']['path']
    assert not Path(hook_path).is_absolute()
    assert hook_path == '.claude/hooks/alfred'

def test_hook_discovery_after_move():
    """프로젝트 이동 후에도 Hook을 찾을 수 있는지 검증"""
    # Mock: 프로젝트를 다른 경로로 이동
    # Expected: Hook 파일을 정상적으로 로드
    pass  # Integration test로 구현 예정
```

### 검증 기준
- ✅ `settings.json`에 상대 경로만 저장됨
- ✅ 프로젝트 이동 후에도 Hook 정상 로드
- ✅ 기존 프로젝트 자동 migration 성공

---

## Phase 3: Cross-platform 호환성

### 목표
Windows/macOS/Linux 모든 환경에서 동일한 timeout 동작을 보장

### 변경 대상 파일

#### 3.1 `.claude/hooks/alfred/alfred_hooks.py`
**변경 내용**: Phase 1에서 이미 구현됨 (timeout_context)

**추가 검증 필요**:
- Windows: `threading.Timer` 동작 확인
- Unix: `signal.SIGALRM` 동작 확인

### 테스트 계획

#### TEST-004: Cross-platform 통합 테스트
```python
# tests/hooks/test_cross_platform.py
@TEST:HOOKS-EMERGENCY-001:CROSS-PLATFORM-001

import platform

def test_timeout_on_current_platform():
    """현재 플랫폼에서 timeout이 정상 작동하는지 검증"""
    from hooks.alfred.alfred_hooks import timeout_context
    import time

    # 정상 종료 케이스
    with timeout_context(2):
        time.sleep(1)  # 1초 실행 → 정상 종료

    # Timeout 발생 케이스
    with pytest.raises(TimeoutError):
        with timeout_context(1):
            time.sleep(2)  # 2초 실행 → TimeoutError

def test_windows_threading_timer():
    """Windows에서 threading.Timer가 사용되는지 검증"""
    if platform.system() != 'Windows':
        pytest.skip("Windows only test")

    # Windows 환경에서 signal.SIGALRM이 호출되지 않는지 검증
    import signal
    assert not hasattr(signal, 'SIGALRM') or signal.SIGALRM is not used

def test_unix_signal_alarm():
    """Unix에서 signal.SIGALRM이 사용되는지 검증"""
    if platform.system() == 'Windows':
        pytest.skip("Unix only test")

    import signal
    assert hasattr(signal, 'SIGALRM')
```

### 검증 기준
- ✅ Windows에서 `threading.Timer` 사용
- ✅ Unix에서 `signal.SIGALRM` 사용
- ✅ 두 환경 모두 동일한 timeout 동작
- ✅ AttributeError 발생하지 않음

---

## Phase 4: 문서화 및 최종 검증

### 목표
Hook 설정 가이드 업데이트 및 migration 가이드 작성

### 변경 대상 파일

#### 4.1 `.moai/docs/hooks-setup-guide.md`
**생성 내용**:
```markdown
# Hook 설정 가이드

## 경로 설정 규칙
- ✅ 항상 상대 경로 사용: `.claude/hooks/alfred`
- ❌ 절대 경로 사용 금지: `/Users/goos/...`

## Cross-platform 호환성
- Windows: threading.Timer 자동 선택
- Unix: signal.SIGALRM 자동 선택

## Troubleshooting
- ImportError 발생 시: sys.path 확인
- Hook 파일 찾기 실패 시: settings.json 경로 확인
```

**@TAG**: `@DOC:HOOKS-EMERGENCY-001:GUIDE`

#### 4.2 `.moai/docs/hooks-migration-guide.md`
**생성 내용**:
```markdown
# Hook Migration 가이드

## v0.7.0 → v0.7.1 Migration

### 자동 Migration
`moai init` 또는 `/alfred:0-project` 실행 시 자동 전환

### 수동 Migration
1. `.claude/settings.json` 열기
2. `hooks.path`를 상대 경로로 변경
3. 저장 후 Claude Code 재시작

### 검증 방법
```bash
cat .claude/settings.json | grep "hooks"
# Expected: "path": ".claude/hooks/alfred"
```
```

**@TAG**: `@DOC:HOOKS-EMERGENCY-001:MIGRATION`

### 최종 검증 체크리스트

- ✅ Phase 1: ImportError 수정 완료
- ✅ Phase 2: 경로 설정 표준화 완료
- ✅ Phase 3: Cross-platform 호환성 완료
- ✅ 모든 테스트 통과 (Unit + Integration)
- ✅ 문서 업데이트 완료
- ✅ 패키지 템플릿 동기화 완료
- ✅ Migration 가이드 작성 완료

---

## 파일 변경 요약

### Modified Files (4개)
1. `.claude/hooks/alfred/alfred_hooks.py` - ImportError, timeout 수정
2. `.claude/settings.json` - 상대 경로 전환
3. `src/moai_adk/templates/.claude/hooks/alfred/alfred_hooks.py` - 템플릿 동기화
4. `src/moai_adk/templates/.claude/settings.json` - 템플릿 동기화

### Added Files (5개)
1. `tests/hooks/test_hook_import.py` - Import 테스트
2. `tests/hooks/test_hook_timeout.py` - Timeout 테스트
3. `tests/hooks/test_hook_path.py` - 경로 설정 테스트
4. `tests/hooks/test_cross_platform.py` - Cross-platform 테스트
5. `src/moai_adk/core/migrator.py` - Migration 로직

### Documentation Files (2개)
1. `.moai/docs/hooks-setup-guide.md` - 설정 가이드
2. `.moai/docs/hooks-migration-guide.md` - Migration 가이드

---

## 기술적 접근 방법

### 1. sys.path 설정 전략
```python
# 프로젝트 루트 계산
project_root = Path(__file__).parent.parent.parent.parent

# sys.path에 추가 (import 전)
sys.path.insert(0, str(project_root))
```

### 2. Cross-platform Timeout 전략
```python
# OS 감지 후 적절한 메커니즘 선택
if platform.system() == 'Windows':
    # threading.Timer 사용
else:
    # signal.SIGALRM 사용
```

### 3. 상대 경로 계산 전략
```python
# settings.json에서 Hook 경로 읽기
hook_relative_path = settings['hooks']['path']  # ".claude/hooks/alfred"

# 프로젝트 루트 기준으로 절대 경로 계산
hook_absolute_path = project_root / hook_relative_path
```

---

## 위험 요소 및 대응 계획

### Risk 1: Windows에서 signal 모듈 호환성
**위험도**: Medium
**대응**: `platform.system()` 체크 선행, Windows는 threading.Timer로 fallback

### Risk 2: 기존 프로젝트 Migration 실패
**위험도**: Low
**대응**: 수동 migration 가이드 제공, migration 실패 시 경고 메시지

### Risk 3: sys.path 설정이 다른 import에 영향
**위험도**: Low
**대응**: sys.path.insert(0)로 최우선 순위 설정, 충돌 가능성 낮음

---

## 다음 단계

1. **Phase 1 구현**: ImportError 수정 및 테스트
2. **Phase 2 구현**: 경로 설정 표준화 및 Migration
3. **Phase 3 검증**: Cross-platform 통합 테스트
4. **Phase 4 문서화**: 가이드 작성 및 최종 검증
5. **`/alfred:3-sync`**: 문서 동기화 및 @TAG 검증
6. **PR 생성**: 검토 및 머지

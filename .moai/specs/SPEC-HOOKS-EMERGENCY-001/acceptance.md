# @SPEC:HOOKS-EMERGENCY-001: 수락 기준 (Acceptance Criteria)

> **SPEC ID**: HOOKS-EMERGENCY-001
> **Version**: 0.0.1
> **Status**: draft
> **Priority**: critical

---

## 📋 수락 기준 개요

이 문서는 Hook 시스템 긴급 복구가 성공적으로 완료되었는지 검증하기 위한 수락 기준을 정의합니다.

**검증 범위**:
1. ImportError 해결
2. 경로 설정 표준화
3. Cross-platform 호환성
4. Migration 동작

---

## AC-001: ImportError 해결

### 수락 조건
Hook 시스템이 ImportError 없이 정상적으로 초기화되어야 함

### Given-When-Then 시나리오

#### Scenario 1.1: SessionStart Hook 초기화
```gherkin
Given: 사용자가 MoAI-ADK 프로젝트를 열었을 때
When: Claude Code가 SessionStart 이벤트를 트리거할 때
Then:
  - ImportError가 발생하지 않아야 함
  - HookResult를 성공적으로 import해야 함
  - 프로젝트 정보 카드가 출력되어야 함
```

**검증 방법**:
```bash
# 1. Claude Code 실행
cd /Users/goos/MoAI/MoAI-ADK-v1.0
claude-code

# 2. SessionStart Hook 실행 확인
# Expected Output:
# ╭─────────────────────────────────╮
# │ 🗿 MoAI-ADK v0.7.0              │
# │ Mode: Personal | Lang: python   │
# ╰─────────────────────────────────╯

# 3. 에러 없이 정상 실행 확인
```

#### Scenario 1.2: sys.path 설정 검증
```gherkin
Given: alfred_hooks.py가 실행될 때
When: HookResult를 import하려고 시도할 때
Then:
  - sys.path에 프로젝트 루트가 추가되어 있어야 함
  - from core import HookResult가 성공해야 함
```

**검증 방법**:
```python
# tests/hooks/test_hook_import.py
def test_sys_path_contains_project_root():
    from hooks.alfred import alfred_hooks
    import sys
    from pathlib import Path

    project_root = str(Path(__file__).parent.parent.parent)
    assert project_root in sys.path
```

#### Scenario 1.3: HookTimeoutError 처리
```gherkin
Given: Hook이 5초 이상 실행될 때
When: Timeout이 발생할 때
Then:
  - TimeoutError가 정상적으로 발생해야 함
  - Claude Code가 동결되지 않아야 함
  - 경고 메시지가 출력되어야 함
```

**검증 방법**:
```python
# tests/hooks/test_hook_timeout.py
def test_hook_timeout_triggers():
    from hooks.alfred.alfred_hooks import timeout_context
    import time

    with pytest.raises(TimeoutError):
        with timeout_context(1):
            time.sleep(2)  # 1초 timeout → TimeoutError 발생
```

### 수락 체크리스트
- ✅ SessionStart Hook이 ImportError 없이 실행됨
- ✅ HookResult import 성공
- ✅ sys.path에 프로젝트 루트가 포함됨
- ✅ Timeout 메커니즘이 정상 작동함
- ✅ NameError 발생하지 않음

---

## AC-002: 경로 설정 표준화

### 수락 조건
Hook 경로가 상대 경로로 설정되어 프로젝트 이동/클론 시에도 정상 작동해야 함

### Given-When-Then 시나리오

#### Scenario 2.1: settings.json 상대 경로 검증
```gherkin
Given: 새로운 프로젝트를 초기화했을 때
When: .claude/settings.json을 확인할 때
Then:
  - hooks.path가 상대 경로로 설정되어 있어야 함
  - 절대 경로가 아니어야 함
```

**검증 방법**:
```bash
# 1. settings.json 확인
cat .claude/settings.json | jq '.hooks.path'

# Expected Output:
# ".claude/hooks/alfred"

# 2. 절대 경로가 아님을 확인
cat .claude/settings.json | grep -v "^/" | grep "hooks.path"
```

#### Scenario 2.2: 프로젝트 이동 후 Hook 탐색
```gherkin
Given: 프로젝트가 /path/A에 있었을 때
When: 프로젝트를 /path/B로 이동한 후 Claude Code를 실행할 때
Then:
  - Hook 파일을 정상적으로 찾아야 함
  - SessionStart Hook이 성공적으로 실행되어야 함
```

**검증 방법**:
```bash
# 1. 프로젝트 이동
mv /Users/goos/MoAI/MoAI-ADK-v1.0 /tmp/MoAI-ADK-v1.0

# 2. Claude Code 실행
cd /tmp/MoAI-ADK-v1.0
claude-code

# 3. Hook 정상 로드 확인
# Expected: SessionStart Hook 실행 성공
```

#### Scenario 2.3: 프로젝트 클론 후 Hook 탐색
```gherkin
Given: 프로젝트를 GitHub에서 클론했을 때
When: 다른 머신에서 프로젝트를 열었을 때
Then:
  - Hook 파일을 자동으로 찾아야 함
  - settings.json의 경로 설정이 유효해야 함
```

**검증 방법**:
```bash
# 1. 프로젝트 클론 (시뮬레이션)
git clone <repo> /tmp/cloned-project

# 2. Claude Code 실행
cd /tmp/cloned-project
claude-code

# 3. Hook 정상 로드 확인
```

### 수락 체크리스트
- ✅ settings.json에 상대 경로만 저장됨
- ✅ 프로젝트 이동 후에도 Hook 정상 로드
- ✅ 프로젝트 클론 후에도 Hook 정상 로드
- ✅ 절대 경로 사용 금지 규칙 준수

---

## AC-003: Cross-platform 호환성

### 수락 조건
Windows, macOS, Linux 모든 환경에서 동일한 Hook 동작을 보장해야 함

### Given-When-Then 시나리오

#### Scenario 3.1: Windows 환경 timeout
```gherkin
Given: Windows 환경에서 Hook을 실행할 때
When: Hook이 5초 이상 실행될 때
Then:
  - threading.Timer를 사용해야 함
  - signal.SIGALRM이 호출되지 않아야 함
  - TimeoutError가 발생하거나 graceful 종료되어야 함
```

**검증 방법**:
```python
# tests/hooks/test_cross_platform.py
import platform
import pytest

@pytest.mark.skipif(platform.system() != 'Windows', reason="Windows only")
def test_windows_timeout():
    from hooks.alfred.alfred_hooks import timeout_context
    import time

    # threading.Timer 사용 확인
    with timeout_context(2):
        time.sleep(1)  # 정상 종료

    # signal.SIGALRM이 사용되지 않음 확인
    import signal
    with timeout_context(1):
        assert not hasattr(signal, 'SIGALRM') or signal.SIGALRM is not used
```

#### Scenario 3.2: Unix 환경 timeout
```gherkin
Given: macOS 또는 Linux 환경에서 Hook을 실행할 때
When: Hook이 5초 이상 실행될 때
Then:
  - signal.SIGALRM을 사용해야 함
  - TimeoutError가 발생해야 함
```

**검증 방법**:
```python
@pytest.mark.skipif(platform.system() == 'Windows', reason="Unix only")
def test_unix_timeout():
    from hooks.alfred.alfred_hooks import timeout_context
    import time
    import signal

    # signal.SIGALRM 사용 확인
    assert hasattr(signal, 'SIGALRM')

    with pytest.raises(TimeoutError):
        with timeout_context(1):
            time.sleep(2)  # TimeoutError 발생
```

#### Scenario 3.3: Cross-platform 일관성
```gherkin
Given: 어떤 운영체제에서든 Hook을 실행할 때
When: Timeout이 발생할 때
Then:
  - 모든 플랫폼에서 동일한 동작을 보여야 함
  - AttributeError가 발생하지 않아야 함
```

**검증 방법**:
```python
def test_timeout_consistency():
    """모든 플랫폼에서 동일한 timeout 동작 검증"""
    from hooks.alfred.alfred_hooks import timeout_context
    import time

    # 정상 종료 (모든 플랫폼)
    with timeout_context(2):
        time.sleep(1)

    # Timeout 발생 (모든 플랫폼)
    try:
        with timeout_context(1):
            time.sleep(2)
        assert False, "TimeoutError should have been raised"
    except TimeoutError:
        pass  # Expected
```

### 수락 체크리스트
- ✅ Windows에서 threading.Timer 사용
- ✅ Unix에서 signal.SIGALRM 사용
- ✅ 모든 플랫폼에서 동일한 timeout 동작
- ✅ AttributeError 발생하지 않음
- ✅ Cross-platform 통합 테스트 통과

---

## AC-004: Migration 동작

### 수락 조건
기존 프로젝트의 settings.json이 자동으로 상대 경로로 전환되어야 함

### Given-When-Then 시나리오

#### Scenario 4.1: 자동 Migration 트리거
```gherkin
Given: 기존 프로젝트의 settings.json에 절대 경로가 설정되어 있을 때
When: moai init 또는 /alfred:0-project를 실행할 때
Then:
  - settings.json이 상대 경로로 자동 전환되어야 함
  - 기존 설정은 유지되어야 함
```

**검증 방법**:
```bash
# 1. 기존 settings.json 확인 (절대 경로)
cat .claude/settings.json
# {"hooks": {"path": "/Users/goos/MoAI/MoAI-ADK-v1.0/.claude/hooks/alfred"}}

# 2. Migration 실행
moai init

# 3. Migration 결과 확인
cat .claude/settings.json
# {"hooks": {"path": ".claude/hooks/alfred"}}
```

#### Scenario 4.2: Migration 검증
```gherkin
Given: Migration이 완료되었을 때
When: settings.json을 확인할 때
Then:
  - hooks.path가 상대 경로여야 함
  - Migration 완료 메시지가 출력되어야 함
```

**검증 방법**:
```python
# tests/hooks/test_migration.py
def test_migration_absolute_to_relative():
    """절대 경로 → 상대 경로 Migration 검증"""
    import json
    from pathlib import Path
    from moai_adk.core.migrator import migrate_settings_json

    # 절대 경로 settings.json 생성
    settings_path = Path('.claude/settings.json')
    settings_path.write_text(json.dumps({
        "hooks": {
            "path": "/absolute/path/.claude/hooks/alfred"
        }
    }))

    # Migration 실행
    migrate_settings_json(settings_path)

    # 상대 경로로 전환 확인
    settings = json.loads(settings_path.read_text())
    assert settings['hooks']['path'] == '.claude/hooks/alfred'
```

#### Scenario 4.3: Migration 중복 실행 안전성
```gherkin
Given: Migration이 이미 완료된 프로젝트에서
When: Migration을 다시 실행할 때
Then:
  - 상대 경로가 유지되어야 함
  - 오류가 발생하지 않아야 함
```

**검증 방법**:
```python
def test_migration_idempotency():
    """Migration 중복 실행 안전성 검증"""
    from moai_adk.core.migrator import migrate_settings_json

    settings_path = Path('.claude/settings.json')

    # 첫 번째 Migration
    migrate_settings_json(settings_path)
    first_result = json.loads(settings_path.read_text())

    # 두 번째 Migration (중복)
    migrate_settings_json(settings_path)
    second_result = json.loads(settings_path.read_text())

    # 동일한 결과 확인
    assert first_result == second_result
```

### 수락 체크리스트
- ✅ 절대 경로 → 상대 경로 자동 전환
- ✅ Migration 완료 메시지 출력
- ✅ 기존 설정 유지
- ✅ 중복 실행 시 안전성 보장
- ✅ Migration 가이드 문서 제공

---

## AC-005: 테스트 커버리지

### 수락 조건
모든 핵심 기능이 테스트로 커버되어야 함

### Given-When-Then 시나리오

#### Scenario 5.1: Unit 테스트 커버리지
```gherkin
Given: Hook 시스템의 모든 함수와 클래스가 정의되었을 때
When: pytest를 실행할 때
Then:
  - 모든 Unit 테스트가 통과해야 함
  - 커버리지가 90% 이상이어야 함
```

**검증 방법**:
```bash
# 1. 테스트 실행
pytest tests/hooks/ -v --cov=.claude/hooks/alfred

# 2. 커버리지 확인
# Expected: >= 90%
```

#### Scenario 5.2: Integration 테스트
```gherkin
Given: Hook 시스템이 전체적으로 통합되었을 때
When: 실제 프로젝트에서 Hook을 실행할 때
Then:
  - SessionStart Hook이 성공해야 함
  - PreToolUse Hook이 성공해야 함
  - PostToolUse Hook이 성공해야 함
```

**검증 방법**:
```bash
# 1. Integration 테스트 실행
pytest tests/integration/test_hooks_e2e.py -v

# 2. 모든 Hook 실행 성공 확인
```

### 수락 체크리스트
- ✅ Unit 테스트 커버리지 >= 90%
- ✅ Integration 테스트 통과
- ✅ Cross-platform 테스트 통과
- ✅ 모든 시나리오 테스트 통과

---

## 최종 수락 체크리스트

### Critical Requirements
- ✅ **AC-001**: ImportError 해결 완료
- ✅ **AC-002**: 경로 설정 표준화 완료
- ✅ **AC-003**: Cross-platform 호환성 완료

### High Priority Requirements
- ✅ **AC-004**: Migration 동작 완료
- ✅ **AC-005**: 테스트 커버리지 >= 90%

### Documentation Requirements
- ✅ Hook 설정 가이드 작성 완료
- ✅ Migration 가이드 작성 완료
- ✅ Troubleshooting 가이드 작성 완료

### Quality Gates
- ✅ 모든 Unit 테스트 통과
- ✅ 모든 Integration 테스트 통과
- ✅ Linting 통과 (ruff, black)
- ✅ Type checking 통과 (mypy)

---

## Definition of Done

### 구현 완료 조건
1. ✅ 모든 Phase (1-4) 구현 완료
2. ✅ 모든 파일 변경 사항 적용 완료
3. ✅ 패키지 템플릿 동기화 완료

### 테스트 완료 조건
1. ✅ 모든 수락 시나리오 테스트 통과
2. ✅ Unit 테스트 커버리지 >= 90%
3. ✅ Integration 테스트 통과
4. ✅ Cross-platform 테스트 통과

### 문서 완료 조건
1. ✅ `.moai/docs/hooks-setup-guide.md` 작성
2. ✅ `.moai/docs/hooks-migration-guide.md` 작성
3. ✅ README.md 업데이트
4. ✅ CHANGELOG.md 업데이트

### 배포 준비 조건
1. ✅ Git 커밋 메시지 표준 준수
2. ✅ @TAG 체인 검증 완료
3. ✅ PR 생성 및 리뷰 완료
4. ✅ main 브랜치 머지 완료

---

## 롤백 기준

다음 조건 중 하나라도 발생하면 SPEC 구현을 롤백합니다:

### 롤백 트리거
1. ❌ Windows 환경에서 Hook 실행 실패
2. ❌ 기존 프로젝트 Migration 실패율 > 10%
3. ❌ Unit 테스트 커버리지 < 80%
4. ❌ Critical Bug 발견 (데이터 손실, 보안 문제)

### 롤백 절차
1. 변경 사항 revert
2. 기존 버전으로 패키지 템플릿 복구
3. 롤백 사유 문서화
4. 재계획 및 재구현

---

## 다음 단계

1. **`/alfred:2-run SPEC-HOOKS-EMERGENCY-001`**: TDD 구현 시작
2. **RED → GREEN → REFACTOR**: 각 수락 시나리오별로 TDD 사이클 실행
3. **`/alfred:3-sync`**: 문서 동기화 및 @TAG 검증
4. **PR 생성**: 검토 및 머지
5. **Release**: v0.7.1 배포

---

**최종 수락 조건**: 위의 모든 체크리스트 항목이 ✅ 상태여야 SPEC 구현이 완료된 것으로 간주합니다.

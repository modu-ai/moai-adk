---
spec_id: SPEC-STATUSLINE-UVX-001
title: "uvx 기반 Statusline 실행 - 수용 기준 및 테스트 시나리오"
version: 1.0
status: draft
created_at: 2025-11-17
---

## 수용 기준 (Acceptance Criteria)

### AC-1: CLI 진입점이 모든 OS에서 실행 가능해야 함

**Given** (주어진 조건):
- MoAI-ADK 패키지가 설치됨 (`pip install moai-adk` 또는 로컬 설치)
- uvx가 설치됨 (uv ≥ 0.4.0 포함)

**When** (수행 조건):
- 사용자가 `moai-adk statusline` 명령어 실행
- stdin으로 Claude Code의 JSON 컨텍스트 전달

**Then** (예상 결과):
- 명령어가 0 종료 코드로 완료
- stdout으로 올바른 형식의 statusline 문자열 출력
- stderr에 오류 메시지 없음

**테스트 코드**:
```python
@pytest.mark.parametrize("os_name", ["windows", "linux", "darwin"])
def test_moai_adk_statusline_command_on_all_os(os_name):
    """Test: moai-adk statusline command works on all OS"""

    # Mock stdout input (Claude Code JSON context)
    session_context = {
        "model": {"display_name": "Claude Haiku 4.5"},
        "cwd": "/path/to/project"
    }

    # Run command
    result = subprocess.run(
        ["moai-adk", "statusline"],
        input=json.dumps(session_context),
        capture_output=True,
        text=True
    )

    # Assert: Command succeeds
    assert result.returncode == 0, f"Failed on {os_name}: {result.stderr}"

    # Assert: Output is not empty
    assert result.stdout.strip(), "Statusline output is empty"

    # Assert: No error output
    assert not result.stderr, f"Unexpected stderr: {result.stderr}"
```

---

### AC-2: Claude Code statusline이 uvx 명령어로 자동 실행되어야 함

**Given** (주어진 조건):
- `.claude/settings.json`이 다음과 같이 설정됨:
  ```json
  {
    "statusLine": {
      "type": "command",
      "command": "uvx moai-adk statusline",
      "padding": 0
    }
  }
  ```
- MoAI-ADK 프로젝트에서 Claude Code 세션 시작

**When** (수행 조건):
- Claude Code 세션 초기화 시점 (SessionStart 훅)
- statusline 렌더링 필요 시점

**Then** (예상 결과):
- Claude Code 하단에 statusline 표시
- 표시 내용이 정확함 (모델, 버전, 브랜치, git 상태 등)
- statusline 렌더링에 성능 저하 없음

**검증 방법**:
1. Claude Code UI에서 하단 statusline 시각 확인
2. `.moai/cache/statusline-perf.json` 로그로 실행 시간 확인
3. 모든 필드가 정확하게 표시되는지 확인

---

### AC-3: uvx 캐시가 효과적으로 작동해야 함

**Given** (주어진 조건):
- 첫 번째 실행에서 moai-adk가 uvx 캐시에 저장됨
- 환경 변수가 기본값

**When** (수행 조건):
- `moai-adk statusline` 명령어를 연속으로 10회 실행
- uvx 캐시 상태 모니터링

**Then** (예상 결과):
- 첫 실행: 2-3초 (다운로드 + 캐시)
- 2-10번째 실행: 100-200ms (캐시 히트)
- 캐시 히트율: ≥ 90%

**테스트 코드**:
```python
def test_uvx_cache_effectiveness():
    """Test: uvx cache improves performance on repeated runs"""

    execution_times = []

    for i in range(10):
        start = time.time()
        result = subprocess.run(
            ["moai-adk", "statusline"],
            input=json.dumps({"model": {}, "cwd": "."}),
            capture_output=True,
            text=True
        )
        elapsed = time.time() - start
        execution_times.append(elapsed)

        assert result.returncode == 0

    # First run should be slower (cache miss)
    # Subsequent runs should be much faster (cache hit)
    avg_first = execution_times[0]
    avg_rest = sum(execution_times[1:]) / 9

    assert avg_rest < avg_first / 2, \
        f"Cache not effective: first={avg_first:.2f}s, rest={avg_rest:.2f}s"
```

---

### AC-4: 모든 OS에서 statusline 내용이 일관되어야 함

**Given** (주어진 조건):
- 동일한 프로젝트 상태 (Windows, macOS, Linux에서)
- 동일한 Claude Code 세션 컨텍스트

**When** (수행 조건):
- 각 OS에서 `moai-adk statusline` 실행
- 출력 내용 비교

**Then** (예상 결과):
- 포함된 필드가 동일 (모델, 버전, 브랜치 등)
- 형식이 동일 (이모지, 공백 등)
- 렌더링 방식이 동일 (색상, 패딩 등)

**테스트 코드**:
```python
@pytest.mark.parametrize("os_platform", ["win32", "linux", "darwin"])
def test_statusline_consistency_across_os(os_platform, monkeypatch):
    """Test: Statusline output is consistent across Windows, macOS, Linux"""

    monkeypatch.setattr("sys.platform", os_platform)

    session_context = {
        "model": {"display_name": "Claude Haiku 4.5"},
        "cwd": "/Users/goos/MoAI/MoAI-ADK"
    }

    result = subprocess.run(
        ["moai-adk", "statusline"],
        input=json.dumps(session_context),
        capture_output=True,
        text=True
    )

    output = result.stdout.strip()

    # Assert: Output contains all expected fields
    assert "Haiku" in output or "Sonnet" in output, "Model not in output"
    assert "|" in output, "Field separator not found"

    # Store for cross-platform comparison
    # (Actual comparison done in integration tests)
```

---

### AC-5: 템플릿이 동기화되어 신규 프로젝트에 적용되어야 함

**Given** (주어진 조건):
- `src/moai_adk/templates/.claude/settings.json`이 업데이트됨
- 신규 MoAI-ADK 프로젝트 생성

**When** (수행 조건):
- `moai-adk init` 또는 템플릿 복사로 새 프로젝트 생성
- `.claude/settings.json` 확인

**Then** (예상 결과):
- 생성된 프로젝트의 `.claude/settings.json`에 `uvx moai-adk statusline` 포함
- statusline 명령어가 uvx 방식

**테스트 코드**:
```python
def test_template_sync_statusline_uvx(tmp_path):
    """Test: New projects get uvx statusline from template"""

    # Copy template to temp directory
    template_dir = Path("src/moai_adk/templates/.claude")
    settings_file = template_dir / "settings.json"

    # Parse template settings
    with open(settings_file) as f:
        template_settings = json.load(f)

    # Assert: Template contains uvx command
    assert template_settings["statusLine"]["command"] == "uvx moai-adk statusline", \
        "Template not updated with uvx command"
```

---

### AC-6: 기존 uv run 방식과의 호환성 유지

**Given** (주어진 조건):
- 기존 프로젝트가 `uv run python -m moai_adk.statusline.main` 사용 중
- 새로운 uvx 방식이 구현됨

**When** (수행 조건):
- 기존 방식으로 statusline 실행 시도
- 새로운 방식으로 statusline 실행 시도

**Then** (예상 결과):
- 둘 다 정상 작동 (역호환성)
- 새로운 방식이 더 빠름 (성능 개선)
- 오류 메시지 없음

**테스트 코드**:
```python
def test_backward_compatibility_uv_run():
    """Test: Old uv run method still works"""

    session_context = {
        "model": {"display_name": "Test"},
        "cwd": "."
    }

    # Old way
    result_old = subprocess.run(
        ["uv", "run", "python", "-m", "moai_adk.statusline.main"],
        input=json.dumps(session_context),
        capture_output=True,
        text=True
    )

    # New way
    result_new = subprocess.run(
        ["moai-adk", "statusline"],
        input=json.dumps(session_context),
        capture_output=True,
        text=True
    )

    # Both should succeed
    assert result_old.returncode == 0, f"Old method failed: {result_old.stderr}"
    assert result_new.returncode == 0, f"New method failed: {result_new.stderr}"

    # Both should produce output
    assert result_old.stdout.strip(), "Old method produced no output"
    assert result_new.stdout.strip(), "New method produced no output"
```

---

## Given-When-Then 형식의 테스트 시나리오

### Scenario 1: 첫 번째 statusline 렌더링 (Windows 환경)

```gherkin
Given MoAI-ADK 프로젝트가 클론됨
  And .claude/settings.json이 uvx 방식으로 설정됨
  And Windows 10 또는 11 환경
  And uvx가 설치됨 (uv >= 0.4.0)

When Claude Code 세션 시작 시 statusline 렌더링

Then 다음이 보장되어야 함:
  - statusline이 하단에 표시됨
  - 내용이 정확함 (모델명, 버전 등)
  - 렌더링 시간 < 3초 (첫 실행)
  - stderr 오류 없음

Test:
  1. 프로젝트 디렉토리로 이동
  2. Claude Code 시작
  3. 하단 statusline 확인 (시각)
  4. 콘솔 로그에서 오류 확인
```

### Scenario 2: 반복 실행 성능 (macOS/Linux)

```gherkin
Given uvx 캐시가 초기화됨
  And moai-adk가 설치됨
  And macOS 또는 Linux 환경

When 'moai-adk statusline' 명령어를 10회 연속 실행

Then 다음이 보장되어야 함:
  - 첫 실행: 2-3초 (초기 캐시)
  - 2-10번째: 100-200ms 이내 (캐시 히트)
  - 캐시 히트율: >= 90%
  - 모든 실행 결과 동일

Test:
  1. for i in {1..10}; do time moai-adk statusline; done
  2. 평균 실행 시간 계산
  3. 첫 실행 vs 나머지 실행 비교
```

### Scenario 3: 오류 처리 (uvx 미설치)

```gherkin
Given uvx가 설치되지 않은 환경
  And uv는 설치됨 (하지만 uvx 미포함)

When 'moai-adk statusline' 명령어 실행

Then 다음이 보장되어야 함:
  - 명령어 실패 (종료 코드 != 0)
  - 명확한 오류 메시지 표시
  - 사용자가 해결 방법 이해 가능

Test:
  1. uvx 제거 (임시)
  2. moai-adk statusline 실행
  3. 오류 메시지 확인
  4. uvx 재설치
```

### Scenario 4: 템플릿 일관성 (신규 프로젝트)

```gherkin
Given moai-adk 0.25.11 이상 설치
  And 신규 프로젝트 템플릿 생성

When 신규 프로젝트 생성 완료

Then 다음이 보장되어야 함:
  - .claude/settings.json에 'uvx moai-adk statusline' 포함
  - statusline이 즉시 작동
  - 설정 수정 불필요

Test:
  1. moai-adk init <project-name>
  2. cat .claude/settings.json | grep "uvx moai-adk statusline"
  3. moai-adk statusline (테스트)
```

---

## 품질 게이트 (Quality Gate Criteria)

### 테스트 커버리지
- **최소 기준**: ≥ 85%
- **대상 파일**:
  - `src/moai_adk/statusline/main.py`
  - `src/moai_adk/cli.py` (새 서브커맨드)
  - `src/moai_adk/__main__.py` (진입점)

### 성능 기준
- **첫 실행**: < 5초 (uvx 초기 캐시)
- **캐시 히트**: < 300ms
- **Windows 개선**: 기존 대비 30-40% 개선

### 호환성 기준
- **Python 버전**: 3.11, 3.12, 3.13, 3.14 모두 작동
- **OS**: Windows, macOS, Linux 모두 작동
- **uv 버전**: ≥ 0.4.0 (uvx 지원)

### 코드 품질 기준
- **Linting**: ruff 모든 규칙 통과
- **Type Checking**: mypy 0 오류
- **Documentation**: 모든 함수에 docstring

---

## 검증 체크리스트

### 기능 검증
- [ ] CLI 진입점이 모든 OS에서 실행됨
- [ ] statusline이 Claude Code에 표시됨
- [ ] uvx 캐시가 효과적으로 작동함
- [ ] 모든 OS에서 출력이 일관됨
- [ ] 템플릿이 동기화됨
- [ ] 기존 uv run 방식 호환성 유지

### 성능 검증
- [ ] Windows에서 30-40% 성능 개선 확인
- [ ] 캐시 히트 시 < 300ms 확인
- [ ] 첫 실행 < 5초 확인

### 호환성 검증
- [ ] Python 3.11+ 모두 테스트
- [ ] Windows, macOS, Linux 모두 테스트
- [ ] uv ≥ 0.4.0 확인

### 코드 품질 검증
- [ ] 테스트 커버리지 ≥ 85%
- [ ] ruff 정검사 통과
- [ ] mypy 타입 검사 통과
- [ ] 모든 공개 함수에 docstring

---

## 성공 기준 요약

| 항목 | 기준 | 검증 방법 |
|------|------|---------|
| **기능** | 모든 OS에서 uvx로 실행 | 수동 테스트 + 자동 테스트 |
| **성능** | 기존 대비 30-40% 개선 | 벤치마크 스크립트 |
| **호환성** | Python 3.11+, 모든 OS | CI/CD 테스트 |
| **품질** | 커버리지 ≥ 85% | pytest 리포트 |
| **문서** | 마이그레이션 가이드 작성 | 리뷰 및 병합 |

**모든 기준이 만족되어야 SPEC-STATUSLINE-UVX-001이 완료됩니다.**


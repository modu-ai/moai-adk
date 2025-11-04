# ✅ Acceptance Criteria: UV Tool Upgrade Cache Refresh Auto-Retry

> **SPEC Reference**: @SPEC:UPDATE-CACHE-FIX-001
> **Related Tests**: @TEST:UPDATE-CACHE-FIX-001
> **Author**: @goos
> **Created**: 2025-10-30
> **Version**: v0.0.1

## 테스트 시나리오

### 시나리오 1: 캐시 오래됨 자동 감지 및 재시도

**Given** (전제 조건)
- PyPI에 moai-adk 0.9.1 버전이 존재함
- 로컬에 moai-adk 0.8.3 버전이 설치되어 있음
- 로컬 uv 캐시가 오래되어 0.9.0 정보만 가지고 있음
- uv CLI 0.9.3+ 버전이 설치되어 있음

**When** (실행 조건)
- 사용자가 `moai-adk update` 명령을 실행함

**Then** (기대 결과)
- ✅ 첫 번째 업그레이드 시도에서 "Nothing to upgrade" 메시지 감지
- ✅ 시스템이 캐시 오래됨을 인식하고 "[yellow]⚠️ Cache outdated, refreshing...[/yellow]" 메시지 표시
- ✅ `uv cache clean moai-adk` 명령 자동 실행
- ✅ "[cyan]♻️ Cache cleared, retrying upgrade...[/cyan]" 메시지 표시
- ✅ 업그레이드를 재시도하여 0.9.1 설치에 성공
- ✅ 최종 버전 확인: 0.9.1
- ✅ 전체 프로세스가 1회 명령 실행으로 완료됨

**Expected Output**:
```
⚠️ Cache outdated, refreshing...
♻️ Cache cleared, retrying upgrade...
✓ Updated moai-adk 0.8.3 -> 0.9.1
```

**Test Code**:
```python
def test_scenario_stale_cache_auto_retry(mocker):
    """
    캐시 스테일 시 자동 재시도 시나리오 테스트

    @TEST:UPDATE-CACHE-FIX-001-SCENARIO-1
    """
    # Setup: Mock subprocess calls
    mock_run = mocker.patch("subprocess.run")

    # 첫 번째 호출: "Nothing to upgrade" (캐시 오래됨)
    first_call = mocker.Mock()
    first_call.returncode = 0
    first_call.stdout = "Nothing to upgrade"

    # 두 번째 호출: 업그레이드 성공
    second_call = mocker.Mock()
    second_call.returncode = 0
    second_call.stdout = "Updated moai-adk v0.8.3 -> v0.9.1"

    mock_run.side_effect = [first_call, second_call]

    # Mock version functions
    mocker.patch(
        "moai_adk.cli.commands.update._get_current_version",
        return_value="0.8.3"
    )
    mocker.patch(
        "moai_adk.cli.commands.update._get_latest_version",
        return_value="0.9.1"
    )
    mocker.patch(
        "moai_adk.cli.commands.update._clear_uv_package_cache",
        return_value=True
    )

    # Execute
    from moai_adk.cli.commands.update import _execute_upgrade_with_retry
    result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])

    # Assert
    assert result is True
    assert mock_run.call_count == 2  # 첫 시도 + 재시도
```

**수동 검증 절차**:
1. PyPI에 새 버전 존재 확인: `uv tool list | grep moai-adk`
2. 로컬 캐시 강제 오래되게 만들기: `touch ~/.cache/uv/simple-v18/pypi/moai-adk.rkyv --date='2024-01-01'`
3. `moai-adk update` 실행
4. 출력 메시지 확인: "Cache outdated" → "Cache cleared" → "Updated"
5. 최종 버전 확인: `moai-adk --version`

---

### 시나리오 2: 캐시 최신 상태 (재시도 불필요)

**Given** (전제 조건)
- PyPI에 moai-adk 0.9.0 버전이 최신임
- 로컬에 moai-adk 0.9.0 버전이 설치되어 있음
- 로컬 uv 캐시도 0.9.0 정보를 가지고 있음 (최신 상태)

**When** (실행 조건)
- 사용자가 `moai-adk update` 명령을 실행함

**Then** (기대 결과)
- ✅ 시스템은 "Already up to date" 또는 "Nothing to upgrade" 메시지 표시
- ✅ 재시도 로직은 실행되지 않음
- ✅ 캐시 정리 메시지가 표시되지 않음
- ✅ 정상 종료 (returncode == 0)
- ✅ 버전 변경 없음

**Expected Output**:
```
✓ moai-adk is already up to date (0.9.0)
```

**Test Code**:
```python
def test_scenario_fresh_cache_no_retry(mocker):
    """
    캐시 최신 상태 시 재시도 불필요 시나리오 테스트

    @TEST:UPDATE-CACHE-FIX-001-SCENARIO-2
    """
    # Setup: Mock subprocess call
    mock_run = mocker.patch("subprocess.run")
    mock_run.return_value.returncode = 0
    mock_run.return_value.stdout = "Already up to date"

    # Mock version functions (same version)
    mocker.patch(
        "moai_adk.cli.commands.update._get_current_version",
        return_value="0.9.0"
    )
    mocker.patch(
        "moai_adk.cli.commands.update._get_latest_version",
        return_value="0.9.0"
    )

    # Mock cache clear (should not be called)
    mock_clear = mocker.patch(
        "moai_adk.cli.commands.update._clear_uv_package_cache"
    )

    # Execute
    from moai_adk.cli.commands.update import _execute_upgrade_with_retry
    result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])

    # Assert
    assert result is True
    assert mock_run.call_count == 1  # 첫 시도만
    mock_clear.assert_not_called()  # 캐시 정리 호출 없음
```

**수동 검증 절차**:
1. 최신 버전 설치 확인: `moai-adk --version`
2. PyPI 최신 버전과 동일한지 확인: `curl -s https://pypi.org/pypi/moai-adk/json | jq -r .info.version`
3. `moai-adk update` 실행
4. "Already up to date" 메시지 확인
5. 캐시 정리 메시지 없음 확인

---

### 시나리오 3: 재시도 후에도 실패 (영구 오류)

**Given** (전제 조건)
- PyPI 접근이 불가능함 (네트워크 오류 또는 PyPI 다운)
- 또는 uv 명령어가 손상됨

**When** (실행 조건)
- 사용자가 `moai-adk update` 명령을 실행함
- 첫 시도: "Nothing to upgrade" 출력
- 캐시 스테일 감지됨
- 캐시 정리 실패 또는 재시도 업그레이드 실패

**Then** (기대 결과)
- ✅ 캐시 정리 시도 메시지 표시
- ✅ 재시도 로직 실행
- ✅ 재시도가 실패하면 "[red]✗ Upgrade failed after retry[/red]" 메시지 표시
- ✅ 수동 해결 방법 안내: "[cyan]uv cache clean moai-adk && moai-adk update[/cyan]"
- ✅ 무한 루프 방지 확인 (max_retries=1 보장)
- ✅ returncode != 0 반환

**Expected Output** (캐시 정리 실패):
```
⚠️ Cache outdated, refreshing...
✗ Cache clear failed. Manual workaround:
  uv cache clean moai-adk && moai-adk update
```

**Expected Output** (재시도 업그레이드 실패):
```
⚠️ Cache outdated, refreshing...
♻️ Cache cleared, retrying upgrade...
✗ Upgrade failed after retry
```

**Test Code**:
```python
def test_scenario_persistent_failure_cache_clear(mocker):
    """
    캐시 정리 실패 시나리오 테스트

    @TEST:UPDATE-CACHE-FIX-001-SCENARIO-3A
    """
    # Setup: Mock subprocess call
    mock_run = mocker.patch("subprocess.run")
    mock_run.return_value.returncode = 0
    mock_run.return_value.stdout = "Nothing to upgrade"

    # Mock version functions (stale cache detected)
    mocker.patch(
        "moai_adk.cli.commands.update._get_current_version",
        return_value="0.8.3"
    )
    mocker.patch(
        "moai_adk.cli.commands.update._get_latest_version",
        return_value="0.9.1"
    )

    # Mock cache clear failure
    mocker.patch(
        "moai_adk.cli.commands.update._clear_uv_package_cache",
        return_value=False
    )

    # Execute
    from moai_adk.cli.commands.update import _execute_upgrade_with_retry
    result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])

    # Assert
    assert result is False
    assert mock_run.call_count == 1  # 첫 시도만 (재시도 없음)


def test_scenario_persistent_failure_retry(mocker):
    """
    재시도 업그레이드 실패 시나리오 테스트

    @TEST:UPDATE-CACHE-FIX-001-SCENARIO-3B
    """
    # Setup: Mock subprocess calls
    mock_run = mocker.patch("subprocess.run")

    # 첫 번째 호출: "Nothing to upgrade"
    first_call = mocker.Mock()
    first_call.returncode = 0
    first_call.stdout = "Nothing to upgrade"

    # 두 번째 호출: 업그레이드 실패
    second_call = mocker.Mock()
    second_call.returncode = 1
    second_call.stdout = ""
    second_call.stderr = "Network timeout"

    mock_run.side_effect = [first_call, second_call]

    # Mock version functions
    mocker.patch(
        "moai_adk.cli.commands.update._get_current_version",
        return_value="0.8.3"
    )
    mocker.patch(
        "moai_adk.cli.commands.update._get_latest_version",
        return_value="0.9.1"
    )
    mocker.patch(
        "moai_adk.cli.commands.update._clear_uv_package_cache",
        return_value=True
    )

    # Execute
    from moai_adk.cli.commands.update import _execute_upgrade_with_retry
    result = _execute_upgrade_with_retry(["uv", "tool", "upgrade", "moai-adk"])

    # Assert
    assert result is False
    assert mock_run.call_count == 2  # 첫 시도 + 재시도 (2회만)
```

**수동 검증 절차**:
1. 네트워크 연결 끊기 또는 PyPI 차단
2. `moai-adk update` 실행
3. 캐시 정리 시도 메시지 확인
4. 수동 해결 방법 안내 메시지 확인
5. 무한 루프 없이 종료되는지 확인

---

### 시나리오 4: 다른 패키지 업그레이드 (영향 없음)

**Given** (전제 조건)
- 사용자가 다른 패키지 (예: pytest, requests)를 업그레이드 중
- moai-adk의 재시도 로직이 구현되어 있음

**When** (실행 조건)
- `uv tool upgrade pytest` 또는 `pip install --upgrade pytest` 실행

**Then** (기대 결과)
- ✅ 이 로직은 moai-adk 업그레이드에만 적용됨
- ✅ pytest 업그레이드는 캐시 재시도 로직의 영향을 받지 않음
- ✅ 정상 동작
- ✅ moai-adk 관련 메시지 표시 없음

**Test Code**:
```python
def test_scenario_other_package_not_affected(mocker):
    """
    다른 패키지 업그레이드 시 영향 없음 시나리오 테스트

    @TEST:UPDATE-CACHE-FIX-001-SCENARIO-4
    """
    # Setup: Mock subprocess call for different package
    mock_run = mocker.patch("subprocess.run")
    mock_run.return_value.returncode = 0
    mock_run.return_value.stdout = "Successfully upgraded pytest"

    # Mock cache clear (should not be called for other packages)
    mock_clear = mocker.patch(
        "moai_adk.cli.commands.update._clear_uv_package_cache"
    )

    # Execute upgrade for different package
    # (이 테스트는 실제 구현에서 moai-adk만 재시도 로직 적용하는지 확인)
    result = subprocess.run(["uv", "tool", "upgrade", "pytest"])

    # Assert
    assert result.returncode == 0
    mock_clear.assert_not_called()  # moai-adk가 아니므로 캐시 정리 없음
```

**수동 검증 절차**:
1. 다른 패키지 설치: `uv tool install pytest`
2. `uv tool upgrade pytest` 실행
3. moai-adk 관련 메시지가 표시되지 않는지 확인
4. pytest 정상 업그레이드 확인

---

### 시나리오 5: 버전 파싱 실패 (Graceful Degradation)

**Given** (전제 조건)
- 로컬 또는 PyPI 버전 정보가 잘못된 형식임 (예: "dev", "unknown")
- `_get_current_version()` 또는 `_get_latest_version()`이 비정상 값을 반환

**When** (실행 조건)
- 사용자가 `moai-adk update` 명령을 실행함
- 버전 파싱 중 `InvalidVersion` 예외 발생

**Then** (기대 결과)
- ✅ 예외가 조용히 처리됨 (DEBUG 로그 기록)
- ✅ 재시도 로직이 건너뛰어짐
- ✅ 원래 업그레이드 결과 반환
- ✅ 사용자에게 에러 표시 없음 (graceful degradation)

**Test Code**:
```python
def test_scenario_version_parsing_failure(mocker):
    """
    버전 파싱 실패 시 graceful degradation 시나리오 테스트

    @TEST:UPDATE-CACHE-FIX-001-SCENARIO-5
    """
    # Setup: Mock subprocess call
    mock_run = mocker.patch("subprocess.run")
    mock_run.return_value.returncode = 0
    mock_run.return_value.stdout = "Nothing to upgrade"

    # Mock version functions returning invalid versions
    mocker.patch(
        "moai_adk.cli.commands.update._get_current_version",
        return_value="dev"
    )
    mocker.patch(
        "moai_adk.cli.commands.update._get_latest_version",
        return_value="unknown"
    )

    # Mock logger to verify DEBUG log
    mock_logger = mocker.patch("moai_adk.cli.commands.update.logger")

    # Execute
    from moai_adk.cli.commands.update import _detect_stale_cache
    result = _detect_stale_cache("Nothing to upgrade", "dev", "unknown")

    # Assert
    assert result is False  # 재시도 건너뛰기
    mock_logger.debug.assert_called()  # DEBUG 로그 확인
```

---

## 검증 체크리스트

### 기능 검증

- [ ] **1회 실행 성공**: 캐시 오래됨 상태에서 `moai-adk update` 1회로 업그레이드 완료
- [ ] **자동 감지**: "Nothing to upgrade" 메시지를 자동으로 감지
- [ ] **버전 비교**: 현재 버전 < 최신 버전일 때만 재시도
- [ ] **캐시 정리**: `uv cache clean moai-adk` 명령이 자동 실행됨
- [ ] **사용자 피드백**: 진행 상황이 명확한 메시지로 표시됨 (⚠️, ♻️, ✓, ✗)
- [ ] **에러 처리**: 실패 시 수동 해결 방법 제시
- [ ] **재시도 제한**: 최대 1회만 재시도 (무한 루프 방지)
- [ ] **다른 패키지 미영향**: moai-adk 이외 패키지에 영향 없음
- [ ] **Graceful Degradation**: 버전 파싱 실패 시 조용히 실패

### 코드 품질

- [ ] **테스트 커버리지**: 90%+ 달성 (캐시 fix 코드 대상)
- [ ] **Error handling**: 모든 예외 처리됨 (TimeoutExpired, FileNotFoundError, InvalidVersion)
- [ ] **Logging**: DEBUG/INFO/WARNING 레벨로 적절히 로깅
- [ ] **Code style**: Ruff 린트 통과 (no warnings)
- [ ] **Type hints**: 모든 함수에 타입 힌트 적용
- [ ] **Docstrings**: 모든 함수에 docstring 작성 (Args, Returns, @TAG 포함)
- [ ] **단위 테스트**: 5개 핵심 시나리오 테스트 통과
- [ ] **Integration test**: 기존 update 테스트 여전히 통과 (회귀 방지)

### 문서

- [ ] **README.md**: Troubleshooting 섹션 추가 (@DOC:UPDATE-CACHE-FIX-001-001)
- [ ] **CHANGELOG.md**: Bug fix 기록 (@DOC:UPDATE-CACHE-FIX-001-002)
- [ ] **SPEC 문서**: spec.md, plan.md, acceptance.md 완성
- [ ] **TAG 참조**: @CODE/@TEST/@DOC 링크 완성
- [ ] **수동 검증 절차**: 각 시나리오별 검증 절차 문서화

### CI/CD

- [ ] **Unit tests**: 모든 테스트 통과
- [ ] **Linting**: Ruff 체크 통과
- [ ] **Type checking**: mypy 검증 통과
- [ ] **Build**: 빌드 성공
- [ ] **Integration tests**: 기존 update 테스트 여전히 통과
- [ ] **Platform tests**: Linux, macOS, Windows 모두 테스트 통과

### 사용자 경험

- [ ] **명확한 메시지**: 진행 상황이 사용자에게 명확히 전달됨
- [ ] **투명한 재시도**: 사용자가 무슨 일이 일어나는지 이해할 수 있음
- [ ] **수동 해결 방법**: 실패 시 명확한 해결 방법 제시
- [ ] **응답 시간**: 캐시 정리 + 재시도가 10초 이내 완료
- [ ] **무음 실패**: 버전 파싱 실패 등은 사용자에게 에러 표시 안 함

---

## 성공 기준 (Definition of Done)

**모든 항목이 체크되어야 SPEC이 완료됨**:

### Phase 1: 개발 완료
1. ✅ 3개 신규 함수 구현 완료 (`_detect_stale_cache`, `_clear_uv_package_cache`, `_execute_upgrade_with_retry`)
2. ✅ 5개 시나리오 테스트 통과 (시나리오 1~5)
3. ✅ 테스트 커버리지 90%+
4. ✅ Ruff 린트 통과
5. ✅ mypy 타입 체크 통과

### Phase 2: 품질 보증
6. ✅ CI/CD 파이프라인 성공 (Linux, macOS, Windows)
7. ✅ 기존 update 테스트 회귀 방지 확인
8. ✅ 수동 E2E 테스트 완료 (각 시나리오별)
9. ✅ 코드 리뷰 승인 (최소 1명)

### Phase 3: 문서화
10. ✅ README.md 업데이트 (Troubleshooting 섹션)
11. ✅ CHANGELOG.md 업데이트 (v0.9.1 Fixed 섹션)
12. ✅ @TAG 체인 검증 완료
13. ✅ 모든 함수에 docstring 작성

### Phase 4: 배포
14. ✅ GitHub Issue 생성 (Team 모드) 또는 Draft PR 생성 (Personal 모드)
15. ✅ CodeRabbit 자동 승인 (AI 코드 리뷰)
16. ✅ PR 병합 후 v0.9.1 릴리즈 준비

---

## 품질 게이트

### Gate 1: 단위 테스트 (필수)
- **조건**: 모든 단위 테스트 통과 + 90%+ 커버리지
- **책임자**: tdd-implementer
- **도구**: pytest, pytest-cov
- **기준**: `pytest tests/unit/test_update_uv_cache_fix.py -v --cov=src/moai_adk/cli/commands/update`

### Gate 2: 코드 품질 (필수)
- **조건**: Ruff 린트 통과 + mypy 타입 체크 통과
- **책임자**: trust-checker
- **도구**: ruff, mypy
- **기준**: `ruff check src/ tests/` + `mypy src/`

### Gate 3: Integration 테스트 (필수)
- **조건**: 기존 update 테스트 여전히 통과
- **책임자**: tdd-implementer
- **도구**: pytest
- **기준**: `pytest tests/integration/test_update.py -v`

### Gate 4: E2E 테스트 (권장)
- **조건**: 실제 환경에서 5개 시나리오 수동 검증
- **책임자**: @goos (프로젝트 오너)
- **도구**: 수동 테스트
- **기준**: 각 시나리오별 검증 절차 완료

---

## 롤백 계획

### 롤백 트리거 조건

다음 중 하나라도 발생하면 롤백 고려:

1. **치명적 버그**: 업그레이드가 완전히 실패하는 경우
2. **무한 루프**: max_retries=1 제한이 작동하지 않는 경우
3. **성능 저하**: 캐시 정리로 인해 업그레이드 시간이 30초 이상 소요
4. **호환성 문제**: 특정 플랫폼(Windows)에서 작동하지 않는 경우
5. **회귀 버그**: 기존 update 기능이 손상된 경우

### 롤백 옵션

#### 옵션 1: 기능 플래그로 비활성화 (선호)
```json
// .moai/config.json
{
  "update": {
    "enable_cache_retry": false
  }
}
```

**장점**: 코드 변경 없이 기능 비활성화 가능
**단점**: 사용자가 수동으로 config 수정 필요

#### 옵션 2: Git revert (빠른 롤백)
```bash
git revert <commit-hash>
git push origin feature/update-cache-fix-001
```

**장점**: 즉시 롤백 가능
**단점**: 커밋 히스토리에 revert 기록 남음

#### 옵션 3: 사용자 다운그레이드 안내
```bash
uv tool install moai-adk==0.9.0 --force
```

**장점**: 사용자가 직접 안정 버전으로 복구 가능
**단점**: 사용자 불편 증가

#### 옵션 4: Hotfix 릴리즈 (v0.9.1.1)
- 버그 수정 후 즉시 패치 버전 릴리즈
- 자동 업그레이드로 사용자에게 배포

**장점**: 문제 해결과 동시에 개선 가능
**단점**: 추가 개발 시간 필요

### 롤백 후 조치

1. **이슈 트래킹**: GitHub Issue 생성하여 롤백 이유 기록
2. **사용자 안내**: README에 수동 해결 방법 명시
3. **근본 원인 분석**: 롤백 원인 분석 후 재구현 계획 수립
4. **테스트 강화**: 롤백 원인에 대한 테스트 추가

---

## 참고 자료

### 관련 문서
- **SPEC**: `.moai/specs/SPEC-UPDATE-CACHE-FIX-001/spec.md`
- **Plan**: `.moai/specs/SPEC-UPDATE-CACHE-FIX-001/plan.md`
- **README**: `README.md#troubleshooting-uv-tool-upgrade-issues`
- **CHANGELOG**: `CHANGELOG.md#v0.9.1-fixed`

### 관련 코드
- **구현**: `src/moai_adk/cli/commands/update.py`
- **테스트**: `tests/unit/test_update_uv_cache_fix.py`

### 외부 참조
- **uv 문서**: https://github.com/astral-sh/uv
- **uv cache 명령어**: https://docs.astral.sh/uv/reference/cli/#uv-cache-clean
- **packaging.version**: https://packaging.pypa.io/en/stable/version.html

### 관련 SPEC
- **UPDATE-REFACTOR-001**: 기존 update 명령어 리팩토링
- **TRUST-001**: TRUST 5 원칙 (Test First, Readable, Unified, Secured, Trackable)

---

## 후속 작업

### v0.9.2 (단기)
- [ ] `--no-retry` CLI 플래그 구현
- [ ] `MOAI_ADK_UPDATE_NO_RETRY` 환경 변수 지원
- [ ] `--verbose` 모드에서 상세한 디버그 로그 표시

### v0.10.0 (중기)
- [ ] 캐시 만료 시간 설정 가능 (`update.cache_ttl`)
- [ ] 여러 패키지에 대한 일괄 캐시 정리
- [ ] 캐시 상태 진단 명령어 (`moai-adk cache status`)
- [ ] 자동 캐시 정리 주기 설정

### v1.0.0 (장기)
- [ ] 캐시 전략 고도화 (LRU, TTL 기반)
- [ ] 오프라인 모드 지원
- [ ] 프록시 환경 지원 강화

---

**문서 상태**: DRAFT
**승인 상태**: STEP 2 완료, git-manager로 전달 대기
**다음 단계**: `/alfred:2-run SPEC-UPDATE-CACHE-FIX-001` 실행 (구현 단계)
